package app

import (
	"context"
	"github.com/webitel/engine/b2bua"
	"github.com/webitel/engine/b2bua/account"
	"github.com/webitel/engine/model"
	"github.com/webitel/wlog"
)

var ErrDisabledB2b = model.NewBadRequestError("app.b2b.disabled", "B2B disabled")

func (app *App) OnB2B(sockId string, domainId int64, userId int64, sipId string, sdp b2bua.SdpDescription) {
	h, ok := app.Hubs.Get(domainId)
	if !ok {
		wlog.Error("not found domain")
		return
	}

	e := model.NewWebSocketEvent("sdp")
	e.UserId = userId
	e.SockId = sockId
	e.Data = map[string]interface{}{
		"sip_id": sipId,
		"sdp":    sdp,
	}
	h.broadcast <- e
}

func (app *App) Dial(sockId string, domainId int64, userId int64, sdp string, destination string) {
	if app.b2b == nil {
		return
	}
	app.b2b.Dial(sockId, domainId, userId, sdp, destination)
}

func (app *App) SipDial(sockId string, domainId int64, userId int64, sdp string, destination string) (string, model.AppError) {
	if app.b2b == nil {
		return "", ErrDisabledB2b
	}

	sipId, rErr := app.b2b.Dial(sockId, domainId, userId, sdp, destination)
	if rErr != nil {
		return "", model.NewInternalError("app.sip.dial.app_err", rErr.Error())
	}

	return sipId, nil
}

func (app *App) SipRemoteSdp(userId int64, wid string) (b2bua.SdpDescription, model.AppError) {
	if app.b2b == nil {
		return b2bua.SdpDescription{}, ErrDisabledB2b
	}

	sdp, rErr := app.b2b.RemoteSdp(userId, wid)
	if rErr != nil {
		return b2bua.SdpDescription{}, model.NewInternalError("app.sip.remote_sdp.app_err", rErr.Error())
	}

	return sdp, nil
}

func (app *App) SipRecovery(sockId string, domainId int64, userId int64, callId string, sdp string) (string, model.AppError) {
	if app.b2b == nil {
		return "", ErrDisabledB2b
	}

	sipId, err := app.Store.Call().GetSipId(context.Background(), domainId, userId, callId)
	if err != nil {
		return "", err
	}

	_, rErr := app.b2b.Recovery(sockId, userId, sipId, sdp)
	if rErr != nil {
		return "", model.NewInternalError("app.sip.recovery.app_err", rErr.Error())
	}

	return sipId, nil
}

func (app *App) SipAnswer(domainId int64, userId int64, callId string, sdp string) (string, model.AppError) {
	if app.b2b == nil {
		return "", ErrDisabledB2b
	}

	remSdp, rErr := app.b2b.Answer(int(userId), callId, sdp)
	if rErr != nil {
		return "", model.NewInternalError("app.sip.answer.app_err", rErr.Error())
	}

	return remSdp, nil
}

func (app *App) SipRegister(ctx context.Context, name string, domainId, userId int64) model.AppError {
	if app.b2b == nil {
		return ErrDisabledB2b
	}

	sipConf, appErr := app.GetUserDefaultSipCDeviceConfig(ctx, userId, domainId)
	if appErr != nil {
		return appErr
	}

	err := app.b2b.Register(userId, b2bua.AuthInfo{
		DisplayName: name,
		Expires:     3200,
		DomainId:    domainId,
		UserId:      userId,
		AuthInfo: account.AuthInfo{
			AuthUser: sipConf.Auth,
			Realm:    sipConf.Domain,
			Password: sipConf.Password,
		},
	})
	if err != nil {
		return model.NewInternalError("app.sip.register.app_err", err.Error())
	}

	return nil
}
