package mq

import (
	"github.com/webitel/engine/model"
)

type MQ interface {
	SendJSON(name string, data []byte) *model.AppError
	BindCallEvents(domainId, userId int64) error
	UnBindCallEvents(domainId, userId int64) error
	Start()
	Close()
	NewDomainQueue(domainId int64, bindings model.GetAllBindings) (DomainQueue, *model.AppError)
}

type DomainQueue interface {
	Start()
	Stop()
	CallEvents() <-chan *model.Call
	BindUserCall(id string, userId int64) *model.BindQueueEvent
	BulkUnbind(b []*model.BindQueueEvent) *model.AppError
	Unbind(bind *model.BindQueueEvent) *model.AppError

	Listen() error
}
