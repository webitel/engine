package grpc_api

import (
	"context"
	"github.com/webitel/engine/app"
	"github.com/webitel/engine/auth_manager"
	"github.com/webitel/engine/model"
	"github.com/webitel/protos/engine"
)

type queueRouting struct {
	app *app.App
}

func NewQueueRoutingApi(app *app.App) *queueRouting {
	return &queueRouting{app: app}
}

func (api *queueRouting) CreateQueueRouting(ctx context.Context, in *engine.CreateQueueRoutingRequest) (*engine.QueueRouting, error) {
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

	var routing = &model.QueueRouting{
		QueueId:  in.GetQueueId(),
		Pattern:  in.GetPattern(),
		Priority: int(in.GetPriority()),
		Disabled: in.GetDisabled(),
	}

	if err = routing.IsValid(); err != nil {
		return nil, err
	}

	routing, err = api.app.CreateQueueRouting(routing)
	if err != nil {
		return nil, err
	}

	return toEngineQueueRouting(routing), nil
}

func (api *queueRouting) SearchQueueRouting(ctx context.Context, in *engine.SearchQueueRoutingRequest) (*engine.ListQueueRouting, error) {
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

	var list []*model.QueueRouting
	list, err = api.app.GetQueueRoutingPage(session.Domain(int64(in.DomainId)), in.GetQueueId(), int(in.Page), int(in.Size))
	if err != nil {
		return nil, err
	}

	items := make([]*engine.QueueRouting, 0, len(list))
	for _, v := range list {
		items = append(items, &engine.QueueRouting{
			Id:       v.Id,
			QueueId:  v.QueueId,
			Pattern:  v.Pattern,
			Priority: int32(v.Priority),
			Disabled: v.Disabled,
		})
	}
	return &engine.ListQueueRouting{
		Items: items,
	}, nil
}

func (api *queueRouting) ReadQueueRouting(ctx context.Context, in *engine.ReadQueueRoutingRequest) (*engine.QueueRouting, error) {
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

	var out *model.QueueRouting
	out, err = api.app.GetQueueRoutingById(session.Domain(in.GetDomainId()), in.GetQueueId(), in.GetId())
	if err != nil {
		return nil, err
	}
	return toEngineQueueRouting(out), nil
}

func (api *queueRouting) UpdateQueueRouting(ctx context.Context, in *engine.UpdateQueueRoutingRequest) (*engine.QueueRouting, error) {
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

	qr := &model.QueueRouting{
		Id:       in.GetId(),
		QueueId:  in.GetQueueId(),
		Pattern:  in.GetPattern(),
		Priority: int(in.GetPriority()),
		Disabled: in.GetDisabled(),
	}

	if err = qr.IsValid(); err != nil {
		return nil, err
	}

	qr, err = api.app.UpdateQueueRouting(session.Domain(in.GetDomainId()), qr)
	if err != nil {
		return nil, err
	}

	return toEngineQueueRouting(qr), nil
}

func (api *queueRouting) DeleteQueueRouting(ctx context.Context, in *engine.DeleteQueueRoutingRequest) (*engine.QueueRouting, error) {
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

	var qr *model.QueueRouting
	qr, err = api.app.RemoveQueueRouting(session.Domain(in.GetDomainId()), in.GetQueueId(), in.GetId())
	if err != nil {
		return nil, err
	}
	return toEngineQueueRouting(qr), nil
}

func toEngineQueueRouting(src *model.QueueRouting) *engine.QueueRouting {
	return &engine.QueueRouting{
		Id:       src.Id,
		QueueId:  src.QueueId,
		Pattern:  src.Pattern,
		Priority: int32(src.Priority),
		Disabled: src.Disabled,
	}
}
