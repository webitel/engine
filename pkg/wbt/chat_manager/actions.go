package chat_manager

import (
	"context"
	"errors"
	"fmt"
	"github.com/webitel/engine/pkg/wbt"
	proto "github.com/webitel/engine/pkg/wbt/gen/chat"
	msg "github.com/webitel/engine/pkg/wbt/gen/chat/messages"
	"github.com/webitel/wlog"
	"strconv"
	"time"
)

func (cc *chatConnection) Decline(authUserId int64, inviteId string, cause string) error {
	_, err := cc.cm.chatCli.Api.DeclineInvitation(context.Background(), &proto.DeclineInvitationRequest{
		InviteId:   inviteId,
		AuthUserId: authUserId,
		Cause:      cause,
	})

	return err
}

func (cc *chatConnection) Join(authUserId int64, inviteId string) (string, error) {
	res, err := cc.cm.chatCli.Api.JoinConversation(context.Background(), &proto.JoinConversationRequest{
		InviteId:   inviteId,
		AuthUserId: authUserId,
	})

	if err != nil {
		return "", err
	}

	return res.ChannelId, nil
}

func (cc *chatConnection) Leave(authUserId int64, channelId, conversationId string, cause LeaveCause) error {

	_, err := cc.cm.chatCli.Api.LeaveConversation(context.Background(), &proto.LeaveConversationRequest{
		ChannelId:      channelId,
		ConversationId: conversationId,
		AuthUserId:     authUserId,
		Cause:          findLeaveConversationCause(cause.String()),
	})

	return err
}

func (cc *chatConnection) SetVariables(channelId string, vars map[string]string) error {
	_, err := cc.cm.chatCli.Api.SetVariables(context.Background(), &proto.SetVariablesRequest{
		ChannelId: channelId,
		Variables: vars,
	})

	return err
}

func (cc *chatConnection) CloseConversation(authUserId int64, channelId, conversationId string, cause CloseCause) error {
	_, err := cc.cm.chatCli.Api.CloseConversation(context.Background(), &proto.CloseConversationRequest{
		ConversationId:  conversationId,
		CloserChannelId: channelId,
		Cause:           findCloseChatCause(cause.String()),
		AuthUserId:      authUserId,
	})

	return err
}

// findLeaveConversationCause tries to find leave reason in the proto.LeaveConversationCause enum.
// If not found returns default_cause
func findLeaveConversationCause(cause string) proto.LeaveConversationCause {
	for name, i := range proto.LeaveConversationCause_value {
		if name == cause {
			return proto.LeaveConversationCause(i)
		}
	}
	return 0
}

// findCloseChatCause tries to find close reason in the proto.CloseConversationCause enum
// If not found returns no_cause
func findCloseChatCause(cause string) proto.CloseConversationCause {
	for name, i := range proto.CloseConversationCause_value {
		if name == cause {
			return proto.CloseConversationCause(i)
		}
	}
	return 0
}

func (cc *chatConnection) SendText(authUserId int64, channelId, conversationId, text string) error {
	_, err := cc.cm.chatCli.Api.SendMessage(context.Background(), &proto.SendMessageRequest{
		ConversationId: conversationId,
		ChannelId:      channelId,
		Message: &proto.Message{
			Type: "text", // TODO
			Text: text,
		},
		AuthUserId: authUserId,
	})

	if err != nil {
		wlog.Error(fmt.Sprintf("error: %s", err.Error()))
	}

	return err
}

func (cc *chatConnection) SendFile(authUserId int64, channelId, conversationId string, file *ChatFile) error {
	_, err := cc.cm.chatCli.Api.SendMessage(context.Background(), &proto.SendMessageRequest{
		ConversationId: conversationId,
		ChannelId:      channelId,
		Message: &proto.Message{
			Type: "file", // TODO
			File: &proto.File{
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
		wlog.Error(fmt.Sprintf("error: %s", err.Error()))
	}

	return err
}

func (cc *chatConnection) AddToChat(authUserId, userId int64, channelId, conversationId, title string) error { // запросити
	_, err := cc.cm.chatCli.Api.InviteToConversation(context.Background(), &proto.InviteToConversationRequest{
		User: &proto.User{
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
	res, err := cc.cm.chatCli.Api.StartConversation(context.Background(), &proto.StartConversationRequest{
		User: &proto.User{ // caller
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

	c, err := cc.cm.chatCli.Api.InviteToConversation(context.Background(), &proto.InviteToConversationRequest{
		User: &proto.User{
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
	_, err = cc.cm.chatCli.Api.JoinConversation(context.Background(), &proto.JoinConversationRequest{
		InviteId:   c.InviteId,
		AuthUserId: authUserId,
	})

	return err
}

func (cc *chatConnection) UpdateChannel(authUserId int64, channelId string, readUntil int64) error { // запросити
	_, err := cc.cm.chatCli.Api.UpdateChannel(context.Background(), &proto.UpdateChannelRequest{
		ChannelId:  channelId,
		AuthUserId: authUserId,
		ReadUntil:  readUntil, //
	})
	return err
}

func (cc *chatConnection) ListActive(token string, domainId, userId int64, page, size int) (*proto.GetConversationsResponse, error) {
	ctx := wbt.WithToken(context.Background(), token)
	// this is the critical step that includes your headers
	return cc.cm.chatCli.Api.GetConversations(ctx, &proto.GetConversationsRequest{
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

	res, err := cc.cm.chatCli.Api.InviteToConversation(ctx, &proto.InviteToConversationRequest{
		User: &proto.User{
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
	_, err := cc.cm.chatCli.Api.BlindTransfer(ctx, &proto.ChatTransferRequest{
		ConversationId: conversationId,
		ChannelId:      channelId,
		SchemaId:       schemaId,
		Variables:      vars,
	})

	return err
}

// TODO check domainId
func (cc *chatConnection) BlindTransferToUser(ctx context.Context, conversationId, channelId string, userId int64, vars map[string]string) error {
	_, err := cc.cm.chatCli.Api.BlindTransfer(ctx, &proto.ChatTransferRequest{
		ConversationId: conversationId,
		ChannelId:      channelId,
		Variables:      vars,
		UserId:         userId,
	})

	return err
}

func (cc *chatConnection) BroadcastMessage(ctx context.Context, message *proto.Message, profileId int64, peer []string) error {
	return errors.New("deprecated")
}

func (cc *chatConnection) SetContact(ctx context.Context, channelId string, conversationId string, contactId int64) error {
	c := fmt.Sprintf("%v", contactId)
	_, err := cc.cm.contactCli.Api.LinkContactToClientNA(ctx, &msg.LinkContactToClientNARequest{
		ConversationId: conversationId,
		ContactId:      c,
	})

	if err != nil {
		return err
	}

	return nil
}
