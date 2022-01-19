package app

import "github.com/webitel/engine/model"

func (a *App) CreateRoutingSchema(scheme *model.RoutingSchema) (*model.RoutingSchema, *model.AppError) {
	return a.Store.RoutingSchema().Create(scheme)
}

func (a *App) GetRoutingSchemaPage(domainId int64, search *model.SearchRoutingSchema) ([]*model.RoutingSchema, bool, *model.AppError) {
	list, err := a.Store.RoutingSchema().GetAllPage(domainId, search)
	if err != nil {
		return nil, false, err
	}
	search.RemoveLastElemIfNeed(&list)
	return list, search.EndOfList(), nil
}

func (app *App) GetRoutingSchemaById(domainId, id int64) (*model.RoutingSchema, *model.AppError) {
	return app.Store.RoutingSchema().Get(domainId, id)
}

func (a *App) UpdateRoutingSchema(scheme *model.RoutingSchema) (*model.RoutingSchema, *model.AppError) {
	oldScheme, err := a.GetRoutingSchemaById(scheme.DomainId, scheme.Id)
	if err != nil {
		return nil, err
	}

	oldScheme.Name = scheme.Name
	oldScheme.Type = scheme.Type
	oldScheme.Debug = scheme.Debug
	oldScheme.Description = scheme.Description
	oldScheme.Payload = scheme.Payload
	oldScheme.Schema = scheme.Schema

	oldScheme.UpdatedAt = scheme.UpdatedAt
	oldScheme.UpdatedBy = scheme.UpdatedBy

	oldScheme, err = a.Store.RoutingSchema().Update(oldScheme)
	if err != nil {
		return nil, err
	}

	return oldScheme, nil
}

func (a *App) RemoveRoutingSchema(domainId, id int64) (*model.RoutingSchema, *model.AppError) {
	scheme, err := a.Store.RoutingSchema().Get(domainId, id)

	if err != nil {
		return nil, err
	}

	err = a.Store.RoutingSchema().Delete(domainId, id)
	if err != nil {
		return nil, err
	}
	return scheme, nil
}

func (a *App) PatchRoutingSchema(domainId, id int64, patch *model.RoutingSchemaPath) (*model.RoutingSchema, *model.AppError) {
	old, err := a.GetRoutingSchemaById(domainId, id)
	if err != nil {
		return nil, err
	}

	old.Patch(patch)

	old.UpdatedAt = model.GetMillis()
	old.UpdatedBy = &model.Lookup{
		Id: patch.UpdatedById,
	}

	if err = old.IsValid(); err != nil {
		return nil, err
	}

	old, err = a.Store.RoutingSchema().Update(old)
	if err != nil {
		return nil, err
	}

	return old, nil
}
