package wsapi

import (
	"context"
	"github.com/webitel/engine/app"
	"github.com/webitel/engine/model"
)

func (api *API) InitScreenShare() {
	api.Router.Handle("ss_invite", api.ApiWebSocketHandler(api.ssInvite))
	api.Router.Handle("ss_accept", api.ApiWebSocketHandler(api.ssAccept))
}

func (api *API) ssInvite(ctx context.Context, conn *app.WebConn, req *model.WebSocketRequest) (map[string]interface{}, model.AppError) {
	var toUserId float64
	var sdp, id string
	var ok bool

	if toUserId, ok = req.Data["to_user_id"].(float64); !ok {
		return nil, NewInvalidWebSocketParamError("ss_invite", "to_user_id")
	}
	if sdp, ok = req.Data["sdp"].(string); !ok {
		return nil, NewInvalidWebSocketParamError("ss_invite", "sdp")
	}
	if id, ok = req.Data["id"].(string); !ok {
		return nil, NewInvalidWebSocketParamError("ss_invite", "id")
	}

	err := api.App.RequestScreenShare(conn.DomainId, conn.UserId, int64(toUserId), conn.Id(), sdp, id)
	if err != nil {
		return nil, err
	}
	return map[string]any{}, nil
}

func (api *API) ssAccept(ctx context.Context, conn *app.WebConn, req *model.WebSocketRequest) (map[string]interface{}, model.AppError) {
	var toUserId float64
	var sdp, sockId, sess string
	var ok bool

	if toUserId, ok = req.Data["to_user_id"].(float64); !ok {
		return nil, NewInvalidWebSocketParamError("ss_accept", "to_user_id")
	}
	if sdp, ok = req.Data["sdp"].(string); !ok {
		return nil, NewInvalidWebSocketParamError("ss_accept", "sdp")
	}
	if sockId, ok = req.Data["sock_id"].(string); !ok {
		return nil, NewInvalidWebSocketParamError("ss_accept", "sock_id")
	}
	if sess, ok = req.Data["session_id"].(string); !ok {
		return nil, NewInvalidWebSocketParamError("ss_accept", "session_id")
	}

	err := api.App.AcceptScreenShare(conn.DomainId, int64(toUserId), sockId, sess, sdp)
	if err != nil {
		return nil, err
	}
	return make(map[string]interface{}), nil
}
