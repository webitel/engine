package app

import (
	"context"
	"errors"

	"github.com/webitel/wlog"

	"github.com/webitel/engine/model"
)

func (app *App) initDomainEventListener() {
	app.MessageQueue.SetDomainsEventHandler(app.handleDomainEventCreated)
}

func (app *App) handleDomainEventCreated(ctx context.Context, e *model.DomainEvent) error {
	err := errors.Join(
		app.Store.OnlineSkills().CreateSystem(ctx, e.ID),
		app.Store.SystemSettings().CreateDefaults(ctx, e.ID),
	)
	if err != nil {
		app.Log.Error("processing domain created event", wlog.Err(err))

		return err
	}

	return nil
}
