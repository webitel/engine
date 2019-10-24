package app

import "github.com/webitel/engine/model"

func (app *App) GetUserCallInfo(userId, domainId int64) (*model.UserCallInfo, *model.AppError) {
	return app.Store.User().GetCallInfo(userId, domainId)
}
