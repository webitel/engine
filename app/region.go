package app

import (
	"context"

	"github.com/webitel/engine/model"
)

func (a *App) GetRegionsPage(ctx context.Context, domainId int64, search *model.SearchRegion) ([]*model.Region, bool, model.AppError) {
	list, err := a.Store.Region().GetAllPage(ctx, domainId, search)
	if err != nil {
		return nil, false, err
	}
	search.RemoveLastElemIfNeed(&list)
	return list, search.EndOfList(), nil
}

func (a *App) CreateRegion(ctx context.Context, domainId int64, region *model.Region) (*model.Region, model.AppError) {
	return a.Store.Region().Create(ctx, domainId, region)
}

func (a *App) GetRegion(ctx context.Context, domainId int64, id int64) (*model.Region, model.AppError) {
	return a.Store.Region().Get(ctx, domainId, id)
}

func (a *App) UpdateRegion(ctx context.Context, domainId int64, region *model.Region) (*model.Region, model.AppError) {
	oldRegion, err := a.GetRegion(ctx, domainId, region.Id)
	if err != nil {
		return nil, err
	}

	oldRegion.Name = region.Name
	oldRegion.Description = region.Description
	oldRegion.Timezone = region.Timezone

	oldRegion, err = a.Store.Region().Update(ctx, domainId, oldRegion)
	if err != nil {
		return nil, err
	}

	return oldRegion, nil
}

func (a *App) PatchRegion(ctx context.Context, domainId int64, id int64, patch *model.RegionPatch) (*model.Region, model.AppError) {
	oldRegion, err := a.GetRegion(ctx, domainId, id)
	if err != nil {
		return nil, err
	}

	oldRegion.Patch(patch)

	if err = oldRegion.IsValid(); err != nil {
		return nil, err
	}

	oldRegion, err = a.Store.Region().Update(ctx, domainId, oldRegion)
	if err != nil {
		return nil, err
	}

	return oldRegion, nil
}

func (a *App) RemoveRegion(ctx context.Context, domainId int64, id int64) (*model.Region, model.AppError) {
	region, err := a.Store.Region().Get(ctx, domainId, id)

	if err != nil {
		return nil, err
	}

	err = a.Store.Region().Delete(ctx, domainId, id)
	if err != nil {
		return nil, err
	}
	return region, nil
}
