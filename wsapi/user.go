package wsapi

import (
	"github.com/webitel/engine/app"
	"github.com/webitel/engine/model"
)

func (api *API) InitUser() {
	api.Router.Handle("user_typing", api.ApiWebSocketHandler(api.userTyping))
	api.Router.Handle("user_default_device", api.ApiWebSocketHandler(api.userDefaultDeviceConfig))
}

func (api *API) userTyping(conn *app.WebConn, req *model.WebSocketRequest) (map[string]interface{}, *model.AppError) {
	//return nil, NewInvalidWebSocketParamError(req.Action, "channel_id")

	data := map[string]interface{}{}
	data["text"] = "pong"
	data["server_time"] = model.GetMillis()
	data["node_id"] = ""
	return data, nil
}

func (api *API) userDefaultDeviceConfig(conn *app.WebConn, req *model.WebSocketRequest) (map[string]interface{}, *model.AppError) {
	config, err := api.App.GetUserDefaultDeviceConfig(conn.GetSession().UserId, conn.GetSession().DomainId)
	if err != nil {
		return nil, err
	}
	return config.ToMap(), nil
}
