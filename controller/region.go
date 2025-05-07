package controller

import (
	"context"

	"github.com/webitel/engine/model"
	"github.com/webitel/engine/pkg/wbt/auth_manager"
)

func (c *Controller) SearchRegion(ctx context.Context, session *auth_manager.Session, search *model.SearchRegion) ([]*model.Region, bool, model.AppError) {
	permission := session.GetPermission(model.PERMISSION_SCOPE_DICTIONARIES)
	if !permission.CanRead() {
		return nil, false, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	return c.app.GetRegionsPage(ctx, session.Domain(search.DomainId), search)
}

func (c *Controller) CreateRegion(ctx context.Context, session *auth_manager.Session, region *model.Region) (*model.Region, model.AppError) {
	permission := session.GetPermission(model.PERMISSION_SCOPE_DICTIONARIES)
	if !permission.CanCreate() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_CREATE)
	}

	if err := region.IsValid(); err != nil {
		return nil, err
	}

	return c.app.CreateRegion(ctx, session.Domain(0), region)
}

func (c *Controller) GetRegion(ctx context.Context, session *auth_manager.Session, id int64) (*model.Region, model.AppError) {
	permission := session.GetPermission(model.PERMISSION_SCOPE_DICTIONARIES)
	if !permission.CanRead() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	return c.app.GetRegion(ctx, session.Domain(0), id)
}

func (c *Controller) UpdateRegion(ctx context.Context, session *auth_manager.Session, region *model.Region) (*model.Region, model.AppError) {
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

	return c.app.UpdateRegion(ctx, session.DomainId, region)
}

func (c *Controller) PatchRegion(ctx context.Context, session *auth_manager.Session, id int64, patch *model.RegionPatch) (*model.Region, model.AppError) {
	permission := session.GetPermission(model.PERMISSION_SCOPE_DICTIONARIES)
	if !permission.CanRead() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if !permission.CanUpdate() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
	}

	return c.app.PatchRegion(ctx, session.DomainId, id, patch)
}

func (c *Controller) DeleteRegion(ctx context.Context, session *auth_manager.Session, id int64) (*model.Region, model.AppError) {
	permission := session.GetPermission(model.PERMISSION_SCOPE_DICTIONARIES)
	if !permission.CanDelete() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_DELETE)
	}

	return c.app.RemoveRegion(ctx, session.Domain(0), id)
}
