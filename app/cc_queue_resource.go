package app

import "github.com/webitel/engine/model"

func (app *App) CreateQueueResourceGroup(queueResourceGroup *model.QueueResourceGroup) (*model.QueueResourceGroup, *model.AppError) {
	return app.Store.QueueResource().Create(queueResourceGroup)
}

func (app *App) GetQueueResourceGroupPage(domainId, queueId int64, page, perPage int) ([]*model.QueueResourceGroup, *model.AppError) {
	return app.Store.QueueResource().GetAllPage(domainId, queueId, page*perPage, perPage)
}

func (app *App) GetQueueResourceGroup(domainId, queueId, id int64) (*model.QueueResourceGroup, *model.AppError) {
	return app.Store.QueueResource().Get(domainId, queueId, id)
}

func (app *App) UpdateQueueResourceGroup(domainId int64, qr *model.QueueResourceGroup) (*model.QueueResourceGroup, *model.AppError) {
	oldQr, err := app.GetQueueResourceGroup(domainId, qr.QueueId, qr.Id)
	if err != nil {
		return nil, err
	}

	oldQr.ResourceGroup = qr.ResourceGroup

	oldQr, err = app.Store.QueueResource().Update(domainId, oldQr)
	if err != nil {
		return nil, err
	}

	return oldQr, nil
}

func (app *App) RemoveQueueResourceGroup(domainId, queueId, id int64) (*model.QueueResourceGroup, *model.AppError) {
	qr, err := app.GetQueueResourceGroup(domainId, queueId, id)

	if err != nil {
		return nil, err
	}

	err = app.Store.QueueResource().Delete(queueId, id)
	if err != nil {
		return nil, err
	}
	return qr, nil
}
