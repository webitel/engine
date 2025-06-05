package controller

import (
	"context"
	"github.com/webitel/engine/model"
	"github.com/webitel/engine/pkg/wbt/auth_manager"
	"strconv"
)

func (c *Controller) SearchQueueSkill(ctx context.Context, session *auth_manager.Session, search *model.SearchQueueSkill) ([]*model.QueueSkill, bool, model.AppError) {
	var err model.AppError
	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_QUEUE)
	if !permission.CanRead() {
		return nil, false, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if session.UseRBAC(auth_manager.PERMISSION_ACCESS_READ, permission) {
		var perm bool
		if perm, err = c.app.QueueCheckAccess(ctx, session.Domain(0), int64(search.QueueId), session.GetAclRoles(), auth_manager.PERMISSION_ACCESS_READ); err != nil {
			return nil, false, err
		} else if !perm {
			return nil, false, c.app.MakeResourcePermissionError(session, int64(search.QueueId), permission, auth_manager.PERMISSION_ACCESS_READ)
		}
	}

	return c.app.SearchQueueSkill(ctx, session.Domain(0), search)
}

func (c *Controller) CreateQueueSkill(ctx context.Context, session *auth_manager.Session, qs *model.QueueSkill) (*model.QueueSkill, model.AppError) {
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
		if perm, err = c.app.QueueCheckAccess(ctx, session.Domain(0), int64(qs.QueueId), session.GetAclRoles(), auth_manager.PERMISSION_ACCESS_UPDATE); err != nil {
			return nil, err
		} else if !perm {
			return nil, c.app.MakeResourcePermissionError(session, int64(qs.QueueId), permission, auth_manager.PERMISSION_ACCESS_CREATE)
		}
	}

	if err := qs.IsValid(); err != nil {
		return nil, err
	}

	qs, err = c.app.CreateQueueSkill(ctx, session.Domain(0), qs)
	if err != nil {
		return nil, err
	}

	c.app.AuditUpdate(ctx, session, model.PERMISSION_SCOPE_CC_QUEUE, strconv.Itoa(int(qs.QueueId)), qs)

	return qs, nil
}

func (c *Controller) GetQueueSkill(ctx context.Context, session *auth_manager.Session, queueId, id uint32) (*model.QueueSkill, model.AppError) {
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

	return c.app.GetQueueSkill(ctx, session.Domain(0), queueId, id)
}

func (c *Controller) UpdateQueueSkill(ctx context.Context, session *auth_manager.Session, qs *model.QueueSkill) (*model.QueueSkill, model.AppError) {
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
		if perm, err = c.app.QueueCheckAccess(ctx, session.Domain(0), int64(qs.QueueId), session.GetAclRoles(), auth_manager.PERMISSION_ACCESS_UPDATE); err != nil {
			return nil, err
		} else if !perm {
			return nil, c.app.MakeResourcePermissionError(session, int64(qs.QueueId), permission, auth_manager.PERMISSION_ACCESS_UPDATE)
		}
	}

	if err := qs.IsValid(); err != nil {
		return nil, err
	}

	qs, err = c.app.UpdateQueueSkill(ctx, session.DomainId, qs)
	if err != nil {
		return nil, err
	}

	c.app.AuditUpdate(ctx, session, model.PERMISSION_SCOPE_CC_QUEUE, strconv.Itoa(int(qs.QueueId)), qs)

	return qs, nil
}

func (c *Controller) PatchQueueSkill(ctx context.Context, session *auth_manager.Session, queueId, id uint32, patch *model.QueueSkillPatch) (*model.QueueSkill, model.AppError) {
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

	var qs *model.QueueSkill

	qs, err = c.app.PatchQueueSkill(ctx, session.DomainId, queueId, id, patch)
	if err != nil {
		return nil, err
	}

	c.app.AuditUpdate(ctx, session, model.PERMISSION_SCOPE_CC_QUEUE, strconv.Itoa(int(queueId)), qs)

	return qs, nil
}

func (c *Controller) DeleteQueueSkill(ctx context.Context, session *auth_manager.Session, queueId, id uint32) (*model.QueueSkill, model.AppError) {
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

	var qs *model.QueueSkill

	qs, err = c.app.RemoveQueueSkill(ctx, session.DomainId, queueId, id)
	if err != nil {
		return nil, err
	}

	c.app.AuditUpdate(ctx, session, model.PERMISSION_SCOPE_CC_QUEUE, strconv.Itoa(int(queueId)), qs)

	return qs, nil
}
