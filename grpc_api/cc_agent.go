package grpc_api

import (
	"context"
	"github.com/webitel/engine/app"
	"github.com/webitel/engine/grpc_api/engine"
	"github.com/webitel/engine/model"
)

type agent struct {
	app *app.App
}

func NewAgentApi(app *app.App) *agent {
	return &agent{app: app}
}

func (api *agent) CreateAgent(ctx context.Context, in *engine.CreateAgentRequest) (*engine.Agent, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_AGENT)
	if !permission.CanCreate() {
		return nil, api.app.MakePermissionError(session, permission, model.PERMISSION_ACCESS_CREATE)
	}

	agent := &model.Agent{
		DomainId: session.Domain(in.GetDomainId()),
		User: model.Lookup{
			Id: int(in.GetUser().GetId()),
		},
		Description: in.Description,
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
		return nil, api.app.MakePermissionError(session, permission, model.PERMISSION_ACCESS_READ)
	}

	var list []*model.Agent

	if permission.Rbac {
		list, err = api.app.GetAgentsPageByGroups(session.Domain(in.DomainId), session.RoleIds, int(in.Page), int(in.Size))
	} else {
		list, err = api.app.GetAgentsPage(session.Domain(in.DomainId), int(in.Page), int(in.Size))
	}

	if err != nil {
		return nil, err
	}

	items := make([]*engine.Agent, 0, len(list))
	for _, v := range list {
		items = append(items, transformAgent(v))
	}

	return &engine.ListAgent{
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
		return nil, api.app.MakePermissionError(session, permission, model.PERMISSION_ACCESS_READ)
	}

	var agent *model.Agent

	if permission.Rbac {
		var perm bool
		if perm, err = api.app.AgentCheckAccess(session.Domain(in.GetDomainId()), in.GetId(), session.RoleIds, model.PERMISSION_ACCESS_READ); err != nil {
			return nil, err
		} else if !perm {
			return nil, api.app.MakeResourcePermissionError(session, in.GetId(), permission, model.PERMISSION_ACCESS_READ)
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
		return nil, api.app.MakePermissionError(session, permission, model.PERMISSION_ACCESS_READ)
	}

	if !permission.CanUpdate() {
		return nil, api.app.MakePermissionError(session, permission, model.PERMISSION_ACCESS_UPDATE)
	}

	if permission.Rbac {
		var perm bool
		if perm, err = api.app.AgentCheckAccess(session.Domain(in.GetDomainId()), in.GetId(), session.RoleIds, model.PERMISSION_ACCESS_UPDATE); err != nil {
			return nil, err
		} else if !perm {
			return nil, api.app.MakeResourcePermissionError(session, in.GetId(), permission, model.PERMISSION_ACCESS_UPDATE)
		}
	}

	var agent *model.Agent

	agent, err = api.app.UpdateAgent(&model.Agent{
		Id:       in.Id,
		DomainId: session.Domain(in.GetDomainId()),
		User: model.Lookup{
			Id: int(in.GetUser().GetId()),
		},
		Description: in.Description,
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
		return nil, api.app.MakePermissionError(session, permission, model.PERMISSION_ACCESS_DELETE)
	}

	if permission.Rbac {
		var perm bool
		if perm, err = api.app.AgentCheckAccess(session.Domain(in.GetDomainId()), in.GetId(), session.RoleIds, model.PERMISSION_ACCESS_DELETE); err != nil {
			return nil, err
		} else if !perm {
			return nil, api.app.MakeResourcePermissionError(session, in.GetId(), permission, model.PERMISSION_ACCESS_DELETE)
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
			return nil, api.app.MakePermissionError(session, permission, model.PERMISSION_ACCESS_UPDATE)
		}

		if permission.Rbac {
			var perm bool
			if perm, err = api.app.AgentCheckAccess(session.Domain(in.GetDomainId()), in.GetId(), session.RoleIds, model.PERMISSION_ACCESS_UPDATE); err != nil {
				return nil, err
			} else if !perm {
				return nil, api.app.MakeResourcePermissionError(session, in.GetId(), permission, model.PERMISSION_ACCESS_UPDATE)
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
		return nil, api.app.MakePermissionError(session, permission, model.PERMISSION_ACCESS_READ)
	}

	if permission.Rbac {
		var perm bool
		if perm, err = api.app.AgentCheckAccess(session.Domain(in.GetDomainId()), in.GetId(), session.RoleIds, model.PERMISSION_ACCESS_READ); err != nil {
			return nil, err
		} else if !perm {
			return nil, api.app.MakeResourcePermissionError(session, in.GetId(), permission, model.PERMISSION_ACCESS_READ)
		}
	}

	var list []*model.AgentInTeam
	list, err = api.app.GetAgentInTeamPage(session.Domain(in.GetDomainId()), in.GetId(), int(in.GetPage()), int(in.GetSize()))
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
		return nil, api.app.MakePermissionError(session, permission, model.PERMISSION_ACCESS_READ)
	}

	if permission.Rbac {
		var perm bool
		if perm, err = api.app.AgentCheckAccess(session.Domain(in.GetDomainId()), in.GetId(), session.RoleIds, model.PERMISSION_ACCESS_READ); err != nil {
			return nil, err
		} else if !perm {
			return nil, api.app.MakeResourcePermissionError(session, in.GetId(), permission, model.PERMISSION_ACCESS_READ)
		}
	}

	var list []*model.AgentInQueue
	list, err = api.app.GetAgentInQueuePage(session.Domain(in.GetDomainId()), in.GetId(), int(in.GetPage()), int(in.GetSize()))
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
		return nil, api.app.MakePermissionError(session, permission, model.PERMISSION_ACCESS_READ)
	}

	if permission.Rbac {
		var perm bool
		if perm, err = api.app.AgentCheckAccess(session.Domain(in.GetDomainId()), in.GetAgentId(), session.RoleIds, model.PERMISSION_ACCESS_READ); err != nil {
			return nil, err
		} else if !perm {
			return nil, api.app.MakeResourcePermissionError(session, in.GetAgentId(), permission, model.PERMISSION_ACCESS_READ)
		}
	}

	var list []*model.AgentState
	list, err = api.app.GetAgentStateHistoryPage(in.GetAgentId(), in.GetFromTime(), in.GetToTime(), int(in.GetPage()), int(in.GetSize()))
	if err != nil {
		return nil, err
	}

	items := make([]*engine.AgentState, 0, len(list))

	for _, v := range list {
		items = append(items, toEngineAgentState(v))
	}

	return &engine.ListAgentStateHistory{
		Items: items,
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
		LastStateChange: src.LastStateChange,
		Status:          src.Status,
		State:           src.State,
		Description:     src.Description,
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
		QueueId:   0,
	}

	if src.QueueId != nil {
		st.QueueId = *src.QueueId
	}

	if src.TimeoutAt != nil {
		st.TimeoutAt = *src.TimeoutAt
	}

	return st
}
