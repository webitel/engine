package grpc_api

import (
	engine "buf.build/gen/go/webitel/engine/protocolbuffers/go"
	"github.com/webitel/engine/model"
)

var ResponseOk = &engine.Response{
	Status: model.STATUS_OK,
}
