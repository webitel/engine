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
	api.Router.Handle("send_file_chat", api.ApiWebSocketHandler(api.sendFileChat))
	api.Router.Handle("add_to_chat", api.ApiWebSocketHandler(api.addToChat))
	api.Router.Handle("start_chat", api.ApiWebSocketHandler(api.startChat))
	api.Router.Handle("update_channel_chat", api.ApiWebSocketHandler(api.updateChannelChat))
	api.Router.Handle("list_active_chat", api.ApiWebSocketHandler(api.listActiveChat))
	api.Router.Handle("blind_transfer_chat", api.ApiWebSocketHandler(api.blindTransfer))
	api.Router.Handle("transfer_user_chat", api.ApiWebSocketHandler(api.blindTransferToUser))
}

func (api *API) subscribeSelfChat(conn *app.WebConn, req *model.WebSocketRequest) (map[string]interface{}, *model.AppError) {
	h, e := api.App.GetHubById(req.Session.Domain(0))
	if e != nil {
		return nil, e
	}

	e = h.SubscribeSessionChat(conn)
	if e != nil {
		return nil, e
	}

	list, err := api.ctrl.ListActiveChat(conn.GetSession(), 0, model.PER_PAGE_DEFAULT)

	if err != nil {
		return nil, err
	}

	return listChatResponse(list), nil
}

func (api *API) declineChat(conn *app.WebConn, req *model.WebSocketRequest) (map[string]interface{}, *model.AppError) {
	var inviteId string
	var cause string
	var ok bool

	inviteId, ok = req.Data["invite_id"].(string)

	if !ok {
		return nil, NewInvalidWebSocketParamError(req.Action, "invite_id")
	}

	cause, _ = req.Data["cause"].(string)

	err := api.ctrl.DeclineChat(conn.GetSession(), inviteId, cause)
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

func (api *API) sendFileChat(conn *app.WebConn, req *model.WebSocketRequest) (map[string]interface{}, *model.AppError) {
	var channelId, conversationId, url, mimeType, name string
	var id, size float64
	var ok bool

	id, ok = req.Data["id"].(float64)
	if !ok || id == 0 {
		return nil, NewInvalidWebSocketParamError(req.Action, "id")
	}
	name, _ = req.Data["name"].(string)

	size, ok = req.Data["size"].(float64)
	if !ok {
		return nil, NewInvalidWebSocketParamError(req.Action, "size")
	}

	channelId, ok = req.Data["channel_id"].(string)
	if !ok || channelId == "" {
		return nil, NewInvalidWebSocketParamError(req.Action, "channel_id")
	}

	conversationId, ok = req.Data["conversation_id"].(string)
	if !ok || conversationId == "" {
		return nil, NewInvalidWebSocketParamError(req.Action, "conversation_id")
	}

	url, ok = req.Data["url"].(string)
	if !ok || url == "" {
		return nil, NewInvalidWebSocketParamError(req.Action, "url")
	}
	mimeType, _ = req.Data["mime"].(string)

	err := api.ctrl.SendFileChat(conn.GetSession(), channelId, conversationId, &model.ChatFile{
		Id:   int64(id),
		Name: name,
		Url:  url,
		Mime: mimeType,
		Size: int64(size),
	})
	return nil, err
}

func (api *API) addToChat(conn *app.WebConn, req *model.WebSocketRequest) (map[string]interface{}, *model.AppError) {
	var channelId, conversationId, title string
	var userId float64
	var ok bool

	channelId, ok = req.Data["channel_id"].(string)
	if !ok || channelId == "" {
		return nil, NewInvalidWebSocketParamError(req.Action, "channel_id")
	}

	title, _ = req.Data["title"].(string)

	conversationId, ok = req.Data["conversation_id"].(string)
	if !ok || conversationId == "" {
		return nil, NewInvalidWebSocketParamError(req.Action, "conversation_id")
	}

	userId, ok = req.Data["user_id"].(float64)
	if !ok || userId == 0 {
		return nil, NewInvalidWebSocketParamError(req.Action, "channel_id")
	}

	err := api.ctrl.AddToChat(conn.GetSession(), int64(userId), channelId, conversationId, title)
	return nil, err
}

func (api *API) startChat(conn *app.WebConn, req *model.WebSocketRequest) (map[string]interface{}, *model.AppError) {
	var userId float64
	var ok bool

	userId, ok = req.Data["user_id"].(float64)
	if !ok || userId == 0 {
		return nil, NewInvalidWebSocketParamError(req.Action, "user_id")
	}

	err := api.ctrl.StartChat(conn.GetSession(), int64(userId))
	return nil, err
}

func (api *API) updateChannelChat(conn *app.WebConn, req *model.WebSocketRequest) (map[string]interface{}, *model.AppError) {
	var channelId string
	var ok bool
	var readUntil float64

	channelId, ok = req.Data["channel_id"].(string)
	if !ok || channelId == "" {
		return nil, NewInvalidWebSocketParamError(req.Action, "channel_id")
	}

	readUntil, _ = req.Data["read_until"].(float64)

	err := api.ctrl.UpdateChannelChat(conn.GetSession(), channelId, int64(readUntil))
	return nil, err
}

func (api *API) listActiveChat(conn *app.WebConn, req *model.WebSocketRequest) (map[string]interface{}, *model.AppError) {
	var page, size float64
	var ok bool

	page, _ = req.Data["page"].(float64)

	if size, ok = req.Data["size"].(float64); !ok {
		size = model.PER_PAGE_DEFAULT
	}

	list, err := api.ctrl.ListActiveChat(conn.GetSession(), int(page), int(size))

	if err != nil {
		return nil, err
	}

	return listChatResponse(list), nil
}

func (api *API) blindTransfer(conn *app.WebConn, req *model.WebSocketRequest) (map[string]interface{}, *model.AppError) {
	var conversationId, channelId string
	var planId float64

	conversationId, _ = req.Data["conversation_id"].(string)
	if conversationId == "" {
		return nil, NewInvalidWebSocketParamError(req.Action, "conversation_id")
	}
	channelId, _ = req.Data["channel_id"].(string)

	planId, _ = req.Data["plan_id"].(float64)
	if planId == 0 {
		return nil, NewInvalidWebSocketParamError(req.Action, "plan_id")
	}

	return nil, api.ctrl.BlindTransferChat(conn.GetSession(), conversationId, channelId, int32(planId), nil)
}

func (api *API) blindTransferToUser(conn *app.WebConn, req *model.WebSocketRequest) (map[string]interface{}, *model.AppError) {
	var conversationId, channelId string
	var userId float64

	conversationId, _ = req.Data["conversation_id"].(string)
	if conversationId == "" {
		return nil, NewInvalidWebSocketParamError(req.Action, "conversation_id")
	}
	channelId, _ = req.Data["channel_id"].(string)

	userId, _ = req.Data["user_id"].(float64)
	if userId == 0 {
		return nil, NewInvalidWebSocketParamError(req.Action, "user_id")
	}

	return nil, api.ctrl.BlindTransferChatToUser(conn.GetSession(), conversationId, channelId, int64(userId), nil)
}

func listChatResponse(list []*model.Conversation) map[string]interface{} {
	res := make(map[string]interface{})
	res["items"] = list

	return res
}
