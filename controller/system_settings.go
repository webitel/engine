package controller

import (
	"context"
	"github.com/webitel/engine/auth_manager"
	"github.com/webitel/engine/model"
)

func (c *Controller) CreateSystemSetting(ctx context.Context, session *auth_manager.Session, s *model.SystemSetting) (*model.SystemSetting, model.AppError) {
	var err model.AppError

	// TODO Permission

	if err = s.IsValid(); err != nil {
		return nil, err
	}

	return c.app.CreateSystemSetting(ctx, session.DomainId, s)
}

func (c *Controller) SearchSystemSetting(ctx context.Context, session *auth_manager.Session, search *model.SearchSystemSetting) ([]*model.SystemSetting, bool, model.AppError) {
	// TODO Permission
	return c.app.GetSystemSettingPage(ctx, session.Domain(0), search)
}

func (c *Controller) ReadSystemSetting(ctx context.Context, session *auth_manager.Session, id int32) (*model.SystemSetting, model.AppError) {
	// TODO Permission
	return c.app.GetSystemSetting(ctx, session.Domain(0), id)
}

func (c *Controller) UpdateSystemSetting(ctx context.Context, session *auth_manager.Session, s *model.SystemSetting) (*model.SystemSetting, model.AppError) {
	// TODO Permission
	if err := s.IsValid(); err != nil {
		return nil, err
	}

	return c.app.UpdateSystemSetting(ctx, session.DomainId, s)
}

func (c *Controller) PatchSystemSetting(ctx context.Context, session *auth_manager.Session, id int32, patch *model.SystemSettingPath) (*model.SystemSetting, model.AppError) {
	// TODO Permission

	return c.app.PatchSystemSetting(ctx, session.Domain(0), id, patch)
}

func (c *Controller) DeleteSystemSetting(ctx context.Context, session *auth_manager.Session, id int32) (*model.SystemSetting, model.AppError) {
	// TODO Permission
	return c.app.RemoveSystemSetting(ctx, session.Domain(0), id)
}
