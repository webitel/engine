package grpc_api

import (
	"context"
	"fmt"
	"github.com/webitel/engine/auth_manager"
	"github.com/webitel/engine/model"
	"github.com/webitel/protos/engine"
	"net/http"
	"strings"
)

type call struct {
	*API
	minimumNumberMaskLen int
	prefixNumberMaskLen  int
	suffixNumberMaskLen  int
	engine.UnsafeCallServiceServer
}

func NewCallApi(api *API, minimumNumberMaskLen, prefixNumberMaskLen, suffixNumberMaskLen int) *call {
	return &call{
		API:                  api,
		minimumNumberMaskLen: minimumNumberMaskLen,
		prefixNumberMaskLen:  prefixNumberMaskLen,
		suffixNumberMaskLen:  suffixNumberMaskLen,
	}
}

func (api *call) SearchHistoryCall(ctx context.Context, in *engine.SearchHistoryCallRequest) (*engine.ListHistoryCall, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	if in.GetCreatedAt() == nil && in.GetStoredAt() == nil && in.GetNumber() == "" && in.GetQ() == "" && in.GetDependencyId() == nil && in.GetId() == nil {
		return nil, model.NewAppError("GRPC.SearchHistoryCall", "grpc.call.search_history", nil, "filter created_at or stored_at or q is required", http.StatusBadRequest)
	}

	var list []*model.HistoryCall
	var endList bool
	req := &model.SearchHistoryCall{
		ListRequest: model.ListRequest{
			DomainId: in.GetDomainId(),
			Page:     int(in.GetPage()),
			PerPage:  int(in.GetSize()),
			Q:        in.GetQ(),
			Sort:     in.Sort,
			Fields:   in.Fields,
		},
		SkipParent:       in.GetSkipParent(),
		UserIds:          in.GetUserId(),
		QueueIds:         in.GetQueueId(),
		TeamIds:          in.GetTeamId(),
		AgentIds:         in.GetAgentId(),
		MemberIds:        in.GetMemberId(),
		GatewayIds:       in.GetGatewayId(),
		Ids:              in.GetId(),
		TransferFromIds:  in.GetTransferFrom(),
		TransferToIds:    in.GetTransferTo(),
		DependencyIds:    in.GetDependencyId(),
		Tags:             in.GetTags(),
		CauseArr:         in.GetCause(),
		Variables:        in.GetVariables(),
		Number:           in.GetNumber(),
		AmdResult:        in.GetAmdResult(),
		HasFile:          GetBool(in.GetHasFile()),
		HasTranscript:    GetBool(in.GetHasTranscript()),
		AgentDescription: in.GetAgentDescription(),
		OwnerIds:         in.GetOwnerId(),
		GranteeIds:       in.GetGranteeId(),
		AmdAiResult:      in.GetAmdAiResult(),
	}

	if in.GetDuration() != nil {
		req.Duration = &model.FilterBetween{
			From: in.GetDuration().GetFrom(),
			To:   in.GetDuration().GetTo(),
		}
	}

	if in.GetAnsweredAt() != nil {
		req.AnsweredAt = &model.FilterBetween{
			From: in.GetAnsweredAt().GetFrom(),
			To:   in.GetAnsweredAt().GetTo(),
		}
	}

	if in.GetCreatedAt() != nil {
		req.CreatedAt = &model.FilterBetween{
			From: in.GetCreatedAt().GetFrom(),
			To:   in.GetCreatedAt().GetTo(),
		}
	}

	if in.GetStoredAt() != nil {
		req.StoredAt = &model.FilterBetween{
			From: in.GetStoredAt().GetFrom(),
			To:   in.GetStoredAt().GetTo(),
		}
	}

	if in.GetDirection() != "" {
		req.Direction = &in.Direction
	}

	if in.GetParentId() != "" {
		req.ParentId = &in.ParentId
	}

	if in.GetMissed() {
		req.Missed = model.NewBool(true)
	}

	if in.Fts != "" {
		req.Fts = &in.Fts
	}

	if list, endList, err = api.ctrl.SearchHistoryCall(session, req); err != nil {
		return nil, err
	}

	items := make([]*engine.HistoryCall, 0, len(list))

	//todo
	accessString := !session.HasAction(auth_manager.PERMISSION_VIEW_NUMBERS) &&
		!((len(in.UserId) == 1 && in.UserId[0] == session.UserId) && in.Missed && (len(in.Cause) == 2 && in.Cause[0] == "NO_ANSWER" && in.Cause[1] == "ORIGINATOR_CANCEL"))
	for _, v := range list {
		items = append(items, toEngineHistoryCall(
			v,
			api.minimumNumberMaskLen,
			api.prefixNumberMaskLen,
			api.suffixNumberMaskLen,
			accessString,
			session.HasAction(auth_manager.PERMISSION_RECORD_FILE),
		))
	}

	return &engine.ListHistoryCall{
		Next:  !endList,
		Items: items,
	}, nil
}

