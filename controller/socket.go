package controller

import (
	"context"
	"github.com/webitel/engine/model"
	"github.com/webitel/engine/pkg/wbt/auth_manager"
)

func (ctrl *Controller) GetWebSocketsPage(ctx context.Context, session *auth_manager.Session, userId int64, search *model.ListRequest) ([]*model.SocketSessionView, bool, model.AppError) {
	permission := session.GetPermission(model.PERMISSION_SCOPE_USERS)
	if !permission.CanRead() {
		return nil, false, ctrl.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	return ctrl.app.GetWebSocketsPage(ctx, session.Domain(0), &model.SearchSocketSessionView{
		ListRequest: *search,
		UserIds:     []int64{userId},
	})
}
