package app

import (
	"context"

	"github.com/webitel/engine/model"
)

func (app *App) CreateRoutingOutboundCall(ctx context.Context, routing *model.RoutingOutboundCall) (*model.RoutingOutboundCall, model.AppError) {
	return app.Store.RoutingOutboundCall().Create(ctx, routing)
}

func (app *App) GetRoutingOutboundCallPage(ctx context.Context, domainId int64, search *model.SearchRoutingOutboundCall) ([]*model.RoutingOutboundCall, bool, model.AppError) {
	list, err := app.Store.RoutingOutboundCall().GetAllPage(ctx, domainId, search)
	if err != nil {
		return nil, false, err
	}
	search.RemoveLastElemIfNeed(&list)
	return list, search.EndOfList(), nil
}

func (app *App) GetRoutingOutboundCallById(ctx context.Context, domainId, id int64) (*model.RoutingOutboundCall, model.AppError) {
	return app.Store.RoutingOutboundCall().Get(ctx, domainId, id)
}

func (app *App) UpdateRoutingOutboundCall(ctx context.Context, routing *model.RoutingOutboundCall) (*model.RoutingOutboundCall, model.AppError) {
	oldRouting, err := app.GetRoutingOutboundCallById(ctx, routing.DomainId, routing.Id)
	if err != nil {
		return nil, err
	}

	oldRouting.Name = routing.Name
	oldRouting.Description = routing.Description
	oldRouting.Pattern = routing.Pattern
	oldRouting.Disabled = routing.Disabled
	oldRouting.UpdatedAt = routing.UpdatedAt

	if routing.GetSchemaId() != nil {
		oldRouting.Schema.Id = routing.Schema.Id
	}

	oldRouting, err = app.Store.RoutingOutboundCall().Update(ctx, oldRouting)
	if err != nil {
		return nil, err
	}

	return oldRouting, nil
}

func (app *App) ChangePositionOutboundCall(ctx context.Context, domainId, fromId, toId int64) model.AppError {
	return app.Store.RoutingOutboundCall().ChangePosition(ctx, domainId, fromId, toId)
}

func (a *App) RemoveRoutingOutboundCall(ctx context.Context, domainId, id int64) (*model.RoutingOutboundCall, model.AppError) {
	routing, err := a.Store.RoutingOutboundCall().Get(ctx, domainId, id)

	if err != nil {
		return nil, err
	}

	err = a.Store.RoutingOutboundCall().Delete(ctx, domainId, id)
	if err != nil {
		return nil, err
	}
	return routing, nil
}

func (a *App) PatchRoutingOutboundCall(ctx context.Context, domainId, id int64, patch *model.RoutingOutboundCallPatch) (*model.RoutingOutboundCall, model.AppError) {
	old, err := a.GetRoutingOutboundCallById(ctx, domainId, id)
	if err != nil {
		return nil, err
	}
	old.Patch(patch)

	old.UpdatedBy = &model.Lookup{
		Id: patch.UpdatedById,
	}

	old.UpdatedAt = model.GetMillis()

	if err = old.IsValid(); err != nil {
		return nil, err
	}

	old, err = a.Store.RoutingOutboundCall().Update(ctx, old)
	if err != nil {
		return nil, err
	}

	return old, nil
}
