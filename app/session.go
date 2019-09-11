package app

import (
	"github.com/webitel/engine/model"
)

func (app *App) GetSession(token string) (*model.Session, *model.AppError) {
	return app.sessionManager.GetSession(token)
}
