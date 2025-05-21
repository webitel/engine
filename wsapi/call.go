package wsapi

import (
	"context"
	"fmt"
	"github.com/webitel/engine/app"
	"github.com/webitel/engine/model"
	"github.com/webitel/wlog"
	"strings"
)

func (api *API) InitCall() {
	api.Router.Handle("subscribe_call", api.ApiWebSocketHandler(api.subscribeSelfCalls))
	api.Router.Handle("un_subscribe_call", api.ApiWebSocketHandler(api.unSubscribeSelfCalls))

	api.Router.Handle("call_invite", api.ApiAsyncWebSocketHandler(api.callInvite))
	api.Router.Handle("call_eavesdrop", api.ApiAsyncWebSocketHandler(api.callEavesdrop))
	api.Router.Handle("call_eavesdrop_state", api.ApiAsyncWebSocketHandler(api.callEavesdropState))
	api.Router.Handle("call_user", api.ApiAsyncWebSocketHandler(api.callToUser))
	api.Router.Handle("call_hangup", api.ApiWebSocketHandler(api.callHangup))
	api.Router.Handle("call_hold", api.ApiWebSocketHandler(api.callHold))
	api.Router.Handle("call_unhold", api.ApiWebSocketHandler(api.callUnHold))
	api.Router.Handle("call_dtmf", api.ApiWebSocketHandler(api.callDTMF))
	api.Router.Handle("call_mute", api.ApiWebSocketHandler(api.callMute))

	api.Router.Handle("call_bt_queue", api.ApiWebSocketHandler(api.callBTQueue))
	api.Router.Handle("call_to_queue", api.ApiWebSocketHandler(api.callToQueue))

	api.Router.Handle("call_blind_transfer", api.ApiWebSocketHandler(api.callBlindTransfer))
	api.Router.Handle("call_blind_transfer_ext", api.ApiWebSocketHandler(api.callBlindTransferExt))
	api.Router.Handle("call_bridge", api.ApiWebSocketHandler(api.callBridge))
	api.Router.Handle("call_recordings", api.ApiWebSocketHandler(api.callRecording))
	api.Router.Handle("call_set_params", api.ApiWebSocketHandler(api.callSetParams))
	api.Router.Handle("call_set_contact", api.ApiWebSocketHandler(api.callSetContact))

	api.Router.Handle("call_by_user", api.ApiAsyncWebSocketHandler(api.callByUser))

	api.Router.Handle("sip_proxy", api.ApiWebSocketHandler(api.sipProxy))
}

func (api *API) callByUser(ctx context.Context, conn *app.WebConn, req *model.WebSocketRequest) (map[string]interface{}, model.AppError) {
	calls, err := api.ctrl.UserActiveCall(ctx, conn.GetSession())

	if err != nil {
		return nil, err
	}

	res := make(map[string]interface{})
	res["items"] = calls
	return res, nil
}

func (api *API) sipProxy(ctx context.Context, conn *app.WebConn, req *model.WebSocketRequest) (map[string]interface{}, model.AppError) {
	var ok bool
	var data string
	if data, ok = req.Data["data"].(string); !ok {
		return nil, NewInvalidWebSocketParamError(req.Action, "data")
	}
	//conn.Sip.Send([]byte(data))
	if data != "" {
	}
	return nil, nil
}

func (api *API) subscribeSelfCalls(ctx context.Context, conn *app.WebConn, req *model.WebSocketRequest) (map[string]interface{}, model.AppError) {
	h, e := api.App.GetHubById(req.Session.Domain(0))
	if e != nil {
		return nil, e
	}

	e = h.SubscribeSessionCalls(conn)
	if e != nil {
		return nil, e
	}

	return api.callByUser(ctx, conn, req)
}

func (api *API) unSubscribeSelfCalls(ctx context.Context, conn *app.WebConn, req *model.WebSocketRequest) (map[string]interface{}, model.AppError) {
	h, e := api.App.GetHubById(req.Session.Domain(0))
	if e != nil {
		return nil, e
	}

	return nil, h.UnSubscribeCalls(conn)
}

