package external_commands

import (
	"github.com/webitel/engine/model"
)

type AuthClient interface {
	Name() string
	Close() error
	Ready() bool
	GetSession(token string) (*model.Session, *model.AppError)
}
