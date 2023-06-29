package controller

import (
	"context"

	"github.com/webitel/engine/auth_manager"
	"github.com/webitel/engine/model"
)

func (c *Controller) CreateSkill(ctx context.Context, session *auth_manager.Session, s *model.Skill) (*model.Skill, model.AppError) {
	var err model.AppError
	permission := session.GetPermission(model.PERMISSION_SCOPE_DICTIONARIES)
	if !permission.CanCreate() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_CREATE)
	}

	if err = s.IsValid(); err != nil {
		return nil, err
	}

	return c.app.CreateSkill(ctx, s)
}

func (c *Controller) SearchSkill(ctx context.Context, session *auth_manager.Session, search *model.SearchSkill) ([]*model.Skill, bool, model.AppError) {
	permission := session.GetPermission(model.PERMISSION_SCOPE_DICTIONARIES)
	if !permission.CanRead() {
		return nil, false, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	return c.app.GetSkillsPage(ctx, session.Domain(0), search)
}

func (c *Controller) ReadSkill(ctx context.Context, session *auth_manager.Session, id int64) (*model.Skill, model.AppError) {
	permission := session.GetPermission(model.PERMISSION_SCOPE_DICTIONARIES)
	if !permission.CanRead() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	return c.app.GetSkill(ctx, id, session.Domain(0))
}

func (c *Controller) UpdateSkill(ctx context.Context, session *auth_manager.Session, s *model.Skill) (*model.Skill, model.AppError) {
	permission := session.GetPermission(model.PERMISSION_SCOPE_DICTIONARIES)
	if !permission.CanUpdate() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
	}

	if err := s.IsValid(); err != nil {
		return nil, err
	}

	return c.app.UpdateSkill(ctx, s)
}

func (c *Controller) DeleteSkill(ctx context.Context, session *auth_manager.Session, id int64) (*model.Skill, model.AppError) {
	permission := session.GetPermission(model.PERMISSION_SCOPE_DICTIONARIES)
	if !permission.CanDelete() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_DELETE)
	}

	return c.app.RemoveSkill(ctx, session.Domain(0), id)
}
