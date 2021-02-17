package app

import (
	"github.com/webitel/engine/auth_manager"
	"github.com/webitel/engine/model"
)

func (a *App) QueueCheckAccess(domainId, id int64, groups []int, access auth_manager.PermissionAccess) (bool, *model.AppError) {
	return a.Store.Queue().CheckAccess(domainId, id, groups, access)
}

func (a *App) CreateQueue(queue *model.Queue) (*model.Queue, *model.AppError) {
	return a.Store.Queue().Create(queue)
}

func (a *App) GetQueuePage(domainId int64, search *model.SearchQueue) ([]*model.Queue, bool, *model.AppError) {
	list, err := a.Store.Queue().GetAllPage(domainId, search)
	if err != nil {
		return nil, false, err
	}
	search.RemoveLastElemIfNeed(&list)
	return list, search.EndOfList(), nil
}

func (a *App) GetQueuePageByGroups(domainId int64, groups []int, search *model.SearchQueue) ([]*model.Queue, bool, *model.AppError) {
	list, err := a.Store.Queue().GetAllPageByGroups(domainId, groups, search)
	if err != nil {
		return nil, false, err
	}
	search.RemoveLastElemIfNeed(&list)
	return list, search.EndOfList(), nil
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
	oldQueue.Schema = queue.Schema
	oldQueue.Ringtone = queue.Ringtone
	oldQueue.DoSchema = queue.DoSchema
	oldQueue.AfterSchema = queue.AfterSchema
	oldQueue.Description = queue.Description
	oldQueue.StickyAgent = queue.StickyAgent
	oldQueue.Processing = queue.Processing
	oldQueue.ProcessingSec = queue.ProcessingSec
	oldQueue.ProcessingRenewalSec = queue.ProcessingRenewalSec

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

func (a *App) GetQueueReportGeneral(domainId int64, search *model.SearchQueueReportGeneral) ([]*model.QueueReportGeneral, bool, *model.AppError) {
	list, err := a.Store.Queue().QueueReportGeneral(domainId, search)
	if err != nil {
		return nil, false, err
	}
	search.RemoveLastElemIfNeed(&list)
	return list, search.EndOfList(), nil
}
