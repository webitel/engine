package grpc_api

import (
	"context"
	"github.com/webitel/engine/model"
	"github.com/webitel/protos/engine"
)

type schemaVariable struct {
	*API
	engine.UnsafeSchemaVariablesServiceServer
}

func NewSchemeVariableApi(api *API) *schemaVariable {
	return &schemaVariable{API: api}
}

func (api *schemaVariable) CreateSchemaVariable(ctx context.Context, in *engine.CreateSchemaVariableRequest) (*engine.SchemaVariable, error) {
	session, err := api.ctrl.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	s := &model.SchemeVariable{
		Name:    in.GetName(),
		Encrypt: in.GetEncrypt(),
		Value:   MarshalJsonpb(in.Value),
	}

	s, err = api.ctrl.CreateSchemeVariable(ctx, session, s)
	if err != nil {
		return nil, err
	}
	return toSchemeVariable(s), nil
}

func (api *schemaVariable) SearchSchemaVariable(ctx context.Context, in *engine.SearchSchemaVariableRequest) (*engine.ListSchemaVariable, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	var list []*model.SchemeVariable
	var endList bool
	req := &model.SearchSchemeVariable{
		ListRequest: model.ExtractSearchOptions(in),
	}

	list, endList, err = api.ctrl.SearchSchemeVariable(ctx, session, req)

	if err != nil {
		return nil, err
	}

	items := make([]*engine.SchemaVariable, 0, len(list))
	for _, v := range list {
		items = append(items, toSchemeVariable(v))
	}
	return &engine.ListSchemaVariable{
		Next:  !endList,
		Items: items,
	}, nil
}

func (api *schemaVariable) ReadSchemaVariable(ctx context.Context, in *engine.ReadSchemaVariableRequest) (*engine.SchemaVariable, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	var s *model.SchemeVariable
	s, err = api.ctrl.GetSchemeVariable(ctx, session, in.Id)

	if err != nil {
		return nil, err
	}

	return toSchemeVariable(s), nil
}

func (api *schemaVariable) UpdateSchemaVariable(ctx context.Context, in *engine.UpdateSchemaVariableRequest) (*engine.SchemaVariable, error) {
	session, err := api.ctrl.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	s := &model.SchemeVariable{
		Name:  in.GetName(),
		Value: MarshalJsonpb(in.Value),
	}

	s, err = api.ctrl.UpdateSchemaVariable(ctx, session, in.GetId(), s)
	if err != nil {
		return nil, err
	}
	return toSchemeVariable(s), nil
}

func (api *schemaVariable) PatchSchemaVariable(ctx context.Context, in *engine.PatchSchemaVariableRequest) (*engine.SchemaVariable, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	var sv *model.SchemeVariable
	patch := &model.PatchSchemeVariable{}

	//TODO
	for _, v := range in.Fields {
		switch v {
		case "name":
			patch.Name = &in.Name
		case "value":
			patch.Value = MarshalJsonpb(in.Value)
		case "encrypt":
			patch.Encrypt = &in.Encrypt
		}
	}

	if sv, err = api.ctrl.PatchSchemaVariable(ctx, session, in.Id, patch); err != nil {
		return nil, err
	}

	return toSchemeVariable(sv), nil
}

func (api *schemaVariable) DeleteSchemaVariable(ctx context.Context, in *engine.DeleteSchemaVariableRequest) (*engine.SchemaVariable, error) {
	session, err := api.ctrl.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	var s *model.SchemeVariable
	s, err = api.ctrl.DeleteSchemaVariable(ctx, session, in.Id)
	if err != nil {
		return nil, err
	}
	return toSchemeVariable(s), nil
}

func toSchemeVariable(src *model.SchemeVariable) *engine.SchemaVariable {
	res := &engine.SchemaVariable{
		Id:      src.Id,
		Name:    src.Name,
		Encrypt: src.Encrypt,
		Value:   UnmarshalJsonpb(src.Value),
	}

	// TODO
	if src.Encrypt {
		res.Value = nil
	}

	return res
}
