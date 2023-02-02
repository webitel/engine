package app

import (
	"context"
	"fmt"
	"github.com/webitel/engine/call_manager"
	"github.com/webitel/engine/model"
	"github.com/webitel/protos/cc"
	"github.com/webitel/wlog"
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

	if req.From != nil && (req.From.UserId != nil || req.From.Extension != nil) {
		if from, err = app.GetCallInfoEndpoint(domainId, req.From, req.Params.IsOnline); err != nil {
			return "", err
		}
	} else {
		return "", model.NewAppError("CreateOutboundCall", "app.call.create.valid.from", nil, "", http.StatusBadRequest)
	}

	if from.IsBusy {
		return "", model.NewAppError("CreateOutboundCall", "app.call.create.valid.busy", nil, "", http.StatusBadRequest)
	}

	invite := inviteFromUser(domainId, req, from)
	for k, v := range req.Params.Variables {
		invite.AddUserVariable(k, v)
	}
	for k, v := range variables {
		invite.AddVariable(k, v)
	}

	//invite.AddVariable("media_webrtc", "true")

	if req.Params.Video {
		invite.AddVariable(model.CALL_VARIABLE_USE_VIDEO, "true")
	}

	if req.Params.Screen {
		invite.AddVariable(model.CALL_VARIABLE_USE_SCREEN, "true")
	}

	if !req.Params.DisableAutoAnswer {
		invite.AddVariable(model.CALL_VARIABLE_SIP_AUTO_ANSWER, "true")
		//FIXME
		invite.AddVariable("wbt_auto_answer", "true")
	}

	if from.HasPush {
		invite.AddVariable("execute_on_originate", "wbt_send_hook")
	}

	if req.Params.DisableStun {
		invite.AddVariable("wbt_disable_stun", "true")
	}

	if !(req.Params.Video || req.Params.Screen) {
		invite.AddVariable("absolute_codec_string", "opus,pcmu,pcma")
	}

	if req.Params.Display != "" {
		invite.AddVariable("sip_h_X-Webitel-Display", req.Params.Display)
	}

	if req.Params.CancelDistribute {
		var stat *model.DistributeAgentInfo
		if stat, err = app.Store.Agent().DistributeInfoByUserId(domainId, from.Id); err != nil {
			wlog.Error(err.Error())
		} else if stat.Busy {
			return "", model.NewAppError("CreateOutboundCall", "app.call.create.valid.agent", nil, "Agent in call", http.StatusBadRequest)
		} else if stat.Distribute {
			if _, err := app.cc.Member().CancelAgentDistribute(context.Background(), &cc.CancelAgentDistributeRequest{
				AgentId: stat.AgentId,
			}); err != nil {
				wlog.Error(err.Error())
			}
		}
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
	var err *model.AppError
	var usr *model.UserCallInfo

	if req.From != nil && (req.From.UserId != nil || req.From.Extension != nil) {
		if usr, err = app.GetCallInfoEndpoint(domainId, req.From, false); err != nil {
			return "", err
		}
	} else {
		usr, err = app.GetUserCallInfo(userId, domainId)
		if err != nil {
			return "", err
		}
	}

	if usr == nil {

	}

	call, err = app.GetCall(domainId, req.Id)
	if err != nil {
		return "", err
	}

	cli, err = app.getCallCli(domainId, req.Id, req.AppId)
	if err != nil {
		return "", err
	}

	var agent, client model.Endpoint

	if call.Direction == model.CALL_DIRECTION_INBOUND {
		client = *call.From
		if call.To != nil {
			agent = *call.To
		}
	} else {
		agent = *call.From
		if call.To != nil {
			client = *call.To
		}
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
				"wbt_auto_answer":                     "true",
				"wbt_parent_id":                       call.Id,

				"wbt_destination": call.Destination,
				"wbt_from_id":     fmt.Sprintf("%v", usr.Id),
				"wbt_from_number": usr.Endpoint,
				"wbt_from_name":   usr.Name,
				"wbt_from_type":   model.EndpointTypeUser,

				"wbt_to_id":     fmt.Sprintf("%v", agent.Id),
				"wbt_to_name":   agent.Name,
				"wbt_to_number": agent.Number,
				"wbt_to_type":   agent.Type,

				"effective_caller_id_number": usr.Extension,
				"effective_caller_id_name":   usr.Name,

				"effective_callee_id_name":   agent.Name,
				"effective_callee_id_number": agent.Number,

				"origination_caller_id_name":   agent.Name,
				"origination_caller_id_number": agent.Number,
				"origination_callee_id_name":   usr.Name,
				"origination_callee_id_number": usr.Extension,
			},
		),
		CallerName:   agent.Name,
		CallerNumber: agent.Number,
		Applications: []*model.CallRequestApplication{
			{
				AppName: "eavesdrop",
				Args:    call.Id,
			},
		},
	}

	if req.Dtmf {
		invite.AddVariable("eavesdrop_enable_dtmf", "true")
	} else {
		invite.AddVariable("eavesdrop_enable_dtmf", "false")
	}

	if req.ALeg {
		invite.AddVariable("eavesdrop_bridge_aleg", "true")
	} else {
		invite.AddVariable("eavesdrop_bridge_aleg", "false")
	}

	if req.BLeg {
		invite.AddVariable("eavesdrop_bridge_bleg", "true")
	} else {
		invite.AddVariable("eavesdrop_bridge_bleg", "false")
	}

	if req.WhisperALeg {
		invite.AddVariable("eavesdrop_whisper_aleg", "true")
	} else {
		invite.AddVariable("eavesdrop_whisper_aleg", "false")
	}

	if req.WhisperBLeg {
		invite.AddVariable("eavesdrop_whisper_bleg", "true")
	} else {
		invite.AddVariable("eavesdrop_whisper_bleg", "false")
	}

	invite.AddUserVariable("wbt_eavesdrop_type", "hide")
	invite.AddUserVariable("wbt_eavesdrop_name", client.Name)
	invite.AddUserVariable("wbt_eavesdrop_number", client.Number)
	invite.AddUserVariable("wbt_eavesdrop_duration", fmt.Sprintf("%d", call.Duration))

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

				//"effective_caller_id_number": usr.Extension,
				//"effective_caller_id_name":   usr.Name,
				//"effective_callee_id_name":   req.Destination,
				//"effective_callee_id_number": req.Destination,

				//"origination_caller_id_name":   req.Destination,
				//"origination_caller_id_number": req.Destination,
				//"origination_callee_id_name":   usr.Name,
				//"origination_callee_id_number": usr.Extension,
			},
		),
		//CallerName:   usr.Name,
		//CallerNumber: usr.Extension,
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

