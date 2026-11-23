package app

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/webitel/engine/model"
	"github.com/webitel/engine/pkg/wbt/auth_manager"
	"github.com/webitel/wlog"
)

// Email events of Workspace (EM-26, docs/email-events.md "WebSocket"). The hub never waits
// for the network: a worker of every subscribed socket checks the session and the read
// right, then hands the event back to the hub, the only writer of conn.Send.

const (
	emailQueueSize   = 64
	emailSeenSize    = 256
	emailOutSize     = 1024
	emailBindTimeout = 5 * time.Second
	emailAuthTimeout = 5 * time.Second
)

// The pause after a failed session check doubles from emailAuthRetry to emailAuthRetryMax.
var (
	emailAuthRetry    = 5 * time.Second
	emailAuthRetryMax = 60 * time.Second
)

type emailSubscription struct {
	conn     *WebConn
	bind     atomic.Pointer[model.BindQueueEvent]
	in       chan *model.EmailEvent
	kick     chan struct{}
	overflow atomic.Bool
	ended    atomic.Bool
	stop     chan struct{}
	stopOnce sync.Once
	seen     emailSeen // used by the hub alone
}

// emailOut is a message of a worker for the hub; id is empty for resync_required, which
// goes even after the subscription ended.
type emailOut struct {
	sub *emailSubscription
	msg *model.WebSocketEvent
	id  string
}

// emailSeen keeps the last ids sent to one socket.
type emailSeen struct {
	ids  map[string]struct{}
	ring []string
	next int
}

func (s *emailSeen) has(id string) bool {
	_, ok := s.ids[id]

	return ok
}

func (s *emailSeen) add(id string) {
	if s.ids == nil {
		s.ids, s.ring = make(map[string]struct{}, emailSeenSize), make([]string, emailSeenSize)
	}
	delete(s.ids, s.ring[s.next])
	s.ring[s.next] = id
	s.ids[id] = struct{}{}
	s.next = (s.next + 1) % emailSeenSize
}

// SubscribeEmail subscribes the socket to the Email events of its own session; it returns
// after the broker confirmed the binding.
func (a *App) SubscribeEmail(ctx context.Context, conn *WebConn, session *auth_manager.Session) model.AppError {
	if !a.config.EmailWorkspace {
		return model.NewBadRequestError("app.email.subscribe.disabled", "Email events are not enabled")
	}
	permission := session.GetPermission(model.PERMISSION_SCOPE_EMAIL_THREAD)
	if !permission.CanRead() {
		return a.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}
	if session.UserId != conn.UserId || session.DomainId != conn.DomainId {
		return model.NewForbiddenError("app.email.subscribe.identity", "the session belongs to another user")
	}
	h, err := a.GetHubById(session.DomainId)
	if err != nil {
		return err
	}

	return h.subscribeEmail(ctx, conn)
}

// UnsubscribeEmail ends the subscription of this socket only.
func (a *App) UnsubscribeEmail(conn *WebConn) model.AppError {
	h, err := a.GetHubById(conn.DomainId)
	if err != nil {
		return err
	}
	conn.emailMx.Lock()
	defer conn.emailMx.Unlock()
	if sub := conn.emailSub.Load(); sub != nil {
		h.endEmail(sub)
	}

	return nil
}

func (wh *Hub) subscribeEmail(ctx context.Context, conn *WebConn) model.AppError {
	conn.emailMx.Lock()
	defer conn.emailMx.Unlock()
	if conn.emailSub.Load() != nil {
		return nil
	}

	// Registered first: an event that comes with the binding is not lost.
	sub := &emailSubscription{conn: conn, in: make(chan *model.EmailEvent, emailQueueSize),
		kick: make(chan struct{}, 1), stop: make(chan struct{})}
	conn.emailSub.Store(sub)
	go wh.runEmail(sub)

	ctx, cancel := context.WithTimeout(ctx, emailBindTimeout)
	defer cancel()
	// Every subscription binds under its own id: the cleanup of an older one never
	// removes a newer one of the same socket.
	b, err := wh.domainQueue.BindUserEmailWait(ctx, conn.Id()+"/"+model.NewId()[:8], conn.UserId)
	if err != nil {
		wh.endEmail(sub)

		return model.NewInternalError("app.email.subscribe.bind", err.Error())
	}
	sub.bind.Store(b)
	if sub.ended.Load() {
		// The socket closed meanwhile; whoever takes the binding removes it.
		if b := sub.bind.Swap(nil); b != nil {
			wh.domainQueue.UnbindEmail(b)
		}

		return model.NewInternalError("app.email.subscribe.closed", "the socket closed")
	}

	return nil
}

// endEmail stops the worker and removes the binding of this socket without waiting for the
// broker; it may run more than once.
func (wh *Hub) endEmail(sub *emailSubscription) {
	sub.ended.Store(true)
	sub.conn.emailSub.CompareAndSwap(sub, nil)
	sub.stopOnce.Do(func() { close(sub.stop) })
	if b := sub.bind.Swap(nil); b != nil {
		wh.domainQueue.UnbindEmail(b)
	}
}

// dispatchEmail hands the event to the worker of every subscribed socket of the user; a full
// worker queue loses the event and asks for a resync.
func (wh *Hub) dispatchEmail(connections *hubConnectionIndex, e *model.EmailEvent) {
	for _, webCon := range connections.ForUser(e.UserId) {
		sub := webCon.emailSub.Load()
		if sub == nil || webCon.DomainId != e.DomainId {
			continue
		}
		select {
		case sub.in <- e:
		default:
			if !sub.overflow.Swap(true) {
				webCon.Log().Warn("webhub.email: queue is full, resync", wlog.String("event_type", "email"))
			}
			select {
			case sub.kick <- struct{}{}:
			default:
			}
		}
	}
}

