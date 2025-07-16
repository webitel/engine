package app

import (
	"context"
	"github.com/webitel/engine/model"
	"github.com/webitel/engine/pkg/wbt/auth_manager"
	"google.golang.org/grpc/status"
	"net/http"
)

func (app *App) AgentCheckAccess(ctx context.Context, domainId, id int64, groups []int, access auth_manager.PermissionAccess) (bool, model.AppError) {
	return app.Store.Agent().CheckAccess(ctx, domainId, id, groups, access)
}
func (app *App) AccessAgentsIds(ctx context.Context, domainId int64, agentIds []int64, groups []int, access auth_manager.PermissionAccess) ([]int64, model.AppError) {
	return app.Store.Agent().AccessAgents(ctx, domainId, agentIds, groups, access)
}

func (app *App) CreateAgent(ctx context.Context, agent *model.Agent) (*model.Agent, model.AppError) {
	return app.Store.Agent().Create(ctx, agent)
}

func (app *App) GetAgentsPage(ctx context.Context, domainId int64, search *model.SearchAgent) ([]*model.Agent, bool, model.AppError) {
	list, err := app.Store.Agent().GetAllPage(ctx, domainId, search)
	if err != nil {
		return nil, false, err
	}
	search.RemoveLastElemIfNeed(&list)
	return list, search.EndOfList(), nil
}

func (app *App) GetAgentActiveTasks(ctx context.Context, domainId, agentId int64) ([]*model.CCTask, model.AppError) {
	return app.Store.Agent().GetActiveTask(ctx, domainId, agentId)
}

func (app *App) GetAgentsPageByGroups(ctx context.Context, domainId int64, groups []int, search *model.SearchAgent) ([]*model.Agent, bool, model.AppError) {
	list, err := app.Store.Agent().GetAllPageByGroups(ctx, domainId, groups, search)
	if err != nil {
		return nil, false, err
	}
	search.RemoveLastElemIfNeed(&list)
	return list, search.EndOfList(), nil
}

func (app *App) GetAgentStateHistoryPage(ctx context.Context, domainId int64, search *model.SearchAgentState) ([]*model.AgentState, bool, model.AppError) {
	list, err := app.Store.Agent().HistoryState(ctx, domainId, search)
	if err != nil {
		return nil, false, err
	}
	search.RemoveLastElemIfNeed(&list)
	return list, search.EndOfList(), nil
}

func (app *App) GetAgentById(ctx context.Context, domainId, id int64) (*model.Agent, model.AppError) {
	return app.Store.Agent().Get(ctx, domainId, id)
}

func (app *App) UpdateAgent(ctx context.Context, agent *model.Agent) (*model.Agent, model.AppError) {
	oldAgent, err := app.GetAgentById(ctx, agent.DomainId, agent.Id)
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
	oldAgent.TaskCount = agent.TaskCount
	oldAgent.ScreenControl = agent.ScreenControl

	if err = oldAgent.IsValid(); err != nil {
		return nil, err
	}

	oldAgent, err = app.Store.Agent().Update(ctx, oldAgent)
	if err != nil {
		return nil, err
	}

	return oldAgent, nil
}

func (a *App) PatchAgent(ctx context.Context, domainId, id int64, patch *model.AgentPatch) (*model.Agent, model.AppError) {
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

func (a *App) RemoveAgent(ctx context.Context, domainId, id int64) (*model.Agent, model.AppError) {
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

func (a *App) GetAgentInQueuePage(ctx context.Context, domainId, id int64, search *model.SearchAgentInQueue) ([]*model.AgentInQueue, bool, model.AppError) {
	list, err := a.Store.Agent().InQueue(ctx, domainId, id, search)
	if err != nil {
		return nil, false, err
	}
	search.RemoveLastElemIfNeed(&list)
	return list, search.EndOfList(), nil
}

func (a *App) GetAgentInQueueStatistics(ctx context.Context, domainId, agentId int64) ([]*model.AgentInQueueStatistic, model.AppError) {
	return a.Store.Agent().QueueStatistic(ctx, domainId, agentId)
}

func (a *App) AgentsLookupNotExistsUsers(ctx context.Context, domainId int64, search *model.SearchAgentUser) ([]*model.AgentUser, bool, model.AppError) {
	list, err := a.Store.Agent().LookupNotExistsUsers(ctx, domainId, search)
	if err != nil {
		return nil, false, err
	}
	search.RemoveLastElemIfNeed(&list)
	return list, search.EndOfList(), nil
}

func (a *App) AgentsLookupNotExistsUsersByGroups(ctx context.Context, domainId int64, groups []int, search *model.SearchAgentUser) ([]*model.AgentUser, bool, model.AppError) {
	list, err := a.Store.Agent().LookupNotExistsUsersByGroups(ctx, domainId, groups, search)
	if err != nil {
		return nil, false, err
	}
	search.RemoveLastElemIfNeed(&list)
	return list, search.EndOfList(), nil
}

func (a *App) GetAgentSession(ctx context.Context, domainId, id int64) (*model.AgentSession, model.AppError) {
	return a.Store.Agent().GetSession(ctx, domainId, id)
}

func (a *App) AgentCC(ctx context.Context, domainId int64, userId int64) (*model.AgentCC, model.AppError) {
	return a.Store.Agent().AgentCC(ctx, domainId, userId)
}

func (a *App) LoginAgent(domainId, agentId int64, onDemand bool) model.AppError {
	err := a.cc.Agent().Online(domainId, agentId, onDemand)
	if err != nil {
		return model.NewBadRequestError("app.agent.login.app_err", err.Error())
	}

	return nil
}

func (a *App) LogoutAgent(domainId, agentId int64) model.AppError {
	err := a.cc.Agent().Offline(domainId, agentId)
	if err != nil {
		return model.NewBadRequestError("app.agent.logout.app_err", err.Error())
	}

	return nil
}

func (a *App) WaitingAgentChannel(domainId int64, agentId int64, channel string) (int64, model.AppError) {
	timestamp, err := a.cc.Agent().WaitingChannel(int(agentId), channel)
	if err != nil {
		return 0, model.NewBadRequestError("app.agent.waiting.app_err", err.Error())
	}

	return timestamp, nil
}

func (a *App) PauseAgent(domainId, agentId int64, payload string, timeout int) model.AppError {
	err := a.cc.Agent().Pause(domainId, agentId, payload, timeout)
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			appErr := model.AppErrorFromJson(st.Message())
			if appErr != nil {
				appErr.SetStatusCode(http.StatusBadRequest)
				return appErr
			}
		}

		return model.NewBadRequestError("app.agent.pause.app_err", err.Error())
	}

	return nil
}

