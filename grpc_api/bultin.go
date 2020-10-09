package grpc_api

import (
	"github.com/webitel/engine/model"
	"github.com/webitel/protos/engine"
)

var ResponseOk = &engine.Response{
	Status: model.STATUS_OK,
}
