package auth_manager

import (
	"context"
	"github.com/webitel/engine/discovery"
	"testing"
)

const (
	TOKEN = "USER_TOKEN"
)

var TEST_CONSUL = "localhost:8500"

func TestAuthManager(t *testing.T) {
	t.Log("AuthManager")

	sd, err := discovery.NewServiceDiscovery(AUTH_SERVICE_NAME, TEST_CONSUL, func() (b bool, appError error) {
		return true, nil
	})
	if err != nil {
		panic(err.Error())
	}

	am := NewAuthManager(1, 10000, sd)
	am.Start()
	defer am.Stop()

	for i := 0; i < 1; i++ {
		testGetSession(t, TOKEN, am)
	}
}

func testGetSession(t *testing.T, token string, am AuthManager) {

	session, err := am.GetSession(context.Background(), token)
	if err != nil {
		t.Errorf("get session \"%s\" error: %s", token, err.Error())
		return
	}

	if session == nil {
		t.Errorf("get session \"%s\" is nil", token)
		return
	}

	if err = session.IsValid(); err != nil {
		t.Errorf("bad session \"%s\": %v", token, err.Error())
		return
	}

	if session.Token != TOKEN {
		t.Errorf("bad session \"%s\": %v", token, session.Token)
		return
	}
}