func (api *call) AggregateHistoryCall(ctx context.Context, in *engine.AggregateHistoryCallRequest) (*engine.ListAggregate, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	if in.GetCreatedAt() == nil && in.GetStoredAt() == nil {
		return nil, model.NewAppError("GRPC.SearchHistoryCall", "grpc.call.search_history", nil, "filter created_at or stored_at is required", http.StatusBadRequest)
	}

	//var list []*model.HistoryCall
	//var endList bool
	req := &model.CallAggregate{
		SearchHistoryCall: model.SearchHistoryCall{
			ListRequest: model.ListRequest{
				DomainId: in.GetDomainId(),
				Page:     int(in.GetPage()),
				PerPage:  int(in.GetSize()),
				Q:        in.GetQ(),
			},
			SkipParent:       in.GetSkipParent(),
			UserIds:          in.GetUserId(),
			QueueIds:         in.GetQueueId(),
			TeamIds:          in.GetTeamId(),
			AgentIds:         in.GetAgentId(),
			MemberIds:        in.GetMemberId(),
			GatewayIds:       in.GetGatewayId(),
			Ids:              in.GetId(),
			TransferFromIds:  in.GetTransferFrom(),
			TransferToIds:    in.GetTransferTo(),
			DependencyIds:    in.GetDependencyId(),
			Tags:             in.GetTags(),
			CauseArr:         in.GetCause(),
			Variables:        in.GetVariables(),
			Number:           in.GetNumber(),
			AmdResult:        in.GetAmdResult(),
			HasFile:          GetBool(in.GetHasFile()),
			HasTranscript:    GetBool(in.GetHasTranscript()),
			AgentDescription: in.GetAgentDescription(),
		},
	}

	if in.GetDuration() != nil {
		req.Duration = &model.FilterBetween{
			From: in.GetDuration().GetFrom(),
			To:   in.GetDuration().GetTo(),
		}
	}

	if in.GetAnsweredAt() != nil {
		req.AnsweredAt = &model.FilterBetween{
			From: in.GetAnsweredAt().GetFrom(),
			To:   in.GetAnsweredAt().GetTo(),
		}
	}

	if in.GetCreatedAt() != nil {
		req.CreatedAt = &model.FilterBetween{
			From: in.GetCreatedAt().GetFrom(),
			To:   in.GetCreatedAt().GetTo(),
		}
	}

	if in.GetStoredAt() != nil {
		req.StoredAt = &model.FilterBetween{
			From: in.GetStoredAt().GetFrom(),
			To:   in.GetStoredAt().GetTo(),
		}
	}

	if in.GetDirection() != "" {
		req.Direction = &in.Direction
	}

	if in.GetParentId() != "" {
		req.ParentId = &in.ParentId
	}

	if in.GetMissed() {
		req.Missed = model.NewBool(true)
	}

	if in.Fts != "" {
		req.Fts = &in.Fts
	}

	for _, v := range in.Aggs {
		a := model.Aggregate{
			Name: v.Name,
			AggregateMetrics: model.AggregateMetrics{
				Min:   v.Min,
				Max:   v.Max,
				Avg:   v.Avg,
				Sum:   v.Sum,
				Count: v.Count,
			},
			Sort:  v.Sort,
			Limit: v.Limit,
		}

		if v.Group != nil {
			a.Group = make([]model.AggregateGroup, 0, len(v.Group))
			for _, j := range v.Group {
				a.Group = append(a.Group, model.AggregateGroup{
					Id:       j.Id,
					Interval: getInterval(j.Interval), //TODO

					Aggregate: j.Aggregate,
					Field:     j.Field,
					Top:       j.Top,
					Desc:      j.Desc,
				})
			}
		}
		req.Aggs = append(req.Aggs, a)
	}

	list, err := api.ctrl.AggregateHistoryCall(session, req)
	if err != nil {
		return nil, err
	}

	items := make([]*engine.AggregateResult, 0, len(list))

	for _, v := range list {
		i := &engine.AggregateResult{
			Name: v.Name,
			Data: UnmarshalJsonpb(v.Data),
		}
		items = append(items, i)
	}

	return &engine.ListAggregate{
		Items: items,
	}, nil
}

