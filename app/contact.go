package app

import (
	"context"
	"strconv"

	"github.com/webitel/engine/model"
)

// TODO channel: call/chat/task

const (
	setContactNotification = "set_contact"
)

func (app *App) SetCallContactId(ctx context.Context, domainId, userId int64, id string, contactId int64) model.AppError {
	appId, err := app.Store.Call().SetContactId(ctx, domainId, id, contactId)
	if err != nil {
		return err
	}

	if appId != nil {
		if cli, err := app.callManager.CallClientById(*appId); err == nil {
			cli.SetCallVariables(id, map[string]string{
				"wbt_contact_id": strconv.FormatInt(contactId, 10),
			})
		} else {
			app.Log.Warn("not found active call")
		}
	}

	return app.MessageQueue.SendNotification(domainId, &model.Notification{
		DomainId:  domainId,
		Action:    setContactNotification,
		CreatedAt: model.GetMillis(),
		ForUsers:  []int64{userId},
		Body: map[string]any{
			"id":         id,
			"contact_id": contactId,
			"channel":    model.CallExchange,
		},
	})
}

// TODO
func (app *App) SetChatContactId(ctx context.Context, domainId, userId, contactId int64, channelId, conversationId string) model.AppError {
	cli, err := app.chatManager.Client()
	if err != nil {
		return model.NewInternalError("chat.set_contact.cli_err", err.Error())
	}

	err = cli.SetContact(ctx, channelId, conversationId, contactId)
	if err != nil {
		return model.NewInternalError("chat.set_contact.app_err", err.Error())
	}

	return app.MessageQueue.SendNotification(domainId, &model.Notification{
		DomainId:  domainId,
		Action:    setContactNotification,
		CreatedAt: model.GetMillis(),
		ForUsers:  []int64{userId},
		Body: map[string]any{
			"id":         channelId,
			"contact_id": contactId,
			"channel":    model.ChatExchange,
		},
	})
}
