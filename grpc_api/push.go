package grpc_api

import (
	gogrpc "buf.build/gen/go/webitel/engine/grpc/go/_gogrpc"
	engine "buf.build/gen/go/webitel/engine/protocolbuffers/go"
	"context"
	"github.com/webitel/engine/model"
)

type push struct {
	*API
	minimumNumberMaskLen int
	prefixNumberMaskLen  int
	suffixNumberMaskLen  int
	gogrpc.UnsafePushServiceServer
}

func NewPushApi(api *API, minimumNumberMaskLen, prefixNumberMaskLen, suffixNumberMaskLen int) *push {
	return &push{
		API:                  api,
		minimumNumberMaskLen: minimumNumberMaskLen,
		prefixNumberMaskLen:  prefixNumberMaskLen,
		suffixNumberMaskLen:  suffixNumberMaskLen,
	}
}

func (api *push) SendPush(ctx context.Context, in *engine.SendPushRequest) (*engine.SendPushResponse, error) {
	m := in.Data
	// TODO WMA-84
	if tmp, ok := m["hide_number"]; ok && tmp == "true" {
		tmp, _ = m["from_number"]
		m["from_number"] = model.HideString(tmp,
			api.minimumNumberMaskLen,
			api.prefixNumberMaskLen,
			api.suffixNumberMaskLen,
		)
	}

	c, err := api.app.SendPush(ctx, &model.SendPush{
		Android:    in.Android,
		Apple:      in.Apple,
		Data:       m,
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
