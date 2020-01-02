package app

import (
	"github.com/webitel/engine/model"
	"net/http"
)

func (a *App) AgentCheckAccess(domainId, id int64, groups []int, access model.PermissionAccess) (bool, *model.AppError) {
	return a.Store.Agent().CheckAccess(domainId, id, groups, access)
}

func (app *App) CreateAgent(agent *model.Agent) (*model.Agent, *model.AppError) {
	return app.Store.Agent().Create(agent)
}

func (a *App) GetAgentsPage(domainId int64, page, perPage int) ([]*model.Agent, *model.AppError) {
	return a.Store.Agent().GetAllPage(domainId, page*perPage, perPage)
}

func (a *App) GetAgentsPageByGroups(domainId int64, groups []int, page, perPage int) ([]*model.Agent, *model.AppError) {
	return a.Store.Agent().GetAllPageByGroups(domainId, groups, page*perPage, perPage)
}

func (a *App) GetAgentStateHistoryPage(agentId, from, to int64, page, perPage int) ([]*model.AgentState, *model.AppError) {
	return a.Store.Agent().HistoryState(agentId, from, to, page*perPage, perPage)
}

func (a *App) GetAgentById(domainId, id int64) (*model.Agent, *model.AppError) {
	return a.Store.Agent().Get(domainId, id)
}

func (a *App) UpdateAgent(agent *model.Agent) (*model.Agent, *model.AppError) {
	oldAgent, err := a.GetAgentById(agent.DomainId, agent.Id)
	if err != nil {
		return nil, err
	}

	oldAgent.Description = agent.Description
	oldAgent.User.Id = agent.User.Id

	oldAgent, err = a.Store.Agent().Update(oldAgent)
	if err != nil {
		return nil, err
	}

	return oldAgent, nil
}

func (a *App) RemoveAgent(domainId, id int64) (*model.Agent, *model.AppError) {
	agent, err := a.GetAgentById(domainId, id)

	if err != nil {
		return nil, err
	}

	err = a.Store.Agent().Delete(domainId, id)
	if err != nil {
		return nil, err
	}
	return agent, nil
}

func (a *App) SetAgentStatus(domainId, id int64, status model.AgentStatus) *model.AppError {
	switch status {
	case model.AgentStatusOnline, model.AgentStatusOffline, model.AgentStatusPause:
		//FIXME fire event ?
		_, err := a.Store.Agent().SetStatus(domainId, id, status.String(), nil)
		return err
	default:
		return model.NewAppError("SetAgentStatus.IsValid", "app.set_agent_status.is_valid.status.app_error", nil, status.String(), http.StatusBadRequest)
	}
}

func (a *App) GetAgentInTeamPage(domainId, id int64, page, perPage int) ([]*model.AgentInTeam, *model.AppError) {
	return a.Store.Agent().InTeam(domainId, id, page*perPage, perPage)
}

func (a *App) GetAgentInQueuePage(domainId, id int64, page, perPage int) ([]*model.AgentInQueue, *model.AppError) {
	return a.Store.Agent().InQueue(domainId, id, page*perPage, perPage)
}
