package wsapi

import (
	"context"
	"github.com/webitel/engine/app"
	"github.com/webitel/engine/model"
)

func (api *API) InitDevice() {
	api.Router.Handle("device_default", api.ApiWebSocketHandler(api.deviceDefault))
}

func (api *API) deviceDefault(ctx context.Context, conn *app.WebConn, req *model.WebSocketRequest) (map[string]interface{}, model.AppError) {
	typeName, _ := req.Data["name"].(string)
	config, err := api.App.GetUserDefaultDeviceConfig(ctx, conn.GetSession().UserId, conn.GetSession().DomainId, typeName)
	if err != nil {
		return nil, err
	}
	return config, nil
}
