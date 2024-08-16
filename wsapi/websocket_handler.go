package wsapi

import (
	"fmt"
	"time"

	"github.com/webitel/engine/app"
	"github.com/webitel/engine/localization"
	"github.com/webitel/engine/model"
	"github.com/webitel/wlog"
)

func (api *API) ApiWebSocketHandler(wh func(*app.WebConn, *model.WebSocketRequest) (map[string]interface{}, model.AppError)) webSocketHandler {
	return webSocketHandler{api.App, wh, false}
}
func (api *API) ApiAsyncWebSocketHandler(wh func(*app.WebConn, *model.WebSocketRequest) (map[string]interface{}, model.AppError)) webSocketHandler {
	return webSocketHandler{api.App, wh, true}
}

type webSocketHandler struct {
	app         *app.App
	handlerFunc func(*app.WebConn, *model.WebSocketRequest) (map[string]interface{}, model.AppError)
	async       bool
}

func (wh webSocketHandler) ServeWebSocket(conn *app.WebConn, r *model.WebSocketRequest) {
	start := time.Now()
	session, sessionErr := wh.app.GetSession(conn.GetSessionToken())
	if sessionErr != nil {
		wlog.Error(fmt.Sprintf("%v:%v seq=%v uid=%v %v [details: %v]", "websocket", r.Action, r.Seq, conn.UserId, sessionErr.SystemMessage(localization.T), sessionErr.Error()))
		sessionErr.SetDetailedError("")
		errResp := model.NewWebSocketError(r.Seq, sessionErr)

		conn.Send <- errResp
		//conn.Close()
		return
	}
	r.Session = *session

	r.T = conn.T
	r.Locale = conn.Locale

	var data map[string]interface{}
	var err model.AppError

	if wh.async {

	}

	if data, err = wh.handlerFunc(conn, r); err != nil {
		conn.Log().With(
			wlog.String("method", r.Action),
			wlog.Float64("duration_ms", float64(time.Since(start).Microseconds())/1000),
		).Error(err.Error(), wlog.Err(err))
		//err.DetailedError = ""
		errResp := model.NewWebSocketError(r.Seq, err)

		conn.Send <- errResp
		return
	}

	resp := model.NewWebSocketResponse(model.STATUS_OK, r.Seq, data)
	conn.Log().With(
		wlog.String("method", r.Action),
		wlog.Float64("duration_ms", float64(time.Since(start).Microseconds())/1000),
	).Debug("send response action " + r.Action)

	conn.Send <- resp
}

func NewInvalidWebSocketParamError(action string, name string) model.AppError {
	return model.NewBadRequestError("api.websocket_handler.invalid_param.app_error", "").SetTranslationParams(map[string]interface{}{"Name": name})
}
