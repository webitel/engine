package controller

import (
	"context"

	"github.com/webitel/engine/auth_manager"
	"github.com/webitel/engine/model"
)

func (c *Controller) CreateAuditForm(ctx context.Context, session *auth_manager.Session, form *model.AuditForm) (*model.AuditForm, model.AppError) {
	var err model.AppError
	permission := session.GetPermission(model.PermissionAuditFrom)
	if !permission.CanRead() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if !permission.CanUpdate() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
	}

	form.CreatedBy = &model.Lookup{
		Id: int(session.UserId),
	}
	form.UpdatedBy = form.CreatedBy

	form.CreatedAt = model.GetTime()
	form.UpdatedAt = form.CreatedAt

	if err = form.IsValid(); err != nil {
		return nil, err
	}

	return c.app.CreateAuditForm(ctx, session.Domain(0), form)
}

func (c *Controller) SearchAuditForm(ctx context.Context, session *auth_manager.Session, search *model.SearchAuditForm) ([]*model.AuditForm, bool, model.AppError) {
	permission := session.GetPermission(model.PermissionAuditFrom)
	if !permission.CanRead() {
		return nil, true, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if session.UseRBAC(auth_manager.PERMISSION_ACCESS_READ, permission) {
		return c.app.GetAuditFormPageByGroups(ctx, session.Domain(0), session.GetAclRoles(), search)
	} else {
		return c.app.GetAuditFormPage(ctx, session.Domain(0), search)
	}
}

func (c *Controller) ReadAuditForm(ctx context.Context, session *auth_manager.Session, id int32) (*model.AuditForm, model.AppError) {
	permission := session.GetPermission(model.PermissionAuditFrom)
	if !permission.CanRead() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if session.UseRBAC(auth_manager.PERMISSION_ACCESS_READ, permission) {
		var perm bool
		var err model.AppError

		if perm, err = c.app.AuditFormCheckAccess(ctx, session.Domain(0), id, session.GetAclRoles(),
			auth_manager.PERMISSION_ACCESS_READ); err != nil {
			return nil, err
		} else if !perm {
			return nil, c.app.MakeResourcePermissionError(session, int64(id), permission, auth_manager.PERMISSION_ACCESS_READ)
		}
	}

	return c.app.GetAuditForm(ctx, session.Domain(0), id)
}

func (c *Controller) PutAuditForm(ctx context.Context, session *auth_manager.Session, form *model.AuditForm) (*model.AuditForm, model.AppError) {
	var err model.AppError
	permission := session.GetPermission(model.PermissionAuditFrom)
	if !permission.CanRead() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if !permission.CanUpdate() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
	}

	if session.UseRBAC(auth_manager.PERMISSION_ACCESS_UPDATE, permission) {
		var perm bool

		if perm, err = c.app.AuditFormCheckAccess(ctx, session.Domain(0), form.Id, session.GetAclRoles(),
			auth_manager.PERMISSION_ACCESS_UPDATE); err != nil {
			return nil, err
		} else if !perm {
			return nil, c.app.MakeResourcePermissionError(session, int64(form.Id), permission, auth_manager.PERMISSION_ACCESS_UPDATE)
		}
	}

	form.UpdatedAt = model.GetTime()
	form.UpdatedBy = &model.Lookup{
		Id: int(session.UserId),
	}

	if err = form.IsValid(); err != nil {
		return nil, err
	}

	return c.app.UpdateAuditForm(ctx, session.Domain(0), form)
}

func (c *Controller) PatchAuditForm(ctx context.Context, session *auth_manager.Session, id int32, patch *model.AuditFormPatch) (*model.AuditForm, model.AppError) {
	var err model.AppError
	permission := session.GetPermission(model.PermissionAuditFrom)
	if !permission.CanRead() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if !permission.CanUpdate() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
	}

	if session.UseRBAC(auth_manager.PERMISSION_ACCESS_UPDATE, permission) {
		var perm bool

		if perm, err = c.app.AuditFormCheckAccess(ctx, session.Domain(0), id, session.GetAclRoles(),
			auth_manager.PERMISSION_ACCESS_UPDATE); err != nil {
			return nil, err
		} else if !perm {
			return nil, c.app.MakeResourcePermissionError(session, int64(id), permission, auth_manager.PERMISSION_ACCESS_UPDATE)
		}
	}

	patch.UpdatedAt = *model.GetTime()
	patch.UpdatedBy = model.Lookup{
		Id: int(session.UserId),
	}

	return c.app.PatchAuditForm(ctx, session.Domain(0), id, patch)
}

