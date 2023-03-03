package app

import (
	"context"
	"github.com/webitel/engine/auth_manager"
	"github.com/webitel/engine/model"
	"net/http"
)

func (a *App) AgentCheckAccess(ctx context.Context, domainId, id int64, groups []int, access auth_manager.PermissionAccess) (bool, *model.AppError) {
	return a.Store.Agent().CheckAccess(ctx, domainId, id, groups, access)
}

func (a *App) CreateAgent(ctx context.Context, agent *model.Agent) (*model.Agent, *model.AppError) {
	return a.Store.Agent().Create(ctx, agent)
}

func (a *App) GetAgentsPage(ctx context.Context, domainId int64, search *model.SearchAgent) ([]*model.Agent, bool, *model.AppError) {
	list, err := a.Store.Agent().GetAllPage(ctx, domainId, search)
	if err != nil {
		return nil, false, err
	}
	search.RemoveLastElemIfNeed(&list)
	return list, search.EndOfList(), nil
}

func (a *App) GetAgentActiveTasks(ctx context.Context, domainId, agentId int64) ([]*model.CCTask, *model.AppError) {
	return a.Store.Agent().GetActiveTask(ctx, domainId, agentId)
}

func (a *App) GetAgentsPageByGroups(ctx context.Context, domainId int64, groups []int, search *model.SearchAgent) ([]*model.Agent, bool, *model.AppError) {
	list, err := a.Store.Agent().GetAllPageByGroups(ctx, domainId, groups, search)
	if err != nil {
		return nil, false, err
	}
	search.RemoveLastElemIfNeed(&list)
	return list, search.EndOfList(), nil
}

func (a *App) GetAgentStateHistoryPage(ctx context.Context, domainId int64, search *model.SearchAgentState) ([]*model.AgentState, bool, *model.AppError) {
	list, err := a.Store.Agent().HistoryState(ctx, domainId, search)
	if err != nil {
		return nil, false, err
	}
	search.RemoveLastElemIfNeed(&list)
	return list, search.EndOfList(), nil
}

func (a *App) GetAgentById(ctx context.Context, domainId, id int64) (*model.Agent, *model.AppError) {
	return a.Store.Agent().Get(ctx, domainId, id)
}

func (a *App) UpdateAgent(ctx context.Context, agent *model.Agent) (*model.Agent, *model.AppError) {
	oldAgent, err := a.GetAgentById(ctx, agent.DomainId, agent.Id)
	if err != nil {
		return nil, err
	}

	oldAgent.Description = agent.Description
	oldAgent.ProgressiveCount = agent.ProgressiveCount
	oldAgent.GreetingMedia = agent.GreetingMedia
	oldAgent.User.Id = agent.User.Id

	oldAgent.UpdatedAt = agent.UpdatedAt
	oldAgent.UpdatedBy = agent.UpdatedBy

	oldAgent.AllowChannels = agent.AllowChannels
	oldAgent.ChatCount = agent.ChatCount
	oldAgent.Supervisor = agent.Supervisor
	oldAgent.Team = agent.Team
	oldAgent.Region = agent.Region
	oldAgent.Auditor = agent.Auditor
	oldAgent.IsSupervisor = agent.IsSupervisor

	if err = oldAgent.IsValid(); err != nil {
		return nil, err
	}

	oldAgent, err = a.Store.Agent().Update(ctx, oldAgent)
	if err != nil {
		return nil, err
	}

	return oldAgent, nil
}

func (a *App) PatchAgent(ctx context.Context, domainId, id int64, patch *model.AgentPatch) (*model.Agent, *model.AppError) {
	oldAgent, err := a.GetAgentById(ctx, domainId, id)
	if err != nil {
		return nil, err
	}

	oldAgent.Patch(patch)

	if err = oldAgent.IsValid(); err != nil {
		return nil, err
	}

	oldAgent, err = a.Store.Agent().Update(ctx, oldAgent)
	if err != nil {
		return nil, err
	}

	return oldAgent, nil
}

func (a *App) RemoveAgent(ctx context.Context, domainId, id int64) (*model.Agent, *model.AppError) {
	agent, err := a.GetAgentById(ctx, domainId, id)

	if err != nil {
		return nil, err
	}

	err = a.Store.Agent().Delete(ctx, domainId, id)
	if err != nil {
		return nil, err
	}
	return agent, nil
}

func (a *App) GetAgentInQueuePage(ctx context.Context, domainId, id int64, search *model.SearchAgentInQueue) ([]*model.AgentInQueue, bool, *model.AppError) {
	list, err := a.Store.Agent().InQueue(ctx, domainId, id, search)
	if err != nil {
		return nil, false, err
	}
	search.RemoveLastElemIfNeed(&list)
	return list, search.EndOfList(), nil
}

func (a *App) GetAgentInQueueStatistics(ctx context.Context, domainId, agentId int64) ([]*model.AgentInQueueStatistic, *model.AppError) {
	return a.Store.Agent().QueueStatistic(ctx, domainId, agentId)
}

