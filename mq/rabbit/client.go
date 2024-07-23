package rabbit

import (
	"context"
	"encoding/json"
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
	go a.Listen()
}

func (a *AMQP) addDomainQueue(id int64, q mq.DomainQueue) {
	a.mx.Lock()
	defer a.mx.Unlock()

	a.domainQueues[id] = q
	wlog.Debug(fmt.Sprintf("added domain queue[%d]", id))
}

func (a *AMQP) RemoveDomainQueue(q *DomainQueue) {
	a.mx.Lock()
	defer a.mx.Unlock()

	delete(a.domainQueues, q.Id())
	wlog.Debug(fmt.Sprintf("remove domain queue[%d] %s", q.Id(), q.Name()))
}

func (a *AMQP) initConnection() {
	var err error

	if a.connectionAttempts >= MAX_ATTEMPTS_CONNECT {
		wlog.Critical(fmt.Sprintf("failed to open AMQP connection..."))
		time.Sleep(time.Second)
		os.Exit(1)
	}
	a.connectionAttempts++
	a.connection, err = amqp.Dial(a.settings.Url)
	if err != nil {
		wlog.Critical(fmt.Sprintf("failed to open AMQP connection to err:%v", err.Error()))
		time.Sleep(time.Second * RECONNECT_SEC)
		a.initConnection()
	} else {
		a.connectionAttempts = 0
		a.channel, err = a.connection.Channel()
		a.errorChan = make(chan *amqp.Error, 1)

		a.channel.NotifyClose(a.errorChan)
		if err != nil {
			wlog.Critical(fmt.Sprintf("failed to open AMQP channel to err:%v", err.Error()))
			time.Sleep(time.Second)
			os.Exit(1)
		} else {
			if err := a.createAppExchange(); err != nil {
				panic(err.Error())
			}
			wlog.Info(fmt.Sprintf("success opened AMQP connection"))
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
		false,
		nil,
	); err != nil {
		return model.NewInternalError("amqp.declare.event_exchange.app_err", err.Error())
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

func (a *AMQP) init() {
	err := a.DeclareExchanges()
	if err != nil {
		wlog.Critical(fmt.Sprintf("failed to declare exchanges: %v", err.Error()))
		time.Sleep(time.Second)
		os.Exit(1)
	}

	//err = a.DeclareQueues()
	//if err != nil {
	//	wlog.Critical(fmt.Sprintf("failed to declare queues: %v", err.Error()))
	//	time.Sleep(time.Second)
	//	os.Exit(1)
	//}
}

func (a *AMQP) Close() {
	wlog.Debug("AMQP receive stop client")
	close(a.stop)
	<-a.stopped

	if a.channel != nil {
		a.channel.Close()
		wlog.Debug("close AMQP channel")
	}

	if a.connection != nil {
		a.connection.Close()
		wlog.Debug("close AMQP connection")
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
