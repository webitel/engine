package controller

import (
	"context"

	"github.com/webitel/engine/model"
	"github.com/webitel/engine/pkg/wbt/auth_manager"
)

// emailCommandAllowed: every Email command of an operator needs email_thread update.
func emailCommandAllowed(session *auth_manager.Session) bool {
	return session.GetPermission(model.PERMISSION_SCOPE_EMAIL_THREAD).CanUpdate()
}

// emailTargetQueue is the queue a transfer moves the email to; it needs cc_queue read.
func emailTargetQueue(cmd *model.EmailAttemptCommand) int32 {
	if cmd.Op == model.EmailAttemptTransfer {
		return cmd.ToQueueId
	}

	return 0
}

// EmailAttemptCommand checks the session, its rights and its own agent; call_center checks the owner,
// the version and a repeat of the command.
func (c *Controller) EmailAttemptCommand(ctx context.Context, session *auth_manager.Session, cmd *model.EmailAttemptCommand) (*model.EmailAttemptResult, model.AppError) {
	if !emailCommandAllowed(session) {
		return nil, c.app.MakePermissionError(session, session.GetPermission(model.PERMISSION_SCOPE_EMAIL_THREAD), auth_manager.PERMISSION_ACCESS_UPDATE)
	}
	if queueId := emailTargetQueue(cmd); queueId > 0 {
		permission := session.GetPermission(model.PERMISSION_SCOPE_CC_QUEUE)
		if !permission.CanRead() {
			return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
		}
		if session.UseRBAC(auth_manager.PERMISSION_ACCESS_READ, permission) {
			allowed, err := c.app.QueueCheckAccess(ctx, session.DomainId, int64(queueId), session.GetAclRoles(), auth_manager.PERMISSION_ACCESS_READ)
			if err != nil {
				return nil, err
			}
			if !allowed {
				return nil, c.app.MakeResourcePermissionError(session, int64(queueId), permission, auth_manager.PERMISSION_ACCESS_READ)
			}
		}
	}

	agentId, err := c.app.EmailOwnAgent(ctx, session.DomainId, session.UserId)
	if err != nil {
		return nil, err
	}

	return c.app.EmailAttemptCommand(ctx, session.DomainId, session.UserId, agentId, cmd)
}

// EmailTransferCandidates lists the agents the current owner may transfer the email to.
func (c *Controller) EmailTransferCandidates(ctx context.Context, session *auth_manager.Session, attemptId int64) ([]*model.EmailTransferCandidate, model.AppError) {
	if !emailCommandAllowed(session) {
		return nil, c.app.MakePermissionError(session, session.GetPermission(model.PERMISSION_SCOPE_EMAIL_THREAD), auth_manager.PERMISSION_ACCESS_UPDATE)
	}
	agentId, err := c.app.EmailOwnAgent(ctx, session.DomainId, session.UserId)
	if err != nil {
		return nil, err
	}

	return c.app.EmailTransferCandidates(ctx, session.DomainId, attemptId, agentId)
}
