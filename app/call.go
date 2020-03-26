package app

import (
	"fmt"
	"github.com/webitel/engine/call_manager"
	"github.com/webitel/engine/model"
	"net/http"
)

func (app *App) CreateOutboundCall(domainId int64, req *model.OutboundCallRequest, variables map[string]string) (string, *model.AppError) {
	var callCli call_manager.CallClient
	var err *model.AppError
	var id string

	var from *model.UserCallInfo

	if req.From.AppId != nil {
		callCli, err = app.CallManager().CallClientById(*req.From.AppId)
	} else {
		callCli, err = app.CallManager().CallClient()
	}

	if err != nil {
		return "", err
	}

	if callCli == nil {
		return "", model.NewAppError("CreateOutboundCall", "app.call.create.not_found", nil, "", http.StatusNotFound)
	}

	if req.From.UserId != nil {
		if from, err = app.GetUserCallInfo(*req.From.UserId, domainId); err != nil {
			return "", err
		}
	} else {
		return "", model.NewAppError("CreateOutboundCall", "app.call.create.valid.from", nil, "", http.StatusBadRequest)
	}

	invite := inviteFromUser(domainId, req, from)
	for k, v := range req.Params.Variables {
		invite.AddUserVariable(k, v)
	}
	for k, v := range variables {
		invite.AddVariable(k, v)
	}

	if req.Params.Video {
		invite.AddVariable(model.CALL_VARIABLE_USE_VIDEO, "true")
	}

	if req.Params.Screen {
		invite.AddVariable(model.CALL_VARIABLE_USE_SCREEN, "true")
	}

	id, err = callCli.MakeOutboundCall(invite)
	if err != nil {
		return "", err
	}

	return id, nil

}

func (app *App) GetCall(domainId int64, callId string) (*model.Call, *model.AppError) {
	return app.Store.Call().Get(domainId, callId)
}

func (app *App) EavesdropCall(domainId, userId int64, req *model.EavesdropCall, variables map[string]string) (string, *model.AppError) {
	var call *model.Call
	var cli call_manager.CallClient

	usr, err := app.GetUserCallInfo(userId, domainId)
	if err != nil {
		return "", err
	}

	call, err = app.GetCall(domainId, req.Id)
	if err != nil {
		return "", err
	}

	cli, err = app.getCallCli(domainId, req.Id, req.AppId)
	if err != nil {
		return "", err
	}

	invite := &model.CallRequest{
		Endpoints:   usr.GetCallEndpoints(),
		Destination: call.Destination,
		Variables: model.UnionStringMaps(
			usr.GetVariables(),
			variables,
			map[string]string{
				model.CALL_VARIABLE_DIRECTION:         model.CALL_DIRECTION_INTERNAL,
				model.CALL_VARIABLE_DISPLAY_DIRECTION: model.CALL_DIRECTION_OUTBOUND,
				model.CALL_VARIABLE_USER_ID:           fmt.Sprintf("%v", usr.Id),
				model.CALL_VARIABLE_DOMAIN_ID:         fmt.Sprintf("%v", domainId),
				"hangup_after_bridge":                 "true",

				"variable_sip_h_X-Webitel-User-Id": fmt.Sprintf("%d", usr.Id),
				"wbt_destination":                  call.Destination,
				"wbt_from_id":                      fmt.Sprintf("%v", usr.Id),
				"wbt_from_number":                  usr.Endpoint,
				"wbt_from_name":                    usr.Name,
				"wbt_from_type":                    model.EndpointTypeUser,
				"wbt_parent_id":                    call.Id,

				"wbt_to_id":     fmt.Sprintf("%v", call.From.Id),
				"wbt_to_name":   call.From.Name,
				"wbt_to_number": call.From.Number,
				"wbt_to_type":   call.From.Type,

				"effective_caller_id_number": call.From.Number,
				"effective_caller_id_name":   call.From.Name,

				"effective_callee_id_name":   usr.Name,
				"effective_callee_id_number": usr.Extension,

				"origination_caller_id_name":   call.From.Name,
				"origination_caller_id_number": call.From.Number,
				"origination_callee_id_name":   usr.Name,
				"origination_callee_id_number": usr.Extension,
			},
		),
		CallerName:   usr.Name,
		CallerNumber: usr.Extension,
		Applications: []*model.CallRequestApplication{
			{
				AppName: "eavesdrop",
				Args:    call.Id,
			},
		},
	}

	if req.Dtmf {
		invite.AddVariable("eavesdrop_enable_dtmf", "true")
	}

	if req.ALeg {
		invite.AddVariable("eavesdrop_bridge_aleg", "true")
	}

	if req.BLeg {
		invite.AddVariable("eavesdrop_bridge_bleg", "true")
	}

	if req.WhisperALeg {
		invite.AddVariable("eavesdrop_whisper_aleg", "true")
	}

	if req.WhisperBLeg {
		invite.AddVariable("eavesdrop_whisper_bleg", "true")
	}

	var id string
	id, err = cli.MakeOutboundCall(invite)
	if err != nil {
		return "", err
	}

	return id, nil
}

