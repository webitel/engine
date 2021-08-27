package app

import (
	"github.com/webitel/engine/auth_manager"
	"github.com/webitel/engine/model"
)

func (a *App) OutboundResourceGroupCheckAccess(domainId, id int64, groups []int, access auth_manager.PermissionAccess) (bool, *model.AppError) {
	return a.Store.OutboundResourceGroup().CheckAccess(domainId, id, groups, access)
}

func (a *App) CreateOutboundResourceGroup(group *model.OutboundResourceGroup) (*model.OutboundResourceGroup, *model.AppError) {
	return a.Store.OutboundResourceGroup().Create(group)
}

func (a *App) GetOutboundResourceGroupPage(domainId int64, search *model.SearchOutboundResourceGroup) ([]*model.OutboundResourceGroup, bool, *model.AppError) {
	list, err := a.Store.OutboundResourceGroup().GetAllPage(domainId, search)
	if err != nil {
		return nil, false, err
	}
	search.RemoveLastElemIfNeed(&list)
	return list, search.EndOfList(), nil
}

func (a *App) GetOutboundResourceGroupPageByGroups(domainId int64, groups []int, search *model.SearchOutboundResourceGroup) ([]*model.OutboundResourceGroup, bool, *model.AppError) {
	list, err := a.Store.OutboundResourceGroup().GetAllPageByGroups(domainId, groups, search)
	if err != nil {
		return nil, false, err
	}
	search.RemoveLastElemIfNeed(&list)
	return list, search.EndOfList(), nil
}

func (app *App) GetOutboundResourceGroup(domainId, id int64) (*model.OutboundResourceGroup, *model.AppError) {
	return app.Store.OutboundResourceGroup().Get(domainId, id)
}

func (a *App) UpdateOutboundResourceGroup(group *model.OutboundResourceGroup) (*model.OutboundResourceGroup, *model.AppError) {
	oldGroup, err := a.GetOutboundResourceGroup(group.DomainId, group.Id)
	if err != nil {
		return nil, err
	}

	oldGroup.Name = group.Name
	oldGroup.Strategy = group.Strategy
	oldGroup.Description = group.Description
	oldGroup.Communication = group.Communication
	oldGroup.UpdatedBy = group.UpdatedBy
	oldGroup.UpdatedAt = group.UpdatedAt
	oldGroup.Time = group.Time

	oldGroup, err = a.Store.OutboundResourceGroup().Update(oldGroup)
	if err != nil {
		return nil, err
	}

	return oldGroup, nil
}

func (a *App) RemoveOutboundResourceGroup(domainId, id int64) (*model.OutboundResourceGroup, *model.AppError) {
	group, err := a.Store.OutboundResourceGroup().Get(domainId, id)

	if err != nil {
		return nil, err
	}

	err = a.Store.OutboundResourceGroup().Delete(domainId, id)
	if err != nil {
		return nil, err
	}
	return group, nil
}

func (a *App) CreateOutboundResourceInGroup(domainId int64, res *model.OutboundResourceInGroup) (*model.OutboundResourceInGroup, *model.AppError) {
	return a.Store.OutboundResourceInGroup().Create(domainId, res)
}

func (a *App) GetOutboundResourceInGroupPage(domainId, groupId int64, search *model.SearchOutboundResourceInGroup) ([]*model.OutboundResourceInGroup, bool, *model.AppError) {
	list, err := a.Store.OutboundResourceInGroup().GetAllPage(domainId, groupId, search)
	if err != nil {
		return nil, false, err
	}
	search.RemoveLastElemIfNeed(&list)
	return list, search.EndOfList(), nil
}

func (app *App) GetOutboundResourceInGroup(domainId, groupId, id int64) (*model.OutboundResourceInGroup, *model.AppError) {
	return app.Store.OutboundResourceInGroup().Get(domainId, groupId, id)
}

func (a *App) UpdateOutboundResourceInGroup(domainId int64, res *model.OutboundResourceInGroup) (*model.OutboundResourceInGroup, *model.AppError) {
	oldRes, err := a.GetOutboundResourceInGroup(domainId, res.GroupId, res.Id)
	if err != nil {
		return nil, err
	}
	oldRes.Resource = res.Resource
	oldRes.Priority = res.Priority
	oldRes.ReserveResource = res.ReserveResource

	oldRes, err = a.Store.OutboundResourceInGroup().Update(domainId, oldRes)
	if err != nil {
		return nil, err
	}

	return oldRes, nil
}

func (a *App) RemoveOutboundResourceInGroup(domainId, groupId, id int64) (*model.OutboundResourceInGroup, *model.AppError) {
	res, err := a.Store.OutboundResourceInGroup().Get(domainId, groupId, id)

	if err != nil {
		return nil, err
	}

	err = a.Store.OutboundResourceInGroup().Delete(domainId, groupId, id)
	if err != nil {
		return nil, err
	}
	return res, nil
}
