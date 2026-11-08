package app

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/webitel/engine/model"
	"github.com/webitel/engine/mq"
	"github.com/webitel/engine/pkg/wbt/auth_manager"
	"github.com/webitel/engine/store"
	"github.com/webitel/wlog"
)

// fakeEmailQueue is the node queue of domain 1: the test pushes events and reconnects.
type fakeEmailQueue struct {
	mq.DomainQueue
	email   chan *model.EmailEvent
	chat    chan *model.ChatEvent
	rebound chan struct{}

	mu      sync.Mutex
	bound   map[string]bool
	binds   int
	unbinds []string
	bindErr error
}

func newFakeEmailQueue() *fakeEmailQueue {
	return &fakeEmailQueue{email: make(chan *model.EmailEvent), chat: make(chan *model.ChatEvent),
		rebound: make(chan struct{}), bound: map[string]bool{}}
}

func (q *fakeEmailQueue) Start()                                              {}
func (q *fakeEmailQueue) Stop()                                               {}
func (q *fakeEmailQueue) Events() <-chan *model.WebSocketEvent                { return nil }
func (q *fakeEmailQueue) CallEvents() <-chan *model.CallEvent                 { return nil }
func (q *fakeEmailQueue) UserStateEvents() <-chan *model.UserState            { return nil }
func (q *fakeEmailQueue) NotificationEvents() <-chan *model.Notification      { return nil }
func (q *fakeEmailQueue) ChatEvents() <-chan *model.ChatEvent                 { return q.chat }
func (q *fakeEmailQueue) EmailEvents() <-chan *model.EmailEvent           { return q.email }
func (q *fakeEmailQueue) Rebound() <-chan struct{}                            { return q.rebound }
func (q *fakeEmailQueue) BulkUnbind(_ []*model.BindQueueEvent) model.AppError { return nil }

func (q *fakeEmailQueue) BindUserEmailWait(_ context.Context, id string, userId int64) (*model.BindQueueEvent, error) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.bindErr != nil {
		return nil, q.bindErr
	}
	q.binds++
	q.bound[id] = true

	return &model.BindQueueEvent{Id: id, UserId: userId, Exchange: model.EmailServiceExchange,
		Routing: model.EmailBindingKey(1, userId)}, nil
}

func (q *fakeEmailQueue) HasBinding(b *model.BindQueueEvent) bool {
	q.mu.Lock()
	defer q.mu.Unlock()

	return q.bound[b.Id]
}

func (q *fakeEmailQueue) Unbind(b *model.BindQueueEvent) model.AppError {
	q.mu.Lock()
	defer q.mu.Unlock()
	delete(q.bound, b.Id)
	q.unbinds = append(q.unbinds, b.Id)

	return nil
}

func (q *fakeEmailQueue) UnbindEmail(b *model.BindQueueEvent) { _ = q.Unbind(b) }

func (q *fakeEmailQueue) state() (bound, unbound int) {
	q.mu.Lock()
	defer q.mu.Unlock()

	return len(q.bound), len(q.unbinds)
}

func (q *fakeEmailQueue) unbound(id string) bool {
	q.mu.Lock()
	defer q.mu.Unlock()

	return !q.bound[id]
}

func (q *fakeEmailQueue) bindCalls() int {
	q.mu.Lock()
	defer q.mu.Unlock()

	return q.binds
}

type fakeMQ struct {
	mq.MQ
	q *fakeEmailQueue
}

func (m *fakeMQ) NewDomainQueue(int64, model.GetAllBindings) (mq.DomainQueue, model.AppError) {
	return m.q, nil
}
func (m *fakeMQ) RegisterWebsocket(int64, *model.RegisterToWebsocketEvent) model.AppError { return nil }
func (m *fakeMQ) UnRegisterWebsocket(int64, *model.RegisterToWebsocketEvent) model.AppError {
	return nil
}

// fakeAuth answers by token; block holds every check until it is closed.
type fakeAuth struct {
	auth_manager.AuthManager
	mu       sync.Mutex
	sessions map[string]*auth_manager.Session
	err      error
	block    chan struct{}
}

func (f *fakeAuth) GetSession(ctx context.Context, token string) (*auth_manager.Session, error) {
	f.mu.Lock()
	block, err, s := f.block, f.err, f.sessions[token]
	f.mu.Unlock()
	if block != nil {
		select {
		case <-block:
		case <-ctx.Done():
			return nil, auth_manager.ErrInternal
		}
	}
	if err != nil {
		return nil, err
	}
	if s == nil {
		return nil, auth_manager.ErrStatusUnauthenticated
	}
	c := *s

	return &c, nil
}

