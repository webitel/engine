package app

import "github.com/webitel/engine/model"

func (a *App) CreateResourceTeam(resource *model.ResourceInTeam) (*model.ResourceInTeam, *model.AppError) {
	return a.Store.ResourceTeam().Create(resource)
}

func (a *App) GetResourceTeamPage(domainId, teamId int64, page, perPage int) ([]*model.ResourceInTeam, *model.AppError) {
	return a.Store.ResourceTeam().GetAllPage(domainId, teamId, page*perPage, perPage)
}

func (app *App) GetResourceTeam(domainId, teamId, id int64) (*model.ResourceInTeam, *model.AppError) {
	return app.Store.ResourceTeam().Get(domainId, teamId, id)
}

func (app *App) UpdateResourceTeam(domainId int64, resource *model.ResourceInTeam) (*model.ResourceInTeam, *model.AppError) {
	oldRes, err := app.Store.ResourceTeam().Get(domainId, resource.TeamId, resource.Id)

	if err != nil {
		return nil, err
	}

	oldRes.Agent = resource.Agent
	oldRes.Skill = resource.Skill
	oldRes.Lvl = resource.Lvl
	oldRes.MinCapacity = resource.MinCapacity
	oldRes.MaxCapacity = resource.MaxCapacity
	oldRes.Bucket = resource.Bucket

	_, err = app.Store.ResourceTeam().Update(oldRes)
	if err != nil {
		return nil, err
	}

	return oldRes, nil
}

func (a *App) RemoveResourceTeam(domainId, teamId, id int64) (*model.ResourceInTeam, *model.AppError) {
	resource, err := a.Store.ResourceTeam().Get(domainId, teamId, id)

	if err != nil {
		return nil, err
	}

	err = a.Store.ResourceTeam().Delete(domainId, teamId, id)
	if err != nil {
		return nil, err
	}
	return resource, nil
}
