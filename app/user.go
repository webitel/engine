package app

import (
	"context"

	"github.com/webitel/engine/model"
	"github.com/webitel/engine/pkg/wbt/auth_manager"
)

func (a *App) UserCheckAccess(ctx context.Context, domainId, id int64, groups []int, access auth_manager.PermissionAccess) (bool, model.AppError) {
	return a.Store.User().CheckAccess(ctx, domainId, id, groups, access)
}

func (app *App) GetUserCallInfo(ctx context.Context, userId, domainId int64) (*model.UserCallInfo, model.AppError) {
	return app.Store.User().GetCallInfo(ctx, userId, domainId)
}

func (app *App) GetCallInfoEndpoint(ctx context.Context, domainId int64, e *model.EndpointRequest, isOnline bool) (*model.UserCallInfo, model.AppError) {
	return app.Store.User().GetCallInfoEndpoint(ctx, domainId, e, isOnline)
}

func (app *App) GetUserDefaultWebRTCDeviceConfig(ctx context.Context, userId, domainId int64) (*model.UserDeviceConfig, model.AppError) {
	conf, err := app.Store.User().DefaultWebRTCDeviceConfig(ctx, userId, domainId)
	if err != nil {
		return nil, err
	}
	conf.Server = app.CallManager().SipWsAddress()
	return conf, nil
}

func (app *App) GetUserDefaultSipCDeviceConfig(ctx context.Context, userId, domainId int64) (*model.UserSipDeviceConfig, model.AppError) {
	conf, err := app.Store.User().DefaultSipDeviceConfig(ctx, userId, domainId)
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

func (app *App) GetUserDefaultDeviceConfig(ctx context.Context, userId, domainId int64, typeName string) (map[string]interface{}, model.AppError) {
	if typeName == model.DeviceTypeSip {
		if res, err := app.GetUserDefaultSipCDeviceConfig(ctx, userId, domainId); err != nil {
			return nil, err
		} else {
			return res.ToMap(), nil
		}
	} else {
		if res, err := app.GetUserDefaultWebRTCDeviceConfig(ctx, userId, domainId); err != nil {
			return nil, err
		} else {
			return res.ToMap(), nil
		}
	}
}

func (app *App) GetWebSocketsPage(ctx context.Context, domainId int64, search *model.SearchSocketSessionView) ([]*model.SocketSessionView, bool, model.AppError) {
	list, err := app.Store.SocketSession().Search(ctx, domainId, search)
	if err != nil {
		return nil, false, err
	}
	search.RemoveLastElemIfNeed(&list)
	return list, search.EndOfList(), nil
}
