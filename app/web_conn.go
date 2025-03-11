package app

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	"github.com/nicksnyder/go-i18n/i18n"
	"github.com/webitel/engine/auth_manager"
	"github.com/webitel/engine/model"
	wlog "github.com/webitel/wlog"
)

const (
	SEND_QUEUE_SIZE    = 256
	SEND_DEADLOCK_WARN = (SEND_QUEUE_SIZE * 95) / 100
	WRITE_WAIT         = 10 * time.Second
	PONG_WAIT          = 60 * time.Second
	PING_PERIOD        = (PONG_WAIT * 9) / 10
	AUTH_TIMEOUT       = 20 * time.Second
)

var (
	spamMessage = []byte{0x0, 0x0, 0x0, 0x0}
)

type WebConn struct {
	id                 string
	sessionExpiresAt   int64 // This should stay at the top for 64-bit alignment of 64-bit words accessed atomically
	App                *App
	WebSocket          *websocket.Conn
	sessionToken       atomic.Value
	session            atomic.Value
	LastUserActivityAt int64
	UserId             int64
	DomainId           int64
	T                  i18n.TranslateFunc
	Locale             string
	Send               chan model.WebSocketMessage
	Sequence           int64
	closeOnce          sync.Once
	endWritePump       chan struct{}
	pumpFinished       chan struct{}
	listenEvents       map[string]*model.BindQueueEvent
	mx                 sync.RWMutex
	ip                 string
	lastLatencyTime    atomic.Int64
	log                *wlog.Logger
	logMx              sync.RWMutex
	Ctx                context.Context
	Span               trace.Span

	//Sip *SipProxy
}

func (a *App) NewWebConn(ws *websocket.Conn, session auth_manager.Session, t i18n.TranslateFunc, locale string, ip string) *WebConn {

	id := model.NewId()
	log := a.Log.With(
		wlog.Namespace("context"),
		wlog.String("ip_address", ip),
		wlog.String("protocol", "wss"),
		wlog.String("sock_id", id),
	)
	// Disable TCP_NO_DELAY for higher throughput
	var tcpConn *net.TCPConn
	switch conn := ws.NetConn().(type) {
	case *net.TCPConn:
		tcpConn = conn
	case *tls.Conn:
		newConn, ok := conn.NetConn().(*net.TCPConn)
		if ok {
			tcpConn = newConn
		}
	}

	if tcpConn != nil {
		err := tcpConn.SetNoDelay(false)
		if err != nil {
			log.Warn("Error in setting NoDelay socket opts", wlog.Err(err))
		}
	}

	wc := &WebConn{
		id:                 id,
		App:                a,
		WebSocket:          ws,
		Send:               make(chan model.WebSocketMessage, SEND_QUEUE_SIZE),
		LastUserActivityAt: model.GetMillis(),
		UserId:             session.UserId,
		T:                  t,
		Locale:             locale,
		endWritePump:       make(chan struct{}),
		pumpFinished:       make(chan struct{}),
		listenEvents:       make(map[string]*model.BindQueueEvent),
		ip:                 ip,
		log:                log,
	}
	wc.Ctx, wc.Span = a.Tracer().Start(context.Background(), "websocket")
	wc.Span.SetAttributes(
		attribute.Int64("domain_id", session.DomainId),
		attribute.Int64("user_id", session.UserId),
		attribute.String("ip_address", ip),
		attribute.String("sock_id", wc.id),
	)

	//wc.Sip = NewSipProxy(wc)

	wc.SetSession(&session)
	wc.SetSessionToken(session.Token)
	wc.SetSessionExpiresAt(session.Expire)

	return wc
}

func (wc *WebConn) SetLog(l *wlog.Logger) {
	wc.logMx.Lock()
	wc.log = l
	wc.logMx.Unlock()
}

func (wc *WebConn) Log() *wlog.Logger {
	wc.logMx.RLock()
	l := wc.log
	wc.logMx.RUnlock()
	return l
}

func (wc *WebConn) Id() string {
	return wc.id
}

func (wc *WebConn) Ip() string {
	return wc.ip
}

