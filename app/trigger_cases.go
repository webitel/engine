package app

import (
	workflow "buf.build/gen/go/webitel/workflow/protocolbuffers/go"
	"context"
	"encoding/json"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/webitel/engine/model"
	"github.com/webitel/engine/store"
	"github.com/webitel/engine/utils"
	flow "github.com/webitel/flow_manager/client"
	"github.com/webitel/wlog"
	"strings"
	"sync/atomic"
	"time"
)

const (
	AMQPMaxAttemptsConnect = 100
	FlowMaxAttemptsToStart = 5
	FlowMinDuration        = 100 * time.Millisecond
	FlowMaxDuration        = 1 * time.Second
)

type TriggersByExpression map[string][]*model.TriggerWithDomainID

type messageCase struct {
	Case     string `json:"case"`
	DomainId int64  `json:"domain_id"`
}

type TriggerCase interface {
	Start() error
	Stop()
	NotifyUpdateTrigger()
}

type TriggerCaseMQ struct {
	log           *wlog.Logger
	config        *model.CaseTriggersSettings
	triggers      atomic.Value
	stopChan      chan struct{}
	stopQueueChan chan struct{}
	reloadChan    chan struct{}
	errorChan     chan *amqp.Error
	connection    *amqp.Connection
	channel       *amqp.Channel
	store         store.Store
	Queue         amqp.Queue
	flowManager   flow.FlowManager
}

func (ct *TriggerCaseMQ) loadTriggersByExpression() TriggersByExpression {
	return ct.triggers.Load().(TriggersByExpression)
}

func (ct *TriggerCaseMQ) storeTriggersByExpression(triggers TriggersByExpression) {
	ct.triggers.Store(triggers)
}

func (ct *TriggerCaseMQ) Start() error {
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

func (ct *TriggerCaseMQ) NotifyUpdateTrigger() {
	if ct == nil {
		return
	}
	go func() { ct.reloadChan <- struct{}{} }()
}

func (ct *TriggerCaseMQ) reloadTriggers() {
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
				ct.log.Error(fmt.Sprintf("Could not reload triggers: %s", err.Error()))
			}
		}
	}
}

func (ct *TriggerCaseMQ) Stop() {
	if ct == nil {
		return
	}
	ct.log.Debug("Trying to stopping TriggerCaseMQ")
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

func (ct *TriggerCaseMQ) listen() error {
	messages, err := ct.channel.Consume(ct.Queue.Name, "", true, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("could not consume messages from %s: %w", ct.Queue.Name, err)
	}

	go ct.processedMessages(messages)
	return nil
}

func (ct *TriggerCaseMQ) processedMessages(messages <-chan amqp.Delivery) {
	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()
	for {
		select {
		case <-ct.stopChan:
			close(ct.stopQueueChan)
			return
		case amqpErr, ok := <-ct.errorChan:
			if !ok {
				ct.log.Info("Closed rabbit error channel")
				return
			}
			ct.log.Error(fmt.Sprintf("amqp connection error: %s", amqpErr.Error()))
			err := ct.initConnection()
			if err != nil {
				// TODO reconnect
				ct.log.Error(fmt.Sprintf("Could not reconnect ro amqp: %s", err.Error()))
				return
			}

			err = ct.listen()
			if err != nil {
				ct.log.Error(fmt.Sprintf("Could not start listen messages: %s", err.Error()))
			}
			return

		case msg := <-messages:
			ct.log.Debug(fmt.Sprintf("Received a message: %s; by routiong key: %s", string(msg.Body), msg.RoutingKey))
			expression := ct.getExpressionByRoutingKey(msg.RoutingKey)

			triggers := ct.loadTriggersByExpression()[expression]
			if len(triggers) == 0 {
				ct.log.Debug(fmt.Sprintf("No trigger found for key %s and expression %s", msg.RoutingKey, expression))
				continue
			}
			message := &messageCase{}
			err := json.Unmarshal(msg.Body, message)
			if err != nil {
				ct.log.Error(fmt.Sprintf("Could not unmarshal message  %s: %s", msg.Body, err.Error()))
				continue
			}

			for _, trigger := range triggers {
				if trigger.DomainId != message.DomainId {
					ct.log.Debug(fmt.Sprintf("Skipping trigger %d because domain ID does not match: %d, %dd", trigger.Id, message.DomainId, trigger.DomainId))
					continue
				}
				variables := trigger.Variables
				if variables == nil {
					variables = make(model.StringMap, 1)
				}
				variables["case"] = message.Case

				request := workflow.StartFlowRequest{DomainId: trigger.DomainId, SchemaId: uint32(trigger.Schema.Id), Variables: variables}
				go func(r *workflow.StartFlowRequest) {
					id, err := ct.startFlowRequestWithContext(ctx, r)
					if err != nil {
						ct.log.Error(fmt.Sprintf("Could not start flow request: %s: %s", r, err.Error()))
						return
					}
					ct.log.Info(fmt.Sprintf("Started flow for with id : %s", id))
				}(&request)
			}
		}
	}
}

func (ct *TriggerCaseMQ) startFlowRequestWithContext(context context.Context, request *workflow.StartFlowRequest) (id string, err error) {
	var backoff = utils.NewBackoff(FlowMinDuration, FlowMaxDuration, 2, false)
	for {
		select {
		case <-context.Done():
			return "", context.Err()
		default:
			id, err = ct.flowManager.Queue().StartFlow(request)
			if err != nil {
				if backoff.Attempt() > FlowMaxAttemptsToStart {
					return "", err
				}
				ct.log.Debug(fmt.Sprintf("Retrying to start flow with request %s: %s. Attempt: %d", request, err.Error(), backoff.Attempt()))
				time.Sleep(backoff.Duration())
				continue
			}
			return id, err
		}
	}
}

func (ct *TriggerCaseMQ) init() error {
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

func (ct *TriggerCaseMQ) loadTriggers() error {
	ctx := context.Background()
	triggerSlice, err := ct.store.Trigger().GetAllByType(ctx, model.TriggerTypeCase)
	if err != nil {
		return fmt.Errorf("could not load triggers: %v", err)
	}
	triggersMap := make(TriggersByExpression, 3)
	for _, trigger := range triggerSlice {
		triggersMap[trigger.Expression] = append(triggersMap[trigger.Expression], trigger)
	}
	ct.log.Debug(fmt.Sprintf("Loaded %d triggers: %+v", len(triggersMap), triggersMap))
	ct.storeTriggersByExpression(triggersMap)
	return nil
}

func (ct *TriggerCaseMQ) initConnection() error {
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

func (ct *TriggerCaseMQ) initQueue() error {
	var err error
	ct.Queue, err = ct.channel.QueueDeclare(ct.config.Queue, true, false, false, false, map[string]interface{}{
		"x-message-ttl": ct.config.QueueMessagesTTL,
	})
	if err != nil {
		return fmt.Errorf("could not create queue %s: %w", ct.config.Queue, err)
	}

	return nil
}

func (ct *TriggerCaseMQ) getExpressionByRoutingKey(routingKey string) string {
	topic := strings.Replace(ct.config.Topic, "*", "", 1)
	return strings.Replace(routingKey, topic, "", 1)
}

func NewTriggerCases(log *wlog.Logger, store store.Store, flow flow.FlowManager, cfg *model.CaseTriggersSettings) *TriggerCaseMQ {
	return &TriggerCaseMQ{
		log:           log,
		store:         store,
		flowManager:   flow,
		config:        cfg,
		reloadChan:    make(chan struct{}, 32),
		stopChan:      make(chan struct{}),
		stopQueueChan: make(chan struct{})}
}
