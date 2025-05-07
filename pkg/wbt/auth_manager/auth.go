package auth_manager

import (
	"context"
	"github.com/hashicorp/golang-lru/v2/expirable"
	"github.com/webitel/engine/pkg/wbt"
	api "github.com/webitel/engine/pkg/wbt/gen"
	"github.com/webitel/wlog"
	"sync"
	"time"
)

var (
	authServiceName = "go.webitel.app"
)

type AuthManager interface {
	Start() error
	Stop()
	GetSession(ctx context.Context, token string) (*Session, error)
	ProductLimit(ctx context.Context, token string, productName string) (int, error)
}

type authManager struct {
	session    *expirable.LRU[string, *Session]
	startOnce  sync.Once
	consulAddr string
	auth       *wbt.Client[api.AuthClient]
	customer   *wbt.Client[api.CustomersClient]

	log *wlog.Logger
}

func NewAuthManager(cacheSize int, cacheTime int64, consulAddr string, log *wlog.Logger) AuthManager {
	if cacheTime < 1 {
		// 0 disabled cache
		cacheTime = 1
	}
	return &authManager{
		consulAddr: consulAddr,
		session:    expirable.NewLRU[string, *Session](cacheSize, nil, time.Second*time.Duration(cacheTime)),
		log:        log.With(wlog.Namespace("context")).With(wlog.String("scope", "auth_manager")),
	}
}

func (am *authManager) Start() error {
	am.log.Debug("starting")
	var err error

	am.startOnce.Do(func() {
		am.auth, err = wbt.NewClient(am.consulAddr, authServiceName, api.NewAuthClient)
		if err != nil {
			return
		}

		am.customer, err = wbt.NewClient(am.consulAddr, authServiceName, api.NewCustomersClient)
		if err != nil {
			return
		}
	})
	return err
}

func (am *authManager) Stop() {
	am.log.Debug("stopping")
	_ = am.auth.Close()
	_ = am.customer.Close()
}
