package chat_manager

import (
	"context"
	"fmt"
	"github.com/webitel/engine/model"
	client "github.com/webitel/protos/engine/chat"
	"github.com/webitel/wlog"
	"google.golang.org/grpc/metadata"
	"strconv"
	"time"
)

func (cc *chatConnection) Decline(authUserId int64, inviteId string, cause string) error {
	_, err := cc.api.DeclineInvitation(context.Background(), &client.DeclineInvitationRequest{
		InviteId:   inviteId,
		AuthUserId: authUserId,
		Cause:      cause,
	})

	return err
}

func (cc *chatConnection) Join(authUserId int64, inviteId string) (string, error) {
	res, err := cc.api.JoinConversation(context.Background(), &client.JoinConversationRequest{
		InviteId:   inviteId,
		AuthUserId: authUserId,
	})

	if err != nil {
		return "", err
	}

	return res.ChannelId, nil
}

func (cc *chatConnection) Leave(authUserId int64, channelId, conversationId string) error {
	_, err := cc.api.LeaveConversation(context.Background(), &client.LeaveConversationRequest{
		ChannelId:      channelId,
		ConversationId: conversationId,
		AuthUserId:     authUserId,
		Cause:          "",
	})

	return err
}
func (cc *chatConnection) SetVariables(channelId string, vars map[string]string) error {
	_, err := cc.api.SetVariables(context.Background(), &client.SetVariablesRequest{
		ChannelId: channelId,
		Variables: vars,
	})

	return err
}

func (cc *chatConnection) CloseConversation(authUserId int64, channelId, conversationId, cause string) error {
	_, err := cc.api.CloseConversation(context.Background(), &client.CloseConversationRequest{
		ConversationId:  conversationId,
		CloserChannelId: channelId,
		Cause:           cause,
		AuthUserId:      authUserId,
	})

	return err
}

func (cc *chatConnection) SendText(authUserId int64, channelId, conversationId, text string) error {
	_, err := cc.api.SendMessage(context.Background(), &client.SendMessageRequest{
		ConversationId: conversationId,
		ChannelId:      channelId,
		Message: &client.Message{
			Type: "text", // TODO
			Text: text,
		},
		AuthUserId: authUserId,
	})

	if err != nil {
		wlog.Error(fmt.Sprintf("[%s] error: %s", cc.host, err.Error()))
	}

	return err
}

func (cc *chatConnection) SendFile(authUserId int64, channelId, conversationId string, file *model.ChatFile) error {
	_, err := cc.api.SendMessage(context.Background(), &client.SendMessageRequest{
		ConversationId: conversationId,
		ChannelId:      channelId,
		Message: &client.Message{
			Type: "file", // TODO
			File: &client.File{
				Id:   file.Id,
				Url:  file.Url,
				Mime: file.Mime,
				Size: file.Size,
				Name: file.Name,
			},
		},
		AuthUserId: authUserId,
	})

	if err != nil {
		wlog.Error(fmt.Sprintf("[%s] error: %s", cc.host, err.Error()))
	}

	return err
}

func (cc *chatConnection) AddToChat(authUserId, userId int64, channelId, conversationId, title string) error { // запросити
	_, err := cc.api.InviteToConversation(context.Background(), &client.InviteToConversationRequest{
		User: &client.User{
			UserId:     userId,
			Type:       "webitel",
			Connection: "", // profile
			Internal:   true,
		},
		ConversationId:   conversationId,
		InviterChannelId: channelId,
		AuthUserId:       authUserId,
		Title:            title,
		DomainId:         1, // todo add
	})
	return err
}

func (cc *chatConnection) NewInternalChat(domainId, authUserId, userId int64) error {
	res, err := cc.api.StartConversation(context.Background(), &client.StartConversationRequest{
		User: &client.User{ // caller
			UserId:     authUserId,
			Type:       "webitel",
			Connection: "", // profile
			Internal:   true,
		},
		DomainId: domainId,
	})

	if err != nil {
		return err
	}

	c, err := cc.api.InviteToConversation(context.Background(), &client.InviteToConversationRequest{
		User: &client.User{
			UserId:     userId,
			Type:       "webitel",
			Connection: "", // profile
			Internal:   true,
		},
		ConversationId:   res.ConversationId,
		InviterChannelId: res.ChannelId,
		TimeoutSec:       30,
		DomainId:         domainId,
		AuthUserId:       authUserId,
		Title:            "test",
	})

	if err != nil {
		return err
	}
	time.Sleep(time.Second)
	_, err = cc.api.JoinConversation(context.Background(), &client.JoinConversationRequest{
		InviteId:   c.InviteId,
		AuthUserId: authUserId,
	})

	return err
}

func (cc *chatConnection) UpdateChannel(authUserId int64, channelId string, readUntil int64) error { // запросити
	_, err := cc.api.UpdateChannel(context.Background(), &client.UpdateChannelRequest{
		ChannelId:  channelId,
		AuthUserId: authUserId,
		ReadUntil:  readUntil, //
	})
	return err
}

func (cc *chatConnection) ListActive(token string, domainId, userId int64, page, size int) (*client.GetConversationsResponse, error) {
	header := metadata.New(map[string]string{model.HEADER_TOKEN: token})
	// this is the critical step that includes your headers
	return cc.api.GetConversations(metadata.NewOutgoingContext(context.Background(), header), &client.GetConversationsRequest{
		Active:      true,
		Page:        int32(page),
		Size:        int32(size),
		DomainId:    domainId,
		UserId:      userId,
		MessageSize: 40, //TODO
	})
}

func (cc *chatConnection) InviteToConversation(ctx context.Context, domainId, userId int64, conversationId, inviterId, invUserId, title string, timeout int, vars map[string]string) (string, error) {

	inviterUserId, _ := strconv.Atoi(invUserId)

	res, err := cc.api.InviteToConversation(ctx, &client.InviteToConversationRequest{
		User: &client.User{
			UserId:   userId,
			Type:     "webitel",
			Internal: true,
		},
		InviterChannelId: inviterId,
		AuthUserId:       int64(inviterUserId),
		ConversationId:   conversationId,
		Variables:        vars,
		TimeoutSec:       int64(timeout),
		DomainId:         domainId,
		Title:            title,
		AppId:            "",
	})

	if err != nil {
		return "", err
	}

	return res.InviteId, nil
}

func (cc *chatConnection) BlindTransfer(ctx context.Context, conversationId, channelId string, schemaId int64, vars map[string]string) error {
	_, err := cc.api.BlindTransfer(ctx, &client.ChatTransferRequest{
		ConversationId: conversationId,
		ChannelId:      channelId,
		SchemaId:       schemaId,
		Variables:      vars,
	})

	return err
}

//TODO check domainId
func (cc *chatConnection) BlindTransferToUser(ctx context.Context, conversationId, channelId string, userId int64, vars map[string]string) error {
	_, err := cc.api.BlindTransfer(ctx, &client.ChatTransferRequest{
		ConversationId: conversationId,
		ChannelId:      channelId,
		Variables:      vars,
		UserId:         userId,
	})

	return err
}
