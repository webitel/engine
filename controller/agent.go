package controller

import (
	"github.com/webitel/engine/auth_manager"
	"github.com/webitel/engine/model"
)

func (c *Controller) GetAgentSession(session *auth_manager.Session, domainId, userId int64) (*model.AgentSession, *model.AppError) {
	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_AGENT)
	if !permission.CanRead() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if permission.Rbac {
		//FIXME agentID!
		if perm, err := c.app.AgentCheckAccess(session.Domain(domainId), userId, session.GetAclRoles(), auth_manager.PERMISSION_ACCESS_READ); err != nil {
			return nil, err
		} else if !perm {
			return nil, c.app.MakeResourcePermissionError(session, userId, permission, auth_manager.PERMISSION_ACCESS_READ)
		}
	}

	return c.app.GetAgentSession(session.Domain(domainId), userId)
}

func (c *Controller) LoginAgent(session *auth_manager.Session, domainId, agentId int64, channels []string, onDemand bool) *model.AppError {
	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_AGENT)
	if !permission.CanRead() {
		return c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if permission.Rbac {
		if perm, err := c.app.AgentCheckAccess(session.Domain(domainId), agentId, session.GetAclRoles(), auth_manager.PERMISSION_ACCESS_READ); err != nil {
			return err
		} else if !perm {
			return c.app.MakeResourcePermissionError(session, agentId, permission, auth_manager.PERMISSION_ACCESS_READ)
		}
	}

	return c.app.LoginAgent(session.Domain(domainId), agentId, channels, onDemand)
}

func (c *Controller) LogoutAgent(session *auth_manager.Session, domainId, agentId int64) *model.AppError {
	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_AGENT)
	if !permission.CanRead() {
		return c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if permission.Rbac {
		if perm, err := c.app.AgentCheckAccess(session.Domain(domainId), agentId, session.GetAclRoles(), auth_manager.PERMISSION_ACCESS_READ); err != nil {
			return err
		} else if !perm {
			return c.app.MakeResourcePermissionError(session, agentId, permission, auth_manager.PERMISSION_ACCESS_READ)
		}
	}

	return c.app.LogoutAgent(session.Domain(domainId), agentId)
}

func (c *Controller) PauseAgent(session *auth_manager.Session, domainId, agentId int64, payload string, timeout int) *model.AppError {
	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_AGENT)
	if !permission.CanRead() {
		return c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if permission.Rbac {
		if perm, err := c.app.AgentCheckAccess(session.Domain(domainId), agentId, session.GetAclRoles(), auth_manager.PERMISSION_ACCESS_READ); err != nil {
			return err
		} else if !perm {
			return c.app.MakeResourcePermissionError(session, agentId, permission, auth_manager.PERMISSION_ACCESS_READ)
		}
	}

	return c.app.PauseAgent(session.Domain(domainId), agentId, payload, timeout)
}

func (c *Controller) WaitingAgent(session *auth_manager.Session, domainId, agentId int64, channel string) (int64, *model.AppError) {
	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_AGENT)
	if !permission.CanRead() {
		return 0, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if permission.Rbac {
		if perm, err := c.app.AgentCheckAccess(session.Domain(domainId), agentId, session.GetAclRoles(), auth_manager.PERMISSION_ACCESS_READ); err != nil {
			return 0, err
		} else if !perm {
			return 0, c.app.MakeResourcePermissionError(session, agentId, permission, auth_manager.PERMISSION_ACCESS_READ)
		}
	}

	return c.app.WaitingAgentChannel(session.Domain(domainId), agentId, channel)
}

func (c *Controller) GetAgentInQueueStatistics(session *auth_manager.Session, domainId, agentId int64) ([]*model.AgentInQueueStatistic, *model.AppError) {
	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_AGENT)
	if !permission.CanRead() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if permission.Rbac {
		if perm, err := c.app.AgentCheckAccess(session.Domain(domainId), agentId, session.GetAclRoles(), auth_manager.PERMISSION_ACCESS_READ); err != nil {
			return nil, err
		} else if !perm {
			return nil, c.app.MakeResourcePermissionError(session, agentId, permission, auth_manager.PERMISSION_ACCESS_READ)
		}
	}

	return c.app.GetAgentInQueueStatistics(session.Domain(domainId), agentId)
}
