package controller

import (
	"context"

	"github.com/webitel/engine/model"
	"github.com/webitel/engine/pkg/wbt/auth_manager"
)

func (c *Controller) SearchUserNotification(ctx context.Context, session *auth_manager.Session, search *model.SearchUserNotification) ([]*model.UserNotification, bool, model.AppError) {
	return c.app.GetUserNotificationPage(ctx, session.Domain(0), session.UserId, search)
}

func (c *Controller) CountUserNotification(ctx context.Context, session *auth_manager.Session, search *model.SearchUserNotification) (int64, model.AppError) {
	return c.app.CountUserNotification(ctx, session.Domain(0), session.UserId, search)
}

func (c *Controller) MarkReadUserNotification(ctx context.Context, session *auth_manager.Session, ids []int64, all bool) (int64, model.AppError) {
	return c.app.MarkReadUserNotification(ctx, session.Domain(0), session.UserId, ids, all)
}
