package app

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/webitel/engine/call_manager"
	"github.com/webitel/engine/gen/cc"
	"github.com/webitel/engine/model"
	sqloptions "github.com/webitel/engine/store/sql_options"
	"github.com/webitel/wlog"
)

const (
	refreshMissedNotification = "refresh_missed"
)

func (app *App) CreateOutboundCall(ctx context.Context, domainId int64, req *model.OutboundCallRequest, variables map[string]string) (string, model.AppError) {
	var callCli call_manager.CallClient
	var err model.AppError
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
		return "", model.NewNotFoundError("app.call.create.not_found", "")
	}

	if req.From != nil && (req.From.UserId != nil || req.From.Extension != nil) {
		if from, err = app.GetCallInfoEndpoint(ctx, domainId, req.From, req.Params.IsOnline); err != nil {
			return "", err
		}
	} else {
		return "", model.NewBadRequestError("app.call.create.valid.from", "")
	}

	if from.IsBusy {
		return "", model.NewBadRequestError("app.call.create.valid.busy", "")
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
		if stat, err = app.Store.Agent().DistributeInfoByUserId(ctx, domainId, from.Id, "call"); err != nil {
			wlog.Error(err.Error())
		} else if stat.Busy {
			return "", model.NewBadRequestError("app.call.create.valid.agent", "Agent in call")
		} else if stat.Distribute {
			if _, err := app.cc.Member().CancelAgentDistribute(context.Background(), &cc.CancelAgentDistributeRequest{
				AgentId: stat.AgentId,
			}); err != nil {
				wlog.Error(err.Error())
			}
		}
	}

	if req.Params.HideNumber {
		invite.AddVariable("wbt_hide_number", "true")
	}

	if req.Params.ContactId > 0 {
		invite.AddVariable("wbt_contact_id", strconv.Itoa(int(req.Params.ContactId)))
	}

	id, err = callCli.MakeOutboundCall(invite)
	if err != nil {
		return "", err
	}

	return id, nil

}

func (app *App) RedialCall(ctx context.Context, domainId int64, userId int64, callId string) (string, model.AppError) {

	from, err := app.Store.Call().FromNumberWithUserIds(ctx, domainId, userId, callId)
	if err != nil {
		return "", err
	}

	req := &model.OutboundCallRequest{
		CreatedAt:   model.GetMillis(),
		CreatedById: userId,
		From: &model.EndpointRequest{
			UserId: &userId,
		},
		To:          nil,
		Destination: from.Number,
		Params: model.CallParameters{
			DisableStun: false,
			HideNumber:  false,
			Variables: map[string]string{
				"wbt_redial": callId,
			},
		},
	}

	var dest string
	dest, err = app.CreateOutboundCall(ctx, domainId, req, nil)
	if err != nil {
		return "", err
	}

	err = app.Store.Call().SetHideMissedAllParent(ctx, domainId, userId, callId)
	if err != nil {
		return "", err
	}

	if len(from.UserIds) != 0 {
		err = app.MessageQueue.SendNotification(domainId, &model.Notification{
			DomainId:  domainId,
			Action:    refreshMissedNotification,
			CreatedAt: model.GetMillis(),
			ForUsers:  from.UniqueUsers(),
			Body: map[string]interface{}{
				"id": callId,
			},
		})
	}

	return dest, nil
}

