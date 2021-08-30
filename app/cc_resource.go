package app

import (
	"github.com/webitel/engine/auth_manager"
	"github.com/webitel/engine/model"
)

func (a *App) CreateOutboundResource(resource *model.OutboundCallResource) (*model.OutboundCallResource, *model.AppError) {
	return a.Store.OutboundResource().Create(resource)
}

func (a *App) OutboundResourceCheckAccess(domainId, id int64, groups []int, access auth_manager.PermissionAccess) (bool, *model.AppError) {
	return a.Store.OutboundResource().CheckAccess(domainId, id, groups, access)
}

func (app *App) GetOutboundResource(domainId, id int64) (*model.OutboundCallResource, *model.AppError) {
	return app.Store.OutboundResource().Get(domainId, id)
}

func (a *App) GetOutboundResourcePage(domainId int64, search *model.SearchOutboundCallResource) ([]*model.OutboundCallResource, bool, *model.AppError) {
	list, err := a.Store.OutboundResource().GetAllPage(domainId, search)
	if err != nil {
		return nil, false, err
	}
	search.RemoveLastElemIfNeed(&list)
	return list, search.EndOfList(), nil
}

func (a *App) GetOutboundResourcePageByGroups(domainId int64, groups []int, search *model.SearchOutboundCallResource) ([]*model.OutboundCallResource, bool, *model.AppError) {
	list, err := a.Store.OutboundResource().GetAllPageByGroups(domainId, groups, search)
	if err != nil {
		return nil, false, err
	}
	search.RemoveLastElemIfNeed(&list)
	return list, search.EndOfList(), nil
}

func (a *App) PatchOutboundResource(domainId, id int64, patch *model.OutboundCallResourcePath) (*model.OutboundCallResource, *model.AppError) {
	oldResource, err := a.GetOutboundResource(domainId, id)
	if err != nil {
		return nil, err
	}

	oldResource.Path(patch)

	if err = oldResource.IsValid(); err != nil {
		return nil, err
	}

	oldResource, err = a.Store.OutboundResource().Update(oldResource)
	if err != nil {
		return nil, err
	}

	return oldResource, nil
}

func (a *App) UpdateOutboundResource(resource *model.OutboundCallResource) (*model.OutboundCallResource, *model.AppError) {
	oldResource, err := a.GetOutboundResource(resource.DomainId, resource.Id)
	if err != nil {
		return nil, err
	}

	oldResource.Limit = resource.Limit
	oldResource.Enabled = resource.Enabled
	oldResource.UpdatedAt = resource.UpdatedAt
	oldResource.UpdatedBy.Id = resource.UpdatedBy.Id
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

	oldResource, err = a.Store.OutboundResource().Update(oldResource)
	if err != nil {
		return nil, err
	}

	return oldResource, nil
}

func (a *App) RemoveOutboundResource(domainId, id int64) (*model.OutboundCallResource, *model.AppError) {
	resource, err := a.Store.OutboundResource().Get(domainId, id)

	if err != nil {
		return nil, err
	}

	err = a.Store.OutboundResource().Delete(domainId, id)
	if err != nil {
		return nil, err
	}
	return resource, nil
}

func (a *App) CreateOutboundResourceDisplay(display *model.ResourceDisplay) (*model.ResourceDisplay, *model.AppError) {
	return a.Store.OutboundResource().SaveDisplay(display)
}

func (a *App) GetOutboundResourceDisplayPage(domainId, resourceId int64, search *model.SearchResourceDisplay) ([]*model.ResourceDisplay, bool, *model.AppError) {
	list, err := a.Store.OutboundResource().GetDisplayAllPage(domainId, resourceId, search)
	if err != nil {
		return nil, false, err
	}
	search.RemoveLastElemIfNeed(&list)
	return list, search.EndOfList(), nil
}

func (a *App) GetOutboundResourceDisplay(domainId, resourceId, id int64) (*model.ResourceDisplay, *model.AppError) {
	return a.Store.OutboundResource().GetDisplay(domainId, resourceId, id)
}

func (a *App) UpdateOutboundResourceDisplay(domainId int64, display *model.ResourceDisplay) (*model.ResourceDisplay, *model.AppError) {
	oldDisplay, err := a.GetOutboundResourceDisplay(domainId, display.ResourceId, display.Id)
	if err != nil {
		return nil, err
	}

	oldDisplay.Display = display.Display

	oldDisplay, err = a.Store.OutboundResource().UpdateDisplay(domainId, oldDisplay)
	if err != nil {
		return nil, err
	}

	return oldDisplay, nil
}

func (a *App) RemoveOutboundResourceDisplay(domainId, resourceId, id int64) (*model.ResourceDisplay, *model.AppError) {
	display, err := a.Store.OutboundResource().GetDisplay(domainId, resourceId, id)

	if err != nil {
		return nil, err
	}

	err = a.Store.OutboundResource().DeleteDisplay(domainId, resourceId, id)
	if err != nil {
		return nil, err
	}
	return display, nil
}
