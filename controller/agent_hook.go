package controller

import (
	"context"

	"github.com/webitel/engine/auth_manager"
	"github.com/webitel/engine/model"
)

func (c *Controller) SearchAgentHook(ctx context.Context, session *auth_manager.Session, agentId int64, search *model.SearchAgentHook) ([]*model.AgentHook, bool, model.AppError) {
	var err model.AppError
	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_AGENT)
	if !permission.CanRead() {
		return nil, false, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if session.UseRBAC(auth_manager.PERMISSION_ACCESS_READ, permission) {
		var perm bool
		if perm, err = c.app.AgentCheckAccess(ctx, session.Domain(0), int64(agentId), session.GetAclRoles(), auth_manager.PERMISSION_ACCESS_READ); err != nil {
			return nil, false, err
		} else if !perm {
			return nil, false, c.app.MakeResourcePermissionError(session, int64(agentId), permission, auth_manager.PERMISSION_ACCESS_READ)
		}
	}

	return c.app.SearchAgentHook(ctx, session.Domain(0), agentId, search)
}

func (c *Controller) CreateAgentHook(ctx context.Context, session *auth_manager.Session, agentId int64, hook *model.AgentHook) (*model.AgentHook, model.AppError) {
	var err model.AppError
	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_AGENT)
	if !permission.CanRead() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if !permission.CanUpdate() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
	}

	if session.UseRBAC(auth_manager.PERMISSION_ACCESS_UPDATE, permission) {
		var perm bool
		if perm, err = c.app.AgentCheckAccess(ctx, session.Domain(0), int64(agentId), session.GetAclRoles(), auth_manager.PERMISSION_ACCESS_UPDATE); err != nil {
			return nil, err
		} else if !perm {
			return nil, c.app.MakeResourcePermissionError(session, int64(agentId), permission, auth_manager.PERMISSION_ACCESS_UPDATE)
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

	hook, err = c.app.CreateAgentHook(ctx, session.Domain(0), agentId, hook)
	if err != nil {
		return nil, err
	}

	return hook, nil
}

func (c *Controller) GetAgentHook(ctx context.Context, session *auth_manager.Session, agentId int64, id int32) (*model.AgentHook, model.AppError) {
	var err model.AppError
	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_AGENT)
	if !permission.CanRead() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if session.UseRBAC(auth_manager.PERMISSION_ACCESS_READ, permission) {
		var perm bool
		if perm, err = c.app.AgentCheckAccess(ctx, session.Domain(0), int64(agentId), session.GetAclRoles(), auth_manager.PERMISSION_ACCESS_READ); err != nil {
			return nil, err
		} else if !perm {
			return nil, c.app.MakeResourcePermissionError(session, int64(agentId), permission, auth_manager.PERMISSION_ACCESS_READ)
		}
	}

	return c.app.GetAgentHook(ctx, session.Domain(0), agentId, id)
}

func (c *Controller) UpdateAgentHook(ctx context.Context, session *auth_manager.Session, agentId int64, hook *model.AgentHook) (*model.AgentHook, model.AppError) {
	var err model.AppError
	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_AGENT)
	if !permission.CanRead() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if !permission.CanUpdate() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
	}

	if session.UseRBAC(auth_manager.PERMISSION_ACCESS_UPDATE, permission) {
		var perm bool
		if perm, err = c.app.AgentCheckAccess(ctx, session.Domain(0), int64(agentId), session.GetAclRoles(), auth_manager.PERMISSION_ACCESS_UPDATE); err != nil {
			return nil, err
		} else if !perm {
			return nil, c.app.MakeResourcePermissionError(session, int64(agentId), permission, auth_manager.PERMISSION_ACCESS_UPDATE)
		}
	}

	hook.UpdatedBy = &model.Lookup{
		Id: int(session.UserId),
	}
	hook.UpdatedAt = *model.GetTime()

	if err := hook.IsValid(); err != nil {
		return nil, err
	}

	hook, err = c.app.UpdateAgentHook(ctx, session.DomainId, agentId, hook)
	if err != nil {
		return nil, err
	}

	return hook, nil
}

func (c *Controller) PatchAgentHook(ctx context.Context, session *auth_manager.Session, agentId int64, id int32, patch *model.AgentHookPatch) (*model.AgentHook, model.AppError) {
	var err model.AppError
	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_AGENT)
	if !permission.CanRead() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if !permission.CanUpdate() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
	}

	if session.UseRBAC(auth_manager.PERMISSION_ACCESS_UPDATE, permission) {
		var perm bool
		if perm, err = c.app.AgentCheckAccess(ctx, session.Domain(0), int64(agentId), session.GetAclRoles(), auth_manager.PERMISSION_ACCESS_UPDATE); err != nil {
			return nil, err
		} else if !perm {
			return nil, c.app.MakeResourcePermissionError(session, int64(agentId), permission, auth_manager.PERMISSION_ACCESS_UPDATE)
		}
	}
	patch.UpdatedBy.Id = int(session.UserId)
	patch.UpdatedAt = *model.GetTime()

	var hook *model.AgentHook
	hook, err = c.app.PatchAgentHook(ctx, session.DomainId, agentId, id, patch)
	if err != nil {
		return nil, err
	}

	return hook, nil
}

func (c *Controller) DeleteAgentHook(ctx context.Context, session *auth_manager.Session, queueId int64, id int32) (*model.AgentHook, model.AppError) {
	var err model.AppError
	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_AGENT)
	if !permission.CanRead() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if !permission.CanUpdate() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
	}

	if session.UseRBAC(auth_manager.PERMISSION_ACCESS_UPDATE, permission) {
		var perm bool
		if perm, err = c.app.AgentCheckAccess(ctx, session.Domain(0), int64(queueId), session.GetAclRoles(), auth_manager.PERMISSION_ACCESS_UPDATE); err != nil {
			return nil, err
		} else if !perm {
			return nil, c.app.MakeResourcePermissionError(session, int64(queueId), permission, auth_manager.PERMISSION_ACCESS_UPDATE)
		}
	}

	var hook *model.AgentHook

	hook, err = c.app.RemoveAgentHook(ctx, session.DomainId, queueId, id)
	if err != nil {
		return nil, err
	}

	return hook, nil
}
