package grpc_api

import (
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

	err = api.ctrl.BroadcastChatBot(session, in.GetProfileId(), in.GetPeer(), in.GetText())
	if err != nil {
		return nil, err
	}

	return &engine.BroadcastResponse{}, nil
}
