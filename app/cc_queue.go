package app

import "github.com/webitel/engine/model"

func (a *App) QueueCheckAccess(domainId, id int64, groups []int, access model.PermissionAccess) (bool, *model.AppError) {
	return a.Store.Queue().CheckAccess(domainId, id, groups, access)
}

func (a *App) CreateQueue(queue *model.Queue) (*model.Queue, *model.AppError) {
	return a.Store.Queue().Create(queue)
}

func (a *App) GetQueuePage(domainId int64, page, perPage int) ([]*model.Queue, *model.AppError) {
	return a.Store.Queue().GetAllPage(domainId, page*perPage, perPage)
}

func (a *App) GetQueuePageByGroups(domainId int64, groups []int, page, perPage int) ([]*model.Queue, *model.AppError) {
	return a.Store.Queue().GetAllPageByGroups(domainId, groups, page*perPage, perPage)
}

func (a *App) GetQueueById(domainId, id int64) (*model.Queue, *model.AppError) {
	return a.Store.Queue().Get(domainId, id)
}

func (a *App) PatchQueue(domainId, id int64, patch *model.QueuePatch) (*model.Queue, *model.AppError) {
	oldQueue, err := a.GetQueueById(domainId, id)
	if err != nil {
		return nil, err
	}

	oldQueue.Patch(patch)

	if err = oldQueue.IsValid(); err != nil {
		return nil, err
	}

	oldQueue, err = a.Store.Queue().Update(oldQueue)
	if err != nil {
		return nil, err
	}

	return oldQueue, nil
}

func (a *App) UpdateQueue(queue *model.Queue) (*model.Queue, *model.AppError) {
	oldQueue, err := a.GetQueueById(queue.DomainId, queue.Id)
	if err != nil {
		return nil, err
	}

	oldQueue.UpdatedAt = queue.UpdatedAt
	oldQueue.UpdatedBy.Id = queue.UpdatedBy.Id
	oldQueue.Strategy = queue.Strategy
	oldQueue.Enabled = queue.Enabled
	oldQueue.Payload = queue.Payload
	oldQueue.Calendar.Id = queue.Calendar.Id
	oldQueue.Priority = queue.Priority
	oldQueue.Name = queue.Name
	oldQueue.Variables = queue.Variables
	oldQueue.Timeout = queue.Timeout
	oldQueue.DncList = queue.DncList
	oldQueue.SecLocateAgent = queue.SecLocateAgent
	oldQueue.Type = queue.Type
	oldQueue.Team = queue.Team
	oldQueue.Description = queue.Description

	oldQueue, err = a.Store.Queue().Update(oldQueue)
	if err != nil {
		return nil, err
	}

	return oldQueue, nil
}

func (a *App) RemoveQueue(domainId, id int64) (*model.Queue, *model.AppError) {
	queue, err := a.Store.Queue().Get(domainId, id)

	if err != nil {
		return nil, err
	}

	err = a.Store.Queue().Delete(domainId, id)
	if err != nil {
		return nil, err
	}
	return queue, nil
}
