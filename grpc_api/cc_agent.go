package grpc_api

import (
	"context"
	"fmt"
	"github.com/webitel/engine/auth_manager"
	"github.com/webitel/engine/grpc_api/engine"
	"github.com/webitel/engine/model"
	"net/http"
)

type agent struct {
	*API
}

func NewAgentApi(api *API) *agent {
	return &agent{api}
}

func (api *agent) CreateAgent(ctx context.Context, in *engine.CreateAgentRequest) (*engine.Agent, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_AGENT)
	if !permission.CanCreate() {
		return nil, api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_CREATE)
	}

	agent := &model.Agent{
		DomainRecord: model.DomainRecord{
			DomainId:  session.Domain(in.GetDomainId()),
			CreatedAt: model.GetMillis(),
			CreatedBy: model.Lookup{
				Id: int(session.UserId),
			},
			UpdatedAt: model.GetMillis(),
			UpdatedBy: model.Lookup{
				Id: int(session.UserId),
			},
		},
		User: model.Lookup{
			Id: int(in.GetUser().GetId()),
		},
		Description:      in.Description,
		ProgressiveCount: int(in.ProgressiveCount),
	}

	err = agent.IsValid()
	if err != nil {
		return nil, err
	}

	agent, err = api.app.CreateAgent(agent)
	if err != nil {
		return nil, err
	}

	return transformAgent(agent), nil
}

func (api *agent) SearchAgent(ctx context.Context, in *engine.SearchAgentRequest) (*engine.ListAgent, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_AGENT)
	if !permission.CanRead() {
		return nil, api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	var list []*model.Agent
	var endList bool
	req := &model.SearchAgent{
		ListRequest: model.ListRequest{
			DomainId: in.GetDomainId(),
			Q:        in.GetQ(),
			Page:     int(in.GetPage()),
			PerPage:  int(in.GetSize()),
			Fields:   in.Fields,
			Sort:     in.Sort,
		},
		Ids: in.Id,
	}

	if permission.Rbac {
		list, endList, err = api.app.GetAgentsPageByGroups(session.Domain(in.DomainId), session.GetAclRoles(), req)
	} else {
		list, endList, err = api.app.GetAgentsPage(session.Domain(in.DomainId), req)
	}

	if err != nil {
		return nil, err
	}

	items := make([]*engine.Agent, 0, len(list))
	for _, v := range list {
		items = append(items, transformAgent(v))
	}

	return &engine.ListAgent{
		Next:  !endList,
		Items: items,
	}, nil
}

func (api *agent) ReadAgent(ctx context.Context, in *engine.ReadAgentRequest) (*engine.Agent, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_AGENT)
	if !permission.CanRead() {
		return nil, api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	var agent *model.Agent

	if permission.Rbac {
		var perm bool
		if perm, err = api.app.AgentCheckAccess(session.Domain(in.GetDomainId()), in.GetId(), session.GetAclRoles(), auth_manager.PERMISSION_ACCESS_READ); err != nil {
			return nil, err
		} else if !perm {
			return nil, api.app.MakeResourcePermissionError(session, in.GetId(), permission, auth_manager.PERMISSION_ACCESS_READ)
		}
	}

	agent, err = api.app.GetAgentById(session.Domain(in.DomainId), in.Id)

	if err != nil {
		return nil, err
	}

	return transformAgent(agent), nil
}

func (api *agent) UpdateAgent(ctx context.Context, in *engine.UpdateAgentRequest) (*engine.Agent, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_AGENT)
	if !permission.CanRead() {
		return nil, api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if !permission.CanUpdate() {
		return nil, api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
	}

	if permission.Rbac {
		var perm bool
		if perm, err = api.app.AgentCheckAccess(session.Domain(in.GetDomainId()), in.GetId(), session.GetAclRoles(), auth_manager.PERMISSION_ACCESS_UPDATE); err != nil {
			return nil, err
		} else if !perm {
			return nil, api.app.MakeResourcePermissionError(session, in.GetId(), permission, auth_manager.PERMISSION_ACCESS_UPDATE)
		}
	}

	var agent *model.Agent

	agent, err = api.app.UpdateAgent(&model.Agent{
		DomainRecord: model.DomainRecord{
			Id:        in.Id,
			DomainId:  session.Domain(in.GetDomainId()),
			UpdatedAt: model.GetMillis(),
			UpdatedBy: model.Lookup{
				Id: int(session.UserId),
			},
		},
		User: model.Lookup{
			Id: int(in.GetUser().GetId()),
		},
		Description:      in.Description,
		ProgressiveCount: int(in.ProgressiveCount),
	})

	if err != nil {
		return nil, err
	}

	return transformAgent(agent), nil
}

