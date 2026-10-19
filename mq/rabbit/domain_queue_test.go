package rabbit

import (
	"cmp"
	"fmt"
	"slices"
	"sync"
	"testing"
	"testing/synctest"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/webitel/wlog"

	"github.com/webitel/engine/model"
)

// These tests take the AMQP channel away from a domain queue in the ways a
// broker does, and check what the queue has bound afterwards. They run on a
// fake clock (testing/synctest), so the 5 s reconnect pause costs nothing.
//
// The real thing — an engine binary against a RabbitMQ that is restarted — is
// exercised by the stand kept with the ticket (WTEL-10528), not from here.

var (
	// What a restarting broker sends: a hard error, the whole connection goes.
	errConnectionForced = &amqp.Error{Code: amqp.ConnectionForced, Reason: "CONNECTION_FORCED - broker shutdown", Server: true}
	// A channel exception for one command, as for a bind to a missing exchange.
	errNotFound = &amqp.Error{Code: amqp.NotFound, Reason: "NOT_FOUND - no exchange", Server: true, Recover: true}
	// Also a channel exception, for a queue the user may not declare.
	errAccessRefused = &amqp.Error{Code: amqp.AccessRefused, Reason: "ACCESS_REFUSED", Server: true, Recover: true}
)

type fakeBind struct {
	exchange, key, sock string
	noWait              bool
}

// fakeChannel records what a domain queue does to its channel and can be
// closed the way a broker closes one.
type fakeChannel struct {
	broker *fakeBroker

	mu         sync.Mutex
	closed     bool
	pending    *amqp.Error // the broker refused a no-wait command; the close has not arrived yet
	queue      string
	binds      []fakeBind
	unbinds    []fakeBind
	deleted    []string
	noWaitCmds []string
	closeCalls int
	delivery   chan amqp.Delivery
	listeners  []chan *amqp.Error
}

// command applies one AMQP command. A refused command closes the channel, as a
// channel exception does. A caller that did not wait for the answer learns
// nothing: the close arrives a moment later, and until then further no-wait
// commands also appear to succeed. The first command that does wait is told
// instead, which is how a failure ends up blamed on the wrong command.
func (f *fakeChannel) command(name string, noWait bool, refuse *amqp.Error, apply func()) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.closed {
		return amqp.ErrClosed
	}

	if noWait {
		f.noWaitCmds = append(f.noWaitCmds, name)
	}

	if f.pending != nil {
		if noWait {
			return nil
		}

		err := f.pending
		f.shutdownLocked(err)

		return err
	}

	if refuse == nil {
		apply()

		return nil
	}

	if noWait {
		f.pending = refuse

		time.AfterFunc(time.Millisecond, func() { f.drop(refuse) })

		return nil
	}

	f.shutdownLocked(refuse)

	return refuse
}

func (f *fakeChannel) QueueDeclare(name string, _, _, _, noWait bool, _ amqp.Table) (amqp.Queue, error) {
	err := f.command("queue.declare", noWait, f.broker.declareRefusal(), func() { f.queue = name })

	return amqp.Queue{Name: name}, err
}

func (f *fakeChannel) QueueBind(_, key, exchange string, noWait bool, args amqp.Table) error {
	sock, _ := args["x-sock-id"].(string)

	return f.command("queue.bind", noWait, f.broker.bindRefusal(exchange, sock), func() {
		f.binds = append(f.binds, fakeBind{exchange, key, sock, noWait})
	})
}

func (f *fakeChannel) QueueUnbind(_, key, exchange string, args amqp.Table) error {
	sock, _ := args["x-sock-id"].(string)

	return f.command("queue.unbind", false, nil, func() {
		f.unbinds = append(f.unbinds, fakeBind{exchange: exchange, key: key, sock: sock})
	})
}

func (f *fakeChannel) QueueDelete(name string, _, _, _ bool) (int, error) {
	return 0, f.command("queue.delete", false, nil, func() { f.deleted = append(f.deleted, name) })
}

func (f *fakeChannel) Consume(_, _ string, _, _, _, noWait bool, _ amqp.Table) (<-chan amqp.Delivery, error) {
	var deliveries chan amqp.Delivery

	err := f.command("basic.consume", noWait, nil, func() {
		deliveries = make(chan amqp.Delivery)
		f.delivery = deliveries
	})

	return deliveries, err
}

func (f *fakeChannel) NotifyClose(c chan *amqp.Error) chan *amqp.Error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.closed {
		close(c)
	} else {
		f.listeners = append(f.listeners, c)
	}

	return c
}

