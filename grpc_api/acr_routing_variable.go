package grpc_api

import (
	"context"
	"github.com/webitel/engine/app"
	"github.com/webitel/engine/grpc_api/engine"
	"github.com/webitel/engine/model"
)

type routingVariable struct {
	app *app.App
}

func NewRoutingVariableApi(app *app.App) *routingVariable {
	return &routingVariable{app: app}
}

func (api *routingVariable) CreateRoutingVariable(ctx context.Context, in *engine.CreateRoutingVariableRequest) (*engine.RoutingVariable, error) {
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

	variable := &model.RoutingVariable{
		DomainId: session.Domain(in.GetDomainId()),
		Key:      in.Key,
		Value:    in.Value,
	}

	if err = variable.IsValid(); err != nil {
		return nil, err
	}

	if variable, err = api.app.CreateRoutingVariable(variable); err != nil {
		return nil, err
	}
	return transformRoutingVariable(variable), nil
}

func (api *routingVariable) SearchRoutingVariable(ctx context.Context, in *engine.SearchRoutingVariableRequest) (*engine.ListRoutingVariable, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	permission := session.GetPermission(model.PERMISSION_SCOPE_ACR_ROUTING)
	if !permission.CanRead() {
		return nil, api.app.MakePermissionError(session, permission, model.PERMISSION_ACCESS_READ)
	}
	var list []*model.RoutingVariable

	list, err = api.app.GetRoutingVariablesPage(session.Domain(in.DomainId), int(in.Page), int(in.Size))

	if err != nil {
		return nil, err
	}

	items := make([]*engine.RoutingVariable, 0, len(list))
	for _, v := range list {
		items = append(items, transformRoutingVariable(v))
	}
	return &engine.ListRoutingVariable{
		Items: items,
	}, nil
}

func (api *routingVariable) ReadRoutingVariable(ctx context.Context, in *engine.ReadRoutingVariableRequest) (*engine.RoutingVariable, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	permission := session.GetPermission(model.PERMISSION_SCOPE_ACR_ROUTING)
	if !permission.CanRead() {
		return nil, api.app.MakePermissionError(session, permission, model.PERMISSION_ACCESS_READ)
	}

	var variable *model.RoutingVariable
	variable, err = api.app.GetRoutingVariableById(session.Domain(in.DomainId), in.Id)
	if err != nil {
		return nil, err
	}
	return transformRoutingVariable(variable), nil
}

func (api *routingVariable) UpdateRoutingVariable(ctx context.Context, in *engine.UpdateRoutingVariableRequest) (*engine.RoutingVariable, error) {
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

	var variable *model.RoutingVariable

	variable, err = api.app.UpdateRoutingVariable(&model.RoutingVariable{
		Id:       in.Id,
		DomainId: session.Domain(in.DomainId),
		Key:      in.Key,
		Value:    in.Value,
	})

	if err != nil {
		return nil, err
	}

	return transformRoutingVariable(variable), nil
}

func (api *routingVariable) DeleteRoutingVariable(ctx context.Context, in *engine.DeleteRoutingVariableRequest) (*engine.RoutingVariable, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	permission := session.GetPermission(model.PERMISSION_SCOPE_ACR_ROUTING)
	if !permission.CanDelete() {
		return nil, api.app.MakePermissionError(session, permission, model.PERMISSION_ACCESS_DELETE)
	}

	var variable *model.RoutingVariable
	variable, err = api.app.RemoveRoutingVariable(session.Domain(in.DomainId), in.Id)
	if err != nil {
		return nil, err
	}

	return transformRoutingVariable(variable), nil
}

func transformRoutingVariable(src *model.RoutingVariable) *engine.RoutingVariable {
	return &engine.RoutingVariable{
		Id:       src.Id,
		DomainId: src.DomainId,
		Key:      src.Key,
		Value:    src.Value,
	}
}
