//go:build integration

package rabbit

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/webitel/engine/model"
	"github.com/webitel/engine/mq"
	"github.com/webitel/wlog"
)

// Needs RabbitMQ 4: EMAIL_TEST_AMQP_URL=amqp://guest:guest@localhost:5672/ go test -tags=integration ./mq/rabbit/
func emailTestConn(t *testing.T) *amqp.Connection {
	t.Helper()
	url := os.Getenv("EMAIL_TEST_AMQP_URL")
	if url == "" {
		t.Fatal("EMAIL_TEST_AMQP_URL is required for the integration tests")
	}
	wlog.InitGlobalLogger(wlog.NewLogger(&wlog.LoggerConfiguration{EnableConsole: false}))
	conn, err := amqp.Dial(url)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	return conn
}

func emailChannel(t *testing.T, conn *amqp.Connection) *amqp.Channel {
	t.Helper()
	ch, err := conn.Channel()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = ch.Close() })

	return ch
}

func waitFor(t *testing.T, what string, ok func() bool) {
	t.Helper()
	for deadline := time.Now().Add(15 * time.Second); time.Now().Before(deadline); time.Sleep(50 * time.Millisecond) {
		if ok() {
			return
		}
	}
	t.Fatalf("timed out: %s", what)
}

// emailNode is the domain queue of domain 1 on one engine node.
func emailNode(t *testing.T, conn *amqp.Connection) *DomainQueue {
	t.Helper()
	client := &AMQP{log: wlog.NewLogger(&wlog.LoggerConfiguration{}), connection: conn,
		domainQueues: map[int64]mq.DomainQueue{}}
	dq := newDomainQueue(client, 1, nil).(*DomainQueue)
	dq.Start()
	t.Cleanup(dq.Stop)

	return dq
}

func publishEmail(t *testing.T, ch *amqp.Channel, id string, user int64) {
	t.Helper()
	body := fmt.Sprintf(`{"event_id":"%s","event_type":"email.workspace.thread.changed","schema_version":1,
"occurred_at":"2026-11-06T09:00:00Z","domain_id":1,"user_id":%d,"thread_id":1024,"access":"current",
"workspace_version":17,"projection_version":905,"message_id":null,"changes":["projection"]}`, id, user)
	key := model.EmailRoutingKey(model.EmailWorkspaceThreadChanged, 1, user)
	if err := ch.Publish(model.EmailServiceExchange, key, false, false, amqp.Publishing{
		ContentType: model.EmailContentType, Body: []byte(body)}); err != nil {
		t.Fatal(err)
	}
}

func expectEmail(t *testing.T, dq *DomainQueue, id string) {
	t.Helper()
	select {
	case e := <-dq.EmailEvents():
		if e.EventId != id {
			t.Fatalf("event %s, want %s", e.EventId, id)
		}
	case <-time.After(10 * time.Second):
		t.Fatalf("no event %s", id)
	}
}

func expectNoEmail(t *testing.T, dq *DomainQueue) {
	t.Helper()
	select {
	case e := <-dq.EmailEvents():
		t.Fatalf("unexpected event %s", e.EventId)
	case <-time.After(500 * time.Millisecond):
	}
}

// barrier waits until the binder of the node removed every binding it was asked to.
func barrier(t *testing.T, dq *DomainQueue) {
	t.Helper()
	waitFor(t, "unbinds done", func() bool {
		dq.RLock()
		defer dq.RUnlock()

		return len(dq.emailUnbinds) == 0
	})
}

// Two tabs on node a and one on node b: each node gets one copy for its sockets; a closed
// tab keeps the other; a node that lost its channel binds again and says so.
func TestEmailNodeSubscriptions(t *testing.T) {
	// Every connection first: emailTestConn sets the global logger the nodes use.
	conn := emailTestConn(t)
	connA := emailTestConn(t)
	connB := emailTestConn(t)
	ch := emailChannel(t, conn)
	for _, e := range []string{model.AppExchange, model.EventExchange} {
		if err := ch.ExchangeDeclare(e, "topic", true, false, false, false, nil); err != nil {
			t.Fatal(err)
		}
	}
	// The binder of a node declares email.events.
	if err := ch.ExchangeDelete(model.EmailServiceExchange, false, false); err != nil {
		t.Fatal(err)
	}
	a, b := emailNode(t, connA), emailNode(t, connB)
	ctx := context.Background()
	a1, err := a.BindUserEmailWait(ctx, "sock-a1/g1", 42)
	if err != nil {
		t.Fatal(err)
	}
	a2, err := a.BindUserEmailWait(ctx, "sock-a2/g1", 42)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := b.BindUserEmailWait(ctx, "sock-b1/g1", 42); err != nil {
		t.Fatal(err)
	}

	const first, second, third, fourth = "00000000-0000-4000-8000-000000000501", "00000000-0000-4000-8000-000000000502",
		"00000000-0000-4000-8000-000000000503", "00000000-0000-4000-8000-000000000504"
	publishEmail(t, ch, first, 42)
	expectEmail(t, a, first)
	expectEmail(t, b, first)
	expectNoEmail(t, a) // one copy per node, not per tab
	publishEmail(t, ch, "00000000-0000-4000-8000-000000000509", 43)
	expectNoEmail(t, b) // the key of another user

	a.UnbindEmail(a1)
	barrier(t, a)
	publishEmail(t, ch, second, 42)
	expectEmail(t, a, second)
	expectEmail(t, b, second)

	a.UnbindEmail(a2)
	barrier(t, a)
	publishEmail(t, ch, third, 42)
	expectEmail(t, b, third)
	expectNoEmail(t, a)

	// The same socket subscribes again under a new id; the late cleanup of the old one
	// does not touch it.
	again, err := a.BindUserEmailWait(ctx, "sock-a2/g2", 42)
	if err != nil {
		t.Fatal(err)
	}
	a.UnbindEmail(a2)
	barrier(t, a)
	const fifth = "00000000-0000-4000-8000-000000000505"
	publishEmail(t, ch, fifth, 42)
	expectEmail(t, a, fifth)
	expectEmail(t, b, fifth)
	a.UnbindEmail(again)

	// After a reconnect the node has no Email bindings; the socket subscribes again.
	_ = b.getChannel().Close()
	select {
	case <-b.Rebound():
	case <-time.After(3 * RECONNECT_SEC * time.Second):
		t.Fatal("no signal after the reconnect")
	}
	if _, err := b.BindUserEmailWait(ctx, "sock-b1/g2", 42); err != nil {
		t.Fatal(err)
	}
	publishEmail(t, ch, fourth, 42)
	expectEmail(t, b, fourth)
}
