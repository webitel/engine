package app

import (
	"github.com/webitel/engine/model"
	"github.com/webitel/engine/utils"
	"github.com/webitel/wlog"
	"sync"
	"time"
)

type Hubs struct {
	app       *App
	hubs      map[int64]*Hub
	storePool *utils.Pool
	sync.RWMutex
}

type taskHub struct {
	a   *App
	log *wlog.Logger
}

type taskHubCreate struct {
	taskHub
	session model.SocketSession
}

type taskHubDelete struct {
	taskHub
	id string
}

type taskHubPong struct {
	taskHub
	id string
	t  time.Time
}

func (ts *taskHubCreate) Execute() {
	err := ts.a.Store.SocketSession().Create(ts.a.ctx, ts.session)
	if err != nil {
		ts.log.Error(err.Error(), wlog.Err(err))
	}
}

func (ts *taskHubPong) Execute() {
	err := ts.a.Store.SocketSession().SetUpdatedAt(ts.a.ctx, ts.id, ts.t)
	if err != nil {
		ts.log.Error(err.Error(), wlog.Err(err))
	}
}

func (ts *taskHubDelete) Execute() {
	err := ts.a.Store.SocketSession().DeleteById(ts.a.ctx, ts.id)
	if err != nil {
		ts.log.Error(err.Error(), wlog.Err(err))
	}
}

func NewHubs(a *App) *Hubs {
	return &Hubs{
		app:       a,
		hubs:      make(map[int64]*Hub),
		storePool: utils.NewPool(4, 1000),
	}
}

func (hs *Hubs) Get(id int64) (*Hub, bool) {
	hs.RLock()
	defer hs.RUnlock()

	h, ok := hs.hubs[id]
	return h, ok
}

func (hs *Hubs) Remove(id int64) {
	hs.Lock()
	defer hs.Unlock()

	delete(hs.hubs, id)
}

func (hs *Hubs) Register(id int64, name string) *Hub {
	hs.Lock()
	defer hs.Unlock()

	if h, ok := hs.hubs[id]; ok {
		return h
	} else {
		h = hs.app.NewWebHub(name, id)
		hs.hubs[id] = h
		return h
	}
}

func (hs *Hubs) Clean() {
	err := hs.app.Store.SocketSession().DeleteByApp(hs.app.ctx, hs.app.nodeId)
	if err != nil {
		hs.app.Log.Error(err.Error(), wlog.Err(err))
	}
}
