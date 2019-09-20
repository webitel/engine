package app

import "github.com/webitel/engine/model"

func (a *App) CreateRoutingScheme(scheme *model.RoutingScheme) (*model.RoutingScheme, *model.AppError) {
	return a.Store.RoutingScheme().Create(scheme)
}

func (a *App) GetRoutingSchemePage(domainId int64, page, perPage int) ([]*model.RoutingScheme, *model.AppError) {
	return a.Store.RoutingScheme().GetAllPage(domainId, page*perPage, perPage)
}

func (app *App) GetRoutingSchemeById(domainId, id int64) (*model.RoutingScheme, *model.AppError) {
	return app.Store.RoutingScheme().Get(domainId, id)
}

func (a *App) UpdateRoutingScheme(scheme *model.RoutingScheme) (*model.RoutingScheme, *model.AppError) {
	oldScheme, err := a.GetRoutingSchemeById(scheme.DomainId, scheme.Id)
	if err != nil {
		return nil, err
	}

	oldScheme.Name = scheme.Name
	oldScheme.Type = scheme.Type
	oldScheme.Debug = scheme.Debug
	oldScheme.Description = scheme.Description

	oldScheme.UpdatedAt = scheme.UpdatedAt
	oldScheme.UpdatedBy = model.Lookup{
		Id: scheme.UpdatedBy.Id,
	}

	oldScheme, err = a.Store.RoutingScheme().Update(oldScheme)
	if err != nil {
		return nil, err
	}

	return oldScheme, nil
}

func (a *App) RemoveRoutingScheme(domainId, id int64) (*model.RoutingScheme, *model.AppError) {
	scheme, err := a.Store.RoutingScheme().Get(domainId, id)

	if err != nil {
		return nil, err
	}

	err = a.Store.RoutingScheme().Delete(domainId, id)
	if err != nil {
		return nil, err
	}
	return scheme, nil
}
