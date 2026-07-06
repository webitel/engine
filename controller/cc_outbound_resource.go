package controller

import (
	"context"

	"github.com/webitel/engine/model"
	"github.com/webitel/engine/pkg/wbt/auth_manager"
)

func (c *Controller) CreateOutboundResourceDisplays(ctx context.Context, resourceID int64, displays []*model.ResourceDisplay) ([]*model.ResourceDisplay, model.AppError) {
	session, err := c.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_OUTBOUND_RESOURCE)
	if !permission.CanRead() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if !permission.CanUpdate() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
	}

	if session.UseRBAC(auth_manager.PERMISSION_ACCESS_UPDATE, permission) {
		canAccess, err := c.app.OutboundResourceCheckAccess(ctx, session.DomainId, resourceID, session.GetAclRoles(), auth_manager.PERMISSION_ACCESS_UPDATE)
		if err != nil {
			return nil, err
		}

		if !canAccess {
			return nil, c.app.MakeResourcePermissionError(session, resourceID, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
		}
	}

	return c.app.CreateOutboundResourceDisplays(ctx, resourceID, displays)
}
