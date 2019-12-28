package app

import "github.com/webitel/engine/model"

func (app *App) CreateRoutingOutboundCall(routing *model.RoutingOutboundCall) (*model.RoutingOutboundCall, *model.AppError) {
	return app.Store.RoutingOutboundCall().Create(routing)
}

func (app *App) GetRoutingOutboundCallPage(domainId int64, page, perPage int) ([]*model.RoutingOutboundCall, *model.AppError) {
	return app.Store.RoutingOutboundCall().GetAllPage(domainId, page*perPage, perPage)
}

func (app *App) GetRoutingOutboundCallById(domainId, id int64) (*model.RoutingOutboundCall, *model.AppError) {
	return app.Store.RoutingOutboundCall().Get(domainId, id)
}

func (app *App) UpdateRoutingOutboundCall(routing *model.RoutingOutboundCall) (*model.RoutingOutboundCall, *model.AppError) {
	oldRouting, err := app.GetRoutingOutboundCallById(routing.DomainId, routing.Id)
	if err != nil {
		return nil, err
	}

	oldRouting.Name = routing.Name
	oldRouting.Description = routing.Description
	oldRouting.Pattern = routing.Pattern
	oldRouting.Priority = routing.Priority
	oldRouting.Disabled = routing.Disabled
	oldRouting.UpdatedAt = routing.UpdatedAt

	if routing.GetSchemaId() != nil {
		oldRouting.Schema.Id = routing.Schema.Id
	}

	oldRouting, err = app.Store.RoutingOutboundCall().Update(oldRouting)
	if err != nil {
		return nil, err
	}

	return oldRouting, nil
}

func (a *App) RemoveRoutingOutboundCall(domainId, id int64) (*model.RoutingOutboundCall, *model.AppError) {
	routing, err := a.Store.RoutingOutboundCall().Get(domainId, id)

	if err != nil {
		return nil, err
	}

	err = a.Store.RoutingOutboundCall().Delete(domainId, id)
	if err != nil {
		return nil, err
	}
	return routing, nil
}

func (a *App) PatchRoutingOutboundCall(domainId, id int64, patch *model.RoutingOutboundCallPatch) (*model.RoutingOutboundCall, *model.AppError) {
	old, err := a.GetRoutingOutboundCallById(domainId, id)
	if err != nil {
		return nil, err
	}
	old.Patch(patch)

	old.UpdatedAt = model.GetMillis()
	old.UpdatedBy.Id = patch.UpdatedById

	if err = old.IsValid(); err != nil {
		return nil, err
	}

	old, err = a.Store.RoutingOutboundCall().Update(old)
	if err != nil {
		return nil, err
	}

	return old, nil
}
