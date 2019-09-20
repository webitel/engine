package grpc_api

import (
	"context"
	"github.com/webitel/engine/app"
	"github.com/webitel/engine/grpc_api/engine"
	"github.com/webitel/engine/model"
)

type routingOutboundCall struct {
	app *app.App
}

func NewRoutingOutboundCallApi(app *app.App) *routingOutboundCall {
	return &routingOutboundCall{app: app}
}

func (api *routingOutboundCall) Create(ctx context.Context, in *engine.RoutingOutboundCall) (*engine.RoutingOutboundCall, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	permission := session.GetPermission(model.PERMISSION_SCOPE_ACR_ROUTING)
	if !permission.CanCreate() {
		return nil, api.app.MakePermissionError(session, permission, model.PERMISSION_ACCESS_CREATE)
	}

	routing := &model.RoutingOutboundCall{
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
		Pattern:     in.Pattern,
		Priority:    int(in.Priority),
		Scheme: model.Lookup{
			Id: int(in.GetScheme().GetId()),
		},
		Disabled: in.Disabled,
	}

	if err = routing.IsValid(); err != nil {
		return nil, err
	}

	if routing, err = api.app.CreateRoutingOutboundCall(routing); err != nil {
		return nil, err
	} else {
		return transformRoutingOutboundCall(routing), nil
	}
}

func (api *routingOutboundCall) List(ctx context.Context, in *engine.ListRequest) (*engine.ListRoutingOutboundCall, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	permission := session.GetPermission(model.PERMISSION_SCOPE_ACR_ROUTING)
	if !permission.CanRead() {
		return nil, api.app.MakePermissionError(session, permission, model.PERMISSION_ACCESS_READ)
	}
	var list []*model.RoutingOutboundCall

	list, err = api.app.GetRoutingOutboundCallPage(session.Domain(in.DomainId), int(in.Page), int(in.Size))

	if err != nil {
		return nil, err
	}

	items := make([]*engine.RoutingOutboundCall, 0, len(list))
	for _, v := range list {
		items = append(items, transformRoutingOutboundCall(v))
	}
	return &engine.ListRoutingOutboundCall{
		Items: items,
	}, nil
}

func (api *routingOutboundCall) Get(ctx context.Context, in *engine.ItemRequest) (*engine.RoutingOutboundCall, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	permission := session.GetPermission(model.PERMISSION_SCOPE_ACR_ROUTING)
	if !permission.CanRead() {
		return nil, api.app.MakePermissionError(session, permission, model.PERMISSION_ACCESS_READ)
	}

	var routing *model.RoutingOutboundCall
	routing, err = api.app.GetRoutingOutboundCallById(session.Domain(in.DomainId), in.Id)
	if err != nil {
		return nil, err
	}
	return transformRoutingOutboundCall(routing), nil
}

func (api *routingOutboundCall) Update(ctx context.Context, in *engine.RoutingOutboundCall) (*engine.RoutingOutboundCall, error) {
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

	var routing = &model.RoutingOutboundCall{
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
		Scheme: model.Lookup{
			Id: int(in.GetScheme().GetId()),
		},
		Pattern:  in.Pattern,
		Priority: int(in.Priority),
		Disabled: in.Disabled,
	}

	if err = routing.IsValid(); err != nil {
		return nil, err
	}

	routing, err = api.app.UpdateRoutingOutboundCall(routing)

	if err != nil {
		return nil, err
	}

	return transformRoutingOutboundCall(routing), nil
}

func (api *routingOutboundCall) Remove(ctx context.Context, in *engine.ItemRequest) (*engine.RoutingOutboundCall, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	permission := session.GetPermission(model.PERMISSION_SCOPE_ACR_ROUTING)
	if !permission.CanDelete() {
		return nil, api.app.MakePermissionError(session, permission, model.PERMISSION_ACCESS_DELETE)
	}

	var routing *model.RoutingOutboundCall
	routing, err = api.app.RemoveRoutingOutboundCall(session.Domain(in.DomainId), in.Id)
	if err != nil {
		return nil, err
	}

	return transformRoutingOutboundCall(routing), nil
}

func transformRoutingOutboundCall(src *model.RoutingOutboundCall) *engine.RoutingOutboundCall {
	dst := &engine.RoutingOutboundCall{
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
		Pattern:     src.Pattern,
		Priority:    int32(src.Priority),
		Disabled:    src.Disabled,
	}

	if src.GetSchemeId() != nil {
		dst.Scheme = &engine.Lookup{
			Id:   int64(*src.GetSchemeId()),
			Name: src.Scheme.Name,
		}
	}

	return dst
}
