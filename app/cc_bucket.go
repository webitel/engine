package app

import (
	"context"

	"github.com/webitel/engine/model"
	"github.com/webitel/engine/pkg/wbt/auth_manager"
)

func (app *App) CreateBucket(ctx context.Context, bucket *model.Bucket) (*model.Bucket, model.AppError) {
	return app.Store.Bucket().Create(ctx, bucket)
}

func (a *App) BucketCheckAccess(ctx context.Context, domainId, id int64, groups []int, access auth_manager.PermissionAccess) (bool, model.AppError) {
	return a.Store.Bucket().CheckAccess(ctx, domainId, id, groups, access)
}

func (app *App) GetBucket(ctx context.Context, id, domainId int64) (*model.Bucket, model.AppError) {
	return app.Store.Bucket().Get(ctx, domainId, id)
}

func (app *App) GetBucketsPage(ctx context.Context, domainId int64, search *model.SearchBucket) ([]*model.Bucket, bool, model.AppError) {
	list, err := app.Store.Bucket().GetAllPage(ctx, domainId, search)
	if err != nil {
		return nil, false, err
	}
	search.RemoveLastElemIfNeed(&list)
	return list, search.EndOfList(), nil
}

func (app *App) UpdateBucket(ctx context.Context, bucket *model.Bucket) (*model.Bucket, model.AppError) {
	oldBucket, err := app.GetBucket(ctx, bucket.Id, bucket.DomainId)

	if err != nil {
		return nil, err
	}

	oldBucket.Name = bucket.Name
	oldBucket.Description = bucket.Description

	oldBucket.UpdatedAt = bucket.UpdatedAt
	oldBucket.UpdatedBy = bucket.UpdatedBy

	_, err = app.Store.Bucket().Update(ctx, oldBucket)
	if err != nil {
		return nil, err
	}

	return oldBucket, nil
}

func (app *App) RemoveBucket(ctx context.Context, domainId, id int64) (*model.Bucket, model.AppError) {
	bucket, err := app.GetBucket(ctx, id, domainId)

	if err != nil {
		return nil, err
	}

	err = app.Store.Bucket().Delete(ctx, domainId, id)
	if err != nil {
		return nil, err
	}
	return bucket, nil
}
