package app

import (
	"github.com/webitel/engine/model"
	"github.com/webitel/engine/pkg/werror"
	"google.golang.org/grpc/status"
)

func errorHandlerFromApi(err error) model.AppError {
	protoErr, ok := status.FromError(err)
	if ok {
		return werror.AppErrorFromJson(protoErr.Message())
	}

	return model.NewInternalError("app.error", err.Error())
}
