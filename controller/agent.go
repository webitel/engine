package controller

import (
	"context"

	"github.com/webitel/engine/auth_manager"
	"github.com/webitel/engine/model"
)

func (c *Controller) GetAgentSession(ctx context.Context, session *auth_manager.Session, domainId, userId int64) (*model.AgentSession, model.AppError) {

	v, err := c.app.AgentCC(ctx, session.Domain(domainId), userId)
	if err != nil {
		return nil, err
	}

	if err = v.Valid(); err != nil {
		return nil, err
	}

	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_AGENT)

	if !session.HasCallCenterLicense() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if !permission.CanRead() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if session.UseRBAC(auth_manager.PERMISSION_ACCESS_READ, permission) {
		if perm, err := c.app.AgentCheckAccess(ctx, session.Domain(domainId), *v.AgentId, session.GetAclRoles(), auth_manager.PERMISSION_ACCESS_READ); err != nil {
			return nil, err
		} else if !perm {
			return nil, c.app.MakeResourcePermissionError(session, userId, permission, auth_manager.PERMISSION_ACCESS_READ)
		}
	}

	return c.app.GetAgentSession(ctx, session.Domain(domainId), userId)
}

func (c *Controller) LoginAgent(ctx context.Context, session *auth_manager.Session, domainId, agentId int64, onDemand bool) model.AppError {
	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_AGENT)
	if !permission.CanRead() {
		return c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if session.UseRBAC(auth_manager.PERMISSION_ACCESS_READ, permission) {
		if perm, err := c.app.AgentCheckAccess(ctx, session.Domain(domainId), agentId, session.GetAclRoles(), auth_manager.PERMISSION_ACCESS_READ); err != nil {
			return err
		} else if !perm {
			return c.app.MakeResourcePermissionError(session, agentId, permission, auth_manager.PERMISSION_ACCESS_READ)
		}
	}

	return c.app.LoginAgent(session.Domain(domainId), agentId, onDemand)
}

func (c *Controller) LogoutAgent(ctx context.Context, session *auth_manager.Session, domainId, agentId int64) model.AppError {
	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_AGENT)
	if !permission.CanRead() {
		return c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if session.UseRBAC(auth_manager.PERMISSION_ACCESS_READ, permission) {
		if perm, err := c.app.AgentCheckAccess(ctx, session.Domain(domainId), agentId, session.GetAclRoles(), auth_manager.PERMISSION_ACCESS_READ); err != nil {
			return err
		} else if !perm {
			return c.app.MakeResourcePermissionError(session, agentId, permission, auth_manager.PERMISSION_ACCESS_READ)
		}
	}

	return c.app.LogoutAgent(session.Domain(domainId), agentId)
}

func (c *Controller) PauseAgent(ctx context.Context, session *auth_manager.Session, domainId, agentId int64, payload string, timeout int) model.AppError {
	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_AGENT)
	if !permission.CanRead() {
		return c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if session.UseRBAC(auth_manager.PERMISSION_ACCESS_READ, permission) {
		if perm, err := c.app.AgentCheckAccess(ctx, session.Domain(domainId), agentId, session.GetAclRoles(), auth_manager.PERMISSION_ACCESS_READ); err != nil {
			return err
		} else if !perm {
			return c.app.MakeResourcePermissionError(session, agentId, permission, auth_manager.PERMISSION_ACCESS_READ)
		}
	}

	return c.app.PauseAgent(session.Domain(domainId), agentId, payload, timeout)
}

func (c *Controller) WaitingAgent(ctx context.Context, session *auth_manager.Session, domainId, agentId int64, channel string) (int64, model.AppError) {
	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_AGENT)
	if !permission.CanRead() {
		return 0, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if session.UseRBAC(auth_manager.PERMISSION_ACCESS_READ, permission) {
		if perm, err := c.app.AgentCheckAccess(ctx, session.Domain(domainId), agentId, session.GetAclRoles(), auth_manager.PERMISSION_ACCESS_READ); err != nil {
			return 0, err
		} else if !perm {
			return 0, c.app.MakeResourcePermissionError(session, agentId, permission, auth_manager.PERMISSION_ACCESS_READ)
		}
	}

	return c.app.WaitingAgentChannel(session.Domain(domainId), agentId, channel)
}

func (c *Controller) ActiveAgentTasks(ctx context.Context, session *auth_manager.Session, domainId, agentId int64) ([]*model.CCTask, model.AppError) {
	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_AGENT)
	if !permission.CanRead() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if session.UseRBAC(auth_manager.PERMISSION_ACCESS_READ, permission) {
		if perm, err := c.app.AgentCheckAccess(ctx, session.Domain(domainId), agentId, session.GetAclRoles(), auth_manager.PERMISSION_ACCESS_READ); err != nil {
			return nil, err
		} else if !perm {
			return nil, c.app.MakeResourcePermissionError(session, agentId, permission, auth_manager.PERMISSION_ACCESS_READ)
		}
	}

	return c.app.GetAgentActiveTasks(ctx, session.Domain(domainId), agentId)
}

