package auth_manager

import (
	"context"
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
	ProductLimit(ctx context.Context, token string, productName string) (int, error)
}

type authManager struct {
	session          utils.ObjectCache
	serviceDiscovery discovery.ServiceDiscovery
	poolConnections  discovery.Pool

	watcher   *discovery.Watcher
	startOnce sync.Once
	stop      chan struct{}
	stopped   chan struct{}

	log *wlog.Logger
}

func NewAuthManager(cacheSize int, cacheTime int64, serviceDiscovery discovery.ServiceDiscovery, log *wlog.Logger) AuthManager {
	if cacheTime < 1 {
		// 0 disabled cache
		cacheTime = 1
	}
	return &authManager{
		stop:             make(chan struct{}),
		stopped:          make(chan struct{}),
		poolConnections:  discovery.NewPoolConnections(),
		session:          utils.NewLruWithParams(cacheSize, "auth manager", cacheTime, ""), //TODO session from config ?
		serviceDiscovery: serviceDiscovery,
		log:              log.With(wlog.Namespace("context")).With(wlog.String("scope", "auth_manager")),
	}
}

func (am *authManager) Start() error {
	am.log.Debug("starting")

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
				am.log.Debug("stopper")
				close(am.stopped)
			}()

			for {
				select {
				case <-am.stop:
					am.log.Debug("received stop signal")
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
		am.log.With(wlog.String("service_id", v.Id), wlog.String("ip_address", addr)).
			Err(err)
		return
	}
	am.poolConnections.Append(client)
	am.log.With(wlog.String("service_id", v.Id), wlog.String("ip_address", addr)).
		Debug("register")
}

func (am *authManager) wakeUp() {
	list, err := am.serviceDiscovery.GetByName(AUTH_SERVICE_NAME)
	if err != nil {
		am.log.Error(err.Error())
		return
	}

	for _, v := range list {
		if _, err := am.poolConnections.GetById(v.Id); err == discovery.ErrNotFoundConnection {
			am.registerConnection(v)
		}
	}
	am.poolConnections.RecheckConnections(list.Ids())
}
