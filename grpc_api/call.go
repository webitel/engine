package grpc_api

import (
	"context"
	"github.com/webitel/engine/grpc_api/engine"
	"github.com/webitel/engine/model"
)

type call struct {
	*API
}

func NewCallApi(app *API) *call {
	return &call{app}
}

func (api *call) ReadCall(ctx context.Context, in *engine.ReadCallRequest) (*engine.Call, error) {
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

func (api *call) SearchCall(ctx context.Context, in *engine.SearchCallRequest) (*engine.ListCall, error) {
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
		},
	}

	list, endList, err = api.ctrl.SearchCall(session, req)

	if err != nil {
		return nil, err
	}

	items := make([]*engine.Call, 0, len(list))
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
		To: model.EndpointRequest{},
		Params: model.CallParameters{
			Timeout:   int(in.GetParams().GetTimeout()),
			Audio:     in.GetParams().GetAudio(),
			Video:     in.GetParams().GetVideo(),
			Screen:    in.GetParams().GetScreen(),
			Record:    in.GetParams().GetRecord(),
			Variables: in.GetParams().GetVariables(),
		},
	}

	if in.To != nil {
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
	}

	var id string
	id, err = api.ctrl.CreateCall(session, req)
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
	return nil, nil
}

func toEngineCall(src *model.Call) *engine.Call {
	item := &engine.Call{
		Id:        src.Id,
		CreatedAt: src.CreatedAt,
		CreatedBy: GetProtoLookup(src.User),
		Timestamp: src.Timestamp,
		State:     src.State,
		Direction: src.Direction,
		From: &engine.Endpoint{
			Type:   src.From.Type,
			Id:     int64(src.From.Id),
			Name:   src.From.Name,
			Number: src.From.Number,
		},
		To: &engine.Endpoint{
			Type:   src.To.Type,
			Id:     int64(src.To.Id),
			Name:   src.To.Name,
			Number: src.To.Number,
		},
	}

	if src.AppId != nil {
		item.AppId = *src.AppId
	}

	if src.ParentId != nil {
		item.ParentId = *src.ParentId
	}

	return item
}
