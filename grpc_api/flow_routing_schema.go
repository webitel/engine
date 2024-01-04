package grpc_api

import (
	"context"
	"github.com/webitel/engine/model"
	"github.com/webitel/protos/engine"
	"strings"
)

type routingSchema struct {
	*API
	engine.UnsafeRoutingSchemaServiceServer
}

func NewRoutingSchemaApi(api *API) *routingSchema {
	return &routingSchema{API: api}
}

func (api *routingSchema) CreateRoutingSchema(ctx context.Context, in *engine.CreateRoutingSchemaRequest) (*engine.RoutingSchema, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	scheme := &model.RoutingSchema{
		Name:        in.Name,
		Type:        in.GetType().String(),
		Debug:       in.Debug,
		Schema:      MarshalJsonpb(in.Schema),
		Payload:     MarshalJsonpb(in.Payload),
		Description: in.Description,
		Editor:      in.Editor,
		Tags:        flowTagsToStrings(in.GetTags()),
	}

	if scheme, err = api.ctrl.CreateRoutingSchema(ctx, session, scheme); err != nil {
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
		Type:   transformTypes(in.GetType()),
		Tags:   in.GetTags(),
	}

	list, endList, err = api.ctrl.SearchSchema(ctx, session, req)

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

	scheme, err := api.ctrl.GetSchema(ctx, session, in.Id)
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

	scheme := &model.RoutingSchema{
		DomainRecord: model.DomainRecord{
			Id: in.Id,
		},
		Name:        in.Name,
		Type:        in.GetType().String(),
		Debug:       in.Debug,
		Schema:      MarshalJsonpb(in.Schema),
		Payload:     MarshalJsonpb(in.Payload),
		Description: in.Description,
		Editor:      in.Editor,
		Tags:        flowTagsToStrings(in.GetTags()),
	}

	scheme, err = api.ctrl.UpdateSchema(ctx, session, scheme)

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

	var scheme *model.RoutingSchema
	patch := &model.RoutingSchemaPath{}

	//TODO
	for _, v := range in.Fields {
		switch v {
		case "name":
			patch.Name = &in.Name
		case "type":
			patch.Type = model.NewString(in.GetType().String())
		case "description":
			patch.Description = &in.Description
		case "tags":
			patch.Tags = flowTagsToStrings(in.Tags)
			if patch.Tags == nil {
				patch.Tags = make([]string, 0, 0)
			}
		case "debug":
			patch.Debug = &in.Debug
		case "editor":
			patch.Editor = &in.Editor
		default:
			if patch.Schema == nil && strings.HasPrefix(v, "schema") {
				patch.Schema = MarshalJsonpb(in.Schema)
			} else if patch.Payload == nil && strings.HasPrefix(v, "payload") {
				patch.Payload = MarshalJsonpb(in.Payload)
			}
		}
	}

	scheme, err = api.ctrl.PatchSchema(ctx, session, in.GetId(), patch)

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

	var scheme *model.RoutingSchema
	scheme, err = api.ctrl.DeleteSchema(ctx, session, in.Id)
	if err != nil {
		return nil, err
	}

	return transformRoutingSchema(scheme), nil
}

func (api *routingSchema) SearchRoutingSchemaTags(ctx context.Context, in *engine.SearchRoutingSchemaTagsRequest) (*engine.ListRoutingSchemaTags, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	var list []*model.RoutingSchemaTag
	var endList bool

	req := &model.SearchRoutingSchemaTag{
		ListRequest: model.ListRequest{
			Q:       in.GetQ(),
			Page:    int(in.GetPage()),
			PerPage: int(in.GetSize()),
			Fields:  in.Fields,
			Sort:    in.Sort,
		},
		Type: transformTypes(in.GetType()),
	}

	list, endList, err = api.ctrl.SearchSchemaTags(ctx, session, req)

	if err != nil {
		return nil, err
	}

	items := make([]*engine.RoutingSchemaTag, 0, len(list))
	for _, v := range list {
		items = append(items, &engine.RoutingSchemaTag{
			Name: v.Name,
			//Count: v.Count,
		})
	}
	return &engine.ListRoutingSchemaTags{
		Next:  !endList,
		Items: items,
	}, nil
}

func transformRoutingSchema(src *model.RoutingSchema) *engine.RoutingSchema {
	s := &engine.RoutingSchema{
		Id:          src.Id,
		CreatedAt:   src.CreatedAt,
		CreatedBy:   GetProtoLookup(src.CreatedBy),
		UpdatedAt:   src.UpdatedAt,
		UpdatedBy:   GetProtoLookup(src.UpdatedBy),
		Description: src.Description,
		Name:        src.Name,
		Debug:       src.Debug,
		Editor:      src.Editor,
		Schema:      UnmarshalJsonpb(src.Schema),
		Payload:     UnmarshalJsonpb(src.Payload),
	}

	if src.Type != "" {
		s.Type = transformTypeToEngine(src.Type)
	}

	if src.Tags != nil {
		s.Tags = make([]*engine.SchemaTag, 0, len(src.Tags))
		for _, v := range src.Tags {
			s.Tags = append(s.Tags, &engine.SchemaTag{
				Name: v,
			})
		}
	}

	return s
}

func transformTypeToEngine(name string) engine.RoutingSchemaType {
	if v, ok := engine.RoutingSchemaType_value[name]; ok {
		return engine.RoutingSchemaType(v)
	}

	return engine.RoutingSchemaType_voice
}

func transformTypes(tps []engine.RoutingSchemaType) []string {
	l := len(tps)
	if l == 0 {
		return nil
	}

	res := make([]string, 0, l)
	for _, v := range tps {
		res = append(res, v.String())
	}

	return res
}

func flowTagsToStrings(tags []*engine.SchemaTag) []string {
	l := len(tags)
	if l == 0 {
		return nil
	}

	res := make([]string, 0, l)

	for _, v := range tags {
		res = append(res, v.Name)
	}

	return res
}
