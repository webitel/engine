package grpc_api

import (
	"context"
	"github.com/webitel/engine/app"
	"github.com/webitel/engine/auth_manager"
	"github.com/webitel/engine/model"
	"github.com/webitel/protos/engine"
	"strings"
)

type routingSchema struct {
	app *app.App
	engine.UnsafeRoutingSchemaServiceServer
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
		return nil, api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if !permission.CanCreate() {
		return nil, api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_CREATE)
	}

	scheme := &model.RoutingSchema{
		DomainRecord: model.DomainRecord{
			Id:        0,
			DomainId:  session.Domain(0),
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
		Type:        in.Type,
		Debug:       in.Debug,
		Schema:      MarshalJsonpb(in.Schema),
		Payload:     MarshalJsonpb(in.Payload),
		Description: in.Description,
		Editor:      in.Editor,
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
		return nil, api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}
	var list []*model.RoutingSchema
	var endList bool

	req := &model.SearchRoutingSchema{
		ListRequest: model.ListRequest{
			Q:       in.GetQ(),
			Page:    int(in.GetPage()),
			PerPage: int(in.GetSize()),
			Fields:  in.Fields,
			Sort:    in.Sort,
		},
		Ids:    in.Id,
		Name:   GetStringPointer(in.Name),
		Editor: in.Editor,
		Type:   in.Type,
	}

	list, endList, err = api.app.GetRoutingSchemaPage(session.Domain(0), req)

	if err != nil {
		return nil, err
	}

	items := make([]*engine.RoutingSchema, 0, len(list))
	for _, v := range list {
		items = append(items, transformRoutingSchema(v))
	}
	return &engine.ListRoutingSchema{
		Next:  !endList,
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
		return nil, api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
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
		return nil, api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if !permission.CanUpdate() {
		return nil, api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
	}

	scheme := &model.RoutingSchema{
		DomainRecord: model.DomainRecord{
			Id:        in.Id,
			DomainId:  session.Domain(0),
			UpdatedAt: model.GetMillis(),
			UpdatedBy: &model.Lookup{
				Id: int(session.UserId),
			},
		},
		Name:        in.Name,
		Type:        in.Type,
		Debug:       in.Debug,
		Schema:      MarshalJsonpb(in.Schema),
		Payload:     MarshalJsonpb(in.Payload),
		Description: in.Description,
		Editor:      in.Editor,
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
		return nil, api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if !permission.CanUpdate() {
		return nil, api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
	}

	var scheme *model.RoutingSchema
	patch := &model.RoutingSchemaPath{}

	//TODO
	for _, v := range in.Fields {
		switch v {
		case "name":
			patch.Name = model.NewString(in.Name)
		case "type":
			patch.Type = &in.Type
		case "description":
			patch.Description = model.NewString(in.Description)
		case "debug":
			patch.Debug = model.NewBool(in.Debug)
		case "editor":
			patch.Editor = model.NewBool(in.Editor)
		default:
			if patch.Schema == nil && strings.HasPrefix(v, "schema") {
				patch.Schema = MarshalJsonpb(in.Schema)
			} else if patch.Payload == nil && strings.HasPrefix(v, "payload") {
				patch.Payload = MarshalJsonpb(in.Payload)
			}
		}
	}
	patch.UpdatedById = int(session.UserId)
	scheme, err = api.app.PatchRoutingSchema(session.Domain(0), in.GetId(), patch)

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
		return nil, api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_DELETE)
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
		Id:          src.Id,
		CreatedAt:   src.CreatedAt,
		CreatedBy:   GetProtoLookup(src.CreatedBy),
		UpdatedAt:   src.UpdatedAt,
		UpdatedBy:   GetProtoLookup(src.UpdatedBy),
		Description: src.Description,
		Name:        src.Name,
		Type:        src.Type,
		Debug:       src.Debug,
		Editor:      src.Editor,
		Schema:      UnmarshalJsonpb(src.Schema),
		Payload:     UnmarshalJsonpb(src.Payload),
	}
}