func (app *App) CallToQueue(ctx context.Context, domainId int64, userId int64, parentId string, cp model.CallParameters, queueId *int, agentId *int) (string, model.AppError) {
	var info *model.TransferInfo
	var cli call_manager.CallClient
	var name, number string

	usr, err := app.GetUserCallInfo(ctx, userId, domainId)
	if err != nil {
		return "", err
	}

	info, err = app.Store.Call().TransferInfo(ctx, parentId, domainId, queueId, agentId)
	if err != nil {
		return "", err
	}

	cli, err = app.getCallCli(ctx, domainId, "", info.AppId)
	if err != nil {
		return "", err
	}

	invite := &model.CallRequest{
		Endpoints: usr.GetCallEndpoints(),
		Variables: model.UnionStringMaps(
			usr.GetVariables(),
			cp.Variables,
			map[string]string{
				model.CALL_VARIABLE_DIRECTION:         model.CALL_DIRECTION_INTERNAL,
				model.CALL_VARIABLE_DISPLAY_DIRECTION: model.CALL_DIRECTION_OUTBOUND,
				model.CALL_VARIABLE_USER_ID:           fmt.Sprintf("%v", usr.Id),
				model.CALL_VARIABLE_DOMAIN_ID:         fmt.Sprintf("%v", domainId),

				"hangup_after_bridge":    "true",
				"ignore_display_updates": "true",
				"wbt_auto_answer":        "true",
				"wbt_parent_id":          info.Id,

				"origination_caller_id_number": usr.Endpoint,
				"effective_caller_id_number":   usr.Endpoint,
				"origination_caller_id_name":   usr.Name,
				"effective_caller_id_namer":    usr.Name,

				"wbt_from_id":     fmt.Sprintf("%v", usr.Id),
				"wbt_from_number": usr.Endpoint,
				"wbt_from_name":   usr.Name,
				"wbt_from_type":   model.EndpointTypeUser,
			},
		),
	}

	if queueId != nil && info.QueueName != nil {
		invite.AddVariable("wbt_bt_queue_id", fmt.Sprintf("%v", *queueId))
		invite.AddVariable("wbt_transfer_form", "false")
		name = *info.QueueName
		number = name
	} else if agentId != nil && info.AgentName != nil {
		invite.AddVariable("wbt_bt_agent_id", fmt.Sprintf("%v", *agentId))
		name = *info.AgentName
		if info.AgentExtension != nil {
			number = *info.AgentExtension
		}
	} else {
		// TODO
		return "", nil
	}

	invite.Destination = number
	invite.AddVariable("effective_callee_id_number", number)
	invite.AddVariable("effective_callee_id_name", name)

	var id string
	id, err = cli.MakeOutboundCall(invite)
	if err != nil {
		return "", err
	}

	return id, nil

}

func (app *App) GetCall(ctx context.Context, domainId int64, callId string) (*model.Call, model.AppError) {
	return app.Store.Call().Get(ctx, domainId, callId)
}

func (app *App) EavesdropCall(ctx context.Context, domainId, userId int64, req *model.EavesdropCall, variables map[string]string) (string, model.AppError) {
	var cli call_manager.CallClient
	var err model.AppError
	var usr *model.UserCallInfo
	var info *model.EavesdropInfo

	if req.From != nil && (req.From.UserId != nil || req.From.Extension != nil) {
		if usr, err = app.GetCallInfoEndpoint(ctx, domainId, req.From, false); err != nil {
			return "", err
		}
	} else {
		usr, err = app.GetUserCallInfo(ctx, userId, domainId)
		if err != nil {
			return "", err
		}
	}

	if usr == nil {
		return "", model.NewBadRequestError("app.call.eavesdrop.valid.user", "No user")
	}

	info, err = app.Store.Call().GetEavesdropInfo(ctx, domainId, req.Id)
	if err != nil {
		return "", err
	}

	cli, err = app.getCallCli(ctx, domainId, info.AgentCallId, &info.AppId)
	if err != nil {
		return "", err
	}

	invite := &model.CallRequest{
		Endpoints:   usr.GetCallEndpoints(),
		Destination: info.Agent.Number,
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
				"wbt_parent_id":                       info.ParentCallId,

				"wbt_destination": info.Agent.Number,
				"wbt_from_id":     fmt.Sprintf("%v", usr.Id),
				"wbt_from_number": usr.Endpoint,
				"wbt_from_name":   usr.Name,
				"wbt_from_type":   model.EndpointTypeUser,

				"wbt_to_id":     fmt.Sprintf("%v", info.Agent.Id),
				"wbt_to_name":   info.Agent.Name,
				"wbt_to_number": info.Agent.Number,
				"wbt_to_type":   info.Agent.Type,

				"effective_caller_id_number": usr.Extension,
				"effective_caller_id_name":   usr.Name,

				"effective_callee_id_name":   info.Agent.Name,
				"effective_callee_id_number": info.Agent.Number,

				"origination_caller_id_name":   info.Agent.Name,
				"origination_caller_id_number": info.Agent.Number,
				"origination_callee_id_name":   usr.Name,
				"origination_callee_id_number": usr.Extension,
			},
		),
		CallerName:   info.Agent.Name,
		CallerNumber: info.Agent.Number,
		Applications: []*model.CallRequestApplication{
			{
				AppName: "eavesdrop",
				Args:    info.AgentCallId,
			},
		},
	}

	/*
	    {
	     "id": "85b9366a-4c2f-45b0-bcea-06a38fe4be37",
	     "control": true,
	     "listenA": true,
	     "listenB": true
	   }
	*/

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

	//if req.Notify {
	invite.AddVariable("wbt_eavesdrop_type", "notify")
	//} else {
	//	invite.AddVariable("wbt_eavesdrop_type", "hide")
	//}

	invite.AddVariable("wbt_eavesdrop_agent_id", info.AgentCallId)
	invite.AddVariable("wbt_eavesdrop_state", req.StateName()) // todo remove WhisperALeg && WhisperBLeg
	invite.AddVariable("wbt_eavesdrop_name", info.Client.Name)
	invite.AddVariable("wbt_eavesdrop_number", info.Client.Number)
	invite.AddVariable("wbt_eavesdrop_duration", fmt.Sprintf("%d", info.Duration))

	var id string
	id, err = cli.MakeOutboundCall(invite)
	if err != nil {
		return "", err
	}

	return id, nil
}

