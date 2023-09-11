package grpc_api

import (
	"context"
	"github.com/webitel/protos/engine"
)

type systemSettings struct {
	*API
	engine.UnsafeSystemSettingServiceServer
}

func NewSystemSettingsApi(api *API) *systemSettings {
	return &systemSettings{API: api}
}

func (s systemSettings) CreateSystemSetting(ctx context.Context, in *engine.CreateSystemSettingRequest) (*engine.SystemSetting, error) {
	//TODO implement me
	panic("implement me")
}

func (s systemSettings) SearchSystemSetting(ctx context.Context, in *engine.SearchSystemSettingRequest) (*engine.ListSystemSetting, error) {
	//TODO implement me
	panic("implement me")
}

func (s systemSettings) ReadSystemSetting(ctx context.Context, in *engine.ReadSystemSettingRequest) (*engine.SystemSetting, error) {
	//TODO implement me
	panic("implement me")
}

func (s systemSettings) UpdateSystemSetting(ctx context.Context, in *engine.UpdateSystemSettingRequest) (*engine.SystemSetting, error) {
	//TODO implement me
	panic("implement me")
}

func (s systemSettings) PatchSystemSetting(ctx context.Context, in *engine.PatchSystemSettingRequest) (*engine.SystemSetting, error) {
	//TODO implement me
	panic("implement me")
}

func (s systemSettings) DeleteSystemSetting(ctx context.Context, in *engine.DeleteSystemSettingRequest) (*engine.SystemSetting, error) {
	//TODO implement me
	panic("implement me")
}
