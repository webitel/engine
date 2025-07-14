package grpc_api

import (
	"context"
	"github.com/webitel/engine/gen/engine"
	"github.com/webitel/engine/model"
)

type userHelper struct {
	*API
	engine.UnsafeUserHelperServiceServer
}

func NewUserHelperApi(api *API) *userHelper {
	return &userHelper{API: api}
}

func (api *userHelper) DefaultDeviceConfig(ctx context.Context, in *engine.DefaultDeviceConfigRequest) (*engine.DefaultDeviceConfigResponse, error) {
	session, err := api.ctrl.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	switch in.Type {
	case model.DeviceTypeSip:
		if res, err := api.app.GetUserDefaultSipCDeviceConfig(ctx, session.UserId, session.DomainId); err != nil {
			return nil, err
		} else {
			return &engine.DefaultDeviceConfigResponse{
				Data: &engine.DefaultDeviceConfigResponse_Sip{
					Sip: &engine.DefaultDeviceConfigResponse_SipDeviceConfig{
						Auth:      res.Auth,
						Domain:    res.Domain,
						Extension: res.Extension,
						Password:  res.Password,
						Proxy:     res.Proxy,
					},
				},
			}, nil
		}

	case model.DeviceTypeWebRTC:
		if res, err := api.app.GetUserDefaultWebRTCDeviceConfig(ctx, session.UserId, session.DomainId); err != nil {
			return nil, err
		} else {
			return &engine.DefaultDeviceConfigResponse{
				Data: &engine.DefaultDeviceConfigResponse_Webrtc{
					Webrtc: &engine.DefaultDeviceConfigResponse_WebRTCDeviceConfig{
						AuthorizationUser: res.AuthorizationUser,
						DisplayName:       res.DisplayName,
						Extension:         res.Extension,
						Ha1:               res.Ha1,
						Realm:             res.Realm,
						Server:            res.Server,
						Uri:               res.Uri,
					},
				},
			}, nil
		}

	default:
		//todo error
		return &engine.DefaultDeviceConfigResponse{}, nil
	}
}

func (api *userHelper) ActivityWorkspaceWidget(ctx context.Context, in *engine.ActivityWorkspaceWidgetRequest) (*engine.ActivityWorkspaceWidgetResponse, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	var stat *model.AgentStatistics
	stat, err = api.ctrl.GetUserTodayStatistics(ctx, session)
	if err != nil {
		return nil, err
	}

	return &engine.ActivityWorkspaceWidgetResponse{
		Utilization:      stat.Utilization,
		Occupancy:        stat.Occupancy,
		CallAbandoned:    stat.CallAbandoned,
		CallHandled:      stat.CallHandled,
		AvgTalkSec:       stat.AvgTalkSec,
		AvgHoldSec:       stat.AvgHoldSec,
		ChatAccepts:      stat.ChatAccepts,
		ChatAht:          stat.ChatAht,
		CallMissed:       stat.CallMissed,
		CallInbound:      stat.CallInbound,
		ScoreRequiredAvg: stat.ScoreRequiredAvg,
		ScoreOptionalAvg: stat.ScoreOptionalAvg,
		ScoreCount:       stat.ScoreCount,
		ScoreRequiredSum: stat.ScoreRequiredSum,
		ScoreOptionalSum: stat.ScoreOptionalSum,
		SumTalkSec:       stat.SumTalkSec,
		VoiceMail:        stat.VoiceMail,
		Available:        stat.Available,
		Online:           stat.Online,
		Processing:       stat.Processing,
		TaskAccepts:      stat.TaskAccepts,
		QueueTalkSec:     stat.QueueTalkSec,
		CallQueueMissed:  stat.CallQueueMissed,
		CallInboundQueue: stat.CallInboundQueue,
		CallDialerQueue:  stat.CallDialerQueue,
		CallManual:       stat.CallManual,
	}, nil
}

func (api *userHelper) OpenedWebSockets(ctx context.Context, in *engine.OpenedWebSocketsRequest) (*engine.ListOpenedWebSocket, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	if len(in.UserId) == 0 {
		return &engine.ListOpenedWebSocket{}, nil
	}

	var list []*model.SocketSessionView
	var endList bool
	list, endList, err = api.ctrl.GetWebSocketsPage(ctx, session, in.UserId[0], &model.ListRequest{
		Q:       in.GetQ(),
		Page:    int(in.GetPage()),
		PerPage: int(in.GetSize()),
		Fields:  in.GetFields(),
		Sort:    in.GetSort(),
	})

	if err != nil {
		return nil, err
	}

	items := make([]*engine.OpenedWebSocket, 0, len(list))
	for _, v := range list {
		items = append(items, &engine.OpenedWebSocket{
			Id:        v.Id,
			CreatedAt: model.TimeToInt64(v.CreatedAt),
			UpdatedAt: model.TimeToInt64(v.UpdatedAt),
			UserAgent: v.UserAgent,
			Ip:        v.Ip,
			Client:    v.Client,
			Duration:  v.Duration,
			Pong:      v.Pong,
		})
	}
	return &engine.ListOpenedWebSocket{
		Next:  !endList,
		Items: items,
	}, nil

}
