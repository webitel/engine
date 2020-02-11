package app

import "github.com/webitel/engine/model"

func (app *App) CreateQueueBucket(queueBucket *model.QueueBucket) (*model.QueueBucket, *model.AppError) {
	return app.Store.BucketInQueue().Create(queueBucket)
}

func (app *App) GetQueueBucketPage(domainId, queueId int64, search *model.SearchQueueBucket) ([]*model.QueueBucket, bool, *model.AppError) {
	list, err := app.Store.BucketInQueue().GetAllPage(domainId, queueId, search)
	if err != nil {
		return nil, false, err
	}
	search.RemoveLastElemIfNeed(&list)
	return list, search.EndOfList(), nil
}

func (app *App) GetQueueBucket(domainId, queueId, id int64) (*model.QueueBucket, *model.AppError) {
	return app.Store.BucketInQueue().Get(domainId, queueId, id)
}

func (app *App) UpdateQueueBucket(domainId int64, qb *model.QueueBucket) (*model.QueueBucket, *model.AppError) {
	oldQb, err := app.GetQueueBucket(domainId, qb.QueueId, qb.Id)
	if err != nil {
		return nil, err
	}

	oldQb.Ratio = qb.Ratio
	oldQb.Bucket = qb.Bucket

	oldQb, err = app.Store.BucketInQueue().Update(domainId, oldQb)
	if err != nil {
		return nil, err
	}

	return oldQb, nil
}

func (app *App) RemoveQueueBucket(domainId, queueId, id int64) (*model.QueueBucket, *model.AppError) {
	qb, err := app.GetQueueBucket(domainId, queueId, id)

	if err != nil {
		return nil, err
	}

	err = app.Store.BucketInQueue().Delete(queueId, id)
	if err != nil {
		return nil, err
	}
	return qb, nil
}