// TODO delete me
func getInterval(in string) string {
	if in == "auto" {
		return "1 hour"
	}

	return in
}

func (api *call) ReadCall(ctx context.Context, in *engine.ReadCallRequest) (*engine.ActiveCall, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	var call *model.Call
	call, err = api.ctrl.GetCall(session, in.DomainId, in.Id)

	if err != nil {
		return nil, err
	}

	return toEngineCall(call), nil
}

func (api *call) SearchActiveCall(ctx context.Context, in *engine.SearchCallRequest) (*engine.ListCall, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	var list []*model.Call
	var endList bool
	req := &model.SearchCall{
		ListRequest: model.ListRequest{
			DomainId: in.GetDomainId(),
			Q:        in.GetQ(),
			Page:     int(in.GetPage()),
			PerPage:  int(in.GetSize()),
			Sort:     in.GetSort(),
		},
		Direction:     in.Direction,
		SkipParent:    in.GetSkipParent(),
		UserIds:       in.GetUserId(),
		QueueIds:      in.GetQueueId(),
		TeamIds:       in.GetTeamId(),
		AgentIds:      in.GetAgentId(),
		MemberIds:     in.GetMemberId(),
		GatewayIds:    in.GetGatewayId(),
		SupervisorIds: in.GetSupervisorId(),
		State:         in.GetState(),
	}

	if in.GetDuration() != nil {
		req.Duration = &model.FilterBetween{
			From: in.GetDuration().GetFrom(),
			To:   in.GetDuration().GetTo(),
		}
	}

	if in.GetAnsweredAt() != nil {
		req.AnsweredAt = &model.FilterBetween{
			From: in.GetAnsweredAt().GetFrom(),
			To:   in.GetAnsweredAt().GetTo(),
		}
	}

	if in.GetCreatedAt() != nil {
		req.CreatedAt = &model.FilterBetween{
			From: in.GetCreatedAt().GetFrom(),
			To:   in.GetCreatedAt().GetTo(),
		}
	}

	if in.GetParentId() != "" {
		req.ParentId = &in.ParentId
	}

	if in.GetNumber() != "" {
		req.Number = model.NewString(in.Number)
	}

	if in.GetMissed() {
		req.Missed = model.NewBool(true)
	}

	list, endList, err = api.ctrl.SearchCall(session, req)

	if err != nil {
		return nil, err
	}

	items := make([]*engine.ActiveCall, 0, len(list))
	for _, v := range list {
		items = append(items, toEngineCall(v))
	}
	return &engine.ListCall{
		Next:  !endList,
		Items: items,
	}, nil
}

