package mq

import (
	"github.com/webitel/engine/model"
)

type MQ interface {
	SendJSON(name string, data []byte) model.AppError
	BindCallEvents(domainId, userId int64) error
	UnBindCallEvents(domainId, userId int64) error

	Start()
	Close()
	NewDomainQueue(domainId int64, bindings model.GetAllBindings) (DomainQueue, model.AppError)

	RegisterWebsocket(domainId int64, event *model.RegisterToWebsocketEvent) model.AppError
	UnRegisterWebsocket(domainId int64, event *model.RegisterToWebsocketEvent) model.AppError
	SendNotification(domainId int64, event *model.Notification) model.AppError

	SendStickingCall(e *model.CallServiceHangup) model.AppError
}

type DomainQueue interface {
	Start()
	Stop()
	Events() <-chan *model.WebSocketEvent
	CallEvents() <-chan *model.CallEvent
	UserStateEvents() <-chan *model.UserState
	ChatEvents() <-chan *model.ChatEvent
	NotificationEvents() <-chan *model.Notification

	BindUserCall(id string, userId int64) *model.BindQueueEvent
	BindUserChat(id string, userId int64) *model.BindQueueEvent
	BulkUnbind(b []*model.BindQueueEvent) model.AppError
	Unbind(bind *model.BindQueueEvent) model.AppError

	BindUsersStatus(id string, userId int64) *model.BindQueueEvent
	BindAgentStatusEvents(id string, userId int64, agentId int) *model.BindQueueEvent
	BindAgentChannelEvents(id string, userId int64, agentId int) *model.BindQueueEvent

	Listen() error
}
