package app

import (
	"github.com/webitel/engine/model"
)

func (app *App) SendNotification(domainId int64, fromUserId *int64, toUsers []int64, action, description string) *model.AppError {
	var err *model.AppError
	n := &model.Notification{
		DomainId:    domainId,
		Action:      action,
		CreatedBy:   fromUserId,
		ForUsers:    toUsers,
		Description: description,
	}

	n, err = app.Store.Notification().Create(n)

	if err != nil {
		return err
	}

	return app.MessageQueue.SendNotification(domainId, n)
}