func (api *agent) DeleteAgent(ctx context.Context, in *engine.DeleteAgentRequest) (*engine.Agent, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_AGENT)
	if !permission.CanDelete() {
		return nil, api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_DELETE)
	}

	if permission.Rbac {
		var perm bool
		if perm, err = api.app.AgentCheckAccess(session.Domain(in.GetDomainId()), in.GetId(), session.GetAclRoles(), auth_manager.PERMISSION_ACCESS_DELETE); err != nil {
			return nil, err
		} else if !perm {
			return nil, api.app.MakeResourcePermissionError(session, in.GetId(), permission, auth_manager.PERMISSION_ACCESS_DELETE)
		}
	}

	var agent *model.Agent
	agent, err = api.app.RemoveAgent(session.Domain(in.DomainId), in.Id)
	if err != nil {
		return nil, err
	}

	return transformAgent(agent), nil
}

func (api *agent) UpdateAgentStatus(ctx context.Context, in *engine.AgentStatusRequest) (*engine.Response, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	if session.UserId != in.Id {
		permission := session.GetPermission(model.PERMISSION_SCOPE_CC_AGENT)
		if !permission.CanUpdate() {
			return nil, api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
		}

		if permission.Rbac {
			var perm bool
			if perm, err = api.app.AgentCheckAccess(session.Domain(in.GetDomainId()), in.GetId(), session.GetAclRoles(), auth_manager.PERMISSION_ACCESS_UPDATE); err != nil {
				return nil, err
			} else if !perm {
				return nil, api.app.MakeResourcePermissionError(session, in.GetId(), permission, auth_manager.PERMISSION_ACCESS_UPDATE)
			}
		}
	}

	switch in.Status {
	case model.AgentStatusOnline:
		err = api.ctrl.LoginAgent(session, session.Domain(in.GetDomainId()), in.GetId(), in.GetChannels(), in.OnDemand)
	case model.AgentStatusPause:
		err = api.ctrl.PauseAgent(session, session.Domain(in.GetDomainId()), in.GetId(), in.GetPayload(), 0)
	case model.AgentStatusOffline:
		err = api.ctrl.LogoutAgent(session, session.Domain(in.GetDomainId()), in.GetId())
	default:
		err = model.NewAppError("GRPC.UpdateAgentStatus", "grpc.agent.update_status", nil, fmt.Sprintf("not found status %s", in.Status),
			http.StatusBadRequest)
	}

	if err != nil {
		return nil, err
	}

	return ResponseOk, nil
}

func (api *agent) SearchAgentInTeam(ctx context.Context, in *engine.SearchAgentInTeamRequest) (*engine.ListAgentInTeam, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_AGENT)
	if !permission.CanRead() {
		return nil, api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if permission.Rbac {
		var perm bool
		if perm, err = api.app.AgentCheckAccess(session.Domain(in.GetDomainId()), in.GetId(), session.GetAclRoles(), auth_manager.PERMISSION_ACCESS_READ); err != nil {
			return nil, err
		} else if !perm {
			return nil, api.app.MakeResourcePermissionError(session, in.GetId(), permission, auth_manager.PERMISSION_ACCESS_READ)
		}
	}

	var list []*model.AgentInTeam
	var endList bool
	req := &model.SearchAgentInTeam{
		ListRequest: model.ListRequest{
			DomainId: in.GetDomainId(),
			Q:        in.GetQ(),
			Page:     int(in.GetPage()),
			PerPage:  int(in.GetSize()),
		},
	}

	list, endList, err = api.app.GetAgentInTeamPage(session.Domain(in.GetDomainId()), in.GetId(), req)
	if err != nil {
		return nil, err
	}

	items := make([]*engine.AgentInTeam, 0, len(list))

	for _, v := range list {
		items = append(items, &engine.AgentInTeam{
			Team:     GetProtoLookup(&v.Team),
			Strategy: v.Strategy,
		})
	}

	return &engine.ListAgentInTeam{
		Next:  !endList,
		Items: items,
	}, nil
}

