package utils

import (
	"github.com/webitel/engine/model"
	"google.golang.org/grpc/status"
)

func DomainErorrFromGRPC(err error) model.AppError {
	st, ok := status.FromError(err)
	if !ok {
		return nil
	}

	jsonString := st.Message()
	if werr := model.AppErrorFromJson(jsonString); werr != nil {
		return werr
	}

	return nil
}