func (wc *WebConn) SetLastLatencyTime(new int64) int64 {
	t := wc.lastLatencyTime.Load()
	wc.lastLatencyTime.Store(new)
	return t
}

func (wc *WebConn) Close() {
	wc.WebSocket.Close()
	wc.closeOnce.Do(func() {
		close(wc.endWritePump)
	})
	<-wc.pumpFinished
}

func (c *WebConn) Pump() {
	ch := make(chan struct{})
	go func() {
		c.writePump()
		close(ch)
	}()
	c.readPump()
	c.closeOnce.Do(func() {
		close(c.endWritePump)
	})

	<-ch
	c.App.HubUnregister(c)
	close(c.pumpFinished)
}

func (c *WebConn) readPump() {
	defer func() {
		c.Log().Debug("websocket.read: close")
		c.WebSocket.Close()

		if c.Span != nil {
			c.Span.End()
		}
	}()

	c.WebSocket.SetReadLimit(int64(c.App.MaxSocketInboundMsgSize()))
	c.WebSocket.SetReadDeadline(time.Now().Add(PONG_WAIT))
	c.WebSocket.SetPongHandler(func(string) error {
		c.WebSocket.SetReadDeadline(time.Now().Add(PONG_WAIT))
		return nil
	})

	for {
		msgType, rd, err := c.WebSocket.NextReader()
		if err != nil {
			c.Log().Debug("websocket.NextReader error", wlog.Err(err))
			return
		}

		var decoder interface {
			Decode(v any) error
		}

		if msgType == websocket.TextMessage {
			decoder = json.NewDecoder(rd)
		} else {
			c.Log().Error("receive bad type message")
			continue
		}

		var req model.WebSocketRequest

		if err = decoder.Decode(&req); err != nil {
			c.Log().Error("decode message error", wlog.Err(err))
			continue
		}

		c.App.Srv.WebSocketRouter.ServeWebSocket(c, &req)
	}
}

// writeMessageBuf is a helper utility that wraps the write to the socket
// along with setting the write deadline.
func (c *WebConn) writeMessageBuf(msgType int, data []byte) error {
	c.WebSocket.SetWriteDeadline(time.Now().Add(WRITE_WAIT))
	return c.WebSocket.WriteMessage(msgType, data)
}

func (c *WebConn) writePump() {
	ticker := time.NewTicker(PING_PERIOD)
	authTicker := time.NewTicker(AUTH_TIMEOUT)

	defer func() {
		ticker.Stop()
		authTicker.Stop()
		c.WebSocket.Close()
	}()

	var buf bytes.Buffer
	buf.Grow(1024 * 2)

	enc := json.NewEncoder(&buf)

	for {
		select {
		case msg, ok := <-c.Send:

			if !ok {
				c.writeMessageBuf(websocket.CloseMessage, []byte{})
				return
			}
			evt, evtOk := msg.(*model.WebSocketEvent)

			buf.Reset()
			var err error

			if evtOk {
				cpyEvt := &model.WebSocketEvent{}
				*cpyEvt = *evt
				cpyEvt.Sequence = c.Sequence
				err = enc.Encode(cpyEvt)
				c.Sequence++
			} else {
				err = enc.Encode(msg)
			}

			if err != nil {
				c.Log().Debug("encoding websocket message", wlog.Err(err))
				continue
			}

			if len(c.Send) >= SEND_DEADLOCK_WARN {
				c.Log().Warn("websocket.full", wlog.String("event", msg.EventType()), wlog.Int("length", buf.Len()))
			}

			if err = c.writeMessageBuf(websocket.TextMessage, buf.Bytes()); err != nil {
				c.Log().Error("send message", wlog.Err(err))
				return
			}

		case <-ticker.C:
			if err := c.writeMessageBuf(websocket.PingMessage, []byte{}); err != nil {
				c.Log().Error("ping message", wlog.Err(err))
				return
			} else if c.App.config.Cloudflare {
				c.WebSocket.WriteMessage(websocket.TextMessage, spamMessage)
			}
		case <-c.endWritePump:
			return
		case <-authTicker.C:
			if c.GetSessionToken() == "" {
				c.Log().Debug("websocket.authTicker: did not authenticate")
				return
			}
			authTicker.Stop()
		}
	}
}

