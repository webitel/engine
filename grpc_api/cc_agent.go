package grpc_api

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/golang/protobuf/ptypes/wrappers"

	"github.com/webitel/engine/gen/engine"
	"github.com/webitel/engine/model"
	"github.com/webitel/engine/pkg/wbt/auth_manager"
)

type agent struct {
	*API
	engine.UnsafeAgentServiceServer
}

func NewAgentApi(api *API) *agent {
	return &agent{API: api}
}

func (api *agent) SearchAgentInTeam(context.Context, *engine.SearchAgentInTeamRequest) (*engine.ListAgentInTeam, error) {
	return nil, model.NewInternalError("deprecated", "deprecated")
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
			CreatedBy: &model.Lookup{
				Id: int(session.UserId),
			},
			UpdatedAt: model.GetMillis(),
			UpdatedBy: &model.Lookup{
				Id: int(session.UserId),
			},
		},
		User: model.Lookup{
			Id: int(in.GetUser().GetId()),
		},
		Description:      in.Description,
		ProgressiveCount: int(in.ProgressiveCount),
		GreetingMedia:    GetLookup(in.GreetingMedia),
		AllowChannels:    in.AllowChannels,
		ChatCount:        in.ChatCount,
		Supervisor:       GetLookups(in.Supervisor),
		Team:             GetLookup(in.Team),
		Region:           GetLookup(in.Region),
		Auditor:          GetLookups(in.Auditor),
		IsSupervisor:     in.GetIsSupervisor(),
		TaskCount:        in.TaskCount,
		ScreenControl:    in.ScreenControl,
	}

	err = agent.IsValid()
	if err != nil {
		return nil, err
	}

	agent, err = api.app.CreateAgent(ctx, agent)
	if err != nil {
		return nil, err
	}

	res := transformAgent(agent)
	api.app.AuditCreate(ctx, session, model.PERMISSION_SCOPE_CC_AGENT, strconv.FormatInt(res.Id, 10), res)

	return res, nil
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
			Q:       in.GetQ(),
			Page:    int(in.GetPage()),
			PerPage: int(in.GetSize()),
			Fields:  in.Fields,
			Sort:    in.Sort,
		},
		Ids:           in.Id,
		AllowChannels: in.GetAllowChannels(),
		SupervisorIds: in.GetSupervisorId(),
		TeamIds:       in.GetTeamId(),
		RegionIds:     in.GetRegionId(),
		AuditorIds:    in.GetAuditorId(),
		SkillIds:      in.GetSkillId(),
		QueueIds:      in.GetQueueId(),
		Extensions:    in.GetExtension(),
		UserIds:       in.GetUserId(),
		NotTeamIds:    in.GetNotTeamId(),
		NotSkillIds:   in.GetNotSkillId(),
	}

	if in.IsSupervisor {
		req.IsSupervisor = &in.IsSupervisor
	}

	if in.NotSupervisor {
		req.NotSupervisor = &in.NotSupervisor
	}

	if session.UseRBAC(auth_manager.PERMISSION_ACCESS_READ, permission) {
		list, endList, err = api.app.GetAgentsPageByGroups(ctx, session.Domain(0), session.GetAclRoles(), req)
	} else {
		list, endList, err = api.app.GetAgentsPage(ctx, session.Domain(0), req)
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

	if session.UseRBAC(auth_manager.PERMISSION_ACCESS_READ, permission) {
		var perm bool
		if perm, err = api.app.AgentCheckAccess(ctx, session.Domain(in.GetDomainId()), in.GetId(), session.GetAclRoles(),
			auth_manager.PERMISSION_ACCESS_READ); err != nil {
			return nil, err
		} else if !perm {
			return nil, api.app.MakeResourcePermissionError(session, in.GetId(), permission, auth_manager.PERMISSION_ACCESS_READ)
		}
	}

	agent, err = api.app.GetAgentById(ctx, session.Domain(in.DomainId), in.Id)

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

	if session.UseRBAC(auth_manager.PERMISSION_ACCESS_UPDATE, permission) {
		var perm bool
		if perm, err = api.app.AgentCheckAccess(ctx, session.Domain(in.GetDomainId()), in.GetId(), session.GetAclRoles(),
			auth_manager.PERMISSION_ACCESS_UPDATE); err != nil {
			return nil, err
		} else if !perm {
			return nil, api.app.MakeResourcePermissionError(session, in.GetId(), permission, auth_manager.PERMISSION_ACCESS_UPDATE)
		}
	}

	var agent *model.Agent

	agent, err = api.app.UpdateAgent(ctx, &model.Agent{
		DomainRecord: model.DomainRecord{
			Id:        in.Id,
			DomainId:  session.Domain(in.GetDomainId()),
			UpdatedAt: model.GetMillis(),
			UpdatedBy: &model.Lookup{
				Id: int(session.UserId),
			},
		},
		User: model.Lookup{
			Id: int(in.GetUser().GetId()),
		},
		Description:      in.Description,
		ProgressiveCount: int(in.ProgressiveCount),
		GreetingMedia:    GetLookup(in.GreetingMedia),
		AllowChannels:    in.AllowChannels,
		ChatCount:        in.ChatCount,
		Supervisor:       GetLookups(in.Supervisor),
		Team:             GetLookup(in.Team),
		Region:           GetLookup(in.Region),
		Auditor:          GetLookups(in.Auditor),
		IsSupervisor:     in.GetIsSupervisor(),
		TaskCount:        in.TaskCount,
		ScreenControl:    in.ScreenControl,
	})

	if err != nil {
		return nil, err
	}

	res := transformAgent(agent)
	api.app.AuditUpdate(ctx, session, model.PERMISSION_SCOPE_CC_AGENT, strconv.FormatInt(res.Id, 10), res)

	return res, nil
}

