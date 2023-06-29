package wsapi

import (
	"context"

	"github.com/webitel/engine/app"
	"github.com/webitel/engine/model"
)

func (api *API) InitDevice() {
	api.Router.Handle("device_default", api.ApiWebSocketHandler(api.deviceDefault))
}

func (api *API) deviceDefault(conn *app.WebConn, req *model.WebSocketRequest) (map[string]interface{}, model.AppError) {
	//return nil, NewInvalidWebSocketParamError(req.Action, "channel_id")
	//d := &DeviceConfig{
	//	Realm:             "webitel.lo",
	//	Uri:               "7005",
	//	AuthorizationUser: "user",
	//	Ha1:               "865011debb10e1a281d090499180483d",
	//}

	typeName, _ := req.Data["name"].(string)
	config, err := api.App.GetUserDefaultDeviceConfig(context.TODO(), conn.GetSession().UserId, conn.GetSession().DomainId, typeName)
	if err != nil {
		return nil, err
	}
	return config, nil
}
