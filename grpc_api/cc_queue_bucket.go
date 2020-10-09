package grpc_api

import (
	"context"
	"github.com/webitel/engine/app"
	"github.com/webitel/engine/auth_manager"
	"github.com/webitel/engine/model"
	"github.com/webitel/protos/engine"
)

type queueBucket struct {
	app *app.App
}

func NewQueueBucketApi(app *app.App) *queueBucket {
	return &queueBucket{app: app}
}

func (api *queueBucket) CreateQueueBucket(ctx context.Context, in *engine.CreateQueueBucketRequest) (*engine.QueueBucket, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_QUEUE)
	if !permission.CanRead() {
		return nil, api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if !permission.CanUpdate() {
		return nil, api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
	}

	if permission.Rbac {
		var perm bool
		if perm, err = api.app.QueueCheckAccess(session.Domain(in.GetDomainId()), in.GetQueueId(), session.GetAclRoles(), auth_manager.PERMISSION_ACCESS_UPDATE); err != nil {
			return nil, err
		} else if !perm {
			return nil, api.app.MakeResourcePermissionError(session, in.GetQueueId(), permission, auth_manager.PERMISSION_ACCESS_UPDATE)
		}
	}

	var queueBucket = &model.QueueBucket{
		QueueId: in.GetQueueId(),
		Bucket: model.Lookup{
			Id: int(in.GetBucket().GetId()),
		},
		Ratio: int(in.GetRatio()),
	}

	if err = queueBucket.IsValid(); err != nil {
		return nil, err
	}

	queueBucket, err = api.app.CreateQueueBucket(queueBucket)
	if err != nil {
		return nil, err
	}

	return toEngineQueueBucket(queueBucket), nil
}

func (api *queueBucket) ReadQueueBucket(ctx context.Context, in *engine.ReadQueueBucketRequest) (*engine.QueueBucket, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_QUEUE)
	if !permission.CanRead() {
		return nil, api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if permission.Rbac {
		var perm bool
		if perm, err = api.app.QueueCheckAccess(session.Domain(in.GetDomainId()), in.GetQueueId(), session.GetAclRoles(), auth_manager.PERMISSION_ACCESS_READ); err != nil {
			return nil, err
		} else if !perm {
			return nil, api.app.MakeResourcePermissionError(session, in.GetQueueId(), permission, auth_manager.PERMISSION_ACCESS_READ)
		}
	}

	var out *model.QueueBucket
	out, err = api.app.GetQueueBucket(session.Domain(in.GetDomainId()), in.GetQueueId(), in.GetId())
	if err != nil {
		return nil, err
	}
	return toEngineQueueBucket(out), nil
}

func (api *queueBucket) SearchQueueBucket(ctx context.Context, in *engine.SearchQueueBucketRequest) (*engine.ListQueueBucket, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_QUEUE)
	if !permission.CanRead() {
		return nil, api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if permission.Rbac {
		var perm bool
		if perm, err = api.app.QueueCheckAccess(session.Domain(in.GetDomainId()), in.GetQueueId(), session.GetAclRoles(), auth_manager.PERMISSION_ACCESS_READ); err != nil {
			return nil, err
		} else if !perm {
			return nil, api.app.MakeResourcePermissionError(session, in.GetQueueId(), permission, auth_manager.PERMISSION_ACCESS_READ)
		}
	}

	var list []*model.QueueBucket
	var endList bool
	req := &model.SearchQueueBucket{
		ListRequest: model.ListRequest{
			DomainId: in.GetDomainId(),
			Q:        in.GetQ(),
			Page:     int(in.GetPage()),
			PerPage:  int(in.GetSize()),
		},
	}

	list, endList, err = api.app.GetQueueBucketPage(session.Domain(in.DomainId), in.GetQueueId(), req)
	if err != nil {
		return nil, err
	}

	items := make([]*engine.QueueBucket, 0, len(list))
	for _, v := range list {
		items = append(items, &engine.QueueBucket{
			Id:    v.Id,
			Ratio: int32(v.Ratio),
			Bucket: &engine.Lookup{
				Id:   int64(v.Bucket.Id),
				Name: v.Bucket.Name,
			},
		})
	}
	return &engine.ListQueueBucket{
		Next:  !endList,
		Items: items,
	}, nil
}

func (api *queueBucket) UpdateQueueBucket(ctx context.Context, in *engine.UpdateQueueBucketRequest) (*engine.QueueBucket, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_QUEUE)
	if !permission.CanRead() {
		return nil, api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if !permission.CanUpdate() {
		return nil, api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
	}

	if permission.Rbac {
		var perm bool
		if perm, err = api.app.QueueCheckAccess(session.Domain(in.GetDomainId()), in.GetQueueId(), session.GetAclRoles(), auth_manager.PERMISSION_ACCESS_UPDATE); err != nil {
			return nil, err
		} else if !perm {
			return nil, api.app.MakeResourcePermissionError(session, in.GetQueueId(), permission, auth_manager.PERMISSION_ACCESS_UPDATE)
		}
	}

	qb := &model.QueueBucket{
		Id:      in.GetId(),
		QueueId: in.GetQueueId(),
		Bucket: model.Lookup{
			Id: int(in.GetBucket().GetId()),
		},
		Ratio: int(in.GetRatio()),
	}

	if err = qb.IsValid(); err != nil {
		return nil, err
	}

	qb, err = api.app.UpdateQueueBucket(session.Domain(in.GetDomainId()), qb)
	if err != nil {
		return nil, err
	}

	return toEngineQueueBucket(qb), nil
}

func (api *queueBucket) DeleteQueueBucket(ctx context.Context, in *engine.DeleteQueueBucketRequest) (*engine.QueueBucket, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_QUEUE)
	if !permission.CanRead() {
		return nil, api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if !permission.CanUpdate() {
		return nil, api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
	}

	if permission.Rbac {
		var perm bool
		if perm, err = api.app.QueueCheckAccess(session.Domain(in.GetDomainId()), in.GetQueueId(), session.GetAclRoles(), auth_manager.PERMISSION_ACCESS_UPDATE); err != nil {
			return nil, err
		} else if !perm {
			return nil, api.app.MakeResourcePermissionError(session, in.GetQueueId(), permission, auth_manager.PERMISSION_ACCESS_UPDATE)
		}
	}

	var qb *model.QueueBucket
	qb, err = api.app.RemoveQueueBucket(session.Domain(in.GetDomainId()), in.GetQueueId(), in.GetId())
	if err != nil {
		return nil, err
	}
	return toEngineQueueBucket(qb), nil
}

func toEngineQueueBucket(src *model.QueueBucket) *engine.QueueBucket {
	return &engine.QueueBucket{
		Id:    src.Id,
		Ratio: int32(src.Ratio),
		Bucket: &engine.Lookup{
			Id:   int64(src.Bucket.Id),
			Name: src.Bucket.Name,
		},
	}
}