func (f *fakeChannel) IsClosed() bool {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.closed
}

func (f *fakeChannel) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.closeCalls++
	f.shutdownLocked(nil)

	return nil
}

// shutdownLocked does what amqp091 does when a channel ends: the error goes to
// the NotifyClose listeners, deliveries stop, the listeners are closed.
func (f *fakeChannel) shutdownLocked(err *amqp.Error) {
	if f.closed {
		return
	}

	f.closed = true

	for _, l := range f.listeners {
		if err != nil {
			l <- err
		}
	}

	if f.delivery != nil {
		close(f.delivery)
		f.delivery = nil
	}

	for _, l := range f.listeners {
		close(l)
	}
}

// drop closes the channel from the broker's side.
func (f *fakeChannel) drop(err *amqp.Error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.shutdownLocked(err)
}

// cancelConsumer is the broker canceling the consumer, for example because
// the queue was deleted: deliveries stop, the channel itself stays open.
func (f *fakeChannel) cancelConsumer() {
	f.mu.Lock()
	defer f.mu.Unlock()

	close(f.delivery)
	f.delivery = nil
}

func (f *fakeChannel) snapshot() (binds, unbinds []fakeBind, noWaitCmds []string) {
	f.mu.Lock()
	defer f.mu.Unlock()

	return slices.Clone(f.binds), slices.Clone(f.unbinds), slices.Clone(f.noWaitCmds)
}

// bound reports whether the binding exists on this channel's queue, however
// often it was applied.
func (f *fakeChannel) bound(b *model.BindQueueEvent) bool {
	binds, _, _ := f.snapshot()

	return slices.ContainsFunc(binds, func(x fakeBind) bool {
		return x.exchange == b.Exchange && x.key == b.Routing && x.sock == b.Id
	})
}

// fakeBroker hands out channels and decides what it refuses.
type fakeBroker struct {
	mu            sync.Mutex
	down          bool
	refuseDeclare *amqp.Error
	missing       map[string]bool // exchanges that do not exist
	failReplay    int             // the next n socket binds die with the connection
	channels      []*fakeChannel
	overlap       bool // a channel was opened while an earlier one was still open
	runaway       bool // the queue reconnected without end
	budget        int  // channels it may still open
}

// More channels than any test needs: a queue that reconnects in a loop runs
// out of them, so the test fails with a message instead of spinning.
const maxChannels = 40

func newFakeBroker() *fakeBroker { return &fakeBroker{budget: maxChannels} }

func (b *fakeBroker) set(change func()) {
	b.mu.Lock()
	defer b.mu.Unlock()

	change()
}

func (b *fakeBroker) open() (amqpChannel, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.down {
		return nil, amqp.ErrClosed
	}

	if b.budget == 0 {
		b.runaway = true

		return nil, amqp.ErrClosed
	}

	b.budget--

	for _, c := range b.channels {
		b.overlap = b.overlap || !c.IsClosed()
	}

	f := &fakeChannel{broker: b}
	b.channels = append(b.channels, f)

	return f, nil
}

func (b *fakeBroker) declareRefusal() *amqp.Error {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.refuseDeclare
}

func (b *fakeBroker) bindRefusal(exchange, sock string) *amqp.Error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.missing[exchange] {
		return errNotFound
	}

	if sock != "" && b.failReplay > 0 {
		b.failReplay--

		return errConnectionForced
	}

	return nil
}

func (b *fakeBroker) opened() int {
	b.mu.Lock()
	defer b.mu.Unlock()

	return len(b.channels)
}

// last is the channel the queue should be working on now.
func (b *fakeBroker) last(t *testing.T) *fakeChannel {
	t.Helper()

	b.mu.Lock()
	defer b.mu.Unlock()

	if len(b.channels) == 0 {
		t.Fatal("the queue never opened a channel")
	}

	return b.channels[len(b.channels)-1]
}

