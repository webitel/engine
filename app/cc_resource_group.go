package app

import (
	"context"
	"github.com/webitel/engine/auth_manager"
	"github.com/webitel/engine/model"
)

func (a *App) OutboundResourceGroupCheckAccess(ctx context.Context, domainId, id int64, groups []int, access auth_manager.PermissionAccess) (bool, *model.AppError) {
	return a.Store.OutboundResourceGroup().CheckAccess(ctx, domainId, id, groups, access)
}

func (a *App) CreateOutboundResourceGroup(ctx context.Context, group *model.OutboundResourceGroup) (*model.OutboundResourceGroup, *model.AppError) {
	return a.Store.OutboundResourceGroup().Create(ctx, group)
}

func (a *App) GetOutboundResourceGroupPage(ctx context.Context, domainId int64, search *model.SearchOutboundResourceGroup) ([]*model.OutboundResourceGroup, bool, *model.AppError) {
	list, err := a.Store.OutboundResourceGroup().GetAllPage(ctx, domainId, search)
	if err != nil {
		return nil, false, err
	}
	search.RemoveLastElemIfNeed(&list)
	return list, search.EndOfList(), nil
}

func (a *App) GetOutboundResourceGroupPageByGroups(ctx context.Context, domainId int64, groups []int, search *model.SearchOutboundResourceGroup) ([]*model.OutboundResourceGroup, bool, *model.AppError) {
	list, err := a.Store.OutboundResourceGroup().GetAllPageByGroups(ctx, domainId, groups, search)
	if err != nil {
		return nil, false, err
	}
	search.RemoveLastElemIfNeed(&list)
	return list, search.EndOfList(), nil
}

func (app *App) GetOutboundResourceGroup(ctx context.Context, domainId, id int64) (*model.OutboundResourceGroup, *model.AppError) {
	return app.Store.OutboundResourceGroup().Get(ctx, domainId, id)
}

func (a *App) UpdateOutboundResourceGroup(ctx context.Context, group *model.OutboundResourceGroup) (*model.OutboundResourceGroup, *model.AppError) {
	oldGroup, err := a.GetOutboundResourceGroup(ctx, group.DomainId, group.Id)
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

	oldGroup, err = a.Store.OutboundResourceGroup().Update(ctx, oldGroup)
	if err != nil {
		return nil, err
	}

	return oldGroup, nil
}

func (a *App) RemoveOutboundResourceGroup(ctx context.Context, domainId, id int64) (*model.OutboundResourceGroup, *model.AppError) {
	group, err := a.Store.OutboundResourceGroup().Get(ctx, domainId, id)

	if err != nil {
		return nil, err
	}

	err = a.Store.OutboundResourceGroup().Delete(ctx, domainId, id)
	if err != nil {
		return nil, err
	}
	return group, nil
}

func (a *App) CreateOutboundResourceInGroup(ctx context.Context, domainId int64, res *model.OutboundResourceInGroup) (*model.OutboundResourceInGroup, *model.AppError) {
	return a.Store.OutboundResourceInGroup().Create(ctx, domainId, res)
}

func (a *App) GetOutboundResourceInGroupPage(ctx context.Context, domainId, groupId int64, search *model.SearchOutboundResourceInGroup) ([]*model.OutboundResourceInGroup, bool, *model.AppError) {
	list, err := a.Store.OutboundResourceInGroup().GetAllPage(ctx, domainId, groupId, search)
	if err != nil {
		return nil, false, err
	}
	search.RemoveLastElemIfNeed(&list)
	return list, search.EndOfList(), nil
}

func (app *App) GetOutboundResourceInGroup(ctx context.Context, domainId, groupId, id int64) (*model.OutboundResourceInGroup, *model.AppError) {
	return app.Store.OutboundResourceInGroup().Get(ctx, domainId, groupId, id)
}

func (a *App) UpdateOutboundResourceInGroup(ctx context.Context, domainId int64, res *model.OutboundResourceInGroup) (*model.OutboundResourceInGroup, *model.AppError) {
	oldRes, err := a.GetOutboundResourceInGroup(ctx, domainId, res.GroupId, res.Id)
	if err != nil {
		return nil, err
	}
	oldRes.Resource = res.Resource
	oldRes.Priority = res.Priority
	oldRes.ReserveResource = res.ReserveResource

	oldRes, err = a.Store.OutboundResourceInGroup().Update(ctx, domainId, oldRes)
	if err != nil {
		return nil, err
	}

	return oldRes, nil
}

func (a *App) RemoveOutboundResourceInGroup(ctx context.Context, domainId, groupId, id int64) (*model.OutboundResourceInGroup, *model.AppError) {
	res, err := a.Store.OutboundResourceInGroup().Get(ctx, domainId, groupId, id)

	if err != nil {
		return nil, err
	}

	err = a.Store.OutboundResourceInGroup().Delete(ctx, domainId, groupId, id)
	if err != nil {
		return nil, err
	}
	return res, nil
}