func (api *call) CreateCall(ctx context.Context, in *engine.CreateCallRequest) (*engine.CreateCallResponse, error) {
	session, err := api.ctrl.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	var req = &model.OutboundCallRequest{
		Destination: in.GetDestination(),
		Params: model.CallParameters{
			Timeout:           int(in.GetParams().GetTimeout()),
			Audio:             in.GetParams().GetAudio(),
			Video:             in.GetParams().GetVideo(),
			Screen:            in.GetParams().GetScreen(),
			Record:            in.GetParams().GetRecord(),
			Variables:         in.GetParams().GetVariables(),
			DisableAutoAnswer: in.GetParams().GetDisableAutoAnswer(),
			Display:           in.GetParams().GetDisplay(),
			DisableStun:       in.GetParams().GetDisableStun(),
			CancelDistribute:  in.GetParams().GetCancelDistribute(),
			IsOnline:          in.GetParams().GetIsOnline(),
		},
	}

	if in.To != nil {
		req.To = &model.EndpointRequest{}
		if in.To.AppId != "" {
			req.To.AppId = model.NewString(in.To.AppId)
		}

		if in.To.Id != 0 {
			req.To.UserId = model.NewInt64(in.To.Id)
		}

	}

	if in.From != nil {
		req.From = &model.EndpointRequest{}
		if in.From.AppId != "" {
			req.From.AppId = model.NewString(in.From.AppId)
		}

		if in.From.Id != 0 {
			req.From.UserId = model.NewInt64(in.From.Id)
		}

		if in.From.Extension != "" {
			req.From.Extension = model.NewString(in.From.Extension)
		}
	}

	var id string
	id, err = api.ctrl.CreateCall(session, req, in.GetParams().GetVariables())
	if err != nil {
		return nil, err
	}

	return &engine.CreateCallResponse{
		Id: id,
	}, nil
}

func (api *call) HangupCall(ctx context.Context, in *engine.HangupCallRequest) (*engine.HangupCallResponse, error) {
	session, err := api.ctrl.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	req := model.HangupCall{
		UserCallRequest: model.UserCallRequest{
			Id: in.GetId(),
		},
	}
	if in.GetAppId() != "" {
		req.AppId = model.NewString(in.GetAppId())
	}
	if in.GetCause() != "" {
		req.Cause = model.NewString(in.GetCause())
	}

	err = api.ctrl.HangupCall(session, session.Domain(in.DomainId), &req)
	if err != nil {
		return nil, err
	}
	return &engine.HangupCallResponse{}, nil
}

func (api *call) HoldCall(ctx context.Context, in *engine.UserCallRequest) (*engine.HoldCallResponse, error) {
	session, err := api.ctrl.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	req := model.UserCallRequest{
		Id: in.GetId(),
	}
	if in.GetAppId() != "" {
		req.AppId = model.NewString(in.GetAppId())
	}

	err = api.ctrl.HoldCall(session, session.Domain(in.DomainId), &req)
	if err != nil {
		return nil, err
	}
	return &engine.HoldCallResponse{
		State: "hold",
	}, nil
}

func (api *call) UnHoldCall(ctx context.Context, in *engine.UserCallRequest) (*engine.HoldCallResponse, error) {
	session, err := api.ctrl.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	req := model.UserCallRequest{
		Id: in.GetId(),
	}
	if in.GetAppId() != "" {
		req.AppId = model.NewString(in.GetAppId())
	}

	err = api.ctrl.UnHoldCall(session, session.Domain(in.DomainId), &req)
	if err != nil {
		return nil, err
	}
	return &engine.HoldCallResponse{
		State: "active",
	}, nil
}

func (api *call) DtmfCall(ctx context.Context, in *engine.DtmfCallRequest) (*engine.DtmfCallResponse, error) {
	session, err := api.ctrl.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	req := model.DtmfCall{
		UserCallRequest: model.UserCallRequest{
			Id: in.GetId(),
		},
	}
	if in.GetAppId() != "" {
		req.AppId = model.NewString(in.GetAppId())
	}

	if len(in.Digit) > 1 {
		req.Digit = rune(in.Digit[0])
	}

	err = api.ctrl.DtmfCall(session, session.Domain(in.DomainId), &req)
	if err != nil {
		return nil, err
	}
	return &engine.DtmfCallResponse{}, nil
}

