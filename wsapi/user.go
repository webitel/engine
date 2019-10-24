package wsapi

import (
	"github.com/webitel/engine/app"
	"github.com/webitel/engine/model"
)

func (api *API) InitUser() {
	api.Router.Handle("user_typing", api.ApiWebSocketHandler(api.userTyping))
}

func (api *API) userTyping(conn *app.WebConn, req *model.WebSocketRequest) (map[string]interface{}, *model.AppError) {
	//return nil, NewInvalidWebSocketParamError(req.Action, "channel_id")

	data := map[string]interface{}{}
	data["text"] = "pong"
	data["server_time"] = model.GetMillis()
	data["node_id"] = ""
	return data, nil
}
