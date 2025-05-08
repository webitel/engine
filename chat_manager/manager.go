package chat_manager

import (
	"github.com/webitel/engine/gen/chat"
	msg "github.com/webitel/engine/gen/chat/messages"
	"github.com/webitel/engine/pkg/wbt"
	"github.com/webitel/wlog"
	"sync"
)

const (
	ChatServiceName = "webitel.chat.server"
)

type ChatManager interface {
	Start() error
	Client() (Chat, error)
	Stop()
}

type chatManager struct {
	consulAddr string
	startOnce  sync.Once
	chatCli    *wbt.Client[chat.ChatServiceClient]
	contactCli *wbt.Client[msg.ContactLinkingServiceClient]
	chat       Chat
}

func NewChatManager(consulAddr string) ChatManager {
	return &chatManager{
		consulAddr: consulAddr,
	}
}

func (cm *chatManager) Start() error {
	wlog.Debug("starting chat service client")
	var err error
	cm.startOnce.Do(func() {
		cm.chatCli, err = wbt.NewClient(cm.consulAddr, ChatServiceName, chat.NewChatServiceClient)
		if err != nil {
			return
		}
		cm.contactCli, err = wbt.NewClient(cm.consulAddr, ChatServiceName, msg.NewContactLinkingServiceClient)
		if err != nil {
			return
		}
		cm.chat = NewChat(cm)
	})
	return err
}

func (cm *chatManager) Stop() {
	if cm.contactCli != nil {
		_ = cm.contactCli.Close()
	}
	if cm.chatCli != nil {
		_ = cm.chatCli.Close()
	}
}

func (cm *chatManager) Client() (Chat, error) {
	return cm.chat, nil
}
