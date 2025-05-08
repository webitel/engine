package grpc_api

import (
	"context"
	"github.com/webitel/engine/gen/engine"
	"github.com/webitel/engine/model"
)

type webHook struct {
	*API
	engine.UnsafeWebHookServiceServer
}

func NewWebHookApi(api *API) *webHook {
	return &webHook{API: api}
}

func (api *webHook) CreateWebHook(ctx context.Context, in *engine.CreateWebHookRequest) (*engine.WebHook, error) {
	session, err := api.ctrl.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	hook := &model.WebHook{
		Name:        in.GetName(),
		Description: in.GetDescription(),
		Origin:      in.GetOrigin(),
		Schema: &model.Lookup{
			Id: int(in.GetSchema().GetId()),
		},
		Enabled:       in.GetEnabled(),
		Authorization: in.GetAuthorization(),
	}

	hook, err = api.ctrl.CreateWebHook(ctx, session, hook)
	if err != nil {
		return nil, err
	}

	return toEngineWebHook(hook), nil
}

func (api *webHook) SearchWebHook(ctx context.Context, in *engine.SearchWebHookRequest) (*engine.ListWebHook, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	var list []*model.WebHook
	var endList bool
	req := &model.SearchWebHook{
		ListRequest: model.ListRequest{
			Q:       in.GetQ(),
			Page:    int(in.GetPage()),
			PerPage: int(in.GetSize()),
			Fields:  in.Fields,
			Sort:    in.Sort,
		},
		Ids: in.GetId(),
	}

	list, endList, err = api.ctrl.SearchWebHook(ctx, session, req)

	if err != nil {
		return nil, err
	}

	items := make([]*engine.WebHook, 0, len(list))
	for _, v := range list {
		items = append(items, toEngineWebHook(v))
	}
	return &engine.ListWebHook{
		Next:  !endList,
		Items: items,
	}, nil
}

func (api *webHook) ReadWebHook(ctx context.Context, in *engine.ReadWebHookRequest) (*engine.WebHook, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	var hook *model.WebHook
	hook, err = api.ctrl.GetWebHook(ctx, session, in.Id)

	if err != nil {
		return nil, err
	}

	return toEngineWebHook(hook), nil
}

func (api *webHook) UpdateWebHook(ctx context.Context, in *engine.UpdateWebHookRequest) (*engine.WebHook, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	hook := &model.WebHook{
		Id:          in.Id,
		Name:        in.Name,
		Description: in.Description,
		Origin:      in.Origin,
		Schema: &model.Lookup{
			Id: int(in.GetSchema().GetId()),
		},
		Enabled:       in.GetEnabled(),
		Authorization: in.GetAuthorization(),
	}

	hook, err = api.ctrl.UpdateWebHook(ctx, session, hook)

	if err != nil {
		return nil, err
	}

	return toEngineWebHook(hook), nil
}

func (api *webHook) PatchWebHook(ctx context.Context, in *engine.PatchWebHookRequest) (*engine.WebHook, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	var hook *model.WebHook
	patch := &model.WebHookPatch{}

	//TODO
	for _, v := range in.Fields {
		switch v {
		case "name":
			patch.Name = &in.Name
		case "description":
			patch.Description = &in.Description
		case "enabled":
			patch.Enabled = &in.Enabled
		case "origin":
			patch.Origin = in.Origin
		case "authorization":
			patch.Authorization = &in.Authorization
		case "schema.id":
			patch.Schema = &model.Lookup{
				Id: int(in.GetSchema().GetId()),
			}
		}
	}

	hook, err = api.ctrl.PatchWebHook(ctx, session, in.GetId(), patch)

	if err != nil {
		return nil, err
	}

	return toEngineWebHook(hook), nil
}

func (api *webHook) DeleteWebHook(ctx context.Context, in *engine.DeleteWebHookRequest) (*engine.WebHook, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	var hook *model.WebHook
	hook, err = api.ctrl.DeleteWebHook(ctx, session, in.Id)
	if err != nil {
		return nil, err
	}

	return toEngineWebHook(hook), nil
}

func toEngineWebHook(src *model.WebHook) *engine.WebHook {
	return &engine.WebHook{
		Id:            src.Id,
		Key:           src.Key,
		CreatedAt:     model.TimeToInt64(src.CreatedAt),
		CreatedBy:     GetProtoLookup(src.CreatedBy),
		UpdatedAt:     model.TimeToInt64(src.UpdatedAt),
		UpdatedBy:     GetProtoLookup(src.UpdatedBy),
		Name:          src.Name,
		Description:   src.Description,
		Origin:        src.Origin,
		Schema:        GetProtoLookup(src.Schema),
		Enabled:       src.Enabled,
		Authorization: src.Authorization,
	}
}