func (api *agent) PatchAgent(ctx context.Context, in *engine.PatchAgentRequest) (*engine.Agent, error) {
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

	if session.UseRBAC(auth_manager.PERMISSION_ACCESS_UPDATE, permission) {
		var perm bool
		if perm, err = api.app.AgentCheckAccess(ctx, session.Domain(0), in.GetId(), session.GetAclRoles(),
			auth_manager.PERMISSION_ACCESS_UPDATE); err != nil {
			return nil, err
		} else if !perm {
			return nil, api.app.MakeResourcePermissionError(session, in.GetId(), permission, auth_manager.PERMISSION_ACCESS_UPDATE)
		}
	}

	var agent *model.Agent
	patch := model.AgentPatch{
		UpdatedBy: model.Lookup{
			Id: int(session.UserId),
		},
		UpdatedAt: model.GetMillis(),
	}

	for _, v := range in.Fields {
		switch v {
		case "user.id", "user":
			patch.User = &model.Lookup{
				Id: int(in.GetUser().GetId()),
			}
		case "description":
			patch.Description = model.NewString(in.Description)
		case "progressive_count":
			patch.ProgressiveCount = model.NewInt(int(in.ProgressiveCount))
		case "greeting_media.id", "greeting_media":
			patch.GreetingMedia = &model.Lookup{
				Id: int(in.GetGreetingMedia().GetId()),
			}
		case "chat_count":
			patch.ChatCount = &in.ChatCount
		case "supervisor.id", "supervisor":
			patch.Supervisor = GetLookups(in.GetSupervisor())
			if patch.Supervisor == nil {
				patch.Supervisor = make([]*model.Lookup, 0, 0)
			}

		case "team.id", "team":
			patch.Team = &model.Lookup{
				Id: int(in.GetTeam().GetId()),
			}
		case "region.id", "region":
			patch.Region = &model.Lookup{
				Id: int(in.GetRegion().GetId()),
			}
		case "auditor":
			patch.Auditor = GetLookups(in.Auditor)
		case "is_supervisor":
			patch.IsSupervisor = &in.IsSupervisor
		case "screen_control":
			patch.ScreenControl = &in.ScreenControl
		}
	}

	agent, err = api.app.PatchAgent(ctx, session.Domain(0), in.GetId(), &patch)

	if err != nil {
		return nil, err
	}

	res := transformAgent(agent)
	api.app.AuditUpdate(ctx, session, model.PERMISSION_SCOPE_CC_AGENT, strconv.FormatInt(res.Id, 10), res)

	return res, nil
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

	if session.UseRBAC(auth_manager.PERMISSION_ACCESS_DELETE, permission) {
		var perm bool
		if perm, err = api.app.AgentCheckAccess(ctx, session.Domain(in.GetDomainId()), in.GetId(), session.GetAclRoles(),
			auth_manager.PERMISSION_ACCESS_DELETE); err != nil {
			return nil, err
		} else if !perm {
			return nil, api.app.MakeResourcePermissionError(session, in.GetId(), permission, auth_manager.PERMISSION_ACCESS_DELETE)
		}
	}

	var agent *model.Agent
	agent, err = api.app.RemoveAgent(ctx, session.Domain(in.DomainId), in.Id)
	if err != nil {
		return nil, err
	}

	res := transformAgent(agent)
	api.app.AuditDelete(ctx, session, model.PERMISSION_SCOPE_CC_AGENT, strconv.FormatInt(res.Id, 10), res)

	return res, nil
}

