package app

import (
	"context"

	"github.com/webitel/engine/model"
)

func (a *App) CreateRoutingSchema(ctx context.Context, scheme *model.RoutingSchema) (*model.RoutingSchema, model.AppError) {
	return a.Store.RoutingSchema().Create(ctx, scheme)
}

func (a *App) GetRoutingSchemaPage(ctx context.Context, domainId int64, search *model.SearchRoutingSchema) ([]*model.RoutingSchema, bool, model.AppError) {
	list, err := a.Store.RoutingSchema().GetAllPage(ctx, domainId, search)
	if err != nil {
		return nil, false, err
	}
	search.RemoveLastElemIfNeed(&list)
	return list, search.EndOfList(), nil
}

func (app *App) GetRoutingSchemaById(ctx context.Context, domainId, id int64) (*model.RoutingSchema, model.AppError) {
	return app.Store.RoutingSchema().Get(ctx, domainId, id)
}

func (a *App) UpdateRoutingSchema(ctx context.Context, scheme *model.RoutingSchema) (*model.RoutingSchema, model.AppError) {
	oldScheme, err := a.GetRoutingSchemaById(ctx, scheme.DomainId, scheme.Id)
	if err != nil {
		return nil, err
	}

	oldScheme.Name = scheme.Name
	oldScheme.Type = scheme.Type
	oldScheme.Debug = scheme.Debug
	oldScheme.Description = scheme.Description
	oldScheme.Payload = scheme.Payload
	oldScheme.Schema = scheme.Schema
	oldScheme.Tags = scheme.Tags

	oldScheme.UpdatedAt = scheme.UpdatedAt
	oldScheme.UpdatedBy = scheme.UpdatedBy

	oldScheme, err = a.Store.RoutingSchema().Update(ctx, oldScheme)
	if err != nil {
		return nil, err
	}

	return oldScheme, nil
}

func (a *App) RemoveRoutingSchema(ctx context.Context, domainId, id int64) (*model.RoutingSchema, model.AppError) {
	scheme, err := a.Store.RoutingSchema().Get(ctx, domainId, id)

	if err != nil {
		return nil, err
	}

	err = a.Store.RoutingSchema().Delete(ctx, domainId, id)
	if err != nil {
		return nil, err
	}
	return scheme, nil
}

func (a *App) PatchRoutingSchema(ctx context.Context, domainId, id int64, patch *model.RoutingSchemaPath) (*model.RoutingSchema, model.AppError) {
	old, err := a.GetRoutingSchemaById(ctx, domainId, id)
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

	old, err = a.Store.RoutingSchema().Update(ctx, old)
	if err != nil {
		return nil, err
	}

	return old, nil
}

func (a *App) GetRoutingSchemaTagsPage(ctx context.Context, domainId int64, search *model.SearchRoutingSchemaTag) ([]*model.RoutingSchemaTag, bool, model.AppError) {
	list, err := a.Store.RoutingSchema().ListTags(ctx, domainId, search)
	if err != nil {
		return nil, false, err
	}
	search.RemoveLastElemIfNeed(&list)
	return list, search.EndOfList(), nil
}
