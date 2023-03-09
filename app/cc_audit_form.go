package app

import (
	"context"
	"github.com/webitel/engine/auth_manager"
	"github.com/webitel/engine/model"
)

func (app *App) AuditFormCheckAccess(ctx context.Context, domainId int64, id int32, groups []int, access auth_manager.PermissionAccess) (bool, *model.AppError) {
	return app.Store.AuditForm().CheckAccess(ctx, domainId, id, groups, access)
}

func (app *App) CreateAuditForm(ctx context.Context, domainId int64, form *model.AuditForm) (*model.AuditForm, *model.AppError) {
	return app.Store.AuditForm().Create(ctx, domainId, form)
}

func (app *App) GetAuditFormPage(ctx context.Context, domainId int64, search *model.SearchAuditForm) ([]*model.AuditForm, bool, *model.AppError) {
	list, err := app.Store.AuditForm().GetAllPage(ctx, domainId, search)
	if err != nil {
		return nil, false, err
	}
	search.RemoveLastElemIfNeed(&list)
	return list, search.EndOfList(), nil
}

func (app *App) GetAuditFormPageByGroups(ctx context.Context, domainId int64, groups []int, search *model.SearchAuditForm) ([]*model.AuditForm, bool, *model.AppError) {
	list, err := app.Store.AuditForm().GetAllPageByGroup(ctx, domainId, groups, search)
	if err != nil {
		return nil, false, err
	}
	search.RemoveLastElemIfNeed(&list)
	return list, search.EndOfList(), nil
}

func (app *App) GetAuditForm(ctx context.Context, domainId int64, id int32) (*model.AuditForm, *model.AppError) {
	return app.Store.AuditForm().Get(ctx, domainId, id)
}

func (app *App) UpdateAuditForm(ctx context.Context, domainId int64, form *model.AuditForm) (*model.AuditForm, *model.AppError) {
	oldForm, err := app.GetAuditForm(ctx, domainId, form.Id)
	if err != nil {
		return nil, err
	}

	oldForm.Name = form.Name
	oldForm.Description = form.Description
	oldForm.Enabled = form.Enabled
	oldForm.Questions = form.Questions
	oldForm.Teams = form.Teams
	oldForm.UpdatedBy = form.UpdatedBy
	oldForm.UpdatedAt = form.UpdatedAt

	if err = oldForm.IsValid(); err != nil {
		return nil, err
	}

	oldForm, err = app.Store.AuditForm().Update(ctx, domainId, oldForm)
	if err != nil {
		return nil, err
	}

	return oldForm, nil
}

func (app *App) PatchAuditForm(ctx context.Context, domainId int64, id int32, patch *model.AuditFormPatch) (*model.AuditForm, *model.AppError) {
	oldForm, err := app.GetAuditForm(ctx, domainId, id)
	if err != nil {
		return nil, err
	}

	oldForm.Patch(patch)

	if err = oldForm.IsValid(); err != nil {
		return nil, err
	}

	oldForm, err = app.Store.AuditForm().Update(ctx, domainId, oldForm)
	if err != nil {
		return nil, err
	}

	return oldForm, nil
}

func (app *App) RemoveAuditForm(ctx context.Context, domainId int64, id int32) (*model.AuditForm, *model.AppError) {
	form, err := app.Store.AuditForm().Get(ctx, domainId, id)

	if err != nil {
		return nil, err
	}

	err = app.Store.AuditForm().Delete(ctx, domainId, id)
	if err != nil {
		return nil, err
	}
	return form, nil
}