func (app *App) GetActiveCallPageByGroups(domainId int64, userSupervisorId int64, groups []int, search *model.SearchCall) ([]*model.Call, bool, *model.AppError) {
	list, err := app.Store.Call().GetActiveByGroups(domainId, userSupervisorId, groups, search)
	if err != nil {
		return nil, false, err
	}
	search.RemoveLastElemIfNeed(&list)
	return list, search.EndOfList(), nil
}

func (app *App) GetUserActiveCalls(domainId, userId int64) ([]*model.Call, *model.AppError) {
	return app.Store.Call().GetUserActiveCall(domainId, userId)
}

func (app *App) GetHistoryCallPage(domainId int64, search *model.SearchHistoryCall) ([]*model.HistoryCall, bool, *model.AppError) {
	list, err := app.Store.Call().GetHistory(domainId, search)
	if err != nil {
		return nil, false, err
	}
	search.RemoveLastElemIfNeed(&list)
	return list, search.EndOfList(), nil
}

func (app *App) GetHistoryCallPageByGroups(domainId int64, userSupervisorId int64, groups []int, search *model.SearchHistoryCall) ([]*model.HistoryCall, bool, *model.AppError) {
	list, err := app.Store.Call().GetHistoryByGroups(domainId, userSupervisorId, groups, search)
	if err != nil {
		return nil, false, err
	}
	search.RemoveLastElemIfNeed(&list)
	return list, search.EndOfList(), nil
}

