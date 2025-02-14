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
		case _, ok := <-ct.reloadChan:
			if !ok {
				return
			}
			err := ct.loadTriggers()
			if err != nil {
				ct.log.Error(fmt.Sprintf("Could not reload triggers: %s", err.Error()))
			}
		default:
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
	createMessages, err := ct.channel.Consume(ct.createQueue.Name, "", false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("could not consume messages from %s: %w", ct.createQueue.Name, err)
	}

	updateMessages, err := ct.channel.Consume(ct.updateQueue.Name, "", false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("could not consume messages from %s: %w", ct.updateQueue.Name, err)
	}

	deleteMessages, err := ct.channel.Consume(ct.deleteQueue.Name, "", false, false, false, false, nil)
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
			triggers := ct.loadTriggersByExpression()[expression]
			if len(triggers) == 0 {
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
					ct.log.Debug(fmt.Sprintf("Skipping trigger %s because domain ID does not match: %d, %dd", message.DomainId, trigger.DomainId))
					continue
				}
				variables := trigger.Variables
				if variables == nil {
					variables = make(model.StringMap, 1)
				}
				variables["$case"] = string(msg.Body)

				request := workflow.StartFlowRequest{DomainId: trigger.DomainId, SchemaId: uint32(trigger.Schema.Id), Variables: variables}
				id, err := ct.flowManager.Queue().StartFlow(&request)
				if err != nil {
					ct.log.Error(fmt.Sprintf("Could not start flow: %s", err.Error()))
					continue
				}
				ct.log.Info(fmt.Sprintf("Started flow with id %s", id))
				err = msg.Ack(false)
				if err != nil {
					ct.log.Error(fmt.Sprintf("Could not ack message: %s", err.Error()))
				}
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
		ct.connection, err = amqp.Dial(ct.config.AMQPUrl)
		if err != nil {
			if ct.backoff.Attempt() > AMQPMaxAttemptsConnect {
				return fmt.Errorf("failed to open AMQP connection for %s: %w", ct.config.AMQPUrl, err)
			}
			time.Sleep(ct.backoff.Duration())
			continue
		}
		break
	}
	ct.backoff.Reset()
	ct.channel, err = ct.connection.Channel()
	if err != nil {
		return fmt.Errorf("failed to open channel for %s: %w", ct.config.AMQPUrl, err)
	}
	ct.errorChan = make(chan *amqp.Error, 1)
	ct.channel.NotifyClose(ct.errorChan)

	return nil
}

func (ct *TriggerCaseMQ) initExchangeQueues() error {
	// create exchanges
	err := ct.channel.ExchangeDeclare(ct.config.Exchange, "direct", true, false, false, true, nil)
	if err != nil {
		return fmt.Errorf("could not create exchange %s: %w", ct.config.Exchange, err)
	}

	// create queues
	ct.createQueue, err = ct.channel.QueueDeclare(ct.config.CreateQueue, false, false, true, true, nil)
	if err != nil {
		return fmt.Errorf("could not create Create queue %s: %w", ct.config.CreateQueue, err)
	}

	ct.updateQueue, err = ct.channel.QueueDeclare(ct.config.UpdateQueue, false, false, true, true, nil)
	if err != nil {
		return fmt.Errorf("could not create Update queue %s: %w", ct.config.UpdateQueue, err)
	}

	ct.deleteQueue, err = ct.channel.QueueDeclare(ct.config.DeleteQueue, false, false, true, true, nil)
	if err != nil {
		return fmt.Errorf("could not create Delete queue %s: %w", ct.config.CreateQueue, err)
	}

	//bind queues
	err = ct.channel.QueueBind(ct.config.Exchange, ct.config.CreateQueue, "create_case_key", false, nil)
	if err != nil {
		return fmt.Errorf("could not bind create create_queue: %w", err)
	}

	err = ct.channel.QueueBind(ct.config.Exchange, ct.config.UpdateQueue, "update_case_key", false, nil)
	if err != nil {
		return fmt.Errorf("could not bind create update_queue: %w", err)
	}

	err = ct.channel.QueueBind(ct.config.Exchange, ct.config.DeleteQueue, "delete_case_key", false, nil)
	if err != nil {
		return fmt.Errorf("could not bind delete create_queue: %w", err)
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
