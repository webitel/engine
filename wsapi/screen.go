package wsapi

import (
	"context"
	"github.com/webitel/engine/app"
	"github.com/webitel/engine/model"
)

func (api *API) InitScreenShare() {
	api.Router.Handle("ss_invite", api.ApiWebSocketHandler(api.ssInvite))
	api.Router.Handle("ss_accept", api.ApiWebSocketHandler(api.ssAccept))
	api.Router.Handle("screenshot", api.ApiWebSocketHandler(api.screenshot))
	api.Router.Handle("start_record_screen", api.ApiWebSocketHandler(api.ssStartRecord))
	api.Router.Handle("stop_record_screen", api.ApiWebSocketHandler(api.ssStopRecord))
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

	spySockId, err := api.ctrl.RequestScreenShare(ctx, conn.GetSession(), conn.UserId, int64(toUserId), conn.Id(), sdp, id)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"sock_id": spySockId,
	}, nil
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

	err := api.ctrl.AcceptScreenShare(ctx, conn.GetSession(), int64(toUserId), sockId, sess, sdp, conn.Id())
	if err != nil {
		return nil, err
	}
	return make(map[string]interface{}), nil
}

func (api *API) screenshot(ctx context.Context, conn *app.WebConn, req *model.WebSocketRequest) (map[string]interface{}, model.AppError) {
	var toUserId float64
	var ok bool

	if toUserId, ok = req.Data["to_user_id"].(float64); !ok {
		return nil, NewInvalidWebSocketParamError("screenshot", "to_user_id")
	}

	err := api.ctrl.Screenshot(ctx, conn.GetSession(), int64(toUserId), conn.Id())
	if err != nil {
		return nil, err
	}

	return make(map[string]interface{}), nil
}

func (api *API) ssStartRecord(ctx context.Context, conn *app.WebConn, req *model.WebSocketRequest) (map[string]interface{}, model.AppError) {
	var toUserId float64
	var sessionId string
	var ok bool

	if toUserId, ok = req.Data["to_user_id"].(float64); !ok {
		return nil, NewInvalidWebSocketParamError("start_record_screen", "to_user_id")
	}

	if sessionId, ok = req.Data["root_id"].(string); !ok {
		return nil, NewInvalidWebSocketParamError("start_record_screen", "root_id")
	}

	err := api.ctrl.ScreenShareRecordStart(ctx, conn.GetSession(), int64(toUserId), sessionId)
	if err != nil {
		return nil, err
	}

	return make(map[string]interface{}), nil
}

func (api *API) ssStopRecord(ctx context.Context, conn *app.WebConn, req *model.WebSocketRequest) (map[string]interface{}, model.AppError) {
	var toUserId float64
	var sessionId string
	var ok bool

	if toUserId, ok = req.Data["to_user_id"].(float64); !ok {
		return nil, NewInvalidWebSocketParamError("screenshot", "to_user_id")
	}

	if sessionId, ok = req.Data["root_id"].(string); !ok {
		return nil, NewInvalidWebSocketParamError("screenshot", "root_id")
	}

	err := api.ctrl.ScreenShareRecordStop(ctx, conn.GetSession(), int64(toUserId), sessionId)
	if err != nil {
		return nil, err
	}

	return make(map[string]interface{}), nil
}
