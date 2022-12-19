package app

import (
	"github.com/webitel/engine/model"
	"golang.org/x/oauth2"
	"net/http"
	"strings"
)

func (app *App) CreateEmailProfile(domainId int64, profile *model.EmailProfile) (*model.EmailProfile, *model.AppError) {
	return app.Store.EmailProfile().Create(domainId, profile)
}

func (a *App) GetEmailProfilesPage(domainId int64, search *model.SearchEmailProfile) ([]*model.EmailProfile, bool, *model.AppError) {
	list, err := a.Store.EmailProfile().GetAllPage(domainId, search)
	if err != nil {
		return nil, false, err
	}
	search.RemoveLastElemIfNeed(&list)
	return list, search.EndOfList(), nil
}

func (a *App) GetEmailProfile(domainId int64, id int) (*model.EmailProfile, *model.AppError) {
	return a.Store.EmailProfile().Get(domainId, id)
}

func (a *App) UpdateEmailProfile(domainId int64, p *model.EmailProfile) (*model.EmailProfile, *model.AppError) {
	oldProfile, err := a.GetEmailProfile(domainId, int(p.Id))
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

	oldProfile, err = a.Store.EmailProfile().Update(domainId, oldProfile)
	if err != nil {
		return nil, err
	}

	return oldProfile, nil
}

func (a *App) PatchEmailProfile(domainId int64, id int, patch *model.EmailProfilePatch) (*model.EmailProfile, *model.AppError) {
	oldProfile, err := a.GetEmailProfile(domainId, id)
	if err != nil {
		return nil, err
	}

	oldProfile.Patch(patch)

	if err = oldProfile.IsValid(); err != nil {
		return nil, err
	}

	oldProfile, err = a.Store.EmailProfile().Update(domainId, oldProfile)
	if err != nil {
		return nil, err
	}

	return oldProfile, nil
}

func (a *App) loginEmailProfileOAuth2(profile *model.EmailProfile) (*model.EmailProfileLogin, *model.AppError) {

	var oauthConf oauth2.Config
	var ok bool

	if strings.Index(profile.ImapHost, model.MailGmail+".com") > -1 {
		oauthConf, ok = a.MailOauthConfig(model.MailGmail)
	} else if strings.Index(profile.ImapHost, model.MailOutlook) == 0 {
		oauthConf, ok = a.MailOauthConfig(model.MailOutlook)
	}

	if !ok {
		return nil, model.NewAppError("Email", "app.email.profile.login.not_found_oauth", nil,
			"Not found server oauth config to "+profile.ImapHost, http.StatusForbidden)
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

func (a *App) LoginEmailProfile(domainId int64, id int) (*model.EmailProfileLogin, *model.AppError) {
	profile, err := a.GetEmailProfile(domainId, id)
	if err != nil {
		return nil, err
	}

	switch profile.AuthType {
	case model.EmailAuthTypeOAuth2:
		return a.loginEmailProfileOAuth2(profile)
	}

	return nil, model.NewAppError("Email", "app.email.profile.login.not_found_auth_type", nil,
		"Not found auth type to "+profile.ImapHost, http.StatusForbidden)
}

func (app *App) RemoveEmailProfile(domainId int64, id int) (*model.EmailProfile, *model.AppError) {
	profile, err := app.Store.EmailProfile().Get(domainId, id)

	if err != nil {
		return nil, err
	}

	err = app.Store.EmailProfile().Delete(domainId, id)
	if err != nil {
		return nil, err
	}
	return profile, nil
}

func (app *App) EmailLoginOAuth(id int, token *oauth2.Token) *model.AppError {
	return app.Store.EmailProfile().SetupOAuth2(id, &model.MailProfileParams{
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
