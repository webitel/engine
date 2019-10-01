package app

import "github.com/webitel/engine/model"

func (app *App) CreateQueueRouting(routing *model.QueueRouting) (*model.QueueRouting, *model.AppError) {
	return app.Store.QueueRouting().Create(routing)
}

func (a *App) GetQueueRoutingPage(domainId, queueId int64, page, perPage int) ([]*model.QueueRouting, *model.AppError) {
	return a.Store.QueueRouting().GetAllPage(domainId, queueId, page*perPage, perPage)
}

func (a *App) GetQueueRoutingById(domainId, queueId, id int64) (*model.QueueRouting, *model.AppError) {
	return a.Store.QueueRouting().Get(domainId, queueId, id)
}

func (a *App) UpdateQueueRouting(domainId int64, qr *model.QueueRouting) (*model.QueueRouting, *model.AppError) {
	oldQr, err := a.GetQueueRoutingById(domainId, qr.QueueId, qr.Id)
	if err != nil {
		return nil, err
	}

	oldQr.Pattern = qr.Pattern
	oldQr.Priority = qr.Priority
	oldQr.Disabled = qr.Disabled

	oldQr, err = a.Store.QueueRouting().Update(oldQr)
	if err != nil {
		return nil, err
	}

	return oldQr, nil
}

func (a *App) RemoveQueueRouting(domainId, queueId, id int64) (*model.QueueRouting, *model.AppError) {
	qr, err := a.Store.QueueRouting().Get(domainId, queueId, id)

	if err != nil {
		return nil, err
	}

	err = a.Store.QueueRouting().Delete(queueId, id)
	if err != nil {
		return nil, err
	}
	return qr, nil
}