func (app *App) EavesdropCallState(ctx context.Context, domainId, userId int64, req *model.EavesdropCall) model.AppError {
	cli, err := app.getCallCli(ctx, domainId, req.Id, nil)
	if err != nil {
		return err
	}

	err = cli.SetEavesdropState(req.Id, req.State)
	if err != nil {
		return err
	}

	return nil
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

func (app *App) GetActiveCallPage(ctx context.Context, domainId int64, search *model.SearchCall) ([]*model.Call, bool, model.AppError) {
	list, err := app.Store.Call().GetActive(ctx, domainId, search)
	if err != nil {
		return nil, false, err
	}
	search.RemoveLastElemIfNeed(&list)
	return list, search.EndOfList(), nil
}

func (app *App) GetActiveCallPageByGroups(ctx context.Context, domainId int64, userSupervisorId int64, groups []int, search *model.SearchCall) ([]*model.Call, bool, model.AppError) {
	list, err := app.Store.Call().GetActiveByGroups(ctx, domainId, userSupervisorId, groups, search)
	if err != nil {
		return nil, false, err
	}
	search.RemoveLastElemIfNeed(&list)
	return list, search.EndOfList(), nil
}

func (app *App) GetUserActiveCalls(ctx context.Context, domainId, userId int64) ([]*model.Call, model.AppError) {
	return app.Store.Call().GetUserActiveCall(ctx, domainId, userId)
}

func (app *App) GetHistoryCallPage(ctx context.Context, domainId int64, userId int64, search *model.SearchHistoryCall) ([]*model.HistoryCall, bool, model.AppError) {
	userGlobalGrantOption := sqloptions.WithUserGrantFilterOption(uint(userId), model.GLOBAL_SELECT_GRANT)

	list, err := app.Store.Call().GetHistory(ctx, domainId, search, userGlobalGrantOption)
	if err != nil {
		return nil, false, err
	}
	search.RemoveLastElemIfNeed(&list)
	return list, search.EndOfList(), nil
}

func (app *App) GetHistoryCallPageByGroups(ctx context.Context, domainId int64, userSupervisorId int64, groups []int, search *model.SearchHistoryCall) ([]*model.HistoryCall, bool, model.AppError) {
	list, err := app.Store.Call().GetHistoryByGroups(ctx, domainId, userSupervisorId, groups, search)
	if err != nil {
		return nil, false, err
	}
	search.RemoveLastElemIfNeed(&list)
	return list, search.EndOfList(), nil
}

func (app *App) GetAggregateHistoryCallPage(ctx context.Context, domainId int64, aggs *model.CallAggregate) ([]*model.AggregateResult, model.AppError) {
	return app.Store.Call().Aggregate(ctx, domainId, aggs)
}

func (app *App) getCallCli(ctx context.Context, domainId int64, id string, appId *string) (cli call_manager.CallClient, err model.AppError) {

	if appId != nil {
		cli, err = app.CallManager().CallClientById(*appId)
	} else {
		var info *model.CallInstance
		info, err = app.Store.Call().GetInstance(ctx, domainId, id)
		if err != nil {
			return nil, err
		}
		cli, err = app.CallManager().CallClientById(*info.AppId)
	}
	return
}

func (app *App) HangupCall(ctx context.Context, domainId int64, req *model.HangupCall) model.AppError {
	var cli call_manager.CallClient
	var err model.AppError
	var cause = ""

	cli, err = app.getCallCli(ctx, domainId, req.Id, req.AppId)
	if err != nil {
		return err
	}

	if req.Cause != nil {
		cause = *req.Cause
	}

	err = cli.HangupCall(req.Id, cause)
	if err == call_manager.NotFoundCall {
		var e *model.CallServiceHangup
		if e, err = app.Store.Call().SetEmptySeverCall(ctx, domainId, req.Id); err == nil {
			//fixme rollback
			err = app.MessageQueue.SendStickingCall(e)
		} else if err.GetStatusCode() == http.StatusNotFound {
			err = nil
		}
	}

	return err
}

func (app *App) ConfirmPushCall(domainId int64, callId string) model.AppError {
	var cli call_manager.CallClient
	var err model.AppError

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

func (app *App) HoldCall(ctx context.Context, domainId int64, req *model.UserCallRequest) model.AppError {
	var cli call_manager.CallClient
	var err model.AppError

	cli, err = app.getCallCli(ctx, domainId, req.Id, req.AppId)
	if err != nil {
		return err
	}

	return cli.Hold(req.Id)
}

func (app *App) UnHoldCall(ctx context.Context, domainId int64, req *model.UserCallRequest) model.AppError {
	var cli call_manager.CallClient
	var err model.AppError

	cli, err = app.getCallCli(ctx, domainId, req.Id, req.AppId)
	if err != nil {
		return err
	}

	return cli.UnHold(req.Id)
}

func (app *App) DtmfCall(ctx context.Context, domainId int64, req *model.DtmfCall) model.AppError {
	var cli call_manager.CallClient
	var err model.AppError

	cli, err = app.getCallCli(ctx, domainId, req.Id, req.AppId)
	if err != nil {
		return err
	}

	return cli.DTMF(req.Id, req.Digit)
}

func (app *App) BlindTransferCallToQueue(ctx context.Context, domainId int64, req *model.BlindTransferCallToQueue) model.AppError {
	if req.Variables == nil {
		req.Variables = make(map[string]string)
	}

	q := strconv.Itoa(req.QueueId)

	req.Variables["wbt_bt_queue_id"] = q
	req.Variables["wbt_bt_queue"] = "true"

	return app.BlindTransferCallExt(ctx, domainId, &model.BlindTransferCall{
		UserCallRequest: req.UserCallRequest,
		Destination:     q,
		Variables:       req.Variables,
	})
}

func (app *App) BlindTransferCall(ctx context.Context, domainId int64, req *model.BlindTransferCall) model.AppError {
	var cli call_manager.CallClient
	var err model.AppError
	var info *model.BlindTransferInfo

	cli, err = app.getCallCli(ctx, domainId, req.Id, req.AppId)
	if err != nil {
		return err
	}

	info, err = app.Store.Call().BlindTransferInfo(ctx, req.Id)
	if err != nil {
		return err
	}

	if info.QueueUnanswered {
		return model.NewBadRequestError("app.call.transfer", "сannot transfer an unanswered call from the queue.")
	}

	var v map[string]string
	if info.ContactId != nil {
		v = map[string]string{
			"wbt_contact_id": fmt.Sprintf("%d", *info.ContactId),
		}
	}

	return cli.BlindTransferExt(info.Id, req.Destination, v)
}

func (app *App) BlindTransferCallExt(ctx context.Context, domainId int64, req *model.BlindTransferCall) model.AppError {
	var cli call_manager.CallClient
	var err model.AppError
	var id string

	cli, err = app.getCallCli(ctx, domainId, req.Id, req.AppId)
	if err != nil {
		return err
	}

	id, err = app.Store.Call().BridgedId(ctx, req.Id)
	if err != nil {
		return err
	}

	return cli.BlindTransferExt(id, req.Destination, req.Variables)
}

func (app *App) BridgeCall(ctx context.Context, domainId int64, fromId, toId string, vars map[string]string) model.AppError {
	var cli call_manager.CallClient
	info, err := app.Store.Call().BridgeInfo(ctx, domainId, fromId, toId)
	if err != nil {
		return err
	}

	cli, err = app.getCallCli(ctx, domainId, info.FromId, &info.AppId)
	if err != nil {
		return err
	}

	/* TODO https://webitel.atlassian.net/browse/WTEL-5591
	if info.ContactId != nil {
		if vars == nil {
			vars = make(map[string]string)
		}
		vars["wbt_contact_id"] = fmt.Sprintf("%d", *info.ContactId)
	}
	*/

	_, err = cli.BridgeCall(info.FromId, info.ToId, vars)
	return err
}

func (app *App) SetCallVariables(ctx context.Context, domainId int64, id string, vars map[string]string) model.AppError {
	domain, err := app.Store.Call().SetVariables(ctx, domainId, id, vars)
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

func (app *App) SetCallParams(ctx context.Context, domainId int64, id string, params model.CallParameters) model.AppError {
	var cli call_manager.CallClient
	var err model.AppError

	cli, err = app.getCallCli(ctx, domainId, id, nil)
	if err != nil {
		return err
	}

	if len(params.Variables) != 0 {
		err = cli.SetCallVariables(id, params.Variables)
	}

	if err != nil {
		return err
	}

	return nil
}

func (app *App) GetLastCallFile(ctx context.Context, domainId int64, callId string) (int64, model.AppError) {
	return app.Store.Call().LastFile(ctx, domainId, callId)
}

func (app *App) CreateCallAnnotation(ctx context.Context, domainId int64, annotation *model.CallAnnotation) (*model.CallAnnotation, model.AppError) {
	_, err := app.Store.Call().GetHistory(ctx, domainId, &model.SearchHistoryCall{
		Ids: []string{annotation.CallId},
	})
	if err != nil {
		return nil, err
	}

	return app.Store.Call().CreateAnnotation(ctx, annotation)
}

func (app *App) UpdateCallAnnotation(ctx context.Context, domainId int64, annotation *model.CallAnnotation) (*model.CallAnnotation, model.AppError) {
	_, err := app.Store.Call().GetHistory(ctx, domainId, &model.SearchHistoryCall{
		Ids: []string{annotation.CallId},
	})
	if err != nil {
		return nil, err
	}
	var oldAnnotation *model.CallAnnotation
	oldAnnotation, err = app.Store.Call().GetAnnotation(ctx, annotation.Id)
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

	return app.Store.Call().UpdateAnnotation(ctx, domainId, oldAnnotation)
}

func (app *App) DeleteCallAnnotation(ctx context.Context, domainId, id int64, callId string) (*model.CallAnnotation, model.AppError) {
	_, err := app.Store.Call().GetHistory(ctx, domainId, &model.SearchHistoryCall{
		Ids: []string{callId},
	})
	if err != nil {
		return nil, err
	}
	var annotation *model.CallAnnotation
	annotation, err = app.Store.Call().GetAnnotation(ctx, id)
	if err != nil {
		return nil, err
	}

	err = app.Store.Call().DeleteAnnotation(ctx, id)
	if err != nil {
		return nil, err
	}

	return annotation, nil
}

func (app *App) UpdateHistoryCall(ctx context.Context, domainId int64, id string, p *model.HistoryCallPatch) (*model.HistoryCall, model.AppError) {
	err := app.Store.Call().UpdateHistoryCall(ctx, domainId, id, p)
	if err != nil {
		return nil, err
	}

	var list []*model.HistoryCall
	list, err = app.Store.Call().GetHistory(ctx, domainId, &model.SearchHistoryCall{
		Ids: []string{id},
		ListRequest: model.ListRequest{
			Page:    1,
			PerPage: 1,
		},
	})

	if err != nil {
		return nil, err
	}

	if len(list) == 0 {
		return nil, model.NewNotFoundError("app.call.update.not_found", "Not found call")
	}

	return list[0], nil
}

/*

func (app *App) createOutboundCallToUser(domainId int64, req *model.OutboundCallRequest, from, to *model.UserCallInfo) (*model.CallRequest, model.AppError) {
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

func (app *App) createOutboundCallToDestination(domainId int64, req *model.OutboundCallRequest, from *model.UserCallInfo) (*model.CallRequest, model.AppError) {
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
