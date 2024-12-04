package grpc_api

import (
	chat "buf.build/gen/go/webitel/chat/protocolbuffers/go"
	gogrpc "buf.build/gen/go/webitel/engine/grpc/go/_gogrpc"
	engine "buf.build/gen/go/webitel/engine/protocolbuffers/go"
	"context"
)

type chatHelper struct {
	*API
	gogrpc.UnsafeChatHelperServiceServer
}

func NewChatHelperApi(api *API) *chatHelper {
	return &chatHelper{API: api}
}

func (api *chatHelper) Broadcast(ctx context.Context, in *engine.BroadcastRequest) (*engine.BroadcastResponse, error) {
	session, err := api.ctrl.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	msg := chat.Message{
		Text: in.GetText(),
	}

	if in.File != nil {
		msg.File = &chat.File{
			Id:   in.File.GetId(),
			Url:  in.File.GetUrl(),
			Mime: in.File.GetMime(),
			Name: in.File.GetName(),
			Size: in.File.GetSize(),
		}
	}

	y := len(in.Buttons)

	if y != 0 {
		msg.Buttons = make([]*chat.Buttons, 0, y)
		for _, v := range in.Buttons {
			if v.Button != nil {
				x := make([]*chat.Button, 0, len(v.Button))
				for _, b := range v.Button {
					x = append(x, &chat.Button{
						Caption: b.Caption,
						Text:    b.Text,
						Type:    b.Type,
						Url:     b.Url,
						Code:    b.Code,
					})
				}
				msg.Buttons = append(msg.Buttons, &chat.Buttons{
					Button: x,
				})
			}

		}
	}

	err = api.ctrl.BroadcastChatBot(session, in.GetProfileId(), in.GetPeer(), &msg)
	if err != nil {
		return nil, err
	}

	return &engine.BroadcastResponse{}, nil
}
