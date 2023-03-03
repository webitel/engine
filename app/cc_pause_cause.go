package app

import (
	"context"
	"github.com/webitel/engine/model"
)

func (a *App) GetPauseCausePage(ctx context.Context, domainId int64, search *model.SearchPauseCause) ([]*model.PauseCause, bool, *model.AppError) {
	list, err := a.Store.PauseCause().GetAllPage(ctx, domainId, search)
	if err != nil {
		return nil, false, err
	}
	search.RemoveLastElemIfNeed(&list)
	return list, search.EndOfList(), nil
}

func (a *App) CreatePauseCause(ctx context.Context, domainId int64, cause *model.PauseCause) (*model.PauseCause, *model.AppError) {
	return a.Store.PauseCause().Create(ctx, domainId, cause)
}

func (a *App) GetPauseCause(ctx context.Context, domainId int64, id uint32) (*model.PauseCause, *model.AppError) {
	return a.Store.PauseCause().Get(ctx, domainId, id)
}

func (a *App) UpdatePauseCause(ctx context.Context, domainId int64, cause *model.PauseCause) (*model.PauseCause, *model.AppError) {
	oldCause, err := a.GetPauseCause(ctx, domainId, uint32(cause.Id))
	if err != nil {
		return nil, err
	}

	oldCause.UpdatedBy = cause.UpdatedBy
	oldCause.UpdatedAt = cause.UpdatedAt

	oldCause.Name = cause.Name
	oldCause.Description = cause.Description
	oldCause.AllowAgent = cause.AllowAgent
	oldCause.AllowSupervisor = cause.AllowSupervisor
	oldCause.AllowAdmin = cause.AllowAdmin
	oldCause.LimitMin = cause.LimitMin

	oldCause, err = a.Store.PauseCause().Update(ctx, domainId, oldCause)
	if err != nil {
		return nil, err
	}

	return oldCause, nil
}

func (a *App) PatchPauseCause(ctx context.Context, domainId int64, id uint32, patch *model.PauseCausePatch) (*model.PauseCause, *model.AppError) {
	oldCause, err := a.GetPauseCause(ctx, domainId, id)
	if err != nil {
		return nil, err
	}

	oldCause.Patch(patch)

	if err = oldCause.IsValid(); err != nil {
		return nil, err
	}

	oldCause, err = a.Store.PauseCause().Update(ctx, domainId, oldCause)
	if err != nil {
		return nil, err
	}

	return oldCause, nil
}

func (a *App) RemovePauseCause(ctx context.Context, domainId int64, id uint32) (*model.PauseCause, *model.AppError) {
	cause, err := a.GetPauseCause(ctx, domainId, id)

	if err != nil {
		return nil, err
	}

	err = a.Store.PauseCause().Delete(ctx, domainId, id)
	if err != nil {
		return nil, err
	}
	return cause, nil
}
