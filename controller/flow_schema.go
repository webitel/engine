package controller

import (
	"github.com/webitel/engine/auth_manager"
	"github.com/webitel/engine/model"
)

func (c *Controller) CreateRoutingSchema(session *auth_manager.Session, schema *model.RoutingSchema) (*model.RoutingSchema, *model.AppError) {
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

	return c.app.CreateRoutingSchema(schema)
}

func (c *Controller) SearchSchema(session *auth_manager.Session, search *model.SearchRoutingSchema) ([]*model.RoutingSchema, bool, *model.AppError) {
	permission := session.GetPermission(model.PERMISSION_SCOPE_SCHEMA)
	if !permission.CanRead() {
		return nil, false, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	return c.app.GetRoutingSchemaPage(session.Domain(search.DomainId), search)
}

func (c *Controller) GetSchema(session *auth_manager.Session, id int64) (*model.RoutingSchema, *model.AppError) {
	permission := session.GetPermission(model.PERMISSION_SCOPE_SCHEMA)
	if !permission.CanRead() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	return c.app.GetRoutingSchemaById(session.Domain(0), id)
}

func (c *Controller) UpdateSchema(session *auth_manager.Session, schema *model.RoutingSchema) (*model.RoutingSchema, *model.AppError) {
	permission := session.GetPermission(model.PERMISSION_SCOPE_SCHEMA)
	if !permission.CanRead() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if !permission.CanUpdate() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
	}

	if err := schema.IsValid(); err != nil {
		return nil, err
	}

	schema.DomainRecord.DomainId = session.DomainId
	schema.DomainRecord.UpdatedAt = model.GetMillis()
	schema.DomainRecord.UpdatedBy = &model.Lookup{
		Id: int(session.UserId),
	}

	return c.app.UpdateRoutingSchema(schema)
}

func (c *Controller) PatchSchema(session *auth_manager.Session, id int64, patch *model.RoutingSchemaPath) (*model.RoutingSchema, *model.AppError) {
	permission := session.GetPermission(model.PERMISSION_SCOPE_SCHEMA)
	if !permission.CanRead() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if !permission.CanUpdate() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
	}

	patch.UpdatedById = int(session.UserId)

	return c.app.PatchRoutingSchema(session.DomainId, id, patch)
}

func (c *Controller) DeleteSchema(session *auth_manager.Session, id int64) (*model.RoutingSchema, *model.AppError) {
	permission := session.GetPermission(model.PERMISSION_SCOPE_SCHEMA)
	if !permission.CanDelete() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_DELETE)
	}

	return c.app.RemoveRoutingSchema(session.Domain(0), id)
}

func (c *Controller) SearchSchemaTags(session *auth_manager.Session, search *model.SearchRoutingSchemaTag) ([]*model.RoutingSchemaTag, bool, *model.AppError) {
	permission := session.GetPermission(model.PERMISSION_SCOPE_SCHEMA)
	if !permission.CanRead() {
		return nil, false, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	return c.app.GetRoutingSchemaTagsPage(session.Domain(search.DomainId), search)
}
