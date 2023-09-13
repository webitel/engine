package app

import (
	"context"
	"fmt"
	"github.com/webitel/engine/auth_manager"
	"github.com/webitel/wlog"
)

func (app *App) AuditCreate(ctx context.Context, session *auth_manager.Session, object string, recordId int64, data interface{}) {
	err := app.audit.Create(ctx, session, object, recordId, data)
	if err != nil {
		wlog.Error(fmt.Sprintf("audit [create] object=%s, recordId=%d, error: %s", object, recordId, err.Error()))
	}
}

func (app *App) AuditUpdate(ctx context.Context, session *auth_manager.Session, object string, recordId int64, data interface{}) {
	err := app.audit.Update(ctx, session, object, recordId, data)
	if err != nil {
		wlog.Error(fmt.Sprintf("audit [update] object=%s, recordId=%d, error: %s", object, recordId, err.Error()))
	}
}

func (app *App) AuditDelete(ctx context.Context, session *auth_manager.Session, object string, recordId int64, data interface{}) {
	err := app.audit.Delete(ctx, session, object, recordId, data)
	if err != nil {
		wlog.Error(fmt.Sprintf("audit [delete] object=%s, recordId=%d, error: %s", object, recordId, err.Error()))
	}
}