func (app *App) GetAggregateHistoryCallPage(domainId int64, aggs *model.CallAggregate) ([]*model.AggregateResult, *model.AppError) {
	return app.Store.Call().Aggregate(domainId, aggs)
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

	err = cli.HangupCall(req.Id, cause)
	if err == call_manager.NotFoundCall {
		var e *model.CallServiceHangup
		if e, err = app.Store.Call().SetEmptySeverCall(domainId, req.Id); err == nil {
			//fixme rollback
			err = app.MessageQueue.SendStickingCall(e)
		}
	}

	return err
}

func (app *App) ConfirmPushCall(domainId int64, callId string) *model.AppError {
	var cli call_manager.CallClient
	var err *model.AppError

	//todo get from store
	cli, err = app.CallManager().CallClient() //app.getCallCli(domainId, callId, nil)
	if err != nil {
		return err
	}

	err = cli.ConfirmPushCall(callId)
	if err != nil {
		return err
	}

	return nil
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
	var id string

	cli, err = app.getCallCli(domainId, req.Id, req.AppId)
	if err != nil {
		return err
	}

	id, err = app.Store.Call().BridgedId(req.Id)
	if err != nil {
		return err
	}

	return cli.BlindTransfer(id, req.Destination)
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

func (app *App) SetCallVariables(domainId int64, id string, vars map[string]string) *model.AppError {
	domain, err := app.Store.Call().SetVariables(domainId, id, vars)
	if err != nil {
		return err
	}
	if domain.AppId != nil {
		//var cli call_manager.CallClient
		//cli, err = app.getCallCli(domainId, id, domain.AppId)
		//if err != nil {
		//	return err
		//}
		//err = cli.SetCallVariables(id, vars)
	}

	return err
}

func (app *App) GetLastCallFile(domainId int64, callId string) (int64, *model.AppError) {
	return app.Store.Call().LastFile(domainId, callId)
}

func (app *App) CreateCallAnnotation(domainId int64, annotation *model.CallAnnotation) (*model.CallAnnotation, *model.AppError) {
	_, err := app.Store.Call().GetHistory(domainId, &model.SearchHistoryCall{
		Ids: []string{annotation.CallId},
	})
	if err != nil {
		return nil, err
	}

	return app.Store.Call().CreateAnnotation(annotation)
}

func (app *App) UpdateCallAnnotation(domainId int64, annotation *model.CallAnnotation) (*model.CallAnnotation, *model.AppError) {
	_, err := app.Store.Call().GetHistory(domainId, &model.SearchHistoryCall{
		Ids: []string{annotation.CallId},
	})
	if err != nil {
		return nil, err
	}
	var oldAnnotation *model.CallAnnotation
	oldAnnotation, err = app.Store.Call().GetAnnotation(annotation.Id)
	if err != nil {
		return nil, err
	}

	oldAnnotation.UpdatedAt = annotation.UpdatedAt
	oldAnnotation.UpdatedBy = annotation.UpdatedBy
	oldAnnotation.Note = annotation.Note
	oldAnnotation.StartSec = annotation.StartSec
	oldAnnotation.EndSec = annotation.EndSec

	if err = oldAnnotation.IsValid(); err != nil {
		return nil, err
	}

	return app.Store.Call().UpdateAnnotation(domainId, oldAnnotation)
}

func (app *App) DeleteCallAnnotation(domainId, id int64, callId string) (*model.CallAnnotation, *model.AppError) {
	_, err := app.Store.Call().GetHistory(domainId, &model.SearchHistoryCall{
		Ids: []string{callId},
	})
	if err != nil {
		return nil, err
	}
	var annotation *model.CallAnnotation
	annotation, err = app.Store.Call().GetAnnotation(id)
	if err != nil {
		return nil, err
	}

	err = app.Store.Call().DeleteAnnotation(id)
	if err != nil {
		return nil, err
	}

	return annotation, nil
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
