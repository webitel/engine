package app

import (
	"strconv"
	"sync/atomic"
	"time"

	"github.com/webitel/engine/model"
	"github.com/webitel/engine/mq"
	"github.com/webitel/wlog"
)

const (
	SHUTDOWN_TICKER      = 5 * time.Minute
	BROADCAST_QUEUE_SIZE = 4096
)

type Hub struct {
	id               int64
	name             string
	connectionCount  int64
	app              *App
	register         chan *WebConn
	unregister       chan *WebConn
	broadcast        chan *model.WebSocketEvent
	stop             chan struct{}
	didStop          chan struct{}
	invalidateUser   chan string
	ExplicitStop     bool
	domainQueue      mq.DomainQueue
	lastUnregisterAt int64

	log *wlog.Logger
}

func (a *App) NewWebHub(name string, id int64) *Hub {
	hub := &Hub{
		id:               id,
		app:              a,
		name:             name,
		register:         make(chan *WebConn, 10),
		unregister:       make(chan *WebConn, 10),
		broadcast:        make(chan *model.WebSocketEvent, BROADCAST_QUEUE_SIZE),
		stop:             make(chan struct{}),
		didStop:          make(chan struct{}),
		lastUnregisterAt: model.GetMillis(),
		invalidateUser:   make(chan string),
		ExplicitStop:     false,
		log: a.Log.With(wlog.Namespace("context"),
			wlog.String("scope", "wsHub"),
			wlog.Int64("domain_id", id),
		),
	}
	dq, _ := a.MessageQueue.NewDomainQueue(id, hub.GetAllBindings)
	hub.domainQueue = dq

	hub.domainQueue.Start()
	go hub.start()

	return hub
}

func (wh *Hub) GetAllBindings() []*model.BindQueueEvent {
	a := make([]*model.BindQueueEvent, 0, 0)
	return a
}

