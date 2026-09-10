package rabbit

import (
	"cmp"
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/pkg/errors"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/webitel/engine/model"
	"github.com/webitel/wlog"
)

type (
	BrokerQueueOption func(*BrokerQueue)
	ChannelProvider   func() (*amqp.Channel, error)
	Handler           func(ctx context.Context, d amqp.Delivery) error
)

func WithConsumerTag(tag string) BrokerQueueOption {
	return func(bq *BrokerQueue) {
		bq.consumerTag = tag
	}
}

func WithConcurrency(n int) BrokerQueueOption {
	return func(bq *BrokerQueue) {
		bq.concurrency = n
	}
}

type BrokerQueue struct {
	queueCfg   *QueueConfig
	bindings   []*QueueBindConfig
	getChannel ChannelProvider

	consumerTag  string
	concurrency  int
	reconnectSec int

	log     *wlog.Logger
	mx      sync.Mutex
	channel *amqp.Channel
	stop    chan struct{}
	stopped chan struct{}
	handler Handler
}

func NewBrokerQueue(cfg *QueueConfig, getChannel ChannelProvider, handler Handler, opts ...BrokerQueueOption) *BrokerQueue {
	b := &BrokerQueue{
		queueCfg:     cfg,
		getChannel:   getChannel,
		concurrency:  1,
		reconnectSec: RECONNECT_SEC,
		log: wlog.GlobalLogger().With(
			wlog.Namespace("context"),
			wlog.String("protocol", "amqp"),
			wlog.String("queue", cfg.Name),
		),
		stop:    make(chan struct{}),
		stopped: make(chan struct{}),
		handler: handler,
	}

	for _, o := range opts {
		o(b)
	}

	return b
}

func (b *BrokerQueue) Bind(cfg *QueueBindConfig) *BrokerQueue {
	b.bindings = append(b.bindings, cfg)

	return b
}

func (b *BrokerQueue) setup(ch *amqp.Channel) model.AppError {
	_, err := ch.QueueDeclare(
		b.queueCfg.Name,
		b.queueCfg.Durable,
		b.queueCfg.AutoDelete,
		b.queueCfg.Exclusive,
		b.queueCfg.NoWait,
		amqp.Table(b.queueCfg.Args),
	)

	if err != nil {
		return model.NewCustomCodeError(
			"rabbit.broker.queue.setup.declare",
			fmt.Sprintf("declaring new queue: %+v", err),
			500,
		)
	}

	for _, bind := range b.bindings {
		name := cmp.Or(bind.Name, b.queueCfg.Name)
		if err := ch.QueueBind(name, bind.Key, bind.Exchange, bind.NoWait, amqp.Table(bind.Args)); err != nil {
			return model.NewCustomCodeError(
				"rabbit.broker.queue.setup.bind",
				fmt.Sprintf("bind %q -> %q (%q): %+v", name, bind.Exchange, bind.Key, err),
				500,
			)
		}
	}

	return nil
}

func (b *BrokerQueue) Run(ctx context.Context) model.AppError {
	defer close(b.stopped)

	for {
		select {
		case <-b.stop:
			return nil
		case <-ctx.Done():
			return model.NewCustomCodeError(
				"rabbit.broker.queue.run_context_done",
				fmt.Sprintf("context done on running broker queue %s: %+v", b.queueCfg.Name, ctx.Err()),
				408,
			)
		default:
		}

		ch, err := b.getChannel()
		if err != nil {
			b.log.Error("getting AMQP channel", wlog.Err(err))
			if !b.sleep(ctx) {
				return nil
			}

			continue
		}

		if err := b.setup(ch); err != nil {
			b.log.Error("setup queue", wlog.Err(err))
			b.closeChan(ch)
			if !b.sleep(ctx) {
				return nil
			}

			continue
		}

		if err := ch.Qos(b.concurrency, 0, false); err != nil {
			b.log.Error("qos", wlog.Err(err))
		}

		deliveries, err := ch.Consume(
			b.queueCfg.Name,
			b.consumerTag,
			false,
			b.queueCfg.Exclusive,
			false,
			b.queueCfg.NoWait,
			nil,
		)

		if err != nil {
			b.log.Error("consuming queue deliveries", wlog.Err(err))
			b.closeChan(ch)
			if !b.sleep(ctx) {
				return nil
			}

			continue
		}

		b.mx.Lock()
		b.channel = ch
		b.mx.Unlock()

		closeErr := make(chan *amqp.Error, 1)
		ch.NotifyClose(closeErr)

		b.log.Info("consuming started")

		cont := b.consumeLoop(ctx, deliveries, closeErr)
		b.closeChan(ch)

		if !cont {
			return nil
		}

		b.log.Warn("channel lost, reconnecting")
	}
}

func (b *BrokerQueue) closeChan(ch *amqp.Channel) {
	if ch == nil {
		return
	}

	if err := ch.Close(); err != nil && !errors.Is(err, amqp.ErrClosed) {
		b.log.Warn("closing amqp channel", wlog.Err(err))
	}
}

func (b *BrokerQueue) process(ctx context.Context, d amqp.Delivery) {
	defer func() {
		if r := recover(); r != nil {
			b.log.Error(fmt.Sprintf("panic in handler: %v", r))
			_ = d.Nack(false, false)
		}
	}()

	if err := b.handler(ctx, d); err != nil {
		b.log.Error("handler error", wlog.Err(err))

		_ = d.Nack(false, false)

		return
	}

	_ = d.Ack(false)
}

func (b *BrokerQueue) consumeLoop(ctx context.Context, deliveries <-chan amqp.Delivery, closeErr <-chan *amqp.Error) bool {
	sem := make(chan struct{}, b.concurrency)

	var wg sync.WaitGroup

	for {
		select {
		case <-b.stop:
			wg.Wait()
			return false
		case <-ctx.Done():
			wg.Wait()
			return false
		case <-closeErr:
			wg.Wait()
			return true
		case d, ok := <-deliveries:
			if !ok {
				wg.Wait()
				return true
			}

			select {
			case sem <- struct{}{}:
			case <-ctx.Done():
				wg.Wait()
				return false
			case <-b.stop:
				wg.Wait()
				return false
			case <-closeErr:
				wg.Wait()
				return true
			}

			wg.Add(1)
			go func(d amqp.Delivery) {
				defer wg.Done()
				defer func() { <-sem }()

				b.process(ctx, d)
			}(d)
		}
	}
}

func (b *BrokerQueue) sleep(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return false
	case <-time.After(time.Duration(b.reconnectSec) * time.Second):
		return true
	case <-b.stop:
		return false
	}
}

func (b *BrokerQueue) Close() {
	close(b.stop)

	select {
	case <-b.stopped:
	case <-time.After(time.Duration(5) * time.Second):
		b.log.Warn("close timed out waiting for consume loop to stop")
	}

	b.mx.Lock()
	defer b.mx.Unlock()

	if b.channel != nil && !b.channel.IsClosed() {
		_ = b.channel.Cancel(b.consumerTag, false)
		_ = b.channel.Close()
	}
}
