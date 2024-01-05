package controller

import (
	"context"
	"github.com/webitel/engine/auth_manager"
	"github.com/webitel/engine/model"
)

func (c *Controller) SearchWebHook(ctx context.Context, session *auth_manager.Session, search *model.SearchWebHook) ([]*model.WebHook, bool, model.AppError) {
	permission := session.GetPermission(model.PermissionWebHook)
	if !permission.CanRead() {
		return nil, false, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	return c.app.SearchWebHook(ctx, session.Domain(search.DomainId), search)
}

func (c *Controller) CreateWebHook(ctx context.Context, session *auth_manager.Session, hook *model.WebHook) (*model.WebHook, model.AppError) {
	permission := session.GetPermission(model.PermissionWebHook)
	if !permission.CanCreate() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_CREATE)
	}

	if err := hook.IsValid(); err != nil {
		return nil, err
	}
	hook.CreatedBy = &model.Lookup{
		Id: int(session.UserId),
	}
	hook.UpdatedBy = hook.CreatedBy

	hook.CreatedAt = model.GetTime()
	hook.UpdatedAt = hook.CreatedAt

	return c.app.CreateWebHook(ctx, session.Domain(0), hook)
}

func (c *Controller) GetWebHook(ctx context.Context, session *auth_manager.Session, id int32) (*model.WebHook, model.AppError) {
	permission := session.GetPermission(model.PermissionWebHook)
	if !permission.CanRead() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	return c.app.GetWebHook(ctx, session.Domain(0), id)
}

func (c *Controller) UpdateWebHook(ctx context.Context, session *auth_manager.Session, hook *model.WebHook) (*model.WebHook, model.AppError) {
	permission := session.GetPermission(model.PermissionWebHook)
	if !permission.CanRead() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if !permission.CanUpdate() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
	}

	if err := hook.IsValid(); err != nil {
		return nil, err
	}

	hook.UpdatedBy = &model.Lookup{
		Id: int(session.UserId),
	}
	hook.UpdatedAt = model.GetTime()

	return c.app.UpdateWebHook(ctx, session.DomainId, hook)
}

func (c *Controller) PatchWebHook(ctx context.Context, session *auth_manager.Session, id int32, patch *model.WebHookPatch) (*model.WebHook, model.AppError) {
	permission := session.GetPermission(model.PermissionWebHook)
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

	return c.app.PatchWebHook(ctx, session.DomainId, id, patch)
}

func (c *Controller) DeleteWebHook(ctx context.Context, session *auth_manager.Session, id int32) (*model.WebHook, model.AppError) {
	permission := session.GetPermission(model.PermissionWebHook)
	if !permission.CanDelete() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_DELETE)
	}

	return c.app.RemoveWebHook(ctx, session.Domain(0), id)
}
