package discovery

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/hashicorp/consul/api"

	"github.com/webitel/wlog"
)

type consul struct {
	id              string
	cli             *api.Client
	kv              *api.KV
	agent           *api.Agent
	stop            chan struct{}
	check           CheckFunction
	checkId         string
	registerService bool
	ttl             time.Duration
	as              *api.AgentServiceRegistration
}

type CheckFunction func() (bool, error)

func NewConsul(id, addr string, check CheckFunction) (*consul, error) {
	conf := api.DefaultConfig()
	conf.Address = addr

	cli, err := api.NewClient(conf)
	if err != nil {
		return nil, err
	}

	c := &consul{
		id:              id,
		registerService: true,
		cli:             cli,
		agent:           cli.Agent(),
		stop:            make(chan struct{}),
		check:           check,
		kv:              cli.KV(),
	}

	return c, nil
}

func (c *consul) GetByName(serviceName string) (ListConnections, error) {
	entries, _, err := c.cli.Health().Service(serviceName, "", true, nil)
	if err != nil {
		return nil, err
	}

	result := make([]*ServiceConnection, 0, len(entries))
	for _, entry := range entries {
		svc := entry.Service
		result = append(result, &ServiceConnection{
			Id:      svc.ID,
			Service: svc.Service,
			Host:    svc.Address,
			Port:    svc.Port,
		})
	}

	return result, nil
}

// RegisterService TODO
func (c *consul) RegisterService(name, pubHost string, pubPort int, ttl, criticalTtl time.Duration) error {
	if !c.registerService {
		return nil
	}

	c.ttl = ttl
	c.id = fmt.Sprintf("%s-%s", name, c.id)
	c.as = &api.AgentServiceRegistration{
		Name:    name,
		ID:      c.id,
		Tags:    []string{c.id},
		Address: pubHost,
		Port:    pubPort,
		Check: &api.AgentServiceCheck{
			DeregisterCriticalServiceAfter: criticalTtl.String(),
			TTL:                            c.ttl.String(),
			CheckID:                        c.id,
		},
	}
	c.agent.ServiceRegister(c.as)
	return nil
	// return c.register(c.as)
}

func (c *consul) register(as *api.AgentServiceRegistration) error {
	var err error
	if err = c.agent.ServiceRegister(as); err != nil {
		return err
	}

	var checks map[string]*api.AgentCheck
	if checks, err = c.agent.Checks(); err != nil {
		return err
	}

	var serviceCheck *api.AgentCheck
	for _, check := range checks {
		if check.ServiceID == c.id {
			serviceCheck = check
		}
	}

	if serviceCheck == nil {
		return errors.New("serviceCheck is null")
	}
	c.checkId = serviceCheck.CheckID
	c.update(as)

	wlog.Info(fmt.Sprintf("started consul service id: %s", c.id))

	go c.updateTTL(c.ttl/2, as)

	return nil
}

func (c *consul) update(as *api.AgentServiceRegistration) {
	ok, err := c.check()
	if !ok {
		if agentErr := c.agent.FailTTL(c.checkId, err.Error()); agentErr != nil {
			c.handlePassTTLError(agentErr, as)
		}
	} else {
		if agentErr := c.agent.PassTTL(c.checkId, "ready..."); agentErr != nil {
			c.handlePassTTLError(agentErr, as)
		}
	}
}

func (c *consul) handlePassTTLError(err error, as *api.AgentServiceRegistration) {
	switch wrapErr := err.(type) {
	case api.StatusError:
		if wrapErr.Code == http.StatusInternalServerError {
			// reconnect ?
			wlog.Info(fmt.Sprintf("reconnect consul service id: %s", c.id))
			if e := c.register(as); e != nil {
				wlog.Error(fmt.Sprintf("reconnect consul service %s error: %s", c.id, e.Error()))
			}
		}
	default:
		wlog.Error(err.Error())
	}
}

func (c *consul) updateTTL(ttl time.Duration, as *api.AgentServiceRegistration) {
	defer wlog.Info("stopped consul checker")

	ticker := time.NewTicker(ttl / 2)
	for {
		select {
		case <-c.stop:
			return
		case <-ticker.C:
			c.update(as)
		}
	}
}

func (c *consul) Shutdown() {
	close(c.stop)
	c.agent.ServiceDeregister(c.id)
}

func (c *consul) GetValueByKey(key string) string {
	kvp, _, _ := c.kv.Get(key, nil)
	return string(kvp.Value)
}

func (c *consul) SaveValue(key, value string) {
	c.kv.Put(&api.KVPair{Key: key, Value: []byte(value)}, nil)
}
