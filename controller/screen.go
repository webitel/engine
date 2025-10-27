package controller

import (
	"context"
	"github.com/webitel/engine/model"
	"github.com/webitel/engine/pkg/wbt/auth_manager"
)

func (c *Controller) Screenshot(ctx context.Context, session *auth_manager.Session, toUserId int64, fromSockId string, ackId string) model.AppError {
	if !session.HasAction(model.PermissionControlAgentScreen) {
		return c.app.MakeActionPermissionError(session, model.PermissionControlAgentScreen, auth_manager.PERMISSION_ACCESS_READ)
	}

	return c.app.Screenshot(ctx, session.Domain(0), toUserId, fromSockId, ackId)
}

func (c *Controller) ACK(ctx context.Context, session *auth_manager.Session, ackId, errText string) model.AppError {
	if !session.HasAction(model.PermissionControlAgentScreen) {
		return c.app.MakeActionPermissionError(session, model.PermissionControlAgentScreen, auth_manager.PERMISSION_ACCESS_READ)
	}

	return c.app.ACK(ctx, session.Domain(0), ackId, errText)
}

func (c *Controller) RequestScreenShare(ctx context.Context, session *auth_manager.Session, fromUserId, toUserId int64, sockId string, sdp, id string, ackId string) (string, model.AppError) {
	if !session.HasAction(model.PermissionControlAgentScreen) {
		return "", c.app.MakeActionPermissionError(session, model.PermissionControlAgentScreen, auth_manager.PERMISSION_ACCESS_READ)
	}

	return c.app.RequestScreenShare(ctx, session.Domain(0), fromUserId, toUserId, sockId, sdp, id, ackId)
}

func (c *Controller) AcceptScreenShare(_ context.Context, session *auth_manager.Session, toUserId int64, sockId string, sess, sdp string, fromSockId string, ackId string) model.AppError {

	return c.app.AcceptScreenShare(session.Domain(0), toUserId, sockId, sess, sdp, fromSockId, ackId)
}

func (c *Controller) ScreenShareRecordStart(ctx context.Context, session *auth_manager.Session, toUserId int64, rootSessionId string, ackId string) model.AppError {
	if !session.HasAction(model.PermissionControlAgentScreen) {
		return c.app.MakeActionPermissionError(session, model.PermissionControlAgentScreen, auth_manager.PERMISSION_ACCESS_READ)
	}

	return c.app.ScreenShareRecordStart(ctx, session.Domain(0), session.UserId, toUserId, rootSessionId, ackId)
}

func (c *Controller) ScreenShareRecordStop(ctx context.Context, session *auth_manager.Session, toUserId int64, rootSessionId string, ackId string) model.AppError {
	if !session.HasAction(model.PermissionControlAgentScreen) {
		return c.app.MakeActionPermissionError(session, model.PermissionControlAgentScreen, auth_manager.PERMISSION_ACCESS_READ)
	}

	return c.app.ScreenShareRecordStop(ctx, session.Domain(0), toUserId, rootSessionId, ackId)
}
