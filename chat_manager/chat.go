package chat_manager

import (
	"context"
	proto "github.com/webitel/engine/gen/chat"
	"github.com/webitel/engine/model"
)

type Chat interface {
	Join(authUserId int64, inviteId string) (string, error)
	Decline(authUserId int64, inviteId string, cause string) error
	Leave(authUserId int64, channelId, conversationId string, cause model.LeaveCause) error
	CloseConversation(authUserId int64, channelId, conversationId string, cause model.CloseCause) error

	SendText(authUserId int64, channelId, conversationId, text string) error
	SendFile(authUserId int64, channelId, conversationId string, file *model.ChatFile) error

	AddToChat(authUserId, userId int64, channelId, conversationId, title string) error
	NewInternalChat(domainId, authUserId, userId int64) error
	UpdateChannel(authUserId int64, channelId string, readUntil int64) error
	ListActive(token string, domainId, userId int64, page, size int) (*proto.GetConversationsResponse, error)
	InviteToConversation(ctx context.Context, domainId, userId int64, conversationId, inviterId, invUserId, title string, timeout int, vars map[string]string) (string, error)
	BlindTransfer(ctx context.Context, conversationId, channelId string, schemaId int64, vars map[string]string) error
	BlindTransferToUser(ctx context.Context, conversationId, channelId string, userId int64, vars map[string]string) error
	SetVariables(channelId string, vars map[string]string) error
	SetContact(ctx context.Context, channelId string, conversationId string, contactId int64) error
}

type chatConnection struct {
	cm *chatManager
}

func NewChat(cm *chatManager) Chat {
	return &chatConnection{
		cm: cm,
	}
}
