package rabbit

import (
	"encoding/json"
	"fmt"
	"github.com/streadway/amqp"
	"github.com/webitel/engine/model"
	"github.com/webitel/engine/mq"
	"github.com/webitel/wlog"
	"strconv"
	"strings"
	"sync"
	"time"
)

type DomainQueue struct {
	name         string
	id           int64
	client       *AMQP
	channel      *amqp.Channel
	closeChannel chan *amqp.Error

	queue  amqp.Queue
	events chan *model.WebSocketEvent

	callEvents        chan *model.CallEvent
	userStateEvents   chan *model.UserState
	chatEvents        chan *model.ChatEvent
	notificationEvent chan *model.Notification

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
		events: make(chan *model.WebSocketEvent),

		callEvents:        make(chan *model.CallEvent),
		userStateEvents:   make(chan *model.UserState),
		chatEvents:        make(chan *model.ChatEvent),
		notificationEvent: make(chan *model.Notification),
		fnGetAllBindings:  bindings,

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
		Routing:  fmt.Sprintf(model.CallRoutingTemplate, dq.Id(), userId),
		Exchange: model.CallExchange,
	}

	dq.bindChan <- b
	return b
}

func (dq *DomainQueue) BindUsersStatus(id string, userId int64) *model.BindQueueEvent {
	b := &model.BindQueueEvent{
		UserId:   userId,
		Id:       id,
		Routing:  fmt.Sprintf(model.MQ_USER_STATUS_TEMPLATE_ROUTING_KEY, dq.Id()),
		Exchange: model.MQ_USER_STATUS_EXCHANGE,
	}

	dq.bindChan <- b
	return b
}

func (dq *DomainQueue) BindUserChat(id string, userId int64) *model.BindQueueEvent {
	b := &model.BindQueueEvent{
		UserId:   userId,
		Id:       id,
		Routing:  fmt.Sprintf("event.*.%d.%d", dq.Id(), userId),
		Exchange: model.ChatExchange,
	}

	dq.bindChan <- b
	return b
}

func (dq *DomainQueue) BindAgentStatusEvents(id string, userId int64, agentId int) *model.BindQueueEvent {
	b := &model.BindQueueEvent{
		UserId:   userId,
		Id:       id,
		Routing:  fmt.Sprintf(model.CallCenterAgentStateTemplate, dq.id, userId),
		Exchange: model.CallCenterExchange,
	}

	dq.bindChan <- b

	return b
}

func (dq *DomainQueue) BindAgentChannelEvents(id string, userId int64, agentId int) *model.BindQueueEvent {
	b2 := &model.BindQueueEvent{
		UserId:   userId,
		Id:       id,
		Routing:  fmt.Sprintf("events.channel.*.%d.*.%d", dq.id, userId),
		Exchange: model.CallCenterExchange,
	}

	dq.bindChan <- b2

	return b2
}

func (dq *DomainQueue) Unbind(bind *model.BindQueueEvent) *model.AppError {
	/*
		2020-03-24T01:11:42.129+0200    debug   app/web_hub.go:67       hub TODO stopped
		panic: runtime error: invalid memory address or nil pointer dereference

	*/
	dq.channel.QueueUnbind(dq.queue.Name, bind.Routing, bind.Exchange, amqp.Table{
		"x-sock-id": bind.Id,
	})
	wlog.Debug(fmt.Sprintf("DomainQueue [%d] unbind userId=%d sockId=%s from %s", dq.Id(), bind.UserId, bind.Id, bind.Exchange))
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
		wlog.Error(fmt.Sprintf("DomainQueue [%d] bind userId=%d sockId=%s to %s events: %s", dq.Id(), b.UserId, b.Id, b.Routing, err.Error()))
	} else {
		wlog.Debug(fmt.Sprintf("DomainQueue [%d] bind userId=%d sockId=%s to %s events", dq.Id(), b.UserId, b.Id, b.Routing))
	}
}

func (dq *DomainQueue) readMessage(m amqp.Delivery) {
	if m.ContentType != "text/json" && m.Exchange != model.ChatExchange {
		wlog.Warn(fmt.Sprintf("DomainQueue [%d] failed receive event content type: %v\n%s", dq.Id(), m.ContentType, m.Body))
		return
	}

	switch m.Exchange {
	case model.CallExchange:
		dq.readCallMessage(m.Body, m.RoutingKey)

	case model.CallCenterExchange:
		dq.readAgentStatusEvent(m.Body, m.RoutingKey)

	case model.ChatExchange:
		dq.readChatEvent(m.Body, m.RoutingKey)

	case model.AppExchange:
		dq.readAppMessage(m.Body, m.RoutingKey)

	case model.MQ_USER_STATUS_EXCHANGE:
		dq.readUserStateMessage(m.Body, m.RoutingKey)

	default:
		wlog.Error(fmt.Sprintf("DomainQueue [%d] not implement parser from exchange %s", dq.Id(), m.Exchange))
	}
}

func parseNotification(data []byte) (*model.Notification, error) {
	var n model.Notification
	err := json.Unmarshal(data, &n)
	if err != nil {
		return nil, err
	}

	return &n, nil
}

