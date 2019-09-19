package grpc_api

import (
	"context"
	"github.com/webitel/engine/app"
	"github.com/webitel/engine/grpc_api/engine"
	"github.com/webitel/engine/model"
)

type routingInboundCall struct {
	app *app.App
}

func NewRoutingInboundCallApi(app *app.App) *routingInboundCall {
	return &routingInboundCall{app: app}
}

func (api *routingInboundCall) Create(ctx context.Context, in *engine.RoutingInboundCall) (*engine.RoutingInboundCall, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	permission := session.GetPermission(model.PERMISSION_SCOPE_ACR_ROUTING)
	if !permission.CanCreate() {
		return nil, api.app.MakePermissionError(session, permission, model.PERMISSION_ACCESS_CREATE)
	}

	routing := &model.RoutingInboundCall{
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
		Name:        in.Name,
		Description: in.Description,
		Numbers:     in.Numbers,
		Host:        in.Host,
		Timezone: model.Lookup{
			Id: int(in.GetTimezone().GetId()),
		},
		StartScheme: model.Lookup{
			Id: int(in.GetStartScheme().GetId()),
		},
		Debug:    in.Debug,
		Disabled: in.Disabled,
	}

	if in.StopScheme != nil {
		routing.StopScheme = &model.Lookup{
			Id: int(in.StopScheme.Id),
		}
	}

	if err = routing.IsValid(); err != nil {
		return nil, err
	}

	if routing, err = api.app.CreateRoutingInboundCall(routing); err != nil {
		return nil, err
	} else {
		return transformRoutingInboundCall(routing), nil
	}
}

func (api *routingInboundCall) List(ctx context.Context, in *engine.ListRequest) (*engine.ListRoutingInboundCall, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	permission := session.GetPermission(model.PERMISSION_SCOPE_ACR_ROUTING)
	if !permission.CanRead() {
		return nil, api.app.MakePermissionError(session, permission, model.PERMISSION_ACCESS_READ)
	}
	var list []*model.RoutingInboundCall

	list, err = api.app.GetRoutingInboundCallPage(session.Domain(in.DomainId), int(in.Page), int(in.Size))

	if err != nil {
		return nil, err
	}

	items := make([]*engine.RoutingInboundCall, 0, len(list))
	for _, v := range list {
		items = append(items, transformRoutingInboundCall(v))
	}
	return &engine.ListRoutingInboundCall{
		Items: items,
	}, nil
}

func (api *routingInboundCall) Get(ctx context.Context, in *engine.ItemRequest) (*engine.RoutingInboundCall, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	permission := session.GetPermission(model.PERMISSION_SCOPE_ACR_ROUTING)
	if !permission.CanRead() {
		return nil, api.app.MakePermissionError(session, permission, model.PERMISSION_ACCESS_READ)
	}

	var routing *model.RoutingInboundCall
	routing, err = api.app.GetRoutingInboundCallById(session.Domain(in.DomainId), in.Id)
	if err != nil {
		return nil, err
	}
	return transformRoutingInboundCall(routing), nil
}

func (api *routingInboundCall) Update(ctx context.Context, in *engine.RoutingInboundCall) (*engine.RoutingInboundCall, error) {
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

	var routing = &model.RoutingInboundCall{
		DomainRecord: model.DomainRecord{
			Id:        in.Id,
			DomainId:  session.Domain(in.GetDomainId()),
			UpdatedAt: model.GetMillis(),
			UpdatedBy: model.Lookup{
				Id: int(session.UserId),
			},
		},
		Name:        in.Name,
		Description: in.Description,
		Numbers:     in.GetNumbers(),
		StartScheme: model.Lookup{
			Id: int(in.GetStartScheme().GetId()),
		},
		Timezone: model.Lookup{
			Id: int(in.GetTimezone().GetId()),
		},
		Host:     in.Host,
		Debug:    in.Debug,
		Disabled: in.Disabled,
	}

	if in.StopScheme != nil {
		routing.StopScheme = &model.Lookup{
			Id: int(in.StopScheme.Id),
		}
	}

	if err = routing.IsValid(); err != nil {
		return nil, err
	}

	routing, err = api.app.UpdateRoutingInboundCall(routing)

	if err != nil {
		return nil, err
	}

	return transformRoutingInboundCall(routing), nil
}

func (api *routingInboundCall) Remove(ctx context.Context, in *engine.ItemRequest) (*engine.RoutingInboundCall, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	permission := session.GetPermission(model.PERMISSION_SCOPE_ACR_ROUTING)
	if !permission.CanDelete() {
		return nil, api.app.MakePermissionError(session, permission, model.PERMISSION_ACCESS_DELETE)
	}

	var routing *model.RoutingInboundCall
	routing, err = api.app.RemoveRoutingInboundCall(session.Domain(in.DomainId), in.Id)
	if err != nil {
		return nil, err
	}

	return transformRoutingInboundCall(routing), nil
}

func transformRoutingInboundCall(src *model.RoutingInboundCall) *engine.RoutingInboundCall {
	dst := &engine.RoutingInboundCall{
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
		Numbers:     src.Numbers,
		Host:        src.Host,
		Timezone: &engine.Lookup{
			Id:   int64(src.Timezone.Id),
			Name: src.Timezone.Name,
		},
		Debug:    src.Debug,
		Disabled: src.Disabled,
	}

	if src.GetStopSchemeId() != nil {
		dst.StopScheme = &engine.Lookup{
			Id:   int64(src.StopScheme.Id),
			Name: src.StopScheme.Name,
		}
	}

	if src.GetStartSchemeId() != nil {
		dst.StartScheme = &engine.Lookup{
			Id:   int64(*src.GetStartSchemeId()),
			Name: src.StartScheme.Name,
		}
	}

	return dst
}
