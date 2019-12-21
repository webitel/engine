package wsapi

import (
	"encoding/json"
	"github.com/webitel/engine/app"
	"github.com/webitel/engine/model"
)

func (api *API) InitDevice() {
	api.Router.Handle("device_default", api.ApiWebSocketHandler(api.deviceDefault))
}

type DeviceConfig struct {
	Realm             string `json:"realm"`
	Uri               string `json:"uri"`
	AuthorizationUser string `json:"authorization_user"`
	Ha1               string `json:"ha1"`
}

func (d DeviceConfig) ToMap() map[string]interface{} {
	out := make(map[string]interface{})
	data, _ := json.Marshal(d)
	_ = json.Unmarshal(data, &out)
	return out
}

func (api *API) deviceDefault(conn *app.WebConn, req *model.WebSocketRequest) (map[string]interface{}, *model.AppError) {
	//return nil, NewInvalidWebSocketParamError(req.Action, "channel_id")
	d := &DeviceConfig{
		Realm:             "",
		Uri:               "",
		AuthorizationUser: "",
		Ha1:               "",
	}

	return d.ToMap(), nil
}