func (api *API) callEavesdrop(ctx context.Context, conn *app.WebConn, req *model.WebSocketRequest) (map[string]interface{}, model.AppError) {
	var ok bool
	reqEa := &model.EavesdropCall{}

	if reqEa.Id, ok = req.Data["id"].(string); !ok {
		return nil, NewInvalidWebSocketParamError(req.Action, "id")
	}

	reqEa.Dtmf, _ = req.Data["control"].(bool)
	reqEa.ALeg, _ = req.Data["listenA"].(bool)
	reqEa.BLeg, _ = req.Data["listenB"].(bool)
	reqEa.WhisperALeg, _ = req.Data["whisperA"].(bool)
	reqEa.WhisperBLeg, _ = req.Data["whisperB"].(bool)

	reqEa.Notify, _ = req.Data["notify"].(bool)

	reqEa.State, _ = req.Data["state"].(string)

	vars := make(map[string]string)
	vars[model.CALL_VARIABLE_SOCK_ID] = conn.Id()

	callId, err := api.ctrl.EavesdropCall(ctx, conn.GetSession(), req.Session.DomainId, reqEa, vars)
	if err != nil {
		return nil, err
	}
	res := make(map[string]interface{})
	res["id"] = callId

	return res, nil
}

func (api *API) callEavesdropState(ctx context.Context, conn *app.WebConn, req *model.WebSocketRequest) (map[string]interface{}, model.AppError) {
	var ok bool
	reqEa := &model.EavesdropCall{}

	if reqEa.Id, ok = req.Data["id"].(string); !ok {
		return nil, NewInvalidWebSocketParamError(req.Action, "id")
	}
	if reqEa.State, ok = req.Data["state"].(string); !ok {
		return nil, NewInvalidWebSocketParamError(req.Action, "state")
	}

	err := api.ctrl.EavesdropStateCall(ctx, conn.GetSession(), req.Session.DomainId, reqEa)
	if err != nil {
		return nil, err
	}

	res := make(map[string]interface{})
	return res, nil
}

func (api *API) callHangup(ctx context.Context, conn *app.WebConn, req *model.WebSocketRequest) (map[string]interface{}, model.AppError) {
	var ok bool
	var id, nodeId string

	if id, ok = req.Data["id"].(string); !ok {
		return nil, NewInvalidWebSocketParamError(req.Action, "id")
	}

	if nodeId, ok = req.Data["app_id"].(string); !ok {
		return nil, NewInvalidWebSocketParamError(req.Action, "app_id")
	}

	var cause = req.GetFieldString("cause")

	cr := model.HangupCall{
		UserCallRequest: model.UserCallRequest{
			Id:    id,
			AppId: &nodeId,
		},
		Cause: nil,
	}

	if cause != "" {
		cr.Cause = &cause
	}

	err := api.App.HangupCall(ctx, conn.GetSession().DomainId, &cr)

	if err == nil {
		// DEV-4198
		conn.Log().Debug("success hangup",
			wlog.String("call_id", id),
			wlog.String("cause", cause),
		)
	}

	return nil, err
}

