package app

import "github.com/webitel/engine/model"

func (a *App) CreateRoutingVariable(variable *model.RoutingVariable) (*model.RoutingVariable, *model.AppError) {
	return a.Store.RoutingVariable().Create(variable)
}

func (a *App) GetRoutingVariablesPage(domainId int64, page, perPage int) ([]*model.RoutingVariable, *model.AppError) {
	return a.Store.RoutingVariable().GetAllPage(domainId, page*perPage, perPage)
}

func (app *App) GetRoutingVariableById(domainId, id int64) (*model.RoutingVariable, *model.AppError) {
	return app.Store.RoutingVariable().Get(domainId, id)
}

func (a *App) UpdateRoutingVariable(variable *model.RoutingVariable) (*model.RoutingVariable, *model.AppError) {
	oldVar, err := a.GetRoutingVariableById(variable.DomainId, variable.Id)
	if err != nil {
		return nil, err
	}

	oldVar.Key = variable.Key
	oldVar.Value = variable.Value

	oldVar, err = a.Store.RoutingVariable().Update(oldVar)
	if err != nil {
		return nil, err
	}

	return oldVar, nil
}

func (a *App) RemoveRoutingVariable(domainId, id int64) (*model.RoutingVariable, *model.AppError) {
	variable, err := a.Store.RoutingVariable().Get(domainId, id)

	if err != nil {
		return nil, err
	}

	err = a.Store.RoutingVariable().Delete(domainId, id)
	if err != nil {
		return nil, err
	}
	return variable, nil
}
