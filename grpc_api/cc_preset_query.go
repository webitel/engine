package grpc_api

import (
	gogrpc "buf.build/gen/go/webitel/engine/grpc/go/_gogrpc"
	engine "buf.build/gen/go/webitel/engine/protocolbuffers/go"
	"context"
	"github.com/webitel/engine/model"
	"strings"
)

type presetQuery struct {
	*API
	gogrpc.UnsafePresetQueryServiceServer
}

func (api *presetQuery) CreatePresetQuery(ctx context.Context, in *engine.CreatePresetQueryRequest) (*engine.PresetQuery, error) {
	session, err := api.ctrl.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	preset := &model.PresetQuery{
		Name:        in.Name,
		Description: in.GetDescription(),
		Section:     in.Section.String(),
		Preset:      MarshalJsonpbToMap(in.Preset),
	}

	preset, err = api.ctrl.CreatePresetQuery(ctx, session, preset)
	if err != nil {
		return nil, err
	}

	return transformPresetQuery(preset), nil
}

func (api *presetQuery) SearchPresetQuery(ctx context.Context, in *engine.SearchPresetQueryRequest) (*engine.ListPresetQuery, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	var list []*model.PresetQuery
	var endList bool
	req := &model.SearchPresetQuery{
		ListRequest: model.ListRequest{
			Q:       in.GetQ(),
			Page:    int(in.GetPage()),
			PerPage: int(in.GetSize()),
			Fields:  in.Fields,
			Sort:    in.Sort,
		},
		Ids: in.Id,
	}

	for _, v := range in.Section {
		req.Section = append(req.Section, v.String())
	}

	list, endList, err = api.ctrl.SearchPresetQuery(ctx, session, req)

	if err != nil {
		return nil, err
	}

	items := make([]*engine.PresetQuery, 0, len(list))
	for _, v := range list {
		items = append(items, transformPresetQuery(v))
	}
	return &engine.ListPresetQuery{
		Next:  !endList,
		Items: items,
	}, nil
}

func (api *presetQuery) ReadPresetQuery(ctx context.Context, in *engine.ReadPresetQueryRequest) (*engine.PresetQuery, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	var preset *model.PresetQuery
	preset, err = api.ctrl.ReadPresetQuery(ctx, session, in.Id)

	if err != nil {
		return nil, err
	}

	return transformPresetQuery(preset), nil
}

func (api *presetQuery) UpdatePresetQuery(ctx context.Context, in *engine.UpdatePresetQueryRequest) (*engine.PresetQuery, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	preset := &model.PresetQuery{
		Id:          in.Id,
		Name:        in.Name,
		Description: in.GetDescription(),
		Section:     in.Section.String(),
		Preset:      MarshalJsonpbToMap(in.Preset),
	}

	preset, err = api.ctrl.UpdatePresetQuery(ctx, session, preset)

	if err != nil {
		return nil, err
	}

	return transformPresetQuery(preset), nil
}

func (api *presetQuery) PatchPresetQuery(ctx context.Context, in *engine.PatchPresetQueryRequest) (*engine.PresetQuery, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	var preset *model.PresetQuery
	patch := &model.PresetQueryPatch{}

	//TODO
	for _, v := range in.Fields {
		switch v {
		case "name":
			patch.Name = &in.Name
		case "description":
			patch.Description = &in.Description
		default:
			if strings.HasPrefix(v, "preset.") {
				patch.Preset = MarshalJsonpbToMap(in.Preset)
			}
		}
	}

	preset, err = api.ctrl.PatchPresetQuery(ctx, session, in.GetId(), patch)

	if err != nil {
		return nil, err
	}

	return transformPresetQuery(preset), nil
}

func (api *presetQuery) DeletePresetQuery(ctx context.Context, in *engine.DeletePresetQueryRequest) (*engine.PresetQuery, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	var preset *model.PresetQuery
	preset, err = api.ctrl.RemovePresetQuery(ctx, session, in.Id)
	if err != nil {
		return nil, err
	}

	return transformPresetQuery(preset), nil
}

func NewPresetQueryApi(api *API) *presetQuery {
	return &presetQuery{API: api}
}

func transformPresetQuery(src *model.PresetQuery) *engine.PresetQuery {
	s, _ := engine.PresetQuerySection_value[src.Section]
	return &engine.PresetQuery{
		Id:          src.Id,
		Name:        src.Name,
		Description: src.Description,
		Preset:      UnmarshalJsonpb(src.Preset.ToSafeBytes()),
		CreatedAt:   model.TimeToInt64(src.CreatedAt),
		UpdatedAt:   model.TimeToInt64(src.UpdatedAt),
		Section:     engine.PresetQuerySection(s),
	}
}
