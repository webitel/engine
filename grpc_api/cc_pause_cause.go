package grpc_api

import (
	"context"
	"github.com/webitel/engine/model"
	"github.com/webitel/protos/engine"
)

type pauseCause struct {
	*API
}

func NewPauseCause(api *API) *pauseCause {
	return &pauseCause{api}
}

func (api *pauseCause) CreateAgentPauseCause(ctx context.Context, in *engine.CreateAgentPauseCauseRequest) (*engine.AgentPauseCause, error) {
	session, err := api.ctrl.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	cause := &model.AgentPauseCause{
		Name:            in.Name,
		Description:     in.Description,
		LimitPerDay:     in.LimitPerDay,
		AllowSupervisor: in.AllowSupervisor,
		AllowAgent:      in.AllowAgent,
	}

	cause, err = api.ctrl.CreatePauseCause(session, cause)
	if err != nil {
		return nil, err
	}

	return toEnginePauseCause(cause), nil
}

func (api *pauseCause) SearchAgentPauseCause(ctx context.Context, in *engine.SearchAgentPauseCauseRequest) (*engine.ListAgentPauseCause, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	var list []*model.AgentPauseCause
	var endList bool
	req := &model.SearchAgentPauseCause{
		ListRequest: model.ListRequest{
			Q:       in.GetQ(),
			Page:    int(in.GetPage()),
			PerPage: int(in.GetSize()),
			Fields:  in.Fields,
		},
		Ids: in.GetId(),
	}

	list, endList, err = api.ctrl.SearchPauseCause(session, req)

	if err != nil {
		return nil, err
	}

	items := make([]*engine.AgentPauseCause, 0, len(list))
	for _, v := range list {
		items = append(items, toEnginePauseCause(v))
	}
	return &engine.ListAgentPauseCause{
		Next:  !endList,
		Items: items,
	}, nil
}

func (api *pauseCause) ReadAgentPauseCause(ctx context.Context, in *engine.ReadAgentPauseCauseRequest) (*engine.AgentPauseCause, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	var cause *model.AgentPauseCause
	cause, err = api.ctrl.GetPauseCause(session, in.Id)

	if err != nil {
		return nil, err
	}

	return toEnginePauseCause(cause), nil
}

func (api *pauseCause) PatchAgentPauseCause(ctx context.Context, in *engine.PatchAgentPauseCauseRequest) (*engine.AgentPauseCause, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	var cause *model.AgentPauseCause
	patch := &model.AgentPauseCausePatch{}

	//TODO
	for _, v := range in.Fields {
		switch v {
		case "name":
			patch.Name = model.NewString(in.Name)
		case "description":
			patch.Description = model.NewString(in.Description)
		case "limit_per_day":
			patch.LimitPerDay = &in.LimitPerDay
		case "allow_supervisor":
			patch.AllowSupervisor = &in.AllowSupervisor
		case "allow_agent":
			patch.AllowAgent = &in.AllowAgent

		}
	}

	if cause, err = api.ctrl.PatchPauseCause(session, in.Id, patch); err != nil {
		return nil, err
	}

	return toEnginePauseCause(cause), nil
}

func (api *pauseCause) UpdateAgentPauseCause(ctx context.Context, in *engine.UpdateAgentPauseCauseRequest) (*engine.AgentPauseCause, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	cause := &model.AgentPauseCause{
		AclRecord: model.AclRecord{
			Id: int(in.Id),
		},
		Name:            in.Name,
		Description:     in.Description,
		LimitPerDay:     in.LimitPerDay,
		AllowSupervisor: in.AllowSupervisor,
		AllowAgent:      in.AllowAgent,
	}

	cause, err = api.ctrl.UpdatePauseCause(session, cause)

	if err != nil {
		return nil, err
	}

	return toEnginePauseCause(cause), nil
}

func (api *pauseCause) DeleteAgentPauseCause(ctx context.Context, in *engine.DeleteAgentPauseCauseRequest) (*engine.AgentPauseCause, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	var cause *model.AgentPauseCause
	cause, err = api.ctrl.DeletePauseCause(session, in.Id)
	if err != nil {
		return nil, err
	}

	return toEnginePauseCause(cause), nil
}

func toEnginePauseCause(src *model.AgentPauseCause) *engine.AgentPauseCause {
	return &engine.AgentPauseCause{
		Id:              uint32(src.Id),
		CreatedAt:       model.TimeToInt64(src.CreatedAt),
		CreatedBy:       GetProtoLookup(&src.CreatedBy),
		UpdatedAt:       model.TimeToInt64(src.UpdatedAt),
		UpdatedBy:       GetProtoLookup(&src.UpdatedBy),
		Name:            src.Name,
		LimitPerDay:     src.LimitPerDay,
		AllowSupervisor: src.AllowSupervisor,
		AllowAgent:      src.AllowAgent,
		Description:     src.Description,
	}
}