func (webCon *WebConn) SendHello() {
	sess := webCon.GetSession()
	msg := model.NewWebSocketEvent(model.WEBSOCKET_EVENT_HELLO)
	msg.Add("server_node_id", webCon.App.nodeId)
	msg.Add("server_build_commit", model.BuildNumber)
	msg.Add("server_version", model.CurrentVersion)
	msg.Add("server_time", model.GetMillis())
	msg.Add("sock_id", webCon.id)
	msg.Add("session", sess)
	if webCon.App.b2b != nil {
		msg.Add("b2bua", true)
	}
	if webCon.App.config.PingClientInterval > 0 {
		msg.Add("ping_interval", webCon.App.config.PingClientInterval)
	}

	if sess.HasChatLicense() {
		msg.Add("use_chat", true)
	}

	if sess.HasCallCenterLicense() {
		msg.Add("use_cc", true)
	}

	webCon.Send <- msg
}

func (webCon *WebConn) SendError(err model.AppError) {
	webCon.log.Error(err.Error(),
		wlog.Err(err),
	)
	msg := model.NewWebSocketEvent(model.WebsocketError)
	msg.Add("sock_id", webCon.id)
	msg.Add("error", err)
	webCon.Send <- msg
}

func (c *WebConn) GetSessionExpiresAt() int64 {
	return atomic.LoadInt64(&c.sessionExpiresAt)
}

func (c *WebConn) SetSessionExpiresAt(v int64) {
	atomic.StoreInt64(&c.sessionExpiresAt, v)
}

func (c *WebConn) GetSessionToken() string {
	return c.sessionToken.Load().(string)
}

func (c *WebConn) SetSessionToken(v string) {
	c.sessionToken.Store(v)
}

func (c *WebConn) GetSession() *auth_manager.Session {
	return c.session.Load().(*auth_manager.Session)
}

func (c *WebConn) SetSession(v *auth_manager.Session) {
	c.session.Store(v)
}

func (webCon *WebConn) IsAuthenticated() bool {
	// Check the expiry to see if we need to check for a new session
	if webCon.GetSessionExpiresAt() < model.GetMillis() {
		if webCon.GetSessionToken() == "" {
			return false
		}

		session, err := webCon.App.GetSessionWitchContext(webCon.Ctx, webCon.GetSessionToken())
		if err == nil && session.CountLicenses() == 0 {
			err = model.SocketPermissionError
		}
		if err != nil {
			webCon.Log().Error("invalid session", wlog.Err(err))
			webCon.SetSessionToken("")
			webCon.SetSession(nil)
			webCon.SetSessionExpiresAt(0)
			webCon.SendError(err)
			return false
		}

		webCon.SetSession(session)
		webCon.SetSessionExpiresAt(session.Expire)
	}

	return true
}

func (webCon *WebConn) SetListenEvent(name string, value *model.BindQueueEvent) {
	webCon.mx.Lock()
	defer webCon.mx.Unlock()

	webCon.listenEvents[name] = value
}

func (webCon *WebConn) GetListenEvent(name string) (*model.BindQueueEvent, bool) {
	webCon.mx.RLock()
	v, ok := webCon.listenEvents[name]
	webCon.mx.RUnlock()

	return v, ok
}

func (webCon *WebConn) ShouldSendEvent(msg *model.WebSocketEvent) bool {
	if !webCon.IsAuthenticated() {
		return false
	}

	if _, ok := webCon.GetListenEvent(msg.EventType()); !ok {
		return true
	}

	switch msg.EventType() {
	//case model.WEBSOCKET_EVENT_CALL:
	//
	//	return false
	}

	return true
}

func (webCon *WebConn) GetAllBindings() []*model.BindQueueEvent {
	webCon.mx.RLock()
	defer webCon.mx.RUnlock()

	arr := make([]*model.BindQueueEvent, 0, len(webCon.listenEvents))
	for _, v := range webCon.listenEvents {
		arr = append(arr, v)
	}
	return arr
}
