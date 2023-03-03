package app

import (
	"context"
	"github.com/webitel/engine/auth_manager"
	"github.com/webitel/engine/model"
)

func (a *App) QueueCheckAccess(ctx context.Context, domainId, id int64, groups []int, access auth_manager.PermissionAccess) (bool, *model.AppError) {
	return a.Store.Queue().CheckAccess(ctx, domainId, id, groups, access)
}

func (a *App) CreateQueue(ctx context.Context, queue *model.Queue) (*model.Queue, *model.AppError) {
	return a.Store.Queue().Create(ctx, queue)
}

func (a *App) GetQueuePage(ctx context.Context, domainId int64, search *model.SearchQueue) ([]*model.Queue, bool, *model.AppError) {
	list, err := a.Store.Queue().GetAllPage(ctx, domainId, search)
	if err != nil {
		return nil, false, err
	}
	search.RemoveLastElemIfNeed(&list)
	return list, search.EndOfList(), nil
}

func (a *App) GetQueuePageByGroups(ctx context.Context, domainId int64, groups []int, search *model.SearchQueue) ([]*model.Queue, bool, *model.AppError) {
	list, err := a.Store.Queue().GetAllPageByGroups(ctx, domainId, groups, search)
	if err != nil {
		return nil, false, err
	}
	search.RemoveLastElemIfNeed(&list)
	return list, search.EndOfList(), nil
}

func (a *App) GetQueueById(ctx context.Context, domainId, id int64) (*model.Queue, *model.AppError) {
	return a.Store.Queue().Get(ctx, domainId, id)
}

func (a *App) PatchQueue(ctx context.Context, domainId, id int64, patch *model.QueuePatch) (*model.Queue, *model.AppError) {
	oldQueue, err := a.GetQueueById(ctx, domainId, id)
	if err != nil {
		return nil, err
	}

	oldQueue.Patch(patch)

	if err = oldQueue.IsValid(); err != nil {
		return nil, err
	}

	oldQueue, err = a.Store.Queue().Update(ctx, oldQueue)
	if err != nil {
		return nil, err
	}

	return oldQueue, nil
}

func (a *App) UpdateQueue(ctx context.Context, queue *model.Queue) (*model.Queue, *model.AppError) {
	oldQueue, err := a.GetQueueById(ctx, queue.DomainId, queue.Id)
	if err != nil {
		return nil, err
	}

	oldQueue.UpdatedAt = queue.UpdatedAt
	oldQueue.UpdatedBy = queue.UpdatedBy
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
	oldQueue.FormSchema = queue.FormSchema
	oldQueue.Grantee = queue.Grantee

	oldQueue, err = a.Store.Queue().Update(ctx, oldQueue)
	if err != nil {
		return nil, err
	}

	return oldQueue, nil
}

func (a *App) RemoveQueue(ctx context.Context, domainId, id int64) (*model.Queue, *model.AppError) {
	queue, err := a.Store.Queue().Get(ctx, domainId, id)

	if err != nil {
		return nil, err
	}

	err = a.Store.Queue().Delete(ctx, domainId, id)
	if err != nil {
		return nil, err
	}
	return queue, nil
}

func (a *App) GetQueueReportGeneral(ctx context.Context, domainId int64, supervisorId int64, groups []int, access auth_manager.PermissionAccess, search *model.SearchQueueReportGeneral) (*model.QueueReportGeneralAgg, bool, *model.AppError) {
	list, err := a.Store.Queue().QueueReportGeneral(ctx, domainId, supervisorId, groups, access, search)
	if err != nil {
		return nil, false, err
	}
	search.RemoveLastElemIfNeed(&list.Items)
	return list, search.EndOfList(), nil
}