func inviteFromUser(domainId int64, req *model.OutboundCallRequest, usr *model.UserCallInfo) *model.CallRequest {
	return &model.CallRequest{
		Endpoints:   usr.GetCallEndpoints(),
		Timeout:     uint16(req.Params.Timeout),
		Destination: req.Destination,
		Variables: model.UnionStringMaps(
			usr.GetVariables(),
			map[string]string{
				model.CALL_VARIABLE_DIRECTION:         model.CALL_DIRECTION_INTERNAL,
				model.CALL_VARIABLE_DISPLAY_DIRECTION: model.CALL_DIRECTION_OUTBOUND,
				model.CALL_VARIABLE_USER_ID:           fmt.Sprintf("%v", usr.Id),
				model.CALL_VARIABLE_DOMAIN_ID:         fmt.Sprintf("%v", domainId),
				"hangup_after_bridge":                 "true",

				"sip_h_X-Webitel-Origin": "request",
				"wbt_created_by":         fmt.Sprintf("%v", usr.Id),
				"wbt_destination":        req.Destination,
				"wbt_from_id":            fmt.Sprintf("%v", usr.Id),
				"wbt_from_number":        usr.Endpoint,
				"wbt_from_name":          usr.Name,
				"wbt_from_type":          model.EndpointTypeUser,

				//"wbt_to_id":   fmt.Sprintf("%v", toEndpoint.Id),
				//"wbt_to_name": toEndpoint.Name,
				//"wbt_to_type": toEndpoint.Type,

				"effective_caller_id_number": usr.Extension,
				"effective_caller_id_name":   usr.Name,
				"effective_callee_id_name":   req.Destination,
				"effective_callee_id_number": req.Destination,

				"origination_caller_id_name":   req.Destination,
				"origination_caller_id_number": req.Destination,
				"origination_callee_id_name":   usr.Name,
				"origination_callee_id_number": usr.Extension,
			},
		),
		CallerName:   usr.Name,
		CallerNumber: usr.Extension,
	}
}

func (app *App) GetActiveCallPage(domainId int64, search *model.SearchCall) ([]*model.Call, bool, *model.AppError) {
	list, err := app.Store.Call().GetActive(domainId, search)
	if err != nil {
		return nil, false, err
	}
	search.RemoveLastElemIfNeed(&list)
	return list, search.EndOfList(), nil
}

func (app *App) GetHistoryCallPage(domainId int64, search *model.SearchHistoryCall) ([]*model.HistoryCall, bool, *model.AppError) {
	list, err := app.Store.Call().GetHistory(domainId, search)
	if err != nil {
		return nil, false, err
	}
	search.RemoveLastElemIfNeed(&list)
	return list, search.EndOfList(), nil
}

func (app *App) getCallCli(domainId int64, id string, appId *string) (cli call_manager.CallClient, err *model.AppError) {

	if appId != nil {
		cli, err = app.CallManager().CallClientById(*appId)
	} else {
		var info *model.CallInstance
		info, err = app.Store.Call().GetInstance(domainId, id)
		if err != nil {
			return nil, err
		}
		cli, err = app.CallManager().CallClientById(*info.AppId)
	}
	return
}

func (app *App) HangupCall(domainId int64, req *model.HangupCall) *model.AppError {
	var cli call_manager.CallClient
	var err *model.AppError
	var cause = ""

	cli, err = app.getCallCli(domainId, req.Id, req.AppId)
	if err != nil {
		return err
	}

	if req.Cause != nil {
		cause = *req.Cause
	}

	return cli.HangupCall(req.Id, cause)
}

func (app *App) HoldCall(domainId int64, req *model.UserCallRequest) *model.AppError {
	var cli call_manager.CallClient
	var err *model.AppError

	cli, err = app.getCallCli(domainId, req.Id, req.AppId)
	if err != nil {
		return err
	}

	return cli.Hold(req.Id)
}

