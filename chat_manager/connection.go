package chat_manager

import (
	gogrpc "buf.build/gen/go/webitel/chat/grpc/go/_gogrpc"
	messgrpc "buf.build/gen/go/webitel/chat/grpc/go/messages/messagesgrpc"
	proto "buf.build/gen/go/webitel/chat/protocolbuffers/go"
	"context"
	"github.com/webitel/engine/model"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"time"
)

const (
	CONNECTION_TIMEOUT = 2 * time.Second
)

type Chat interface {
	Name() string
	Close() error
	Ready() bool

	Join(authUserId int64, inviteId string) (string, error)
	Decline(authUserId int64, inviteId string, cause string) error
	Leave(authUserId int64, channelId, conversationId string) error
	CloseConversation(authUserId int64, channelId, conversationId, cause string) error

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
	BroadcastMessage(ctx context.Context, message *proto.Message, profileId int64, peer []string) error
	SetContact(token string, channelId string, conversationId string, contactId int64) error
}

type chatConnection struct {
	name    string
	host    string
	client  *grpc.ClientConn
	api     gogrpc.ChatServiceClient
	mess    gogrpc.MessagesClient
	contact messgrpc.ContactLinkingServiceClient
}

func NewChatServiceConnection(name, url string) (Chat, error) {
	var err error
	connection := &chatConnection{
		name: name,
		host: url,
	}

	connection.client, err = grpc.Dial(url, grpc.WithInsecure(), grpc.WithBlock(), grpc.WithTimeout(CONNECTION_TIMEOUT))

	if err != nil {
		return nil, err
	}

	connection.api = gogrpc.NewChatServiceClient(connection.client)
	connection.mess = gogrpc.NewMessagesClient(connection.client)

	return connection, nil
}

func (cc *chatConnection) Ready() bool {
	switch cc.client.GetState() {
	case connectivity.Idle, connectivity.Ready:
		return true
	}
	return false
}

func (cc *chatConnection) Name() string {
	return cc.name
}

func (cc *chatConnection) Close() error {
	err := cc.client.Close()
	if err != nil {
		return ErrInternal
	}
	return nil
}
