package wsapi

import (
	"context"
	"fmt"

	"github.com/webitel/webitel-go-kit/logging/wlog"

	"github.com/webitel/engine/app"
	"github.com/webitel/engine/model"
)

func (api *API) InitUser() {
	api.Router.Handle("user_typing", api.ApiWebSocketHandler(api.userTyping))
	api.Router.Handle("user_default_device", api.ApiWebSocketHandler(api.userDefaultDeviceConfig))

	api.Router.Handle("subscribe_users_status", api.ApiWebSocketHandler(api.subscribeUsersStatus))
	api.Router.Handle("ping", api.ApiWebSocketHandler(api.ping))

	api.Router.Handle("latency_start", api.ApiWebSocketHandler(api.latencyStart))
	api.Router.Handle("latency_ack", api.ApiWebSocketHandler(api.latencyAck))
}

func (api *API) userTyping(conn *app.WebConn, req *model.WebSocketRequest) (map[string]interface{}, model.AppError) {
	// return nil, NewInvalidWebSocketParamError(req.Action, "channel_id")

	data := map[string]interface{}{}
	data["text"] = "pong"
	data["server_time"] = model.GetMillis()
	data["node_id"] = ""
	return data, nil
}

func (api *API) userDefaultDeviceConfig(conn *app.WebConn, req *model.WebSocketRequest) (map[string]interface{}, model.AppError) {
	typeName, _ := req.Data["name"].(string)
	config, err := api.App.GetUserDefaultDeviceConfig(context.TODO(), conn.GetSession().UserId, conn.GetSession().DomainId, typeName)
	if err != nil {
		return nil, err
	}
	return config, nil
}

func (api *API) subscribeUsersStatus(conn *app.WebConn, req *model.WebSocketRequest) (map[string]interface{}, model.AppError) {
	h, e := api.App.GetHubById(req.Session.Domain(0)) // FIXME
	if e != nil {
		return nil, e
	}

	return nil, h.SubscribeSessionUsersStatus(conn)
}

func (api *API) ping(conn *app.WebConn, req *model.WebSocketRequest) (map[string]interface{}, model.AppError) {
	data := map[string]interface{}{}
	data["pong"] = 1
	if api.pingClientLatency {
		data["server_ts"] = model.GetMillis()
	}
	return data, nil
}

func (api *API) latencyStart(conn *app.WebConn, req *model.WebSocketRequest) (map[string]interface{}, model.AppError) {
	return map[string]interface{}{
		"server_ts": model.GetMillis(),
	}, nil
}

func (api *API) latencyAck(conn *app.WebConn, req *model.WebSocketRequest) (map[string]interface{}, model.AppError) {
	t := model.GetMillis()
	req.Data["server_ack_ts"] = t
	if v, ok := req.Data["last_latency"].(float64); ok && v > 0 {
		old := conn.SetLastLatencyTime(t)
		if old > 0 {
			wlog.Debug(fmt.Sprintf("[%s] user_id=%d sock=%s latency=%f diff=%ds", conn.Ip(), conn.UserId, conn.Id(), v, (t-old)/1000))
		} else {
			wlog.Debug(fmt.Sprintf("[%s] user_id=%d sock=%s latency=%f", conn.Ip(), conn.UserId, conn.Id(), v))
		}

	}
	return req.Data, nil
}
