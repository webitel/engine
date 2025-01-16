package app

import (
	"context"

	"github.com/webitel/engine/model"
)

func (a *App) GetQuickReplyPage(ctx context.Context, domainId int64, search *model.SearchQuickReply) ([]*model.QuickReply, bool, model.AppError) {
	list, err := a.Store.QuickReply().GetAllPage(ctx, domainId, search)
	if err != nil {
		return nil, false, err
	}
	search.RemoveLastElemIfNeed(&list)
	return list, search.EndOfList(), nil
}

func (a *App) CreateQuickReply(ctx context.Context, domainId int64, cause *model.QuickReply) (*model.QuickReply, model.AppError) {
	return a.Store.QuickReply().Create(ctx, domainId, cause)
}

func (a *App) GetQuickReply(ctx context.Context, domainId int64, id uint32) (*model.QuickReply, model.AppError) {
	return a.Store.QuickReply().Get(ctx, domainId, id)
}

func (a *App) UpdateQuickReply(ctx context.Context, domainId int64, cause *model.QuickReply) (*model.QuickReply, model.AppError) {
	oldCause, err := a.GetQuickReply(ctx, domainId, uint32(cause.Id))
	if err != nil {
		return nil, err
	}

	oldCause.UpdatedBy = cause.UpdatedBy
	oldCause.UpdatedAt = cause.UpdatedAt

	oldCause.Name = cause.Name
	oldCause.Text = cause.Text

	oldCause, err = a.Store.QuickReply().Update(ctx, domainId, oldCause)
	if err != nil {
		return nil, err
	}

	return oldCause, nil
}

func (a *App) PatchQuickReply(ctx context.Context, domainId int64, id uint32, patch *model.QuickReplyPatch) (*model.QuickReply, model.AppError) {
	oldCause, err := a.GetQuickReply(ctx, domainId, id)
	if err != nil {
		return nil, err
	}

	oldCause.Patch(patch)

	if err = oldCause.IsValid(); err != nil {
		return nil, err
	}

	oldCause, err = a.Store.QuickReply().Update(ctx, domainId, oldCause)
	if err != nil {
		return nil, err
	}

	return oldCause, nil
}

func (a *App) RemoveQuickReply(ctx context.Context, domainId int64, id uint32) (*model.QuickReply, model.AppError) {
	cause, err := a.GetQuickReply(ctx, domainId, id)

	if err != nil {
		return nil, err
	}

	err = a.Store.QuickReply().Delete(ctx, domainId, id)
	if err != nil {
		return nil, err
	}
	return cause, nil
}
