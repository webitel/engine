package controller

import (
	"context"
	"strconv"

	"github.com/webitel/engine/model"
	"github.com/webitel/engine/pkg/wbt/auth_manager"
)

func (c *Controller) CreateQueue(ctx context.Context, session *auth_manager.Session, queue *model.Queue) (*model.Queue, model.AppError) {
	var err model.AppError
	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_QUEUE)
	if !permission.CanCreate() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_CREATE)
	}

	queue.DomainRecord = model.DomainRecord{
		DomainId:  session.Domain(0),
		CreatedAt: model.GetMillis(),
		CreatedBy: &model.Lookup{
			Id: int(session.UserId),
		},
		UpdatedAt: model.GetMillis(),
		UpdatedBy: &model.Lookup{
			Id: int(session.UserId),
		},
	}

	if err = queue.IsValid(); err != nil {
		return nil, err
	}

	if err := queue.TaskProcessing.ProlongationOptions.IsValid(); err != nil {
		return nil, err
	}

	queue, err = c.app.CreateQueue(ctx, queue)
	if err != nil {
		return nil, err
	}

	c.app.AuditCreate(ctx, session, model.PERMISSION_SCOPE_CC_QUEUE, strconv.FormatInt(queue.Id, 10), queue)

	return queue, nil
}

func (c *Controller) SearchQueue(ctx context.Context, session *auth_manager.Session, search *model.SearchQueue) ([]*model.Queue, bool, model.AppError) {
	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_QUEUE)
	if !permission.CanRead() {
		return nil, false, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if session.UseRBAC(auth_manager.PERMISSION_ACCESS_READ, permission) {
		return c.app.GetQueuePageByGroups(ctx, session.Domain(search.DomainId), session.GetAclRoles(), search)
	} else {
		return c.app.GetQueuePage(ctx, session.Domain(search.DomainId), search)
	}
}

func (c *Controller) GetQueue(ctx context.Context, session *auth_manager.Session, id int64) (*model.Queue, model.AppError) {
	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_QUEUE)
	if !permission.CanRead() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if session.UseRBAC(auth_manager.PERMISSION_ACCESS_READ, permission) {
		if perm, err := c.app.QueueCheckAccess(ctx, session.Domain(0), id, session.GetAclRoles(), auth_manager.PERMISSION_ACCESS_READ); err != nil {
			return nil, err
		} else if !perm {
			return nil, c.app.MakeResourcePermissionError(session, id, permission, auth_manager.PERMISSION_ACCESS_READ)
		}
	}

	return c.app.GetQueueById(ctx, session.Domain(0), id)
}

func (c *Controller) UpdateQueue(ctx context.Context, session *auth_manager.Session, queue *model.Queue) (*model.Queue, model.AppError) {
	var err model.AppError
	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_QUEUE)
	if !permission.CanRead() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if !permission.CanUpdate() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
	}

	if session.UseRBAC(auth_manager.PERMISSION_ACCESS_UPDATE, permission) {
		var perm bool
		if perm, err = c.app.QueueCheckAccess(ctx, session.Domain(0), queue.Id, session.GetAclRoles(),
			auth_manager.PERMISSION_ACCESS_UPDATE); err != nil {
			return nil, err
		} else if !perm {
			return nil, c.app.MakeResourcePermissionError(session, queue.Id, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
		}
	}

	queue.DomainId = session.Domain(0)
	queue.UpdatedAt = model.GetMillis()
	queue.UpdatedBy = &model.Lookup{
		Id: int(session.UserId),
	}

	if err := queue.IsValid(); err != nil {
		return nil, err
	}

	if err := queue.TaskProcessing.ProlongationOptions.IsValid(); err != nil {
		return nil, err
	}

	queue, err = c.app.UpdateQueue(ctx, queue)
	if err != nil {
		return nil, err
	}

	c.app.AuditUpdate(ctx, session, model.PERMISSION_SCOPE_CC_QUEUE, strconv.FormatInt(queue.Id, 10), queue)

	return queue, nil
}

