package grpc_api

import (
	gogrpc "buf.build/gen/go/webitel/engine/grpc/go/_gogrpc"
	engine "buf.build/gen/go/webitel/engine/protocolbuffers/go"
	"context"
	"github.com/webitel/engine/model"
)

type push struct {
	*API
	gogrpc.UnsafePushServiceServer
}

func NewPushApi(api *API) *push {
	return &push{API: api}
}

func (p push) SendPush(ctx context.Context, in *engine.SendPushRequest) (*engine.SendPushResponse, error) {
	c, err := p.app.SendPush(ctx, &model.SendPush{
		Android:    in.Android,
		Apple:      in.Apple,
		Data:       in.Data,
		Expiration: in.Expiration,
		Priority:   in.Priority,
	})

	if err != nil {
		return nil, err
	}

	return &engine.SendPushResponse{
		Send: int32(c),
	}, nil
}