// startQueue starts a domain queue for domain 1 on the fake broker. Every test
// also ends with the two checks in the cleanup: the queue stops, and it never
// had two channels open at once (two consumers would duplicate every event).
func startQueue(t *testing.T, b *fakeBroker) *DomainQueue {
	t.Helper()

	client := &AMQP{log: wlog.NewLogger(&wlog.LoggerConfiguration{})}
	dq := newDomainQueue(client, 1, nil).(*DomainQueue)
	dq.newChannel = b.open
	dq.Start()

	t.Cleanup(func() {
		// Whatever the test did to the broker, let the queue connect, because
		// only a connected queue can be stopped. A queue whose listener is gone
		// for good never gets there: such a regression ends in synctest's
		// deadlock panic, and its goroutine dump shows where the queue is stuck.
		b.set(func() { b.down, b.refuseDeclare, b.missing, b.failReplay, b.budget = false, nil, nil, 0, maxChannels })
		pass(1)
		dq.Stop()

		b.set(func() {
			if b.overlap {
				t.Error("a channel was opened while the previous one was still open")
			}

			if b.runaway {
				t.Errorf("the queue reconnected in a loop: more than %d channels", maxChannels)
			}
		})
	})

	synctest.Wait()

	return dq
}

// pass lets n reconnect pauses go by.
func pass(n int) {
	synctest.Wait()
	time.Sleep(time.Duration(n)*RECONNECT_SEC*time.Second + time.Millisecond)
	synctest.Wait()
}

func registered(dq *DomainQueue) int {
	dq.RLock()
	defer dq.RUnlock()

	return len(dq.bindings)
}

func down(dq *DomainQueue) int32 { return dq.client.domainQueuesDown.Load() }

// expectEventReachesHub publishes an agent status event on the channel and
// expects the queue to hand it to the hub: the listener is running and reads
// from this channel.
func expectEventReachesHub(t *testing.T, dq *DomainQueue, f *fakeChannel) {
	t.Helper()

	f.mu.Lock()
	deliveries := f.delivery
	f.mu.Unlock()

	if deliveries == nil {
		t.Fatal("the channel has no consumer")
	}

	go func() {
		select {
		case deliveries <- amqp.Delivery{
			Exchange:    model.CallCenterExchange,
			RoutingKey:  "events.status.1.10",
			ContentType: "text/json",
			Body:        []byte(`{"event":"agent_status","user_id":10,"data":{"status":"online"}}`),
		}:
		case <-time.After(time.Second):
		}
	}()

	select {
	case ev := <-dq.Events():
		if ev.Event != model.WebsocketCCEventAgentStatus || ev.UserId != 10 {
			t.Fatalf("unexpected event handed to the hub: %+v", ev)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("an event published after the reconnect never reached the hub: the listener is not running")
	}
}

func expectBound(t *testing.T, f *fakeChannel, when string, bindings ...*model.BindQueueEvent) {
	t.Helper()

	for _, b := range bindings {
		if !f.bound(b) {
			binds, _, _ := f.snapshot()
			t.Fatalf("%s: %s %s for socket %s is not bound; bound: %+v", when, b.Exchange, b.Routing, b.Id, binds)
		}
	}
}

// The ticket: RabbitMQ restarts under open Workspaces. The queue must come
// back on a new channel with every subscription the open sockets had, without
// the sockets doing anything.
func TestDomainQueueRestoresBindingsAfterChannelLoss(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		b := newFakeBroker()
		dq := startQueue(t, b)

		call := dq.BindUserCall("sock-1", 10)
		chat := dq.BindUserChat("sock-1", 10)
		status := dq.BindAgentStatusEvents("sock-2", 11, 5)
		again := dq.BindUserCall("sock-1", 10) // the same subscription sent twice

		synctest.Wait()

		first := b.last(t)
		expectBound(t, first, "before the loss", call, chat, status)

		first.drop(errConnectionForced)
		synctest.Wait()

		second := b.last(t)
		if b.opened() != 2 || second == first {
			t.Fatalf("expected one reconnect, the queue opened %d channels", b.opened())
		}

		expectBound(t, second, "after the reconnect", call, chat, status, again)

		binds, _, noWait := second.snapshot()

		want := []fakeBind{
			{model.AppExchange, "notification.1", "", false},
			{call.Exchange, call.Routing, "sock-1", false},
			{chat.Exchange, chat.Routing, "sock-1", false},
			{status.Exchange, status.Routing, "sock-2", false},
		}

		key := func(x fakeBind) string { return fmt.Sprint(x) }
		slices.SortFunc(binds, func(x, y fakeBind) int { return cmp.Compare(key(x), key(y)) })
		slices.SortFunc(want, func(x, y fakeBind) int { return cmp.Compare(key(x), key(y)) })

		// Exactly these, each once, each waited for: a restore that does not
		// wait cannot tell which binding the broker refused.
		if !slices.Equal(binds, want) {
			t.Fatalf("bindings on the new queue:\n got %+v\nwant %+v", binds, want)
		}

		// Declare, the notification bind and consume must wait as well, or their
		// failure surfaces on the first restored binding and is blamed on it.
		if len(noWait) != 0 {
			t.Fatalf("commands sent without waiting for the broker's answer: %v", noWait)
		}

		if second.queue == first.queue {
			t.Fatalf("the queue name did not change: %s", second.queue)
		}

		expectEventReachesHub(t, dq, second)

		pass(3)

		if b.opened() != 2 {
			t.Fatalf("the queue kept reconnecting: %d channels opened", b.opened())
		}
	})
}

