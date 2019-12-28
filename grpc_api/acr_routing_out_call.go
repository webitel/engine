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

func (api *routingOutboundCall) CreateRoutingOutboundCall(ctx context.Context, in *engine.CreateRoutingOutboundCallRequest) (*engine.RoutingOutboundCall, error) {
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
		Schema: model.Lookup{
			Id: int(in.GetSchema().GetId()),
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

func (api *routingOutboundCall) SearchRoutingOutboundCall(ctx context.Context, in *engine.SearchRoutingOutboundCallRequest) (*engine.ListRoutingOutboundCall, error) {
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

func (api *routingOutboundCall) ReadRoutingOutboundCall(ctx context.Context, in *engine.ReadRoutingOutboundCallRequest) (*engine.RoutingOutboundCall, error) {
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

func (api *routingOutboundCall) UpdateRoutingOutboundCall(ctx context.Context, in *engine.UpdateRoutingOutboundCallRequest) (*engine.RoutingOutboundCall, error) {
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
		Schema: model.Lookup{
			Id: int(in.GetSchema().GetId()),
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

func (api *routingOutboundCall) PatchRoutingOutboundCall(ctx context.Context, in *engine.PatchRoutingOutboundCallRequest) (*engine.RoutingOutboundCall, error) {
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

	var routing *model.RoutingOutboundCall
	patch := &model.RoutingOutboundCallPatch{}

	//TODO
	for _, v := range in.Fields {
		switch v {
		case "name":
			patch.Name = model.NewString(in.Name)
		case "description":
			patch.Description = model.NewString(in.Description)
		case "schema":
			patch.Schema = &model.Lookup{
				Id: int(in.GetSchema().GetId()),
			}
		case "priority":
			patch.Priority = model.NewInt(int(in.GetPriority()))
		case "string":
			patch.Pattern = model.NewString(in.GetPattern())
		case "disabled":
			patch.Disabled = model.NewBool(in.GetDisabled())
		}
	}
	patch.UpdatedById = int(session.UserId)
	routing, err = api.app.PatchRoutingOutboundCall(session.Domain(in.GetDomainId()), in.GetId(), patch)

	if err != nil {
		return nil, err
	}

	return transformRoutingOutboundCall(routing), nil
}

func (api *routingOutboundCall) DeleteRoutingOutboundCall(ctx context.Context, in *engine.DeleteRoutingOutboundCallRequest) (*engine.RoutingOutboundCall, error) {
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

	if src.GetSchemaId() != nil {
		dst.Schema = &engine.Lookup{
			Id:   int64(*src.GetSchemaId()),
			Name: src.Schema.Name,
		}
	}

	return dst
}
