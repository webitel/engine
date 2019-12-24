package app

import "github.com/webitel/engine/model"

func (app *App) GetUserCallInfo(userId, domainId int64) (*model.UserCallInfo, *model.AppError) {
	return app.Store.User().GetCallInfo(userId, domainId)
}

func (app *App) GetUserDefaultDeviceConfig(userId, domainId int64) (*model.UserDeviceConfig, *model.AppError) {
	conf, err := app.Store.User().DefaultDeviceConfig(userId, domainId)
	if err != nil {
		return nil, err
	}
	conf.Server = app.CallManager().SipWsAddress()
	return conf, nil
}