func (c *Controller) PatchQueue(ctx context.Context, session *auth_manager.Session, id int64, patch *model.QueuePatch) (*model.Queue, model.AppError) {
	var err model.AppError
	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_QUEUE)
	if !permission.CanRead() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if !permission.CanUpdate() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
	}

	if session.UseRBAC(auth_manager.PERMISSION_ACCESS_UPDATE, permission) {
		var perm bool
		if perm, err = c.app.QueueCheckAccess(ctx, session.Domain(0), id, session.GetAclRoles(),
			auth_manager.PERMISSION_ACCESS_UPDATE); err != nil {
			return nil, err
		} else if !perm {
			return nil, c.app.MakeResourcePermissionError(session, id, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
		}
	}

	patch.UpdatedBy = model.Lookup{
		Id: int(session.UserId),
	}

	var queue *model.Queue
	queue, err = c.app.PatchQueue(ctx, session.Domain(0), id, patch)
	if err != nil {
		return nil, err
	}

	c.app.AuditUpdate(ctx, session, model.PERMISSION_SCOPE_CC_QUEUE, strconv.FormatInt(queue.Id, 10), queue)

	return queue, nil
}

func (c *Controller) GetQueuesGlobalState(ctx context.Context, session *auth_manager.Session) (bool, model.AppError) {
	if !session.HasAdminPermission(auth_manager.PERMISSION_ACCESS_READ) {
		perm := session.GetPermission(model.PERMISSION_SCOPE_CC_QUEUE)
		return false, c.app.MakePermissionError(session, perm, auth_manager.PERMISSION_ACCESS_READ)
	}

	return c.app.GetQueuesGlobalState(ctx, session.Domain(0))
}

func (c *Controller) SetQueuesGlobalState(ctx context.Context, session *auth_manager.Session, newState bool) (int32, model.AppError) {
	if !session.HasAdminPermission(auth_manager.PERMISSION_ACCESS_UPDATE) {
		perm := session.GetPermission(model.PERMISSION_SCOPE_CC_QUEUE)
		return -1, c.app.MakePermissionError(session, perm, auth_manager.PERMISSION_ACCESS_UPDATE)
	}

	updatedBy := &model.Lookup{
		Id: int(session.UserId),
	}

	return c.app.SetQueuesGlobalState(ctx, session.Domain(0), newState, updatedBy)
}

func (c *Controller) DeleteQueue(ctx context.Context, session *auth_manager.Session, id int64) (*model.Queue, model.AppError) {
	var err model.AppError
	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_QUEUE)
	if !permission.CanDelete() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_DELETE)
	}

	if session.UseRBAC(auth_manager.PERMISSION_ACCESS_DELETE, permission) {
		var perm bool
		if perm, err = c.app.QueueCheckAccess(ctx, session.Domain(0), id, session.GetAclRoles(),
			auth_manager.PERMISSION_ACCESS_DELETE); err != nil {
			return nil, err
		} else if !perm {
			return nil, c.app.MakeResourcePermissionError(session, id, permission, auth_manager.PERMISSION_ACCESS_DELETE)
		}
	}

	var queue *model.Queue
	queue, err = c.app.RemoveQueue(ctx, session.Domain(0), id)

	if err != nil {
		return nil, err
	}

	c.app.AuditDelete(ctx, session, model.PERMISSION_SCOPE_CC_QUEUE, strconv.FormatInt(queue.Id, 10), queue)

	return queue, nil
}

func (c *Controller) QueueReportGeneral(ctx context.Context, session *auth_manager.Session, search *model.SearchQueueReportGeneral) (*model.QueueReportGeneralAgg, bool, model.AppError) {
	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_QUEUE)
	if !permission.CanRead() {
		return nil, false, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	return c.app.GetQueueReportGeneral(ctx, session.Domain(0), session.UserId, session.RoleIds,
		auth_manager.PERMISSION_ACCESS_READ, search)

}

func (c *Controller) SearchQueueTags(ctx context.Context, session *auth_manager.Session, search *model.ListRequest) ([]*model.Tag, bool, model.AppError) {
	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_QUEUE)
	if !permission.CanRead() {
		return nil, false, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	return c.app.SearchQueueTags(ctx, session.Domain(search.DomainId), search)
}
