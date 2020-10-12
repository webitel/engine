package chat_manager

import (
	client "github.com/webitel/protos/chat"
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

	Join(inviteId string) (string, error)
	Decline(userId int64, inviteId string) error
	Leave(channelId, conversationId string) error
	CloseConversation(channelId, conversationId, cause string) error

	SendText(channelId, conversationId, text string) error
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
