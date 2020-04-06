package app

import (
	"github.com/webitel/engine/auth_manager"
	"github.com/webitel/engine/model"
	"net/http"
)

func (a *App) AgentCheckAccess(domainId, id int64, groups []int, access auth_manager.PermissionAccess) (bool, *model.AppError) {
	return a.Store.Agent().CheckAccess(domainId, id, groups, access)
}

func (app *App) CreateAgent(agent *model.Agent) (*model.Agent, *model.AppError) {
	return app.Store.Agent().Create(agent)
}

func (a *App) GetAgentsPage(domainId int64, search *model.SearchAgent) ([]*model.Agent, bool, *model.AppError) {
	list, err := a.Store.Agent().GetAllPage(domainId, search)
	if err != nil {
		return nil, false, err
	}
	search.RemoveLastElemIfNeed(&list)
	return list, search.EndOfList(), nil
}

func (a *App) GetAgentsPageByGroups(domainId int64, groups []int, search *model.SearchAgent) ([]*model.Agent, bool, *model.AppError) {
	list, err := a.Store.Agent().GetAllPageByGroups(domainId, groups, search)
	if err != nil {
		return nil, false, err
	}
	search.RemoveLastElemIfNeed(&list)
	return list, search.EndOfList(), nil
}

func (a *App) GetAgentStateHistoryPage(agentId int64, search *model.SearchAgentState) ([]*model.AgentState, bool, *model.AppError) {
	list, err := a.Store.Agent().HistoryState(agentId, search)
	if err != nil {
		return nil, false, err
	}
	search.RemoveLastElemIfNeed(&list)
	return list, search.EndOfList(), nil
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
	oldAgent.ProgressiveCount = agent.ProgressiveCount
	oldAgent.User.Id = agent.User.Id

	oldAgent.UpdatedAt = agent.UpdatedAt
	oldAgent.UpdatedBy.Id = agent.UpdatedBy.Id

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

func (a *App) GetAgentInTeamPage(domainId, id int64, search *model.SearchAgentInTeam) ([]*model.AgentInTeam, bool, *model.AppError) {
	list, err := a.Store.Agent().InTeam(domainId, id, search)
	if err != nil {
		return nil, false, err
	}
	search.RemoveLastElemIfNeed(&list)
	return list, search.EndOfList(), nil
}

func (a *App) GetAgentInQueuePage(domainId, id int64, search *model.SearchAgentInQueue) ([]*model.AgentInQueue, bool, *model.AppError) {
	list, err := a.Store.Agent().InQueue(domainId, id, search)
	if err != nil {
		return nil, false, err
	}
	search.RemoveLastElemIfNeed(&list)
	return list, search.EndOfList(), nil
}

func (a *App) GetAgentInQueueStatistics(domainId, agentId int64) ([]*model.AgentInQueueStatistic, *model.AppError) {
	return a.Store.Agent().QueueStatistic(domainId, agentId)
}

func (a *App) AgentsLookupNotExistsUsers(domainId int64, search *model.SearchAgentUser) ([]*model.AgentUser, bool, *model.AppError) {
	list, err := a.Store.Agent().LookupNotExistsUsers(domainId, search)
	if err != nil {
		return nil, false, err
	}
	search.RemoveLastElemIfNeed(&list)
	return list, search.EndOfList(), nil
}

func (a *App) AgentsLookupNotExistsUsersByGroups(domainId int64, groups []int, search *model.SearchAgentUser) ([]*model.AgentUser, bool, *model.AppError) {
	list, err := a.Store.Agent().LookupNotExistsUsersByGroups(domainId, groups, search)
	if err != nil {
		return nil, false, err
	}
	search.RemoveLastElemIfNeed(&list)
	return list, search.EndOfList(), nil
}

func (a *App) GetAgentSession(domainId, id int64) (*model.AgentSession, *model.AppError) {
	return a.Store.Agent().GetSession(domainId, id)
}

func (a *App) LoginAgent(domainId, agentId int64) *model.AppError {
	err := a.cc.Agent().Login(domainId, agentId)
	if err != nil {
		return model.NewAppError("LoginAgent", "app.agent.login.app_err", nil, err.Error(), http.StatusBadRequest)
	}

	return nil
}

func (a *App) LogoutAgent(domainId, agentId int64) *model.AppError {
	err := a.cc.Agent().Logout(domainId, agentId)
	if err != nil {
		return model.NewAppError("LogoutAgent", "app.agent.logout.app_err", nil, err.Error(), http.StatusBadRequest)
	}

	return nil
}

func (a *App) PauseAgent(domainId, agentId int64, payload string, timeout int) *model.AppError {
	err := a.cc.Agent().Pause(domainId, agentId, payload, timeout)
	if err != nil {
		return model.NewAppError("PauseAgent", "app.agent.pause.app_err", nil, err.Error(), http.StatusBadRequest)
	}

	return nil
}
