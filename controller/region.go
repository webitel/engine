package controller

import (
	"github.com/webitel/engine/auth_manager"
	"github.com/webitel/engine/model"
)

func (c *Controller) SearchRegion(session *auth_manager.Session, search *model.SearchRegion) ([]*model.Region, bool, *model.AppError) {
	permission := session.GetPermission(model.PERMISSION_SCOPE_DICTIONARIES)
	if !permission.CanRead() {
		return nil, false, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	return c.app.GetRegionsPage(session.Domain(search.DomainId), search)
}

func (c *Controller) CreateRegion(session *auth_manager.Session, region *model.Region) (*model.Region, *model.AppError) {
	permission := session.GetPermission(model.PERMISSION_SCOPE_DICTIONARIES)
	if !permission.CanCreate() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_CREATE)
	}

	if err := region.IsValid(); err != nil {
		return nil, err
	}

	return c.app.CreateRegion(session.Domain(0), region)
}

func (c *Controller) GetRegion(session *auth_manager.Session, id uint32) (*model.Region, *model.AppError) {
	permission := session.GetPermission(model.PERMISSION_SCOPE_DICTIONARIES)
	if !permission.CanRead() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	return c.app.GetRegion(session.Domain(0), id)
}

func (c *Controller) UpdateRegion(session *auth_manager.Session, region *model.Region) (*model.Region, *model.AppError) {
	permission := session.GetPermission(model.PERMISSION_SCOPE_DICTIONARIES)
	if !permission.CanRead() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if !permission.CanUpdate() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
	}

	if err := region.IsValid(); err != nil {
		return nil, err
	}

	return c.app.UpdateRegion(session.DomainId, region)
}

func (c *Controller) PatchRegion(session *auth_manager.Session, id uint32, patch *model.RegionPatch) (*model.Region, *model.AppError) {
	permission := session.GetPermission(model.PERMISSION_SCOPE_DICTIONARIES)
	if !permission.CanRead() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if !permission.CanUpdate() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
	}

	return c.app.PatchRegion(session.DomainId, id, patch)
}

func (c *Controller) DeleteRegion(session *auth_manager.Session, id uint32) (*model.Region, *model.AppError) {
	permission := session.GetPermission(model.PERMISSION_SCOPE_DICTIONARIES)
	if !permission.CanDelete() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_DELETE)
	}

	return c.app.RemoveRegion(session.Domain(0), id)
}
