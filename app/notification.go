package app

import (
	"context"
	"github.com/webitel/engine/model"
)

func (app *App) SendNotification(ctx context.Context, domainId int64, fromUserId *int64, toUsers []int64, action, description string) *model.AppError {
	var err *model.AppError
	n := &model.Notification{
		DomainId:    domainId,
		Action:      action,
		CreatedBy:   fromUserId,
		ForUsers:    toUsers,
		Description: description,
	}

	n, err = app.Store.Notification().Create(ctx, n)

	if err != nil {
		return err
	}

	return app.MessageQueue.SendNotification(domainId, n)
}
