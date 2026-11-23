package controller

import (
	"testing"

	"github.com/webitel/engine/model"
	"github.com/webitel/engine/pkg/wbt/auth_manager"
)

func TestEmailCommandAccess(t *testing.T) {
	scope := func(access uint32) *auth_manager.Session {
		return &auth_manager.Session{Scopes: []auth_manager.SessionPermission{
			{Name: model.PERMISSION_SCOPE_EMAIL_THREAD, Obac: true, Access: access},
		}}
	}
	if !emailCommandAllowed(scope(auth_manager.PERMISSION_ACCESS_UPDATE.Value() | auth_manager.PERMISSION_ACCESS_READ.Value())) {
		t.Error("update was refused")
	}
	if emailCommandAllowed(scope(auth_manager.PERMISSION_ACCESS_READ.Value())) {
		t.Error("read only was allowed")
	}
	if emailCommandAllowed(&auth_manager.Session{}) {
		t.Error("a session without the scope was allowed")
	}

	if emailTargetQueue(&model.EmailAttemptCommand{Op: model.EmailAttemptTransfer, ToQueueId: 5}) != 5 ||
		emailTargetQueue(&model.EmailAttemptCommand{Op: model.EmailAttemptTransfer, ToAgentId: 4}) != 0 {
		t.Error("only a transfer to a queue needs cc_queue read")
	}
}