func (wh *Hub) start() {
	wh.log.Debug("started")

	ticker := time.NewTicker(SHUTDOWN_TICKER)
	defer func() {
		ticker.Stop()
		wh.log.Debug("stopped")
	}()

	connections := newHubConnectionIndex()

	for {
		select {
		case <-ticker.C:

			if wh.connectionCount == 0 && (wh.lastUnregisterAt+(24*60*60*1000)) < model.GetMillis() {
				wh.domainQueue.Stop()
				wh.app.DeleteHub(wh.id)
				wh.log.Debug("shutdown")
				return
			}
		case webCon := <-wh.register:
			connections.Add(webCon)
			atomic.StoreInt64(&wh.connectionCount, int64(len(connections.All())))

			webCon.SetLog(webCon.Log().With(wlog.Int64("user_id", webCon.UserId),
				wlog.Int64("domain_id", webCon.DomainId),
			))

			err := wh.app.MessageQueue.RegisterWebsocket(webCon.DomainId, &model.RegisterToWebsocketEvent{
				AppId:     wh.app.nodeId,
				Timestamp: model.GetMillis(),
				UserId:    webCon.UserId,
				Addr:      webCon.WebSocket.RemoteAddr().String(),
				SocketId:  webCon.id,
			})
			if err != nil {
				webCon.Log().Err(err)
			}

			wlog.Debug("register", wlog.Int("count", len(connections.ForUser(webCon.UserId))))

		case webCon := <-wh.unregister:
			connections.Remove(webCon)

			atomic.StoreInt64(&wh.connectionCount, int64(len(connections.All())))
			wh.lastUnregisterAt = model.GetMillis()

			wh.domainQueue.BulkUnbind(webCon.GetAllBindings())

			err := wh.app.MessageQueue.UnRegisterWebsocket(webCon.DomainId, &model.RegisterToWebsocketEvent{
				AppId:     wh.app.nodeId,
				Timestamp: model.GetMillis(),
				UserId:    webCon.UserId,
				Addr:      webCon.WebSocket.RemoteAddr().String(),
				SocketId:  webCon.id,
			})
			if err != nil {
				webCon.Log().Err(err)
			}

			wlog.Debug("deregister", wlog.Int("count", len(connections.ForUser(webCon.UserId))))

			if wh.app.b2b != nil && !connections.HasUser(webCon.UserId) {
				wh.app.b2b.Unregister(webCon.UserId, 5)
			}

		case msg := <-wh.domainQueue.Events():
			candidates := connections.ForUser(msg.UserId)
			for _, webCon := range candidates {
				if webCon.ShouldSendEvent(msg) {
					select {
					case webCon.Send <- msg:
					default:
						webCon.Log().Error("webcon.broadcast.dq: cannot send, closing websocket", wlog.String("event_type", "dq"))
						close(webCon.Send)
						connections.Remove(webCon)
					}
				}
			}

		case ev := <-wh.domainQueue.CallEvents():

			msg := model.NewWebSocketCallEvent(ev)

			usr, _ := strconv.Atoi(ev.UserId)

			msg.PrecomputeJSON()
			candidates := connections.ForUser(int64(usr))
			for _, webCon := range candidates {
				//FIXME permission call events
				if webCon.ShouldSendEvent(msg) {
					select {
					case webCon.Send <- msg:
					default:
						webCon.Log().Error("webcon.broadcast.call: cannot send, closing websocket", wlog.String("event_type", "call"))
						close(webCon.Send)
						connections.Remove(webCon)
					}
					// todo delete me
					webCon.Log().Debug("send event", wlog.String("event_type", "call"), wlog.String("event_name", ev.Event), wlog.String("event_id", ev.Id))
				}
			}

		case ev := <-wh.domainQueue.ChatEvents():
			candidates := connections.ForUser(ev.UserId)
			msg := model.NewWebSocketChatEvent(ev)
			for _, webCon := range candidates {
				select {
				case webCon.Send <- msg:
				default:
					webCon.Log().Error("webcon.broadcast.chat: cannot send, closing websocket", wlog.String("event_type", "chat"))
					close(webCon.Send)
					connections.Remove(webCon)
				}
			}

		case ev := <-wh.broadcast:
			candidates := connections.ForUser(ev.UserId)
			for _, webCon := range candidates {
				if ev.SockId != "" && ev.SockId != webCon.id {
					//continue
				}
				select {
				case webCon.Send <- ev:
				default:
					webCon.Log().Error("webhub.broadcast: cannot send, closing websocket", wlog.String("event_type", "webhub"))
					close(webCon.Send)
					connections.Remove(webCon)
				}
			}

		case ev := <-wh.domainQueue.UserStateEvents():

			msg := model.NewWebSocketUserStateEvent(ev)

			msg.PrecomputeJSON()
			candidates := connections.ForUser(msg.UserId)

			for _, webCon := range candidates {
				select {
				case webCon.Send <- msg:
				default:
					webCon.Log().Error("webhub.broadcast.user_status: cannot send, closing websocket", wlog.String("event_type", "user_status"))
					close(webCon.Send)
					connections.Remove(webCon)
				}
			}

		case ev := <-wh.domainQueue.NotificationEvents():
			msg := model.NewWebSocketNotificationEvent(ev)
			msg.PrecomputeJSON()

			if len(ev.ForUsers) != 0 {
				for _, u := range ev.ForUsers {
					candidates := connections.ForUser(u)
					for _, webCon := range candidates {
						if webCon.ShouldSendEvent(msg) {
							select {
							case webCon.Send <- msg:
							default:
								webCon.Log().Error("webhub.broadcast.notification: cannot send, closing websocket", wlog.String("event_type", "notification"))
								close(webCon.Send)
								connections.Remove(webCon)
							}
						}
					}
				}
			}
		}
	}
}

func (wh *Hub) UnSubscribeCalls(conn *WebConn) model.AppError {
	if b, ok := conn.GetListenEvent("call"); ok {
		wh.domainQueue.Unbind(b)
	} else {
		//NOTFOUND
	}

	return nil
}

func (wh *Hub) SubscribeSessionCalls(conn *WebConn) model.AppError {

	b := wh.domainQueue.BindUserCall(conn.Id(), conn.GetSession().UserId)
	//TODO
	conn.SetListenEvent("call", b)

	return nil
}

