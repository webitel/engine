package call_manager

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/webitel/engine/model"
	"github.com/webitel/engine/pkg/discovery"
	"github.com/webitel/wlog"
)

const (
	CLUSTER_CALL_SERVICE_NAME = "freeswitch"
	WATCHER_INTERVAL          = 1000 * 5
)

type CallManager interface {
	Start() error
	Stop()
	MakeOutboundCall(req *model.CallRequest) (string, model.AppError)
	Bridge(legA, legANode, legB, legBNode string) model.AppError
	CallClient() (CallClient, model.AppError)
	CallClientById(id string) (CallClient, model.AppError)

	SipWsAddress() string
	SipRouteUri() string
}

type CallClient interface {
	Name() string
	Ready() bool

	Host() string

	Execute(app string, args string) model.AppError

	GetServerVersion() (string, model.AppError)
	SetConnectionSps(sps int) (int, model.AppError)
	GetRemoteSps() (int, model.AppError)

	MakeOutboundCall(req *model.CallRequest) (string, model.AppError)

	NewCall(settings *model.CallRequest) (string, string, model.AppError)
	NewCallContext(ctx context.Context, settings *model.CallRequest) (string, string, model.AppError)

	HangupCall(id, cause string) model.AppError
	Hold(id string) model.AppError
	UnHold(id string) model.AppError
	SetCallVariables(id string, variables map[string]string) model.AppError
	BridgeCall(legAId, legBId string, vars map[string]string) (string, model.AppError)
	DTMF(id string, ch rune) model.AppError
	SetEavesdropState(id string, state string) model.AppError
	Mute(id string, val bool) model.AppError
	BlindTransfer(id, destination string) model.AppError
	BlindTransferExt(id, destination string, vars map[string]string) model.AppError
	ConfirmPushCall(id string) model.AppError

	Close() error
}

type callManager struct {
	sipServerAddr    string
	sipProxy         string
	serviceDiscovery discovery.ServiceDiscovery
	poolConnections  discovery.Pool

	watcher   *discovery.Watcher
	startOnce sync.Once
	stop      chan struct{}
	stopped   chan struct{}
}

func NewCallManager(addr, proxy string, serviceDiscovery discovery.ServiceDiscovery) CallManager {
	return &callManager{
		sipServerAddr:    getWsAddress(addr),
		sipProxy:         proxy,
		serviceDiscovery: serviceDiscovery,
		poolConnections:  discovery.NewPoolConnections(),

		stop:    make(chan struct{}),
		stopped: make(chan struct{}),
	}
}

func (cm *callManager) SipRouteUri() string {
	return "sip:" + cm.sipProxy
}

// todo bug if ""
func getWsAddress(addr string) string {
	if addr == "" {
		return "/sip"
	} else if strings.HasPrefix(addr, "wss://") || strings.HasPrefix(addr, "ws://") {
		return addr
	} else {
		return "wss://" + addr + "/sip"
	}
}

func (cm *callManager) SipWsAddress() string {
	return cm.sipServerAddr
}

func (c *callManager) CallClient() (CallClient, model.AppError) {
	cli, err := c.poolConnections.Get(discovery.StrategyRoundRobin)
	if err != nil {
		return nil, model.NewNotFoundError("call.get_client.not_found", err.Error())
	}
	return cli.(CallClient), nil
}

func (c *callManager) CallClientById(id string) (CallClient, model.AppError) {
	cli, err := c.poolConnections.GetById(id)
	if err != nil {
		return nil, model.NewNotFoundError("call.get_client_by_id.not_found", err.Error())
	}
	return cli.(CallClient), nil
}

func (c *callManager) Start() error {
	wlog.Debug(fmt.Sprintf("starting call manager [ws: %s, proxy: %s]", c.SipWsAddress(), c.SipRouteUri()))

	if services, err := c.serviceDiscovery.GetByName(CLUSTER_CALL_SERVICE_NAME); err != nil {
		return model.NewInternalError("", err.Error()) //
	} else {
		for _, v := range services {
			c.registerConnection(v)
		}
	}

	c.startOnce.Do(func() {
		c.watcher = discovery.MakeWatcher("call manager", WATCHER_INTERVAL, c.wakeUp)
		go c.watcher.Start()
		go func() {
			defer func() {
				wlog.Debug("stopper call manager")
				close(c.stopped)
			}()

			for {
				select {
				case <-c.stop:
					wlog.Debug("call manager received stop signal")
					return
				}
			}
		}()
	})
	return nil
}

func (c *callManager) Stop() {
	wlog.Debug("callManager Stopping")

	if c.watcher != nil {
		c.watcher.Stop()
	}

	if c.poolConnections != nil {
		c.poolConnections.CloseAllConnections()
	}

	close(c.stop)
	<-c.stopped
}

func (c *callManager) wakeUp() {
	list, err := c.serviceDiscovery.GetByName(CLUSTER_CALL_SERVICE_NAME)
	if err != nil {
		wlog.Error(err.Error())
		return
	}

	for _, v := range list {
		if _, err := c.poolConnections.GetById(v.Id); err == discovery.ErrNotFoundConnection {
			c.registerConnection(v)
		}
	}
	c.poolConnections.RecheckConnections(list.Ids())
}

func (c *callManager) registerConnection(v *discovery.ServiceConnection) {
	var version string
	var sps int

	client, err := NewCallConnection(v.Id, v.Host, c.SipRouteUri(), v.Port)
	if err != nil {
		wlog.Error(fmt.Sprintf("connection %s error: %s", v.Id, err.Error()))
		return
	}

	if version, err = client.GetServerVersion(); err != nil {
		wlog.Error(fmt.Sprintf("connection %s get version error: %s", v.Id, err.Error()))
		return
	}

	if sps, err = client.GetRemoteSps(); err != nil {
		wlog.Error(fmt.Sprintf("connection %s get SPS error: %s", v.Id, err.Error()))
		return
	}
	client.SetConnectionSps(sps)

	c.poolConnections.Append(client)
	wlog.Debug(fmt.Sprintf("register connection %s [%s] [sps=%d]", client.Name(), version, sps))
}