func (api *agent) SearchAgentInQueue(ctx context.Context, in *engine.SearchAgentInQueueRequest) (*engine.ListAgentInQueue, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_AGENT)
	if !permission.CanRead() {
		return nil, api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if permission.Rbac {
		var perm bool
		if perm, err = api.app.AgentCheckAccess(session.Domain(in.GetDomainId()), in.GetId(), session.GetAclRoles(), auth_manager.PERMISSION_ACCESS_READ); err != nil {
			return nil, err
		} else if !perm {
			return nil, api.app.MakeResourcePermissionError(session, in.GetId(), permission, auth_manager.PERMISSION_ACCESS_READ)
		}
	}

	var list []*model.AgentInQueue
	var endList bool
	req := &model.SearchAgentInQueue{
		ListRequest: model.ListRequest{
			DomainId: in.GetDomainId(),
			Q:        in.GetQ(),
			Page:     int(in.GetPage()),
			PerPage:  int(in.GetSize()),
		},
	}
	list, endList, err = api.app.GetAgentInQueuePage(session.Domain(in.GetDomainId()), in.GetId(), req)
	if err != nil {
		return nil, err
	}

	items := make([]*engine.AgentInQueue, 0, len(list))

	for _, v := range list {
		items = append(items, &engine.AgentInQueue{
			Queue:         GetProtoLookup(&v.Queue),
			Priority:      int32(v.Priority),
			Type:          int32(v.Type),
			Strategy:      v.Strategy,
			Enabled:       v.Enabled,
			CountMember:   int32(v.CountMembers),
			WaitingMember: int32(v.WaitingMembers),
			ActiveMember:  int32(v.ActiveMembers),
		})
	}

	return &engine.ListAgentInQueue{
		Next:  !endList,
		Items: items,
	}, nil
}

func (api *agent) AgentStateHistory(ctx context.Context, in *engine.AgentStateHistoryRequest) (*engine.ListAgentStateHistory, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_AGENT)
	if !permission.CanRead() {
		return nil, api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if permission.Rbac {
		var perm bool
		if perm, err = api.app.AgentCheckAccess(session.Domain(in.GetDomainId()), in.GetAgentId(), session.GetAclRoles(), auth_manager.PERMISSION_ACCESS_READ); err != nil {
			return nil, err
		} else if !perm {
			return nil, api.app.MakeResourcePermissionError(session, in.GetAgentId(), permission, auth_manager.PERMISSION_ACCESS_READ)
		}
	}

	var list []*model.AgentState
	var endList bool
	req := &model.SearchAgentState{
		ListRequest: model.ListRequest{
			DomainId: in.GetDomainId(),
			Q:        in.GetQ(),
			Page:     int(in.GetPage()),
			PerPage:  int(in.GetSize()),
		},
		JoinedAt: model.FilterBetween{
			From: in.GetTimeFrom(),
			To:   in.GetTimeTo(),
		},
		AgentIds: []int64{in.AgentId},
	}
	list, endList, err = api.app.GetAgentStateHistoryPage(session.Domain(in.GetDomainId()), req)
	if err != nil {
		return nil, err
	}

	items := make([]*engine.AgentState, 0, len(list))

	for _, v := range list {
		items = append(items, toEngineAgentState(v))
	}

	return &engine.ListAgentStateHistory{
		Next:  !endList,
		Items: items,
	}, nil
}

//FIXME RBAC
func (api *agent) SearchAgentStateHistory(ctx context.Context, in *engine.SearchAgentStateHistoryRequest) (*engine.ListAgentStateHistory, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_AGENT)
	if !permission.CanRead() {
		return nil, api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	var list []*model.AgentState
	var endList bool
	req := &model.SearchAgentState{
		ListRequest: model.ListRequest{
			DomainId: in.GetDomainId(),
			Page:     int(in.GetPage()),
			PerPage:  int(in.GetSize()),
			Sort:     in.Sort,
		},
		JoinedAt: model.FilterBetween{
			From: in.GetJoinedAt().GetFrom(),
			To:   in.GetJoinedAt().GetTo(),
		},
		AgentIds: in.AgentId,
	}
	list, endList, err = api.app.GetAgentStateHistoryPage(session.Domain(in.GetDomainId()), req)
	if err != nil {
		return nil, err
	}

	items := make([]*engine.AgentState, 0, len(list))

	for _, v := range list {
		items = append(items, toEngineAgentState(v))
	}

	return &engine.ListAgentStateHistory{
		Next:  !endList,
		Items: items,
	}, nil
}

