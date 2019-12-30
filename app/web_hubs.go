package app

import "sync"

type Hubs struct {
	app  *App
	hubs map[int64]*Hub
	sync.RWMutex
}

func NewHubs(a *App) *Hubs {
	return &Hubs{
		app:  a,
		hubs: make(map[int64]*Hub),
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
