package grpc_api

import (
	"context"
	"github.com/webitel/engine/gen/engine"
	"github.com/webitel/engine/model"
	"github.com/webitel/engine/pkg/wbt/auth_manager"
)

type routingOutboundCall struct {
	*API
	engine.UnsafeRoutingOutboundCallServiceServer
}

func NewRoutingOutboundCallApi(api *API) *routingOutboundCall {
	return &routingOutboundCall{
		API: api,
	}
}

func (api *routingOutboundCall) CreateRoutingOutboundCall(ctx context.Context, in *engine.CreateRoutingOutboundCallRequest) (*engine.RoutingOutboundCall, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	routing := &model.RoutingOutboundCall{
		DomainRecord: model.DomainRecord{
			DomainId:  session.Domain(in.GetDomainId()),
			CreatedAt: model.GetMillis(),
			CreatedBy: &model.Lookup{
				Id: int(session.UserId),
			},
			UpdatedAt: model.GetMillis(),
			UpdatedBy: &model.Lookup{
				Id: int(session.UserId),
			},
		},
		Name:        in.Name,
		Description: in.Description,
		Pattern:     in.Pattern,
		Schema: model.Lookup{
			Id: int(in.GetSchema().GetId()),
		},
		Disabled: in.Disabled,
	}

	if routing, err = api.ctrl.CreateRoutingOutboundCall(ctx, session, routing); err != nil {
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

	var list []*model.RoutingOutboundCall
	var isEndList bool
	req := &model.SearchRoutingOutboundCall{
		ListRequest: model.ListRequest{
			Q:       in.GetQ(),
			Page:    int(in.GetPage()),
			PerPage: int(in.GetSize()),
			Fields:  in.Fields,
			Sort:    in.Sort,
		},
		Ids:         in.Id,
		Name:        GetStringPointer(in.Name),
		SchemaIds:   in.SchemaId,
		Pattern:     GetStringPointer(in.Pattern),
		Description: GetStringPointer(in.Description),
	}

	list, isEndList, err = api.ctrl.SearchRoutingOutboundCall(ctx, session, req)

	if err != nil {
		return nil, err
	}

	items := make([]*engine.RoutingOutboundCallCompact, 0, len(list))
	for _, v := range list {
		items = append(items, toRoutingOutboundCallCompact(v))
	}
	return &engine.ListRoutingOutboundCall{
		Next:  !isEndList,
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
		return nil, api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	var routing *model.RoutingOutboundCall
	routing, err = api.ctrl.GetRoutingOutboundCall(ctx, session, in.DomainId, in.Id)
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

	var routing = &model.RoutingOutboundCall{
		DomainRecord: model.DomainRecord{
			Id:        in.Id,
			DomainId:  session.Domain(in.GetDomainId()),
			UpdatedAt: model.GetMillis(),
			UpdatedBy: &model.Lookup{
				Id: int(session.UserId),
			},
		},
		Name:        in.Name,
		Description: in.Description,
		Schema: model.Lookup{
			Id: int(in.GetSchema().GetId()),
		},
		Pattern:  in.Pattern,
		Disabled: in.Disabled,
	}

	routing, err = api.ctrl.UpdateRoutingOutboundCall(ctx, session, routing)

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

	var routing *model.RoutingOutboundCall
	patch := &model.RoutingOutboundCallPatch{}

	//TODO
	for _, v := range in.Fields {
		switch v {
		case "name":
			patch.Name = model.NewString(in.Name)
		case "description":
			patch.Description = model.NewString(in.Description)
		case "schema.id":
			patch.Schema = GetLookup(in.Schema)
		case "pattern":
			patch.Pattern = model.NewString(in.GetPattern())
		case "disabled":
			patch.Disabled = model.NewBool(in.GetDisabled())
		}
	}

	routing, err = api.ctrl.PatchRoutingOutboundCall(ctx, session, in.GetDomainId(), in.GetId(), patch)

	if err != nil {
		return nil, err
	}

	return transformRoutingOutboundCall(routing), nil
}

func (api *routingOutboundCall) MovePositionRoutingOutboundCall(ctx context.Context,
	in *engine.MovePositionRoutingOutboundCallRequest) (*engine.MovePositionRoutingOutboundCallResponse, error) {

	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	err = api.ctrl.ChangePositionOutboundCall(ctx, session, in.GetDomainId(), in.GetFromId(), in.GetToId())

	if err != nil {
		return nil, err
	}

	return &engine.MovePositionRoutingOutboundCallResponse{Success: true}, nil
}

func (api *routingOutboundCall) DeleteRoutingOutboundCall(ctx context.Context, in *engine.DeleteRoutingOutboundCallRequest) (*engine.RoutingOutboundCall, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	var routing *model.RoutingOutboundCall
	routing, err = api.ctrl.DeleteRoutingOutboundCall(ctx, session, in.DomainId, in.Id)
	if err != nil {
		return nil, err
	}

	return transformRoutingOutboundCall(routing), nil
}

func transformRoutingOutboundCall(src *model.RoutingOutboundCall) *engine.RoutingOutboundCall {
	dst := &engine.RoutingOutboundCall{
		Id:          src.Id,
		DomainId:    src.DomainId,
		CreatedAt:   src.CreatedAt,
		CreatedBy:   GetProtoLookup(src.CreatedBy),
		UpdatedAt:   src.UpdatedAt,
		UpdatedBy:   GetProtoLookup(src.UpdatedBy),
		Description: src.Description,
		Name:        src.Name,
		Pattern:     src.Pattern,
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

func toRoutingOutboundCallCompact(src *model.RoutingOutboundCall) *engine.RoutingOutboundCallCompact {
	dst := &engine.RoutingOutboundCallCompact{
		Id:          src.Id,
		DomainId:    src.DomainId,
		Description: src.Description,
		Name:        src.Name,
		Pattern:     src.Pattern,
		Disabled:    src.Disabled,
		Position:    int32(src.Position),
	}

	if src.GetSchemaId() != nil {
		dst.Schema = &engine.Lookup{
			Id:   int64(*src.GetSchemaId()),
			Name: src.Schema.Name,
		}
	}

	return dst
}
