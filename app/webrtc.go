package app

import (
	"context"
	"github.com/webitel/engine/b2bua"
	"github.com/webitel/engine/b2bua/account"
	"github.com/webitel/engine/model"
)

func (app *App) OnB2B(sipId string, sdp b2bua.SdpDescription) {
	h, _ := app.Hubs.Get(1)
	e := model.NewWebSocketEvent("sdp")
	e.Data = map[string]interface{}{
		"sip_id": sipId,
		"sdp":    sdp,
	}
	h.broadcast <- e
}

func (app *App) Dial(userId int, sdp string, destination string) {
	app.b2b.Dial(userId, sdp, destination)
}

func (app *App) SipDial(userId int, sdp string, destination string) (string, model.AppError) {
	sipId, rErr := app.b2b.Dial(userId, sdp, destination)
	if rErr != nil {
		return "", model.NewInternalError("app.sip.dial.app_err", rErr.Error())
	}

	return sipId, nil
}

func (app *App) SipRemoteSdp(userId int, wid string) (b2bua.SdpDescription, model.AppError) {
	sdp, rErr := app.b2b.RemoteSdp(userId, wid)
	if rErr != nil {
		return b2bua.SdpDescription{}, model.NewInternalError("app.sip.remote_sdp.app_err", rErr.Error())
	}

	return sdp, nil
}

func (app *App) SipRecovery(domainId int64, userId int64, callId string, sdp string) (string, model.AppError) {
	sipId, err := app.Store.Call().GetSipId(context.Background(), domainId, userId, callId)
	if err != nil {
		return "", err
	}

	_, rErr := app.b2b.Recovery(int(userId), sipId, sdp)
	if rErr != nil {
		return "", model.NewInternalError("app.sip.recovery.app_err", rErr.Error())
	}

	return sipId, nil
}

func (app *App) SipAnswer(domainId int64, userId int64, callId string, sdp string) (string, model.AppError) {
	remSdp, rErr := app.b2b.Answer(int(userId), callId, sdp)
	if rErr != nil {
		return "", model.NewInternalError("app.sip.answer.app_err", rErr.Error())
	}

	return remSdp, nil
}

func (app *App) SipRegister(ctx context.Context, domainId, userId int64) model.AppError {
	sipConf, appErr := app.GetUserDefaultSipCDeviceConfig(ctx, userId, domainId)
	if appErr != nil {
		return appErr
	}

	err := app.b2b.Register(int(userId), b2bua.AuthInfo{
		DisplayName: "igor",
		Expires:     3200,
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
