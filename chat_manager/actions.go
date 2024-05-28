package chat_manager

import (
	proto "buf.build/gen/go/webitel/chat/protocolbuffers/go"
	"buf.build/gen/go/webitel/chat/protocolbuffers/go/messages"
	"context"
	"errors"
	"fmt"
	"github.com/webitel/engine/model"
	"github.com/webitel/wlog"
	"google.golang.org/grpc/metadata"
	"strconv"
	"time"
)

func (cc *chatConnection) Decline(authUserId int64, inviteId string, cause string) error {
	_, err := cc.api.DeclineInvitation(context.Background(), &proto.DeclineInvitationRequest{
		InviteId:   inviteId,
		AuthUserId: authUserId,
		Cause:      cause,
	})

	return err
}

func (cc *chatConnection) Join(authUserId int64, inviteId string) (string, error) {
	res, err := cc.api.JoinConversation(context.Background(), &proto.JoinConversationRequest{
		InviteId:   inviteId,
		AuthUserId: authUserId,
	})

	if err != nil {
		return "", err
	}

	return res.ChannelId, nil
}

func (cc *chatConnection) Leave(authUserId int64, channelId, conversationId string) error {
	_, err := cc.api.LeaveConversation(context.Background(), &proto.LeaveConversationRequest{
		ChannelId:      channelId,
		ConversationId: conversationId,
		AuthUserId:     authUserId,
		Cause:          "",
	})

	return err
}

func (cc *chatConnection) SetVariables(channelId string, vars map[string]string) error {
	_, err := cc.api.SetVariables(context.Background(), &proto.SetVariablesRequest{
		ChannelId: channelId,
		Variables: vars,
	})

	return err
}

func (cc *chatConnection) CloseConversation(authUserId int64, channelId, conversationId, cause string) error {
	_, err := cc.api.CloseConversation(context.Background(), &proto.CloseConversationRequest{
		ConversationId:  conversationId,
		CloserChannelId: channelId,
		Cause:           cause,
		AuthUserId:      authUserId,
	})

	return err
}

func (cc *chatConnection) SendText(authUserId int64, channelId, conversationId, text string) error {
	_, err := cc.api.SendMessage(context.Background(), &proto.SendMessageRequest{
		ConversationId: conversationId,
		ChannelId:      channelId,
		Message: &proto.Message{
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
	_, err := cc.api.SendMessage(context.Background(), &proto.SendMessageRequest{
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
		wlog.Error(fmt.Sprintf("[%s] error: %s", cc.host, err.Error()))
	}

	return err
}

func (cc *chatConnection) AddToChat(authUserId, userId int64, channelId, conversationId, title string) error { // запросити
	_, err := cc.api.InviteToConversation(context.Background(), &proto.InviteToConversationRequest{
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
	res, err := cc.api.StartConversation(context.Background(), &proto.StartConversationRequest{
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

	c, err := cc.api.InviteToConversation(context.Background(), &proto.InviteToConversationRequest{
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
	_, err = cc.api.JoinConversation(context.Background(), &proto.JoinConversationRequest{
		InviteId:   c.InviteId,
		AuthUserId: authUserId,
	})

	return err
}

func (cc *chatConnection) UpdateChannel(authUserId int64, channelId string, readUntil int64) error { // запросити
	_, err := cc.api.UpdateChannel(context.Background(), &proto.UpdateChannelRequest{
		ChannelId:  channelId,
		AuthUserId: authUserId,
		ReadUntil:  readUntil, //
	})
	return err
}

func (cc *chatConnection) ListActive(token string, domainId, userId int64, page, size int) (*proto.GetConversationsResponse, error) {
	header := metadata.New(map[string]string{model.HEADER_TOKEN: token})
	// this is the critical step that includes your headers
	return cc.api.GetConversations(metadata.NewOutgoingContext(context.Background(), header), &proto.GetConversationsRequest{
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

	res, err := cc.api.InviteToConversation(ctx, &proto.InviteToConversationRequest{
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
	_, err := cc.api.BlindTransfer(ctx, &proto.ChatTransferRequest{
		ConversationId: conversationId,
		ChannelId:      channelId,
		SchemaId:       schemaId,
		Variables:      vars,
	})

	return err
}

// TODO check domainId
func (cc *chatConnection) BlindTransferToUser(ctx context.Context, conversationId, channelId string, userId int64, vars map[string]string) error {
	_, err := cc.api.BlindTransfer(ctx, &proto.ChatTransferRequest{
		ConversationId: conversationId,
		ChannelId:      channelId,
		Variables:      vars,
		UserId:         userId,
	})

	return err
}

func (cc *chatConnection) BroadcastMessage(ctx context.Context, message *proto.Message, profileId int64, peer []string) error {
	res, err := cc.mess.BroadcastMessage(ctx, &proto.BroadcastMessageRequest{
		Message: message,
		From:    profileId,
		Peer:    peer,
	})

	if err != nil {
		return err
	}

	if len(res.Failure) > 0 {
		return errors.New(res.Failure[0].String())
	}

	return nil
}

func (cc *chatConnection) SetContact(token string, channelId string, conversationId string, contactId int64) error {
	header := metadata.New(map[string]string{"x-webitel-access": token})
	ctx := metadata.NewOutgoingContext(context.TODO(), header)
	c := fmt.Sprintf("%v", contactId)
	_, err := cc.contact.LinkContactToClient(ctx, &messages.LinkContactToClientRequest{
		ConversationId: conversationId,
		ContactId:      c,
	})

	if err != nil {
		return err
	}

	cc.SetVariables(channelId, map[string]string{
		"wbt_contact_id":   c,
		"wbt_hide_contact": "false",
	})

	return nil
}
