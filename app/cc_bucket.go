package app

import "github.com/webitel/engine/model"

func (app *App) CreateBucket(bucket *model.Bucket) (*model.Bucket, *model.AppError) {
	return app.Store.Bucket().Create(bucket)
}

func (app *App) GetBucket(id, domainId int64) (*model.Bucket, *model.AppError) {
	return app.Store.Bucket().Get(domainId, id)
}

func (app *App) GetBucketsPage(domainId int64, page, perPage int) ([]*model.Bucket, *model.AppError) {
	return app.Store.Bucket().GetAllPage(domainId, page*perPage, perPage)
}

func (app *App) UpdateBucket(bucket *model.Bucket) (*model.Bucket, *model.AppError) {
	oldBucket, err := app.GetBucket(bucket.Id, bucket.DomainId)

	if err != nil {
		return nil, err
	}

	oldBucket.Name = bucket.Name
	oldBucket.Description = bucket.Description

	_, err = app.Store.Bucket().Update(oldBucket)
	if err != nil {
		return nil, err
	}

	return oldBucket, nil
}

func (app *App) RemoveBucket(domainId, id int64) (*model.Bucket, *model.AppError) {
	bucket, err := app.GetBucket(id, domainId)

	if err != nil {
		return nil, err
	}

	err = app.Store.Bucket().Delete(domainId, id)
	if err != nil {
		return nil, err
	}
	return bucket, nil
}
