package app

import "github.com/webitel/engine/model"

func (a *App) CreateOutboundResource(resource *model.OutboundCallResource) (*model.OutboundCallResource, *model.AppError) {
	return a.Store.OutboundResource().Create(resource)
}

func (a *App) OutboundResourceCheckAccess(domainId, id int64, groups []int, access model.PermissionAccess) (bool, *model.AppError) {
	return a.Store.OutboundResource().CheckAccess(domainId, id, groups, access)
}

func (app *App) GetOutboundResource(domainId, id int64) (*model.OutboundCallResource, *model.AppError) {
	return app.Store.OutboundResource().Get(domainId, id)
}

func (a *App) GetOutboundResourcePage(domainId int64, page, perPage int) ([]*model.OutboundCallResource, *model.AppError) {
	return a.Store.OutboundResource().GetAllPage(domainId, page*perPage, perPage)
}

func (a *App) GetOutboundResourcePageByGroups(domainId int64, groups []int, page, perPage int) ([]*model.OutboundCallResource, *model.AppError) {
	return a.Store.OutboundResource().GetAllPageByGroups(domainId, groups, page*perPage, perPage)
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
	oldResource.DialString = resource.DialString
	oldResource.ErrorIds = resource.ErrorIds
	oldResource.Gateway = resource.Gateway

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
