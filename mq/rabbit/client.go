package rabbit

import (
	"context"
	"encoding/json"
	stderrors "errors"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/pkg/errors"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/webitel/engine/model"
	"github.com/webitel/engine/mq"
	"github.com/webitel/wlog"
)

const (
	MAX_ATTEMPTS_CONNECT = 100
	RECONNECT_SEC        = 5

	MAX_QUEUE_REGISTER_SIZE = 10000
)

const (
	EXIT_DECLARE_EXCHANGE = 110
	EXIT_DECLARE_QUEUE    = 111
	EXIT_BIND             = 112
)

const (
	callServiceHangupData = `{"hangup_by":"service","cause":"SYSTEM_SHUTDOWN","sip":501}`
)

// Stdlib errors: this file's `errors` is pkg/errors, which attaches stack traces.
var (
	errConnectionClosed = stderrors.New("amqp: connection is closed")
	errChannelClosed    = stderrors.New("amqp: channel is closed")
)

var errMaxRegisterQueueSize = model.NewInternalError("amqp.register_domain.max_queue_size", "")
var errMaxUnRegisterQueueSize = model.NewInternalError("amqp.un_register_domain.max_queue_size", "")

type AMQP struct {
	connection         *amqp.Connection
	channel            *amqp.Channel
	settings           *model.MessageQueueSettings
	queueForCall       amqp.Queue
	nodeName           string
	connectionAttempts int
	errorChan          chan *amqp.Error
	stop               chan struct{}
	stopped            chan struct{}
	domainQueues       map[int64]mq.DomainQueue

	registerDomainQueue   chan mq.DomainQueue
	unRegisterDomainQueue chan mq.DomainQueue
	log                   *wlog.Logger
	domainEventHandler    mq.DomainEventHandler

	mx sync.Mutex
}

func NewRabbitMQ(nodeName string, settings *model.MessageQueueSettings) mq.LayeredMQLayer {
	mq_ := &AMQP{
		settings:     settings,
		errorChan:    make(chan *amqp.Error, 1),
		domainQueues: make(map[int64]mq.DomainQueue),
		stop:         make(chan struct{}),
		stopped:      make(chan struct{}),

		registerDomainQueue:   make(chan mq.DomainQueue, MAX_QUEUE_REGISTER_SIZE),
		unRegisterDomainQueue: make(chan mq.DomainQueue, MAX_QUEUE_REGISTER_SIZE),
		nodeName:              nodeName,
		log: wlog.GlobalLogger().With(
			wlog.Namespace("context"),
			wlog.String("protocol", "amqp"),
		),
	}

	return mq_
}

func (a *AMQP) NewDomainQueue(domainId int64, bindings model.GetAllBindings) (mq.DomainQueue, model.AppError) {
	if len(a.registerDomainQueue) > MAX_QUEUE_REGISTER_SIZE {
		return nil, errMaxRegisterQueueSize
	}
	q := newDomainQueue(a, domainId, bindings)
	a.registerDomainQueue <- q
	return q, nil
}

func (a *AMQP) Start() {
	a.initConnection()

	if err := a.initDomains(context.Background()); err != nil {
		a.log.Critical("initializing domains consumer", wlog.Err(err))
		panic(err)
	}

	go a.Listen()
}

// Ping reads the cached connection's state; it does not dial.
func (a *AMQP) Ping(context.Context) error {
	a.mx.Lock()
	defer a.mx.Unlock()

	if a.connection == nil || a.connection.IsClosed() {
		return errConnectionClosed
	}

	if a.channel == nil || a.channel.IsClosed() {
		return errChannelClosed
	}

	return nil
}

func (a *AMQP) BrokerChannelProvider() ChannelProvider {
	return func() (*amqp.Channel, error) {
		a.mx.Lock()
		conn := a.connection
		a.mx.Unlock()

		if conn == nil || conn.IsClosed() {
			return nil, model.NewCustomCodeError(
				"rabbit.client.broker_channel_provider",
				"connection closed",
				412,
			)
		}

		return conn.Channel()
	}
}

