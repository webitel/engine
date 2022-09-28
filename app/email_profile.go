package app

import (
	"github.com/webitel/engine/model"
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