func (app *App) UnHoldCall(domainId int64, req *model.UserCallRequest) *model.AppError {
	var cli call_manager.CallClient
	var err *model.AppError

	cli, err = app.getCallCli(domainId, req.Id, req.AppId)
	if err != nil {
		return err
	}

	return cli.UnHold(req.Id)
}

func (app *App) DtmfCall(domainId int64, req *model.DtmfCall) *model.AppError {
	var cli call_manager.CallClient
	var err *model.AppError

	cli, err = app.getCallCli(domainId, req.Id, req.AppId)
	if err != nil {
		return err
	}

	return cli.DTMF(req.Id, req.Digit)
}

func (app *App) BlindTransferCall(domainId int64, req *model.BlindTransferCall) *model.AppError {
	var cli call_manager.CallClient
	var err *model.AppError

	cli, err = app.getCallCli(domainId, req.Id, req.AppId)
	if err != nil {
		return err
	}

	return cli.BlindTransfer(req.Id, req.Destination)
}

func (app *App) BridgeCall(domainId int64, fromId, toId string) *model.AppError {
	var cli call_manager.CallClient
	info, err := app.Store.Call().BridgeInfo(domainId, fromId, toId)
	if err != nil {
		return err
	}

	cli, err = app.getCallCli(domainId, info.FromId, &info.AppId)
	if err != nil {
		return err
	}

	_, err = cli.BridgeCall(info.FromId, info.ToId, "")
	return err
}

/*

func (app *App) createOutboundCallToUser(domainId int64, req *model.OutboundCallRequest, from, to *model.UserCallInfo) (*model.CallRequest, *model.AppError) {
	invite := &model.CallRequest{
		Endpoints: from.GetCallEndpoints(),
		Variables: map[string]string{
			model.CALL_VARIABLE_DIRECTION:         model.CALL_DIRECTION_INTERNAL,
			model.CALL_VARIABLE_DISPLAY_DIRECTION: model.CALL_DIRECTION_OUTBOUND,
			model.CALL_VARIABLE_USER_ID:           fmt.Sprintf("%v", req.FromId),
			model.CALL_VARIABLE_DOMAIN_ID:         fmt.Sprintf("%v", domainId),

			"sip_h_X-Webitel-Destination": to.Extension,

			"hangup_after_bridge":        "true",
			"effective_caller_id_number": from.Extension,
			"effective_caller_id_name":   from.Name,
			"effective_callee_id_name":   to.Name,
			"effective_callee_id_number": to.Extension,

			"origination_caller_id_name":   to.Name,
			"origination_caller_id_number": to.Extension,
			"origination_callee_id_name":   from.Name,
			"origination_callee_id_number": from.Extension,
		},
		Timeout:      req.Timeout,
		CallerName:   to.Name,
		CallerNumber: to.Extension,
		Applications: []*model.CallRequestApplication{
			{
				AppName: "bridge",
				Args:    to.BridgeEndpoint(),
			},
		},
	}

	return invite, nil
}

func (app *App) createOutboundCallToDestination(domainId int64, req *model.OutboundCallRequest, from *model.UserCallInfo) (*model.CallRequest, *model.AppError) {
	invite := &model.CallRequest{
		Endpoints: from.GetCallEndpoints(),
		Variables: map[string]string{
			model.CALL_VARIABLE_DIRECTION:         model.CALL_DIRECTION_INTERNAL,
			model.CALL_VARIABLE_DISPLAY_DIRECTION: model.CALL_DIRECTION_OUTBOUND,
			model.CALL_VARIABLE_USER_ID:           fmt.Sprintf("%v", req.FromId),
			model.CALL_VARIABLE_DOMAIN_ID:         fmt.Sprintf("%v", domainId),

			"sip_h_X-Webitel-Destination": req.Destination,

			"hangup_after_bridge":        "true",
			"effective_caller_id_number": from.Extension,
			"effective_caller_id_name":   from.Name,
			"effective_callee_id_name":   req.Destination,
			"effective_callee_id_number": req.Destination,

			"origination_caller_id_name":   req.Destination,
			"origination_caller_id_number": req.Destination,
			"origination_callee_id_name":   from.Name,
			"origination_callee_id_number": from.Extension,
		},
		Destination:  req.Destination,
		Timeout:      req.Timeout,
		CallerName:   req.Destination,
		CallerNumber: req.Destination,
	}

	return invite, nil
}


*/
