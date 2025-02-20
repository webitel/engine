package grpc_api

import (
	gogrpc "buf.build/gen/go/webitel/engine/grpc/go/_gogrpc"
	engine "buf.build/gen/go/webitel/engine/protocolbuffers/go"
	"context"
	"errors"
)

type chatHelper struct {
	*API
	gogrpc.UnsafeChatHelperServiceServer
}

func NewChatHelperApi(api *API) *chatHelper {
	return &chatHelper{API: api}
}

func (api *chatHelper) Broadcast(ctx context.Context, in *engine.BroadcastRequest) (*engine.BroadcastResponse, error) {
	return nil, errors.New("deprecated")
}
