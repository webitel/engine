package app

import (
	"context"
	"fmt"
	"strings"

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

	var oauthConf oauth2.Config
	var ok bool

	if strings.Index(profile.ImapHost, model.MailGmail+".com") > -1 {
		oauthConf, ok = a.MailOauthConfig(model.MailGmail)
	} else if strings.Index(profile.ImapHost, model.MailOutlook) == 0 {
		oauthConf, ok = a.MailOauthConfig(model.MailOutlook)
	}

	if !ok {
		return nil, model.NewForbiddenError("app.email.profile.login.not_found_oauth", "Not found server oauth config to "+profile.ImapHost)
	}

	oauthState, err := a.EncryptId(int64(profile.Id))
	if err != nil {
		return nil, err
	}

	return &model.EmailProfileLogin{
		AuthType:    profile.AuthType,
		RedirectUrl: oauthConf.AuthCodeURL(oauthState, oauth2.AccessTypeOffline, oauth2.SetAuthURLParam("approval_prompt", "force")),
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
	return app.Store.EmailProfile().SetupOAuth2(ctx, id, &model.MailProfileParams{
		OAuth2: token,
	})
}

func (app *App) MailOauthConfig(name string) (oauth2.Config, bool) {
	if app.config.EmailOAuth == nil {
		return oauth2.Config{}, false
	}

	p, ok := app.config.EmailOAuth[name]
	return p, ok
}
