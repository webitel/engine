package app

import (
	"context"
	"github.com/webitel/engine/model"
)

func (app *App) CreateQueueResourceGroup(ctx context.Context, queueResourceGroup *model.QueueResourceGroup) (*model.QueueResourceGroup, *model.AppError) {
	return app.Store.QueueResource().Create(ctx, queueResourceGroup)
}

func (app *App) GetQueueResourceGroupPage(ctx context.Context, domainId, queueId int64, search *model.SearchQueueResourceGroup) ([]*model.QueueResourceGroup, bool, *model.AppError) {
	list, err := app.Store.QueueResource().GetAllPage(ctx, domainId, queueId, search)
	if err != nil {
		return nil, false, err
	}
	search.RemoveLastElemIfNeed(&list)
	return list, search.EndOfList(), nil
}

func (app *App) GetQueueResourceGroup(ctx context.Context, domainId, queueId, id int64) (*model.QueueResourceGroup, *model.AppError) {
	return app.Store.QueueResource().Get(ctx, domainId, queueId, id)
}

func (app *App) UpdateQueueResourceGroup(ctx context.Context, domainId int64, qr *model.QueueResourceGroup) (*model.QueueResourceGroup, *model.AppError) {
	oldQr, err := app.GetQueueResourceGroup(ctx, domainId, qr.QueueId, qr.Id)
	if err != nil {
		return nil, err
	}

	oldQr.ResourceGroup = qr.ResourceGroup

	oldQr, err = app.Store.QueueResource().Update(ctx, domainId, oldQr)
	if err != nil {
		return nil, err
	}

	return oldQr, nil
}

func (app *App) RemoveQueueResourceGroup(ctx context.Context, domainId, queueId, id int64) (*model.QueueResourceGroup, *model.AppError) {
	qr, err := app.GetQueueResourceGroup(ctx, domainId, queueId, id)

	if err != nil {
		return nil, err
	}

	err = app.Store.QueueResource().Delete(ctx, queueId, id)
	if err != nil {
		return nil, err
	}
	return qr, nil
}
