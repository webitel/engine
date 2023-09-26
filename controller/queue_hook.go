package controller

import (
	"context"

	"github.com/webitel/engine/auth_manager"
	"github.com/webitel/engine/model"
)

func (c *Controller) SearchQueueHook(ctx context.Context, session *auth_manager.Session, queueId uint32, search *model.SearchQueueHook) ([]*model.QueueHook, bool, model.AppError) {
	var err model.AppError
	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_QUEUE)
	if !permission.CanRead() {
		return nil, false, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if session.UseRBAC(auth_manager.PERMISSION_ACCESS_READ, permission) {
		var perm bool
		if perm, err = c.app.QueueCheckAccess(ctx, session.Domain(0), int64(queueId), session.GetAclRoles(), auth_manager.PERMISSION_ACCESS_READ); err != nil {
			return nil, false, err
		} else if !perm {
			return nil, false, c.app.MakeResourcePermissionError(session, int64(queueId), permission, auth_manager.PERMISSION_ACCESS_READ)
		}
	}

	return c.app.SearchQueueHook(ctx, session.Domain(0), queueId, search)
}

func (c *Controller) CreateQueueHook(ctx context.Context, session *auth_manager.Session, queueId uint32, hook *model.QueueHook) (*model.QueueHook, model.AppError) {
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
		if perm, err = c.app.QueueCheckAccess(ctx, session.Domain(0), int64(queueId), session.GetAclRoles(), auth_manager.PERMISSION_ACCESS_UPDATE); err != nil {
			return nil, err
		} else if !perm {
			return nil, c.app.MakeResourcePermissionError(session, int64(queueId), permission, auth_manager.PERMISSION_ACCESS_UPDATE)
		}
	}
	hook.CreatedBy = &model.Lookup{
		Id: int(session.UserId),
	}
	hook.UpdatedBy = hook.CreatedBy

	hook.CreatedAt = *model.GetTime()
	hook.UpdatedAt = hook.CreatedAt

	if err := hook.IsValid(); err != nil {
		return nil, err
	}

	hook, err = c.app.CreateQueueHook(ctx, session.Domain(0), queueId, hook)
	if err != nil {
		return nil, err
	}

	// todo
	c.app.AuditUpdate(ctx, session, model.PERMISSION_SCOPE_CC_QUEUE, int64(queueId), hook)

	return hook, nil
}

func (c *Controller) GetQueueHook(ctx context.Context, session *auth_manager.Session, queueId, id uint32) (*model.QueueHook, model.AppError) {
	var err model.AppError
	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_QUEUE)
	if !permission.CanRead() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if session.UseRBAC(auth_manager.PERMISSION_ACCESS_READ, permission) {
		var perm bool
		if perm, err = c.app.QueueCheckAccess(ctx, session.Domain(0), int64(queueId), session.GetAclRoles(), auth_manager.PERMISSION_ACCESS_READ); err != nil {
			return nil, err
		} else if !perm {
			return nil, c.app.MakeResourcePermissionError(session, int64(queueId), permission, auth_manager.PERMISSION_ACCESS_READ)
		}
	}

	return c.app.GetQueueHook(ctx, session.Domain(0), queueId, id)
}

func (c *Controller) UpdateQueueHook(ctx context.Context, session *auth_manager.Session, queueId uint32, hook *model.QueueHook) (*model.QueueHook, model.AppError) {
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
		if perm, err = c.app.QueueCheckAccess(ctx, session.Domain(0), int64(queueId), session.GetAclRoles(), auth_manager.PERMISSION_ACCESS_UPDATE); err != nil {
			return nil, err
		} else if !perm {
			return nil, c.app.MakeResourcePermissionError(session, int64(queueId), permission, auth_manager.PERMISSION_ACCESS_UPDATE)
		}
	}

	hook.UpdatedBy = &model.Lookup{
		Id: int(session.UserId),
	}
	hook.UpdatedAt = *model.GetTime()

	if err := hook.IsValid(); err != nil {
		return nil, err
	}

	hook, err = c.app.UpdateQueueHook(ctx, session.DomainId, queueId, hook)
	if err != nil {
		return nil, err
	}

	// todo
	c.app.AuditUpdate(ctx, session, model.PERMISSION_SCOPE_CC_QUEUE, int64(queueId), hook)

	return hook, nil
}

func (c *Controller) PatchQueueHook(ctx context.Context, session *auth_manager.Session, queueId, id uint32, patch *model.QueueHookPatch) (*model.QueueHook, model.AppError) {
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
		if perm, err = c.app.QueueCheckAccess(ctx, session.Domain(0), int64(queueId), session.GetAclRoles(), auth_manager.PERMISSION_ACCESS_UPDATE); err != nil {
			return nil, err
		} else if !perm {
			return nil, c.app.MakeResourcePermissionError(session, int64(queueId), permission, auth_manager.PERMISSION_ACCESS_UPDATE)
		}
	}
	patch.UpdatedBy.Id = int(session.UserId)
	patch.UpdatedAt = *model.GetTime()

	var hook *model.QueueHook
	hook, err = c.app.PatchQueueHook(ctx, session.DomainId, queueId, id, patch)
	if err != nil {
		return nil, err
	}

	// todo
	c.app.AuditUpdate(ctx, session, model.PERMISSION_SCOPE_CC_QUEUE, int64(queueId), hook)

	return hook, nil
}

func (c *Controller) DeleteQueueHook(ctx context.Context, session *auth_manager.Session, queueId, id uint32) (*model.QueueHook, model.AppError) {
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
		if perm, err = c.app.QueueCheckAccess(ctx, session.Domain(0), int64(queueId), session.GetAclRoles(), auth_manager.PERMISSION_ACCESS_UPDATE); err != nil {
			return nil, err
		} else if !perm {
			return nil, c.app.MakeResourcePermissionError(session, int64(queueId), permission, auth_manager.PERMISSION_ACCESS_UPDATE)
		}
	}

	var hook *model.QueueHook

	hook, err = c.app.RemoveQueueHook(ctx, session.DomainId, queueId, id)
	if err != nil {
		return nil, err
	}

	// todo
	c.app.AuditUpdate(ctx, session, model.PERMISSION_SCOPE_CC_QUEUE, int64(queueId), hook)

	return hook, nil
}
