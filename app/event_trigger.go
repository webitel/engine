package app

import (
	"context"
	"encoding/json"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/webitel/engine/app/flow"
	"github.com/webitel/engine/model"
	"github.com/webitel/engine/store"
	"github.com/webitel/engine/utils"
	"github.com/webitel/wlog"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

const (
	AMQPMaxAttemptsConnect = 100
)

type TriggersByExpression map[string][]*model.TriggerWithDomainID

type EventMessage map[string]json.RawMessage

type EventTrigger interface {
	Start() error
	Stop()
	NotifyUpdateTrigger()
}

type TriggerEventMQ struct {
	log           *wlog.Logger
	config        *model.TriggersSettings
	triggers      atomic.Value
	stopChan      chan struct{}
	stopQueueChan chan struct{}
	reloadChan    chan struct{}
	errorChan     chan *amqp.Error
	connection    *amqp.Connection
	channel       *amqp.Channel
	store         store.Store
	Queue         amqp.Queue
	Exchange      string
	flowManager   flow.FlowManager
}

func NewEventTrigger(log *wlog.Logger, store store.Store, flow flow.FlowManager, cfg *model.TriggersSettings) *TriggerEventMQ {
	return &TriggerEventMQ{
		log: log.With(wlog.Namespace("context"),
			wlog.String("scope", "event-trigger"),
		),
		store:         store,
		flowManager:   flow,
		config:        cfg,
		reloadChan:    make(chan struct{}, 16),
		stopChan:      make(chan struct{}),
		stopQueueChan: make(chan struct{}),
	}
}

func (ct *TriggerEventMQ) loadTriggersByExpression() TriggersByExpression {
	return ct.triggers.Load().(TriggersByExpression)
}

func (ct *TriggerEventMQ) storeTriggersByExpression(triggers TriggersByExpression) {
	ct.triggers.Store(triggers)
}

func (ct *TriggerEventMQ) Start() error {
	if ct == nil {
		return nil
	}
	err := ct.init()
	if err != nil {
		return err
	}

	if err := ct.listen(); err != nil {
		return err
	}

	go ct.reloadTriggers()

	return nil

}

func (ct *TriggerEventMQ) NotifyUpdateTrigger() {
	if ct == nil {
		return
	}
	go func() { ct.reloadChan <- struct{}{} }()
}

func (ct *TriggerEventMQ) reloadTriggers() {
	for {
		select {
		case <-ct.stopChan:
			return

		case _, ok := <-ct.reloadChan:
			if !ok {
				return
			}
			err := ct.loadTriggers()
			if err != nil {
				ct.log.Error(fmt.Sprintf("could not reload triggers: %s", err.Error()))
			}
		}
	}
}

func (ct *TriggerEventMQ) Stop() {
	if ct == nil {
		return
	}
	ct.log.Debug("trying to stopping TriggerCaseMQ")
	close(ct.stopChan)
	close(ct.reloadChan)
	<-ct.stopQueueChan

	if ct.channel != nil {
		_ = ct.channel.Close()
	}

	if ct.connection != nil {
		_ = ct.connection.Close()
	}
}

func (ct *TriggerEventMQ) listen() error {
	messages, err := ct.channel.Consume(ct.Queue.Name, "", true, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("could not consume messages from %s: %w", ct.Queue.Name, err)
	}

	go ct.processedMessages(messages)
	return nil
}

type triggerRequest struct {
	name      string
	triggerId int32
	variables model.StringMap
	schemaId  int
	domainId  int64
}

func (ct *TriggerEventMQ) getFlowRequests(domainId int64, event string) []triggerRequest {
	triggers := ct.loadTriggersByExpression()[event]
	if len(triggers) == 0 {
		return nil
	}

	res := make([]triggerRequest, 0, 3)

	for _, tr := range triggers {
		if tr.DomainId == domainId {
			res = append(res, triggerRequest{
				name:      tr.Name,
				triggerId: tr.Id,
				schemaId:  tr.Schema.Id,
				domainId:  domainId,
				variables: make(model.StringMap),
			})
		}
	}

	return res
}

func (ct *TriggerEventMQ) processedMessages(messages <-chan amqp.Delivery) {
	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()
	reconnect := false
	defer func() {
		if !reconnect {
			return
		}
		reconnect = false

		err := ct.initConnection()
		if err != nil {
			// TODO fatal
			ct.log.Error(fmt.Sprintf("could not reconnect ro amqp: %s", err.Error()))
			return
		}

		err = ct.listen()
		if err != nil {
			ct.log.Error(fmt.Sprintf("could not start listen messages: %s", err.Error()))
		}
	}()

	for {
		select {
		case <-ct.stopChan:
			close(ct.stopQueueChan)
			return
		case amqpErr, ok := <-ct.errorChan:
			if !ok {
				ct.log.Info("closed rabbit error channel")
				return
			}
			ct.log.Error(fmt.Sprintf("amqp connection error: %s", amqpErr.Error()))
			reconnect = true

		case msg, ok := <-messages:
			if !ok {
				return
			}

			object, event, domainId := ct.getExpressionByRoutingKey(msg.RoutingKey)

			// TODO
			requests := ct.getFlowRequests(domainId, triggerHash(object, event))

			if len(requests) == 0 {
				ct.log.Debug(fmt.Sprintf("no trigger found for key %s and expression %s", msg.RoutingKey, event))
				continue
			}

			message := EventMessage{}
			err := json.Unmarshal(msg.Body, &message)
			if err != nil {
				ct.log.Error(fmt.Sprintf("could not unmarshal message  %s: %s", string(msg.Body), err.Error()))
				continue
			}

			for _, rs := range requests {
				for k, v := range message {
					rs.variables[k] = string(v)
				}
				rs.variables["action"] = event
				job, err := ct.store.Trigger().CreateJob(ctx, rs.domainId, rs.triggerId, rs.variables)
				if err != nil {
					ct.log.Error(fmt.Sprintf("could not create job: %v: %s", rs, err.Error()))
					return
				}
				ct.log.Info(fmt.Sprintf("started trigger \"%s\" job_id : %d", rs.name, job.Id))
			}
		}
	}
}

func (ct *TriggerEventMQ) init() error {
	if err := ct.initConnection(); err != nil {
		return err
	}

	if err := ct.initQueue(); err != nil {
		return err
	}

	if err := ct.loadTriggers(); err != nil {
		return err
	}

	return nil
}

func (ct *TriggerEventMQ) loadTriggers() error {
	ctx := context.Background()
	triggerSlice, err := ct.store.Trigger().GetAllByType(ctx, model.TriggerTypeEvent)
	if err != nil {
		return fmt.Errorf("could not load triggers: %v", err)
	}
	triggersMap := make(TriggersByExpression, 4)
	for _, trigger := range triggerSlice {
		ex := triggerHash(trigger.Object, trigger.Event)
		triggersMap[ex] = append(triggersMap[ex], trigger)
	}
	ct.log.Debug(fmt.Sprintf("loaded %d triggers: %+v", len(triggersMap), triggersMap))
	ct.storeTriggersByExpression(triggersMap)
	return nil
}

func triggerHash(object string, event string) string {
	return object + "." + event
}

func (ct *TriggerEventMQ) initConnection() error {
	var err error
	var backoff = utils.NewDefaultBackoff()
	for {
		ct.connection, err = amqp.Dial(ct.config.BrokerUrl)
		if err != nil {
			if backoff.Attempt() > AMQPMaxAttemptsConnect {
				return fmt.Errorf("failed to open AMQP connection for %s: %w", ct.config.BrokerUrl, err)
			}
			time.Sleep(backoff.Duration())
			continue
		}
		break
	}
	ct.channel, err = ct.connection.Channel()
	if err != nil {
		return fmt.Errorf("failed to open channel for %s: %w", ct.config.BrokerUrl, err)
	}
	ct.errorChan = make(chan *amqp.Error, 1)
	ct.channel.NotifyClose(ct.errorChan)

	return nil
}

func (ct *TriggerEventMQ) initQueue() error {
	var err error

	err = ct.channel.ExchangeDeclare(ct.config.Exchange, "topic", true, false, false, true, nil)
	if err != nil {
		return err
	}

	ct.Queue, err = ct.channel.QueueDeclare(ct.config.Queue,
		true,
		false,
		false,
		false, amqp.Table{"x-queue-type": "quorum"},
	)
	if err != nil {
		return fmt.Errorf("could not create queue %s: %w", ct.config.Queue, err)
	}

	err = ct.channel.QueueBind(ct.Queue.Name, "#", ct.config.Exchange, true, nil)
	if err != nil {
		return err
	}

	return nil
}

func (ct *TriggerEventMQ) getExpressionByRoutingKey(routingKey string) (object string, expression string, domainId int64) {
	s := strings.Split(routingKey, ".")
	if len(s) < 3 {
		return "", "", 0
	}
	object = s[1]
	expression = s[2]
	domainId, _ = strconv.ParseInt(s[3], 10, 64)
	return
}
