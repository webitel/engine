package controller

import (
	"context"

	"github.com/webitel/engine/model"
	"github.com/webitel/engine/pkg/wbt/auth_manager"
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

func (c *Controller) ReportingAttempt(
	session *auth_manager.Session,
	attemptId int64,
	status, description string,
	nextOffering *int64,
	expireAt *int64,
	vars map[string]string,
	stickyDisplay bool,
	agentId int32,
	exclDes bool,
	waitBetweenRetries *int32,
	onlyComm bool,
	draft bool,
) model.AppError {
	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_QUEUE)
	if !permission.CanRead() {
		return c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	return c.app.ReportingAttempt(
		attemptId,
		status,
		description,
		nextOffering,
		expireAt,
		vars,
		stickyDisplay,
		agentId,
		exclDes,
		waitBetweenRetries,
		onlyComm,
		draft,
	)
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

func (c *Controller) ProcessingActionComponentAttempt(session *auth_manager.Session, attemptId int64, appId string,
	formId, componentId string, action string, vars map[string]string, sync bool) model.AppError {
	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_QUEUE)
	if !permission.CanRead() {
		return c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	return c.app.ProcessingActionComponent(session.DomainId, attemptId, appId, formId, componentId, action, vars, sync)
}

func (c *Controller) ProcessingSaveForm(session *auth_manager.Session, attemptId int64, fields map[string]string, form []byte) model.AppError {
	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_QUEUE)
	if !permission.CanRead() {
		return c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	return c.app.ProcessingSaveForm(session.DomainId, attemptId, fields, form)
}

func (c *Controller) InterceptAttempt(session *auth_manager.Session, attemptId int64, agentId int32) model.AppError {
	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_QUEUE)
	if !permission.CanRead() {
		return c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	return c.app.InterceptAttempt(session.DomainId, attemptId, agentId)
}

// TODO(WTEL-9876): permission checks (self-assign vs. force-assign-to-anyone)
func (c *Controller) AssignAttempt(ctx context.Context, session *auth_manager.Session, attemptId int64, agentId *int64) model.AppError {
	var targetAgentId int32
	if agentId == nil {
		ownAgentId, err := c.app.GetOwnAgentId(ctx, session.DomainId, session.UserId)
		if err != nil {
			return err
		}
		targetAgentId = ownAgentId
	} else {
		targetAgentId = int32(*agentId)
	}

	online, err := c.app.IsAgentChatChannelOnline(ctx, targetAgentId)
	if err != nil {
		return err
	}
	if !online {
		return model.NewBadRequestError("controller.cc_member.assign_attempt.agent_offline", "agent's chat channel is not online")
	}

	return c.app.InterceptAttempt(session.DomainId, attemptId, targetAgentId)
}

func (c *Controller) ResetMembersCount(ctx context.Context, query *model.ResetMembersCountQuery) (int64, model.AppError) {
	session, err := c.app.GetSessionFromCtx(ctx)
	if err != nil {
		return 0, err
	}

	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_QUEUE)
	if !permission.CanRead() {
		return 0, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if session.UseRBAC(auth_manager.PERMISSION_ACCESS_READ, permission) {
		canRead, err := c.app.QueueCheckAccess(ctx, session.Domain(0), query.QueueId, session.GetAclRoles(), auth_manager.PERMISSION_ACCESS_READ)
		if err != nil {
			return 0, err
		}

		if !canRead {
			return 0, c.app.MakeResourcePermissionError(session, query.QueueId, permission, auth_manager.PERMISSION_ACCESS_READ)
		}
	}

	return c.app.ResetMembersCount(ctx, session.Domain(0), query)
}
