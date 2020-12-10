package chat_manager

import (
	client "github.com/webitel/engine/chat_manager/chat"
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
	Decline(authUserId int64, inviteId string) error
	Leave(authUserId int64, channelId, conversationId string) error
	CloseConversation(authUserId int64, channelId, conversationId, cause string) error

	SendText(authUserId int64, channelId, conversationId, text string) error
	SendFile(authUserId int64, channelId, conversationId, url, mimeType string) error

	AddToChat(authUserId, userId int64, channelId, conversationId, title string) error
	NewInternalChat(domainId, authUserId, userId int64) error
	UpdateChannel(authUserId int64, channelId string) error
	ListActive(token string, domainId, userId int64, page, size int) (*client.GetConversationsResponse, error)
}

type chatConnection struct {
	name   string
	host   string
	client *grpc.ClientConn
	api    client.ChatServiceClient
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

	connection.api = client.NewChatServiceClient(connection.client)

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
