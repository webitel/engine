package app

import (
	"context"

	"github.com/webitel/engine/model"
)

func (app *App) GetUserNotificationPage(ctx context.Context, domainId, userId int64, search *model.SearchUserNotification) ([]*model.UserNotification, bool, model.AppError) {
	list, err := app.Store.UserNotification().GetAllPage(ctx, domainId, userId, search)
	if err != nil {
		return nil, false, err
	}

	search.RemoveLastElemIfNeed(&list)

	return list, search.EndOfList(), nil
}

func (app *App) CountUserNotification(ctx context.Context, domainId, userId int64, search *model.SearchUserNotification) (int64, model.AppError) {
	return app.Store.UserNotification().Count(ctx, domainId, userId, search)
}

func (app *App) MarkReadUserNotification(ctx context.Context, domainId, userId int64, ids []int64, all bool) (int64, model.AppError) {
	if !all && len(ids) == 0 {
		return 0, model.NewBadRequestError("app.user_notification.mark_read.valid.id", "id or all is required")
	}

	return app.Store.UserNotification().MarkRead(ctx, domainId, userId, ids, all)
}