func (f *fakeAuth) set(change func()) {
	f.mu.Lock()
	defer f.mu.Unlock()
	change()
}

type fakeStore struct{ store.Store }

func (fakeStore) SocketSession() store.SocketSessionStore { return fakeSockets{} }

type fakeSockets struct{ store.SocketSessionStore }

func (fakeSockets) Create(context.Context, model.SocketSession) model.AppError     { return nil }
func (fakeSockets) SetUpdatedAt(context.Context, string, time.Time) model.AppError { return nil }
func (fakeSockets) DeleteById(context.Context, string) model.AppError              { return nil }

func emailSession(token string, user int64, read bool) *auth_manager.Session {
	s := &auth_manager.Session{Id: token, Token: token, UserId: user, DomainId: 1, Expire: model.GetMillis() + 3600_000}
	if read {
		s.Scopes = []auth_manager.SessionPermission{{Name: model.PERMISSION_SCOPE_EMAIL_THREAD, Obac: true,
			Access: auth_manager.PERMISSION_ACCESS_READ.Value()}}
	}

	return s
}

type emailEnv struct {
	app  *App
	q    *fakeEmailQueue
	auth *fakeAuth
	srv  *httptest.Server
	hub  *Hub
}

func newEmailEnv(t *testing.T) *emailEnv {
	t.Helper()
	q := newFakeEmailQueue()
	auth := &fakeAuth{sessions: map[string]*auth_manager.Session{
		"tab-a": emailSession("tab-a", 42, true), "tab-b": emailSession("tab-b", 42, true),
		"other": emailSession("other", 43, true)}}
	a := &App{config: &model.Config{EmailWorkspace: true, AuthCacheExpire: 30}, ctx: context.Background(),
		Log: wlog.NewLogger(&wlog.LoggerConfiguration{}), tracer: NewTrace(), MessageQueue: &fakeMQ{q: q},
		sessionManager: auth, Store: fakeStore{}}
	a.Hubs = NewHubs(a)
	env := &emailEnv{app: a, q: q, auth: auth}
	env.hub, _ = a.GetHubById(1)
	up := websocket.Upgrader{}
	env.srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ws, err := up.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		token := r.URL.Query().Get("token")
		wc := a.NewWebConn(ws, *auth.sessions[token], "test", "1", "ua", "127.0.0.1")
		wc.DomainId = 1
		// Hub.Register, not App.HubRegister: that one reads conn.log while the hub sets it,
		// a race of the existing code outside the Email events.
		env.hub.Register(wc)
		sockets.Store(token, wc)
		wc.Pump()
	}))
	t.Cleanup(env.srv.Close)
	emailAuthRetry, emailAuthRetryMax = 50*time.Millisecond, 200*time.Millisecond

	return env
}

var sockets sync.Map

// tab is the client side of a socket: it keeps the "email" events and the chat ones.
type tab struct {
	t      *testing.T
	ws     *websocket.Conn
	conn   *WebConn
	events chan map[string]any
	closed chan struct{}
}

func (env *emailEnv) open(t *testing.T, token string) *tab {
	t.Helper()
	ws, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(env.srv.URL, "http")+"/?token="+token, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = ws.Close() })
	c := &tab{t: t, ws: ws, events: make(chan map[string]any, 1000), closed: make(chan struct{})}
	go func() {
		defer close(c.closed)
		for {
			var m struct {
				Event string         `json:"event"`
				Data  map[string]any `json:"data"`
			}
			if err := ws.ReadJSON(&m); err != nil {
				return
			}
			if m.Event == model.WebsocketEmailEvent || m.Event == model.WEBSOCKET_EVENT_CHAT {
				c.events <- m.Data
			}
		}
	}()
	for deadline := time.Now().Add(5 * time.Second); c.conn == nil; time.Sleep(5 * time.Millisecond) {
		if v, ok := sockets.Load(token); ok && v.(*WebConn).UserId != 0 {
			c.conn = v.(*WebConn)
		}
		if time.Now().After(deadline) {
			t.Fatal("no socket")
		}
	}
	sockets.Delete(token)

	return c
}

func (c *tab) subscribe(env *emailEnv) {
	c.t.Helper()
	if err := env.app.SubscribeEmail(context.Background(), c.conn, c.conn.GetSession()); err != nil {
		c.t.Fatal(err)
	}
}

func (c *tab) next() map[string]any {
	c.t.Helper()
	select {
	case m := <-c.events:
		return m
	case <-time.After(5 * time.Second):
		c.t.Fatal("no event")
	}

	return nil
}

