package rabbit

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"testing"
	"testing/synctest"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/webitel/engine/model"
	"github.com/webitel/wlog"
)

func TestDomainQueueEmailEvents(t *testing.T) {
	dq := newDomainQueue(&AMQP{log: wlog.NewLogger(&wlog.LoggerConfiguration{})}, 1, nil).(*DomainQueue)

	body := func(domain string) string {
		return `{"event_id":"5b247aa1-9433-53ce-b37d-346cec14bff5","event_type":"email.workspace.thread.changed",
"schema_version":1,"occurred_at":"2026-11-06T09:00:00Z","domain_id":` + domain + `,"user_id":42,"thread_id":1024,
"access":"current","workspace_version":17,"projection_version":905,"message_id":null,"changes":["projection"]}`
	}
	deliver := func(key, body string) *model.EmailEvent {
		got := make(chan *model.EmailEvent, 1)
		go func() {
			select {
			case e := <-dq.EmailEvents():
				got <- e
			case <-time.After(100 * time.Millisecond):
				got <- nil
			}
		}()
		dq.readMessage(amqp.Delivery{Exchange: model.EmailServiceExchange, RoutingKey: key,
			ContentType: model.EmailContentType, Body: []byte(body)})

		return <-got
	}

	if e := deliver("email.workspace.thread.changed.v1.1.42", body("1")); e == nil || e.UserId != 42 || e.WorkspaceVersion != 17 {
		t.Fatalf("event %+v", e)
	}
	if e := deliver("email.workspace.thread.changed.v1.2.42", body("2")); e != nil {
		t.Errorf("other domain: delivered %+v", e)
	}
}

// emailQueue starts a domain queue whose Email binder has a broker of its own, so the
// checks of startQueue still see one consumer channel at a time.
func emailQueue(t *testing.T) (dq *DomainQueue, consumer, binder *fakeBroker) {
	t.Helper()
	consumer, binder = newFakeBroker(), newFakeBroker()
	dq = startQueue(t, consumer)
	dq.newEmailChannel = binder.open

	return dq, consumer, binder
}

func emailBind(id string) fakeBind {
	return fakeBind{model.EmailServiceExchange, "email.workspace.*.*.v1.1.42", id, false}
}

// binderLog is every bind and unbind the binder made, over all its channels.
func binderLog(b *fakeBroker) (binds, unbinds []fakeBind) {
	b.mu.Lock()
	channels := slices.Clone(b.channels)
	b.mu.Unlock()
	for _, c := range channels {
		bs, us, _ := c.snapshot()
		binds, unbinds = append(binds, bs...), append(unbinds, us...)
	}
	for i := range unbinds {
		unbinds[i].noWait = false
	}

	return binds, unbinds
}

// The reply to subscribe_email waits for the broker: a confirmed binding is kept, a refused
// one is not left behind.
func TestDomainQueueEmailBindWait(t *testing.T) {
	t.Run("confirmed", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			dq, _, binder := emailQueue(t)
			bind, err := dq.BindUserEmailWait(context.Background(), "sock-1/a", 42)
			if err != nil || bind.Exchange != model.EmailServiceExchange || bind.Routing != "email.workspace.*.*.v1.1.42" {
				t.Fatalf("bind %+v, %v", bind, err)
			}
			if binds, _ := binderLog(binder); !slices.Contains(binds, emailBind("sock-1/a")) || !dq.HasBinding(bind) {
				t.Fatalf("bound %+v", binds)
			}
		})
	})

	t.Run("refused", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			dq, _, binder := emailQueue(t)
			binder.set(func() { binder.missing = map[string]bool{model.EmailServiceExchange: true} })
			if _, err := dq.BindUserEmailWait(context.Background(), "sock-1/a", 42); !errors.Is(err, errNotFound) {
				t.Fatalf("err = %v", err)
			}
			if registered(dq) != 0 {
				t.Fatal("a refused binding is kept")
			}
		})
	})
}

// A subscribe that timed out and finished late never removes the next subscription of the
// same socket, and a slow broker holds neither the listener nor UnbindEmail.
func TestDomainQueueEmailLateBind(t *testing.T) {
	for _, late := range []string{"success", "error"} {
		t.Run(late, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				dq, consumer, binder := emailQueue(t)
				hold := make(chan struct{})
				binder.set(func() { binder.hold = hold })

				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				if _, err := dq.BindUserEmailWait(ctx, "sock-1/a", 42); !errors.Is(err, context.DeadlineExceeded) {
					t.Fatalf("first: %v", err)
				}
				var retry *model.BindQueueEvent
				var retryErr error
				done := make(chan struct{})
				go func() {
					retry, retryErr = dq.BindUserEmailWait(context.Background(), "sock-1/b", 42)
					close(done)
				}()
				synctest.Wait()

				expectEventReachesHub(t, dq, consumer.last(t))
				dq.UnbindEmail(&model.BindQueueEvent{Id: "sock-9/x", Exchange: model.EmailServiceExchange, Routing: "email.workspace.*.*.v1.1.9"})

				binder.set(func() {
					binder.hold = nil
					if late == "error" {
						binder.failReplay = 1
					}
				})
				close(hold)
				<-done
				if retryErr != nil {
					t.Fatalf("retry: %v", retryErr)
				}

				binds, unbinds := binderLog(binder)
				if !slices.Contains(binds, emailBind("sock-1/b")) || slices.Contains(unbinds, emailBind("sock-1/b")) {
					t.Fatalf("binds %+v, unbinds %+v", binds, unbinds)
				}
				if slices.Contains(binds, emailBind("sock-1/a")) && !slices.Contains(unbinds, emailBind("sock-1/a")) {
					t.Fatalf("the late binding stays: binds %+v, unbinds %+v", binds, unbinds)
				}
				if registered(dq) != 1 || !dq.HasBinding(retry) {
					t.Fatalf("registered %d", registered(dq))
				}
			})
		})
	}
}

