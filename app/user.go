package app

import (
	"github.com/webitel/engine/auth_manager"
	"github.com/webitel/engine/model"
)

func (a *App) UserCheckAccess(domainId, id int64, groups []int, access auth_manager.PermissionAccess) (bool, *model.AppError) {
	return a.Store.User().CheckAccess(domainId, id, groups, access)
}

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
