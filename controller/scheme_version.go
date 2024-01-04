package controller

import (
	"context"
	"github.com/webitel/engine/auth_manager"
	"github.com/webitel/engine/model"
)

func (c *Controller) SearchSchemeVersions(ctx context.Context, session *auth_manager.Session, search *model.SearchSchemeVersion) ([]*model.SchemeVersion, bool, model.AppError) {
	permission := session.GetPermission(model.PERMISSION_SCOPE_SCHEMA)
	if !permission.CanRead() {
		return nil, false, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	return c.app.SearchSchemeVersions(ctx, search)
}
