package controller

import (
	"github.com/webitel/engine/auth_manager"
	"github.com/webitel/engine/model"
)

func (c *Controller) DirectAgentToMember(session *auth_manager.Session, domainId, memberId int64, communicationId int, agentId int64) (int64, *model.AppError) {
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

	return c.app.DirectAgentToMember(session.Domain(domainId), memberId, communicationId, agentId)
}

func (c *Controller) ListOfflineQueueForAgent(session *auth_manager.Session, search *model.SearchOfflineQueueMembers) ([]*model.OfflineMember, bool, *model.AppError) {
	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_QUEUE)
	if !permission.CanRead() {
		return nil, false, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	return c.app.ListOfflineQueueForAgent(session.DomainId, search)
}
