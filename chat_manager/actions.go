package chat_manager

import (
	"context"
	client "github.com/webitel/protos/chat"
)

func (cc *chatConnection) Decline(userId int64, inviteId string) error {
	_, err := cc.api.DeclineInvitation(context.Background(), &client.DeclineInvitationRequest{
		InviteId: inviteId,
		UserId:   userId,
	})

	return err
}

func (cc *chatConnection) Join(inviteId string) (string, error) {
	res, err := cc.api.JoinConversation(context.Background(), &client.JoinConversationRequest{
		InviteId: inviteId,
	})

	if err != nil {
		return "", err
	}

	return res.ChannelId, nil
}

func (cc *chatConnection) Leave(channelId, conversationId string) error {
	_, err := cc.api.LeaveConversation(context.Background(), &client.LeaveConversationRequest{
		ChannelId:      channelId,
		ConversationId: conversationId,
	})

	return err
}

func (cc *chatConnection) CloseConversation(channelId, conversationId, cause string) error {
	_, err := cc.api.CloseConversation(context.Background(), &client.CloseConversationRequest{
		ConversationId:  conversationId,
		CloserChannelId: channelId,
		FromFlow:        false,
		Cause:           cause,
	})

	return err
}

func (cc *chatConnection) SendText(channelId, conversationId, text string) error {
	_, err := cc.api.SendMessage(context.Background(), &client.SendMessageRequest{
		ConversationId: conversationId,
		ChannelId:      channelId,
		FromFlow:       false,
		Message: &client.Message{
			Type: "text", // TODO
			Value: &client.Message_Text{
				Text: text,
			},
		},
	})

	return err
}

func (cc *chatConnection) AddToChat() { // запросити
	cc.api.InviteToConversation(context.Background(), &client.InviteToConversationRequest{
		User: &client.User{
			UserId:     0,
			Type:       "webitel",
			Connection: "", // profile
			Internal:   true,
		},
		ConversationId:   "",
		InviterChannelId: "",
	})
}

func (cc *chatConnection) NewInternalChat() {
	res, _ := cc.api.StartConversation(context.Background(), &client.StartConversationRequest{
		User: &client.User{ // caller
			UserId:     0,
			Type:       "webitel",
			Connection: "", // profile
			Internal:   true,
		},
		DomainId: 0,
	})

	cc.api.InviteToConversation(context.Background(), &client.InviteToConversationRequest{
		User: &client.User{
			UserId:     0,
			Type:       "webitel",
			Connection: "", // profile
			Internal:   true,
		},
		ConversationId:   res.ConversationId,
		InviterChannelId: res.ChannelId,
	})
}
