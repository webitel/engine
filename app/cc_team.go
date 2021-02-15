package app

import (
	"github.com/webitel/engine/auth_manager"
	"github.com/webitel/engine/model"
)

func (a *App) AgentTeamCheckAccess(domainId, id int64, groups []int, access auth_manager.PermissionAccess) (bool, *model.AppError) {
	return a.Store.AgentTeam().CheckAccess(domainId, id, groups, access)
}

func (app *App) CreateAgentTeam(team *model.AgentTeam) (*model.AgentTeam, *model.AppError) {
	return app.Store.AgentTeam().Create(team)
}

func (a *App) GetAgentTeamsPage(domainId int64, search *model.SearchAgentTeam) ([]*model.AgentTeam, bool, *model.AppError) {
	list, err := a.Store.AgentTeam().GetAllPage(domainId, search)
	if err != nil {
		return nil, false, err
	}
	search.RemoveLastElemIfNeed(&list)
	return list, search.EndOfList(), nil
}

func (a *App) GetAgentTeamsPageByGroups(domainId int64, groups []int, search *model.SearchAgentTeam) ([]*model.AgentTeam, bool, *model.AppError) {
	list, err := a.Store.AgentTeam().GetAllPageByGroups(domainId, groups, search)
	if err != nil {
		return nil, false, err
	}
	search.RemoveLastElemIfNeed(&list)
	return list, search.EndOfList(), nil
}

func (a *App) GetAgentTeamById(domainId, id int64) (*model.AgentTeam, *model.AppError) {
	return a.Store.AgentTeam().Get(domainId, id)
}

func (a *App) UpdateAgentTeam(domainId int64, team *model.AgentTeam) (*model.AgentTeam, *model.AppError) {
	oldTeam, err := a.GetAgentTeamById(team.DomainId, team.Id)
	if err != nil {
		return nil, err
	}

	oldTeam.Name = team.Name
	oldTeam.Description = team.Description
	oldTeam.Strategy = team.Strategy
	oldTeam.MaxNoAnswer = team.MaxNoAnswer
	oldTeam.WrapUpTime = team.WrapUpTime
	oldTeam.NoAnswerDelayTime = team.NoAnswerDelayTime
	oldTeam.CallTimeout = team.CallTimeout
	oldTeam.Administrator = team.Administrator

	oldTeam.UpdatedAt = team.UpdatedAt
	oldTeam.UpdatedBy.Id = team.UpdatedBy.Id

	oldTeam, err = a.Store.AgentTeam().Update(domainId, oldTeam)
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
