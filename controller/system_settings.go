package controller

import (
	"context"
	"github.com/webitel/engine/auth_manager"
	"github.com/webitel/engine/model"
)

func (c *Controller) CreateSystemSetting(ctx context.Context, session *auth_manager.Session, s *model.SystemSetting) (*model.SystemSetting, model.AppError) {
	var err model.AppError

	if !session.HasAction(auth_manager.PermissionSystemSetting) {
		return nil, c.app.MakeActionPermissionError(session, auth_manager.PermissionSystemSetting, auth_manager.PERMISSION_ACCESS_CREATE)
	}

	if err = s.IsValid(); err != nil {
		return nil, err
	}

	return c.app.CreateSystemSetting(ctx, session.UserId, session.DomainId, s)
}

func (c *Controller) SearchSystemSetting(ctx context.Context, session *auth_manager.Session, search *model.SearchSystemSetting) ([]*model.SystemSetting, bool, model.AppError) {

	return c.app.GetSystemSettingPage(ctx, session.Domain(0), search)
}

func (c *Controller) SearchAvailableSystemSetting(ctx context.Context, session *auth_manager.Session, search *model.ListRequest) ([]string, model.AppError) {
	if !session.HasAction(auth_manager.PermissionSystemSetting) {
		return nil, c.app.MakeActionPermissionError(session, auth_manager.PermissionSystemSetting, auth_manager.PERMISSION_ACCESS_READ)
	}

	return c.app.GetAvailableSystemSetting(ctx, session.Domain(0), search)
}

func (c *Controller) ReadSystemSetting(ctx context.Context, session *auth_manager.Session, id int32) (*model.SystemSetting, model.AppError) {
	if !session.HasAction(auth_manager.PermissionSystemSetting) {
		return nil, c.app.MakeActionPermissionError(session, auth_manager.PermissionSystemSetting, auth_manager.PERMISSION_ACCESS_READ)
	}

	return c.app.GetSystemSetting(ctx, session.Domain(0), id)
}

func (c *Controller) UpdateSystemSetting(ctx context.Context, session *auth_manager.Session, s *model.SystemSetting) (*model.SystemSetting, model.AppError) {
	if !session.HasAction(auth_manager.PermissionSystemSetting) {
		return nil, c.app.MakeActionPermissionError(session, auth_manager.PermissionSystemSetting, auth_manager.PERMISSION_ACCESS_UPDATE)
	}

	return c.app.UpdateSystemSetting(ctx, session.UserId, session.DomainId, s)
}

func (c *Controller) PatchSystemSetting(ctx context.Context, session *auth_manager.Session, id int32, patch *model.SystemSettingPath) (*model.SystemSetting, model.AppError) {
	if !session.HasAction(auth_manager.PermissionSystemSetting) {
		return nil, c.app.MakeActionPermissionError(session, auth_manager.PermissionSystemSetting, auth_manager.PERMISSION_ACCESS_UPDATE)
	}

	return c.app.PatchSystemSetting(ctx, session.UserId, session.Domain(0), id, patch)
}

func (c *Controller) DeleteSystemSetting(ctx context.Context, session *auth_manager.Session, id int32) (*model.SystemSetting, model.AppError) {
	if !session.HasAction(auth_manager.PermissionSystemSetting) {
		return nil, c.app.MakeActionPermissionError(session, auth_manager.PermissionSystemSetting, auth_manager.PERMISSION_ACCESS_DELETE)
	}
	return c.app.RemoveSystemSetting(ctx, session.Domain(0), id)
}
