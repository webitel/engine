package app

import (
	"context"

	"github.com/webitel/engine/model"
)

func (a *App) GetQuickReplyPage(ctx context.Context, domainId int64, search *model.SearchQuickReply) ([]*model.QuickReply, bool, model.AppError) {
	_, session, _ := a.getSessionFromCtx(ctx)
	list, err := a.Store.QuickReply().GetAllPage(ctx, domainId, search, session.UserId)
	if err != nil {
		return nil, false, err
	}
	search.RemoveLastElemIfNeed(&list)
	return list, search.EndOfList(), nil
}

func (a *App) CreateQuickReply(ctx context.Context, domainId int64, reply *model.QuickReply) (*model.QuickReply, model.AppError) {
	return a.Store.QuickReply().Create(ctx, domainId, reply)
}

func (a *App) GetQuickReply(ctx context.Context, domainId int64, id uint32) (*model.QuickReply, model.AppError) {
	return a.Store.QuickReply().Get(ctx, domainId, id)
}

func (a *App) UpdateQuickReply(ctx context.Context, domainId int64, reply *model.QuickReply) (*model.QuickReply, model.AppError) {
	oldReply, err := a.GetQuickReply(ctx, domainId, uint32(reply.Id))
	if err != nil {
		return nil, err
	}

	oldReply.UpdatedBy = reply.UpdatedBy
	oldReply.UpdatedAt = reply.UpdatedAt

	oldReply.Name = reply.Name
	oldReply.Text = reply.Text
	oldReply.Queues = reply.Queues
	oldReply.Teams = reply.Teams
	oldReply.Article = reply.Article

	oldReply, err = a.Store.QuickReply().Update(ctx, domainId, oldReply)
	if err != nil {
		return nil, err
	}

	return oldReply, nil
}

func (a *App) PatchQuickReply(ctx context.Context, domainId int64, id uint32, patch *model.QuickReplyPatch) (*model.QuickReply, model.AppError) {
	oldReply, err := a.GetQuickReply(ctx, domainId, id)
	if err != nil {
		return nil, err
	}

	oldReply.Patch(patch)

	if err = oldReply.IsValid(); err != nil {
		return nil, err
	}

	oldReply, err = a.Store.QuickReply().Update(ctx, domainId, oldReply)
	if err != nil {
		return nil, err
	}

	return oldReply, nil
}

func (a *App) RemoveQuickReply(ctx context.Context, domainId int64, id uint32) (*model.QuickReply, model.AppError) {
	reply, err := a.GetQuickReply(ctx, domainId, id)

	if err != nil {
		return nil, err
	}

	err = a.Store.QuickReply().Delete(ctx, domainId, id)
	if err != nil {
		return nil, err
	}
	return reply, nil
}
