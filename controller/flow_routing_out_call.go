package controller

import (
	"github.com/webitel/engine/auth_manager"
	"github.com/webitel/engine/model"
)

func (c *Controller) CreateRoutingOutboundCall(session *auth_manager.Session, routing *model.RoutingOutboundCall) (*model.RoutingOutboundCall, *model.AppError) {
	permission := session.GetPermission(model.PERMISSION_SCOPE_ACR_ROUTING)
	if !permission.CanCreate() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_CREATE)
	}

	if err := routing.IsValid(); err != nil {
		return nil, err
	}

	routing.DomainId = session.Domain(routing.DomainId)

	return c.app.CreateRoutingOutboundCall(routing)
}

func (c *Controller) SearchRoutingOutboundCall(session *auth_manager.Session, search *model.SearchRoutingOutboundCall) ([]*model.RoutingOutboundCall, bool, *model.AppError) {
	permission := session.GetPermission(model.PERMISSION_SCOPE_ACR_ROUTING)
	if !permission.CanRead() {
		return nil, false, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	return c.app.GetRoutingOutboundCallPage(session.Domain(search.DomainId), search)
}

func (c *Controller) GetRoutingOutboundCall(session *auth_manager.Session, domainId, id int64) (*model.RoutingOutboundCall, *model.AppError) {
	permission := session.GetPermission(model.PERMISSION_SCOPE_ACR_ROUTING)
	if !permission.CanRead() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	return c.app.GetRoutingOutboundCallById(session.Domain(domainId), id)
}

func (c *Controller) UpdateRoutingOutboundCall(session *auth_manager.Session, routing *model.RoutingOutboundCall) (*model.RoutingOutboundCall, *model.AppError) {
	permission := session.GetPermission(model.PERMISSION_SCOPE_ACR_ROUTING)
	if !permission.CanRead() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if !permission.CanUpdate() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
	}
	if err := routing.IsValid(); err != nil {
		return nil, err
	}

	routing.DomainId = session.Domain(routing.DomainId)
	return c.app.UpdateRoutingOutboundCall(routing)
}

func (c *Controller) PatchRoutingOutboundCall(session *auth_manager.Session, domainId, id int64, patch *model.RoutingOutboundCallPatch) (*model.RoutingOutboundCall, *model.AppError) {
	permission := session.GetPermission(model.PERMISSION_SCOPE_ACR_ROUTING)
	if !permission.CanRead() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if !permission.CanUpdate() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
	}

	patch.UpdatedById = int(session.UserId)

	return c.app.PatchRoutingOutboundCall(session.Domain(domainId), id, patch)
}

func (c *Controller) ChangePositionOutboundCall(session *auth_manager.Session, domainId, fromId, toId int64) *model.AppError {
	permission := session.GetPermission(model.PERMISSION_SCOPE_ACR_ROUTING)
	if !permission.CanRead() {
		return c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if !permission.CanUpdate() {
		return c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
	}

	return c.app.ChangePositionOutboundCall(session.Domain(domainId), fromId, toId)
}

func (c *Controller) DeleteRoutingOutboundCall(session *auth_manager.Session, domainId, id int64) (*model.RoutingOutboundCall, *model.AppError) {
	permission := session.GetPermission(model.PERMISSION_SCOPE_ACR_ROUTING)
	if !permission.CanDelete() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_DELETE)
	}

	return c.app.RemoveRoutingOutboundCall(session.Domain(domainId), id)
}
