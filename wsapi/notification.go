package wsapi

import (
	"context"
	"github.com/webitel/engine/app"
	"github.com/webitel/engine/model"
)

func (api *API) InitNotification() {
	api.Router.Handle("notification_send", api.ApiWebSocketHandler(api.sendNotification))
}

func (api *API) sendNotification(conn *app.WebConn, req *model.WebSocketRequest) (map[string]interface{}, *model.AppError) {
	desc, _ := req.Data["description"].(string)
	action, _ := req.Data["action"].(string)

	err := api.App.SendNotification(context.Background(), conn.DomainId, &conn.UserId, []int64{1}, action, desc)
	if err != nil {
		return nil, err
	}

	return nil, nil
}
