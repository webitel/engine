package app

import (
	"github.com/webitel/call_center/discovery"
	"github.com/webitel/engine/model"
)

type cluster struct {
	app       *App
	discovery discovery.ServiceDiscovery
}

func NewCluster(app *App) *cluster {
	return &cluster{
		app: app,
	}
}

func (c *cluster) Start() error {
	sd, err := discovery.NewServiceDiscovery(c.app.nodeId, "192.168.177.199:8500", func() (b bool, appError error) {
		return true, nil
	})
	if err != nil {
		return err
	}
	c.discovery = sd

	err = sd.RegisterService(model.APP_SERVICE_NAME, "10.10.10.25", 8081, model.APP_SERVICE_TTL, model.APP_DEREGESTER_CRITICAL_TTL)
	if err != nil {
		return err
	}

	return nil
}

func (c *cluster) Stop() {
	c.discovery.Shutdown()
}
