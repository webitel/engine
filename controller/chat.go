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
	return c.app.JoinChat(session.UserId, inviteId)
}

func (c *Controller) LeaveChat(session *auth_manager.Session, channelId, conversationId string) *model.AppError {
	// FIXME PERMISSION
	return c.app.LeaveChat(session.UserId, channelId, conversationId)
}

func (c *Controller) CloseChat(session *auth_manager.Session, channelId, conversationId, cause string) *model.AppError {
	// FIXME PERMISSION
	return c.app.CloseChat(session.UserId, channelId, conversationId, cause)
}

func (c *Controller) SendTextChat(session *auth_manager.Session, channelId, conversationId, text string) *model.AppError {
	// FIXME PERMISSION
	return c.app.SendTextMessage(session.UserId, channelId, conversationId, text)
}

func (c *Controller) SendFileChat(session *auth_manager.Session, channelId, conversationId string, file *model.ChatFile) *model.AppError {
	// FIXME PERMISSION
	return c.app.SendFileMessage(session.UserId, channelId, conversationId, file)
}

func (c *Controller) AddToChat(session *auth_manager.Session, userId int64, channelId, conversationId, title string) *model.AppError {
	// FIXME PERMISSION
	return c.app.AddToChat(session.UserId, userId, channelId, conversationId, title)
}

func (c *Controller) StartChat(session *auth_manager.Session, userId int64) *model.AppError {
	// FIXME PERMISSION
	return c.app.StartChat(session.DomainId, session.UserId, userId)
}

func (c *Controller) UpdateChannelChat(session *auth_manager.Session, channelId string, readUntil int64) *model.AppError {
	// FIXME PERMISSION
	return c.app.UpdateChannelChat(session.UserId, channelId, readUntil)
}

func (c *Controller) ListActiveChat(session *auth_manager.Session, page, size int) ([]*model.Conversation, *model.AppError) {
	// FIXME PERMISSION
	return c.app.ListActiveChat(session.Token, session.DomainId, session.UserId, page, size)
}

func (c *Controller) BlindTransferChat(session *auth_manager.Session, conversationId, channelId string, planId int32, vars map[string]string) *model.AppError {
	// FIXME PERMISSION
	return c.app.BlindTransferChat(session.DomainId, conversationId, channelId, planId, vars)
}
