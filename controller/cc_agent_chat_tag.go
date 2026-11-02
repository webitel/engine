package controller

import (
	"context"
	"github.com/webitel/engine/model"
	"github.com/webitel/engine/pkg/wbt/auth_manager"
)

func (c *Controller) SearchAgentChatTag(ctx context.Context, session *auth_manager.Session, search *model.SearchTeamChatTag) ([]*model.TeamChatTag, bool, model.AppError) {
	userId := session.GetUserId()
	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_AGENT)
	if !permission.CanRead() {
		return nil, false, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	return c.app.SearchAgentChatTag(ctx, session.Domain(0), userId, search)
}
