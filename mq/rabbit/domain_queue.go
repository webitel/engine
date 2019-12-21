package rabbit

import (
	"encoding/json"
	"fmt"
	"github.com/streadway/amqp"
	"github.com/webitel/engine/model"
	"github.com/webitel/engine/mq"
	"github.com/webitel/wlog"
	"sync"
	"time"
)

type DomainQueue struct {
	name         string
	id           int64
	client       *AMQP
	channel      *amqp.Channel
	closeChannel chan *amqp.Error

	queue      amqp.Queue
	callEvents chan *model.Call

	bindChan chan *model.BindQueueEvent

	fnGetAllBindings model.GetAllBindings

	delivery  <-chan amqp.Delivery
	startOnce sync.Once
	reconnect chan error

	stop    chan struct{}
	stopped chan struct{}
}

func (dq *DomainQueue) Name() string {
	return dq.name
}

func (dq *DomainQueue) Id() int64 {
	return dq.id
}

func newDomainQueue(client *AMQP, id int64, bindings model.GetAllBindings) mq.DomainQueue {
	q := &DomainQueue{
		id:     id,
		client: client,
		name:   fmt.Sprintf("domain.%v", id),

		callEvents:       make(chan *model.Call),
		fnGetAllBindings: bindings,

		bindChan: make(chan *model.BindQueueEvent, 100), //TODO

		closeChannel: make(chan *amqp.Error, 1),
		reconnect:    make(chan error, 1),
		stop:         make(chan struct{}),
		stopped:      make(chan struct{}),
	}
	return q
}

func (dq *DomainQueue) Start() {
	wlog.Debug(fmt.Sprintf("DomainQueue [%d] started", dq.Id()))
	dq.startOnce.Do(func() {
		dq.connect()
	})
}

func (dq *DomainQueue) BindUserCall(id string, userId int64) *model.BindQueueEvent {
	b := &model.BindQueueEvent{
		UserId:   userId,
		Id:       id,
		Routing:  fmt.Sprintf(model.MQ_CALL_TEMPLATE_ROUTING_KEY, dq.Id(), userId),
		Exchange: model.MQ_CALL_EXCHANGE,
	}

	dq.bindChan <- b
	return b
}

func (dq *DomainQueue) Unbind(bind *model.BindQueueEvent) *model.AppError {
	dq.channel.QueueUnbind(dq.queue.Name, bind.Routing, bind.Exchange, amqp.Table{
		"x-sock-id": bind.Id,
	})
	wlog.Debug(fmt.Sprintf("DomainQueue [%d] unbind userId=%d sockId=%s to call events", dq.Id(), bind.UserId, bind.Id))
	//TODO check error
	return nil
}

func (dq *DomainQueue) BulkUnbind(b []*model.BindQueueEvent) *model.AppError {
	var err error
	for _, v := range b {
		err = dq.Unbind(v)

		if err != nil {
			//TODO
		}
	}
	return nil
}

func (dq *DomainQueue) bind(b *model.BindQueueEvent) {
	err := dq.channel.QueueBind(
		dq.queue.Name,
		b.Routing,
		b.Exchange,
		true,
		amqp.Table{
			"x-sock-id": b.Id,
		})
	if err != nil {
		wlog.Error(fmt.Sprintf("DomainQueue [%d] bind userId=%d sockId=%s to call events: %s", dq.Id(), b.UserId, b.Id, err.Error()))
	} else {
		wlog.Debug(fmt.Sprintf("DomainQueue [%d] bind userId=%d sockId=%s to call events", dq.Id(), b.UserId, b.Id))
	}
}

func (dq *DomainQueue) readMessage(m amqp.Delivery) {
	if m.ContentType != "text/json" {
		wlog.Warn(fmt.Sprintf("DomainQueue [%d] failed receive event content type: %v\n%s", dq.Id(), m.ContentType, m.Body))
		return
	}

	switch m.Exchange {
	case model.MQ_CALL_EXCHANGE:
		dq.readCallMessage(m.Body, m.RoutingKey)
	default:
		wlog.Error(fmt.Sprintf("DomainQueue [%d] not implement parser from exchange %s", dq.Id(), m.Exchange))
	}
}