func (c *tab) none() {
	c.t.Helper()
	select {
	case m := <-c.events:
		c.t.Fatalf("unexpected event %v", m)
	case <-time.After(200 * time.Millisecond):
	}
}

func emailEvent(id string, user int64) *model.EmailEvent {
	return &model.EmailEvent{EventId: id, DomainId: 1, UserId: user, ThreadId: 1024, EventType: model.EmailWorkspaceThreadChanged,
		Access: "current", WorkspaceVersion: 17, Changes: []string{"projection"}}
}

func (env *emailEnv) push(e *model.EmailEvent) {
	select {
	case env.q.email <- e:
	case <-time.After(5 * time.Second):
		panic("the hub does not read Email events")
	}
}

func TestEmailSubscriptionDelivery(t *testing.T) {
	env := newEmailEnv(t)
	a, b, other := env.open(t, "tab-a"), env.open(t, "tab-b"), env.open(t, "other")
	a.subscribe(env)
	a.subscribe(env) // a repeat binds nothing more
	b.subscribe(env)
	other.subscribe(env)
	if bound, _ := env.q.state(); bound != 3 || env.q.bindCalls() != 3 {
		t.Fatalf("bound %d, binds %d", bound, env.q.bindCalls())
	}

	env.push(emailEvent("e-1", 42))
	for _, c := range []*tab{a, b} {
		if m := c.next(); m["event_id"] != "e-1" || m["thread_id"] != "1024" || m["workspace_version"] != "17" {
			t.Fatalf("event %v", m)
		}
	}
	other.none()

	env.push(emailEvent("e-1", 42)) // a repeat goes nowhere
	a.none()

	if err := env.app.UnsubscribeEmail(b.conn); err != nil {
		t.Fatal(err)
	}
	env.push(emailEvent("e-2", 42))
	if m := a.next(); m["event_id"] != "e-2" {
		t.Fatalf("event %v", m)
	}
	b.none()

	// A new subscription of the same socket binds under a new id and gets the events.
	first := b.conn.Id()
	b.subscribe(env)
	if id := b.conn.emailSub.Load().bind.Load().Id; !strings.HasPrefix(id, first+"/") || env.q.unbound(id) {
		t.Fatalf("binding %s", id)
	}
	env.push(emailEvent("e-3", 42))
	for _, c := range []*tab{a, b} {
		if m := c.next(); m["event_id"] != "e-3" {
			t.Fatalf("event %v", m)
		}
	}

	_ = a.ws.Close()
	waitEmail(t, "binding of the closed socket removed", func() bool { _, unbound := env.q.state(); return unbound == 2 })
}

func TestEmailSubscribeChecks(t *testing.T) {
	env := newEmailEnv(t)
	a := env.open(t, "tab-a")

	env.app.config.EmailWorkspace = false
	if err := env.app.SubscribeEmail(context.Background(), a.conn, a.conn.GetSession()); err == nil || err.GetStatusCode() != 400 {
		t.Fatalf("disabled: %v", err)
	}
	env.app.config.EmailWorkspace = true

	if err := env.app.SubscribeEmail(context.Background(), a.conn, emailSession("tab-a", 42, false)); err == nil || err.GetStatusCode() != 403 {
		t.Fatalf("no read: %v", err)
	}
	if err := env.app.SubscribeEmail(context.Background(), a.conn, emailSession("other", 43, true)); err == nil || err.GetStatusCode() != 403 {
		t.Fatalf("other user: %v", err)
	}
	env.q.mu.Lock()
	env.q.bindErr = errors.New("broker refused")
	env.q.mu.Unlock()
	if err := env.app.SubscribeEmail(context.Background(), a.conn, a.conn.GetSession()); err == nil || a.conn.emailSub.Load() != nil {
		t.Fatalf("refused binding: %v", err)
	}
}

