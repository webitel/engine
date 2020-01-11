package grpc_api

import (
	"context"
	"github.com/webitel/engine/app"
	"github.com/webitel/engine/grpc_api/engine"
	"github.com/webitel/engine/model"
)

type routingSchema struct {
	app *app.App
}

func NewRoutingSchemaApi(app *app.App) *routingSchema {
	return &routingSchema{app: app}
}

func (api *routingSchema) CreateRoutingSchema(ctx context.Context, in *engine.CreateRoutingSchemaRequest) (*engine.RoutingSchema, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	permission := session.GetPermission(model.PERMISSION_SCOPE_ACR_ROUTING)
	if !permission.CanRead() {
		return nil, api.app.MakePermissionError(session, permission, model.PERMISSION_ACCESS_READ)
	}

	if !permission.CanCreate() {
		return nil, api.app.MakePermissionError(session, permission, model.PERMISSION_ACCESS_CREATE)
	}

	scheme := &model.RoutingSchema{
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
		Schema:      MarshalJsonpb(in.Schema),
		Payload:     MarshalJsonpb(in.Payload),
		Description: in.Description,
	}

	if err = scheme.IsValid(); err != nil {
		return nil, err
	}

	if scheme, err = api.app.CreateRoutingSchema(scheme); err != nil {
		return nil, err
	} else {
		return transformRoutingSchema(scheme), nil
	}
}

func (api *routingSchema) SearchRoutingSchema(ctx context.Context, in *engine.SearchRoutingSchemaRequest) (*engine.ListRoutingSchema, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	permission := session.GetPermission(model.PERMISSION_SCOPE_ACR_ROUTING)
	if !permission.CanRead() {
		return nil, api.app.MakePermissionError(session, permission, model.PERMISSION_ACCESS_READ)
	}
	var list []*model.RoutingSchema

	list, err = api.app.GetRoutingSchemaPage(session.Domain(in.DomainId), int(in.Page), int(in.Size))

	if err != nil {
		return nil, err
	}

	items := make([]*engine.RoutingSchema, 0, len(list))
	for _, v := range list {
		items = append(items, transformRoutingSchema(v))
	}
	return &engine.ListRoutingSchema{
		Items: items,
	}, nil
}

func (api *routingSchema) ReadRoutingSchema(ctx context.Context, in *engine.ReadRoutingSchemaRequest) (*engine.RoutingSchema, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	permission := session.GetPermission(model.PERMISSION_SCOPE_ACR_ROUTING)
	if !permission.CanRead() {
		return nil, api.app.MakePermissionError(session, permission, model.PERMISSION_ACCESS_READ)
	}

	scheme, err := api.app.GetRoutingSchemaById(session.Domain(in.DomainId), in.Id)
	if err != nil {
		return nil, err
	}
	return transformRoutingSchema(scheme), nil
}

func (api *routingSchema) UpdateRoutingSchema(ctx context.Context, in *engine.UpdateRoutingSchemaRequest) (*engine.RoutingSchema, error) {
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

	scheme := &model.RoutingSchema{
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
		Schema:      MarshalJsonpb(in.Schema),
		Payload:     MarshalJsonpb(in.Payload),
		Description: in.Description,
	}

	if err = scheme.IsValid(); err != nil {
		return nil, err
	}

	scheme, err = api.app.UpdateRoutingSchema(scheme)

	if err != nil {
		return nil, err
	}

	return transformRoutingSchema(scheme), nil
}

func (api *routingSchema) PatchRoutingSchema(ctx context.Context, in *engine.PatchRoutingSchemaRequest) (*engine.RoutingSchema, error) {
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

	var scheme *model.RoutingSchema
	patch := &model.RoutingSchemaPath{}

	//TODO
	for _, v := range in.Fields {
		switch v {
		case "name":
			patch.Name = model.NewString(in.Name)
		case "type":
			patch.Type = model.NewInt8(int8(in.Type))
		case "schema":
			patch.Schema = MarshalJsonpb(in.Schema)
		case "payload":
			patch.Payload = MarshalJsonpb(in.Payload)
		case "description":
			patch.Description = model.NewString(in.Description)
		case "debug":
			patch.Debug = model.NewBool(in.Debug)
		}
	}
	patch.UpdatedById = int(session.UserId)
	scheme, err = api.app.PatchRoutingSchema(session.Domain(in.GetDomainId()), in.GetId(), patch)

	if err != nil {
		return nil, err
	}

	return transformRoutingSchema(scheme), nil
}

func (api *routingSchema) DeleteRoutingSchema(ctx context.Context, in *engine.DeleteRoutingSchemaRequest) (*engine.RoutingSchema, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	permission := session.GetPermission(model.PERMISSION_SCOPE_ACR_ROUTING)
	if !permission.CanDelete() {
		return nil, api.app.MakePermissionError(session, permission, model.PERMISSION_ACCESS_DELETE)
	}

	var scheme *model.RoutingSchema
	scheme, err = api.app.RemoveRoutingSchema(session.Domain(in.DomainId), in.Id)
	if err != nil {
		return nil, err
	}

	return transformRoutingSchema(scheme), nil
}

func transformRoutingSchema(src *model.RoutingSchema) *engine.RoutingSchema {
	return &engine.RoutingSchema{
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
		Schema:      UnmarshalJsonpb(src.Schema),
		Payload:     UnmarshalJsonpb(src.Payload),
	}
}
