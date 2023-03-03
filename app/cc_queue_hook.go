package app

import (
	"context"
	"github.com/webitel/engine/model"
)

func (app *App) CreateQueueHook(ctx context.Context, domainId int64, queueId uint32, hook *model.QueueHook) (*model.QueueHook, *model.AppError) {
	return app.Store.QueueHook().Create(ctx, domainId, queueId, hook)
}

func (app *App) SearchQueueHook(ctx context.Context, domainId int64, queueId uint32, search *model.SearchQueueHook) ([]*model.QueueHook, bool, *model.AppError) {
	list, err := app.Store.QueueHook().GetAllPage(ctx, domainId, queueId, search)
	if err != nil {
		return nil, false, err
	}
	search.RemoveLastElemIfNeed(&list)
	return list, search.EndOfList(), nil
}

func (app *App) GetQueueHook(ctx context.Context, domainId int64, queueId, id uint32) (*model.QueueHook, *model.AppError) {
	return app.Store.QueueHook().Get(ctx, domainId, queueId, id)
}

func (app *App) UpdateQueueHook(ctx context.Context, domainId int64, queueId uint32, hook *model.QueueHook) (*model.QueueHook, *model.AppError) {
	oldHook, err := app.GetQueueHook(ctx, domainId, queueId, hook.Id)
	if err != nil {
		return nil, err
	}

	oldHook.Schema = hook.Schema
	oldHook.Enabled = hook.Enabled
	oldHook.Event = hook.Event
	oldHook.UpdatedAt = hook.UpdatedAt
	oldHook.UpdatedBy = hook.UpdatedBy

	oldHook, err = app.Store.QueueHook().Update(ctx, domainId, queueId, oldHook)
	if err != nil {
		return nil, err
	}

	return oldHook, nil
}

func (app *App) PatchQueueHook(ctx context.Context, domainId int64, queueId, id uint32, patch *model.QueueHookPatch) (*model.QueueHook, *model.AppError) {
	oldHook, err := app.GetQueueHook(ctx, domainId, queueId, id)
	if err != nil {
		return nil, err
	}

	oldHook.Patch(patch)
	oldHook.UpdatedBy = &patch.UpdatedBy
	oldHook.UpdatedAt = patch.UpdatedAt

	if err = oldHook.IsValid(); err != nil {
		return nil, err
	}

	oldHook, err = app.Store.QueueHook().Update(ctx, domainId, queueId, oldHook)
	if err != nil {
		return nil, err
	}

	return oldHook, nil
}

func (app *App) RemoveQueueHook(ctx context.Context, domainId int64, queueId, id uint32) (*model.QueueHook, *model.AppError) {
	qb, err := app.GetQueueHook(ctx, domainId, queueId, id)

	if err != nil {
		return nil, err
	}

	err = app.Store.QueueHook().Delete(ctx, domainId, queueId, id)
	if err != nil {
		return nil, err
	}
	return qb, nil
}
