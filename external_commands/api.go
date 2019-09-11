package external_commands

import (
	"github.com/webitel/engine/external_commands/grpc"
	"github.com/webitel/engine/model"
)

func NewAuthServiceConnection(name, url string) (model.AuthClient, *model.AppError) {
	return grpc.NewAuthServiceConnection(name, url)
}
