package wsapi

import (
	"context"
	"fmt"
	"go.opentelemetry.io/otel/attribute"
	"time"

	"github.com/webitel/engine/app"
	"github.com/webitel/engine/model"
	"github.com/webitel/wlog"
)

func (api *API) ApiWebSocketHandler(wh func(context.Context, *app.WebConn, *model.WebSocketRequest) (map[string]interface{}, model.AppError)) webSocketHandler {
	return webSocketHandler{api.App, wh, false}
}
func (api *API) ApiAsyncWebSocketHandler(wh func(context.Context, *app.WebConn, *model.WebSocketRequest) (map[string]interface{}, model.AppError)) webSocketHandler {
	return webSocketHandler{api.App, wh, true}
}

type webSocketHandler struct {
	app         *app.App
	handlerFunc func(context.Context, *app.WebConn, *model.WebSocketRequest) (map[string]interface{}, model.AppError)
	async       bool
}

func (wh webSocketHandler) ServeWebSocket(conn *app.WebConn, r *model.WebSocketRequest) {
	start := time.Now()
	session, sessionErr := wh.app.GetSessionWitchContext(conn.Ctx, conn.GetSessionToken())
	if sessionErr != nil {
		wlog.Error(fmt.Sprintf("%v:%v seq=%v uid=%v %v [details: %v]", "websocket", r.Action, r.Seq, conn.UserId, sessionErr.GetDetailedError(), sessionErr.Error()))
		sessionErr.SetDetailedError("")
		errResp := model.NewWebSocketError(r.Seq, sessionErr)

		conn.Send <- errResp
		//conn.Close()
		return
	}
	r.Session = *session

	r.Locale = conn.Locale

	ctx, span := wh.app.Tracer().Start(conn.Ctx, r.Action)
	defer span.End()

	span.SetAttributes(
		attribute.Int64("domain_id", session.DomainId),
		attribute.Int64("user_id", session.UserId),
		attribute.String("ip_address", session.GetUserIp()),
		attribute.String("method", r.Action),
		attribute.String("sock_id", conn.Id()),
	)

	var data map[string]interface{}
	var err model.AppError

	if data, err = wh.handlerFunc(ctx, conn, r); err != nil {
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
