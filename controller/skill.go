package controller

import (
	"context"
	"github.com/webitel/engine/model"
	"github.com/webitel/engine/pkg/wbt/auth_manager"
	"time"
)

func (c *Controller) CreateSkill(ctx context.Context, session *auth_manager.Session, s *model.Skill) (*model.Skill, model.AppError) {
	var err model.AppError
	permission := session.GetPermission(model.PermissionSkill)
	if !permission.CanCreate() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_CREATE)
	}
	t := time.Now()
	s.CreatedAt = &t
	s.UpdatedAt = s.CreatedAt
	s.CreatedBy = &model.Lookup{
		Id: int(session.UserId),
	}
	s.UpdatedBy = s.CreatedBy

	if err = s.IsValid(); err != nil {
		return nil, err
	}

	return c.app.CreateSkill(ctx, s)
}

func (c *Controller) SearchSkill(ctx context.Context, session *auth_manager.Session, search *model.SearchSkill) ([]*model.Skill, bool, model.AppError) {
	permission := session.GetPermission(model.PermissionSkill)
	if !permission.CanRead() {
		return nil, false, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if session.UseRBAC(auth_manager.PERMISSION_ACCESS_READ, permission) {
		return c.app.GetSkillsPageByGroups(ctx, session.Domain(search.DomainId), session.GetAclRoles(), search)
	} else {
		return c.app.GetSkillsPage(ctx, session.Domain(search.DomainId), search)
	}
}

func (c *Controller) ReadSkill(ctx context.Context, session *auth_manager.Session, id int64) (*model.Skill, model.AppError) {
	permission := session.GetPermission(model.PermissionSkill)
	if !permission.CanRead() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if session.UseRBAC(auth_manager.PERMISSION_ACCESS_READ, permission) {
		if perm, err := c.app.SkillCheckAccess(ctx, session.Domain(0), id, session.GetAclRoles(), auth_manager.PERMISSION_ACCESS_READ); err != nil {
			return nil, err
		} else if !perm {
			return nil, c.app.MakeResourcePermissionError(session, id, permission, auth_manager.PERMISSION_ACCESS_READ)
		}
	}

	return c.app.GetSkill(ctx, id, session.Domain(0))
}

func (c *Controller) UpdateSkill(ctx context.Context, session *auth_manager.Session, s *model.Skill) (*model.Skill, model.AppError) {
	permission := session.GetPermission(model.PermissionSkill)
	if !permission.CanUpdate() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
	}

	if session.UseRBAC(auth_manager.PERMISSION_ACCESS_UPDATE, permission) {
		if perm, err := c.app.SkillCheckAccess(ctx, session.Domain(0), s.Id, session.GetAclRoles(), auth_manager.PERMISSION_ACCESS_UPDATE); err != nil {
			return nil, err
		} else if !perm {
			return nil, c.app.MakeResourcePermissionError(session, s.Id, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
		}
	}

	t := time.Now()
	s.UpdatedAt = &t
	s.UpdatedBy = &model.Lookup{
		Id: int(session.UserId),
	}

	if err := s.IsValid(); err != nil {
		return nil, err
	}

	return c.app.UpdateSkill(ctx, s)
}

func (c *Controller) DeleteSkill(ctx context.Context, session *auth_manager.Session, id int64) (*model.Skill, model.AppError) {
	permission := session.GetPermission(model.PermissionSkill)
	if !permission.CanDelete() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_DELETE)
	}

	if session.UseRBAC(auth_manager.PERMISSION_ACCESS_DELETE, permission) {
		if perm, err := c.app.SkillCheckAccess(ctx, session.Domain(0), id, session.GetAclRoles(), auth_manager.PERMISSION_ACCESS_DELETE); err != nil {
			return nil, err
		} else if !perm {
			return nil, c.app.MakeResourcePermissionError(session, id, permission, auth_manager.PERMISSION_ACCESS_DELETE)
		}
	}

	return c.app.RemoveSkill(ctx, session.Domain(0), id)
}
