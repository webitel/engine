package controller

import (
	"context"

	"github.com/webitel/engine/auth_manager"
	"github.com/webitel/engine/model"
)

func (c *Controller) DirectAgentToMember(ctx context.Context, session *auth_manager.Session, domainId, memberId int64, communicationId int, agentId int64) (int64, model.AppError) {
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

	return c.app.DirectAgentToMember(session.Domain(domainId), memberId, communicationId, agentId)
}

func (c *Controller) ListOfflineQueueForAgent(ctx context.Context, session *auth_manager.Session, search *model.SearchOfflineQueueMembers) ([]*model.OfflineMember, bool, model.AppError) {
	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_QUEUE)
	if !permission.CanRead() {
		return nil, false, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	return c.app.ListOfflineQueueForAgent(ctx, session.DomainId, search)
}

func (c *Controller) ReportingAttempt(session *auth_manager.Session, attemptId int64, status, description string, nextOffering *int64,
	expireAt *int64, vars map[string]string, stickyDisplay bool, agentId int32, exclDes bool, waitBetweenRetries *int32, onlyComm bool) model.AppError {
	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_QUEUE)
	if !permission.CanRead() {
		return c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	return c.app.ReportingAttempt(attemptId, status, description, nextOffering, expireAt, vars, stickyDisplay, agentId, exclDes,
		waitBetweenRetries, onlyComm)
}

func (c *Controller) RenewalAttempt(session *auth_manager.Session, attemptId int64, renewal uint32) model.AppError {
	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_QUEUE)
	if !permission.CanRead() {
		return c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	return c.app.RenewalAttempt(session.DomainId, attemptId, renewal)
}

func (c *Controller) ProcessingActionFormAttempt(session *auth_manager.Session, attemptId int64, appId string, formId string, action string, fields map[string]string) model.AppError {
	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_QUEUE)
	if !permission.CanRead() {
		return c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	return c.app.ProcessingActionForm(session.DomainId, attemptId, appId, formId, action, fields)
}

func (c *Controller) ProcessingSaveForm(session *auth_manager.Session, attemptId int64, fields map[string]string) model.AppError {
	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_QUEUE)
	if !permission.CanRead() {
		return c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	return c.app.ProcessingSaveForm(session.DomainId, attemptId, fields)
}

func (c *Controller) InterceptAttempt(session *auth_manager.Session, attemptId int64, agentId int32) model.AppError {
	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_QUEUE)
	if !permission.CanRead() {
		return c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	return c.app.InterceptAttempt(session.DomainId, attemptId, agentId)
}