func (api *agent) SearchLookupUsersAgentNotExists(ctx context.Context, in *engine.SearchLookupUsersAgentNotExistsRequest) (*engine.ListAgentUser, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	permission := session.GetPermission(model.PERMISSION_SCOPE_USERS)
	if !permission.CanRead() {
		return nil, api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	var list []*model.AgentUser
	var endList bool
	req := &model.SearchAgentUser{
		ListRequest: model.ListRequest{
			//DomainId: in.GetDomainId(),
			Q:       in.GetQ(),
			Page:    int(in.GetPage()),
			PerPage: int(in.GetSize()),
		},
	}

	items := make([]*engine.AgentUser, 0, len(list))

	if permission.Rbac {
		list, endList, err = api.app.AgentsLookupNotExistsUsersByGroups(session.Domain(in.GetDomainId()), session.GetAclRoles(), req)
	} else {
		list, endList, err = api.app.AgentsLookupNotExistsUsers(session.Domain(in.GetDomainId()), req)
	}

	if err != nil {
		return nil, err
	}

	for _, v := range list {
		items = append(items, &engine.AgentUser{
			Id:   v.Id,
			Name: v.Name,
		})
	}

	return &engine.ListAgentUser{
		Next:  !endList,
		Items: items,
	}, nil
}

func (api *agent) SearchAgentInQueueStatistics(ctx context.Context, in *engine.SearchAgentInQueueStatisticsRequest) (*engine.AgentInQueueStatisticsList, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	var list []*model.AgentInQueueStatistic

	list, err = api.ctrl.GetAgentInQueueStatistics(session, in.GetDomainId(), in.GetAgentId())

	if err != nil {
		return nil, err
	}

	res := make([]*engine.AgentInQueueStatistics, 0, len(list))

	for _, v := range list {
		res = append(res, &engine.AgentInQueueStatistics{
			Queue:      GetProtoLookup(&v.Queue),
			Statistics: toAgentStats(v.Statistics),
		})
	}

	return &engine.AgentInQueueStatisticsList{
		Items: res,
	}, nil
}

// FIXME RBAC
func (api *agent) SearchAgentCallStatistics(ctx context.Context, in *engine.SearchAgentCallStatisticsRequest) (*engine.AgentCallStatisticsList, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_AGENT)
	if !permission.CanRead() {
		return nil, api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if in.GetTime() == nil {
		return nil, model.NewAppError("GRPC.SearchAgentCallStatistics", "grpc.agent.report.call", nil, "filter time is required", http.StatusBadRequest)
	}

	var list []*model.AgentCallStatistics
	var endList bool
	req := &model.SearchAgentCallStatistics{
		ListRequest: model.ListRequest{
			DomainId: session.Domain(in.DomainId),
			Page:     int(in.GetPage()),
			PerPage:  int(in.GetSize()),
			Q:        in.GetQ(),
			Fields:   in.Fields,
			Sort:     in.Sort,
		},
		Time: model.FilterBetween{
			From: in.GetTime().GetFrom(),
			To:   in.GetTime().GetTo(),
		},
		AgentIds: in.AgentId,
	}

	list, endList, err = api.app.GetAgentReportCall(session.Domain(in.DomainId), req)
	if err != nil {
		return nil, err
	}

	items := make([]*engine.AgentCallStatistics, 0, len(list))
	for _, v := range list {
		items = append(items, toEngineAgentCallStatistics(v))
	}
	return &engine.AgentCallStatisticsList{
		Next:  !endList,
		Items: items,
	}, nil
}

// FIXME RBAC
func (api *agent) SearchAgentStatusStatistic(ctx context.Context, in *engine.SearchAgentStatusStatisticRequest) (*engine.ListAgentStatsStatistic, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_AGENT)
	if !permission.CanRead() {
		return nil, api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if in.GetTime() == nil {
		return nil, model.NewAppError("GRPC.SearchAgentCallStatistics", "grpc.agent.report.call", nil, "filter time is required", http.StatusBadRequest)
	}

	var list []*model.AgentStatusStatistics
	var endList bool
	req := &model.SearchAgentStatusStatistic{
		ListRequest: model.ListRequest{
			DomainId: session.Domain(in.DomainId),
			Page:     int(in.GetPage()),
			PerPage:  int(in.GetSize()),
			Sort:     in.Sort,
		},
		Time: model.FilterBetween{
			From: in.GetTime().GetFrom(),
			To:   in.GetTime().GetTo(),
		},
		AgentIds: in.AgentId,
		Status:   in.Status,
	}

	if in.Utilization != nil {
		req.Utilization = &model.FilterBetween{
			From: in.GetUtilization().GetFrom(),
			To:   in.GetUtilization().GetTo(),
		}
	}

	list, endList, err = api.app.GetAgentStatusStatistic(session.Domain(in.DomainId), req)
	if err != nil {
		return nil, err
	}

	items := make([]*engine.AgentStatsStatistic, 0, len(list))
	for _, v := range list {
		items = append(items, toEngineAgentStatusStatistics(v))
	}
	return &engine.ListAgentStatsStatistic{
		Next:  !endList,
		Items: items,
	}, nil
}

func toEngineAgentCallStatistics(src *model.AgentCallStatistics) *engine.AgentCallStatistics {
	return &engine.AgentCallStatistics{
		Name:       src.Name,
		Count:      src.Count,
		Abandoned:  src.Abandoned,
		Handles:    src.Handles,
		SumTalkSec: src.SumTalkSec,
		AvgTalkSec: src.AvgTalkSec,
		MinTalkSec: src.MinTalkSec,
		MaxTalkSec: src.MaxTalkSec,
		SumHoldSec: src.SumHoldSec,
		AvgHoldSec: src.AvgHoldSec,
		MinHoldSec: src.MinHoldSec,
		MaxHoldSec: src.MaxHoldSec,
	}
}

func toEngineAgentStatusStatistics(src *model.AgentStatusStatistics) *engine.AgentStatsStatistic {
	item := &engine.AgentStatsStatistic{
		AgentId:        src.AgentId,
		Name:           src.Name,
		Status:         src.Status,
		StatusDuration: src.StatusDuration,
		User:           GetProtoLookup(&src.User),
		Extension:      src.Extension,
		Teams:          GetProtoLookups(src.Teams),
		Queues:         GetProtoLookups(src.Queues),
		Online:         src.Online,
		Offline:        src.Offline,
		Pause:          src.Pause,
		Utilization:    src.Utilization,
		CallTime:       src.CallTime,
		Handles:        src.Handles,
		Missed:         src.Missed,
		MaxBridgedAt:   model.TimeToInt64(src.MaxBridgedAt),
		MaxOfferingAt:  model.TimeToInt64(src.MaxOfferingAt),
	}

	if src.ActiveCallId != nil {
		item.ActiveCallId = *src.ActiveCallId
	}

	return item
}

func transformAgent(src *model.Agent) *engine.Agent {
	agent := &engine.Agent{
		Id: src.Id,
		User: &engine.Lookup{
			Id:   int64(src.User.Id),
			Name: src.User.Name,
		},
		LastStatusChange: src.LastStatusChange,
		Status:           src.Status,
		Description:      src.Description,
		ProgressiveCount: int32(src.ProgressiveCount),
		Name:             src.Name,
		StatusDuration:   src.StatusDuration,
	}

	if src.Channels != nil {
		agent.Channels = make([]*engine.AgentChannel, 0, len(src.Channels))
		for _, v := range src.Channels {
			agent.Channels = append(agent.Channels, &engine.AgentChannel{
				Channel:  v.Channel,
				State:    v.State,
				JoinedAt: v.JoinedAt,
				Online:   v.Active,
			})
		}
	}

	return agent
}

func toEngineAgentState(src *model.AgentState) *engine.AgentState {
	st := &engine.AgentState{
		Id:       src.Id,
		JoinedAt: model.TimeToInt64(src.JoinedAt),
		State:    src.State,
		Duration: src.Duration,
	}

	if src.Channel != nil {
		st.Channel = *src.Channel
	}

	if src.Agent != nil {
		st.Agent = &engine.Lookup{
			Id:   int64(src.Agent.Id),
			Name: src.Agent.Name,
		}
	}

	if src.Payload != nil {
		st.Payload = *src.Payload
	}

	return st
}

func toAgentStats(src []*model.AgentInQueueStats) []*engine.AgentInQueueStatistics_AgentInQueueStatisticsItem {
	res := make([]*engine.AgentInQueueStatistics_AgentInQueueStatisticsItem, 0, len(src))

	for _, v := range src {
		res = append(res, &engine.AgentInQueueStatistics_AgentInQueueStatisticsItem{
			Bucket:        GetProtoLookup(v.Bucket),
			Skill:         GetProtoLookup(v.Skill),
			MemberWaiting: int32(v.MemberWaiting),
		})
	}
	return res
}
