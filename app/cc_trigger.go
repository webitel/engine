package app

import (
	"context"

	"github.com/webitel/engine/auth_manager"
	"github.com/webitel/engine/model"
)

func (a *App) TriggerCheckAccess(ctx context.Context, domainId int64, id int32, groups []int, access auth_manager.PermissionAccess) (bool, model.AppError) {
	return a.Store.Trigger().CheckAccess(ctx, domainId, id, groups, access)
}

func (a *App) CreateTrigger(ctx context.Context, domainId int64, trigger *model.Trigger) (*model.Trigger, model.AppError) {
	createdTrigger, err := a.Store.Trigger().Create(ctx, domainId, trigger)
	if err != nil {
		return nil, err
	}

	// notify about triggers were changed
	if a.TriggerCases != nil {
		a.TriggerCases.NotifyUpdateTrigger()
	}

	return createdTrigger, nil
}

func (a *App) GetTriggerList(ctx context.Context, domainId int64, search *model.SearchTrigger) ([]*model.Trigger, bool, model.AppError) {
	list, err := a.Store.Trigger().GetAllPage(ctx, domainId, search)
	if err != nil {
		return nil, false, err
	}
	search.RemoveLastElemIfNeed(&list)
	return list, search.EndOfList(), nil
}

func (a *App) GetTriggerListByGroups(ctx context.Context, domainId int64, groups []int, search *model.SearchTrigger) ([]*model.Trigger, bool, model.AppError) {
	list, err := a.Store.Trigger().GetAllPageByGroup(ctx, domainId, groups, search)
	if err != nil {
		return nil, false, err
	}
	search.RemoveLastElemIfNeed(&list)
	return list, search.EndOfList(), nil
}

func (a *App) GetTrigger(ctx context.Context, domainId int64, id int32) (*model.Trigger, model.AppError) {
	return a.Store.Trigger().Get(ctx, domainId, id)
}

func (a *App) UpdateTrigger(ctx context.Context, domainId int64, trigger *model.Trigger) (*model.Trigger, model.AppError) {
	oldTrigger, err := a.GetTrigger(ctx, domainId, trigger.Id)
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
	oldTrigger.Type = trigger.Type

	oldTrigger, err = a.Store.Trigger().Update(ctx, domainId, oldTrigger)
	if err != nil {
		return nil, err
	}

	// notify about triggers were changed
	if a.TriggerCases != nil {
		a.TriggerCases.NotifyUpdateTrigger()
	}

	return oldTrigger, nil
}

func (a *App) PatchTrigger(ctx context.Context, domainId int64, id int32, patch *model.TriggerPatch) (*model.Trigger, model.AppError) {
	oldTrigger, err := a.GetTrigger(ctx, domainId, id)
	if err != nil {
		return nil, err
	}

	oldTrigger.Patch(patch)

	if err = oldTrigger.IsValid(); err != nil {
		return nil, err
	}

	oldTrigger, err = a.Store.Trigger().Update(ctx, domainId, oldTrigger)
	if err != nil {
		return nil, err
	}

	// notify about triggers were changed
	if a.TriggerCases != nil {
		a.TriggerCases.NotifyUpdateTrigger()
	}

	return oldTrigger, nil
}

func (a *App) RemoveTrigger(ctx context.Context, domainId int64, id int32) (*model.Trigger, model.AppError) {
	trigger, err := a.Store.Trigger().Get(ctx, domainId, id)

	if err != nil {
		return nil, err
	}

	err = a.Store.Trigger().Delete(ctx, domainId, id)
	if err != nil {
		return nil, err
	}

	// notify about triggers were changed
	if a.TriggerCases != nil {
		a.TriggerCases.NotifyUpdateTrigger()
	}

	return trigger, nil
}

func (a *App) GetTriggerJobList(ctx context.Context, domainId int64, triggerId int32, search *model.SearchTriggerJob) ([]*model.TriggerJob, bool, model.AppError) {
	var list []*model.TriggerJob
	_, err := a.Store.Trigger().Get(ctx, domainId, triggerId)

	list, err = a.Store.Trigger().GetAllJobs(ctx, triggerId, search)
	if err != nil {
		return nil, false, err
	}
	search.RemoveLastElemIfNeed(&list)
	return list, search.EndOfList(), nil
}

func (a *App) CreateTriggerJob(ctx context.Context, domainId int64, triggerId int32, vars map[string]string) (*model.TriggerJob, model.AppError) {
	_, err := a.Store.Trigger().Get(ctx, domainId, triggerId)
	if err != nil {
		return nil, err
	}

	return a.Store.Trigger().CreateJob(ctx, domainId, triggerId, vars)
}
