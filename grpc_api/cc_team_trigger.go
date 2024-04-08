package grpc_api

import (
	gogrpc "buf.build/gen/go/webitel/engine/grpc/go/_gogrpc"
	engine "buf.build/gen/go/webitel/engine/protocolbuffers/go"
	"context"
	"github.com/webitel/engine/model"
)

type teamTrigger struct {
	*API
	gogrpc.UnsafeTeamTriggerServiceServer
}

func NewTeamTriggerApi(api *API) *teamTrigger {
	return &teamTrigger{API: api}
}

func (api *teamTrigger) CreateTeamTrigger(ctx context.Context, in *engine.CreateTeamTriggerRequest) (*engine.TeamTrigger, error) {
	session, err := api.ctrl.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	trigger := &model.TeamTrigger{
		Schema: &model.Lookup{
			Id: int(in.GetSchema().GetId()),
		},
		Name:        in.Name,
		Description: in.Description,
		Enabled:     in.Enabled,
	}

	trigger, err = api.ctrl.CreateTeamTrigger(ctx, session, in.TeamId, trigger)
	if err != nil {
		return nil, err
	}

	return toEngineTeamTrigger(trigger), nil
}

func (api *teamTrigger) SearchTeamTrigger(ctx context.Context, in *engine.SearchTeamTriggerRequest) (*engine.ListTeamTrigger, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	var list []*model.TeamTrigger
	var endList bool
	req := &model.SearchTeamTrigger{
		ListRequest: model.ListRequest{
			Q:       in.GetQ(),
			Page:    int(in.GetPage()),
			PerPage: int(in.GetSize()),
			Fields:  in.Fields,
			Sort:    in.Sort,
		},
		Ids:       in.GetId(),
		SchemaIds: in.SchemaId,
	}

	if in.Enabled != nil {
		req.Enabled = &in.Enabled.Value
	}

	list, endList, err = api.ctrl.SearchTeamTrigger(ctx, session, in.TeamId, req)

	if err != nil {
		return nil, err
	}

	items := make([]*engine.TeamTrigger, 0, len(list))
	for _, v := range list {
		items = append(items, toEngineTeamTrigger(v))
	}
	return &engine.ListTeamTrigger{
		Next:  !endList,
		Items: items,
	}, nil
}

func (api *teamTrigger) RunTeamTrigger(ctx context.Context, in *engine.RunTeamTriggerRequest) (*engine.RunTeamTriggerResponse, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	res := &engine.RunTeamTriggerResponse{}

	res.JobId, err = api.ctrl.RunAgentTrigger(ctx, session, in.TriggerId, in.Variables)

	if err != nil {
		return nil, err
	}

	return res, nil
}

func (api *teamTrigger) ReadTeamTrigger(ctx context.Context, in *engine.ReadTeamTriggerRequest) (*engine.TeamTrigger, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	var trigger *model.TeamTrigger
	trigger, err = api.ctrl.GetTeamTrigger(ctx, session, in.TeamId, in.Id)

	if err != nil {
		return nil, err
	}

	return toEngineTeamTrigger(trigger), nil
}

func (api *teamTrigger) UpdateTeamTrigger(ctx context.Context, in *engine.UpdateTeamTriggerRequest) (*engine.TeamTrigger, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	trigger := &model.TeamTrigger{
		Id: in.Id,
		Schema: &model.Lookup{
			Id: int(in.GetSchema().GetId()),
		},
		Enabled:     in.Enabled,
		Name:        in.Name,
		Description: in.Description,
	}

	trigger, err = api.ctrl.UpdateTeamTrigger(ctx, session, in.TeamId, trigger)

	if err != nil {
		return nil, err
	}

	return toEngineTeamTrigger(trigger), nil
}

func (api *teamTrigger) PatchTeamTrigger(ctx context.Context, in *engine.PatchTeamTriggerRequest) (*engine.TeamTrigger, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	var trigger *model.TeamTrigger
	patch := &model.TeamTriggerPatch{}

	//TODO
	for _, v := range in.Fields {
		switch v {
		case "schema.id":
			patch.Schema = &model.Lookup{
				Id: int(in.GetSchema().GetId()),
			}
		case "name":
			patch.Name = &in.Name
		case "description":
			patch.Description = &in.Description
		case "enabled":
			patch.Enabled = &in.Enabled
		}
	}

	if trigger, err = api.ctrl.PatchTeamTrigger(ctx, session, in.TeamId, in.Id, patch); err != nil {
		return nil, err
	}

	return toEngineTeamTrigger(trigger), nil
}

func (api *teamTrigger) DeleteTeamTrigger(ctx context.Context, in *engine.DeleteTeamTriggerRequest) (*engine.TeamTrigger, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	var trigger *model.TeamTrigger
	trigger, err = api.ctrl.DeleteTeamTrigger(ctx, session, in.TeamId, in.Id)
	if err != nil {
		return nil, err
	}

	return toEngineTeamTrigger(trigger), nil
}

func toEngineTeamTrigger(src *model.TeamTrigger) *engine.TeamTrigger {
	return &engine.TeamTrigger{
		Id:          src.Id,
		Schema:      GetProtoLookup(src.Schema),
		Enabled:     src.Enabled,
		Name:        src.Name,
		Description: src.Description,
	}
}
