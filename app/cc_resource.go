package app

import (
	"context"

	"github.com/webitel/engine/auth_manager"
	"github.com/webitel/engine/model"
)

func (a *App) CreateOutboundResource(ctx context.Context, resource *model.OutboundCallResource) (*model.OutboundCallResource, model.AppError) {
	return a.Store.OutboundResource().Create(ctx, resource)
}

func (a *App) OutboundResourceCheckAccess(ctx context.Context, domainId, id int64, groups []int, access auth_manager.PermissionAccess) (bool, model.AppError) {
	return a.Store.OutboundResource().CheckAccess(ctx, domainId, id, groups, access)
}

func (app *App) GetOutboundResource(ctx context.Context, domainId, id int64) (*model.OutboundCallResource, model.AppError) {
	return app.Store.OutboundResource().Get(ctx, domainId, id)
}

func (a *App) GetOutboundResourcePage(ctx context.Context, domainId int64, search *model.SearchOutboundCallResource) ([]*model.OutboundCallResource, bool, model.AppError) {
	list, err := a.Store.OutboundResource().GetAllPage(ctx, domainId, search)
	if err != nil {
		return nil, false, err
	}
	search.RemoveLastElemIfNeed(&list)
	return list, search.EndOfList(), nil
}

func (a *App) GetOutboundResourcePageByGroups(ctx context.Context, domainId int64, groups []int, search *model.SearchOutboundCallResource) ([]*model.OutboundCallResource, bool, model.AppError) {
	list, err := a.Store.OutboundResource().GetAllPageByGroups(ctx, domainId, groups, search)
	if err != nil {
		return nil, false, err
	}
	search.RemoveLastElemIfNeed(&list)
	return list, search.EndOfList(), nil
}

func (a *App) PatchOutboundResource(ctx context.Context, domainId, id int64, patch *model.OutboundCallResourcePath) (*model.OutboundCallResource, model.AppError) {
	oldResource, err := a.GetOutboundResource(ctx, domainId, id)
	if err != nil {
		return nil, err
	}

	oldResource.Path(patch)

	if err = oldResource.IsValid(); err != nil {
		return nil, err
	}

	oldResource, err = a.Store.OutboundResource().Update(ctx, oldResource)
	if err != nil {
		return nil, err
	}

	return oldResource, nil
}

func (a *App) UpdateOutboundResource(ctx context.Context, resource *model.OutboundCallResource) (*model.OutboundCallResource, model.AppError) {
	oldResource, err := a.GetOutboundResource(ctx, resource.DomainId, resource.Id)
	if err != nil {
		return nil, err
	}

	oldResource.Limit = resource.Limit
	oldResource.Enabled = resource.Enabled
	oldResource.UpdatedAt = resource.UpdatedAt
	oldResource.UpdatedBy = resource.UpdatedBy
	oldResource.RPS = resource.RPS
	oldResource.Reserve = resource.Reserve
	oldResource.Variables = resource.Variables
	oldResource.Number = resource.Number
	oldResource.MaxSuccessivelyErrors = resource.MaxSuccessivelyErrors
	oldResource.Name = resource.Name
	oldResource.ErrorIds = resource.ErrorIds
	oldResource.Gateway = resource.Gateway
	oldResource.Description = resource.Description
	oldResource.Patterns = resource.Patterns
	oldResource.FailureDialDelay = resource.FailureDialDelay
	oldResource.Parameters = resource.Parameters

	oldResource, err = a.Store.OutboundResource().Update(ctx, oldResource)
	if err != nil {
		return nil, err
	}

	return oldResource, nil
}

func (a *App) RemoveOutboundResource(ctx context.Context, domainId, id int64) (*model.OutboundCallResource, model.AppError) {
	resource, err := a.Store.OutboundResource().Get(ctx, domainId, id)

	if err != nil {
		return nil, err
	}

	err = a.Store.OutboundResource().Delete(ctx, domainId, id)
	if err != nil {
		return nil, err
	}
	return resource, nil
}

func (a *App) CreateOutboundResourceDisplay(ctx context.Context, display *model.ResourceDisplay) (*model.ResourceDisplay, model.AppError) {
	return a.Store.OutboundResource().SaveDisplay(ctx, display)
}

func (a *App) CreateOutboundResourceDisplays(ctx context.Context, resourceId int64, displays []*model.ResourceDisplay) ([]*model.ResourceDisplay, model.AppError) {
	return a.Store.OutboundResource().SaveDisplays(ctx, resourceId, displays)
}

func (a *App) GetOutboundResourceDisplayPage(ctx context.Context, domainId, resourceId int64, search *model.SearchResourceDisplay) ([]*model.ResourceDisplay, bool, model.AppError) {
	list, err := a.Store.OutboundResource().GetDisplayAllPage(ctx, domainId, resourceId, search)
	if err != nil {
		return nil, false, err
	}
	search.RemoveLastElemIfNeed(&list)
	return list, search.EndOfList(), nil
}

func (a *App) GetOutboundResourceDisplay(ctx context.Context, domainId, resourceId, id int64) (*model.ResourceDisplay, model.AppError) {
	return a.Store.OutboundResource().GetDisplay(ctx, domainId, resourceId, id)
}

func (a *App) UpdateOutboundResourceDisplay(ctx context.Context, domainId int64, display *model.ResourceDisplay) (*model.ResourceDisplay, model.AppError) {
	oldDisplay, err := a.GetOutboundResourceDisplay(ctx, domainId, display.ResourceId, display.Id)
	if err != nil {
		return nil, err
	}

	oldDisplay.Display = display.Display

	oldDisplay, err = a.Store.OutboundResource().UpdateDisplay(ctx, domainId, oldDisplay)
	if err != nil {
		return nil, err
	}

	return oldDisplay, nil
}

func (a *App) RemoveOutboundResourceDisplay(ctx context.Context, domainId, resourceId, id int64) (*model.ResourceDisplay, model.AppError) {
	display, err := a.Store.OutboundResource().GetDisplay(ctx, domainId, resourceId, id)

	if err != nil {
		return nil, err
	}

	err = a.Store.OutboundResource().DeleteDisplay(ctx, domainId, resourceId, id)
	if err != nil {
		return nil, err
	}
	return display, nil
}

func (a *App) RemoveOutboundResourceDisplays(ctx context.Context, resourceId int64, ids []int64) model.AppError {
	return a.Store.OutboundResource().DeleteDisplays(ctx, resourceId, ids)
}
