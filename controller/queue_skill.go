package controller

import "C"
import (
	"github.com/webitel/engine/auth_manager"
	"github.com/webitel/engine/model"
)

func (c *Controller) SearchQueueSkill(session *auth_manager.Session, search *model.SearchQueueSkill) ([]*model.QueueSkill, bool, *model.AppError) {
	var err *model.AppError
	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_QUEUE)
	if !permission.CanRead() {
		return nil, false, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if permission.Rbac {
		var perm bool
		if perm, err = c.app.QueueCheckAccess(session.Domain(0), int64(search.QueueId), session.GetAclRoles(), auth_manager.PERMISSION_ACCESS_READ); err != nil {
			return nil, false, err
		} else if !perm {
			return nil, false, c.app.MakeResourcePermissionError(session, int64(search.QueueId), permission, auth_manager.PERMISSION_ACCESS_READ)
		}
	}

	return c.app.SearchQueueSkill(session.Domain(0), search)
}

func (c *Controller) CreateQueueSkill(session *auth_manager.Session, qs *model.QueueSkill) (*model.QueueSkill, *model.AppError) {
	var err *model.AppError
	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_QUEUE)
	if !permission.CanRead() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if !permission.CanDelete() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_DELETE)
	}

	if permission.Rbac {
		var perm bool
		if perm, err = c.app.QueueCheckAccess(session.Domain(0), int64(qs.QueueId), session.GetAclRoles(), auth_manager.PERMISSION_ACCESS_CREATE); err != nil {
			return nil, err
		} else if !perm {
			return nil, c.app.MakeResourcePermissionError(session, int64(qs.QueueId), permission, auth_manager.PERMISSION_ACCESS_CREATE)
		}
	}

	if err := qs.IsValid(); err != nil {
		return nil, err
	}

	return c.app.CreateQueueSkill(session.Domain(0), qs)
}

func (c *Controller) GetQueueSkill(session *auth_manager.Session, queueId, id uint32) (*model.QueueSkill, *model.AppError) {
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

	return c.app.GetQueueSkill(session.Domain(0), queueId, id)
}

func (c *Controller) UpdateQueueSkill(session *auth_manager.Session, qs *model.QueueSkill) (*model.QueueSkill, *model.AppError) {
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
		if perm, err = c.app.QueueCheckAccess(session.Domain(0), int64(qs.QueueId), session.GetAclRoles(), auth_manager.PERMISSION_ACCESS_UPDATE); err != nil {
			return nil, err
		} else if !perm {
			return nil, c.app.MakeResourcePermissionError(session, int64(qs.QueueId), permission, auth_manager.PERMISSION_ACCESS_UPDATE)
		}
	}

	if err := qs.IsValid(); err != nil {
		return nil, err
	}

	return c.app.UpdateQueueSkill(session.DomainId, qs)
}

func (c *Controller) PatchQueueSkill(session *auth_manager.Session, queueId, id uint32, patch *model.QueueSkillPatch) (*model.QueueSkill, *model.AppError) {
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

	return c.app.PatchQueueSkill(session.DomainId, queueId, id, patch)
}

func (c *Controller) DeleteQueueSkill(session *auth_manager.Session, queueId, id uint32) (*model.QueueSkill, *model.AppError) {
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

	return c.app.RemoveQueueSkill(session.DomainId, queueId, id)
}