func (api *call) BlindTransferCall(ctx context.Context, in *engine.BlindTransferCallRequest) (*engine.BlindTransferCallResponse, error) {
	session, err := api.ctrl.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	req := model.BlindTransferCall{
		UserCallRequest: model.UserCallRequest{
			Id: in.GetId(),
		},
		Destination: in.GetDestination(),
	}
	if in.GetAppId() != "" {
		req.AppId = model.NewString(in.GetAppId())
	}

	err = api.ctrl.BlindTransferCall(session, session.Domain(in.DomainId), &req)
	if err != nil {
		return nil, err
	}
	return &engine.BlindTransferCallResponse{}, nil
}

func (api *call) EavesdropCall(ctx context.Context, in *engine.EavesdropCallRequest) (*engine.CreateCallResponse, error) {
	session, err := api.ctrl.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	req := model.EavesdropCall{
		UserCallRequest: model.UserCallRequest{
			Id: in.GetId(),
		},
		Dtmf:        in.Control,
		ALeg:        in.ListenA,
		BLeg:        in.ListenB,
		WhisperALeg: in.WhisperA,
		WhisperBLeg: in.WhisperB,
	}
	if in.GetAppId() != "" {
		req.AppId = model.NewString(in.GetAppId())
	}

	if in.From != nil {
		req.From = &model.EndpointRequest{}

		if in.From.Id != 0 {
			req.From.UserId = model.NewInt64(in.From.Id)
		}

		if in.From.Extension != "" {
			req.From.Extension = model.NewString(in.From.Extension)
		}
	}

	_, err = api.ctrl.EavesdropCall(session, session.Domain(0), &req, nil)
	if err != nil {
		return nil, err
	}
	return &engine.CreateCallResponse{}, nil
}

func (api *call) CreateCallAnnotation(ctx context.Context, in *engine.CreateCallAnnotationRequest) (*engine.CallAnnotation, error) {
	session, err := api.ctrl.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	annotation := &model.CallAnnotation{
		CallId:   in.GetCallId(),
		Note:     in.GetNote(),
		StartSec: in.GetStartSec(),
		EndSec:   in.GetEndSec(),
	}

	annotation, err = api.ctrl.CreateCallAnnotation(session, annotation)
	if err != nil {
		return nil, err
	}

	return toEngineAnnotation(annotation), nil
}

func (api *call) UpdateCallAnnotation(ctx context.Context, in *engine.UpdateCallAnnotationRequest) (*engine.CallAnnotation, error) {
	session, err := api.ctrl.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	annotation := &model.CallAnnotation{
		Id:       in.GetId(),
		CallId:   in.GetCallId(),
		Note:     in.GetNote(),
		StartSec: in.GetStartSec(),
		EndSec:   in.GetEndSec(),
	}

	annotation, err = api.ctrl.UpdateCallAnnotation(session, annotation)
	if err != nil {
		return nil, err
	}

	return toEngineAnnotation(annotation), nil
}

func (api *call) DeleteCallAnnotation(ctx context.Context, in *engine.DeleteCallAnnotationRequest) (*engine.CallAnnotation, error) {
	session, err := api.ctrl.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	var annotation *model.CallAnnotation
	annotation, err = api.ctrl.DeleteCallAnnotation(session, in.GetId(), in.GetCallId())
	if err != nil {
		return nil, err
	}

	return toEngineAnnotation(annotation), nil
}

func (api *call) ConfirmPush(ctx context.Context, in *engine.ConfirmPushRequest) (*engine.ConfirmPushResponse, error) {
	session, err := api.ctrl.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	err = api.ctrl.ConfirmPushCall(session, in.Id)
	if err != nil {
		return nil, err
	}

	return &engine.ConfirmPushResponse{}, nil
}