func (api *agent) UpdateAgentStatus(ctx context.Context, in *engine.AgentStatusRequest) (*engine.Response, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	var cc *model.AgentCC
	var agentId int64 = 0
	cc, err = api.app.AgentCC(ctx, session.Domain(0), session.UserId)
	if err != nil {
		return nil, err
	}
	if cc.AgentId != nil {
		agentId = *cc.AgentId
	}

	if agentId != in.Id {
		permission := session.GetPermission(model.PERMISSION_SCOPE_CC_AGENT)
		if !permission.CanUpdate() {
			return nil, api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
		}

		if session.UseRBAC(auth_manager.PERMISSION_ACCESS_UPDATE, permission) {
			var perm bool
			if perm, err = api.app.AgentCheckAccess(ctx, session.Domain(in.GetDomainId()), in.GetId(), session.GetAclRoles(),
				auth_manager.PERMISSION_ACCESS_UPDATE); err != nil {
				return nil, err
			} else if !perm {
				return nil, api.app.MakeResourcePermissionError(session, in.GetId(), permission, auth_manager.PERMISSION_ACCESS_UPDATE)
			}
		}
	}

	switch in.Status {
	case model.AgentStatusOnline:
		err = api.ctrl.LoginAgent(ctx, session, session.Domain(in.GetDomainId()), in.GetId(), in.OnDemand)
	case model.AgentStatusPause:
		err = api.ctrl.PauseAgent(ctx, session, session.Domain(in.GetDomainId()), in.GetId(), in.GetPayload(), in.GetStatusComment(), 0)
	case model.AgentStatusOffline:
		err = api.ctrl.LogoutAgent(ctx, session, session.Domain(in.GetDomainId()), in.GetId())
	default:
		err = model.NewBadRequestError("grpc.agent.update_status", fmt.Sprintf("not found status %s", in.Status))
	}

	if err != nil {
		return nil, err
	}

	return ResponseOk, nil
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

	if session.UseRBAC(auth_manager.PERMISSION_ACCESS_READ, permission) {
		var perm bool
		if perm, err = api.app.AgentCheckAccess(ctx, session.Domain(0), in.GetId(), session.GetAclRoles(),
			auth_manager.PERMISSION_ACCESS_READ); err != nil {
			return nil, err
		} else if !perm {
			return nil, api.app.MakeResourcePermissionError(session, in.GetId(), permission, auth_manager.PERMISSION_ACCESS_READ)
		}
	}

	var list []*model.AgentInQueue
	var endList bool
	req := &model.SearchAgentInQueue{
		ListRequest: model.ListRequest{
			Q:       in.GetQ(),
			Page:    int(in.GetPage()),
			PerPage: int(in.GetSize()),
			Fields:  in.Fields,
			Sort:    in.Sort,
		},
	}
	list, endList, err = api.app.GetAgentInQueuePage(ctx, session.Domain(0), in.GetId(), req)
	if err != nil {
		return nil, err
	}

	items := make([]*engine.AgentInQueue, 0, len(list))

	for _, v := range list {
		el := &engine.AgentInQueue{
			Queue:          GetProtoLookup(&v.Queue),
			Priority:       int32(v.Priority),
			Type:           int32(v.Type),
			Strategy:       v.Strategy,
			Enabled:        v.Enabled,
			CountMembers:   int32(v.CountMembers),
			WaitingMembers: int32(v.WaitingMembers),
			ActiveMembers:  int32(v.ActiveMembers),
			MaxMemberLimit: int32(v.MaxMemberLimit),
			Agents: &engine.AgentInQueue_AgentsInQueue{
				Online:  v.Agents.Online,
				Pause:   v.Agents.Pause,
				Offline: v.Agents.Offline,
				Free:    v.Agents.Free,
				Total:   v.Agents.Total,
			},
		}

		if v.Agents.AllowPause != nil {
			el.Agents.AllowPause = &wrappers.Int32Value{
				Value: *v.Agents.AllowPause,
			}
		}

		items = append(items, el)
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

	if session.UseRBAC(auth_manager.PERMISSION_ACCESS_READ, permission) {
		var perm bool
		if perm, err = api.app.AgentCheckAccess(ctx, session.Domain(in.GetDomainId()), in.GetAgentId(), session.GetAclRoles(),
			auth_manager.PERMISSION_ACCESS_READ); err != nil {
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
	list, endList, err = api.app.GetAgentStateHistoryPage(ctx, session.Domain(in.GetDomainId()), req)
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

func (api *agent) AgentSetState(ctx context.Context, in *engine.AgentSetStateRequest) (*engine.AgentSetStateResponse, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	_, err = api.ctrl.WaitingAgent(ctx, session, 0, int64(in.AgentId), "")
	if err != nil {
		return nil, err
	}

	return &engine.AgentSetStateResponse{}, nil
}

// FIXME RBAC
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

	if in.GetFromId() > 0 {
		req.FromId = &in.FromId
	}

	list, endList, err = api.app.GetAgentStateHistoryPage(ctx, session.Domain(in.GetDomainId()), req)
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

	if session.UseRBAC(auth_manager.PERMISSION_ACCESS_READ, permission) {
		list, endList, err = api.app.AgentsLookupNotExistsUsersByGroups(ctx, session.Domain(in.GetDomainId()), session.GetAclRoles(), req)
	} else {
		list, endList, err = api.app.AgentsLookupNotExistsUsers(ctx, session.Domain(in.GetDomainId()), req)
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

	list, err = api.ctrl.GetAgentInQueueStatistics(ctx, session, in.GetDomainId(), in.GetAgentId())

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
		return nil, model.NewBadRequestError("grpc.agent.report.call", "filter time is required")
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

	list, endList, err = api.app.GetAgentReportCall(ctx, session.Domain(in.DomainId), req)
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

func (api *agent) AgentTodayStatistics(ctx context.Context, in *engine.AgentTodayStatisticsRequest) (*engine.AgentTodayStatisticsResponse, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	var stat *model.AgentStatistics
	stat, err = api.ctrl.GetAgentTodayStatistics(ctx, session, in.AgentId)
	if err != nil {
		return nil, err
	}

	return &engine.AgentTodayStatisticsResponse{
		Utilization:      stat.Utilization,
		Occupancy:        stat.Occupancy,
		CallAbandoned:    stat.CallAbandoned,
		CallHandled:      stat.CallHandled,
		AvgTalkSec:       stat.AvgTalkSec,
		AvgHoldSec:       stat.AvgHoldSec,
		ChatAccepts:      stat.ChatAccepts,
		ChatAht:          stat.ChatAht,
		CallMissed:       stat.CallMissed,
		CallInbound:      stat.CallInbound,
		ScoreRequiredAvg: stat.ScoreRequiredAvg,
		ScoreOptionalAvg: stat.ScoreOptionalAvg,
		ScoreCount:       stat.ScoreCount,
		ScoreRequiredSum: stat.ScoreRequiredSum,
		ScoreOptionalSum: stat.ScoreOptionalSum,
		SumTalkSec:       stat.SumTalkSec,
		VoiceMail:        stat.VoiceMail,
		Available:        stat.Available,
		Online:           stat.Online,
		Processing:       stat.Processing,
		TaskAccepts:      stat.TaskAccepts,
		QueueTalkSec:     stat.QueueTalkSec,
		CallQueueMissed:  stat.CallQueueMissed,
		CallInboundQueue: stat.CallInboundQueue,
		CallDialerQueue:  stat.CallDialerQueue,
		CallManual:       stat.CallManual,
	}, nil
}

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
		return nil, model.NewBadRequestError("grpc.agent.report.call", "filter time is required")
	}

	var list []*model.AgentStatusStatistics
	var endList bool
	req := &model.SearchAgentStatusStatistic{
		ListRequest: model.ListRequest{
			DomainId: session.Domain(0),
			Q:        in.GetQ(),
			Page:     int(in.GetPage()),
			PerPage:  int(in.GetSize()),
			Sort:     in.Sort,
			Fields:   in.Fields,
		},
		Time: model.FilterBetween{
			From: in.GetTime().GetFrom(),
			To:   in.GetTime().GetTo(),
		},
		Utilization:   nil,
		AgentIds:      in.AgentId,
		Status:        in.Status,
		TeamIds:       in.TeamId,
		QueueIds:      in.QueueId,
		SkillIds:      in.SkillId,
		RegionIds:     in.RegionId,
		SupervisorIds: in.SupervisorId,
		AuditorIds:    in.AuditorId,
		HasCall:       in.HasCall,
	}

	if in.Utilization != nil {
		req.Utilization = &model.FilterBetween{
			From: in.GetUtilization().GetFrom(),
			To:   in.GetUtilization().GetTo(),
		}
	}

	list, endList, err = api.app.GetAgentStatusStatistic(ctx, session.Domain(0), session.UserId, session.RoleIds, auth_manager.PERMISSION_ACCESS_READ, req)
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

func (api *agent) SearchPauseCauseForAgent(ctx context.Context, in *engine.SearchPauseCauseForAgentRequest) (*engine.ForAgentPauseCauseList, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	var list []*model.AgentPauseCause
	list, err = api.ctrl.GetAgentPauseCause(ctx, session, in.AgentId, in.AllowChange)
	if err != nil {
		return nil, err
	}

	items := make([]*engine.ForAgentPauseCause, 0, len(list))
	for _, v := range list {
		items = append(items, &engine.ForAgentPauseCause{
			Id:          v.Id,
			Name:        v.Name,
			LimitMin:    v.LimitMin,
			DurationMin: v.DurationMin,
		})
	}

	return &engine.ForAgentPauseCauseList{
		Items: items,
	}, nil
}

func (api *agent) SearchAgentStatusStatisticItem(ctx context.Context, in *engine.SearchAgentStatusStatisticItemRequest) (*engine.AgentStatusStatisticItem, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	if in.GetTime() == nil {
		return nil, model.NewBadRequestError("grpc.agent.report.call", "filter time is required")
	}
	var item *model.SupervisorAgentItem

	item, err = api.ctrl.GetSupervisorAgentItem(ctx, session, in.AgentId, &model.FilterBetween{
		From: in.GetTime().GetFrom(),
		To:   in.GetTime().GetTo(),
	})

	if err != nil {
		return nil, err
	}

	return &engine.AgentStatusStatisticItem{
		AgentId:          item.AgentId,
		Name:             item.Name,
		Status:           item.Status,
		StatusDuration:   item.StatusDuration,
		User:             GetProtoLookup(&item.User),
		Extension:        item.Extension,
		Team:             GetProtoLookup(item.Team),
		Supervisor:       GetProtoLookups(item.Supervisor),
		Auditor:          GetProtoLookups(item.Auditor),
		Region:           GetProtoLookup(item.Region),
		ProgressiveCount: item.ProgressiveCount,
		ChatCount:        item.ChatCount,
		PauseCause:       item.PauseCause,
		Online:           item.Online,
		Offline:          item.Offline,
		Pause:            item.Pause,
		ScoreRequiredAvg: item.ScoreRequiredAvg,
		ScoreOptionalAvg: item.ScoreOptionalAvg,
		ScoreCount:       item.ScoreCount,
		DescTrack:        item.DescTrack,
		StatusComment:    item.StatusComment,
		ScreenControl:    item.ScreenControl,
	}, nil
}

func (api *agent) SearchUserStatus(ctx context.Context, in *engine.SearchUserStatusRequest) (*engine.ListUserStatus, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	var list []*model.UserStatus
	var endList bool
	req := &model.SearchUserStatus{
		ListRequest: model.ListRequest{
			Q:       in.GetQ(),
			Page:    int(in.GetPage()),
			PerPage: int(in.GetSize()),
			Fields:  in.Fields,
			Sort:    in.Sort,
		},
	}
	// Commented by WTEL-4615 requirement
	// TODO DEV-4075
	//lq := len(req.Q)
	//if lq != 0 && rune(req.Q[lq-1]) != rune('*') {
	//	req.Q += "*"
	//}

	list, endList, err = api.ctrl.SearchUserStatus(ctx, session, req)
	if err != nil {
		return nil, err
	}

	items := make([]*engine.UserStatus, 0, len(list))
	for _, v := range list {
		items = append(items, toUserStatus(v))
	}
	return &engine.ListUserStatus{
		Next:  !endList,
		Items: items,
	}, nil
}

func toEngineAgentCallStatistics(src *model.AgentCallStatistics) *engine.AgentCallStatistics {
	return &engine.AgentCallStatistics{
		Name:        src.Name,
		Count:       src.Count,
		Abandoned:   src.Abandoned,
		Handles:     src.Handles,
		SumTalkSec:  src.SumTalkSec,
		AvgTalkSec:  src.AvgTalkSec,
		MinTalkSec:  src.MinTalkSec,
		MaxTalkSec:  src.MaxTalkSec,
		SumHoldSec:  src.SumHoldSec,
		AvgHoldSec:  src.AvgHoldSec,
		MinHoldSec:  src.MinHoldSec,
		MaxHoldSec:  src.MaxHoldSec,
		Utilization: src.Utilization,
		Occupancy:   src.Occupancy,
		ChatAccepts: src.ChatAccepts,
		ChatAht:     src.ChatAHT,
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
		Team:           GetProtoLookup(src.Team),
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
		Transferred:    src.Transferred,
		Skills:         GetProtoLookups(src.Skills),
		Supervisor:     GetProtoLookups(src.Supervisor),
		Auditor:        GetProtoLookups(src.Auditor),
		PauseCause:     src.PauseCause,
		ChatCount:      src.ChatCount,
		Occupancy:      src.Occupancy,
		DescTrack:      src.DescTrack,
		ScreenControl:  src.ScreenControl,
	}

	if src.ActiveCallId != nil {
		item.ActiveCallId = *src.ActiveCallId
	}

	if src.StatusComment != nil {
		item.StatusComment = *src.StatusComment
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
		Status:                src.Status,
		Description:           src.Description,
		LastStatusChange:      src.LastStatusChange,
		ProgressiveCount:      int32(src.ProgressiveCount),
		Name:                  src.Name,
		StatusDuration:        src.StatusDuration,
		GreetingMedia:         GetProtoLookup(src.GreetingMedia),
		AllowChannels:         src.AllowChannels,
		ChatCount:             src.ChatCount,
		Supervisor:            GetProtoLookups(src.Supervisor),
		Team:                  GetProtoLookup(src.Team),
		Region:                GetProtoLookup(src.Region),
		Auditor:               GetProtoLookups(src.Auditor),
		IsSupervisor:          src.IsSupervisor,
		Skills:                GetProtoLookups(src.Skills),
		TaskCount:             src.TaskCount,
		ScreenControl:         src.ScreenControl,
		AllowSetScreenControl: src.AllowSetScreenControl,
	}
	agent.Channel = make([]*engine.AgentChannel, 0, len(src.Channel))

	for _, v := range src.Channel {
		c := &engine.AgentChannel{
			Channel:  v.Channel,
			State:    v.State,
			JoinedAt: v.JoinedAt,
		}

		if v.Timeout != nil {
			c.Timeout = *v.Timeout
		}
		agent.Channel = append(agent.Channel, c)
	}

	if src.Extension != nil {
		agent.Extension = *src.Extension
	}

	userPresenceStatus := ""
	if len(src.UserPresenceStatus) != 0 {
		userPresenceStatus = strings.Join(src.UserPresenceStatus, ",")
		agent.UserPresenceStatus = &engine.Agent_UserPresence{
			Status: "{" + userPresenceStatus + "}",
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

func toUserStatus(src *model.UserStatus) *engine.UserStatus {
	s := ""
	if len(src.Presence) != 0 {
		s = strings.Join(src.Presence, ",")
	}
	return &engine.UserStatus{
		Id:        src.Id,
		Name:      src.Name,
		Extension: src.Extension,
		Presence: &engine.UserStatus_UserPresence{
			Status: "{" + s + "}",
		},
		Status: src.Status,
	}
}