// sendEmail writes a checked event to the socket; a full socket is closed as for the other
// events. An id goes once to a socket, counted only once it is in the send buffer.
func (wh *Hub) sendEmail(connections *hubConnectionIndex, out emailOut) {
	conn := out.sub.conn
	if !connections.Has(conn) {
		return
	}
	if out.id != "" && (conn.emailSub.Load() != out.sub || out.sub.seen.has(out.id)) {
		return
	}
	select {
	case conn.Send <- out.msg:
		if out.id != "" {
			out.sub.seen.add(out.id)
		}
	default:
		conn.Log().Error("webhub.broadcast.email: cannot send, closing websocket", wlog.String("event_type", "email"))
		close(conn.Send)
		connections.Remove(conn)
	}
}

// emailRebound: the node queue came back after a loss with no Email bindings, so the events
// in between are lost. Every subscription ends; the client subscribes again.
func (wh *Hub) emailRebound(connections *hubConnectionIndex) {
	ended := 0
	for webCon := range connections.All() {
		if sub := webCon.emailSub.Load(); sub != nil {
			wh.endEmail(sub)
			wh.sendEmail(connections, emailOut{sub: sub, msg: model.NewEmailResyncEvent(model.EmailResyncSubscriptionLost)})
			ended++
		}
	}
	if ended > 0 {
		wh.log.Warn("webhub.email: node queue was lost, subscriptions ended, clients resync", wlog.Int("subscriptions", ended))
	}
}

type emailVerdict int

const (
	emailAllow    emailVerdict = iota
	emailDeny                  // auth is unavailable: nothing goes, a resync follows
	emailRevoked               // the session ended or lost the read right
	emailIdentity              // the token now names another user
)

// emailAuth is the last check of the session of one socket.
type emailAuth struct {
	session *auth_manager.Session
	checked time.Time
	retryAt time.Time
	pause   time.Duration
	lost    bool
}

// access checks the session at most once per auth_cache_expire and when it expires; an
// unavailable auth never lets an event through.
func (wh *Hub) emailAccess(sub *emailSubscription, a *emailAuth) emailVerdict {
	now := time.Now()
	ttl := time.Duration(max(wh.app.config.AuthCacheExpire, 1)) * time.Second
	// Expire is in milliseconds, as WebConn reads it.
	if a.session != nil && now.Sub(a.checked) < ttl && (a.session.Expire == 0 || a.session.Expire > model.GetMillis()) {
		return emailAllow
	}
	if now.Before(a.retryAt) {
		return emailDeny
	}

	ctx, cancel := context.WithTimeout(context.Background(), emailAuthTimeout)
	defer cancel()
	s, err := wh.app.GetSessionWitchContext(ctx, sub.conn.GetSessionToken())
	switch {
	case err != nil && err.GetStatusCode() >= 500:
		sub.conn.Log().Warn("webhub.email: session check failed, events wait", wlog.Err(err))
		a.pause = min(max(2*a.pause, emailAuthRetry), emailAuthRetryMax)
		a.session, a.retryAt, a.lost = nil, now.Add(a.pause), true

		return emailDeny
	case err != nil:
		return emailRevoked
	case s.UserId != sub.conn.UserId || s.DomainId != sub.conn.DomainId:
		return emailIdentity
	case !s.GetPermission(model.PERMISSION_SCOPE_EMAIL_THREAD).CanRead():
		return emailRevoked
	}
	a.session, a.checked, a.pause = s, now, 0

	return emailAllow
}

// runEmail delivers the events of one socket. While auth is unavailable a timer checks the
// session again, so the resync comes also when no other event does.
func (wh *Hub) runEmail(sub *emailSubscription) {
	var auth emailAuth
	retry := time.NewTimer(time.Hour)
	retry.Stop()
	defer retry.Stop()
	retrying := false
	for {
		var e *model.EmailEvent
		recheck := false
		select {
		case <-sub.stop:
			return
		case <-wh.stop:
			return
		case <-sub.kick:
		case e = <-sub.in:
		case <-retry.C:
			retrying, recheck = false, true
		}
		if sub.overflow.Swap(false) && !wh.emitEmail(sub, model.NewEmailResyncEvent(model.EmailResyncOverflow), "") {
			return
		}
		if e == nil && !recheck {
			continue
		}

		switch wh.emailAccess(sub, &auth) {
		case emailDeny:
			if !retrying {
				retry.Reset(time.Until(auth.retryAt))
				retrying = true
			}

			continue
		case emailRevoked:
			sub.conn.Log().Info("webhub.email: session may not read Email, subscription ended")
			wh.emitEmail(sub, model.NewEmailResyncEvent(model.EmailResyncSubscriptionLost), "")
			wh.endEmail(sub)

			return
		case emailIdentity:
			sub.conn.Log().Warn("webhub.email: session of another user, closing websocket")
			wh.endEmail(sub)
			_ = sub.conn.WebSocket.Close()

			return
		}
		if auth.lost {
			auth.lost = false
			if !wh.emitEmail(sub, model.NewEmailResyncEvent(model.EmailResyncSubscriptionLost), "") {
				return
			}
		}
		if e != nil && !wh.emitEmail(sub, e.WebSocketEvent(), e.EventId) {
			return
		}
	}
}

func (wh *Hub) emitEmail(sub *emailSubscription, msg *model.WebSocketEvent, id string) bool {
	select {
	case wh.emailOut <- emailOut{sub: sub, msg: msg, id: id}:
		return true
	case <-sub.stop:
		return false
	case <-wh.stop:
		return false
	}
}