func (api *call) SetVariablesCall(ctx context.Context, in *engine.SetVariablesCallRequest) (*engine.SetVariablesCallResponse, error) {
	session, err := api.ctrl.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	err = api.ctrl.SetCallVariables(session, in.Id, in.Variables)
	if err != nil {
		return nil, err
	}

	return &engine.SetVariablesCallResponse{}, nil
}

func toEngineCall(src *model.Call) *engine.ActiveCall {
	item := &engine.ActiveCall{
		Id:               src.Id,
		AppId:            src.AppId,
		Type:             src.Type,
		State:            src.State,
		Timestamp:        model.TimeToInt64(src.Timestamp),
		User:             GetProtoLookup(src.User),
		Extension:        "",
		Gateway:          GetProtoLookup(src.Gateway),
		Direction:        src.Direction,
		Destination:      src.Destination,
		Variables:        src.Variables,
		CreatedAt:        model.TimeToInt64(&src.CreatedAt),
		AnsweredAt:       model.TimeToInt64(src.AnsweredAt),
		BridgedAt:        model.TimeToInt64(src.BridgedAt),
		Duration:         int32(src.Duration),
		HoldSec:          int32(src.HoldSec),
		WaitSec:          int32(src.WaitSec),
		BillSec:          int32(src.BillSec),
		Queue:            GetProtoLookup(src.Queue),
		Member:           GetProtoLookup(src.Member),
		Team:             GetProtoLookup(src.Team),
		Agent:            GetProtoLookup(src.Agent),
		JoinedAt:         model.TimeToInt64(src.JoinedAt),
		LeavingAt:        model.TimeToInt64(src.LeavingAt),
		ReportingAt:      model.TimeToInt64(src.ReportingAt),
		QueueBridgedAt:   model.TimeToInt64(src.QueueBridgedAt),
		QueueWaitSec:     defaultInt(src.QueueWaitSec),
		QueueDurationSec: defaultInt(src.QueueDurationSec),
		ReportingSec:     defaultInt(src.ReportingSec),
		Display:          defaultString(src.Display),
		Supervisor:       GetProtoLookups(src.Supervisor),
	}

	if src.ParentId != nil {
		item.ParentId = *src.ParentId
	}

	if src.From != nil {
		item.From = &engine.Endpoint{
			Type:   src.From.Type,
			Number: src.From.Number,
			Id:     src.From.Id,
			Name:   src.From.Name,
		}
	}

	if src.To != nil {
		item.To = &engine.Endpoint{
			Type:   src.To.Type,
			Number: src.To.Number,
			Id:     src.To.Id,
			Name:   src.To.Name,
		}
	}

	if src.Extension != nil {
		item.Extension = *src.Extension
	}

	if src.BlindTransfer != nil {
		item.BlindTransfer = *src.BlindTransfer
	}

	return item
}

func toEngineAnnotation(src *model.CallAnnotation) *engine.CallAnnotation {
	return &engine.CallAnnotation{
		Id:        src.Id,
		CallId:    src.CallId,
		CreatedBy: GetProtoLookup(src.CreatedBy),
		CreatedAt: model.TimeToInt64(&src.CreatedAt),
		UpdatedBy: GetProtoLookup(src.UpdatedBy),
		UpdatedAt: model.TimeToInt64(&src.UpdatedAt),
		Note:      src.Note,
		StartSec:  src.StartSec,
		EndSec:    src.EndSec,
	}
}

