package controller

import (
	"github.com/webitel/engine/auth_manager"
	"github.com/webitel/engine/model"
)

func (c *Controller) SearchPauseCause(session *auth_manager.Session, search *model.SearchPauseCause) ([]*model.PauseCause, bool, *model.AppError) {
	permission := session.GetPermission(model.PERMISSION_SCOPE_DICTIONARIES)
	if !permission.CanRead() {
		return nil, false, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	return c.app.GetPauseCausePage(session.Domain(search.DomainId), search)
}

func (c *Controller) CreatePauseCause(session *auth_manager.Session, cause *model.PauseCause) (*model.PauseCause, *model.AppError) {
	permission := session.GetPermission(model.PERMISSION_SCOPE_DICTIONARIES)
	if !permission.CanCreate() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_CREATE)
	}

	if err := cause.IsValid(); err != nil {
		return nil, err
	}
	cause.CreatedBy.Id = int(session.UserId)
	cause.UpdatedBy = cause.CreatedBy

	cause.CreatedAt = model.GetTime()
	cause.UpdatedAt = cause.CreatedAt

	return c.app.CreatePauseCause(session.Domain(0), cause)
}

func (c *Controller) GetPauseCause(session *auth_manager.Session, id uint32) (*model.PauseCause, *model.AppError) {
	permission := session.GetPermission(model.PERMISSION_SCOPE_DICTIONARIES)
	if !permission.CanRead() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	return c.app.GetPauseCause(session.Domain(0), id)
}

func (c *Controller) UpdatePauseCause(session *auth_manager.Session, cause *model.PauseCause) (*model.PauseCause, *model.AppError) {
	permission := session.GetPermission(model.PERMISSION_SCOPE_DICTIONARIES)
	if !permission.CanRead() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if !permission.CanUpdate() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
	}

	if err := cause.IsValid(); err != nil {
		return nil, err
	}

	cause.UpdatedBy.Id = int(session.UserId)
	cause.UpdatedAt = model.GetTime()

	return c.app.UpdatePauseCause(session.DomainId, cause)
}

func (c *Controller) PatchPauseCause(session *auth_manager.Session, id uint32, patch *model.PauseCausePatch) (*model.PauseCause, *model.AppError) {
	permission := session.GetPermission(model.PERMISSION_SCOPE_DICTIONARIES)
	if !permission.CanRead() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if !permission.CanUpdate() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
	}

	patch.UpdatedBy.Id = int(session.UserId)
	patch.UpdatedAt = model.GetTime()

	return c.app.PatchPauseCause(session.DomainId, id, patch)
}

func (c *Controller) DeletePauseCause(session *auth_manager.Session, id uint32) (*model.PauseCause, *model.AppError) {
	permission := session.GetPermission(model.PERMISSION_SCOPE_DICTIONARIES)
	if !permission.CanDelete() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_DELETE)
	}

	return c.app.RemovePauseCause(session.Domain(0), id)
}
