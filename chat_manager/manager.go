package chat_manager

import (
	"fmt"
	"github.com/webitel/engine/discovery"
	"github.com/webitel/wlog"
	"sync"
)

var (
	SERVICE_NAME = "webitel.chat.server"
)

const (
	WATCHER_INTERVAL = 5 * 1000
)

type ChatManager interface {
	Start() error
	Client() (Chat, error)
	Stop()
}

type chatManager struct {
	serviceDiscovery discovery.ServiceDiscovery
	poolConnections  discovery.Pool

	watcher   *discovery.Watcher
	startOnce sync.Once
	stop      chan struct{}
	stopped   chan struct{}
}

func NewChatManager(serviceDiscovery discovery.ServiceDiscovery) ChatManager {
	return &chatManager{
		stop:             make(chan struct{}),
		stopped:          make(chan struct{}),
		poolConnections:  discovery.NewPoolConnections(),
		serviceDiscovery: serviceDiscovery,
	}
}

func (cm *chatManager) Start() error {
	wlog.Debug("starting chat service client")

	if services, err := cm.serviceDiscovery.GetByName(SERVICE_NAME); err != nil {
		return err
	} else {
		for _, v := range services {
			cm.registerConnection(v)
		}
	}

	cm.startOnce.Do(func() {
		cm.watcher = discovery.MakeWatcher("chat manager", WATCHER_INTERVAL, cm.wakeUp)
		go cm.watcher.Start()
		go func() {
			defer func() {
				wlog.Debug("stopped chat manager")
				close(cm.stopped)
			}()

			for {
				select {
				case <-cm.stop:
					wlog.Debug("chat manager received stop signal")
					return
				}
			}
		}()
	})
	return nil
}

func (cm *chatManager) Stop() {
	if cm.watcher != nil {
		cm.watcher.Stop()
	}

	if cm.poolConnections != nil {
		cm.poolConnections.CloseAllConnections()
	}

	close(cm.stop)
	<-cm.stopped
}

func (cm *chatManager) registerConnection(v *discovery.ServiceConnection) {
	addr := fmt.Sprintf("%s:%d", v.Host, v.Port)
	client, err := NewChatServiceConnection(v.Id, addr)
	if err != nil {
		wlog.Error(fmt.Sprintf("connection %s [%s] error: %s", v.Id, addr, err.Error()))
		return
	}
	cm.poolConnections.Append(client)
	wlog.Debug(fmt.Sprintf("register connection %s [%s]", client.Name(), addr))
}

func (cm *chatManager) wakeUp() {
	list, err := cm.serviceDiscovery.GetByName(SERVICE_NAME)
	if err != nil {
		wlog.Error(err.Error())
		return
	}

	for _, v := range list {
		if _, err := cm.poolConnections.GetById(v.Id); err == discovery.ErrNotFoundConnection {
			cm.registerConnection(v)
		}
	}
	cm.poolConnections.RecheckConnections(list.Ids())
}

func (cm *chatManager) Client() (Chat, error) {
	conn, err := cm.poolConnections.Get(discovery.StrategyRoundRobin)
	if err != nil {
		return nil, err
	}
	return conn.(Chat), nil
}
