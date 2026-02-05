package auth_manager

import (
	"context"
	"github.com/webitel/wlog"
	"testing"
	"time"
)

const (
	TOKEN = "IHOR"
)

var TEST_CONSUL = "172.22.22.22:8500"

func TestAuthManager(t *testing.T) {
	t.Log("AuthManager")

	am := NewAuthManager(1, 10000, TEST_CONSUL, wlog.NewLogger(&wlog.LoggerConfiguration{
		EnableExport:  false,
		EnableConsole: true,
		ConsoleJson:   false,
		ConsoleLevel:  "",
		EnableFile:    false,
		FileJson:      false,
		FileLevel:     "",
		FileLocation:  "",
	}))
	am.Start()
	defer am.Stop()

	for i := 0; i < 1000; i++ {
		go testGetSession(t, TOKEN, am.(*authManager))
		println(i)
	}

	time.Sleep(120 * time.Second)
}

func testGetSession(t *testing.T, token string, am *authManager) {

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
