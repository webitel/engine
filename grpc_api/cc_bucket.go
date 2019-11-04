package grpc_api

import (
	"context"
	"github.com/webitel/engine/app"
	"github.com/webitel/engine/grpc_api/engine"
	"github.com/webitel/engine/model"
)

type bucket struct {
	app *app.App
}

func NewBucketApi(app *app.App) *bucket {
	return &bucket{app}
}

func (api *bucket) CreateBucket(ctx context.Context, in *engine.CreateBucketRequest) (*engine.Bucket, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	bucket := &model.Bucket{
		Name:        in.GetName(),
		Description: in.GetDescription(),
		DomainId:    session.Domain(in.DomainId),
	}

	if err = bucket.IsValid(); err != nil {
		return nil, err
	}

	if bucket, err = api.app.CreateBucket(bucket); err != nil {
		return nil, err
	} else {
		return toEngineBucket(bucket), nil
	}
}

func (api *bucket) SearchBucket(ctx context.Context, in *engine.SearchBucketRequest) (*engine.ListBucket, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	var list []*model.Bucket
	list, err = api.app.GetBucketsPage(session.Domain(in.DomainId), int(in.Page), int(in.Size))
	if err != nil {
		return nil, err
	}

	items := make([]*engine.Bucket, 0, len(list))
	for _, v := range list {
		items = append(items, toEngineBucket(v))
	}
	return &engine.ListBucket{
		Items: items,
	}, nil
}

func (api *bucket) ReadBucket(ctx context.Context, in *engine.ReadBucketRequest) (*engine.Bucket, error) {
	var bucket *model.Bucket
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	bucket, err = api.app.GetBucket(in.Id, session.Domain(in.GetDomainId()))
	if err != nil {
		return nil, err
	}

	return toEngineBucket(bucket), nil
}

func (api *bucket) UpdateBucket(ctx context.Context, in *engine.UpdateBucketRequest) (*engine.Bucket, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	var bucket *model.Bucket

	bucket, err = api.app.UpdateBucket(&model.Bucket{
		Id:          in.Id,
		Name:        in.Name,
		DomainId:    session.Domain(in.GetDomainId()),
		Description: in.Description,
	})

	if err != nil {
		return nil, err
	}

	return toEngineBucket(bucket), nil
}

func (api *bucket) DeleteBucket(ctx context.Context, in *engine.DeleteBucketRequest) (*engine.Bucket, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	var bucket *model.Bucket
	bucket, err = api.app.RemoveBucket(session.Domain(in.DomainId), in.Id)
	if err != nil {
		return nil, err
	}

	return toEngineBucket(bucket), nil
}

func toEngineBucket(src *model.Bucket) *engine.Bucket {
	return &engine.Bucket{
		Id:          src.Id,
		Name:        src.Name,
		Description: src.Description,
	}
}
