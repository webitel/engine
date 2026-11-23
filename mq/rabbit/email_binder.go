package rabbit

import (
	"context"
	"errors"
	"maps"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/webitel/engine/model"
	"github.com/webitel/wlog"
)

const emailOpsSize = 1000

// The pause after a failed unbind doubles from emailUnbindRetry to emailUnbindRetryMax.
var (
	emailUnbindRetry    = time.Second
	emailUnbindRetryMax = 30 * time.Second
)

var (
	errBindingGone  = errors.New("binding removed before the broker confirmed it")
	errQueueChanged = errors.New("the node queue changed while binding")
	errNoQueue      = errors.New("the node queue is not declared yet")
)

// emailOp is a bind whose caller waits for the broker.
type emailOp struct {
	b    *model.BindQueueEvent
	done chan error
}

// emailUnbind is a binding to remove from the node queue it was made on.
type emailUnbind struct {
	b     *model.BindQueueEvent
	queue string
}

// BindUserEmailWait binds the Email events of a user to this node's queue for the
// subscription id and returns once the broker confirmed it; a failed or late binding is removed.
func (dq *DomainQueue) BindUserEmailWait(ctx context.Context, id string, userId int64) (*model.BindQueueEvent, error) {
	b := &model.BindQueueEvent{
		UserId:   userId,
		Id:       id,
		Routing:  model.EmailBindingKey(dq.Id(), userId),
		Exchange: model.EmailServiceExchange,
	}
	dq.Lock()
	dq.bindings[bindKey(b)] = b
	dq.Unlock()

	dq.startEmailBinder()
	op := emailOp{b: b, done: make(chan error, 1)}
	select {
	case dq.emailOps <- op:
	case <-ctx.Done():
		dq.forget(b)

		return nil, ctx.Err()
	}

	select {
	case err := <-op.done:
		if err != nil {
			dq.UnbindEmail(b)

			return nil, err
		}

		return b, nil
	case <-ctx.Done():
		dq.UnbindEmail(b)

		return nil, ctx.Err()
	}
}

// UnbindEmail forgets the binding at once and has the binder remove it from the queue it is
// on, again after a failure; it never waits. A binding already forgotten went with its queue.
func (dq *DomainQueue) UnbindEmail(b *model.BindQueueEvent) {
	key := bindKey(b)
	dq.Lock()
	_, ok := dq.bindings[key]
	if ok {
		queue, bound := dq.emailQueues[key]
		if !bound {
			queue = dq.queue.Name
		}
		delete(dq.bindings, key)
		delete(dq.emailQueues, key)
		dq.emailUnbinds[key] = emailUnbind{b: b, queue: queue}
	}
	dq.Unlock()

	if ok {
		dq.startEmailBinder()
		select {
		case dq.emailKick <- struct{}{}:
		default:
		}
	}
}

// startEmailBinder runs the binder from the first Email binding of the domain on.
func (dq *DomainQueue) startEmailBinder() {
	dq.emailOnce.Do(func() { go dq.runEmailBinder() })
}

func (dq *DomainQueue) queueName() string {
	dq.RLock()
	defer dq.RUnlock()

	return dq.queue.Name
}

// runEmailBinder changes the Email bindings on a channel of its own, so a slow broker holds
// neither the listener nor the hub. After every bind it removes the bindings waiting for
// that, so a late bind is always followed by its unbind.
func (dq *DomainQueue) runEmailBinder() {
	var ch amqpChannel
	retry := time.NewTimer(time.Hour)
	retry.Stop()
	var pause time.Duration
	retrying := false
	defer func() {
		retry.Stop()
		if ch != nil {
			_ = ch.Close()
		}
	}()

	for {
		select {
		case <-dq.stop:
			return
		case op := <-dq.emailOps:
			op.done <- dq.emailBind(&ch, op.b)
		case <-dq.emailKick:
		case <-retry.C:
			retrying = false
		}

		if !dq.emailUnbindAll(&ch) {
			pause = 0
		} else if !retrying {
			pause = min(max(2*pause, emailUnbindRetry), emailUnbindRetryMax)
			retry.Reset(pause)
			retrying = true
		}
	}
}

// emailChannel opens a channel when needed and declares on it the exchange of email-service
// with the parameters of its sender, so a node binds before the first event is sent.
func (dq *DomainQueue) emailChannel(ch *amqpChannel) (amqpChannel, error) {
	if *ch == nil || (*ch).IsClosed() {
		c, err := dq.newEmailChannel()
		if err != nil {
			return nil, err
		}
		if err = c.ExchangeDeclare(model.EmailServiceExchange, model.MQ_TOPIC, true, false, false, false, nil); err != nil {
			_ = c.Close()

			return nil, err
		}
		*ch = c
	}

	return *ch, nil
}

// emailDrop closes a channel a command failed on; the next change opens another.
func emailDrop(ch *amqpChannel) {
	if *ch != nil {
		_ = (*ch).Close()
		*ch = nil
	}
}

func (dq *DomainQueue) emailBind(ch *amqpChannel, b *model.BindQueueEvent) error {
	if !dq.HasBinding(b) {
		return errBindingGone
	}
	name := dq.queueName()
	if name == "" {
		return errNoQueue
	}
	c, err := dq.emailChannel(ch)
	if err != nil {
		return err
	}
	if err = c.QueueBind(name, b.Routing, b.Exchange, false, amqp.Table{"x-sock-id": b.Id}); err != nil {
		emailDrop(ch)

		return err
	}
	// The queue name tells the queue generation: the binding is kept for this one only.
	dq.Lock()
	defer dq.Unlock()
	if dq.queue.Name != name {
		return errQueueChanged
	}
	if _, ok := dq.bindings[bindKey(b)]; ok {
		dq.emailQueues[bindKey(b)] = name
	}

	return nil
}

// emailUnbindAll reports whether some unbind waits for a retry. A refusal of the broker is
// final; an unbind for an older queue is dropped, as that queue went with its bindings.
func (dq *DomainQueue) emailUnbindAll(ch *amqpChannel) bool {
	dq.RLock()
	pending := maps.Clone(dq.emailUnbinds)
	dq.RUnlock()

	current := dq.queueName()
	for key, u := range pending {
		if u.queue != current {
			dq.emailUnbound(key)
			continue
		}
		c, err := dq.emailChannel(ch)
		if err == nil {
			err = c.QueueUnbind(u.queue, u.b.Routing, u.b.Exchange, amqp.Table{"x-sock-id": u.b.Id})
		}
		if err == nil {
			dq.emailUnbound(key)
			continue
		}

		emailDrop(ch)
		var refused *amqp.Error
		if errors.As(err, &refused) && refused.Recover {
			dq.log.Warn("email unbind refused", wlog.String("sock_id", u.b.Id), wlog.Err(err))
			dq.emailUnbound(key)
			continue
		}
		dq.log.Warn("email unbind failed, will retry", wlog.String("sock_id", u.b.Id), wlog.Err(err))

		return true
	}

	return false
}

func (dq *DomainQueue) emailUnbound(key string) {
	dq.Lock()
	delete(dq.emailUnbinds, key)
	dq.Unlock()
}
