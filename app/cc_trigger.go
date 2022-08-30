package app

import (
	"github.com/webitel/engine/auth_manager"
	"github.com/webitel/engine/model"
)

func (a *App) TriggerCheckAccess(domainId int64, id int32, groups []int, access auth_manager.PermissionAccess) (bool, *model.AppError) {
	return a.Store.Trigger().CheckAccess(domainId, id, groups, access)
}

func (a *App) CreateTrigger(domainId int64, trigger *model.Trigger) (*model.Trigger, *model.AppError) {
	return a.Store.Trigger().Create(domainId, trigger)
}

func (a *App) GetTriggerList(domainId int64, search *model.SearchTrigger) ([]*model.Trigger, bool, *model.AppError) {
	list, err := a.Store.Trigger().GetAllPage(domainId, search)
	if err != nil {
		return nil, false, err
	}
	search.RemoveLastElemIfNeed(&list)
	return list, search.EndOfList(), nil
}

func (a *App) GetTriggerListByGroups(domainId int64, groups []int, search *model.SearchTrigger) ([]*model.Trigger, bool, *model.AppError) {
	list, err := a.Store.Trigger().GetAllPageByGroup(domainId, groups, search)
	if err != nil {
		return nil, false, err
	}
	search.RemoveLastElemIfNeed(&list)
	return list, search.EndOfList(), nil
}

func (a *App) GetTrigger(domainId int64, id int32) (*model.Trigger, *model.AppError) {
	return a.Store.Trigger().Get(domainId, id)
}

func (a *App) UpdateTrigger(domainId int64, trigger *model.Trigger) (*model.Trigger, *model.AppError) {
	oldTrigger, err := a.GetTrigger(domainId, trigger.Id)
	if err != nil {
		return nil, err
	}

	oldTrigger.UpdatedAt = trigger.UpdatedAt
	oldTrigger.UpdatedBy = trigger.UpdatedBy
	oldTrigger.Name = trigger.Name
	oldTrigger.Enabled = trigger.Enabled
	oldTrigger.Schema = trigger.Schema
	oldTrigger.Variables = trigger.Variables
	oldTrigger.Description = trigger.Description
	oldTrigger.Expression = trigger.Expression
	oldTrigger.Timezone = trigger.Timezone
	oldTrigger.Timeout = trigger.Timeout

	oldTrigger, err = a.Store.Trigger().Update(domainId, oldTrigger)
	if err != nil {
		return nil, err
	}

	return oldTrigger, nil
}

func (a *App) PatchTrigger(domainId int64, id int32, patch *model.TriggerPatch) (*model.Trigger, *model.AppError) {
	oldTrigger, err := a.GetTrigger(domainId, id)
	if err != nil {
		return nil, err
	}

	oldTrigger.Patch(patch)

	if err = oldTrigger.IsValid(); err != nil {
		return nil, err
	}

	oldTrigger, err = a.Store.Trigger().Update(domainId, oldTrigger)
	if err != nil {
		return nil, err
	}

	return oldTrigger, nil
}

func (a *App) RemoveTrigger(domainId int64, id int32) (*model.Trigger, *model.AppError) {
	trigger, err := a.Store.Trigger().Get(domainId, id)

	if err != nil {
		return nil, err
	}

	err = a.Store.Trigger().Delete(domainId, id)
	if err != nil {
		return nil, err
	}

	return trigger, nil
}
