package grpc_api

import (
	"github.com/webitel/engine/gen/engine"
	"github.com/webitel/engine/model"
)

var ResponseOk = &engine.Response{
	Status: model.STATUS_OK,
}
