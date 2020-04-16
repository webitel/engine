package grpc_api

import (
	"context"
	"github.com/webitel/engine/auth_manager"
	"github.com/webitel/engine/grpc_api/engine"
	"github.com/webitel/engine/model"
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

	if err = api.app.SetAgentStatus(session.Domain(in.GetDomainId()), in.GetId(), getAgentStatus(in.Status)); err != nil {
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

func (api *agent) SearchAgentStateHistory(ctx context.Context, in *engine.SearchAgentStateHistoryRequest) (*engine.ListAgentStateHistory, error) {
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
		From: in.GetTimeFrom(),
		To:   in.GetTimeTo(),
	}
	list, endList, err = api.app.GetAgentStateHistoryPage(in.GetAgentId(), req)
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

func getAgentStatus(name string) model.AgentStatus {
	return model.AgentStatus{name}
}

func transformAgent(src *model.Agent) *engine.Agent {
	agent := &engine.Agent{
		Id: src.Id,
		User: &engine.Lookup{
			Id:   int64(src.User.Id),
			Name: src.User.Name,
		},
		LastStateChange:  src.LastStateChange,
		Status:           src.Status,
		State:            src.State,
		Description:      src.Description,
		ProgressiveCount: int32(src.ProgressiveCount),
		Name:             src.Name,
	}

	if src.StateTimeout != nil {
		agent.StateTimeout = *src.StateTimeout
	}

	return agent
}

func toEngineAgentState(src *model.AgentState) *engine.AgentState {
	st := &engine.AgentState{
		Id:        src.Id,
		JoinedAt:  src.JoinedAt,
		State:     src.State,
		TimeoutAt: 0,
	}

	if src.Queue != nil {
		st.Queue = &engine.Lookup{
			Id:   int64(src.Queue.Id),
			Name: src.Queue.Name,
		}
	}

	if src.TimeoutAt != nil {
		st.TimeoutAt = *src.TimeoutAt
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
