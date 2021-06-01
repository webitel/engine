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

func (app *App) GetCallInfoEndpoint(domainId int64, e *model.EndpointRequest) (*model.UserCallInfo, *model.AppError) {
	return app.Store.User().GetCallInfoEndpoint(domainId, e)
}

func (app *App) GetUserDefaultWebRTCDeviceConfig(userId, domainId int64) (map[string]interface{}, *model.AppError) {
	conf, err := app.Store.User().DefaultWebRTCDeviceConfig(userId, domainId)
	if err != nil {
		return nil, err
	}
	conf.Server = app.CallManager().SipWsAddress()
	return conf.ToMap(), nil
}

func (app *App) GetUserDefaultSipCDeviceConfig(userId, domainId int64) (map[string]interface{}, *model.AppError) {
	conf, err := app.Store.User().DefaultSipDeviceConfig(userId, domainId)
	if err != nil {
		return nil, err
	}
	conf.Proxy = app.CallManager().SipRouteUri()
	return conf.ToMap(), nil
}

func (app *App) GetUserDefaultDeviceConfig(userId, domainId int64, typeName string) (map[string]interface{}, *model.AppError) {
	if typeName == model.DeviceTypeSip {
		return app.GetUserDefaultSipCDeviceConfig(userId, domainId)
	} else {
		return app.GetUserDefaultWebRTCDeviceConfig(userId, domainId)
	}
}