func (a *App) GetAgentReportCall(ctx context.Context, domainId int64, search *model.SearchAgentCallStatistics) ([]*model.AgentCallStatistics, bool, model.AppError) {
	list, err := a.Store.Agent().CallStatistics(ctx, domainId, search)
	if err != nil {
		return nil, false, err
	}
	search.RemoveLastElemIfNeed(&list)
	return list, search.EndOfList(), nil
}

func (a *App) GetAgentTodayStatistics(ctx context.Context, domainId, agentId int64) (*model.AgentStatistics, model.AppError) {
	return a.Store.Agent().TodayStatistics(ctx, domainId, &agentId, nil)
}

func (a *App) GetUserTodayStatistics(ctx context.Context, domainId, userId int64) (*model.AgentStatistics, model.AppError) {
	return a.Store.Agent().TodayStatistics(ctx, domainId, nil, &userId)
}

func (a *App) GetAgentStatusStatistic(ctx context.Context, domainId int64, supervisorUserId int64, groups []int, access auth_manager.PermissionAccess, search *model.SearchAgentStatusStatistic) ([]*model.AgentStatusStatistics, bool, model.AppError) {
	list, err := a.Store.Agent().StatusStatistic(ctx, domainId, supervisorUserId, groups, access, search)
	if err != nil {
		return nil, false, err
	}
	search.RemoveLastElemIfNeed(&list)
	return list, search.EndOfList(), nil
}

// agent_id check
func (a *App) AcceptTask(appId string, domainId int64, attemptId int64) model.AppError {
	err := a.cc.Agent().AcceptTask(appId, domainId, attemptId)
	if err != nil {
		return model.NewInternalError("app.cc.accept_task", err.Error())
	}
	return nil
}

func (a *App) CloseTask(appId string, domainId int64, attemptId int64) model.AppError {
	err := a.cc.Agent().CloseTask(appId, domainId, attemptId)
	if err != nil {
		return model.NewInternalError("app.cc.close_task", err.Error())
	}
	return nil
}

func (a *App) GetAgentPauseCause(ctx context.Context, domainId, fromUserId, toAgentId int64, allowChange bool) ([]*model.AgentPauseCause, model.AppError) {
	return a.Store.Agent().PauseCause(ctx, domainId, fromUserId, toAgentId, allowChange)
}

func (a *App) SupervisorAgentItem(ctx context.Context, domainId int64, agentId int64, t *model.FilterBetween) (*model.SupervisorAgentItem, model.AppError) {
	return a.Store.Agent().SupervisorAgentItem(ctx, domainId, agentId, t)
}

func (app *App) GetUsersStatusPage(ctx context.Context, domainId int64, search *model.SearchUserStatus) ([]*model.UserStatus, bool, model.AppError) {
	list, err := app.Store.Agent().UsersStatus(ctx, domainId, search)
	if err != nil {
		return nil, false, err
	}
	search.RemoveLastElemIfNeed(&list)
	return list, search.EndOfList(), nil
}

func (app *App) GetUsersStatusPageByGroups(ctx context.Context, domainId int64, groups []int, search *model.SearchUserStatus) ([]*model.UserStatus, bool, model.AppError) {
	list, err := app.Store.Agent().UsersStatusByGroup(ctx, domainId, groups, search)
	if err != nil {
		return nil, false, err
	}
	search.RemoveLastElemIfNeed(&list)
	return list, search.EndOfList(), nil
}

func (a *App) RunAgentTrigger(ctx context.Context, domainId int64, userId int64, triggerId int32, vars map[string]string) (string, model.AppError) {
	jobId, err := a.cc.Agent().RunTrigger(ctx, domainId, userId, triggerId, vars)
	if err != nil {
		return "", model.NewBadRequestError("app.agent.run_trigger.app_err", err.Error())
	}

	return jobId, nil
}
