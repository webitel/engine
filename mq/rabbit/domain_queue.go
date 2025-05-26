package rabbit

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/webitel/engine/model"
	"github.com/webitel/engine/mq"
	"github.com/webitel/wlog"
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
	log     *wlog.Logger

	sync.RWMutex
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

		bindChan: make(chan *model.BindQueueEvent, 1000), //TODO

		closeChannel: make(chan *amqp.Error, 1),
		reconnect:    make(chan error, 1),
		stop:         make(chan struct{}),
		stopped:      make(chan struct{}),
		log: client.log.With(
			wlog.String("scope", "domain queue"),
			wlog.Int64("domain_id", id),
		),
	}
	return q
}

func (dq *DomainQueue) getChannel() *amqp.Channel {
	dq.RLock()
	ch := dq.channel
	dq.RUnlock()

	return ch
}

func (dq *DomainQueue) setChannel(ch *amqp.Channel) {
	dq.Lock()
	dq.channel = ch
	dq.Unlock()
}

func (dq *DomainQueue) Start() {
	dq.log.Debug(fmt.Sprintf("started"))
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
		Routing:  fmt.Sprintf(model.MQ_USER_STATUS_TEMPLATE_ROUTING_KEY, dq.Id(), userId),
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

func (dq *DomainQueue) Unbind(bind *model.BindQueueEvent) model.AppError {
	/*
		2020-03-24T01:11:42.129+0200    debug   app/web_hub.go:67       hub TODO stopped
		panic: runtime error: invalid memory address or nil pointer dereference

	*/

	ch := dq.getChannel()

	if ch == nil {
		return model.NewInternalError("mq.unbind.valid.channel", "Not found channel")
	}

	err := ch.QueueUnbind(dq.queue.Name, bind.Routing, bind.Exchange, amqp.Table{
		"x-sock-id": bind.Id,
	})

	if err != nil {
		return model.NewInternalError("mq.unbind.queue.error", err.Error())
	}

	dq.log.With(
		wlog.Int64("user_id", bind.UserId),
		wlog.String("sock_id", bind.Id),
		wlog.String("exchange", bind.Exchange),
	).Debug("unbind domain queue")

	return nil
}

func (dq *DomainQueue) BulkUnbind(b []*model.BindQueueEvent) model.AppError {
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
	ch := dq.getChannel()
	if ch == nil {
		dq.log.With(
			wlog.Int64("user_id", b.UserId),
			wlog.String("sock_id", b.Id),
			wlog.String("exchange", b.Exchange),
		).Error("not found active channel")
		return
	}

	err := ch.QueueBind(
		dq.queue.Name,
		b.Routing,
		b.Exchange,
		true,
		amqp.Table{
			"x-sock-id": b.Id,
		})
	if err != nil {
		dq.log.With(
			wlog.Int64("user_id", b.UserId),
			wlog.String("sock_id", b.Id),
			wlog.String("exchange", b.Exchange),
			wlog.String("routing", b.Routing),
		).Error(err.Error())
	} else {
		dq.log.With(
			wlog.Int64("user_id", b.UserId),
			wlog.String("sock_id", b.Id),
			wlog.String("exchange", b.Exchange),
			wlog.String("routing", b.Routing),
		).Debug("bind events")
	}
}

func (dq *DomainQueue) readMessage(m amqp.Delivery) {
	log := dq.log.With(
		wlog.String("content_type", m.ContentType),
		wlog.String("exchange", m.Exchange),
		wlog.String("routing", m.RoutingKey),
	)
	if m.ContentType != "text/json" && m.Exchange != model.ChatExchange {
		log.Warn("failed receive event content type")
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
		log.Error("not implement parser")
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
	log := dq.log.With(
		wlog.String("routing", rk),
	)
	route := strings.Split(rk, ".")
	if len(route) < 1 {
		log.Error("read app message, error: bad routing key")
	}

	switch route[0] {
	case "notification":
		e, err := parseNotification(data)
		if err != nil {
			log.Warn("failed parse json event notification, skip", wlog.String("message", string(data)))
			return
		}
		log.Debug("receive notification event", wlog.Int64("notification_id", e.Id))
		dq.notificationEvent <- e

	default:
		log.Error("read app message, error: no handler")
	}
}

func (dq *DomainQueue) readChatEvent(data []byte, rk string) {
	ev := dq.parseChatMessage(data, rk)
	if ev != nil {
		dq.chatEvents <- ev
	}
}

func (dq *DomainQueue) parseChatMessage(body []byte, rk string) *model.ChatEvent {
	log := dq.log.With(
		wlog.String("routing", rk),
		wlog.String("channel", "chat"),
	)
	rks := strings.Split(rk, ".")
	if len(rks) != 4 {
		log.Error("bad rk format")
		return nil
	}

	domainId, err := strconv.Atoi(rks[2])
	if err != nil {
		log.Error("bad domainId")
		return nil

	}

	userId, err := strconv.Atoi(rks[3])
	if err != nil {
		log.Error("bad userId")
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
		log.Error(err.Error())
		return nil
	}

	return ev
}

func (dq *DomainQueue) readAgentStatusEvent(data []byte, rk string) {
	log := dq.log.With(
		wlog.String("routing", rk),
		wlog.String("channel", "agent"),
	)
	e, err := parseAgentStatusEvent(data)
	if err != nil {
		log.Warn(err.Error())
		return
	}

	if ev, appErr := model.NewWebSocketCallCenterEvent(e); appErr != nil {
		log.Warn(err.Error())
		return
	} else {
		if len(data) < 400 {
			log.Debug("receive cc event", wlog.String("event", e.Event), wlog.String("message", string(data)))
		} else {
			log.Debug("receive cc event", wlog.String("event", e.Event), wlog.String("message", "#big data#"))
		}
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
	log := dq.log.With(
		wlog.String("routing", rk),
		wlog.String("channel", "call"),
	)
	e, err := parseCallEvent(data)
	if err != nil {
		log.Warn(err.Error())
		return
	}

	log.Debug("receive call event "+e.Event,
		wlog.String("event", e.Event),
		wlog.String("call_id", e.Id),
		wlog.String("user_id", e.UserId),
		wlog.String("domain_id", e.DomainId),
		wlog.String("app_id", e.AppId),
	)

	if e.Event != model.CallEventNameHeartbeat {
		dq.callEvents <- e
	}
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
	log := dq.log.With(
		wlog.String("routing", rk),
		wlog.String("channel", "user"),
	)
	e, err := parseUserStateEvent(data)
	if err != nil {
		log.Warn(err.Error())
		return
	}

	log.Debug("receive user status event",
		wlog.String("app_id", e.App),
		wlog.Int64("user_id", e.Id),
		wlog.String("user_status", e.Status),
	)
	dq.userStateEvents <- e

}

func (dq *DomainQueue) connect() error {
	var err error
	dq.log.Debug("trying connect...")
	dq.setChannel(nil)

	defer func() {
		if err != nil {
			dq.log.Error(err.Error())
			time.Sleep(time.Second * RECONNECT_SEC)
			go dq.connect()
		}
	}()

	var ch *amqp.Channel

	ch, err = dq.client.NewChannel()
	if err != nil {
		dq.log.Error(err.Error())
		return err
	}

	dq.setChannel(ch)

	dq.closeChannel = make(chan *amqp.Error, 1)
	ch.NotifyClose(dq.closeChannel)

	dq.queue, err = ch.QueueDeclare(
		fmt.Sprintf("engine.ws.%s.%d", model.NewId()[0:10], dq.id),
		true,
		false,
		false,
		true,
		amqp.Table{
			"x-queue-type": "quorum",
			"x-expires":    10000, // delete after 10s
		},
	)

	if err != nil {
		dq.log.Error("declare error", wlog.Err(err))
		return err
	}

	err = ch.QueueBind(dq.queue.Name,
		fmt.Sprintf("notification.%d", dq.id), model.AppExchange, true, nil)
	if err != nil {
		dq.log.Error("bind error", wlog.Err(err))
		return err
	}

	dq.delivery, err = ch.Consume(
		dq.queue.Name,
		model.NewId(),
		true,
		true,
		false,
		true,
		nil,
	)

	if err != nil {
		dq.log.Error("consume error", wlog.Err(err))
		return err
	}

	dq.log.Debug("connected")

	dq.rebindingUsers()

	go func() {
		if err := dq.Listen(); err != nil {
			dq.log.Error(err.Error())
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
	ch := dq.getChannel()
	if ch != nil {
		ch.QueueDelete(dq.queue.Name, false, false, true)
	}
}

func (dq *DomainQueue) Listen() error {
	var ok bool
	var err error
	dq.log.Debug("start listener")
	wlog.Debug(fmt.Sprintf("DomainQueue [%d] start listener", dq.Id()))
	defer dq.log.Debug("close listener")

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
	dq.log.Debug("stopping")
	close(dq.stop)
	<-dq.stopped
}

func (dq *DomainQueue) getCallEvent(data []byte) *model.CallEvent {
	e := &REvent{}
	err := json.Unmarshal(data, e)
	if err != nil {
		dq.log.Warn("failed parse json event", wlog.Err(err))
		return nil
	}

	return nil
}
