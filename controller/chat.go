package controller

import (
	"context"

	"github.com/webitel/engine/auth_manager"
	"github.com/webitel/engine/model"
)

func (c *Controller) DeclineChat(session *auth_manager.Session, inviteId string, cause string) model.AppError {
	// FIXME PERMISSION
	return c.app.DeclineChat(session.UserId, inviteId, cause)
}

func (c *Controller) JoinChat(session *auth_manager.Session, inviteId string) (string, model.AppError) {
	// FIXME PERMISSION
	return c.app.JoinChat(session.UserId, inviteId)
}

func (c *Controller) LeaveChat(session *auth_manager.Session, channelId, conversationId string) model.AppError {
	// FIXME PERMISSION
	return c.app.LeaveChat(session.UserId, channelId, conversationId)
}

func (c *Controller) CloseChat(session *auth_manager.Session, channelId, conversationId, cause string) model.AppError {
	// FIXME PERMISSION
	return c.app.CloseChat(session.UserId, channelId, conversationId, cause)
}

func (c *Controller) SendTextChat(session *auth_manager.Session, channelId, conversationId, text string) model.AppError {
	// FIXME PERMISSION
	return c.app.SendTextMessage(session.UserId, channelId, conversationId, text)
}

func (c *Controller) SendFileChat(session *auth_manager.Session, channelId, conversationId string, file *model.ChatFile) model.AppError {
	// FIXME PERMISSION
	return c.app.SendFileMessage(session.UserId, channelId, conversationId, file)
}

func (c *Controller) AddToChat(session *auth_manager.Session, userId int64, channelId, conversationId, title string) model.AppError {
	// FIXME PERMISSION
	return c.app.AddToChat(session.UserId, userId, channelId, conversationId, title)
}

func (c *Controller) StartChat(session *auth_manager.Session, userId int64) model.AppError {
	// FIXME PERMISSION
	return c.app.StartChat(session.DomainId, session.UserId, userId)
}

func (c *Controller) UpdateChannelChat(session *auth_manager.Session, channelId string, readUntil int64) model.AppError {
	// FIXME PERMISSION
	return c.app.UpdateChannelChat(session.UserId, channelId, readUntil)
}

func (c *Controller) ListActiveChat(ctx context.Context, session *auth_manager.Session, page, size int) ([]*model.Conversation, model.AppError) {
	// FIXME PERMISSION
	return c.app.ListActiveChat(ctx, session.Token, session.DomainId, session.UserId, page, size)
}

func (c *Controller) BlindTransferChat(ctx context.Context, session *auth_manager.Session, conversationId, channelId string, planId int32, vars map[string]string) model.AppError {
	// FIXME PERMISSION
	return c.app.BlindTransferChat(ctx, session.DomainId, conversationId, channelId, planId, vars)
}

// todo check userId in domain
func (c *Controller) BlindTransferChatToUser(session *auth_manager.Session, conversationId, channelId string, userId int64, vars map[string]string) model.AppError {
	// FIXME PERMISSION
	return c.app.BlindTransferChatToUser(session.DomainId, conversationId, channelId, userId, vars)
}

func (c *Controller) BroadcastChatBot(session *auth_manager.Session, profileId int64, peer []string, text string) model.AppError {
	permission := session.GetPermission(model.PermissionChat)
	if !permission.CanRead() {
		return c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	return c.app.BroadcastChatBot(context.TODO(), session.Domain(0), profileId, peer, text)
}

// todo check userId in domain
func (c *Controller) SetContactChat(session *auth_manager.Session, channelId string, conversationId string, contactId int64) model.AppError {
	// FIXME PERMISSION
	return c.app.SetContactToChat(session.Token, channelId, conversationId, contactId)
}
