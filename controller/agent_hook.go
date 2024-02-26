package controller

import (
	"context"

	"github.com/webitel/engine/auth_manager"
	"github.com/webitel/engine/model"
)

func (c *Controller) SearchTeamHook(ctx context.Context, session *auth_manager.Session, teamId int64, search *model.SearchTeamHook) ([]*model.TeamHook, bool, model.AppError) {
	var err model.AppError
	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_TEAM)
	if !permission.CanRead() {
		return nil, false, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if session.UseRBAC(auth_manager.PERMISSION_ACCESS_READ, permission) {
		var perm bool
		if perm, err = c.app.AgentTeamCheckAccess(ctx, session.Domain(0), teamId, session.GetAclRoles(), auth_manager.PERMISSION_ACCESS_READ); err != nil {
			return nil, false, err
		} else if !perm {
			return nil, false, c.app.MakeResourcePermissionError(session, teamId, permission, auth_manager.PERMISSION_ACCESS_READ)
		}
	}

	return c.app.SearchTeamHook(ctx, session.Domain(0), teamId, search)
}

func (c *Controller) CreateTeamHook(ctx context.Context, session *auth_manager.Session, teamId int64, hook *model.TeamHook) (*model.TeamHook, model.AppError) {
	var err model.AppError
	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_TEAM)
	if !permission.CanRead() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if !permission.CanUpdate() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
	}

	if session.UseRBAC(auth_manager.PERMISSION_ACCESS_UPDATE, permission) {
		var perm bool
		if perm, err = c.app.AgentTeamCheckAccess(ctx, session.Domain(0), teamId, session.GetAclRoles(), auth_manager.PERMISSION_ACCESS_UPDATE); err != nil {
			return nil, err
		} else if !perm {
			return nil, c.app.MakeResourcePermissionError(session, teamId, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
		}
	}
	hook.CreatedBy = &model.Lookup{
		Id: int(session.UserId),
	}
	hook.UpdatedBy = hook.CreatedBy

	hook.CreatedAt = *model.GetTime()
	hook.UpdatedAt = hook.CreatedAt

	if err := hook.IsValid(); err != nil {
		return nil, err
	}

	hook, err = c.app.CreateTeamHook(ctx, session.Domain(0), teamId, hook)
	if err != nil {
		return nil, err
	}

	return hook, nil
}

func (c *Controller) GetTeamHook(ctx context.Context, session *auth_manager.Session, teamId int64, id uint32) (*model.TeamHook, model.AppError) {
	var err model.AppError
	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_TEAM)
	if !permission.CanRead() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if session.UseRBAC(auth_manager.PERMISSION_ACCESS_READ, permission) {
		var perm bool
		if perm, err = c.app.AgentTeamCheckAccess(ctx, session.Domain(0), teamId, session.GetAclRoles(), auth_manager.PERMISSION_ACCESS_READ); err != nil {
			return nil, err
		} else if !perm {
			return nil, c.app.MakeResourcePermissionError(session, teamId, permission, auth_manager.PERMISSION_ACCESS_READ)
		}
	}

	return c.app.GetTeamHook(ctx, session.Domain(0), teamId, id)
}

func (c *Controller) UpdateTeamHook(ctx context.Context, session *auth_manager.Session, teamId int64, hook *model.TeamHook) (*model.TeamHook, model.AppError) {
	var err model.AppError
	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_TEAM)
	if !permission.CanRead() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if !permission.CanUpdate() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
	}

	if session.UseRBAC(auth_manager.PERMISSION_ACCESS_UPDATE, permission) {
		var perm bool
		if perm, err = c.app.AgentTeamCheckAccess(ctx, session.Domain(0), teamId, session.GetAclRoles(), auth_manager.PERMISSION_ACCESS_UPDATE); err != nil {
			return nil, err
		} else if !perm {
			return nil, c.app.MakeResourcePermissionError(session, teamId, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
		}
	}

	hook.UpdatedBy = &model.Lookup{
		Id: int(session.UserId),
	}
	hook.UpdatedAt = *model.GetTime()

	if err := hook.IsValid(); err != nil {
		return nil, err
	}

	hook, err = c.app.UpdateTeamHook(ctx, session.DomainId, teamId, hook)
	if err != nil {
		return nil, err
	}

	return hook, nil
}

func (c *Controller) PatchTeamHook(ctx context.Context, session *auth_manager.Session, teamId int64, id uint32, patch *model.TeamHookPatch) (*model.TeamHook, model.AppError) {
	var err model.AppError
	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_TEAM)
	if !permission.CanRead() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if !permission.CanUpdate() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
	}

	if session.UseRBAC(auth_manager.PERMISSION_ACCESS_UPDATE, permission) {
		var perm bool
		if perm, err = c.app.AgentTeamCheckAccess(ctx, session.Domain(0), teamId, session.GetAclRoles(), auth_manager.PERMISSION_ACCESS_UPDATE); err != nil {
			return nil, err
		} else if !perm {
			return nil, c.app.MakeResourcePermissionError(session, teamId, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
		}
	}
	patch.UpdatedBy.Id = int(session.UserId)
	patch.UpdatedAt = *model.GetTime()

	var hook *model.TeamHook
	hook, err = c.app.PatchTeamHook(ctx, session.DomainId, teamId, id, patch)
	if err != nil {
		return nil, err
	}

	return hook, nil
}

func (c *Controller) DeleteTeamHook(ctx context.Context, session *auth_manager.Session, teamId int64, id uint32) (*model.TeamHook, model.AppError) {
	var err model.AppError
	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_TEAM)
	if !permission.CanRead() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if !permission.CanUpdate() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
	}

	if session.UseRBAC(auth_manager.PERMISSION_ACCESS_UPDATE, permission) {
		var perm bool
		if perm, err = c.app.AgentTeamCheckAccess(ctx, session.Domain(0), teamId, session.GetAclRoles(), auth_manager.PERMISSION_ACCESS_UPDATE); err != nil {
			return nil, err
		} else if !perm {
			return nil, c.app.MakeResourcePermissionError(session, teamId, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
		}
	}

	var hook *model.TeamHook

	hook, err = c.app.RemoveTeamHook(ctx, session.DomainId, teamId, id)
	if err != nil {
		return nil, err
	}

	return hook, nil
}
