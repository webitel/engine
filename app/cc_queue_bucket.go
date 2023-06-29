package app

import (
	"context"

	"github.com/webitel/engine/model"
)

func (app *App) CreateQueueBucket(ctx context.Context, queueBucket *model.QueueBucket) (*model.QueueBucket, model.AppError) {
	return app.Store.BucketInQueue().Create(ctx, queueBucket)
}

func (app *App) GetQueueBucketPage(ctx context.Context, domainId, queueId int64, search *model.SearchQueueBucket) ([]*model.QueueBucket, bool, model.AppError) {
	list, err := app.Store.BucketInQueue().GetAllPage(ctx, domainId, queueId, search)
	if err != nil {
		return nil, false, err
	}
	search.RemoveLastElemIfNeed(&list)
	return list, search.EndOfList(), nil
}

func (app *App) GetQueueBucket(ctx context.Context, domainId, queueId, id int64) (*model.QueueBucket, model.AppError) {
	return app.Store.BucketInQueue().Get(ctx, domainId, queueId, id)
}

func (app *App) UpdateQueueBucket(ctx context.Context, domainId int64, qb *model.QueueBucket) (*model.QueueBucket, model.AppError) {
	oldQb, err := app.GetQueueBucket(ctx, domainId, qb.QueueId, qb.Id)
	if err != nil {
		return nil, err
	}

	oldQb.Ratio = qb.Ratio
	oldQb.Bucket = qb.Bucket
	oldQb.Priority = qb.Priority
	oldQb.Disabled = qb.Disabled

	oldQb, err = app.Store.BucketInQueue().Update(ctx, domainId, oldQb)
	if err != nil {
		return nil, err
	}

	return oldQb, nil
}

func (a *App) PatchQueueBucket(ctx context.Context, domainId, queueId, id int64, patch *model.QueueBucketPatch) (*model.QueueBucket, model.AppError) {
	oldQb, err := a.GetQueueBucket(ctx, domainId, queueId, id)
	if err != nil {
		return nil, err
	}

	oldQb.Patch(patch)

	if err = oldQb.IsValid(); err != nil {
		return nil, err
	}

	oldQb, err = a.Store.BucketInQueue().Update(ctx, domainId, oldQb)
	if err != nil {
		return nil, err
	}

	return oldQb, nil
}

func (app *App) RemoveQueueBucket(ctx context.Context, domainId, queueId, id int64) (*model.QueueBucket, model.AppError) {
	qb, err := app.GetQueueBucket(ctx, domainId, queueId, id)

	if err != nil {
		return nil, err
	}

	err = app.Store.BucketInQueue().Delete(ctx, queueId, id)
	if err != nil {
		return nil, err
	}
	return qb, nil
}
