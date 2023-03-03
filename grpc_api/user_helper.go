package grpc_api

import (
	"context"
	"github.com/webitel/engine/model"
	"github.com/webitel/protos/engine"
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
