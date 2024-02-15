package app

import (
	"context"
	"fmt"
	"github.com/webitel/engine/auth_manager"
	"github.com/webitel/engine/model"
	"golang.org/x/oauth2"
)

func (app *App) CountActiveEmailProfile(ctx context.Context, domainId int64) (int, model.AppError) {
	return app.Store.EmailProfile().CountEnabledByDomain(ctx, domainId)
}

func (app *App) ConstraintEmailProfileLimit(ctx context.Context, domainId int64, token string) model.AppError {
	count, err := app.CountActiveEmailProfile(ctx, domainId)
	if err != nil {
		return err
	}

	limit, errLic := app.sessionManager.ProductLimit(ctx, token, auth_manager.LicenseEmail)
	if errLic != nil {
		return model.NewInternalError("app.email.app_error", errLic.Error())
	}

	if (count + 1) > limit {
		return model.NewInternalError("app.email.valid.license", fmt.Sprintf("mail profile registration is limited; maximum number of active: %d", limit))
	}

	return nil
}

func (app *App) CreateEmailProfile(ctx context.Context, domainId int64, profile *model.EmailProfile) (*model.EmailProfile, model.AppError) {
	return app.Store.EmailProfile().Create(ctx, domainId, profile)
}

func (a *App) GetEmailProfilesPage(ctx context.Context, domainId int64, search *model.SearchEmailProfile) ([]*model.EmailProfile, bool, model.AppError) {
	list, err := a.Store.EmailProfile().GetAllPage(ctx, domainId, search)
	if err != nil {
		return nil, false, err
	}
	search.RemoveLastElemIfNeed(&list)
	return list, search.EndOfList(), nil
}

func (a *App) GetEmailProfile(ctx context.Context, domainId int64, id int) (*model.EmailProfile, model.AppError) {
	return a.Store.EmailProfile().Get(ctx, domainId, id)
}

func (a *App) UpdateEmailProfile(ctx context.Context, domainId int64, p *model.EmailProfile) (*model.EmailProfile, model.AppError) {
	oldProfile, err := a.GetEmailProfile(ctx, domainId, int(p.Id))
	if err != nil {
		return nil, err
	}

	oldProfile.UpdatedBy = p.UpdatedBy
	oldProfile.Name = p.Name
	oldProfile.Description = p.Description
	oldProfile.ImapHost = p.ImapHost
	oldProfile.Login = p.Login
	oldProfile.Password = p.Password
	oldProfile.Mailbox = p.Mailbox
	oldProfile.Schema = p.Schema
	oldProfile.Enabled = p.Enabled
	oldProfile.ImapPort = p.ImapPort
	oldProfile.SmtpPort = p.SmtpPort
	oldProfile.SmtpHost = p.SmtpHost
	oldProfile.FetchInterval = p.FetchInterval
	oldProfile.Params = p.Params

	oldProfile, err = a.Store.EmailProfile().Update(ctx, domainId, oldProfile)
	if err != nil {
		return nil, err
	}

	return oldProfile, nil
}

func (a *App) PatchEmailProfile(ctx context.Context, domainId int64, id int, patch *model.EmailProfilePatch) (*model.EmailProfile, model.AppError) {
	oldProfile, err := a.GetEmailProfile(ctx, domainId, id)
	if err != nil {
		return nil, err
	}

	oldProfile.Patch(patch)

	if err = oldProfile.IsValid(); err != nil {
		return nil, err
	}

	oldProfile, err = a.Store.EmailProfile().Update(ctx, domainId, oldProfile)
	if err != nil {
		return nil, err
	}

	return oldProfile, nil
}

func (a *App) loginEmailProfileOAuth2(profile *model.EmailProfile) (*model.EmailProfileLogin, model.AppError) {
	oauthConf, err := profile.Oauth()
	var domainEncrypt, profileEncrypt string
	if err != nil {
		return nil, err
	}

	profileEncrypt, err = a.EncryptId(int64(profile.Id))
	if err != nil {
		return nil, err
	}

	domainEncrypt, err = a.EncryptId(profile.DomainId)
	if err != nil {
		return nil, err
	}

	oauthState := domainEncrypt + "::" + profileEncrypt

	return &model.EmailProfileLogin{
		AuthType:    profile.AuthType,
		RedirectUrl: oauthConf.AuthCodeURL(oauthState, oauth2.AccessTypeOffline),
		Cookie: map[string]string{
			"oauthstate": oauthState,
		},
	}, nil
}

func (a *App) LoginEmailProfile(ctx context.Context, domainId int64, id int) (*model.EmailProfileLogin, model.AppError) {
	profile, err := a.GetEmailProfile(ctx, domainId, id)
	if err != nil {
		return nil, err
	}

	switch profile.AuthType {
	case model.EmailAuthTypeOAuth2:
		return a.loginEmailProfileOAuth2(profile)
	}

	return nil, model.NewForbiddenError("app.email.profile.login.not_found_auth_type", "Not found auth type to "+profile.ImapHost)
}

func (app *App) RemoveEmailProfile(ctx context.Context, domainId int64, id int) (*model.EmailProfile, model.AppError) {
	profile, err := app.Store.EmailProfile().Get(ctx, domainId, id)

	if err != nil {
		return nil, err
	}

	err = app.Store.EmailProfile().Delete(ctx, domainId, id)
	if err != nil {
		return nil, err
	}
	return profile, nil
}

func (app *App) EmailLoginOAuth(ctx context.Context, id int, token *oauth2.Token) model.AppError {
	return app.Store.EmailProfile().SetupOAuth2(ctx, id, token)
}
