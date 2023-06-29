package app

import (
	"context"

	"github.com/webitel/engine/model"
)

func (a *App) CreateRoutingVariable(ctx context.Context, variable *model.RoutingVariable) (*model.RoutingVariable, model.AppError) {
	return a.Store.RoutingVariable().Create(ctx, variable)
}

func (a *App) GetRoutingVariablesPage(ctx context.Context, domainId int64, page, perPage int) ([]*model.RoutingVariable, model.AppError) {
	return a.Store.RoutingVariable().GetAllPage(ctx, domainId, page*perPage, perPage)
}

func (app *App) GetRoutingVariableById(ctx context.Context, domainId, id int64) (*model.RoutingVariable, model.AppError) {
	return app.Store.RoutingVariable().Get(ctx, domainId, id)
}

func (a *App) UpdateRoutingVariable(ctx context.Context, variable *model.RoutingVariable) (*model.RoutingVariable, model.AppError) {
	oldVar, err := a.GetRoutingVariableById(ctx, variable.DomainId, variable.Id)
	if err != nil {
		return nil, err
	}

	oldVar.Key = variable.Key
	oldVar.Value = variable.Value

	oldVar, err = a.Store.RoutingVariable().Update(ctx, oldVar)
	if err != nil {
		return nil, err
	}

	return oldVar, nil
}

func (a *App) RemoveRoutingVariable(ctx context.Context, domainId, id int64) (*model.RoutingVariable, model.AppError) {
	variable, err := a.Store.RoutingVariable().Get(ctx, domainId, id)

	if err != nil {
		return nil, err
	}

	err = a.Store.RoutingVariable().Delete(ctx, domainId, id)
	if err != nil {
		return nil, err
	}
	return variable, nil
}
