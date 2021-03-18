package controller

import (
	"github.com/webitel/engine/auth_manager"
	"github.com/webitel/engine/model"
)

func (c *Controller) SearchQueueHook(session *auth_manager.Session, queueId uint32, search *model.SearchQueueHook) ([]*model.QueueHook, bool, *model.AppError) {
	var err *model.AppError
	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_QUEUE)
	if !permission.CanRead() {
		return nil, false, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if permission.Rbac {
		var perm bool
		if perm, err = c.app.QueueCheckAccess(session.Domain(0), int64(queueId), session.GetAclRoles(), auth_manager.PERMISSION_ACCESS_READ); err != nil {
			return nil, false, err
		} else if !perm {
			return nil, false, c.app.MakeResourcePermissionError(session, int64(queueId), permission, auth_manager.PERMISSION_ACCESS_READ)
		}
	}

	return c.app.SearchQueueHook(session.Domain(0), queueId, search)
}

func (c *Controller) CreateQueueHook(session *auth_manager.Session, queueId uint32, hook *model.QueueHook) (*model.QueueHook, *model.AppError) {
	var err *model.AppError
	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_QUEUE)
	if !permission.CanRead() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if !permission.CanUpdate() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
	}

	if permission.Rbac {
		var perm bool
		if perm, err = c.app.QueueCheckAccess(session.Domain(0), int64(queueId), session.GetAclRoles(), auth_manager.PERMISSION_ACCESS_UPDATE); err != nil {
			return nil, err
		} else if !perm {
			return nil, c.app.MakeResourcePermissionError(session, int64(queueId), permission, auth_manager.PERMISSION_ACCESS_UPDATE)
		}
	}
	hook.CreatedBy.Id = int(session.UserId)
	hook.UpdatedBy = hook.CreatedBy

	hook.CreatedAt = *model.GetTime()
	hook.UpdatedAt = hook.CreatedAt

	if err := hook.IsValid(); err != nil {
		return nil, err
	}

	return c.app.CreateQueueHook(session.Domain(0), queueId, hook)
}

func (c *Controller) GetQueueHook(session *auth_manager.Session, queueId, id uint32) (*model.QueueHook, *model.AppError) {
	var err *model.AppError
	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_QUEUE)
	if !permission.CanRead() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if permission.Rbac {
		var perm bool
		if perm, err = c.app.QueueCheckAccess(session.Domain(0), int64(queueId), session.GetAclRoles(), auth_manager.PERMISSION_ACCESS_READ); err != nil {
			return nil, err
		} else if !perm {
			return nil, c.app.MakeResourcePermissionError(session, int64(queueId), permission, auth_manager.PERMISSION_ACCESS_READ)
		}
	}

	return c.app.GetQueueHook(session.Domain(0), queueId, id)
}

func (c *Controller) UpdateQueueHook(session *auth_manager.Session, queueId uint32, hook *model.QueueHook) (*model.QueueHook, *model.AppError) {
	var err *model.AppError
	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_QUEUE)
	if !permission.CanRead() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if !permission.CanUpdate() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
	}

	if permission.Rbac {
		var perm bool
		if perm, err = c.app.QueueCheckAccess(session.Domain(0), int64(queueId), session.GetAclRoles(), auth_manager.PERMISSION_ACCESS_UPDATE); err != nil {
			return nil, err
		} else if !perm {
			return nil, c.app.MakeResourcePermissionError(session, int64(queueId), permission, auth_manager.PERMISSION_ACCESS_UPDATE)
		}
	}

	hook.UpdatedBy.Id = int(session.UserId)
	hook.UpdatedAt = *model.GetTime()

	if err := hook.IsValid(); err != nil {
		return nil, err
	}

	return c.app.UpdateQueueHook(session.DomainId, queueId, hook)
}

func (c *Controller) PatchQueueHook(session *auth_manager.Session, queueId, id uint32, patch *model.QueueHookPatch) (*model.QueueHook, *model.AppError) {
	var err *model.AppError
	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_QUEUE)
	if !permission.CanRead() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if !permission.CanUpdate() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
	}

	if permission.Rbac {
		var perm bool
		if perm, err = c.app.QueueCheckAccess(session.Domain(0), int64(queueId), session.GetAclRoles(), auth_manager.PERMISSION_ACCESS_UPDATE); err != nil {
			return nil, err
		} else if !perm {
			return nil, c.app.MakeResourcePermissionError(session, int64(queueId), permission, auth_manager.PERMISSION_ACCESS_UPDATE)
		}
	}
	patch.UpdatedBy.Id = int(session.UserId)
	patch.UpdatedAt = *model.GetTime()

	return c.app.PatchQueueHook(session.DomainId, queueId, id, patch)
}

func (c *Controller) DeleteQueueHook(session *auth_manager.Session, queueId, id uint32) (*model.QueueHook, *model.AppError) {
	var err *model.AppError
	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_QUEUE)
	if !permission.CanRead() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if !permission.CanUpdate() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
	}

	if permission.Rbac {
		var perm bool
		if perm, err = c.app.QueueCheckAccess(session.Domain(0), int64(queueId), session.GetAclRoles(), auth_manager.PERMISSION_ACCESS_UPDATE); err != nil {
			return nil, err
		} else if !perm {
			return nil, c.app.MakeResourcePermissionError(session, int64(queueId), permission, auth_manager.PERMISSION_ACCESS_UPDATE)
		}
	}

	return c.app.RemoveQueueHook(session.DomainId, queueId, id)
}
