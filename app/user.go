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

func (app *App) GetCallInfoEndpoint(domainId int64, e *model.EndpointRequest, isOnline bool) (*model.UserCallInfo, *model.AppError) {
	return app.Store.User().GetCallInfoEndpoint(domainId, e, isOnline)
}

func (app *App) GetUserDefaultWebRTCDeviceConfig(userId, domainId int64) (*model.UserDeviceConfig, *model.AppError) {
	conf, err := app.Store.User().DefaultWebRTCDeviceConfig(userId, domainId)
	if err != nil {
		return nil, err
	}
	conf.Server = app.CallManager().SipWsAddress()
	return conf, nil
}

func (app *App) GetUserDefaultSipCDeviceConfig(userId, domainId int64) (*model.UserSipDeviceConfig, *model.AppError) {
	conf, err := app.Store.User().DefaultSipDeviceConfig(userId, domainId)
	if err != nil {
		return nil, err
	}

	if app.config.SipSettings.PublicProxy != "" {
		conf.Proxy = app.config.SipSettings.PublicProxy
	} else {
		conf.Proxy = app.CallManager().SipRouteUri()
	}
	return conf, nil
}

func (app *App) GetUserDefaultDeviceConfig(userId, domainId int64, typeName string) (map[string]interface{}, *model.AppError) {
	if typeName == model.DeviceTypeSip {
		if res, err := app.GetUserDefaultSipCDeviceConfig(userId, domainId); err != nil {
			return nil, err
		} else {
			return res.ToMap(), nil
		}
	} else {
		if res, err := app.GetUserDefaultWebRTCDeviceConfig(userId, domainId); err != nil {
			return nil, err
		} else {
			return res.ToMap(), nil
		}
	}
}
