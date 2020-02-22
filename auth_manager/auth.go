package auth_manager

import (
	"fmt"
	"github.com/webitel/engine/discovery"
	"github.com/webitel/engine/utils"
	"github.com/webitel/wlog"
	"sync"
)

var (
	AUTH_SERVICE_NAME = "go.webitel.app"
)

const (
	WATCHER_INTERVAL = 5 * 1000
)

type AuthManager interface {
	Start() error
	Stop()
	GetSession(token string) (*Session, error)
}

type authManager struct {
	session          utils.ObjectCache
	serviceDiscovery discovery.ServiceDiscovery
	poolConnections  discovery.Pool

	watcher   *discovery.Watcher
	startOnce sync.Once
	stop      chan struct{}
	stopped   chan struct{}
}

func NewAuthManager(cacheSize int, cacheTime int64, serviceDiscovery discovery.ServiceDiscovery) AuthManager {
	return &authManager{
		stop:             make(chan struct{}),
		stopped:          make(chan struct{}),
		poolConnections:  discovery.NewPoolConnections(),
		session:          utils.NewLruWithParams(cacheSize, "auth manager", cacheTime, ""), //TODO session from config ?
		serviceDiscovery: serviceDiscovery,
	}
}

func (am *authManager) Start() error {
	wlog.Debug("starting auth service")

	if services, err := am.serviceDiscovery.GetByName(AUTH_SERVICE_NAME); err != nil {
		return err
	} else {
		for _, v := range services {
			am.registerConnection(v)
		}
	}

	am.startOnce.Do(func() {
		am.watcher = discovery.MakeWatcher("auth manager", WATCHER_INTERVAL, am.wakeUp)
		go am.watcher.Start()
		go func() {
			defer func() {
				wlog.Debug("stopper auth manager")
				close(am.stopped)
			}()

			for {
				select {
				case <-am.stop:
					wlog.Debug("auth manager received stop signal")
					return
				}
			}
		}()
	})
	return nil
}

func (am *authManager) Stop() {
	if am.watcher != nil {
		am.watcher.Stop()
	}

	if am.poolConnections != nil {
		am.poolConnections.CloseAllConnections()
	}

	close(am.stop)
	<-am.stopped
}

func (am *authManager) getAuthClient() (AuthClient, error) {
	conn, err := am.poolConnections.Get(discovery.StrategyRoundRobin)
	if err != nil {
		return nil, err
	}
	return conn.(AuthClient), nil
}

func (am *authManager) registerConnection(v *discovery.ServiceConnection) {
	addr := fmt.Sprintf("%s:%d", v.Host, v.Port)
	client, err := NewAuthServiceConnection(v.Id, addr)
	if err != nil {
		wlog.Error(fmt.Sprintf("connection %s [%s] error: %s", v.Id, addr, err.Error()))
		return
	}
	am.poolConnections.Append(client)
	wlog.Debug(fmt.Sprintf("register connection %s [%s]", client.Name(), addr))
}

func (am *authManager) wakeUp() {
	list, err := am.serviceDiscovery.GetByName(AUTH_SERVICE_NAME)
	if err != nil {
		wlog.Error(err.Error())
		return
	}

	for _, v := range list {
		if _, err := am.poolConnections.GetById(v.Id); err == discovery.ErrNotFoundConnection {
			am.registerConnection(v)
		}
	}
	am.poolConnections.RecheckConnections()
}
