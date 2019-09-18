package app

import "github.com/webitel/engine/model"

func (a *App) AgentTeamCheckAccess(domainId, id int64, groups []int, access model.PermissionAccess) (bool, *model.AppError) {
	return a.Store.AgentTeam().CheckAccess(domainId, id, groups, access)
}

func (app *App) CreateAgentTeam(team *model.AgentTeam) (*model.AgentTeam, *model.AppError) {
	return app.Store.AgentTeam().Create(team)
}

func (a *App) GetAgentTeamsPage(domainId int64, page, perPage int) ([]*model.AgentTeam, *model.AppError) {
	return a.Store.AgentTeam().GetAllPage(domainId, page*perPage, perPage)
}

func (a *App) GetAgentTeamsPageByGroups(domainId int64, groups []int, page, perPage int) ([]*model.AgentTeam, *model.AppError) {
	return a.Store.AgentTeam().GetAllPageByGroups(domainId, groups, page*perPage, perPage)
}

func (a *App) GetAgentTeamById(domainId, id int64) (*model.AgentTeam, *model.AppError) {
	return a.Store.AgentTeam().Get(domainId, id)
}

func (a *App) UpdateAgentTeam(team *model.AgentTeam) (*model.AgentTeam, *model.AppError) {
	oldTeam, err := a.GetAgentTeamById(team.DomainId, team.Id)
	if err != nil {
		return nil, err
	}

	oldTeam.Name = team.Name
	oldTeam.Description = team.Description
	oldTeam.Strategy = team.Strategy
	oldTeam.MaxNoAnswer = team.MaxNoAnswer
	oldTeam.WrapUpTime = team.WrapUpTime
	oldTeam.RejectDelayTime = team.RejectDelayTime
	oldTeam.BusyDelayTime = team.BusyDelayTime
	oldTeam.NoAnswerDelayTime = team.NoAnswerDelayTime
	oldTeam.CallTimeout = team.CallTimeout

	oldTeam, err = a.Store.AgentTeam().Update(oldTeam)
	if err != nil {
		return nil, err
	}

	return oldTeam, nil
}

func (a *App) RemoveAgentTeam(domainId, id int64) (*model.AgentTeam, *model.AppError) {
	team, err := a.Store.AgentTeam().Get(domainId, id)

	if err != nil {
		return nil, err
	}

	err = a.Store.AgentTeam().Delete(domainId, id)
	if err != nil {
		return nil, err
	}
	return team, nil
}
