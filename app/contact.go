package app

import (
	"context"
	"github.com/webitel/engine/model"
)

// TODO channel: call/chat/task

const (
	setContactNotification = "set_contact"
)

func (app *App) SetContactId(ctx context.Context, domainId int64, userId int64, channel string, id string, contactId int64) model.AppError {
	err := app.Store.Call().SetContactId(ctx, domainId, id, contactId)
	if err != nil {
		return err
	}

	return app.MessageQueue.SendNotification(domainId, &model.Notification{
		DomainId:  domainId,
		Action:    setContactNotification,
		CreatedAt: model.GetMillis(),
		ForUsers:  []int64{userId},
		Body: map[string]interface{}{
			"id":         id,
			"contact_id": contactId,
		},
	})
}