func (api *API) callBTQueue(ctx context.Context, conn *app.WebConn, req *model.WebSocketRequest) (map[string]interface{}, model.AppError) {
	var ok bool
	var id string
	var queueId float64

	if id, ok = req.Data["id"].(string); !ok {
		return nil, NewInvalidWebSocketParamError(req.Action, "id")
	}

	if queueId, ok = req.Data["queue_id"].(float64); !ok {
		return nil, NewInvalidWebSocketParamError(req.Action, "queue_id")
	}

	err := api.ctrl.BlindTransferCallToQueue(ctx, conn.GetSession(), conn.DomainId, &model.BlindTransferCallToQueue{
		UserCallRequest: model.UserCallRequest{
			Id: id,
		},
		Variables: variablesFromMap(req.Data, "variables"),
		QueueId:   int(queueId),
	})

	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (api *API) callToQueue(ctx context.Context, conn *app.WebConn, req *model.WebSocketRequest) (map[string]interface{}, model.AppError) {
	var cp model.CallParameters
	var queueId *int
	var agentId *int

	sess := conn.GetSession()

	parentId, _ := req.Data["parent_id"].(string)

	if tmp, ok := req.Data["queue_id"].(float64); ok {
		queueId = model.NewInt(int(tmp))
	} else if tmp, ok = req.Data["agent_id"].(float64); ok {
		agentId = model.NewInt(int(tmp))
	}

	_, err := api.ctrl.CallToQueue(ctx, sess, sess.UserId, parentId, cp, queueId, agentId)

	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (api *API) callBlindTransfer(ctx context.Context, conn *app.WebConn, req *model.WebSocketRequest) (map[string]interface{}, model.AppError) {
	var ok bool
	var id, destination string

	if id, ok = req.Data["id"].(string); !ok {
		return nil, NewInvalidWebSocketParamError(req.Action, "id")
	}
	if destination, ok = req.Data["destination"].(string); !ok || len(destination) < 1 {
		return nil, NewInvalidWebSocketParamError(req.Action, "destination")
	}

	err := api.ctrl.BlindTransferCall(ctx, conn.GetSession(), conn.DomainId, &model.BlindTransferCall{
		UserCallRequest: model.UserCallRequest{
			Id: id,
		},
		Destination: destination,
	})

	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (api *API) callBlindTransferExt(ctx context.Context, conn *app.WebConn, req *model.WebSocketRequest) (map[string]interface{}, model.AppError) {
	var ok bool
	var id, destination string

	if id, ok = req.Data["id"].(string); !ok {
		return nil, NewInvalidWebSocketParamError(req.Action, "id")
	}
	if destination, ok = req.Data["destination"].(string); !ok || len(destination) < 1 {
		return nil, NewInvalidWebSocketParamError(req.Action, "destination")
	}

	err := api.ctrl.BlindTransferCallExt(ctx, conn.GetSession(), conn.DomainId, &model.BlindTransferCall{
		UserCallRequest: model.UserCallRequest{
			Id: id,
		},
		Destination: destination,
		Variables:   variablesFromMap(req.Data, "variables"),
	})

	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (api *API) callHold(ctx context.Context, conn *app.WebConn, req *model.WebSocketRequest) (map[string]interface{}, model.AppError) {
	var ok bool
	var id, nodeId string

	if id, ok = req.Data["id"].(string); !ok {
		return nil, NewInvalidWebSocketParamError(req.Action, "id")
	}
	if nodeId, ok = req.Data["app_id"].(string); !ok {
		return nil, NewInvalidWebSocketParamError(req.Action, "app_id")
	}

	if cli, err := api.App.CallManager().CallClientById(nodeId); err != nil {
		return nil, err
	} else {
		err = cli.Hold(id)
		if err != nil {
			return nil, err
		}
	}

	return nil, nil
}

func (api *API) callDTMF(ctx context.Context, conn *app.WebConn, req *model.WebSocketRequest) (map[string]interface{}, model.AppError) {
	var ok bool
	var id, nodeId string
	var key string

	if id, ok = req.Data["id"].(string); !ok {
		return nil, NewInvalidWebSocketParamError(req.Action, "id")
	}
	if nodeId, ok = req.Data["app_id"].(string); !ok {
		return nil, NewInvalidWebSocketParamError(req.Action, "app_id")
	}
	if key, ok = req.Data["dtmf"].(string); !ok || len(key) < 1 {
		return nil, NewInvalidWebSocketParamError(req.Action, "dtmf")
	}

	if cli, err := api.App.CallManager().CallClientById(nodeId); err != nil {
		return nil, err
	} else {
		err = cli.DTMF(id, []rune(key)[0])
		if err != nil {
			return nil, err
		}
	}
	return nil, nil
}

func (api *API) callUnHold(ctx context.Context, conn *app.WebConn, req *model.WebSocketRequest) (map[string]interface{}, model.AppError) {
	var ok bool
	var id, nodeId string

	if id, ok = req.Data["id"].(string); !ok {
		return nil, NewInvalidWebSocketParamError(req.Action, "id")
	}
	if nodeId, ok = req.Data["app_id"].(string); !ok {
		return nil, NewInvalidWebSocketParamError(req.Action, "app_id")
	}

	if cli, err := api.App.CallManager().CallClientById(nodeId); err != nil {
		return nil, err
	} else {
		err = cli.UnHold(id)
		if err != nil {
			return nil, err
		}
	}

	return nil, nil
}

func (api *API) callInvite(ctx context.Context, conn *app.WebConn, req *model.WebSocketRequest) (map[string]interface{}, model.AppError) {
	var callReq = &model.OutboundCallRequest{}
	var ok bool
	var props map[string]interface{}
	if callReq.Destination, ok = req.Data["destination"].(string); !ok {
		return nil, NewInvalidWebSocketParamError(req.Action, "destination")
	}

	if props, ok = req.Data["params"].(map[string]interface{}); ok {
		callReq.Params = callParams(props)
	}

	vars := make(map[string]string)
	vars[model.CALL_VARIABLE_SOCK_ID] = conn.Id()

	if id, err := api.ctrl.CreateCall(ctx, conn.GetSession(), callReq, vars); err != nil {
		return nil, err
	} else {
		data := make(map[string]interface{})
		data["id"] = id
		return data, nil
	}
}

func (api *API) callToUser(ctx context.Context, conn *app.WebConn, req *model.WebSocketRequest) (map[string]interface{}, model.AppError) {
	var ok, useVideo, useScreen bool
	var callId, callToId, parentCallId, sendToCallId string
	var toUserId float64
	var variables map[string]interface{}

	if toUserId, ok = req.Data["toUserId"].(float64); !ok {
		return nil, NewInvalidWebSocketParamError(req.Action, "toUserId")
	}
	sendToCallId, _ = req.Data["sendToCallId"].(string)

	callId = model.NewUuid()
	callToId = model.NewUuid()

	info, err := api.App.GetUserCallInfo(ctx, conn.UserId, conn.DomainId)
	if err != nil {
		return nil, err
	}

	infoTo, err := api.App.GetUserCallInfo(ctx, int64(toUserId), conn.DomainId)
	if err != nil {
		return nil, err
	}

	invite := &model.CallRequest{
		Endpoints: info.GetCallEndpoints(),
		Variables: map[string]string{
			model.CALL_VARIABLE_ID:                callId,
			model.CALL_VARIABLE_DIRECTION:         model.CALL_DIRECTION_INTERNAL,
			model.CALL_VARIABLE_DISPLAY_DIRECTION: model.CALL_DIRECTION_OUTBOUND,
			model.CALL_VARIABLE_USER_ID:           fmt.Sprintf("%v", conn.UserId),
			model.CALL_VARIABLE_DOMAIN_ID:         fmt.Sprintf("%v", conn.DomainId),
			model.CALL_VARIABLE_SOCK_ID:           conn.Id(),

			"sip_h_X-Webitel-Destination": infoTo.Extension,

			"origination_uuid": callId,
			//"media_webrtc":     "true",
			//"absolute_codec_string": "VP8",

			"hangup_after_bridge":        "true",
			"hold_music":                 "silence",
			"effective_caller_id_number": info.Extension,
			"effective_caller_id_name":   info.Name,
			"effective_callee_id_name":   infoTo.Name,
			"effective_callee_id_number": infoTo.Extension,

			"origination_caller_id_name":   infoTo.Name,
			"origination_caller_id_number": infoTo.Extension,
			"origination_callee_id_name":   info.Name,
			"origination_callee_id_number": info.Extension,
		},
		Timeout:      0,
		CallerName:   infoTo.Name,
		CallerNumber: infoTo.Extension,
		Applications: []*model.CallRequestApplication{
			{
				AppName: "bridge",
				Args: fmt.Sprintf("{sip_route_uri=%s,request_parent_call_id=%s,origination_uuid=%s,sip_h_X-Webitel-Uuid=%s, sip_h_X-Webitel-User-Id=%d}%s", api.App.CallManager().SipRouteUri(),
					sendToCallId, callToId, callToId, int64(toUserId), strings.Join(infoTo.GetCallEndpoints(), ",")),
			},
		},
	}

	if variables, ok = req.Data["variables"].(map[string]interface{}); ok {
		for k, v := range variables {
			switch v.(type) {
			case string:
				invite.AddUserVariable(k, v.(string))
			case interface{}:
				invite.AddUserVariable(k, fmt.Sprintf("%v", v))
			}
		}
	}

	if useVideo, ok = req.Data["useVideo"].(bool); ok && useVideo {
		invite.AddVariable("video_request", "true")
	}

	if useScreen, ok = req.Data["useScreen"].(bool); ok && useScreen {
		invite.AddVariable("screen_request", "true")
	}

	if parentCallId, ok = req.Data["parentCallId"].(string); ok && parentCallId != "" {
		invite.AddVariable("request_parent_call_id", parentCallId)
	}

	_, err = api.App.CallManager().MakeOutboundCall(invite)

	if err != nil {
		return nil, err
	}

	data := map[string]interface{}{}
	data["call_id"] = callId
	return data, nil
}

func (api *API) callMute(ctx context.Context, conn *app.WebConn, req *model.WebSocketRequest) (map[string]interface{}, model.AppError) {
	var ok, mute bool
	var id, nodeId string

	if id, ok = req.Data["id"].(string); !ok {
		return nil, NewInvalidWebSocketParamError(req.Action, "id")
	}
	if nodeId, ok = req.Data["app_id"].(string); !ok {
		return nil, NewInvalidWebSocketParamError(req.Action, "app_id")
	}
	if mute, ok = req.Data["mute"].(bool); !ok {
		return nil, NewInvalidWebSocketParamError(req.Action, "mute")
	}

	if cli, err := api.App.CallManager().CallClientById(nodeId); err != nil {
		return nil, err
	} else {
		err = cli.Mute(id, mute)
		if err != nil {
			return nil, err
		}
	}

	return nil, nil
}

func (api *API) callBridge(ctx context.Context, conn *app.WebConn, req *model.WebSocketRequest) (map[string]interface{}, model.AppError) {
	var ok bool
	var fromId, toId string

	if fromId, ok = req.Data["from_id"].(string); !ok {
		return nil, NewInvalidWebSocketParamError(req.Action, "from_id")
	}

	if toId, ok = req.Data["to_id"].(string); !ok {
		return nil, NewInvalidWebSocketParamError(req.Action, "to_id")
	}
	res := make(map[string]interface{})
	err := api.App.BridgeCall(ctx, conn.DomainId, fromId, toId, variablesFromMap(req.Data, "variables"))
	//FIXME set result
	return res, err
}

func (api *API) callRecording(ctx context.Context, conn *app.WebConn, req *model.WebSocketRequest) (map[string]interface{}, model.AppError) {
	var id string
	var ok bool

	if id, ok = req.Data["id"].(string); !ok {
		return nil, NewInvalidWebSocketParamError(req.Action, "id")
	}

	fileId, err := api.App.GetLastCallFile(ctx, conn.DomainId, id)
	if err != nil {
		return nil, err
	}

	res := make(map[string]interface{})
	res["file_id"] = fileId
	return res, nil
}

func (api *API) callSetParams(ctx context.Context, conn *app.WebConn, req *model.WebSocketRequest) (map[string]interface{}, model.AppError) {
	var id string
	var ok bool
	var params model.CallParameters
	params.Variables = make(map[string]string)

	if id, ok = req.Data["id"].(string); !ok {
		return nil, NewInvalidWebSocketParamError(req.Action, "id")
	}

	if variables, ok := req.Data["variables"].(map[string]interface{}); ok {
		for k, v := range variables {
			switch v.(type) {
			case string:
				params.Variables[k] = v.(string)
			case interface{}:
				params.Variables[k] = fmt.Sprintf("%v", v)
			}
		}
	}

	err := api.App.SetCallParams(ctx, conn.DomainId, id, params)
	if err != nil {
		return nil, err
	}

	res := make(map[string]interface{})
	return res, nil
}

func (api *API) callSetContact(ctx context.Context, conn *app.WebConn, req *model.WebSocketRequest) (map[string]interface{}, model.AppError) {
	var ok bool
	var id string
	var contactId float64

	if id, ok = req.Data["id"].(string); !ok {
		return nil, NewInvalidWebSocketParamError(req.Action, "id")
	}

	if contactId, ok = req.Data["contact_id"].(float64); !ok {
		return nil, NewInvalidWebSocketParamError(req.Action, "contact_id")
	}

	res := make(map[string]interface{})
	err := api.ctrl.SetContactCall(ctx, conn.GetSession(), id, int64(contactId))
	return res, err
}

func (api *API) callSendVideo(ctx context.Context, conn *app.WebConn, req *model.WebSocketRequest) (map[string]interface{}, model.AppError) {
	var ok bool
	var id, nodeId, id2, nodeId2 string

	if id, ok = req.Data["id"].(string); !ok {
		return nil, NewInvalidWebSocketParamError(req.Action, "id")
	}
	if nodeId, ok = req.Data["app_id"].(string); !ok {
		return nil, NewInvalidWebSocketParamError(req.Action, "app_id")
	}
	if id2, ok = req.Data["parent_id"].(string); !ok {
		return nil, NewInvalidWebSocketParamError(req.Action, "parent_id")
	}
	if nodeId2, ok = req.Data["parent_app_id"].(string); !ok {
		return nil, NewInvalidWebSocketParamError(req.Action, "parent_app_id")
	}

	api.App.CallManager().Bridge(id, nodeId, id2, nodeId2)

	return nil, nil
}

func variablesFromMap(m map[string]interface{}, name string) map[string]string {
	vi, ok := m[name].(map[string]interface{})
	if !ok {
		return nil
	}

	vars := make(map[string]string, len(vi))
	for k, v := range vi {
		vars[k] = fmt.Sprintf("%v", v)
	}

	return vars

}

func callParams(props map[string]any) model.CallParameters {
	var params model.CallParameters

	params.Timeout, _ = props["timeout"].(int)
	params.Video, _ = props["video"].(bool)
	params.Screen, _ = props["screen"].(bool)
	params.Record, _ = props["record"].(bool)
	params.DisableAutoAnswer, _ = props["disableAutoAnswer"].(bool)
	params.Display, _ = props["display"].(string)
	params.HideNumber, _ = props["hide_number"].(bool)

	if variables, ok := props["variables"].(map[string]interface{}); ok {
		params.Variables = make(map[string]string)
		for k, v := range variables {
			switch v.(type) {
			case string:
				params.Variables[k] = v.(string)
			case interface{}:
				params.Variables[k] = fmt.Sprintf("%v", v)
			}
		}
	}

	return params
}
