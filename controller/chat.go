package controller

import (
	"context"

	"github.com/webitel/engine/model"
	"github.com/webitel/engine/pkg/wbt/auth_manager"
)

var (
	noChatAccessError = model.NewForbiddenError("chat.valid.license", "no CALL_CENTER license was found")
)

func (c *Controller) DeclineChat(session *auth_manager.Session, inviteId string, cause string) model.AppError {
	if !session.HasCallCenterLicense() {
		return noChatAccessError
	}

	return c.app.DeclineChat(session.UserId, inviteId, cause)
}

func (c *Controller) JoinChat(session *auth_manager.Session, inviteId string) (string, model.AppError) {
	if !session.HasCallCenterLicense() {
		return "", noChatAccessError
	}

	return c.app.JoinChat(session.UserId, inviteId)
}

func (c *Controller) LeaveChat(session *auth_manager.Session, channelId, conversationId string, reason model.LeaveCause) model.AppError {
	if !session.HasCallCenterLicense() {
		return noChatAccessError
	}

	return c.app.LeaveChat(session.UserId, channelId, conversationId, reason)
}

func (c *Controller) CloseChat(session *auth_manager.Session, channelId, conversationId string, cause model.CloseCause) model.AppError {
	if !session.HasCallCenterLicense() {
		return noChatAccessError
	}

	return c.app.CloseChat(session.UserId, channelId, conversationId, cause)
}

func (c *Controller) SendTextChat(session *auth_manager.Session, channelId, conversationId, text string) model.AppError {
	if !session.HasCallCenterLicense() {
		return noChatAccessError
	}

	return c.app.SendTextMessage(session.UserId, channelId, conversationId, text)
}

func (c *Controller) SendFileChat(session *auth_manager.Session, channelId, conversationId string, file *model.ChatFile) model.AppError {
	if !session.HasCallCenterLicense() {
		return noChatAccessError
	}

	return c.app.SendFileMessage(session.UserId, channelId, conversationId, file)
}

func (c *Controller) AddToChat(session *auth_manager.Session, userId int64, channelId, conversationId, title string) model.AppError {
	if !session.HasCallCenterLicense() {
		return noChatAccessError
	}

	return c.app.AddToChat(session.UserId, userId, channelId, conversationId, title)
}

func (c *Controller) StartChat(session *auth_manager.Session, userId int64) model.AppError {
	if !session.HasCallCenterLicense() {
		return noChatAccessError
	}

	return c.app.StartChat(session.DomainId, session.UserId, userId)
}

func (c *Controller) UpdateChannelChat(session *auth_manager.Session, channelId string, readUntil int64) model.AppError {
	if !session.HasCallCenterLicense() {
		return noChatAccessError
	}

	return c.app.UpdateChannelChat(session.UserId, channelId, readUntil)
}

func (c *Controller) ListActiveChat(ctx context.Context, session *auth_manager.Session, page, size int) ([]*model.Conversation, model.AppError) {
	if !session.HasCallCenterLicense() {
		return nil, noChatAccessError
	}

	permission := session.GetPermission(model.PermissionContacts)
	return c.app.ListActiveChat(ctx, session.Token, session.DomainId, session.UserId, page, size, permission.CanRead())
}

func (c *Controller) BlindTransferChat(ctx context.Context, session *auth_manager.Session, conversationId, channelId string, planId int32, vars map[string]string) model.AppError {
	if !session.HasCallCenterLicense() {
		return noChatAccessError
	}

	return c.app.BlindTransferChat(ctx, session.DomainId, conversationId, channelId, planId, vars)
}

// todo check userId in domain
func (c *Controller) BlindTransferChatToUser(session *auth_manager.Session, conversationId, channelId string, userId int64, vars map[string]string) model.AppError {
	if !session.HasCallCenterLicense() {
		return noChatAccessError
	}

	return c.app.BlindTransferChatToUser(session.DomainId, conversationId, channelId, userId, vars)
}

// todo check userId in domain
func (c *Controller) SetContactChat(session *auth_manager.Session, channelId string, conversationId string, contactId int64) model.AppError {
	if !session.HasCallCenterLicense() {
		return noChatAccessError
	}
	return c.app.SetChatContactId(context.Background(), session.DomainId, session.UserId, contactId, channelId, conversationId)
}