func (c *Controller) GetAgentInQueueStatistics(ctx context.Context, session *auth_manager.Session, domainId, agentId int64) ([]*model.AgentInQueueStatistic, model.AppError) {
	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_AGENT)
	if !permission.CanRead() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if session.UseRBAC(auth_manager.PERMISSION_ACCESS_READ, permission) {
		if perm, err := c.app.AgentCheckAccess(ctx, session.Domain(domainId), agentId, session.GetAclRoles(), auth_manager.PERMISSION_ACCESS_READ); err != nil {
			return nil, err
		} else if !perm {
			return nil, c.app.MakeResourcePermissionError(session, agentId, permission, auth_manager.PERMISSION_ACCESS_READ)
		}
	}

	return c.app.GetAgentInQueueStatistics(ctx, session.Domain(domainId), agentId)
}

func (c *Controller) AcceptAgentTask(session *auth_manager.Session, appId string, attemptId int64) model.AppError {
	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_AGENT)
	if !permission.CanRead() {
		return c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	return c.app.AcceptTask(appId, session.DomainId, attemptId)
}

func (c *Controller) CloseAgentTask(session *auth_manager.Session, appId string, attemptId int64) model.AppError {
	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_AGENT)
	if !permission.CanRead() {
		return c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	return c.app.CloseTask(appId, session.DomainId, attemptId)
}

func (c *Controller) GetAgentPauseCause(ctx context.Context, session *auth_manager.Session, toAgentId int64, allowChange bool) ([]*model.AgentPauseCause, model.AppError) {
	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_AGENT)
	if !permission.CanRead() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if !permission.CanUpdate() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
	}
	// todo RBAC ?

	return c.app.GetAgentPauseCause(ctx, session.Domain(0), session.UserId, toAgentId, allowChange)
}

func (c *Controller) GetSupervisorAgentItem(ctx context.Context, session *auth_manager.Session, agentId int64, t *model.FilterBetween) (*model.SupervisorAgentItem, model.AppError) {
	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_AGENT)
	if !permission.CanRead() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if !permission.CanUpdate() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
	}
	// todo RBAC ?

	return c.app.SupervisorAgentItem(ctx, session.DomainId, agentId, t)
}

func (c *Controller) GetAgentTodayStatistics(ctx context.Context, session *auth_manager.Session, agentId int64) (*model.AgentStatistics, model.AppError) {
	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_AGENT)
	if !permission.CanRead() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if session.UseRBAC(auth_manager.PERMISSION_ACCESS_READ, permission) {
		if perm, err := c.app.AgentCheckAccess(ctx, session.Domain(0), agentId, session.GetAclRoles(), auth_manager.PERMISSION_ACCESS_READ); err != nil {
			return nil, err
		} else if !perm {
			return nil, c.app.MakeResourcePermissionError(session, agentId, permission, auth_manager.PERMISSION_ACCESS_READ)
		}
	}

	return c.app.GetAgentTodayStatistics(ctx, session.Domain(0), agentId)
}

func (c *Controller) GetUserTodayStatistics(ctx context.Context, session *auth_manager.Session) (*model.AgentStatistics, model.AppError) {
	return c.app.GetUserTodayStatistics(ctx, session.Domain(0), session.UserId)
}

func (c *Controller) SearchUserStatus(ctx context.Context, session *auth_manager.Session, search *model.SearchUserStatus) ([]*model.UserStatus, bool, model.AppError) {
	permission := session.GetPermission(model.PERMISSION_SCOPE_USERS)
	if !permission.CanRead() {
		return nil, false, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if session.UseRBAC(auth_manager.PERMISSION_ACCESS_READ, permission) {
		return c.app.GetUsersStatusPageByGroups(ctx, session.Domain(0), session.GetAclRoles(), search)
	} else {
		return c.app.GetUsersStatusPage(ctx, session.Domain(0), search)
	}
}

func (c *Controller) RunAgentTrigger(ctx context.Context, session *auth_manager.Session, triggerId int32, vars map[string]string) (string, model.AppError) {
	if !session.HasCallCenterLicense() {
		return "", model.NewForbiddenError("app.license.cc", "Not found \"CALL_CENTER\" license")
	}
	return c.app.RunAgentTrigger(ctx, session.Domain(0), session.UserId, triggerId, vars)
}
