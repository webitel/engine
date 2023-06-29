package app

import (
	"context"

	"github.com/webitel/engine/auth_manager"
	"github.com/webitel/engine/model"
)

func (a *App) AgentTeamCheckAccess(ctx context.Context, domainId, id int64, groups []int, access auth_manager.PermissionAccess) (bool, model.AppError) {
	return a.Store.AgentTeam().CheckAccess(ctx, domainId, id, groups, access)
}

func (app *App) CreateAgentTeam(ctx context.Context, team *model.AgentTeam) (*model.AgentTeam, model.AppError) {
	return app.Store.AgentTeam().Create(ctx, team)
}

func (a *App) GetAgentTeamsPage(ctx context.Context, domainId int64, search *model.SearchAgentTeam) ([]*model.AgentTeam, bool, model.AppError) {
	list, err := a.Store.AgentTeam().GetAllPage(ctx, domainId, search)
	if err != nil {
		return nil, false, err
	}
	search.RemoveLastElemIfNeed(&list)
	return list, search.EndOfList(), nil
}

func (a *App) GetAgentTeamsPageByGroups(ctx context.Context, domainId int64, groups []int, search *model.SearchAgentTeam) ([]*model.AgentTeam, bool, model.AppError) {
	list, err := a.Store.AgentTeam().GetAllPageByGroups(ctx, domainId, groups, search)
	if err != nil {
		return nil, false, err
	}
	search.RemoveLastElemIfNeed(&list)
	return list, search.EndOfList(), nil
}

func (a *App) GetAgentTeamById(ctx context.Context, domainId, id int64) (*model.AgentTeam, model.AppError) {
	return a.Store.AgentTeam().Get(ctx, domainId, id)
}

func (a *App) UpdateAgentTeam(ctx context.Context, domainId int64, team *model.AgentTeam) (*model.AgentTeam, model.AppError) {
	oldTeam, err := a.GetAgentTeamById(ctx, team.DomainId, team.Id)
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
	oldTeam.InviteChatTimeout = team.InviteChatTimeout
	oldTeam.Admin = team.Admin

	oldTeam.UpdatedAt = team.UpdatedAt
	oldTeam.UpdatedBy = team.UpdatedBy

	oldTeam, err = a.Store.AgentTeam().Update(ctx, domainId, oldTeam)
	if err != nil {
		return nil, err
	}

	return oldTeam, nil
}

func (a *App) RemoveAgentTeam(ctx context.Context, domainId, id int64) (*model.AgentTeam, model.AppError) {
	team, err := a.Store.AgentTeam().Get(ctx, domainId, id)

	if err != nil {
		return nil, err
	}

	err = a.Store.AgentTeam().Delete(ctx, domainId, id)
	if err != nil {
		return nil, err
	}
	return team, nil
}
