package grpc_api

import (
	"context"
	"github.com/webitel/protos/engine"
)

type chatHelper struct {
	*API
	engine.UnsafeChatHelperServiceServer
}

func NewChatHelperApi(api *API) *chatHelper {
	return &chatHelper{API: api}
}

func (api *chatHelper) Broadcast(ctx context.Context, in *engine.BroadcastRequest) (*engine.BroadcastResponse, error) {
	session, err := api.ctrl.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	err = api.ctrl.BroadcastChatBot(session, in.GetProfileId(), in.GetPeer(), in.GetText())
	if err != nil {
		return nil, err
	}

	return &engine.BroadcastResponse{}, nil
}