func (a *AMQP) addDomainQueue(id int64, q mq.DomainQueue) {
	a.mx.Lock()
	defer a.mx.Unlock()

	a.domainQueues[id] = q
	a.log.Debug("added domain queue", wlog.Int64("domain_id", id))
}

func (a *AMQP) RemoveDomainQueue(q *DomainQueue) {
	a.mx.Lock()
	defer a.mx.Unlock()

	delete(a.domainQueues, q.Id())
	a.log.Debug("remove domain queue", wlog.Int64("domain_id", q.Id()))
}

func (a *AMQP) initConnection() {
	var err error

	if a.connectionAttempts >= MAX_ATTEMPTS_CONNECT {
		a.log.Critical(fmt.Sprintf("failed to open AMQP connection..."))
		time.Sleep(time.Second)
		os.Exit(1)
	}
	a.connectionAttempts++
	a.connection, err = amqp.Dial(a.settings.Url)
	if err != nil {
		a.log.Critical(fmt.Sprintf("failed to open AMQP connection to err:%v", err.Error()))
		time.Sleep(time.Second * RECONNECT_SEC)
		a.initConnection()
	} else {
		a.connectionAttempts = 0
		a.channel, err = a.connection.Channel()
		a.errorChan = make(chan *amqp.Error, 1)

		a.channel.NotifyClose(a.errorChan)
		if err != nil {
			a.log.Critical(fmt.Sprintf("failed to open AMQP channel to err:%v", err.Error()))
			time.Sleep(time.Second)
			os.Exit(1)
		} else {
			if err := a.createAppExchange(); err != nil {
				panic(err.Error())
			}
			a.log.Info("success opened AMQP connection")
		}
	}
}

func (a *AMQP) SendNotification(domainId int64, event *model.Notification) model.AppError {
	err := a.channel.Publish(model.AppExchange, fmt.Sprintf("notification.%d", domainId), false, false, amqp.Publishing{
		ContentType: "text/json",
		Body:        []byte(event.ToJson()),
	})
	if err != nil {
		return model.NewInternalError("amqp.notification.publish.app_error", err.Error())
	}
	return nil
}

func (a *AMQP) RegisterWebsocket(domainId int64, event *model.RegisterToWebsocketEvent) model.AppError {
	err := a.channel.Publish(model.AppExchange, fmt.Sprintf("event.open_socket.%d.%d", domainId, event.UserId), false, false, amqp.Publishing{
		ContentType: "text/json",
		Body:        []byte(event.ToJson()),
	})
	if err != nil {
		return model.NewInternalError("amqp.register_socket.publish.app_error", err.Error())
	}
	return nil
}

func (a *AMQP) SendStickingCall(e *model.CallServiceHangup) model.AppError {
	// fixme CC
	e.Subclass = "Event-Subclass"
	e.Event = "hangup"
	e.Data = callServiceHangupData

	err := a.channel.Publish(model.CallExchange, fmt.Sprintf("events.hangup.%s.%s.%s", e.CCAppId, e.DomainId, e.UserId), false, false, amqp.Publishing{
		ContentType: "text/json",
		Body:        e.MarshalJSON(),
	})

	if err != nil {
		return model.NewInternalError("amqp.publish.sticking_call.app_error", err.Error())
	}

	return nil
}

func (a *AMQP) UnRegisterWebsocket(domainId int64, event *model.RegisterToWebsocketEvent) model.AppError {
	err := a.channel.Publish(model.AppExchange, fmt.Sprintf("event.close_socket.%d.%d", domainId, event.UserId), false, false, amqp.Publishing{
		ContentType: "text/json",
		Body:        []byte(event.ToJson()),
	})
	if err != nil {
		return model.NewInternalError("amqp.unregister_socket.publish.app_error", err.Error())
	}
	return nil
}

