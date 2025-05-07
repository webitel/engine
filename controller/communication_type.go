package controller

import (
	"context"

	"github.com/webitel/engine/model"
	"github.com/webitel/engine/pkg/wbt/auth_manager"
)

func (c *Controller) CreateCommunicationType(ctx context.Context, session *auth_manager.Session, ct *model.CommunicationType) (*model.CommunicationType, model.AppError) {
	var err model.AppError
	permission := session.GetPermission(model.PERMISSION_SCOPE_DICTIONARIES)
	if !permission.CanCreate() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_CREATE)
	}

	if err = ct.IsValid(); err != nil {
		return nil, err
	}

	return c.app.CreateCommunicationType(ctx, session.Domain(0), ct)
}

func (c *Controller) GetCommunicationTypePage(ctx context.Context, session *auth_manager.Session, search *model.SearchCommunicationType) ([]*model.CommunicationType, bool, model.AppError) {
	permission := session.GetPermission(model.PERMISSION_SCOPE_DICTIONARIES)
	if !permission.CanRead() {
		return nil, false, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	return c.app.GetCommunicationTypePage(ctx, session.Domain(0), search)
}

func (c *Controller) ReadCommunicationType(ctx context.Context, session *auth_manager.Session, id int64) (*model.CommunicationType, model.AppError) {
	permission := session.GetPermission(model.PERMISSION_SCOPE_DICTIONARIES)
	if !permission.CanRead() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	return c.app.GetCommunicationType(ctx, id, session.Domain(0))
}

func (c *Controller) UpdateCommunicationType(ctx context.Context, session *auth_manager.Session, ct *model.CommunicationType) (*model.CommunicationType, model.AppError) {
	permission := session.GetPermission(model.PERMISSION_SCOPE_DICTIONARIES)
	if !permission.CanUpdate() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
	}

	return c.app.UpdateCommunicationType(ctx, session.Domain(0), ct)
}

func (c *Controller) PatchCommunicationType(ctx context.Context, session *auth_manager.Session, id int64, patch *model.CommunicationTypePatch) (*model.CommunicationType, model.AppError) {
	permission := session.GetPermission(model.PERMISSION_SCOPE_DICTIONARIES)
	if !permission.CanUpdate() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
	}

	return c.app.PatchCommunicationType(ctx, session.Domain(0), id, patch)
}

func (c *Controller) RemoveCommunicationType(ctx context.Context, session *auth_manager.Session, id int64) (*model.CommunicationType, model.AppError) {
	permission := session.GetPermission(model.PERMISSION_SCOPE_DICTIONARIES)
	if !permission.CanDelete() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_DELETE)
	}

	return c.app.RemoveCommunicationType(ctx, session.Domain(0), id)
}
