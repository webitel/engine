package grpc_api

import (
	"context"
	"github.com/webitel/engine/model"
	"github.com/webitel/protos/engine"
)

type queueHook struct {
	*API
	engine.UnsafeQueueHookServiceServer
}

func NewQueueHookApi(api *API) *queueHook {
	return &queueHook{API: api}
}

func (api *queueHook) CreateQueueHook(ctx context.Context, in *engine.CreateQueueHookRequest) (*engine.QueueHook, error) {
	session, err := api.ctrl.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	hook := &model.QueueHook{
		Schema: model.Lookup{
			Id: int(in.GetSchema().GetId()),
		},
		Event:   in.Event,
		Enabled: in.Enabled,
	}

	hook, err = api.ctrl.CreateQueueHook(ctx, session, in.QueueId, hook)
	if err != nil {
		return nil, err
	}

	return toEngineQueueHook(hook), nil
}

func (api *queueHook) SearchQueueHook(ctx context.Context, in *engine.SearchQueueHookRequest) (*engine.ListQueueHook, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	var list []*model.QueueHook
	var endList bool
	req := &model.SearchQueueHook{
		ListRequest: model.ListRequest{
			Q:       in.GetQ(),
			Page:    int(in.GetPage()),
			PerPage: int(in.GetSize()),
			Fields:  in.Fields,
			Sort:    in.Sort,
		},
		Ids:       in.GetId(),
		SchemaIds: in.SchemaId,
		Events:    in.Event,
	}

	list, endList, err = api.ctrl.SearchQueueHook(ctx, session, in.QueueId, req)

	if err != nil {
		return nil, err
	}

	items := make([]*engine.QueueHook, 0, len(list))
	for _, v := range list {
		items = append(items, toEngineQueueHook(v))
	}
	return &engine.ListQueueHook{
		Next:  !endList,
		Items: items,
	}, nil
}

func (api *queueHook) ReadQueueHook(ctx context.Context, in *engine.ReadQueueHookRequest) (*engine.QueueHook, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	var hook *model.QueueHook
	hook, err = api.ctrl.GetQueueHook(ctx, session, in.QueueId, in.Id)

	if err != nil {
		return nil, err
	}

	return toEngineQueueHook(hook), nil
}

func (api *queueHook) PatchQueueHook(ctx context.Context, in *engine.PatchQueueHookRequest) (*engine.QueueHook, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	var hook *model.QueueHook
	patch := &model.QueueHookPatch{}

	//TODO
	for _, v := range in.Fields {
		switch v {
		case "schema.id":
			patch.Schema = &model.Lookup{
				Id: int(in.GetSchema().GetId()),
			}
		case "event":
			patch.Event = &in.Event
		case "enabled":
			patch.Enabled = &in.Enabled
		}
	}

	if hook, err = api.ctrl.PatchQueueHook(ctx, session, in.QueueId, in.Id, patch); err != nil {
		return nil, err
	}

	return toEngineQueueHook(hook), nil
}

func (api *queueHook) UpdateQueueHook(ctx context.Context, in *engine.UpdateQueueHookRequest) (*engine.QueueHook, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	hook := &model.QueueHook{
		Id: in.Id,
		Schema: model.Lookup{
			Id: int(in.GetSchema().GetId()),
		},
		Event:   in.Event,
		Enabled: in.Enabled,
	}

	hook, err = api.ctrl.UpdateQueueHook(ctx, session, in.QueueId, hook)

	if err != nil {
		return nil, err
	}

	return toEngineQueueHook(hook), nil
}

func (api *queueHook) DeleteQueueHook(ctx context.Context, in *engine.DeleteQueueHookRequest) (*engine.QueueHook, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	var hook *model.QueueHook
	hook, err = api.ctrl.DeleteQueueHook(ctx, session, in.QueueId, in.Id)
	if err != nil {
		return nil, err
	}

	return toEngineQueueHook(hook), nil
}

func toEngineQueueHook(src *model.QueueHook) *engine.QueueHook {
	return &engine.QueueHook{
		Id:         src.Id,
		Schema:     GetProtoLookup(&src.Schema),
		Event:      src.Event,
		Enabled:    src.Enabled,
		Properties: nil,
	}
}