func parseCallEvent(data []byte) (*model.Call, error) {
	fmt.Println(string(data))
	var call model.Call
	err := json.Unmarshal(data, &call)
	if err != nil {
		return nil, err
	}

	return &call, nil
}

func (dq *DomainQueue) readCallMessage(data []byte, rk string) {
	e, err := parseCallEvent(data)
	if err != nil {
		wlog.Warn(err.Error())
		wlog.Warn(fmt.Sprintf("DomainQueue [%d] failed parse json event, skip %s", dq.Id(), string(data)))
		return
	}

	wlog.Debug(fmt.Sprintf("DomainQueue [%d] receive event %v:%v [%v] rk=%s", dq.Id(), e.NodeName, e.Id, e.Action, rk))
	dq.callEvents <- e
}

func (dq *DomainQueue) connect() error {
	var err error
	wlog.Debug(fmt.Sprintf("DomainQueue [%d] trying connect...", dq.Id()))

	defer func() {
		if err != nil {
			time.Sleep(time.Second * RECONNECT_SEC)
			go dq.connect()
		}
	}()

	dq.channel, err = dq.client.NewChannel()
	if err != nil {
		return err
	}
	dq.closeChannel = make(chan *amqp.Error, 1)
	dq.channel.NotifyClose(dq.closeChannel)

	dq.queue, err = dq.channel.QueueDeclare(
		fmt.Sprintf("engine.call.%s.%d", model.NewId()[0:10], dq.id),
		false,
		false,
		true,
		false,
		nil,
	)

	if err != nil {
		return err
	}

	dq.delivery, err = dq.channel.Consume(
		dq.queue.Name,
		model.NewId(),
		true,
		true,
		false,
		false,
		nil,
	)

	if err != nil {
		return err
	}

	wlog.Debug(fmt.Sprintf("DomainQueue [%d] connected", dq.Id()))

	dq.rebindingUsers()

	go func() {
		if err := dq.Listen(); err != nil {
			wlog.Error(fmt.Sprintf("DomainQueue [%d] error: %s", dq.Id(), err.Error()))
		}
	}()

	return nil
}

func (dq *DomainQueue) rebindingUsers() {
	for _, v := range dq.fnGetAllBindings() {
		dq.bind(v)
	}

	if len(dq.bindChan) > 0 {
		for v := range dq.bindChan {
			dq.bind(v)
		}
	}
}

func (dq *DomainQueue) Listen() error {
	var ok bool
	var err error
	wlog.Debug(fmt.Sprintf("DomainQueue [%d] start listener", dq.Id()))
	defer wlog.Debug(fmt.Sprintf("DomainQueue [%d] close listener", dq.Id()))

	for {
		select {
		case err, ok = <-dq.closeChannel:
			go dq.connect()
			if !ok {
				return nil
			}
			return err
		case <-dq.stop:
			return nil
		case m, ok := <-dq.delivery:
			if !ok {
				return nil
			}
			dq.readMessage(m)
		case b := <-dq.bindChan:
			dq.bind(b)
		}
	}
}

func (dq *DomainQueue) CallEvents() <-chan *model.Call {
	return dq.callEvents
}

func (dq *DomainQueue) Stop() {
	wlog.Debug(fmt.Sprintf("DomainQueue [%d] stopping", dq.Id()))
	close(dq.stop)
	<-dq.stopped
}

func (dq *DomainQueue) getCallEvent(data []byte) *model.Call {
	e := &REvent{}
	err := json.Unmarshal(data, e)
	if err != nil {
		wlog.Warn(err.Error())
		wlog.Warn(fmt.Sprintf("DomainQueue [%d] failed parse json event, skip %s", dq.Id(), data))
		return nil
	}

	return nil
}
