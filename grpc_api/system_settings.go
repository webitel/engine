package grpc_api

import (
	gogrpc "buf.build/gen/go/webitel/engine/grpc/go/_gogrpc"
	engine "buf.build/gen/go/webitel/engine/protocolbuffers/go"
	"context"
	"github.com/webitel/engine/model"
)

type systemSettings struct {
	*API
	gogrpc.UnsafeSystemSettingServiceServer
}

func NewSystemSettingsApi(api *API) *systemSettings {
	return &systemSettings{API: api}
}

func (api systemSettings) CreateSystemSetting(ctx context.Context, in *engine.CreateSystemSettingRequest) (*engine.SystemSetting, error) {
	session, err := api.ctrl.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	s := &model.SystemSetting{
		Name:  in.GetName().String(),
		Value: MarshalJsonpb(in.Value),
	}

	s, err = api.ctrl.CreateSystemSetting(ctx, session, s)
	if err != nil {
		return nil, err
	}
	return transformSystemSetting(s), nil
}

func (api systemSettings) SearchSystemSetting(ctx context.Context, in *engine.SearchSystemSettingRequest) (*engine.ListSystemSetting, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	var list []*model.SystemSetting
	var endList bool
	var nameFilters []string
	for _, name := range in.GetName() {
		nameFilters = append(nameFilters, name.String())
	}
	req := &model.SearchSystemSetting{
		ListRequest: model.ListRequest{
			Q:       in.GetQ(),
			Page:    int(in.GetPage()),
			PerPage: int(in.GetSize()),
			Fields:  in.Fields,
			Sort:    in.Sort,
		},
		Name: nameFilters,
	}

	list, endList, err = api.ctrl.SearchSystemSetting(ctx, session, req)

	if err != nil {
		return nil, err
	}

	items := make([]*engine.SystemSetting, 0, len(list))
	for _, v := range list {
		items = append(items, transformSystemSetting(v))
	}
	return &engine.ListSystemSetting{
		Next:  !endList,
		Items: items,
	}, nil
}

func (api systemSettings) ReadSystemSetting(ctx context.Context, in *engine.ReadSystemSettingRequest) (*engine.SystemSetting, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	var s *model.SystemSetting
	s, err = api.ctrl.ReadSystemSetting(ctx, session, in.Id)

	if err != nil {
		return nil, err
	}

	return transformSystemSetting(s), nil
}

func (api systemSettings) UpdateSystemSetting(ctx context.Context, in *engine.UpdateSystemSettingRequest) (*engine.SystemSetting, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	s := &model.SystemSetting{
		Id:    in.Id,
		Value: MarshalJsonpb(in.Value),
	}

	s, err = api.ctrl.UpdateSystemSetting(ctx, session, s)

	if err != nil {
		return nil, err
	}

	return transformSystemSetting(s), nil
}

func (api systemSettings) PatchSystemSetting(ctx context.Context, in *engine.PatchSystemSettingRequest) (*engine.SystemSetting, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	var s *model.SystemSetting
	patch := &model.SystemSettingPath{}

	//TODO
	for _, v := range in.Fields {
		switch v {
		case "value":
			patch.Value = MarshalJsonpb(in.Value)
		}
	}

	if s, err = api.ctrl.PatchSystemSetting(ctx, session, in.Id, patch); err != nil {
		return nil, err
	}

	return transformSystemSetting(s), nil
}

func (api systemSettings) DeleteSystemSetting(ctx context.Context, in *engine.DeleteSystemSettingRequest) (*engine.SystemSetting, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	var s *model.SystemSetting
	s, err = api.ctrl.DeleteSystemSetting(ctx, session, in.Id)
	if err != nil {
		return nil, err
	}

	return transformSystemSetting(s), nil
}

func (api systemSettings) SearchAvailableSystemSetting(ctx context.Context, in *engine.SearchAvailableSystemSettingRequest) (*engine.ListAvailableSystemSetting, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	var list []string

	list, err = api.ctrl.SearchAvailableSystemSetting(ctx, session, &model.ListRequest{
		Q: in.Q,
	})

	if err != nil {
		return nil, err
	}

	res := &engine.ListAvailableSystemSetting{
		Items: make([]*engine.AvailableSystemSetting, 0, len(list)),
	}

	for _, v := range list {
		res.Items = append(res.Items, &engine.AvailableSystemSetting{
			Name: v,
		})
	}

	return res, nil
}

func transformSystemSetting(s *model.SystemSetting) *engine.SystemSetting {
	res := &engine.SystemSetting{
		Id:    s.Id,
		Value: UnmarshalJsonpb(s.Value),
	}

	i, _ := engine.SystemSettingName_value[s.Name]
	res.Name = engine.SystemSettingName(i)

	return res
}
