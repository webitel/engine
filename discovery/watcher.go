package discovery

import (
	"github.com/webitel/wlog"
	"time"
)

type WatcherNotify func()

type Watcher struct {
	name            string
	stop            chan struct{}
	stopped         chan struct{}
	pollingInterval int
	PollAndNotify   WatcherNotify
	log             *wlog.Logger
}

func MakeWatcher(name string, pollingInterval int, pollAndNotify WatcherNotify) *Watcher {
	return &Watcher{
		name:            name,
		stop:            make(chan struct{}),
		stopped:         make(chan struct{}),
		pollingInterval: pollingInterval,
		PollAndNotify:   pollAndNotify,
		log: wlog.GlobalLogger().
			With(wlog.Namespace("context")).
			With(wlog.String("scope", "watcher"), wlog.String("name", name)),
	}
}

func (watcher *Watcher) Start() {
	watcher.log.Debug("started")
	//<-time.After(time.Duration(rand.Intn(watcher.pollingInterval)) * time.Millisecond)

	defer func() {
		watcher.log.Debug("finished")
		close(watcher.stopped)
	}()

	for {
		select {
		case <-watcher.stop:
			watcher.log.Debug("received stop signal")
			return
		case <-time.After(time.Duration(watcher.pollingInterval) * time.Millisecond):
			watcher.PollAndNotify()
		}
	}
}

func (watcher *Watcher) Stop() {
	watcher.log.Debug("stopping")
	close(watcher.stop)
	<-watcher.stopped
}
