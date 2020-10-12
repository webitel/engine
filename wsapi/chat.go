package wsapi

import (
	"github.com/webitel/engine/app"
	"github.com/webitel/engine/model"
)

func (api *API) InitChat() {
	api.Router.Handle("subscribe_chat", api.ApiWebSocketHandler(api.subscribeSelfChat))
	api.Router.Handle("decline_chat", api.ApiWebSocketHandler(api.declineChat))
	api.Router.Handle("join_chat", api.ApiWebSocketHandler(api.joinChat))
	api.Router.Handle("close_chat", api.ApiWebSocketHandler(api.closeChat))
	api.Router.Handle("leave_chat", api.ApiWebSocketHandler(api.leaveChat))
	api.Router.Handle("send_text_chat", api.ApiWebSocketHandler(api.sendTextChat))
}

func (api *API) subscribeSelfChat(conn *app.WebConn, req *model.WebSocketRequest) (map[string]interface{}, *model.AppError) {
	h, e := api.App.GetHubById(req.Session.Domain(0))
	if e != nil {
		return nil, e
	}

	return nil, h.SubscribeSessionChat(conn)
}

func (api *API) declineChat(conn *app.WebConn, req *model.WebSocketRequest) (map[string]interface{}, *model.AppError) {
	var inviteId string
	var ok bool

	inviteId, ok = req.Data["invite_id"].(string)

	if !ok {
		return nil, NewInvalidWebSocketParamError(req.Action, "invite_id")
	}

	err := api.ctrl.DeclineChat(conn.GetSession(), inviteId)
	return nil, err
}

func (api *API) joinChat(conn *app.WebConn, req *model.WebSocketRequest) (map[string]interface{}, *model.AppError) {
	var inviteId string
	var ok bool

	inviteId, ok = req.Data["invite_id"].(string)

	if !ok {
		return nil, NewInvalidWebSocketParamError(req.Action, "invite_id")
	}

	channelId, err := api.ctrl.JoinChat(conn.GetSession(), inviteId)
	if err != nil {
		return nil, err
	}
	res := make(map[string]interface{})
	res["channel_id"] = channelId

	return res, nil
}

func (api *API) closeChat(conn *app.WebConn, req *model.WebSocketRequest) (map[string]interface{}, *model.AppError) {
	var channelId, conversationId, cause string
	var ok bool

	channelId, ok = req.Data["channel_id"].(string)
	if !ok || channelId == "" {
		return nil, NewInvalidWebSocketParamError(req.Action, "channel_id")
	}

	conversationId, ok = req.Data["conversation_id"].(string)
	if !ok || conversationId == "" {
		return nil, NewInvalidWebSocketParamError(req.Action, "conversation_id")
	}

	cause, ok = req.Data["cause"].(string)
	if !ok || cause == "" {
		return nil, NewInvalidWebSocketParamError(req.Action, "cause")
	}

	err := api.ctrl.CloseChat(conn.GetSession(), channelId, conversationId, cause)
	return nil, err
}

func (api *API) leaveChat(conn *app.WebConn, req *model.WebSocketRequest) (map[string]interface{}, *model.AppError) {
	var channelId, conversationId string
	var ok bool

	channelId, ok = req.Data["channel_id"].(string)
	if !ok || channelId == "" {
		return nil, NewInvalidWebSocketParamError(req.Action, "channel_id")
	}

	conversationId, ok = req.Data["conversation_id"].(string)
	if !ok || conversationId == "" {
		return nil, NewInvalidWebSocketParamError(req.Action, "conversation_id")
	}

	err := api.ctrl.LeaveChat(conn.GetSession(), channelId, conversationId)
	return nil, err
}

func (api *API) sendTextChat(conn *app.WebConn, req *model.WebSocketRequest) (map[string]interface{}, *model.AppError) {
	var channelId, conversationId, text string
	var ok bool

	channelId, ok = req.Data["channel_id"].(string)
	if !ok || channelId == "" {
		return nil, NewInvalidWebSocketParamError(req.Action, "channel_id")
	}

	conversationId, ok = req.Data["conversation_id"].(string)
	if !ok || conversationId == "" {
		return nil, NewInvalidWebSocketParamError(req.Action, "conversation_id")
	}

	text, ok = req.Data["text"].(string)
	if !ok || text == "" {
		return nil, NewInvalidWebSocketParamError(req.Action, "text")
	}

	err := api.ctrl.SendTextChat(conn.GetSession(), channelId, conversationId, text)
	return nil, err
}
