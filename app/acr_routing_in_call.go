package app

import "github.com/webitel/engine/model"

func (app *App) CreateRoutingInboundCall(routing *model.RoutingInboundCall) (*model.RoutingInboundCall, *model.AppError) {
	return app.Store.RoutingInboundCall().Create(routing)
}

func (app *App) GetRoutingInboundCallPage(domainId int64, page, perPage int) ([]*model.RoutingInboundCall, *model.AppError) {
	return app.Store.RoutingInboundCall().GetAllPage(domainId, page*perPage, perPage)
}

func (app *App) GetRoutingInboundCallById(domainId, id int64) (*model.RoutingInboundCall, *model.AppError) {
	return app.Store.RoutingInboundCall().Get(domainId, id)
}

func (app *App) UpdateRoutingInboundCall(routing *model.RoutingInboundCall) (*model.RoutingInboundCall, *model.AppError) {
	oldRouting, err := app.GetRoutingInboundCallById(routing.DomainId, routing.Id)
	if err != nil {
		return nil, err
	}

	oldRouting.Name = routing.Name
	oldRouting.Description = routing.Description
	oldRouting.Numbers = routing.Numbers
	oldRouting.Host = routing.Host
	oldRouting.Disabled = routing.Disabled
	oldRouting.Debug = routing.Debug
	oldRouting.Timezone.Id = routing.Timezone.Id

	if routing.GetStartSchemaId() != nil {
		oldRouting.StartSchema.Id = routing.StartSchema.Id
	}
	if routing.GetStopSchemaId() != nil {
		oldRouting.StopSchema = &model.Lookup{
			Id: routing.StopSchema.Id,
		}
	} else {
		oldRouting.StopSchema = nil
	}

	oldRouting, err = app.Store.RoutingInboundCall().Update(oldRouting)
	if err != nil {
		return nil, err
	}

	return oldRouting, nil
}

func (a *App) RemoveRoutingInboundCall(domainId, id int64) (*model.RoutingInboundCall, *model.AppError) {
	routing, err := a.Store.RoutingInboundCall().Get(domainId, id)

	if err != nil {
		return nil, err
	}

	err = a.Store.RoutingInboundCall().Delete(domainId, id)
	if err != nil {
		return nil, err
	}
	return routing, nil
}