// Root cause B: a subscribe that arrives while the broker is away used to make
// the reconnect park forever, so nothing was delivered until engine restarted.
func TestDomainQueueAppliesBindsMadeWhileDisconnected(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		b := newFakeBroker()
		b.down = true
		dq := startQueue(t, b)

		call := dq.BindUserCall("sock-1", 10)
		chat := dq.BindUserChat("sock-1", 10)
		status := dq.BindAgentStatusEvents("sock-1", 10, 5)

		pass(2)

		if b.opened() != 0 {
			t.Fatalf("a channel was opened while the broker was down")
		}

		b.set(func() { b.down = false })
		pass(1)

		f := b.last(t)
		expectBound(t, f, "after the broker came back", call, chat, status)
		expectEventReachesHub(t, dq, f)
	})
}

// Root cause C: the broker cancels the consumer (the queue was deleted). Only
// the deliveries end; the channel stays open and nothing reports a close. The
// listener used to exit there and the domain stayed dead.
func TestDomainQueueReconnectsWhenConsumerIsCanceled(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		b := newFakeBroker()
		dq := startQueue(t, b)

		call := dq.BindUserCall("sock-1", 10)

		synctest.Wait()

		first := b.last(t)
		first.cancelConsumer()
		synctest.Wait()

		second := b.last(t)
		if second == first {
			t.Fatal("no reconnect after the consumer was canceled")
		}

		if !first.IsClosed() || first.closeCalls != 1 {
			t.Fatalf("the channel left behind was not closed (Close called %d times)", first.closeCalls)
		}

		expectBound(t, second, "after the reconnect", call)
		expectEventReachesHub(t, dq, second)
	})
}

// A subscription that ended must not come back on the next queue.
func TestDomainQueueDoesNotRestoreWhatWasUnsubscribed(t *testing.T) {
	t.Run("unsubscribed while connected", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			b := newFakeBroker()
			dq := startQueue(t, b)

			// Two tabs of one user: same exchange and routing key, and only
			// one of them goes away.
			gone, kept := dq.BindUserCall("sock-1", 10), dq.BindUserCall("sock-2", 10)

			synctest.Wait()

			first := b.last(t)
			if err := dq.Unbind(gone); err != nil {
				t.Fatalf("unbind on a live channel: %v", err)
			}

			_, unbinds, _ := first.snapshot()
			if !slices.Equal(unbinds, []fakeBind{{exchange: gone.Exchange, key: gone.Routing, sock: "sock-1"}}) {
				t.Fatalf("the unbind did not reach the broker as sent: %+v", unbinds)
			}

			first.drop(errConnectionForced)
			synctest.Wait()

			second := b.last(t)
			expectBound(t, second, "after the reconnect", kept)

			if second.bound(gone) {
				t.Fatal("an unsubscribed binding was restored")
			}
		})
	})

	t.Run("socket closed while the broker was down", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			b := newFakeBroker()
			dq := startQueue(t, b)

			gone, kept := dq.BindUserCall("sock-1", 10), dq.BindUserCall("sock-2", 10)

			synctest.Wait()
			b.set(func() { b.down = true })
			b.last(t).drop(errConnectionForced)
			synctest.Wait()

			if err := dq.BulkUnbind([]*model.BindQueueEvent{gone}); err != nil {
				t.Fatalf("bulk unbind: %v", err)
			}

			b.set(func() { b.down = false })
			pass(1)

			f := b.last(t)
			expectBound(t, f, "after the broker came back", kept)

			if f.bound(gone) {
				t.Fatal("the binding of a socket that closed during the outage was restored")
			}
		})
	})

	t.Run("subscribed and unsubscribed while the broker was down", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			b := newFakeBroker()
			b.down = true
			dq := startQueue(t, b)

			// Still waiting in the queue's buffer when the socket goes away.
			gone := dq.BindUserCall("sock-1", 10)
			if err := dq.Unbind(gone); err == nil {
				t.Fatal("unbind without a channel is expected to report that")
			}

			b.set(func() { b.down = false })
			pass(1)

			f := b.last(t)
			if f.bound(gone) {
				t.Fatal("a binding that was unsubscribed before it was ever applied got bound")
			}

			expectEventReachesHub(t, dq, f)
		})
	})
}