func (wh *Hub) SubscribeSessionChat(conn *WebConn) model.AppError {

	b := wh.domainQueue.BindUserChat(conn.Id(), conn.GetSession().UserId)
	//TODO
	conn.SetListenEvent("chat", b)

	return nil
}

func (wh *Hub) SubscribeSessionUsersStatus(conn *WebConn) model.AppError {

	b := wh.domainQueue.BindUsersStatus(conn.Id(), conn.GetSession().UserId)
	//TODO
	conn.SetListenEvent("status", b)

	return nil
}

func (wh *Hub) SubscribeSessionAgentStatus(conn *WebConn, agentId int) model.AppError {

	b := wh.domainQueue.BindAgentStatusEvents(conn.Id(), conn.GetSession().UserId, agentId)
	//TODO
	conn.SetListenEvent("agent_status", b)

	b2 := wh.domainQueue.BindAgentChannelEvents(conn.Id(), conn.GetSession().UserId, agentId)
	//TODO
	conn.SetListenEvent("agent_channel", b2)

	return nil
}

func (a *App) GetHubById(id int64) (*Hub, model.AppError) {
	if h, ok := a.Hubs.Get(id); ok {
		return h, nil
	} else {
		h = a.Hubs.Register(id, "")
		return h, nil
	}
}

func (a *App) DeleteHub(id int64) {
	a.Hubs.Remove(id)
}

func (a *App) HubRegister(webCon *WebConn) {
	hub, _ := a.GetHubById(webCon.DomainId)
	if hub != nil {
		hub.Register(webCon)
	}
}

func (a *App) HubUnregister(webConn *WebConn) {
	if webConn.UserId == 0 {
		return //TODO user not register
	}
	hub, _ := a.GetHubById(webConn.DomainId)
	if hub != nil {
		hub.Unregister(webConn)
	}
}

func (h *Hub) Register(webConn *WebConn) {
	select {
	case h.register <- webConn:
	case <-h.didStop:
	}

	if webConn.IsAuthenticated() {
		webConn.SendHello()
	}
}

func (h *Hub) Unregister(webConn *WebConn) {
	select {
	case h.unregister <- webConn:
	case <-h.stop:
	}
}

// hubConnectionIndex provides fast addition, removal, and iteration of web connections.
// It requires 3 functionalities which need to be very fast:
// - check if a connection exists or not.
// - get all connections for a given userID.
// - get all connections.
type hubConnectionIndex struct {
	// byUserId stores the list of connections for a given userID
	byUserId map[int64][]*WebConn
	// byConnection serves the dual purpose of storing the index of the webconn
	// in the value of byUserId map, and also to get all connections.
	byConnection map[*WebConn]int
}

func newHubConnectionIndex() *hubConnectionIndex {
	return &hubConnectionIndex{
		byUserId:     make(map[int64][]*WebConn),
		byConnection: make(map[*WebConn]int),
	}
}

func (i *hubConnectionIndex) Add(wc *WebConn) {
	i.byUserId[wc.UserId] = append(i.byUserId[wc.UserId], wc)
	i.byConnection[wc] = len(i.byUserId[wc.UserId]) - 1
}

func (i *hubConnectionIndex) Remove(wc *WebConn) {
	userConnIndex, ok := i.byConnection[wc]
	if !ok {
		return
	}

	// get the conn slice.
	userConnections := i.byUserId[wc.UserId]
	// get the last connection.
	last := userConnections[len(userConnections)-1]
	// set the slot that we are trying to remove to be the last connection.
	userConnections[userConnIndex] = last
	// remove the last connection from the slice.
	i.byUserId[wc.UserId] = userConnections[:len(userConnections)-1]
	// set the index of the connection that was moved to the new index.
	i.byConnection[last] = userConnIndex

	delete(i.byConnection, wc)
}

func (i *hubConnectionIndex) Has(wc *WebConn) bool {
	_, ok := i.byConnection[wc]
	return ok
}

func (i *hubConnectionIndex) ForUser(id int64) []*WebConn {
	return i.byUserId[id]
}

func (i *hubConnectionIndex) HasUser(id int64) bool {
	c, ok := i.byUserId[id]
	return ok && len(c) > 0
}

func (i *hubConnectionIndex) All() map[*WebConn]int {
	return i.byConnection
}
