package mq

import (
	"context"

	"github.com/webitel/engine/model"
)

type LayeredMQLayer interface {
	MQ
}

type LayeredMQ struct {
	context context.Context
	MQLayer LayeredMQLayer
}

func NewMQ(mq LayeredMQLayer) MQ {
	return &LayeredMQ{
		context: context.TODO(),
		MQLayer: mq,
	}
}

func (l *LayeredMQ) SendJSON(name string, data []byte) model.AppError {
	return l.MQLayer.SendJSON(name, data)
}

func (l *LayeredMQ) Start() {
	l.MQLayer.Start()
}

func (l *LayeredMQ) Close() {
	l.MQLayer.Close()
}

func (l *LayeredMQ) BindCallEvents(domainId, userId int64) error {
	return l.MQLayer.BindCallEvents(domainId, userId)
}

func (l *LayeredMQ) UnBindCallEvents(domainId, userId int64) error {
	return l.MQLayer.UnBindCallEvents(domainId, userId)
}

func (l *LayeredMQ) NewDomainQueue(domainId int64, bindings model.GetAllBindings) (DomainQueue, model.AppError) {
	return l.MQLayer.NewDomainQueue(domainId, bindings)
}

func (l *LayeredMQ) RegisterWebsocket(domainId int64, event *model.RegisterToWebsocketEvent) model.AppError {
	return l.MQLayer.RegisterWebsocket(domainId, event)
}

func (l *LayeredMQ) UnRegisterWebsocket(domainId int64, event *model.RegisterToWebsocketEvent) model.AppError {
	return l.MQLayer.UnRegisterWebsocket(domainId, event)
}

func (l *LayeredMQ) SendStickingCall(event *model.CallServiceHangup) model.AppError {
	return l.MQLayer.SendStickingCall(event)
}

func (l *LayeredMQ) SendNotification(domainId int64, event *model.Notification) model.AppError {
	return l.MQLayer.SendNotification(domainId, event)
}

func (l *LayeredMQ) Send(ctx context.Context, exchange string, rk string, body []byte) error {
	return l.MQLayer.Send(ctx, exchange, rk, body)
}

func (l *LayeredMQ) SendStartFlow(ctx context.Context, domainId int64, schemaId int32, in interface{}) model.AppError {
	return l.MQLayer.SendStartFlow(ctx, domainId, schemaId, in)
}
