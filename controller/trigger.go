package controller

import (
	"github.com/webitel/engine/auth_manager"
	"github.com/webitel/engine/model"
)

func (c *Controller) CreateTrigger(session *auth_manager.Session, trigger *model.Trigger) (*model.Trigger, *model.AppError) {
	var err *model.AppError
	permission := session.GetPermission(model.PERMISSION_SCOPE_TRIGGER)
	if !permission.CanCreate() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_CREATE)
	}

	trigger.CreatedBy = &model.Lookup{
		Id: int(session.UserId),
	}
	trigger.UpdatedBy = trigger.CreatedBy
	trigger.CreatedAt = model.GetTime()
	trigger.UpdatedAt = trigger.CreatedAt

	if err = trigger.IsValid(); err != nil {
		return nil, err
	}

	return c.app.CreateTrigger(session.Domain(0), trigger)
}

func (c *Controller) SearchTrigger(session *auth_manager.Session, search *model.SearchTrigger) ([]*model.Trigger, bool, *model.AppError) {
	permission := session.GetPermission(model.PERMISSION_SCOPE_TRIGGER)
	if !permission.CanRead() {
		return nil, false, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if session.UseRBAC(auth_manager.PERMISSION_ACCESS_READ, permission) {
		return c.app.GetTriggerListByGroups(session.Domain(search.DomainId), session.GetAclRoles(), search)
	} else {
		return c.app.GetTriggerList(session.Domain(search.DomainId), search)
	}
}

func (c *Controller) ReadTrigger(session *auth_manager.Session, id int32) (*model.Trigger, *model.AppError) {
	var err *model.AppError
	permission := session.GetPermission(model.PERMISSION_SCOPE_TRIGGER)
	if !permission.CanRead() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if session.UseRBAC(auth_manager.PERMISSION_ACCESS_READ, permission) {
		var perm bool
		if perm, err = c.app.TriggerCheckAccess(session.Domain(0), id, session.GetAclRoles(), auth_manager.PERMISSION_ACCESS_READ); err != nil {
			return nil, err
		} else if !perm {
			return nil, c.app.MakeResourcePermissionError(session, int64(id), permission, auth_manager.PERMISSION_ACCESS_READ)
		}
	}

	return c.app.GetTrigger(session.Domain(0), id)
}

func (c *Controller) UpdateTrigger(session *auth_manager.Session, trigger *model.Trigger) (*model.Trigger, *model.AppError) {
	var err *model.AppError
	permission := session.GetPermission(model.PERMISSION_SCOPE_TRIGGER)
	if !permission.CanRead() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}
	if !permission.CanUpdate() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
	}

	if session.UseRBAC(auth_manager.PERMISSION_ACCESS_UPDATE, permission) {
		var perm bool
		if perm, err = c.app.TriggerCheckAccess(session.Domain(0), trigger.Id, session.GetAclRoles(), auth_manager.PERMISSION_ACCESS_UPDATE); err != nil {
			return nil, err
		} else if !perm {
			return nil, c.app.MakeResourcePermissionError(session, int64(trigger.Id), permission, auth_manager.PERMISSION_ACCESS_UPDATE)
		}
	}

	trigger.UpdatedBy = &model.Lookup{
		Id: int(session.UserId),
	}
	trigger.UpdatedAt = model.GetTime()

	return c.app.UpdateTrigger(session.Domain(0), trigger)
}

func (c *Controller) PatchTrigger(session *auth_manager.Session, id int32, patch *model.TriggerPatch) (*model.Trigger, *model.AppError) {
	var err *model.AppError
	permission := session.GetPermission(model.PERMISSION_SCOPE_TRIGGER)
	if !permission.CanRead() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}
	if !permission.CanUpdate() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
	}

	if session.UseRBAC(auth_manager.PERMISSION_ACCESS_UPDATE, permission) {
		var perm bool
		if perm, err = c.app.TriggerCheckAccess(session.Domain(0), id, session.GetAclRoles(), auth_manager.PERMISSION_ACCESS_UPDATE); err != nil {
			return nil, err
		} else if !perm {
			return nil, c.app.MakeResourcePermissionError(session, int64(id), permission, auth_manager.PERMISSION_ACCESS_UPDATE)
		}
	}

	patch.UpdatedBy = &model.Lookup{
		Id: int(session.UserId),
	}
	patch.UpdatedAt = model.GetTime()

	return c.app.PatchTrigger(session.Domain(0), id, patch)
}

func (c *Controller) RemoveTrigger(session *auth_manager.Session, id int32) (*model.Trigger, *model.AppError) {
	var err *model.AppError
	permission := session.GetPermission(model.PERMISSION_SCOPE_TRIGGER)
	if !permission.CanRead() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}
	if !permission.CanDelete() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_DELETE)
	}

	if session.UseRBAC(auth_manager.PERMISSION_ACCESS_DELETE, permission) {
		var perm bool
		if perm, err = c.app.TriggerCheckAccess(session.Domain(0), id, session.GetAclRoles(), auth_manager.PERMISSION_ACCESS_DELETE); err != nil {
			return nil, err
		} else if !perm {
			return nil, c.app.MakeResourcePermissionError(session, int64(id), permission, auth_manager.PERMISSION_ACCESS_DELETE)
		}
	}

	return c.app.RemoveTrigger(session.Domain(0), id)
}

func (c *Controller) GetTriggerJobList(session *auth_manager.Session, triggerId int32, search *model.SearchTriggerJob) ([]*model.TriggerJob, bool, *model.AppError) {
	var err *model.AppError
	permission := session.GetPermission(model.PERMISSION_SCOPE_TRIGGER)
	if !permission.CanRead() {
		return nil, false, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if session.UseRBAC(auth_manager.PERMISSION_ACCESS_READ, permission) {
		var perm bool
		if perm, err = c.app.TriggerCheckAccess(session.Domain(0), triggerId, session.GetAclRoles(), auth_manager.PERMISSION_ACCESS_READ); err != nil {
			return nil, false, err
		} else if !perm {
			return nil, false, c.app.MakeResourcePermissionError(session, int64(triggerId), permission, auth_manager.PERMISSION_ACCESS_READ)
		}
	}

	return c.app.GetTriggerJobList(session.Domain(0), triggerId, search)
}

func (c *Controller) CreateTriggerJob(session *auth_manager.Session, triggerId int32, vars map[string]string) (*model.TriggerJob, *model.AppError) {
	var err *model.AppError
	permission := session.GetPermission(model.PERMISSION_SCOPE_TRIGGER)
	if !permission.CanRead() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if session.UseRBAC(auth_manager.PERMISSION_ACCESS_READ, permission) {
		var perm bool
		if perm, err = c.app.TriggerCheckAccess(session.Domain(0), triggerId, session.GetAclRoles(), auth_manager.PERMISSION_ACCESS_READ); err != nil {
			return nil, err
		} else if !perm {
			return nil, c.app.MakeResourcePermissionError(session, int64(triggerId), permission, auth_manager.PERMISSION_ACCESS_READ)
		}
	}

	return c.app.CreateTriggerJob(session.Domain(0), triggerId, vars)
}
