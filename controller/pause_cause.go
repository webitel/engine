package controller

import (
	"context"

	"github.com/webitel/engine/auth_manager"
	"github.com/webitel/engine/model"
)

func (c *Controller) SearchPauseCause(ctx context.Context, session *auth_manager.Session, search *model.SearchPauseCause) ([]*model.PauseCause, bool, model.AppError) {
	permission := session.GetPermission(model.PERMISSION_SCOPE_DICTIONARIES)
	if !permission.CanRead() {
		return nil, false, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	return c.app.GetPauseCausePage(ctx, session.Domain(search.DomainId), search)
}

func (c *Controller) CreatePauseCause(ctx context.Context, session *auth_manager.Session, cause *model.PauseCause) (*model.PauseCause, model.AppError) {
	permission := session.GetPermission(model.PERMISSION_SCOPE_DICTIONARIES)
	if !permission.CanCreate() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_CREATE)
	}

	if err := cause.IsValid(); err != nil {
		return nil, err
	}
	cause.CreatedBy = &model.Lookup{
		Id: int(session.UserId),
	}
	cause.UpdatedBy = cause.CreatedBy

	cause.CreatedAt = model.GetTime()
	cause.UpdatedAt = cause.CreatedAt

	return c.app.CreatePauseCause(ctx, session.Domain(0), cause)
}

func (c *Controller) GetPauseCause(ctx context.Context, session *auth_manager.Session, id uint32) (*model.PauseCause, model.AppError) {
	permission := session.GetPermission(model.PERMISSION_SCOPE_DICTIONARIES)
	if !permission.CanRead() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	return c.app.GetPauseCause(ctx, session.Domain(0), id)
}

func (c *Controller) UpdatePauseCause(ctx context.Context, session *auth_manager.Session, cause *model.PauseCause) (*model.PauseCause, model.AppError) {
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

	cause.UpdatedBy = &model.Lookup{
		Id: int(session.UserId),
	}
	cause.UpdatedAt = model.GetTime()

	return c.app.UpdatePauseCause(ctx, session.DomainId, cause)
}

func (c *Controller) PatchPauseCause(ctx context.Context, session *auth_manager.Session, id uint32, patch *model.PauseCausePatch) (*model.PauseCause, model.AppError) {
	permission := session.GetPermission(model.PERMISSION_SCOPE_DICTIONARIES)
	if !permission.CanRead() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if !permission.CanUpdate() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
	}

	patch.UpdatedBy = model.Lookup{
		Id: int(session.UserId),
	}
	patch.UpdatedAt = model.GetTime()

	return c.app.PatchPauseCause(ctx, session.DomainId, id, patch)
}

func (c *Controller) DeletePauseCause(ctx context.Context, session *auth_manager.Session, id uint32) (*model.PauseCause, model.AppError) {
	permission := session.GetPermission(model.PERMISSION_SCOPE_DICTIONARIES)
	if !permission.CanDelete() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_DELETE)
	}

	return c.app.RemovePauseCause(ctx, session.Domain(0), id)
}
