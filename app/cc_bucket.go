package app

import (
	"github.com/webitel/engine/auth_manager"
	"github.com/webitel/engine/model"
)

func (app *App) CreateBucket(bucket *model.Bucket) (*model.Bucket, *model.AppError) {
	return app.Store.Bucket().Create(bucket)
}

func (a *App) BucketCheckAccess(domainId, id int64, groups []int, access auth_manager.PermissionAccess) (bool, *model.AppError) {
	return a.Store.Bucket().CheckAccess(domainId, id, groups, access)
}

func (app *App) GetBucket(id, domainId int64) (*model.Bucket, *model.AppError) {
	return app.Store.Bucket().Get(domainId, id)
}

func (app *App) GetBucketsPage(domainId int64, search *model.SearchBucket) ([]*model.Bucket, bool, *model.AppError) {
	list, err := app.Store.Bucket().GetAllPage(domainId, search)
	if err != nil {
		return nil, false, err
	}
	search.RemoveLastElemIfNeed(&list)
	return list, search.EndOfList(), nil
}

func (a *App) GetBucketsPageByGroups(domainId int64, groups []int, search *model.SearchBucket) ([]*model.Bucket, bool, *model.AppError) {
	list, err := a.Store.Bucket().GetAllPageByGroups(domainId, groups, search)
	if err != nil {
		return nil, false, err
	}
	search.RemoveLastElemIfNeed(&list)
	return list, search.EndOfList(), nil
}

func (app *App) UpdateBucket(bucket *model.Bucket) (*model.Bucket, *model.AppError) {
	oldBucket, err := app.GetBucket(bucket.Id, bucket.DomainId)

	if err != nil {
		return nil, err
	}

	oldBucket.Name = bucket.Name
	oldBucket.Description = bucket.Description

	oldBucket.UpdatedAt = bucket.UpdatedAt
	oldBucket.UpdatedBy.Id = bucket.UpdatedBy.Id

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
