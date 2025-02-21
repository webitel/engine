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
	"sync/atomic"
	"time"
)

const (
	AMQPMaxAttemptsConnect = 100
)

type Backoff interface {
	Duration() time.Duration
	Attempt() uint64
	Reset()
}

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
	log                 *wlog.Logger
	config              *model.CaseTriggersSettings
	triggers            atomic.Value
	stopChan            chan struct{}
	stopCreateQueueChan chan struct{}
	stopUpdateQueueChan chan struct{}
	stopDeleteQueueChan chan struct{}
	reloadChan          chan struct{}
	errorChan           chan *amqp.Error
	connection          *amqp.Connection
	channel             *amqp.Channel
	store               store.Store
	backoff             Backoff
	createQueue         amqp.Queue
	updateQueue         amqp.Queue
	deleteQueue         amqp.Queue
	flowManager         flow.FlowManager
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
	<-ct.stopCreateQueueChan
	<-ct.stopUpdateQueueChan
	<-ct.stopDeleteQueueChan

	if ct.channel != nil {
		_ = ct.channel.Close()
	}

	if ct.connection != nil {
		_ = ct.connection.Close()
	}
}

func (ct *TriggerCaseMQ) listen() error {
	createMessages, err := ct.channel.Consume(ct.createQueue.Name, "", true, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("could not consume messages from %s: %w", ct.createQueue.Name, err)
	}

	updateMessages, err := ct.channel.Consume(ct.updateQueue.Name, "", true, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("could not consume messages from %s: %w", ct.updateQueue.Name, err)
	}

	deleteMessages, err := ct.channel.Consume(ct.deleteQueue.Name, "", true, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("could not consume messages from %s: %w", ct.deleteQueue.Name, err)
	}
	go ct.processedMessages(createMessages, ct.stopCreateQueueChan, "create")
	go ct.processedMessages(updateMessages, ct.stopUpdateQueueChan, "update")
	go ct.processedMessages(deleteMessages, ct.stopDeleteQueueChan, "delete")
	return nil
}

func (ct *TriggerCaseMQ) processedMessages(messages <-chan amqp.Delivery, stopChan chan struct{}, expression string) {
	for {
		select {
		case <-ct.stopChan:
			close(stopChan)
			return
		case msg := <-messages:
			ct.log.Debug(fmt.Sprintf("Received a message: %s", string(msg.Body)))
			triggers := ct.loadTriggersByExpression()[expression]
			if len(triggers) == 0 {
				continue
			}
			message := &messageCase{}
			err := json.Unmarshal(msg.Body, message)
			if err != nil {
				ct.log.Error(fmt.Sprintf("Could not unmarshal message  %s: %s", msg.Body, err.Error()))
				err = msg.Nack(false, false) // drop message
				if err != nil {
					ct.log.Error(fmt.Sprintf("Could not NACK message  %s: %s", msg.Body, err.Error()))
				}
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
				id, err := ct.flowManager.Queue().StartFlow(&request)
				if err != nil {
					ct.log.Error(fmt.Sprintf("Could not start flow: %s", err.Error()))
					err = msg.Nack(false, true)
					if err != nil {
						ct.log.Error(fmt.Sprintf("Could not nack message: %s", err.Error()))
					}
					continue
				}
				ct.log.Info(fmt.Sprintf("Started flow with id %s", id))
			}
		}
	}
}

func (ct *TriggerCaseMQ) init() error {
	if err := ct.initConnection(); err != nil {
		return err
	}

	if err := ct.initExchangeQueues(); err != nil {
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
	ct.storeTriggersByExpression(triggersMap)
	return nil
}

func (ct *TriggerCaseMQ) initConnection() error {
	var err error

	for {
		ct.connection, err = amqp.Dial(ct.config.BrokerUrl)
		if err != nil {
			if ct.backoff.Attempt() > AMQPMaxAttemptsConnect {
				return fmt.Errorf("failed to open AMQP connection for %s: %w", ct.config.BrokerUrl, err)
			}
			time.Sleep(ct.backoff.Duration())
			continue
		}
		break
	}
	ct.backoff.Reset()
	ct.channel, err = ct.connection.Channel()
	if err != nil {
		return fmt.Errorf("failed to open channel for %s: %w", ct.config.BrokerUrl, err)
	}
	ct.errorChan = make(chan *amqp.Error, 1)
	ct.channel.NotifyClose(ct.errorChan)

	return nil
}

func (ct *TriggerCaseMQ) initExchangeQueues() error {
	var err error
	// create queues
	ct.createQueue, err = ct.channel.QueueDeclare(ct.config.CreateQueue, true, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("could not create Create queue %s: %w", ct.config.CreateQueue, err)
	}

	ct.updateQueue, err = ct.channel.QueueDeclare(ct.config.UpdateQueue, true, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("could not create Update queue %s: %w", ct.config.UpdateQueue, err)
	}

	ct.deleteQueue, err = ct.channel.QueueDeclare(ct.config.DeleteQueue, true, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("could not create Delete queue %s: %w", ct.config.CreateQueue, err)
	}
	return nil
}

func NewTriggerCases(log *wlog.Logger, store store.Store, flow flow.FlowManager, cfg *model.CaseTriggersSettings) *TriggerCaseMQ {
	return &TriggerCaseMQ{
		log:                 log,
		store:               store,
		flowManager:         flow,
		config:              cfg,
		reloadChan:          make(chan struct{}, 32),
		stopChan:            make(chan struct{}),
		stopCreateQueueChan: make(chan struct{}),
		stopUpdateQueueChan: make(chan struct{}),
		stopDeleteQueueChan: make(chan struct{}),
		backoff:             utils.NewDefaultBackoff()}
}
