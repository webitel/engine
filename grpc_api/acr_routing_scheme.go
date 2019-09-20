package grpc_api

import (
	"context"
	google_protobuf "github.com/golang/protobuf/ptypes/any"
	"github.com/webitel/engine/app"
	"github.com/webitel/engine/grpc_api/engine"
	"github.com/webitel/engine/model"
)

type routingScheme struct {
	app *app.App
}

func NewRoutingSchemeApi(app *app.App) *routingScheme {
	return &routingScheme{app: app}
}

func (api *routingScheme) Create(ctx context.Context, in *engine.RoutingScheme) (*engine.RoutingScheme, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	permission := session.GetPermission(model.PERMISSION_SCOPE_ACR_ROUTING)
	if !permission.CanRead() {
		return nil, api.app.MakePermissionError(session, permission, model.PERMISSION_ACCESS_READ)
	}

	scheme := &model.RoutingScheme{
		DomainRecord: model.DomainRecord{
			Id:        0,
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
		Name:        in.Name,
		Type:        int8(in.Type),
		Debug:       in.Debug,
		Scheme:      []byte("{}"),
		Payload:     []byte("{}"),
		Description: in.Description, //TODO
	}

	if err = scheme.IsValid(); err != nil {
		return nil, err
	}

	if scheme, err = api.app.CreateRoutingScheme(scheme); err != nil {
		return nil, err
	} else {
		return transformRoutingScheme(scheme), nil
	}
}

func (api *routingScheme) List(ctx context.Context, in *engine.ListRequest) (*engine.ListRoutingScheme, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	permission := session.GetPermission(model.PERMISSION_SCOPE_ACR_ROUTING)
	if !permission.CanRead() {
		return nil, api.app.MakePermissionError(session, permission, model.PERMISSION_ACCESS_READ)
	}
	var list []*model.RoutingScheme

	list, err = api.app.GetRoutingSchemePage(session.Domain(in.DomainId), int(in.Page), int(in.Size))

	if err != nil {
		return nil, err
	}

	items := make([]*engine.RoutingScheme, 0, len(list))
	for _, v := range list {
		items = append(items, transformRoutingScheme(v))
	}
	return &engine.ListRoutingScheme{
		Items: items,
	}, nil
}

func (api *routingScheme) Get(ctx context.Context, in *engine.ItemRequest) (*engine.RoutingScheme, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	permission := session.GetPermission(model.PERMISSION_SCOPE_ACR_ROUTING)
	if !permission.CanRead() {
		return nil, api.app.MakePermissionError(session, permission, model.PERMISSION_ACCESS_READ)
	}

	scheme, err := api.app.GetRoutingSchemeById(session.Domain(in.DomainId), in.Id)
	if err != nil {
		return nil, err
	}
	return transformRoutingScheme(scheme), nil
}

func (api *routingScheme) Update(ctx context.Context, in *engine.RoutingScheme) (*engine.RoutingScheme, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	permission := session.GetPermission(model.PERMISSION_SCOPE_ACR_ROUTING)
	if !permission.CanRead() {
		return nil, api.app.MakePermissionError(session, permission, model.PERMISSION_ACCESS_READ)
	}

	if !permission.CanUpdate() {
		return nil, api.app.MakePermissionError(session, permission, model.PERMISSION_ACCESS_UPDATE)
	}

	var scheme *model.RoutingScheme

	scheme, err = api.app.UpdateRoutingScheme(&model.RoutingScheme{
		DomainRecord: model.DomainRecord{
			Id:        in.Id,
			DomainId:  session.Domain(in.GetDomainId()),
			UpdatedAt: model.GetMillis(),
			UpdatedBy: model.Lookup{
				Id: int(session.UserId),
			},
		},
		Name:        in.Name,
		Type:        int8(in.Type),
		Debug:       in.Debug,
		Scheme:      nil,
		Payload:     nil,
		Description: in.Description,
	})

	if err != nil {
		return nil, err
	}

	return transformRoutingScheme(scheme), nil
}

func (api *routingScheme) Remove(ctx context.Context, in *engine.ItemRequest) (*engine.RoutingScheme, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	permission := session.GetPermission(model.PERMISSION_SCOPE_ACR_ROUTING)
	if !permission.CanDelete() {
		return nil, api.app.MakePermissionError(session, permission, model.PERMISSION_ACCESS_DELETE)
	}

	var scheme *model.RoutingScheme
	scheme, err = api.app.RemoveRoutingScheme(session.Domain(in.DomainId), in.Id)
	if err != nil {
		return nil, err
	}

	return transformRoutingScheme(scheme), nil
}

func transformRoutingScheme(src *model.RoutingScheme) *engine.RoutingScheme {
	return &engine.RoutingScheme{
		Id:        src.Id,
		DomainId:  src.DomainId,
		CreatedAt: src.CreatedAt,
		CreatedBy: &engine.Lookup{
			Id:   int64(src.CreatedBy.Id),
			Name: src.CreatedBy.Name,
		},
		UpdatedAt: src.UpdatedAt,
		UpdatedBy: &engine.Lookup{
			Id:   int64(src.UpdatedBy.Id),
			Name: src.UpdatedBy.Name,
		},
		Description: src.Description,
		Name:        src.Name,
		Type:        int32(src.Type),
		Debug:       src.Debug,
		Scheme: &google_protobuf.Any{
			TypeUrl: "json",
			Value:   src.Scheme,
		},
		Payload: &google_protobuf.Any{
			TypeUrl: "json",
			Value:   src.Payload,
		},
	}
}