func (a *AMQP) createAppExchange() model.AppError {
	if err := a.channel.ExchangeDeclare(model.AppExchange, "topic", true, false, false, true, nil); err != nil {
		return model.NewInternalError("amqp.declare.exchange.app_err", err.Error())
	}
	if err := a.channel.ExchangeDeclare(
		model.EventExchange,
		model.MQ_TOPIC,
		true,
		false,
		false,
		true,
		nil,
	); err != nil {
		return model.NewInternalError("amqp.declare.event_exchange.app_err", err.Error())
	}

	if err := a.channel.ExchangeDeclare(
		model.CallExchange,
		model.MQ_TOPIC,
		true,
		false,
		false,
		true,
		nil,
	); err != nil {
		return model.NewInternalError("amqp.declare.case_exchange.app_err", err.Error())
	}

	return nil
}

func (a *AMQP) Channel() *amqp.Channel {
	return a.channel
}

func (a *AMQP) NewChannel() (*amqp.Channel, error) {
	if a.connection == nil || a.connection.IsClosed() {
		return nil, errors.New("connection closed")
	}
	return a.connection.Channel()
}

func (a *AMQP) Send(ctx context.Context, exchange string, rk string, body []byte) error {
	if a.channel == nil || a.connection == nil || a.connection.IsClosed() {
		return errors.New("connection closed")
	}

	return a.channel.Publish(exchange, rk, true, false, amqp.Publishing{
		ContentType: "text/json",
		Body:        body,
	})
}

func (a *AMQP) Close() {
	a.log.Debug("AMQP receive stop client")
	close(a.stop)
	<-a.stopped

	if a.channel != nil {
		a.channel.Close()
		a.log.Debug("close AMQP channel")
	}

	if a.connection != nil {
		a.connection.Close()
		a.log.Debug("close AMQP connection")
	}
}

func (a *AMQP) SendJSON(key string, data []byte) model.AppError {

	return nil
}

func (a *AMQP) SendStartFlow(ctx context.Context, domainId int64, schemaId int32, in interface{}) model.AppError {
	exe := model.ExecFlow{
		DomainId:  domainId,
		SchemaId:  schemaId,
		Variables: model.StructToVariable(in),
	}

	body, err := json.Marshal(exe)
	if err != nil {
		return model.NewInternalError("amqp.start_flow.parse", err.Error())
	}

	err = a.channel.Publish("flow", "exec", true, false, amqp.Publishing{
		ContentType: "text/json",
		Body:        body,
	})

	if err != nil {
		return model.NewInternalError("amqp.start_flow.publish", err.Error())
	}

	return nil
}

func (a *AMQP) initDomains(ctx context.Context) error {
	queueCfg := NewQueueConfig(
		"engine.domains.consumer",
		WithQueueDurable(true),
		WithQueueArg("x-queue-type", "quorum"),
	)

	bq := NewBrokerQueue(
		queueCfg,
		a.BrokerChannelProvider(),
		a.processDomainDeliveries,
		WithConsumerTag(
			fmt.Sprintf("engine-%s-domains", a.nodeName),
		),
		WithConcurrency(2),
	)

	bq.Bind(NewQueueBindConfig(
		queueCfg.Name,
		"domains.create.*",
		"webitel",
		WithQueueBindNoWait(false),
	))

	go func() {
		if err := bq.Run(ctx); err != nil {
			a.log.Error("broker queue stopped", wlog.Err(err))
		}
	}()

	go func() {
		<-a.stopped

		bq.Close()
	}()

	return nil
}

func (a *AMQP) processDomainDeliveries(ctx context.Context, d amqp.Delivery) error {
	de, err := model.NewDomainEventFromRoutingKey(d.RoutingKey)
	if err != nil {
		return err
	}

	a.mx.Lock()
	h := a.domainEventHandler
	a.mx.Unlock()

	if h == nil {
		a.log.Warn("no domain event handler set, dropping event")
		return nil
	}

	return h(ctx, de)
}

func (a *AMQP) SetDomainsEventHandler(h mq.DomainEventHandler) {
	a.mx.Lock()
	a.domainEventHandler = h
	a.mx.Unlock()
}
