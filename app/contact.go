package app

import (
	"context"
	"github.com/webitel/engine/model"
)

// TODO channel: call/chat/task

func (app *App) SetContactId(ctx context.Context, domainId int64, channel string, id string, contactId int64) model.AppError {

	return app.Store.Call().SetContactId(ctx, domainId, id, contactId)
}