func (dq *DomainQueue) readAppMessage(data []byte, rk string) {
	route := strings.Split(rk, ".")
	if len(route) < 1 {
		wlog.Error(fmt.Sprintf("DomainQueue [%d] read app message, error: bad routing key %s", dq.id, rk))
	}

	switch route[0] {
	case "notification":
		e, err := parseNotification(data)
		if err != nil {
			wlog.Warn(err.Error())
			wlog.Warn(fmt.Sprintf("DomainQueue [%d] failed parse json event notification, skip %s", dq.Id(), string(data)))
			return
		}

		wlog.Debug(fmt.Sprintf("DomainQueue [%d] receive notification event %v rk=%s", dq.Id(), e.Id, rk))
		dq.notificationEvent <- e

	default:
		wlog.Error(fmt.Sprintf("DomainQueue [%d] read app message, error: no handler %s", dq.id, rk))
	}
}

func (dq *DomainQueue) readChatEvent(data []byte, rk string) {
	ev := parseChatMessage(data, rk)
	if ev != nil {
		dq.chatEvents <- ev
	}
}

func parseChatMessage(body []byte, rk string) *model.ChatEvent {
	rks := strings.Split(rk, ".")
	if len(rks) != 4 {
		wlog.Error(fmt.Sprintf("event %s: bad rk format", rk))
		return nil
	}

	domainId, err := strconv.Atoi(rks[2])
	if err != nil {
		wlog.Error(fmt.Sprintf("event %s: bad domainId", rk))
		return nil

	}

	userId, err := strconv.Atoi(rks[3])
	if err != nil {
		wlog.Error(fmt.Sprintf("event %s: bad userId", rk))
		return nil
	}

	ev := &model.ChatEvent{
		Event:    rks[1],
		DomainId: int64(domainId),
		UserId:   int64(userId),
		Data:     nil,
	}

	err = json.Unmarshal(body, &ev.Data)
	if err != nil {
		wlog.Error(fmt.Sprintf("event %s: %s", rk, err.Error()))
		return nil
	}

	return ev
}

func (dq *DomainQueue) readAgentStatusEvent(data []byte, rk string) {
	e, err := parseAgentStatusEvent(data)
	if err != nil {
		wlog.Warn(err.Error())
		wlog.Warn(fmt.Sprintf("DomainQueue [%d] failed parse json event, skip %s", dq.Id(), string(data)))
		return
	}

	if ev, appErr := model.NewWebSocketCallCenterEvent(e); appErr != nil {
		wlog.Warn(appErr.Error())
		wlog.Warn(fmt.Sprintf("DomainQueue [%d] failed parse event, skip %s", dq.Id(), string(data)))
		return
	} else {
		wlog.Debug(fmt.Sprintf("DomainQueue [%d] receive cc event %v <%v> rk=%s", dq.Id(), e.Event, string(data), rk))
		dq.events <- ev
	}
}

func parseAgentStatusEvent(data []byte) (*model.CallCenterEvent, error) {
	var e *model.CallCenterEvent
	err := json.Unmarshal(data, &e)
	if err != nil {
		return nil, err
	}

	return e, nil
}

func parseCallEvent(data []byte) (*model.CallEvent, error) {
	var call model.CallEvent
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

	wlog.Debug(fmt.Sprintf("DomainQueue [%d] receive call event %v:%v [%v] rk=%s", dq.Id(), e.AppId, e.Id, e.Event, rk))
	dq.callEvents <- e
}

func parseUserStateEvent(data []byte) (*model.UserState, error) {
	var state model.UserState
	err := json.Unmarshal(data, &state)
	if err != nil {
		return nil, err
	}

	return &state, nil
}

func (dq *DomainQueue) readUserStateMessage(data []byte, rk string) {
	e, err := parseUserStateEvent(data)
	if err != nil {
		wlog.Warn(err.Error())
		wlog.Warn(fmt.Sprintf("DomainQueue [%d] failed parse json event, skip %s", dq.Id(), string(data)))
		return
	}
	wlog.Debug(fmt.Sprintf("DomainQueue [%d] receive event %v:%v [%v] rk=%s", dq.Id(), e.App, e.Id, e.Status, rk))
	dq.userStateEvents <- e

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

	err = dq.channel.QueueBind(dq.queue.Name, fmt.Sprintf("notification.%d", dq.id), model.AppExchange, false, nil)
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

func (dq *DomainQueue) removeQueue() {
	dq.channel.QueueDelete(dq.queue.Name, false, false, true)
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
			dq.removeQueue()
			dq.client.RemoveDomainQueue(dq)
			close(dq.stopped)
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

func (dq *DomainQueue) Events() <-chan *model.WebSocketEvent {
	return dq.events
}

func (dq *DomainQueue) CallEvents() <-chan *model.CallEvent {
	return dq.callEvents
}

func (dq *DomainQueue) ChatEvents() <-chan *model.ChatEvent {
	return dq.chatEvents
}

func (dq *DomainQueue) UserStateEvents() <-chan *model.UserState {
	return dq.userStateEvents
}

func (dq *DomainQueue) NotificationEvents() <-chan *model.Notification {
	return dq.notificationEvent
}

func (dq *DomainQueue) Stop() {
	wlog.Debug(fmt.Sprintf("DomainQueue [%d] stopping", dq.Id()))
	close(dq.stop)
	<-dq.stopped
}

func (dq *DomainQueue) getCallEvent(data []byte) *model.CallEvent {
	e := &REvent{}
	err := json.Unmarshal(data, e)
	if err != nil {
		wlog.Warn(err.Error())
		wlog.Warn(fmt.Sprintf("DomainQueue [%d] failed parse json event, skip %s", dq.Id(), data))
		return nil
	}

	return nil
}
