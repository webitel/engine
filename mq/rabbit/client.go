package rabbit

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/streadway/amqp"
	"github.com/webitel/engine/model"
	"github.com/webitel/engine/mq"
	"github.com/webitel/wlog"
	"os"
	"sync"
	"time"
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

var errMaxRegisterQueueSize = model.NewAppError("AMQP", "amqp.register_domain.max_queue_size", nil, "", 500)
var errMaxUnRegisterQueueSize = model.NewAppError("AMQP", "amqp.un_register_domain.max_queue_size", nil, "", 500)

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
	callEvent          chan model.CallEvent
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
		callEvent:    make(chan model.CallEvent),
		stop:         make(chan struct{}),
		stopped:      make(chan struct{}),

		registerDomainQueue:   make(chan mq.DomainQueue, MAX_QUEUE_REGISTER_SIZE),
		unRegisterDomainQueue: make(chan mq.DomainQueue, MAX_QUEUE_REGISTER_SIZE),
		nodeName:              nodeName,
	}

	return mq_
}

func (a *AMQP) NewDomainQueue(domainId int64, bindings model.GetAllBindings) (mq.DomainQueue, *model.AppError) {
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
			wlog.Info(fmt.Sprintf("success opened AMQP connection"))
		}
	}
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

func (a *AMQP) SendJSON(key string, data []byte) *model.AppError {

	return nil
}
