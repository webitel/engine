package grpc_api

import (
	"context"
	"github.com/webitel/engine/gen/engine"
	"github.com/webitel/engine/model"
)

type chatPlanApi struct {
	*API
	engine.UnsafeRoutingChatPlanServiceServer
}

func NewChatPlan(api *API) *chatPlanApi {
	return &chatPlanApi{API: api}
}

func (api *chatPlanApi) CreateChatPlan(ctx context.Context, in *engine.CreateChatPlanRequest) (*engine.ChatPlan, error) {
	session, err := api.ctrl.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	plan := &model.ChatPlan{
		Enabled: in.Enabled,
		Name:    in.Name,
		Schema: model.Lookup{
			Id: int(in.GetSchema().GetId()),
		},
		Description: in.Description,
	}

	plan, err = api.ctrl.CreateChatPlan(ctx, session, plan)
	if err != nil {
		return nil, err
	}

	return toEngineChatPlan(plan), nil
}

func (api *chatPlanApi) SearchChatPlan(ctx context.Context, in *engine.SearchChatPlanRequest) (*engine.ListChatPlan, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	var list []*model.ChatPlan
	var endList bool
	req := &model.SearchChatPlan{
		ListRequest: model.ListRequest{
			Q:       in.GetQ(),
			Page:    int(in.GetPage()),
			PerPage: int(in.GetSize()),
			Fields:  in.Fields,
			Sort:    in.Sort,
		},
		Ids: in.GetId(),
	}

	if in.Name != "" {
		req.Name = &in.Name
	}

	if in.Enabled {
		req.Enabled = &in.Enabled
	}

	list, endList, err = api.ctrl.SearchChatPlan(ctx, session, req)

	if err != nil {
		return nil, err
	}

	items := make([]*engine.ChatPlan, 0, len(list))
	for _, v := range list {
		items = append(items, toEngineChatPlan(v))
	}
	return &engine.ListChatPlan{
		Next:  !endList,
		Items: items,
	}, nil
}

func (api *chatPlanApi) ReadChatPlan(ctx context.Context, in *engine.ReadChatPlanRequest) (*engine.ChatPlan, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	var plan *model.ChatPlan
	plan, err = api.ctrl.GetChatPlan(ctx, session, in.Id)

	if err != nil {
		return nil, err
	}

	return toEngineChatPlan(plan), nil
}

func (api *chatPlanApi) UpdateChatPlan(ctx context.Context, in *engine.UpdateChatPlanRequest) (*engine.ChatPlan, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	plan := &model.ChatPlan{
		Id:      in.GetId(),
		Enabled: in.GetEnabled(),
		Name:    in.GetName(),
		Schema: model.Lookup{
			Id: int(in.GetSchema().GetId()),
		},
		Description: in.GetDescription(),
	}

	plan, err = api.ctrl.UpdateChatPlan(ctx, session, plan)

	if err != nil {
		return nil, err
	}

	return toEngineChatPlan(plan), nil
}

func (api *chatPlanApi) PatchChatPlan(ctx context.Context, in *engine.PatchChatPlanRequest) (*engine.ChatPlan, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	var plan *model.ChatPlan
	patch := &model.PatchChatPlan{}

	//TODO
	for _, v := range in.Fields {
		switch v {
		case "name":
			patch.Name = model.NewString(in.Name)
		case "description":
			patch.Description = model.NewString(in.Description)
		case "enabled":
			patch.Enabled = model.NewBool(in.Enabled)
		case "schema":
			patch.Schema = GetLookup(in.Schema)
		}
	}

	if plan, err = api.ctrl.PatchChatPlan(ctx, session, in.Id, patch); err != nil {
		return nil, err
	}

	return toEngineChatPlan(plan), nil
}

func (api *chatPlanApi) DeleteChatPlan(ctx context.Context, in *engine.DeleteChatPlanRequest) (*engine.ChatPlan, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	var plan *model.ChatPlan
	plan, err = api.ctrl.DeleteChatPlan(ctx, session, in.Id)
	if err != nil {
		return nil, err
	}

	return toEngineChatPlan(plan), nil
}

func toEngineChatPlan(src *model.ChatPlan) *engine.ChatPlan {
	return &engine.ChatPlan{
		Id:          src.Id,
		Name:        src.Name,
		Description: src.Description,
		Schema:      GetProtoLookup(&src.Schema),
		Enabled:     src.Enabled,
	}
}