// After a lost channel the queue tells the hub once its bindings are back. The Email
// bindings are not restored there: they belonged to the old queue, and a slow Email bind
// holds neither the reconnect nor the listener.
func TestDomainQueueRebound(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		dq, consumer, binder := emailQueue(t)
		select {
		case <-dq.Rebound():
			t.Fatal("the first connect is no reconnect")
		default:
		}
		bind, err := dq.BindUserEmailWait(context.Background(), "sock-1/a", 42)
		if err != nil {
			t.Fatal(err)
		}
		call := dq.BindUserCall("sock-1", 42)

		hold := make(chan struct{})
		binder.set(func() { binder.hold = hold })
		var pendingErr error
		done := make(chan struct{})
		go func() {
			_, pendingErr = dq.BindUserEmailWait(context.Background(), "sock-2/a", 42)
			close(done)
		}()
		synctest.Wait()

		consumer.last(t).drop(errConnectionForced)
		pass(1)
		select {
		case <-dq.Rebound():
		default:
			t.Fatal("no signal after the reconnect")
		}
		expectBound(t, consumer.last(t), "after the reconnect", call)
		if consumer.last(t).bound(bind) || dq.HasBinding(bind) {
			t.Fatal("an Email binding is restored on the consumer channel")
		}
		expectEventReachesHub(t, dq, consumer.last(t))

		binder.set(func() { binder.hold = nil })
		close(hold)
		<-done
		if !errors.Is(pendingErr, errQueueChanged) && !errors.Is(pendingErr, errBindingGone) {
			t.Fatalf("a bind across the reconnect: %v", pendingErr)
		}
	})
}

// A subscribe confirmed on the new queue while the reconnect still runs keeps its binding,
// so its cleanup removes it from the broker.
func TestDomainQueueEmailBindDuringReconnect(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		dq, consumer, binder := emailQueue(t)
		hold := make(chan struct{})
		consumer.set(func() { consumer.hold = hold })
		consumer.last(t).drop(errConnectionForced)
		pass(1) // the new queue is declared, the reconnect waits on its first bind

		bind, err := dq.BindUserEmailWait(context.Background(), "sock-1/a", 42)
		if err != nil {
			t.Fatal(err)
		}
		consumer.set(func() { consumer.hold = nil })
		close(hold)
		synctest.Wait()
		select {
		case <-dq.Rebound():
		default:
			t.Fatal("no signal after the reconnect")
		}
		if !dq.HasBinding(bind) {
			t.Fatal("the binding on the new queue is forgotten")
		}

		dq.UnbindEmail(bind)
		synctest.Wait()
		if _, unbinds := binderLog(binder); !slices.Contains(unbinds, emailBind("sock-1/a")) {
			t.Fatalf("not removed from the broker: %+v", unbinds)
		}
	})
}

// An unbind is never dropped: not for many at once, not after a broker failure; one for an
// older queue goes without a command.
func TestDomainQueueEmailUnbind(t *testing.T) {
	register := func(dq *DomainQueue, n int) []*model.BindQueueEvent {
		out := make([]*model.BindQueueEvent, n)
		dq.Lock()
		for i := range out {
			out[i] = &model.BindQueueEvent{Id: fmt.Sprintf("sock-%d/a", i), Exchange: model.EmailServiceExchange, Routing: "email.workspace.*.*.v1.1.42"}
			dq.bindings[bindKey(out[i])] = out[i]
		}
		dq.Unlock()

		return out
	}
	pending := func(dq *DomainQueue) int {
		dq.RLock()
		defer dq.RUnlock()

		return len(dq.emailUnbinds)
	}

	t.Run("many while the binder is busy", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			dq, _, binder := emailQueue(t)
			hold := make(chan struct{})
			binder.set(func() { binder.hold = hold })
			go func() { _, _ = dq.BindUserEmailWait(context.Background(), "busy/a", 42) }()
			synctest.Wait()

			for _, b := range register(dq, emailOpsSize+500) {
				dq.UnbindEmail(b)
			}
			binder.set(func() { binder.hold = nil })
			close(hold)
			synctest.Wait()
			if _, unbinds := binderLog(binder); len(unbinds) != emailOpsSize+500 || pending(dq) != 0 {
				t.Fatalf("unbinds %d, pending %d", len(unbinds), pending(dq))
			}
		})
	})

	t.Run("after a broker failure", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			dq, _, binder := emailQueue(t)
			binder.set(func() { binder.down = true })
			dq.UnbindEmail(register(dq, 1)[0])
			synctest.Wait()
			if pending(dq) != 1 {
				t.Fatal("a failed unbind is dropped")
			}

			binder.set(func() { binder.down = false })
			time.Sleep(2 * emailUnbindRetry)
			synctest.Wait()
			if _, unbinds := binderLog(binder); len(unbinds) != 1 || pending(dq) != 0 {
				t.Fatalf("unbinds %d, pending %d", len(unbinds), pending(dq))
			}
		})
	})

	t.Run("an older queue", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			dq, _, binder := emailQueue(t)
			b := register(dq, 1)[0]
			dq.Lock()
			delete(dq.bindings, bindKey(b))
			dq.emailUnbinds[bindKey(b)] = emailUnbind{b: b, queue: "engine.ws.old.1"}
			dq.Unlock()
			dq.startEmailBinder()
			dq.emailKick <- struct{}{}
			synctest.Wait()
			if _, unbinds := binderLog(binder); len(unbinds) != 0 || pending(dq) != 0 {
				t.Fatalf("unbinds %d, pending %d", len(unbinds), pending(dq))
			}
		})
	})
}