// A binding the broker refuses (the exchange does not exist) closes the
// channel. Restoring it on every new queue would close each of them in turn,
// so it is dropped; everything else must keep working.
func TestDomainQueueDropsABindingTheBrokerRefuses(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		b := newFakeBroker()
		b.missing = map[string]bool{model.ChatExchange: true}
		dq := startQueue(t, b)

		call := dq.BindUserCall("sock-1", 10)

		synctest.Wait()

		chat := dq.BindUserChat("sock-1", 10)

		pass(1)

		settled := b.opened()
		f := b.last(t)

		if f.IsClosed() {
			t.Fatal("the queue did not recover from the refused binding")
		}

		expectBound(t, f, "after the refusal", call)

		if f.bound(chat) || registered(dq) != 1 {
			t.Fatalf("the refused binding is still kept (%d remembered)", registered(dq))
		}

		pass(4)

		if b.opened() != settled {
			t.Fatalf("the queue keeps reconnecting: %d channels, then %d", settled, b.opened())
		}

		expectEventReachesHub(t, dq, f)
	})
}

// A restore that fails because the connection went away says nothing about the
// binding: it must be restored on the next attempt, not dropped.
func TestDomainQueueKeepsBindingsWhenTheRestoreIsInterrupted(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		b := newFakeBroker()
		dq := startQueue(t, b)

		call, chat := dq.BindUserCall("sock-1", 10), dq.BindUserChat("sock-1", 10)

		synctest.Wait()
		b.set(func() { b.failReplay = 1 })
		b.last(t).drop(errConnectionForced)
		pass(1)

		if registered(dq) != 2 {
			t.Fatalf("a binding was dropped because the connection died: %d remembered", registered(dq))
		}

		f := b.last(t)
		expectBound(t, f, "on the attempt after the interrupted one", call, chat)
		expectEventReachesHub(t, dq, f)
	})
}

// Found on the stand: the broker refuses to declare the queue. If declare does
// not wait for the answer, the refusal surfaces on the first restored binding,
// which is then dropped as "refused" — one subscription lost per retry, and
// nothing left to restore once the broker is fixed.
func TestDomainQueueKeepsBindingsWhileTheQueueCannotBeDeclared(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		b := newFakeBroker()
		dq := startQueue(t, b)

		call, chat := dq.BindUserCall("sock-1", 10), dq.BindUserChat("sock-1", 10)

		synctest.Wait()
		b.set(func() { b.refuseDeclare = errAccessRefused })
		b.last(t).drop(errConnectionForced)
		pass(4)

		if registered(dq) != 2 {
			t.Fatalf("bindings were dropped while the queue could not be declared: %d of 2 remembered", registered(dq))
		}

		if down(dq) != 1 {
			t.Fatalf("a queue that cannot be declared is not reported down (counter %d)", down(dq))
		}

		b.set(func() { b.refuseDeclare = nil })
		pass(1)

		f := b.last(t)
		expectBound(t, f, "once the queue could be declared", call, chat)
		expectEventReachesHub(t, dq, f)

		if down(dq) != 0 {
			t.Fatalf("still reported down after recovery (counter %d)", down(dq))
		}
	})
}

// Ping reports a domain consumer that is down. The counter behind it must go
// up once per outage, not once per failed attempt, and come back to zero.
func TestDomainQueueDownCounter(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		b := newFakeBroker()
		dq := startQueue(t, b)

		if down(dq) != 0 {
			t.Fatalf("down before anything happened (counter %d)", down(dq))
		}

		for round := 1; round <= 2; round++ {
			b.set(func() { b.down = true })
			b.last(t).drop(errConnectionForced)
			pass(3)

			if down(dq) != 1 {
				t.Fatalf("round %d: after three failed attempts the counter is %d, want 1", round, down(dq))
			}

			b.set(func() { b.down = false })
			pass(1)

			if down(dq) != 0 {
				t.Fatalf("round %d: counter %d after the reconnect, want 0", round, down(dq))
			}
		}

		// A loss the queue recovers from at once must not leave a trace either.
		b.last(t).cancelConsumer()
		synctest.Wait()

		if down(dq) != 0 {
			t.Fatalf("counter %d after an immediate reconnect, want 0", down(dq))
		}
	})
}
