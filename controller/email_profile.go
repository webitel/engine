package controller

import (
	"context"

	"github.com/webitel/engine/auth_manager"
	"github.com/webitel/engine/model"
)

func (c *Controller) CreateEmailProfile(ctx context.Context, session *auth_manager.Session, profile *model.EmailProfile) (*model.EmailProfile, model.AppError) {
	permission := session.GetPermission(model.PERMISSION_SCOPE_EMAIL_PROFILE)
	if !permission.CanCreate() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_CREATE)
	}

	profile.CreatedBy = &model.Lookup{
		Id: int(session.UserId),
	}
	profile.UpdatedBy = profile.CreatedBy
	profile.DomainId = session.Domain(0)

	if err := profile.IsValid(); err != nil {
		return nil, err
	}

	if profile.Enabled {
		if err := c.app.ConstraintEmailProfileLimit(ctx, session.Domain(0), session.Token); err != nil {
			return nil, err
		}
	}

	return c.app.CreateEmailProfile(ctx, session.Domain(0), profile)
}

func (c *Controller) SearchEmailProfile(ctx context.Context, session *auth_manager.Session, search *model.SearchEmailProfile) ([]*model.EmailProfile, bool, model.AppError) {
	permission := session.GetPermission(model.PERMISSION_SCOPE_EMAIL_PROFILE)
	if !permission.CanRead() {
		return nil, false, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	search.DomainId = session.Domain(search.DomainId) //TODO

	return c.app.GetEmailProfilesPage(ctx, session.Domain(search.DomainId), search)
}

func (c *Controller) GetEmailProfile(ctx context.Context, session *auth_manager.Session, id int) (*model.EmailProfile, model.AppError) {
	permission := session.GetPermission(model.PERMISSION_SCOPE_EMAIL_PROFILE)
	if !permission.CanRead() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	return c.app.GetEmailProfile(ctx, session.Domain(0), id)
}

func (c *Controller) UpdateEmailProfile(ctx context.Context, session *auth_manager.Session, profile *model.EmailProfile) (*model.EmailProfile, model.AppError) {
	permission := session.GetPermission(model.PERMISSION_SCOPE_EMAIL_PROFILE)
	if !permission.CanRead() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if !permission.CanUpdate() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
	}

	profile.DomainId = session.Domain(0)

	if err := profile.IsValid(); err != nil {
		return nil, err
	}

	if profile.Enabled {
		if err := c.app.ConstraintEmailProfileLimit(ctx, session.Domain(0), session.Token); err != nil {
			return nil, err
		}
	}

	return c.app.UpdateEmailProfile(ctx, session.Domain(0), profile)
}

func (c *Controller) PatchEmailProfile(ctx context.Context, session *auth_manager.Session, id int, patch *model.EmailProfilePatch) (*model.EmailProfile, model.AppError) {
	permission := session.GetPermission(model.PERMISSION_SCOPE_EMAIL_PROFILE)
	if !permission.CanRead() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if !permission.CanUpdate() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
	}

	patch.UpdatedBy.Id = int(session.UserId)
	patch.UpdatedAt = model.GetMillis()

	if patch.Enabled != nil && *patch.Enabled {
		if err := c.app.ConstraintEmailProfileLimit(ctx, session.Domain(0), session.Token); err != nil {
			return nil, err
		}
	}

	return c.app.PatchEmailProfile(ctx, session.Domain(0), id, patch)
}

func (c *Controller) RemoveEmailProfile(ctx context.Context, session *auth_manager.Session, id int) (*model.EmailProfile, model.AppError) {
	permission := session.GetPermission(model.PERMISSION_SCOPE_EMAIL_PROFILE)
	if !permission.CanRead() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if !permission.CanDelete() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_DELETE)
	}

	return c.app.RemoveEmailProfile(ctx, session.Domain(0), id)
}

func (c *Controller) LoginEmailProfile(ctx context.Context, session *auth_manager.Session, id int) (*model.EmailProfileLogin, model.AppError) {
	permission := session.GetPermission(model.PERMISSION_SCOPE_EMAIL_PROFILE)
	if !permission.CanRead() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if !permission.CanUpdate() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
	}

	return c.app.LoginEmailProfile(ctx, session.Domain(0), id)
}