func toEngineHistoryCall(src *model.HistoryCall, minHideString, pref, suff int, accessString bool, accessFile bool) *engine.HistoryCall {
	item := &engine.HistoryCall{
		Id:               src.Id,
		AppId:            src.AppId,
		Type:             src.Type,
		User:             GetProtoLookup(src.User),
		Extension:        "",
		Gateway:          GetProtoLookup(src.Gateway),
		Direction:        src.Direction,
		Destination:      setAccessString(src.Destination, minHideString, pref, suff, accessString),
		Variables:        prettyVariables(src.Variables),
		CreatedAt:        model.TimeToInt64(&src.CreatedAt),
		AnsweredAt:       model.TimeToInt64(src.AnsweredAt),
		BridgedAt:        model.TimeToInt64(src.BridgedAt),
		HangupAt:         model.TimeToInt64(src.HangupAt),
		StoredAt:         model.TimeToInt64(src.StoredAt),
		HangupBy:         src.HangupBy,
		Cause:            src.Cause,
		Duration:         int32(src.Duration),
		HoldSec:          int32(src.HoldSec),
		WaitSec:          int32(src.WaitSec),
		BillSec:          int32(src.BillSec),
		SipCode:          defaultInt(src.SipCode),
		Annotations:      toCallAnnotation(src.Annotations),
		Queue:            GetProtoLookup(src.Queue),
		Member:           GetProtoLookup(src.Member),
		Team:             GetProtoLookup(src.Team),
		Agent:            GetProtoLookup(src.Agent),
		JoinedAt:         model.TimeToInt64(src.JoinedAt),
		LeavingAt:        model.TimeToInt64(src.LeavingAt),
		ReportingAt:      model.TimeToInt64(src.ReportingAt),
		QueueBridgedAt:   model.TimeToInt64(src.QueueBridgedAt),
		QueueWaitSec:     defaultInt(src.QueueWaitSec),
		QueueDurationSec: defaultInt(src.QueueDurationSec),
		ReportingSec:     defaultInt(src.ReportingSec),
		Result:           src.GetResult(),
		Tags:             src.Tags, // TODO
		Display:          defaultString(src.Display),
		HasChildren:      src.HasChildren,
		Hold:             toCallHold(src.Hold),
		AmdResult:        defaultString(src.AmdResult),
		Transcripts:      toCallFileTranscriptLookups(src.Transcripts),
		TalkSec:          src.TalkSec,
		Grantee:          GetProtoLookup(src.Grantee),
		AmdAiLogs:        src.AmdAiLogs,
	}
	if src.ParentId != nil {
		item.ParentId = *src.ParentId
	}

	if src.TransferFrom != nil {
		item.TransferFrom = *src.TransferFrom
	}

	if src.TransferTo != nil {
		item.TransferTo = *src.TransferTo
	}

	if src.Result != nil {
		item.Result = *src.Result
	}

	if src.From != nil {
		item.From = &engine.Endpoint{
			Type:   src.From.Type,
			Number: setAccessString(src.From.Number, minHideString, pref, suff, accessString),
			Id:     src.From.Id,
			Name:   src.From.Name,
		}
	}

	if src.To != nil {
		item.To = &engine.Endpoint{
			Type:   src.To.Type,
			Number: setAccessString(src.To.Number, minHideString, pref, suff, accessString),
			Id:     src.To.Id,
			Name:   src.To.Name,
		}
	}

	if src.SipCode != nil {
		item.SipCode = int32(*src.SipCode)
	}

	if src.Extension != nil {
		item.Extension = *src.Extension
	}

	if src.AgentDescription != nil {
		item.AgentDescription = *src.AgentDescription
	}

	if src.HangupDisposition != nil {
		item.HangupDisposition = *src.HangupDisposition
	}

	if src.BlindTransfer != nil {
		item.BlindTransfer = *src.BlindTransfer
	}

	if src.AmdAiResult != nil {
		item.AmdAiResult = *src.AmdAiResult
	}

	if accessFile {
		item.Files = toCallFile(src.Files)
		item.FilesJob = toCallFilesJob(src.FilesJob)
	}

	return item
}

