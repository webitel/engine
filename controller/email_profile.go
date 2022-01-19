package controller

import (
	"github.com/webitel/engine/auth_manager"
	"github.com/webitel/engine/model"
)

func (c *Controller) CreateEmailProfile(session *auth_manager.Session, profile *model.EmailProfile) (*model.EmailProfile, *model.AppError) {
	permission := session.GetPermission(model.PERMISSION_SCOPE_EMAIL_PROFILE)
	if !permission.CanCreate() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_CREATE)
	}

	profile.CreatedBy = &model.Lookup{
		Id: int(session.UserId),
	}
	profile.UpdatedBy = profile.CreatedBy
	profile.DomainId = session.Domain(profile.DomainId)

	if err := profile.IsValid(); err != nil {
		return nil, err
	}

	return c.app.CreateEmailProfile(profile)
}

func (c *Controller) SearchEmailProfile(session *auth_manager.Session, search *model.SearchEmailProfile) ([]*model.EmailProfile, bool, *model.AppError) {
	permission := session.GetPermission(model.PERMISSION_SCOPE_EMAIL_PROFILE)
	if !permission.CanRead() {
		return nil, false, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	search.DomainId = session.Domain(search.DomainId) //TODO

	return c.app.GetEmailProfilesPage(session.Domain(search.DomainId), search)
}

func (c *Controller) GetEmailProfile(session *auth_manager.Session, domainId int64, id int) (*model.EmailProfile, *model.AppError) {
	permission := session.GetPermission(model.PERMISSION_SCOPE_EMAIL_PROFILE)
	if !permission.CanRead() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	return c.app.GetEmailProfile(session.Domain(domainId), id)
}

func (c *Controller) UpdateEmailProfile(session *auth_manager.Session, profile *model.EmailProfile) (*model.EmailProfile, *model.AppError) {
	permission := session.GetPermission(model.PERMISSION_SCOPE_EMAIL_PROFILE)
	if !permission.CanRead() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if !permission.CanUpdate() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
	}

	profile.DomainId = session.Domain(profile.DomainId)

	if err := profile.IsValid(); err != nil {
		return nil, err
	}

	return c.app.UpdateEmailProfile(profile)
}

func (c *Controller) RemoveEmailProfile(session *auth_manager.Session, domainId int64, id int) (*model.EmailProfile, *model.AppError) {
	permission := session.GetPermission(model.PERMISSION_SCOPE_EMAIL_PROFILE)
	if !permission.CanRead() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if !permission.CanDelete() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_DELETE)
	}

	return c.app.RemoveEmailProfile(session.Domain(domainId), id)
}
