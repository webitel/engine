package grpc_api

import (
	"context"
	"github.com/webitel/engine/model"
	"github.com/webitel/protos/engine"
)

type region struct {
	*API
}

func NewRegionApi(api *API) *region {
	return &region{api}
}

func (api *region) CreateRegion(ctx context.Context, in *engine.CreateRegionRequest) (*engine.Region, error) {
	session, err := api.ctrl.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	region := &model.Region{
		Name: in.Name,
		Timezone: model.Lookup{
			Id: int(in.GetTimezone().GetId()),
		},
		Description: &in.Description,
	}

	region, err = api.ctrl.CreateRegion(session, region)
	if err != nil {
		return nil, err
	}

	return toEngineRegion(region), nil
}

func (api *region) SearchRegion(ctx context.Context, in *engine.SearchRegionRequest) (*engine.ListRegion, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	var list []*model.Region
	var endList bool
	req := &model.SearchRegion{
		ListRequest: model.ListRequest{
			Q:       in.GetQ(),
			Page:    int(in.GetPage()),
			PerPage: int(in.GetSize()),
		},
		Ids:         in.GetId(),
		TimezoneIds: in.GetTimezoneId(),
	}

	if in.Name != "" {
		req.Name = &in.Name
	}

	if in.Description != "" {
		req.Description = &in.Description
	}

	list, endList, err = api.ctrl.SearchRegion(session, req)

	if err != nil {
		return nil, err
	}

	items := make([]*engine.Region, 0, len(list))
	for _, v := range list {
		items = append(items, toEngineRegion(v))
	}
	return &engine.ListRegion{
		Next:  !endList,
		Items: items,
	}, nil
}

func (api *region) ReadRegion(ctx context.Context, in *engine.ReadRegionRequest) (*engine.Region, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	var region *model.Region
	region, err = api.ctrl.GetRegion(session, in.Id)

	if err != nil {
		return nil, err
	}

	return toEngineRegion(region), nil
}

func (api *region) PatchRegion(ctx context.Context, in *engine.PatchRegionRequest) (*engine.Region, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	var region *model.Region
	patch := &model.RegionPatch{}

	//TODO
	for _, v := range in.Fields {
		switch v {
		case "name":
			patch.Name = model.NewString(in.Name)
		case "description":
			patch.Description = model.NewString(in.Description)
		case "timezone.id":
			patch.Timezone = GetLookup(in.Timezone)
		}
	}

	if region, err = api.ctrl.PatchRegion(session, in.Id, patch); err != nil {
		return nil, err
	}

	return toEngineRegion(region), nil
}

func (api *region) UpdateRegion(ctx context.Context, in *engine.UpdateRegionRequest) (*engine.Region, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	region := &model.Region{
		Id:          in.Id,
		Name:        in.Name,
		Description: &in.Description,
	}

	if in.Timezone != nil {
		region.Timezone.Id = int(in.Timezone.Id)
	}

	region, err = api.ctrl.UpdateRegion(session, region)

	if err != nil {
		return nil, err
	}

	return toEngineRegion(region), nil
}

func (api *region) DeleteRegion(ctx context.Context, in *engine.DeleteRegionRequest) (*engine.Region, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	var region *model.Region
	region, err = api.ctrl.DeleteRegion(session, in.Id)
	if err != nil {
		return nil, err
	}

	return toEngineRegion(region), nil
}

func toEngineRegion(src *model.Region) *engine.Region {
	r := &engine.Region{
		Id:          src.Id,
		Name:        src.Name,
		Description: "",
		Timezone:    GetProtoLookup(&src.Timezone),
	}

	if src.Description != nil {
		r.Description = *src.Description
	}

	return r
}
