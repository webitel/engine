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
	oldRouting.Debug = routing.Debug
	oldRouting.UpdatedAt = routing.UpdatedAt

	if routing.GetSchemeId() != nil {
		oldRouting.Scheme.Id = routing.Scheme.Id
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
