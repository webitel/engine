package controller

import (
	"context"
	"github.com/webitel/engine/auth_manager"
	"github.com/webitel/engine/model"
)

func (c *Controller) SearchTeamTrigger(ctx context.Context, session *auth_manager.Session, teamId int64, search *model.SearchTeamTrigger) ([]*model.TeamTrigger, bool, model.AppError) {
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

	return c.app.SearchTeamTrigger(ctx, session.Domain(0), teamId, search)
}

func (c *Controller) SearchAgentTrigger(ctx context.Context, session *auth_manager.Session, search *model.SearchTeamTrigger) ([]*model.TeamTrigger, bool, model.AppError) {
	//var err model.AppError
	userId := session.GetUserId()
	/* TODO
	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_AGENT)
	if !permission.CanRead() {
		return nil, false, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if session.UseRBAC(auth_manager.PERMISSION_ACCESS_READ, permission) {
		var perm bool
		if perm, err = c.app.AgentCheckAccess(ctx, session.Domain(0), userId, session.GetAclRoles(), auth_manager.PERMISSION_ACCESS_READ); err != nil {
			return nil, false, err
		} else if !perm {
			return nil, false, c.app.MakeResourcePermissionError(session, userId, permission, auth_manager.PERMISSION_ACCESS_READ)
		}
	}
	*/

	return c.app.SearchAgentTrigger(ctx, session.Domain(0), userId, search)
}

func (c *Controller) CreateTeamTrigger(ctx context.Context, session *auth_manager.Session, teamId int64, trigger *model.TeamTrigger) (*model.TeamTrigger, model.AppError) {
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
	trigger.CreatedBy = &model.Lookup{
		Id: int(session.UserId),
	}
	trigger.UpdatedBy = trigger.CreatedBy

	trigger.CreatedAt = model.GetTime()
	trigger.UpdatedAt = trigger.CreatedAt

	if err := trigger.IsValid(); err != nil {
		return nil, err
	}

	trigger, err = c.app.CreateTeamTrigger(ctx, session.Domain(0), teamId, trigger)
	if err != nil {
		return nil, err
	}

	return trigger, nil
}

func (c *Controller) GetTeamTrigger(ctx context.Context, session *auth_manager.Session, teamId int64, id uint32) (*model.TeamTrigger, model.AppError) {
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

	return c.app.GetTeamTrigger(ctx, session.Domain(0), teamId, id)
}

func (c *Controller) UpdateTeamTrigger(ctx context.Context, session *auth_manager.Session, teamId int64, trigger *model.TeamTrigger) (*model.TeamTrigger, model.AppError) {
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

	trigger.UpdatedBy = &model.Lookup{
		Id: int(session.UserId),
	}
	trigger.UpdatedAt = model.GetTime()

	if err := trigger.IsValid(); err != nil {
		return nil, err
	}

	trigger, err = c.app.UpdateTeamTrigger(ctx, session.DomainId, teamId, trigger)
	if err != nil {
		return nil, err
	}

	return trigger, nil
}

func (c *Controller) PatchTeamTrigger(ctx context.Context, session *auth_manager.Session, teamId int64, id uint32, patch *model.TeamTriggerPatch) (*model.TeamTrigger, model.AppError) {
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

	var trigger *model.TeamTrigger
	trigger, err = c.app.PatchTeamTrigger(ctx, session.DomainId, teamId, id, patch)
	if err != nil {
		return nil, err
	}

	return trigger, nil
}

func (c *Controller) DeleteTeamTrigger(ctx context.Context, session *auth_manager.Session, teamId int64, id uint32) (*model.TeamTrigger, model.AppError) {
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

	var trigger *model.TeamTrigger

	trigger, err = c.app.RemoveTeamTrigger(ctx, session.DomainId, teamId, id)
	if err != nil {
		return nil, err
	}

	return trigger, nil
}