func (a *App) AgentsLookupNotExistsUsers(ctx context.Context, domainId int64, search *model.SearchAgentUser) ([]*model.AgentUser, bool, *model.AppError) {
	list, err := a.Store.Agent().LookupNotExistsUsers(ctx, domainId, search)
	if err != nil {
		return nil, false, err
	}
	search.RemoveLastElemIfNeed(&list)
	return list, search.EndOfList(), nil
}

func (a *App) AgentsLookupNotExistsUsersByGroups(ctx context.Context, domainId int64, groups []int, search *model.SearchAgentUser) ([]*model.AgentUser, bool, *model.AppError) {
	list, err := a.Store.Agent().LookupNotExistsUsersByGroups(ctx, domainId, groups, search)
	if err != nil {
		return nil, false, err
	}
	search.RemoveLastElemIfNeed(&list)
	return list, search.EndOfList(), nil
}

func (a *App) GetAgentSession(ctx context.Context, domainId, id int64) (*model.AgentSession, *model.AppError) {
	return a.Store.Agent().GetSession(ctx, domainId, id)
}

func (a *App) HasAgentCC(ctx context.Context, domainId int64, userId int64) *model.AppError {
	v, err := a.Store.Agent().HasAgentCC(ctx, domainId, userId)
	if err != nil {
		return err
	}

	return v.Valid()
}

func (a *App) LoginAgent(domainId, agentId int64, onDemand bool) *model.AppError {
	err := a.cc.Agent().Online(domainId, agentId, onDemand)
	if err != nil {
		return model.NewAppError("LoginAgent", "app.agent.login.app_err", nil, err.Error(), http.StatusBadRequest)
	}

	return nil
}

func (a *App) LogoutAgent(domainId, agentId int64) *model.AppError {
	err := a.cc.Agent().Offline(domainId, agentId)
	if err != nil {
		return model.NewAppError("LogoutAgent", "app.agent.logout.app_err", nil, err.Error(), http.StatusBadRequest)
	}

	return nil
}

func (a *App) WaitingAgentChannel(domainId int64, agentId int64, channel string) (int64, *model.AppError) {
	timestamp, err := a.cc.Agent().WaitingChannel(int(agentId), channel)
	if err != nil {
		return 0, model.NewAppError("WaitingAgentChannel", "app.agent.waiting.app_err", nil, err.Error(), http.StatusBadRequest)
	}

	return timestamp, nil
}

func (a *App) PauseAgent(domainId, agentId int64, payload string, timeout int) *model.AppError {
	err := a.cc.Agent().Pause(domainId, agentId, payload, timeout)
	if err != nil {
		return model.NewAppError("PauseAgent", "app.agent.pause.app_err", nil, err.Error(), http.StatusBadRequest)
	}

	return nil
}

func (a *App) GetAgentReportCall(ctx context.Context, domainId int64, search *model.SearchAgentCallStatistics) ([]*model.AgentCallStatistics, bool, *model.AppError) {
	list, err := a.Store.Agent().CallStatistics(ctx, domainId, search)
	if err != nil {
		return nil, false, err
	}
	search.RemoveLastElemIfNeed(&list)
	return list, search.EndOfList(), nil
}

func (a *App) GetAgentTodayStatistics(ctx context.Context, domainId, agentId int64) (*model.AgentStatistics, *model.AppError) {
	return a.Store.Agent().TodayStatistics(ctx, domainId, agentId)
}

func (a *App) GetAgentStatusStatistic(ctx context.Context, domainId int64, supervisorUserId int64, groups []int, access auth_manager.PermissionAccess, search *model.SearchAgentStatusStatistic) ([]*model.AgentStatusStatistics, bool, *model.AppError) {
	list, err := a.Store.Agent().StatusStatistic(ctx, domainId, supervisorUserId, groups, access, search)
	if err != nil {
		return nil, false, err
	}
	search.RemoveLastElemIfNeed(&list)
	return list, search.EndOfList(), nil
}

// agent_id check
func (a *App) AcceptTask(appId string, domainId int64, attemptId int64) *model.AppError {
	err := a.cc.Agent().AcceptTask(appId, domainId, attemptId)
	if err != nil {
		return model.NewAppError("AcceptTask", "app.cc.accept_task", nil, err.Error(), http.StatusInternalServerError)
	}
	return nil
}

func (a *App) CloseTask(appId string, domainId int64, attemptId int64) *model.AppError {
	err := a.cc.Agent().CloseTask(appId, domainId, attemptId)
	if err != nil {
		return model.NewAppError("CloseTask", "app.cc.close_task", nil, err.Error(), http.StatusInternalServerError)
	}
	return nil
}

func (a *App) GetAgentPauseCause(ctx context.Context, domainId, fromUserId, toAgentId int64, allowChange bool) ([]*model.AgentPauseCause, *model.AppError) {
	return a.Store.Agent().PauseCause(ctx, domainId, fromUserId, toAgentId, allowChange)
}

func (a *App) SupervisorAgentItem(ctx context.Context, domainId int64, agentId int64, t *model.FilterBetween) (*model.SupervisorAgentItem, *model.AppError) {
	return a.Store.Agent().SupervisorAgentItem(ctx, domainId, agentId, t)
}
