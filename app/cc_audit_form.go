package app

import (
	"context"
	"fmt"

	"github.com/webitel/engine/auth_manager"
	"github.com/webitel/engine/model"
	"golang.org/x/sync/singleflight"
)

var (
	formGroupRequest singleflight.Group
)

func (app *App) AuditFormCheckAccess(ctx context.Context, domainId int64, id int32, groups []int, access auth_manager.PermissionAccess) (bool, model.AppError) {
	return app.Store.AuditForm().CheckAccess(ctx, domainId, id, groups, access)
}

func (app *App) CreateAuditForm(ctx context.Context, domainId int64, form *model.AuditForm) (*model.AuditForm, model.AppError) {
	return app.Store.AuditForm().Create(ctx, domainId, form)
}

func (app *App) GetAuditFormPage(ctx context.Context, domainId int64, search *model.SearchAuditForm) ([]*model.AuditForm, bool, model.AppError) {
	list, err := app.Store.AuditForm().GetAllPage(ctx, domainId, search)
	if err != nil {
		return nil, false, err
	}
	search.RemoveLastElemIfNeed(&list)
	return list, search.EndOfList(), nil
}

func (app *App) GetAuditFormPageByGroups(ctx context.Context, domainId int64, groups []int, search *model.SearchAuditForm) ([]*model.AuditForm, bool, model.AppError) {
	list, err := app.Store.AuditForm().GetAllPageByGroup(ctx, domainId, groups, search)
	if err != nil {
		return nil, false, err
	}
	search.RemoveLastElemIfNeed(&list)
	return list, search.EndOfList(), nil
}

func (app *App) GetAuditForm(ctx context.Context, domainId int64, id int32) (*model.AuditForm, model.AppError) {
	v, err, _ := formGroupRequest.Do(fmt.Sprintf("%d-%d", domainId, id), func() (interface{}, error) {
		res, err := app.Store.AuditForm().Get(ctx, domainId, id)
		if err != nil {
			return nil, err
		}

		return res, nil
	})

	if err != nil {
		switch err.(type) {
		case model.AppError:
			return nil, err.(model.AppError)
		default:
			return nil, model.NewInternalError("app.audit_form.get", err.Error())
		}
	}

	return v.(*model.AuditForm), nil
}

func (app *App) UpdateAuditForm(ctx context.Context, domainId int64, form *model.AuditForm) (*model.AuditForm, model.AppError) {
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

func (app *App) PatchAuditForm(ctx context.Context, domainId int64, id int32, patch *model.AuditFormPatch) (*model.AuditForm, model.AppError) {
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

func (app *App) RemoveAuditForm(ctx context.Context, domainId int64, id int32) (*model.AuditForm, model.AppError) {
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

func (app *App) RateAuditForm(ctx context.Context, domainId int64, userId int64, rate model.Rate) (*model.AuditRate, model.AppError) {
	if rate.CallId == nil {
		return nil, model.NewBadRequestError("app.audit.rate.valid.call_id", "call_id is required")
	}

	rateUserId, callCreatedAt, err := app.Store.Call().GetOwnerUserCall(ctx, *rate.CallId)
	if err != nil {
		return nil, err
	}

	var form *model.AuditForm
	form, err = app.GetAuditForm(ctx, domainId, int32(rate.Form.Id))
	if err != nil {
		return nil, err
	}

	if !form.Enabled {
		return nil, model.NewBadRequestError("app.audit.rate.valid.form", "form is disabled")
	}

	if form.Archive {
		return nil, model.NewBadRequestError("app.audit.rate.valid.form", "form is archive")
	}

	if rateUserId != nil {
		rate.RatedUser = &model.Lookup{Id: int(*rateUserId)}
	}

	rate.CallCreatedAt = &callCreatedAt

	auditRate := &model.AuditRate{
		AclRecord: model.AclRecord{
			CreatedAt: model.GetTime(),
			CreatedBy: &model.Lookup{
				Id: int(userId),
			},
		},
	}
	auditRate.UpdatedBy = auditRate.CreatedBy
	auditRate.UpdatedAt = auditRate.CreatedAt
	err = auditRate.SetRate(form, rate)
	if err != nil {
		return nil, err
	}

	if err = auditRate.IsValid(); err != nil {
		return nil, err
	}

	auditRate, err = app.Store.AuditRate().Create(ctx, domainId, auditRate)
	if err != nil {
		return nil, err
	}

	if form.Editable {
		err = app.Store.AuditForm().SetEditable(ctx, form.Id, false)
		if err != nil {
			return nil, err
		}
	}

	return auditRate, nil
}

func (app *App) GetAuditRate(ctx context.Context, domainId int64, id int64) (*model.AuditRate, model.AppError) {
	return app.Store.AuditRate().Get(ctx, domainId, id)
}

func (app *App) GetAuditRatePage(ctx context.Context, domainId int64, search *model.SearchAuditRate) ([]*model.AuditRate, bool, model.AppError) {
	list, err := app.Store.AuditRate().GetAllPage(ctx, domainId, search)
	if err != nil {
		return nil, false, err
	}
	search.RemoveLastElemIfNeed(&list)
	return list, search.EndOfList(), nil
}

func (app *App) GetAuditRateFormId(ctx context.Context, domainId, id int64) (int32, model.AppError) {
	return app.Store.AuditRate().FormId(ctx, domainId, id)
}
