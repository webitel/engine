package app

import (
	"fmt"
	"github.com/webitel/engine/model"
	"github.com/webitel/engine/mq"
	"github.com/webitel/wlog"
	"sync/atomic"
	"time"
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
}

func (a *App) NewWebHub(name string, id int64) *Hub {
	hub := &Hub{
		id:               id,
		app:              a,
		name:             name,
		register:         make(chan *WebConn, 1),
		unregister:       make(chan *WebConn, 1),
		broadcast:        make(chan *model.WebSocketEvent, BROADCAST_QUEUE_SIZE),
		stop:             make(chan struct{}),
		didStop:          make(chan struct{}),
		lastUnregisterAt: model.GetMillis(),
		invalidateUser:   make(chan string),
		ExplicitStop:     false,
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
	wlog.Debug(fmt.Sprintf("hub %s started", wh.name))

	ticker := time.NewTicker(SHUTDOWN_TICKER)
	defer func() {
		ticker.Stop()
		wlog.Debug(fmt.Sprintf("hub %s stopped", wh.name))
	}()

	connections := newHubConnectionIndex()

	for {
		select {
		case <-ticker.C:

			if wh.connectionCount == 0 && (wh.lastUnregisterAt+(5*60*1000)) < model.GetMillis() {
				wh.domainQueue.Stop()
				wh.app.DeleteHub(wh.id)
				wlog.Debug(fmt.Sprintf("shutdown domain=%s hub", wh.name))
				return
			}
		case webCon := <-wh.register:
			connections.Add(webCon)
			atomic.StoreInt64(&wh.connectionCount, int64(len(connections.All())))

			wlog.Debug(fmt.Sprintf("register user %d opened socket %d", webCon.UserId, len(connections.ForUser(webCon.UserId))))

		case webCon := <-wh.unregister:
			connections.Remove(webCon)

			atomic.StoreInt64(&wh.connectionCount, int64(len(connections.All())))
			wh.lastUnregisterAt = model.GetMillis()

			wh.domainQueue.BulkUnbind(webCon.GetAllBindings())

			wlog.Debug(fmt.Sprintf("un-register user %d opened socket %d", webCon.UserId, len(connections.ForUser(webCon.UserId))))

		case ev := <-wh.domainQueue.CallEvents():

			msg := model.NewWebSocketCallEvent(ev)

			msg.PrecomputeJSON()
			candidates := connections.All()
			for _, webCon := range candidates {
				if webCon.ShouldSendEvent(msg) {
					select {
					case webCon.Send <- msg:
					default:
						wlog.Error(fmt.Sprintf("webhub.broadcast: cannot send, closing websocket for userId=%v", webCon.UserId))
						close(webCon.Send)
						connections.Remove(webCon)
					}
				}
			}
		}
	}
}

func (wh *Hub) UnSubscribeCalls(conn *WebConn) *model.AppError {
	if b, ok := conn.GetListenEvent("call"); ok {
		wh.domainQueue.Unbind(b)
	} else {
		//NOTFOUND
	}

	return nil
}

func (wh *Hub) SubscribeSessionCalls(conn *WebConn) *model.AppError {

	b := wh.domainQueue.BindUserCall(conn.Id(), conn.GetSession().UserId)
	//TODO
	conn.SetListenEvent("call", b)

	return nil
}

func (a *App) GetHubById(id int64) (*Hub, *model.AppError) {
	if h, ok := a.Hubs.Get(id); ok {
		return h, nil
	} else {
		h = a.Hubs.Register(id, "TODO")
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

type hubConnectionIndexIndexes struct {
	connections         int
	connectionsByUserId int
}

// hubConnectionIndex provides fast addition, removal, and iteration of web connections.
type hubConnectionIndex struct {
	connections         []*WebConn
	connectionsByUserId map[int64][]*WebConn
	connectionIndexes   map[*WebConn]*hubConnectionIndexIndexes
}

func newHubConnectionIndex() *hubConnectionIndex {
	return &hubConnectionIndex{
		connections:         make([]*WebConn, 0, model.SESSION_CACHE_SIZE),
		connectionsByUserId: make(map[int64][]*WebConn),
		connectionIndexes:   make(map[*WebConn]*hubConnectionIndexIndexes),
	}
}

func (i *hubConnectionIndex) Add(wc *WebConn) {
	i.connections = append(i.connections, wc)
	i.connectionsByUserId[wc.UserId] = append(i.connectionsByUserId[wc.UserId], wc)
	i.connectionIndexes[wc] = &hubConnectionIndexIndexes{
		connections:         len(i.connections) - 1,
		connectionsByUserId: len(i.connectionsByUserId[wc.UserId]) - 1,
	}
}

func (i *hubConnectionIndex) Remove(wc *WebConn) {
	indexes, ok := i.connectionIndexes[wc]
	if !ok {
		return
	}

	last := i.connections[len(i.connections)-1]
	i.connections[indexes.connections] = last
	i.connections = i.connections[:len(i.connections)-1]
	i.connectionIndexes[last].connections = indexes.connections

	userConnections := i.connectionsByUserId[wc.UserId]
	last = userConnections[len(userConnections)-1]
	userConnections[indexes.connectionsByUserId] = last
	i.connectionsByUserId[wc.UserId] = userConnections[:len(userConnections)-1]
	i.connectionIndexes[last].connectionsByUserId = indexes.connectionsByUserId

	delete(i.connectionIndexes, wc)
}

func (i *hubConnectionIndex) ForUser(id int64) []*WebConn {
	return i.connectionsByUserId[id]
}

func (i *hubConnectionIndex) All() []*WebConn {
	return i.connections
}
