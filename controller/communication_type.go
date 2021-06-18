package controller

import (
	"github.com/webitel/engine/auth_manager"
	"github.com/webitel/engine/model"
)

func (c *Controller) CreateCommunicationType(session *auth_manager.Session, ct *model.CommunicationType) (*model.CommunicationType, *model.AppError) {
	var err *model.AppError
	permission := session.GetPermission(model.PERMISSION_SCOPE_DICTIONARIES)
	if !permission.CanCreate() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_CREATE)
	}

	if err = ct.IsValid(); err != nil {
		return nil, err
	}

	return c.app.CreateCommunicationType(ct)
}

func (c *Controller) GetCommunicationTypePage(session *auth_manager.Session, search *model.SearchCommunicationType) ([]*model.CommunicationType, bool, *model.AppError) {
	permission := session.GetPermission(model.PERMISSION_SCOPE_DICTIONARIES)
	if !permission.CanRead() {
		return nil, false, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	return c.app.GetCommunicationTypePage(session.Domain(0), search)
}

func (c *Controller) ReadCommunicationType(session *auth_manager.Session, id int64) (*model.CommunicationType, *model.AppError) {
	permission := session.GetPermission(model.PERMISSION_SCOPE_DICTIONARIES)
	if !permission.CanRead() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	return c.app.GetCommunicationType(id, session.Domain(0))
}

func (c *Controller) UpdateCommunicationType(session *auth_manager.Session, ct *model.CommunicationType) (*model.CommunicationType, *model.AppError) {
	permission := session.GetPermission(model.PERMISSION_SCOPE_DICTIONARIES)
	if !permission.CanUpdate() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
	}

	ct.DomainId = session.Domain(0)
	return c.app.UpdateCommunicationType(ct)
}

func (c *Controller) RemoveCommunicationType(session *auth_manager.Session, id int64) (*model.CommunicationType, *model.AppError) {
	permission := session.GetPermission(model.PERMISSION_SCOPE_DICTIONARIES)
	if !permission.CanDelete() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_DELETE)
	}

	return c.app.RemoveCommunicationType(session.Domain(0), id)
}
