package app

import (
	"context"

	"github.com/webitel/engine/pkg/wbt/auth_manager"
	"github.com/webitel/wlog"
)

func (app *App) AuditCreate(ctx context.Context, session *auth_manager.Session, object string, recordId string, data any) {
	if err := app.audit.Create(ctx, session, object, recordId, data); err != nil {
		wlog.Error("executing audit create request", wlog.String("object", object), wlog.String("record_id", recordId), wlog.Err(err))
	}
}

func (app *App) AuditUpdate(ctx context.Context, session *auth_manager.Session, object string, recordId string, data any) {
	if err := app.audit.Update(ctx, session, object, recordId, data); err != nil {
		wlog.Error("executing audit update request", wlog.String("object", object), wlog.String("record_id", recordId), wlog.Err(err))
	}
}

func (app *App) AuditDelete(ctx context.Context, session *auth_manager.Session, object string, recordId string, data any) {
	if err := app.audit.Delete(ctx, session, object, recordId, data); err != nil {
		wlog.Error("executing audit delete request", wlog.String("object", object), wlog.String("record_id", recordId), wlog.Err(err))
	}
}