func (c *Controller) DeleteAuditForm(ctx context.Context, session *auth_manager.Session, id int32) (*model.AuditForm, model.AppError) {
	var err model.AppError
	permission := session.GetPermission(model.PermissionAuditFrom)
	if !permission.CanRead() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if !permission.CanDelete() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_DELETE)
	}

	if session.UseRBAC(auth_manager.PERMISSION_ACCESS_DELETE, permission) {
		var perm bool

		if perm, err = c.app.AuditFormCheckAccess(ctx, session.Domain(0), id, session.GetAclRoles(),
			auth_manager.PERMISSION_ACCESS_DELETE); err != nil {
			return nil, err
		} else if !perm {
			return nil, c.app.MakeResourcePermissionError(session, int64(id), permission, auth_manager.PERMISSION_ACCESS_DELETE)
		}
	}

	return c.app.RemoveAuditForm(ctx, session.Domain(0), id)
}

func (c *Controller) RateAuditForm(ctx context.Context, session *auth_manager.Session, rate model.Rate) (*model.AuditRate, model.AppError) {
	var err model.AppError
	permission := session.GetPermission(model.PermissionAuditFrom)
	if !permission.CanRead() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if session.UseRBAC(auth_manager.PERMISSION_ACCESS_READ, permission) {
		var perm bool

		if perm, err = c.app.AuditFormCheckAccess(ctx, session.Domain(0), int32(rate.Form.Id), session.GetAclRoles(),
			auth_manager.PERMISSION_ACCESS_DELETE); err != nil {
			return nil, err
		} else if !perm {
			return nil, c.app.MakeResourcePermissionError(session, int64(rate.Form.Id), permission, auth_manager.PERMISSION_ACCESS_READ)
		}
	}

	if session.HasAction(auth_manager.PermissionAuditRate) {
		return nil, c.app.MakeResourcePermissionError(session, int64(rate.Form.Id), permission, auth_manager.PERMISSION_ACCESS_CREATE)
	}

	return c.app.RateAuditForm(ctx, session.Domain(0), session.UserId, rate)
}

func (c *Controller) SearchAuditRate(ctx context.Context, session *auth_manager.Session, formId int32, search *model.SearchAuditRate) ([]*model.AuditRate, bool, model.AppError) {
	var err model.AppError
	permission := session.GetPermission(model.PermissionAuditFrom)
	if !permission.CanRead() {
		return nil, true, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if session.UseRBAC(auth_manager.PERMISSION_ACCESS_READ, permission) {
		var perm bool

		if perm, err = c.app.AuditFormCheckAccess(ctx, session.Domain(0), formId, session.GetAclRoles(),
			auth_manager.PERMISSION_ACCESS_DELETE); err != nil {
			return nil, true, err
		} else if !perm {
			return nil, true, c.app.MakeResourcePermissionError(session, int64(formId), permission, auth_manager.PERMISSION_ACCESS_READ)
		}
	}

	return c.app.GetAuditRatePage(ctx, session.Domain(0), search)
}

func (c *Controller) ReadAuditRate(ctx context.Context, session *auth_manager.Session, id int64) (*model.AuditRate, model.AppError) {
	var err model.AppError
	permission := session.GetPermission(model.PermissionAuditFrom)
	if !permission.CanRead() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if session.UseRBAC(auth_manager.PERMISSION_ACCESS_READ, permission) {
		var perm bool
		var formId int32
		formId, err = c.app.GetAuditRateFormId(ctx, session.Domain(0), id)

		if perm, err = c.app.AuditFormCheckAccess(ctx, session.Domain(0), formId, session.GetAclRoles(),
			auth_manager.PERMISSION_ACCESS_DELETE); err != nil {
			return nil, err
		} else if !perm {
			return nil, c.app.MakeResourcePermissionError(session, int64(formId), permission, auth_manager.PERMISSION_ACCESS_READ)
		}
	}

	return c.app.GetAuditRate(ctx, session.Domain(0), id)
}