func TestEmailSessionChecks(t *testing.T) {
	t.Run("the read right is gone", func(t *testing.T) {
		env := newEmailEnv(t)
		a := env.open(t, "tab-a")
		a.subscribe(env)
		env.auth.set(func() { env.auth.sessions["tab-a"] = emailSession("tab-a", 42, false) })
		env.push(emailEvent("e-1", 42))
		if m := a.next(); m["type"] != "resync_required" || m["reason"] != "subscription_lost" {
			t.Fatalf("event %v", m)
		}
		waitEmail(t, "unbound", func() bool { _, unbound := env.q.state(); return unbound == 1 })
		env.push(emailEvent("e-2", 42))
		a.none()
	})

	// Nothing goes while auth is unavailable; once it is back, the resync comes even when
	// the lost event was the last one, and later events go again.
	t.Run("auth is unavailable", func(t *testing.T) {
		env := newEmailEnv(t)
		a := env.open(t, "tab-a")
		a.subscribe(env)
		env.auth.set(func() { env.auth.err = auth_manager.ErrInternal })
		env.push(emailEvent("e-1", 42))
		a.none()

		env.auth.set(func() { env.auth.err = nil })
		if m := a.next(); m["reason"] != "subscription_lost" {
			t.Fatalf("event %v", m)
		}
		a.none()
		env.push(emailEvent("e-2", 42))
		if m := a.next(); m["event_id"] != "e-2" {
			t.Fatalf("event %v", m)
		}
	})

	t.Run("the token names another user", func(t *testing.T) {
		env := newEmailEnv(t)
		a := env.open(t, "tab-a")
		a.subscribe(env)
		env.auth.set(func() { env.auth.sessions["tab-a"] = emailSession("tab-a", 43, true) })
		env.push(emailEvent("e-1", 42))
		select {
		case <-a.closed:
		case <-time.After(5 * time.Second):
			t.Fatal("the socket stays open")
		}
	})
}

// A worker waiting for auth neither stops the hub nor the other events; its full queue
// asks for a resync.
func TestEmailSlowAuthAndOverflow(t *testing.T) {
	env := newEmailEnv(t)
	a := env.open(t, "tab-a")
	a.subscribe(env)
	block := make(chan struct{})
	env.auth.set(func() { env.auth.block = block })

	for i := range emailQueueSize + 5 {
		env.push(emailEvent("e-"+string(rune('A'+i%26))+string(rune('a'+i/26)), 42))
	}
	select {
	case env.q.chat <- &model.ChatEvent{UserId: 42, Event: "message", Data: map[string]any{}}:
	case <-time.After(5 * time.Second):
		t.Fatal("the hub is blocked")
	}
	if m := a.next(); m["action"] != "message" {
		t.Fatalf("chat event %v", m)
	}

	env.auth.set(func() { env.auth.block = nil })
	close(block)
	got := 0
	resync := false
	for deadline := time.After(5 * time.Second); got < emailQueueSize; {
		select {
		case m := <-a.events:
			if m["type"] == "resync_required" && m["reason"] == "overflow" {
				resync = true
			} else {
				got++
			}
		case <-deadline:
			t.Fatalf("got %d events", got)
		}
	}
	if !resync {
		t.Fatal("no resync after the overflow")
	}
}

// After the node queue came back every subscription ends with subscription_lost.
func TestEmailRebound(t *testing.T) {
	env := newEmailEnv(t)
	a, b := env.open(t, "tab-a"), env.open(t, "tab-b")
	a.subscribe(env)
	b.subscribe(env)

	env.q.rebound <- struct{}{}
	for _, c := range []*tab{a, b} {
		if m := c.next(); m["reason"] != "subscription_lost" {
			t.Fatalf("event %v", m)
		}
	}
	waitEmail(t, "both ended", func() bool { return a.conn.emailSub.Load() == nil && b.conn.emailSub.Load() == nil })
	a.subscribe(env)
	env.push(emailEvent("e-1", 42))
	if m := a.next(); m["event_id"] != "e-1" {
		t.Fatalf("event %v", m)
	}
	b.none()
}

// A full socket buffer closes the socket, as for the other events.
func TestEmailFullSocket(t *testing.T) {
	wh := &Hub{}
	conn := &WebConn{Send: make(chan model.WebSocketMessage), log: wlog.NewLogger(&wlog.LoggerConfiguration{})}
	sub := &emailSubscription{conn: conn}
	conn.emailSub.Store(sub)
	connections := newHubConnectionIndex()
	connections.Add(conn)
	wh.sendEmail(connections, emailOut{sub: sub, msg: emailEvent("e-1", 42).WebSocketEvent(), id: "e-1"})
	if connections.Has(conn) {
		t.Fatal("the full socket stays")
	}
	if _, ok := <-conn.Send; ok {
		t.Fatal("send buffer open")
	}
}

func waitEmail(t *testing.T, what string, ok func() bool) {
	t.Helper()
	for deadline := time.Now().Add(5 * time.Second); time.Now().Before(deadline); time.Sleep(10 * time.Millisecond) {
		if ok() {
			return
		}
	}
	t.Fatalf("timed out: %s", what)
}
