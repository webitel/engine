package controller

import (
	"github.com/webitel/engine/auth_manager"
	"github.com/webitel/engine/model"
)

func (c *Controller) DeclineChat(session *auth_manager.Session, inviteId string) *model.AppError {
	// FIXME PERMISSION
	return c.app.DeclineChat(session.UserId, inviteId)
}

func (c *Controller) JoinChat(session *auth_manager.Session, inviteId string) (string, *model.AppError) {
	// FIXME PERMISSION
	return c.app.JoinChat(inviteId)
}

func (c *Controller) LeaveChat(session *auth_manager.Session, channelId, conversationId string) *model.AppError {
	// FIXME PERMISSION
	return c.app.LeaveChat(channelId, conversationId)
}

func (c *Controller) CloseChat(session *auth_manager.Session, channelId, conversationId, cause string) *model.AppError {
	// FIXME PERMISSION
	return c.app.CloseChat(channelId, conversationId, cause)
}

func (c *Controller) SendTextChat(session *auth_manager.Session, channelId, conversationId, text string) *model.AppError {
	// FIXME PERMISSION
	return c.app.SendTextMessage(channelId, conversationId, text)
}
