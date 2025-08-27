package controller

import (
	"context"

	"github.com/webitel/engine/pkg/wbt/auth_manager"

	"github.com/webitel/engine/model"
)

func (c *Controller) SearchQuickReply(ctx context.Context, session *auth_manager.Session, search *model.SearchQuickReply) ([]*model.QuickReply, bool, model.AppError) {
	permission := session.GetPermission(model.PERMISSION_SCOPE_DICTIONARIES)
	if !permission.CanRead() {
		return nil, false, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	return c.app.GetQuickReplyPage(ctx, session.Domain(search.DomainId), search)
}

func (c *Controller) CreateQuickReply(ctx context.Context, session *auth_manager.Session, cause *model.QuickReply) (*model.QuickReply, model.AppError) {
	permission := session.GetPermission(model.PERMISSION_SCOPE_DICTIONARIES)
	if !permission.CanCreate() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_CREATE)
	}

	if err := cause.IsValid(); err != nil {
		return nil, err
	}

	cause.DomainId = session.Domain(0)
	cause.CreatedAt = model.GetMillis()
	cause.UpdatedAt = cause.CreatedAt
	cause.CreatedBy = &model.Lookup{
		Id: int(session.UserId),
	}

	cause.UpdatedBy = cause.CreatedBy

	return c.app.CreateQuickReply(ctx, session.Domain(0), cause)
}

func (c *Controller) GetQuickReply(ctx context.Context, session *auth_manager.Session, id uint32) (*model.QuickReply, model.AppError) {
	permission := session.GetPermission(model.PERMISSION_SCOPE_DICTIONARIES)
	if !permission.CanRead() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	return c.app.GetQuickReply(ctx, session.Domain(0), id)
}

func (c *Controller) UpdateQuickReply(ctx context.Context, session *auth_manager.Session, cause *model.QuickReply) (*model.QuickReply, model.AppError) {
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

	cause.UpdatedAt = model.GetMillis()
	cause.UpdatedBy = &model.Lookup{
		Id: int(session.UserId),
	}

	return c.app.UpdateQuickReply(ctx, session.DomainId, cause)
}

func (c *Controller) PatchQuickReply(ctx context.Context, session *auth_manager.Session, id uint32, patch *model.QuickReplyPatch) (*model.QuickReply, model.AppError) {
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

	return c.app.PatchQuickReply(ctx, session.DomainId, id, patch)
}

func (c *Controller) DeleteQuickReply(ctx context.Context, session *auth_manager.Session, id uint32) (*model.QuickReply, model.AppError) {
	permission := session.GetPermission(model.PERMISSION_SCOPE_DICTIONARIES)
	if !permission.CanDelete() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_DELETE)
	}

	return c.app.RemoveQuickReply(ctx, session.Domain(0), id)
}
