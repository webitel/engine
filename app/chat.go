package app

import (
	"context"
	"github.com/webitel/engine/model"
	"net/http"
)

func (a *App) DeclineChat(authUserId int64, inviteId string) *model.AppError {
	chat, err := a.chatManager.Client()
	if err != nil {
		return model.NewAppError("DeclineChat", "chat.decline.client_err", nil, err.Error(), http.StatusInternalServerError)
	}

	err = chat.Decline(authUserId, inviteId)
	if err != nil {
		return model.NewAppError("DeclineChat", "chat.decline.app_err", nil, err.Error(), http.StatusInternalServerError)
	}

	return nil
}

func (a *App) JoinChat(authUserId int64, inviteId string) (string, *model.AppError) {
	var channelId string
	chat, err := a.chatManager.Client()
	if err != nil {
		return "", model.NewAppError("AcceptChat", "chat.accept.client_err", nil, err.Error(), http.StatusInternalServerError)
	}

	channelId, err = chat.Join(authUserId, inviteId)
	if err != nil {
		return "", model.NewAppError("AcceptChat", "chat.accept.app_err", nil, err.Error(), http.StatusInternalServerError)
	}

	return channelId, nil
}

func (a *App) LeaveChat(authUserId int64, channelId, conversationId string) *model.AppError {
	chat, err := a.chatManager.Client()
	if err != nil {
		return model.NewAppError("LeaveChat", "chat.leave.client_err", nil, err.Error(), http.StatusInternalServerError)
	}

	err = chat.Leave(authUserId, channelId, conversationId)
	if err != nil {
		return model.NewAppError("LeaveChat", "chat.leave.app_err", nil, err.Error(), http.StatusInternalServerError)
	}

	return nil
}

func (a *App) SendTextMessage(authUserId int64, channelId, conversationId, text string) *model.AppError {
	chat, err := a.chatManager.Client()
	if err != nil {
		return model.NewAppError("SendTextMessage", "chat.send.text.client_err.not_found", nil, err.Error(), http.StatusInternalServerError)
	}

	err = chat.SendText(authUserId, channelId, conversationId, text)
	if err != nil {
		return model.NewAppError("SendTextMessage", "chat.send.text.app_err", nil, err.Error(), http.StatusInternalServerError)
	}

	return nil
}

func (a *App) SendFileMessage(authUserId int64, channelId, conversationId string, file *model.ChatFile) *model.AppError {
	chat, err := a.chatManager.Client()
	if err != nil {
		return model.NewAppError("SendFileMessage", "chat.send.file.client_err.not_found", nil, err.Error(), http.StatusInternalServerError)
	}

	err = chat.SendFile(authUserId, channelId, conversationId, file)
	if err != nil {
		return model.NewAppError("SendFileMessage", "chat.send.file.app_err", nil, err.Error(), http.StatusInternalServerError)
	}

	return nil
}

func (a *App) CloseChat(authUserId int64, channelId, conversationId, cause string) *model.AppError {
	chat, err := a.chatManager.Client()
	if err != nil {
		return model.NewAppError("CloseChat", "chat.close.client_err", nil, err.Error(), http.StatusInternalServerError)
	}

	err = chat.CloseConversation(authUserId, channelId, conversationId, cause)
	if err != nil {
		return model.NewAppError("CloseChat", "chat.close.app_err", nil, err.Error(), http.StatusInternalServerError)
	}

	return nil
}

func (a *App) AddToChat(authUserId, userId int64, channelId, conversationId, title string) *model.AppError {
	chat, err := a.chatManager.Client()
	if err != nil {
		return model.NewAppError("AddToChat", "chat.invite.client_err", nil, err.Error(), http.StatusInternalServerError)
	}

	err = chat.AddToChat(authUserId, userId, channelId, conversationId, title)
	if err != nil {
		return model.NewAppError("AddToChat", "chat.invite.app_err", nil, err.Error(), http.StatusInternalServerError)
	}

	return nil
}

func (a *App) StartChat(domainId, authUserId, userId int64) *model.AppError {
	chat, err := a.chatManager.Client()
	if err != nil {
		return model.NewAppError("StartChat", "chat.start.client_err", nil, err.Error(), http.StatusInternalServerError)
	}

	err = chat.NewInternalChat(domainId, authUserId, userId)
	if err != nil {
		return model.NewAppError("StartChat", "chat.start.app_err", nil, err.Error(), http.StatusInternalServerError)
	}

	return nil
}

func (a *App) UpdateChannelChat(authUserId int64, channelId string, readUntil int64) *model.AppError {
	chat, err := a.chatManager.Client()
	if err != nil {
		return model.NewAppError("StartChat", "chat.start.client_err", nil, err.Error(), http.StatusInternalServerError)
	}

	err = chat.UpdateChannel(authUserId, channelId, readUntil)
	if err != nil {
		return model.NewAppError("StartChat", "chat.start.app_err", nil, err.Error(), http.StatusInternalServerError)
	}

	return nil
}

func (a *App) ListActiveChat(token string, domainId, userId int64, page, size int) ([]*model.Conversation, *model.AppError) {
	return a.Store.Chat().OpenedConversations(domainId, userId)
}

func (a *App) BlindTransferChat(domainId int64, conversationId, channelId string, planId int32, vars map[string]string) *model.AppError {
	schemaId, err := a.Store.ChatPlan().GetSchemaId(domainId, planId)
	if err != nil {
		return err
	}

	chat, errChat := a.chatManager.Client()
	if errChat != nil {
		return model.NewAppError("BlindTransferChat", "chat.transfer.client_err", nil, errChat.Error(), http.StatusInternalServerError)
	}

	errChat = chat.BlindTransfer(context.Background(), conversationId, channelId, int64(schemaId), vars)
	if errChat != nil {
		return model.NewAppError("BlindTransferChat", "chat.transfer.api_err", nil, errChat.Error(), http.StatusInternalServerError)
	}

	return nil
}

func (a *App) BlindTransferChatToUser(domainId int64, conversationId, channelId string, userId int64, vars map[string]string) *model.AppError {
	chat, errChat := a.chatManager.Client()
	if errChat != nil {
		return model.NewAppError("BlindTransferChatToUser", "chat.transfer.client_err", nil, errChat.Error(), http.StatusInternalServerError)
	}

	errChat = chat.BlindTransferToUser(context.Background(), conversationId, channelId, userId, vars)
	if errChat != nil {
		return model.NewAppError("BlindTransferChatToUser", "chat.transfer.api_err", nil, errChat.Error(), http.StatusInternalServerError)
	}

	return nil
}
