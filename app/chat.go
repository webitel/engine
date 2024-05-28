package app

import (
	"context"
	"github.com/webitel/engine/model"

	proto "buf.build/gen/go/webitel/chat/protocolbuffers/go"
	"net/url"
)

var publicStorage *url.URL

func (a *App) DeclineChat(authUserId int64, inviteId string, cause string) model.AppError {
	chat, err := a.chatManager.Client()
	if err != nil {
		return model.NewInternalError("chat.decline.client_err", err.Error())
	}

	err = chat.Decline(authUserId, inviteId, cause)
	if err != nil {
		return model.NewInternalError("chat.decline.app_err", err.Error())
	}

	return nil
}

func (a *App) JoinChat(authUserId int64, inviteId string) (string, model.AppError) {
	var channelId string
	chat, err := a.chatManager.Client()
	if err != nil {
		return "", model.NewInternalError("chat.accept.client_err", err.Error())
	}

	channelId, err = chat.Join(authUserId, inviteId)
	if err != nil {
		return "", model.NewInternalError("chat.accept.app_err", err.Error())
	}

	return channelId, nil
}

func (a *App) LeaveChat(authUserId int64, channelId, conversationId string) model.AppError {
	chat, err := a.chatManager.Client()
	if err != nil {
		return model.NewInternalError("chat.leave.client_err", err.Error())
	}

	err = chat.Leave(authUserId, channelId, conversationId)
	if err != nil {
		return model.NewInternalError("chat.leave.app_err", err.Error())
	}

	return nil
}

func (a *App) SendTextMessage(authUserId int64, channelId, conversationId, text string) model.AppError {
	chat, err := a.chatManager.Client()
	if err != nil {
		return model.NewInternalError("chat.send.text.client_err.not_found", err.Error())
	}

	err = chat.SendText(authUserId, channelId, conversationId, text)
	if err != nil {
		return model.NewInternalError("chat.send.text.app_err", err.Error())
	}

	return nil
}

func setupPublicStorageUrl(storageUrl *string) {
	var err error
	if storageUrl == nil || *storageUrl == "" {
		return
	}

	publicStorage, err = url.Parse(*storageUrl)
	if err != nil {
		panic(err.Error())
	}

}

func (a *App) SendFileMessage(authUserId int64, channelId, conversationId string, file *model.ChatFile) model.AppError {
	chat, err := a.chatManager.Client()
	if err != nil {
		return model.NewInternalError("chat.send.file.client_err.not_found", err.Error())
	}

	// TODO WTEL-3713
	if publicStorage != nil && file.Url != "" {
		var u *url.URL
		u, err = url.Parse(file.Url)
		if err != nil {
			return model.NewInternalError("chat.send.file.valid.url", err.Error())
		}

		u.Host = publicStorage.Host
		u.Scheme = publicStorage.Scheme
		file.Url = u.String()
	}

	err = chat.SendFile(authUserId, channelId, conversationId, file)
	if err != nil {
		return model.NewInternalError("chat.send.file.app_err", err.Error())
	}

	return nil
}

func (a *App) CloseChat(authUserId int64, channelId, conversationId, cause string) model.AppError {
	chat, err := a.chatManager.Client()
	if err != nil {
		return model.NewInternalError("chat.close.client_err", err.Error())
	}

	err = chat.CloseConversation(authUserId, channelId, conversationId, cause)
	if err != nil {
		return model.NewInternalError("chat.close.app_err", err.Error())
	}

	return nil
}

func (a *App) AddToChat(authUserId, userId int64, channelId, conversationId, title string) model.AppError {
	chat, err := a.chatManager.Client()
	if err != nil {
		return model.NewInternalError("chat.invite.client_err", err.Error())
	}

	err = chat.AddToChat(authUserId, userId, channelId, conversationId, title)
	if err != nil {
		return model.NewInternalError("chat.invite.app_err", err.Error())
	}

	return nil
}

func (a *App) StartChat(domainId, authUserId, userId int64) model.AppError {
	chat, err := a.chatManager.Client()
	if err != nil {
		return model.NewInternalError("chat.start.client_err", err.Error())
	}

	err = chat.NewInternalChat(domainId, authUserId, userId)
	if err != nil {
		return model.NewInternalError("chat.start.app_err", err.Error())
	}

	return nil
}

func (a *App) UpdateChannelChat(authUserId int64, channelId string, readUntil int64) model.AppError {
	chat, err := a.chatManager.Client()
	if err != nil {
		return model.NewInternalError("chat.start.client_err", err.Error())
	}

	err = chat.UpdateChannel(authUserId, channelId, readUntil)
	if err != nil {
		return model.NewInternalError("chat.start.app_err", err.Error())
	}

	return nil
}

func (a *App) ListActiveChat(ctx context.Context, token string, domainId, userId int64, page, size int) ([]*model.Conversation, model.AppError) {
	return a.Store.Chat().OpenedConversations(ctx, domainId, userId)
}

func (a *App) BlindTransferChat(ctx context.Context, domainId int64, conversationId, channelId string, planId int32, vars map[string]string) model.AppError {
	schema, err := a.Store.ChatPlan().GetSchemaId(ctx, domainId, planId)
	if err != nil {
		return err
	}

	chat, errChat := a.chatManager.Client()
	if errChat != nil {
		return model.NewInternalError("chat.transfer.client_err", errChat.Error())
	}

	if len(vars) == 0 {
		vars = make(map[string]string)
	}

	vars["chatplan_name"] = schema.Name

	errChat = chat.BlindTransfer(context.Background(), conversationId, channelId, int64(schema.Id), vars)
	if errChat != nil {
		return model.NewInternalError("chat.transfer.api_err", errChat.Error())
	}

	return nil
}

func (a *App) BlindTransferChatToUser(domainId int64, conversationId, channelId string, userId int64, vars map[string]string) model.AppError {
	chat, errChat := a.chatManager.Client()
	if errChat != nil {
		return model.NewInternalError("chat.transfer.client_err", errChat.Error())
	}

	errChat = chat.BlindTransferToUser(context.Background(), conversationId, channelId, userId, vars)
	if errChat != nil {
		return model.NewInternalError("chat.transfer.api_err", errChat.Error())
	}

	return nil
}

func (a *App) BroadcastChatBot(ctx context.Context, domainId int64, profileId int64, peer []string, text string) model.AppError {

	appErr := a.Store.Chat().ValidDomain(ctx, domainId, profileId)
	if appErr != nil {
		return appErr
	}

	cli, err := a.chatManager.Client()
	if err != nil {
		return model.NewInternalError("chat.broadcast.cli_err", err.Error())
	}

	msg := &proto.Message{
		Type: "text", //TODO
		Text: text,
	}

	err = cli.BroadcastMessage(ctx, msg, profileId, peer)
	if err != nil {
		return model.NewInternalError("chat.broadcast.api_err", err.Error())
	}

	return nil
}

func (a *App) SetContactToChat(token string, channelId string, conversationId string, contactId int64) model.AppError {
	cli, err := a.chatManager.Client()
	if err != nil {
		return model.NewInternalError("chat.set_contact.cli_err", err.Error())
	}

	err = cli.SetContact(token, channelId, conversationId, contactId)
	if err != nil {
		return model.NewInternalError("chat.set_contact.app_err", err.Error())
	}

	return nil
}
