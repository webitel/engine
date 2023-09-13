package controller

import (
	"context"

	"github.com/webitel/engine/auth_manager"
	"github.com/webitel/engine/model"
)

func (c *Controller) CreateRoutingSchema(ctx context.Context, session *auth_manager.Session, schema *model.RoutingSchema) (*model.RoutingSchema, model.AppError) {
	var err model.AppError
	permission := session.GetPermission(model.PERMISSION_SCOPE_SCHEMA)
	if !permission.CanCreate() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_CREATE)
	}

	schema.DomainRecord = model.DomainRecord{
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
	}

	if err := schema.IsValid(); err != nil {
		return nil, err
	}

	schema, err = c.app.CreateRoutingSchema(ctx, schema)
	if err != nil {
		return nil, err
	}

	return schema, nil
}

func (c *Controller) SearchSchema(ctx context.Context, session *auth_manager.Session, search *model.SearchRoutingSchema) ([]*model.RoutingSchema, bool, model.AppError) {
	permission := session.GetPermission(model.PERMISSION_SCOPE_SCHEMA)
	if !permission.CanRead() {
		return nil, false, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	return c.app.GetRoutingSchemaPage(ctx, session.Domain(search.DomainId), search)
}

func (c *Controller) GetSchema(ctx context.Context, session *auth_manager.Session, id int64) (*model.RoutingSchema, model.AppError) {
	permission := session.GetPermission(model.PERMISSION_SCOPE_SCHEMA)
	if !permission.CanRead() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	return c.app.GetRoutingSchemaById(ctx, session.Domain(0), id)
}

func (c *Controller) UpdateSchema(ctx context.Context, session *auth_manager.Session, schema *model.RoutingSchema) (*model.RoutingSchema, model.AppError) {
	var err model.AppError
	permission := session.GetPermission(model.PERMISSION_SCOPE_SCHEMA)
	if !permission.CanRead() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if !permission.CanUpdate() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
	}

	if err = schema.IsValid(); err != nil {
		return nil, err
	}

	schema.DomainRecord.DomainId = session.DomainId
	schema.DomainRecord.UpdatedAt = model.GetMillis()
	schema.DomainRecord.UpdatedBy = &model.Lookup{
		Id: int(session.UserId),
	}

	schema, err = c.app.UpdateRoutingSchema(ctx, schema)
	if err != nil {
		return nil, err
	}

	c.app.AuditUpdate(ctx, session, model.PERMISSION_SCOPE_SCHEMA, schema.Id, schema)

	return schema, nil
}

func (c *Controller) PatchSchema(ctx context.Context, session *auth_manager.Session, id int64, patch *model.RoutingSchemaPath) (*model.RoutingSchema, model.AppError) {
	permission := session.GetPermission(model.PERMISSION_SCOPE_SCHEMA)
	if !permission.CanRead() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if !permission.CanUpdate() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
	}

	patch.UpdatedById = int(session.UserId)

	schema, err := c.app.PatchRoutingSchema(ctx, session.DomainId, id, patch)
	if err != nil {
		return nil, err
	}

	c.app.AuditUpdate(ctx, session, model.PERMISSION_SCOPE_SCHEMA, schema.Id, schema)

	return schema, nil
}

func (c *Controller) DeleteSchema(ctx context.Context, session *auth_manager.Session, id int64) (*model.RoutingSchema, model.AppError) {
	permission := session.GetPermission(model.PERMISSION_SCOPE_SCHEMA)
	if !permission.CanDelete() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_DELETE)
	}

	schema, err := c.app.RemoveRoutingSchema(ctx, session.Domain(0), id)
	if err != nil {
		return nil, err
	}

	c.app.AuditDelete(ctx, session, model.PERMISSION_SCOPE_SCHEMA, schema.Id, schema)

	return schema, nil
}

func (c *Controller) SearchSchemaTags(ctx context.Context, session *auth_manager.Session, search *model.SearchRoutingSchemaTag) ([]*model.RoutingSchemaTag, bool, model.AppError) {
	permission := session.GetPermission(model.PERMISSION_SCOPE_SCHEMA)
	if !permission.CanRead() {
		return nil, false, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	return c.app.GetRoutingSchemaTagsPage(ctx, session.Domain(search.DomainId), search)
}