// todo, change proto response
func prettyVariables(src *model.Variables) map[string]string {
	if src == nil {
		return nil
	}
	if len(*src) > 0 {
		res := make(map[string]string)
		for k, v := range *src {
			switch r := v.(type) {
			case string:
				res[k] = r
			case []interface{}:
				t := make([]string, 0, len(r))
				for _, l := range r {
					t = append(t, fmt.Sprintf("%v", l))
				}
				res[k] = strings.Join(t, ", ")
			case []string:
				res[k] = strings.Join(r, ", ")
			default:
				res[k] = fmt.Sprintf("%v", v)

			}
		}
		return res
	}

	return nil
}

func toCallFileTranscriptLookups(src []*model.CallFileTranscriptLookup) []*engine.TranscriptLookup {
	if src == nil {
		return nil
	}

	res := make([]*engine.TranscriptLookup, 0, len(src))

	for _, v := range src {
		res = append(res, &engine.TranscriptLookup{
			Id:     v.Id,
			Locale: v.Locale,
			FileId: v.FileId,
			File:   GetProtoLookup(v.File),
		})
	}

	return res
}

func toCallFilesJob(src []*model.HistoryFileJob) []*engine.HistoryFileJob {
	if src == nil {
		return nil
	}

	res := make([]*engine.HistoryFileJob, 0, len(src))
	for _, v := range src {
		r := &engine.HistoryFileJob{
			Id:        v.Id,
			FileId:    v.FileId,
			CreatedAt: v.CreatedAt,
			Action:    toFileJobAction(v.Action),
			State:     (engine.HistoryFileJob_HistoryFileJobState)(v.State),
		}
		if v.Error != nil {
			r.ErrorDetail = *v.Error
		}
		res = append(res, r)
	}

	return res
}

func toFileJobAction(n string) engine.HistoryFileJob_HistoryFileJobAction {
	switch n {
	case "STT":
		return engine.HistoryFileJob_STT
	case "remove":
		return engine.HistoryFileJob_delete
	default:
		return engine.HistoryFileJob_undefined
	}
}

func toCallFile(src []*model.CallFile) []*engine.CallFile {
	if src == nil {
		return nil
	}

	res := make([]*engine.CallFile, 0, len(src))
	for _, v := range src {
		res = append(res, &engine.CallFile{
			Id:       v.Id,
			Name:     v.Name,
			Size:     v.Size,
			MimeType: v.MimeType,
			StartAt:  v.StartAt,
			StopAt:   v.StopAt,
		})
	}

	return res
}

func toCallAnnotation(src []*model.CallAnnotation) []*engine.CallAnnotation {
	if src == nil {
		return nil
	}

	res := make([]*engine.CallAnnotation, 0, len(src))
	for _, v := range src {
		res = append(res, &engine.CallAnnotation{
			Id:        v.Id,
			CallId:    v.CallId,
			CreatedBy: GetProtoLookup(v.CreatedBy),
			CreatedAt: model.TimeToInt64(&v.CreatedAt),
			UpdatedBy: GetProtoLookup(v.UpdatedBy),
			UpdatedAt: model.TimeToInt64(&v.UpdatedAt),
			Note:      v.Note,
			StartSec:  v.StartSec,
			EndSec:    v.EndSec,
		})
	}

	return res
}

func toCallHold(src []*model.CallHold) []*engine.CallHold {
	if src == nil {
		return nil
	}

	res := make([]*engine.CallHold, 0, len(src))
	for _, v := range src {
		res = append(res, &engine.CallHold{
			Start: v.Start,
			Stop:  v.Finish,
			Sec:   v.Sec,
		})
	}

	return res
}

func defaultInt(i *int) int32 {
	if i != nil {

		return int32(*i)
	}
	return 0
}

func defaultBigInt(i *int64) int64 {
	if i != nil {

		return *i
	}
	return 0
}

func defaultString(s *string) string {
	if s != nil {
		return *s
	}

	return ""
}

func setAccessString(str string, min, p, s int, h bool) string {
	if !h {
		return str
	}

	return model.HideString(str, min, p, s)
}
