package wsapi

import (
	"fmt"
	"github.com/webitel/engine/app"
	"github.com/webitel/engine/localization"
	"github.com/webitel/engine/model"
	"github.com/webitel/wlog"
	"net/http"
)

func (api *API) ApiWebSocketHandler(wh func(*app.WebConn, *model.WebSocketRequest) (map[string]interface{}, *model.AppError)) webSocketHandler {
	return webSocketHandler{api.App, wh, false}
}
func (api *API) ApiAsyncWebSocketHandler(wh func(*app.WebConn, *model.WebSocketRequest) (map[string]interface{}, *model.AppError)) webSocketHandler {
	return webSocketHandler{api.App, wh, true}
}

type webSocketHandler struct {
	app         *app.App
	handlerFunc func(*app.WebConn, *model.WebSocketRequest) (map[string]interface{}, *model.AppError)
	async       bool
}

func (wh webSocketHandler) ServeWebSocket(conn *app.WebConn, r *model.WebSocketRequest) {
	wlog.Debug(fmt.Sprintf("[%s] websock (%s) method %s", conn.Ip(), conn.Id(), r.Action))

	session, sessionErr := wh.app.GetSession(conn.GetSessionToken())
	if sessionErr != nil {
		wlog.Error(fmt.Sprintf("%v:%v seq=%v uid=%v %v [details: %v]", "websocket", r.Action, r.Seq, conn.UserId, sessionErr.SystemMessage(localization.T), sessionErr.Error()))
		sessionErr.DetailedError = ""
		errResp := model.NewWebSocketError(r.Seq, sessionErr)

		conn.Send <- errResp
		//conn.Close()
		return
	}
	r.Session = *session

	r.T = conn.T
	r.Locale = conn.Locale

	var data map[string]interface{}
	var err *model.AppError

	if wh.async {

	}

	if data, err = wh.handlerFunc(conn, r); err != nil {
		wlog.Error(fmt.Sprintf("%v %v seq=%vq [details: %v]", "websocket", r.Action, r.Seq, err.Error()))
		//err.DetailedError = ""
		errResp := model.NewWebSocketError(r.Seq, err)

		conn.Send <- errResp
		return
	}

	resp := model.NewWebSocketResponse(model.STATUS_OK, r.Seq, data)

	conn.Send <- resp
}

func NewInvalidWebSocketParamError(action string, name string) *model.AppError {
	return model.NewAppError("websocket: "+action, "api.websocket_handler.invalid_param.app_error", map[string]interface{}{"Name": name}, "", http.StatusBadRequest)
}
