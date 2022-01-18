package app

import "github.com/webitel/engine/model"

func (app *App) CreateQueueHook(domainId int64, queueId uint32, hook *model.QueueHook) (*model.QueueHook, *model.AppError) {
	return app.Store.QueueHook().Create(domainId, queueId, hook)
}

func (app *App) SearchQueueHook(domainId int64, queueId uint32, search *model.SearchQueueHook) ([]*model.QueueHook, bool, *model.AppError) {
	list, err := app.Store.QueueHook().GetAllPage(domainId, queueId, search)
	if err != nil {
		return nil, false, err
	}
	search.RemoveLastElemIfNeed(&list)
	return list, search.EndOfList(), nil
}

func (app *App) GetQueueHook(domainId int64, queueId, id uint32) (*model.QueueHook, *model.AppError) {
	return app.Store.QueueHook().Get(domainId, queueId, id)
}

func (app *App) UpdateQueueHook(domainId int64, queueId uint32, hook *model.QueueHook) (*model.QueueHook, *model.AppError) {
	oldHook, err := app.GetQueueHook(domainId, queueId, hook.Id)
	if err != nil {
		return nil, err
	}

	oldHook.Schema = hook.Schema
	oldHook.Enabled = hook.Enabled
	oldHook.Event = hook.Event
	oldHook.UpdatedAt = hook.UpdatedAt
	oldHook.UpdatedBy = hook.UpdatedBy

	oldHook, err = app.Store.QueueHook().Update(domainId, queueId, oldHook)
	if err != nil {
		return nil, err
	}

	return oldHook, nil
}

func (app *App) PatchQueueHook(domainId int64, queueId, id uint32, patch *model.QueueHookPatch) (*model.QueueHook, *model.AppError) {
	oldHook, err := app.GetQueueHook(domainId, queueId, id)
	if err != nil {
		return nil, err
	}

	oldHook.Patch(patch)
	oldHook.UpdatedBy = &patch.UpdatedBy
	oldHook.UpdatedAt = patch.UpdatedAt

	if err = oldHook.IsValid(); err != nil {
		return nil, err
	}

	oldHook, err = app.Store.QueueHook().Update(domainId, queueId, oldHook)
	if err != nil {
		return nil, err
	}

	return oldHook, nil
}

func (app *App) RemoveQueueHook(domainId int64, queueId, id uint32) (*model.QueueHook, *model.AppError) {
	qb, err := app.GetQueueHook(domainId, queueId, id)

	if err != nil {
		return nil, err
	}

	err = app.Store.QueueHook().Delete(domainId, queueId, id)
	if err != nil {
		return nil, err
	}
	return qb, nil
}
