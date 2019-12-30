package app

import "github.com/webitel/engine/model"

func (a *App) CreateResourceTeamAgent(resource *model.ResourceInTeam) (*model.ResourceInTeam, *model.AppError) {
	return a.Store.ResourceTeam().Create(resource)
}

func (a *App) GetResourceTeamAgentPage(domainId, teamId int64, page, perPage int) ([]*model.ResourceInTeam, *model.AppError) {
	return a.Store.ResourceTeam().GetAllPage(domainId, teamId, page*perPage, perPage, true)
}

func (app *App) GetResourceTeamAgent(domainId, teamId, id int64) (*model.ResourceInTeam, *model.AppError) {
	return app.Store.ResourceTeam().Get(domainId, teamId, id)
}

func (app *App) UpdateResourceTeamAgent(domainId int64, resource *model.ResourceInTeam) (*model.ResourceInTeam, *model.AppError) {
	oldRes, err := app.Store.ResourceTeam().Get(domainId, resource.TeamId, resource.Id)

	if err != nil {
		return nil, err
	}

	oldRes.Agent = resource.Agent
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

func (a *App) RemoveResourceTeamAgent(domainId, teamId, id int64) (*model.ResourceInTeam, *model.AppError) {
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

//skills
func (a *App) CreateResourceTeamSkill(resource *model.ResourceInTeam) (*model.ResourceInTeam, *model.AppError) {
	return a.Store.ResourceTeam().Create(resource)
}

func (a *App) GetResourceTeamSkillPage(domainId, teamId int64, page, perPage int) ([]*model.ResourceInTeam, *model.AppError) {
	return a.Store.ResourceTeam().GetAllPage(domainId, teamId, page*perPage, perPage, false)
}

func (app *App) GetResourceTeamSkill(domainId, teamId, id int64) (*model.ResourceInTeam, *model.AppError) {
	return app.Store.ResourceTeam().Get(domainId, teamId, id)
}

func (app *App) UpdateResourceTeamSkill(domainId int64, resource *model.ResourceInTeam) (*model.ResourceInTeam, *model.AppError) {
	oldRes, err := app.Store.ResourceTeam().Get(domainId, resource.TeamId, resource.Id)

	if err != nil {
		return nil, err
	}

	oldRes.Skill = resource.Skill
	oldRes.Agent = nil
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

func (a *App) RemoveResourceTeamSkill(domainId, teamId, id int64) (*model.ResourceInTeam, *model.AppError) {
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
