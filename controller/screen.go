package controller

import (
	"context"
	"github.com/webitel/engine/model"
	"github.com/webitel/engine/pkg/wbt/auth_manager"
)

func (c *Controller) Screenshot(ctx context.Context, session *auth_manager.Session, toUserId int64) model.AppError {
	if !session.HasAction(model.PermissionControlAgentScreen) {
		return c.app.MakeActionPermissionError(session, model.PermissionControlAgentScreen, auth_manager.PERMISSION_ACCESS_READ)
	}

	return c.app.Screenshot(ctx, session.Domain(0), toUserId)
}

func (c *Controller) RequestScreenShare(ctx context.Context, session *auth_manager.Session, fromUserId, toUserId int64, sockId string, sdp, id string) model.AppError {
	if !session.HasAction(model.PermissionControlAgentScreen) {
		return c.app.MakeActionPermissionError(session, model.PermissionControlAgentScreen, auth_manager.PERMISSION_ACCESS_READ)
	}

	return c.app.RequestScreenShare(ctx, session.Domain(0), fromUserId, toUserId, sockId, sdp, id)
}

func (c *Controller) AcceptScreenShare(_ context.Context, session *auth_manager.Session, toUserId int64, sockId string, sess, sdp string) model.AppError {

	return c.app.AcceptScreenShare(session.Domain(0), toUserId, sockId, sess, sdp)
}

func (c *Controller) ScreenShareRecordStart(ctx context.Context, session *auth_manager.Session, toUserId int64, rootSessionId string) model.AppError {
	if !session.HasAction(model.PermissionControlAgentScreen) {
		return c.app.MakeActionPermissionError(session, model.PermissionControlAgentScreen, auth_manager.PERMISSION_ACCESS_READ)
	}

	return c.app.ScreenShareRecordStart(ctx, session.Domain(0), session.UserId, toUserId, rootSessionId)
}

func (c *Controller) ScreenShareRecordStop(ctx context.Context, session *auth_manager.Session, toUserId int64, rootSessionId string) model.AppError {
	if !session.HasAction(model.PermissionControlAgentScreen) {
		return c.app.MakeActionPermissionError(session, model.PermissionControlAgentScreen, auth_manager.PERMISSION_ACCESS_READ)
	}

	return c.app.ScreenShareRecordStop(ctx, session.Domain(0), toUserId, rootSessionId)
}
