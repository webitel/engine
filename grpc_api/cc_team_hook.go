package grpc_api

import (
	gogrpc "buf.build/gen/go/webitel/engine/grpc/go/_gogrpc"
	engine "buf.build/gen/go/webitel/engine/protocolbuffers/go"
	"context"
	"github.com/webitel/engine/model"
)

type teamHook struct {
	*API
	gogrpc.UnsafeTeamHookServiceServer
}

func NewTeamHookApi(api *API) *teamHook {
	return &teamHook{API: api}
}

func (api *teamHook) CreateTeamHook(ctx context.Context, in *engine.CreateTeamHookRequest) (*engine.TeamHook, error) {
	session, err := api.ctrl.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	hook := &model.TeamHook{
		Schema: model.Lookup{
			Id: int(in.GetSchema().GetId()),
		},
		Event:   in.GetEvent().String(),
		Enabled: in.Enabled,
	}

	hook, err = api.ctrl.CreateTeamHook(ctx, session, in.TeamId, hook)
	if err != nil {
		return nil, err
	}

	return toEngineTeamHook(hook), nil
}

func (api *teamHook) SearchTeamHook(ctx context.Context, in *engine.SearchTeamHookRequest) (*engine.ListTeamHook, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	var list []*model.TeamHook
	var endList bool
	req := &model.SearchTeamHook{
		ListRequest: model.ListRequest{
			Q:       in.GetQ(),
			Page:    int(in.GetPage()),
			PerPage: int(in.GetSize()),
			Fields:  in.Fields,
			Sort:    in.Sort,
		},
		Ids:       in.GetId(),
		SchemaIds: in.SchemaId,
		//Events:    in.Event,
	}

	list, endList, err = api.ctrl.SearchTeamHook(ctx, session, in.TeamId, req)

	if err != nil {
		return nil, err
	}

	items := make([]*engine.TeamHook, 0, len(list))
	for _, v := range list {
		items = append(items, toEngineTeamHook(v))
	}
	return &engine.ListTeamHook{
		Next:  !endList,
		Items: items,
	}, nil
}

func (api *teamHook) ReadTeamHook(ctx context.Context, in *engine.ReadTeamHookRequest) (*engine.TeamHook, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	var hook *model.TeamHook
	hook, err = api.ctrl.GetTeamHook(ctx, session, in.TeamId, in.Id)

	if err != nil {
		return nil, err
	}

	return toEngineTeamHook(hook), nil
}

func (api *teamHook) PatchTeamHook(ctx context.Context, in *engine.PatchTeamHookRequest) (*engine.TeamHook, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	var hook *model.TeamHook
	patch := &model.TeamHookPatch{}

	//TODO
	for _, v := range in.Fields {
		switch v {
		case "schema.id":
			patch.Schema = &model.Lookup{
				Id: int(in.GetSchema().GetId()),
			}
		case "event":
			patch.Event = model.NewString(in.Event.String())
		case "enabled":
			patch.Enabled = &in.Enabled
		}
	}

	if hook, err = api.ctrl.PatchTeamHook(ctx, session, in.TeamId, in.Id, patch); err != nil {
		return nil, err
	}

	return toEngineTeamHook(hook), nil
}

func (api *teamHook) UpdateTeamHook(ctx context.Context, in *engine.UpdateTeamHookRequest) (*engine.TeamHook, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	hook := &model.TeamHook{
		Id: in.Id,
		Schema: model.Lookup{
			Id: int(in.GetSchema().GetId()),
		},
		Event:   in.Event.String(),
		Enabled: in.Enabled,
	}

	hook, err = api.ctrl.UpdateTeamHook(ctx, session, in.TeamId, hook)

	if err != nil {
		return nil, err
	}

	return toEngineTeamHook(hook), nil
}

func (api *teamHook) DeleteTeamHook(ctx context.Context, in *engine.DeleteTeamHookRequest) (*engine.TeamHook, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	var hook *model.TeamHook
	hook, err = api.ctrl.DeleteTeamHook(ctx, session, in.TeamId, in.Id)
	if err != nil {
		return nil, err
	}

	return toEngineTeamHook(hook), nil
}

func toEngineTeamHook(src *model.TeamHook) *engine.TeamHook {
	e, _ := engine.TeamHookEvent_value[src.Event]
	return &engine.TeamHook{
		Id:         src.Id,
		Schema:     GetProtoLookup(&src.Schema),
		Event:      engine.TeamHookEvent(e),
		Enabled:    src.Enabled,
		Properties: nil,
	}
}
