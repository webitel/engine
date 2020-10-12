package app

import (
	"github.com/webitel/engine/model"
	"net/http"
)

func (a *App) DeclineChat(userId int64, inviteId string) *model.AppError {
	chat, err := a.chatManager.Client()
	if err != nil {
		return model.NewAppError("DeclineChat", "chat.decline.client_err", nil, err.Error(), http.StatusInternalServerError)
	}

	err = chat.Decline(userId, inviteId)
	if err != nil {
		return model.NewAppError("DeclineChat", "chat.decline.app_err", nil, err.Error(), http.StatusInternalServerError)
	}

	return nil
}

func (a *App) JoinChat(inviteId string) (string, *model.AppError) {
	var channelId string
	chat, err := a.chatManager.Client()
	if err != nil {
		return "", model.NewAppError("AcceptChat", "chat.accept.client_err", nil, err.Error(), http.StatusInternalServerError)
	}

	channelId, err = chat.Join(inviteId)
	if err != nil {
		return "", model.NewAppError("AcceptChat", "chat.accept.app_err", nil, err.Error(), http.StatusInternalServerError)
	}

	return channelId, nil
}

func (a *App) LeaveChat(channelId, conversationId string) *model.AppError {
	chat, err := a.chatManager.Client()
	if err != nil {
		return model.NewAppError("LeaveChat", "chat.leave.client_err", nil, err.Error(), http.StatusInternalServerError)
	}

	err = chat.Leave(channelId, conversationId)
	if err != nil {
		return model.NewAppError("LeaveChat", "chat.leave.app_err", nil, err.Error(), http.StatusInternalServerError)
	}

	return nil
}

func (a *App) SendTextMessage(channelId, conversationId, text string) *model.AppError {
	chat, err := a.chatManager.Client()
	if err != nil {
		return model.NewAppError("SendTextMessage", "chat.send.client_err", nil, err.Error(), http.StatusInternalServerError)
	}

	err = chat.SendText(channelId, conversationId, text)
	if err != nil {
		return model.NewAppError("SendTextMessage", "chat.send.app_err", nil, err.Error(), http.StatusInternalServerError)
	}

	return nil
}

func (a *App) CloseChat(channelId, conversationId, cause string) *model.AppError {
	chat, err := a.chatManager.Client()
	if err != nil {
		return model.NewAppError("CloseChat", "chat.close.client_err", nil, err.Error(), http.StatusInternalServerError)
	}

	err = chat.CloseConversation(channelId, conversationId, cause)
	if err != nil {
		return model.NewAppError("CloseChat", "chat.close.app_err", nil, err.Error(), http.StatusInternalServerError)
	}

	return nil
}
