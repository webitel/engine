package grpc_api

import (
	"context"
	"github.com/webitel/engine/gen/engine"
	"github.com/webitel/engine/model"
)

type pauseCause struct {
	*API
	engine.UnsafeAgentPauseCauseServiceServer
}

func NewPauseCause(api *API) *pauseCause {
	return &pauseCause{API: api}
}

func (api *pauseCause) CreateAgentPauseCause(ctx context.Context, in *engine.CreateAgentPauseCauseRequest) (*engine.AgentPauseCause, error) {
	session, err := api.ctrl.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	cause := &model.PauseCause{
		Name:            in.Name,
		Description:     in.Description,
		LimitMin:        in.LimitMin,
		AllowSupervisor: in.AllowSupervisor,
		AllowAdmin:      in.AllowAdmin,
		AllowAgent:      in.AllowAgent,
	}

	cause, err = api.ctrl.CreatePauseCause(ctx, session, cause)
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

	var list []*model.PauseCause
	var endList bool
	req := &model.SearchPauseCause{
		ListRequest: model.ListRequest{
			Q:       in.GetQ(),
			Page:    int(in.GetPage()),
			PerPage: int(in.GetSize()),
			Fields:  in.Fields,
			Sort:    in.Sort,
		},
		Ids: in.GetId(),
	}

	list, endList, err = api.ctrl.SearchPauseCause(ctx, session, req)

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
	var cause *model.PauseCause
	cause, err = api.ctrl.GetPauseCause(ctx, session, in.Id)

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

	var cause *model.PauseCause
	patch := &model.PauseCausePatch{}

	//TODO
	for _, v := range in.Fields {
		switch v {
		case "name":
			patch.Name = model.NewString(in.Name)
		case "description":
			patch.Description = model.NewString(in.Description)
		case "limit_min":
			patch.LimitMin = &in.LimitMin
		case "allow_supervisor":
			patch.AllowSupervisor = &in.AllowSupervisor
		case "allow_admin":
			patch.AllowAdmin = &in.AllowAdmin
		case "allow_agent":
			patch.AllowAgent = &in.AllowAgent

		}
	}

	if cause, err = api.ctrl.PatchPauseCause(ctx, session, in.Id, patch); err != nil {
		return nil, err
	}

	return toEnginePauseCause(cause), nil
}

func (api *pauseCause) UpdateAgentPauseCause(ctx context.Context, in *engine.UpdateAgentPauseCauseRequest) (*engine.AgentPauseCause, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	cause := &model.PauseCause{
		AclRecord:       model.AclRecord{},
		Id:              int(in.Id),
		Name:            in.Name,
		Description:     in.Description,
		LimitMin:        in.LimitMin,
		AllowSupervisor: in.AllowSupervisor,
		AllowAgent:      in.AllowAgent,
		AllowAdmin:      in.AllowAdmin,
	}

	cause, err = api.ctrl.UpdatePauseCause(ctx, session, cause)

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

	var cause *model.PauseCause
	cause, err = api.ctrl.DeletePauseCause(ctx, session, in.Id)
	if err != nil {
		return nil, err
	}

	return toEnginePauseCause(cause), nil
}

func toEnginePauseCause(src *model.PauseCause) *engine.AgentPauseCause {
	return &engine.AgentPauseCause{
		Id:              uint32(src.Id),
		CreatedAt:       model.TimeToInt64(src.CreatedAt),
		CreatedBy:       GetProtoLookup(src.CreatedBy),
		UpdatedAt:       model.TimeToInt64(src.UpdatedAt),
		UpdatedBy:       GetProtoLookup(src.UpdatedBy),
		Name:            src.Name,
		LimitMin:        src.LimitMin,
		AllowSupervisor: src.AllowSupervisor,
		AllowAdmin:      src.AllowAdmin,
		AllowAgent:      src.AllowAgent,
		Description:     src.Description,
	}
}
