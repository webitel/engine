package app

import (
	"context"

	"github.com/webitel/engine/model"
)

func (a *App) SearchQueueSkill(ctx context.Context, domainId int64, search *model.SearchQueueSkill) ([]*model.QueueSkill, bool, model.AppError) {
	list, err := a.Store.QueueSkill().GetAllPage(ctx, domainId, search)
	if err != nil {
		return nil, false, err
	}
	search.RemoveLastElemIfNeed(&list)
	return list, search.EndOfList(), nil
}

func (a *App) CreateQueueSkill(ctx context.Context, domainId int64, qs *model.QueueSkill) (*model.QueueSkill, model.AppError) {
	return a.Store.QueueSkill().Create(ctx, domainId, qs)
}

func (a *App) GetQueueSkill(ctx context.Context, domainId int64, queueId, id uint32) (*model.QueueSkill, model.AppError) {
	return a.Store.QueueSkill().Get(ctx, domainId, queueId, id)
}

func (a *App) UpdateQueueSkill(ctx context.Context, domainId int64, qs *model.QueueSkill) (*model.QueueSkill, model.AppError) {
	oldQs, err := a.GetQueueSkill(ctx, domainId, qs.QueueId, qs.Id)
	if err != nil {
		return nil, err
	}

	oldQs.Lvl = qs.Lvl
	oldQs.MaxCapacity = qs.MaxCapacity
	oldQs.MinCapacity = qs.MinCapacity
	oldQs.Skill = qs.Skill
	oldQs.Buckets = qs.Buckets
	oldQs.Enabled = qs.Enabled

	oldQs, err = a.Store.QueueSkill().Update(ctx, domainId, oldQs)
	if err != nil {
		return nil, err
	}

	return oldQs, nil
}

func (a *App) PatchQueueSkill(ctx context.Context, domainId int64, queueId, id uint32, patch *model.QueueSkillPatch) (*model.QueueSkill, model.AppError) {
	oldQs, err := a.GetQueueSkill(ctx, domainId, queueId, id)
	if err != nil {
		return nil, err
	}

	oldQs.Patch(patch)

	if err = oldQs.IsValid(); err != nil {
		return nil, err
	}

	oldQs, err = a.Store.QueueSkill().Update(ctx, domainId, oldQs)
	if err != nil {
		return nil, err
	}

	return oldQs, nil
}

func (a *App) RemoveQueueSkill(ctx context.Context, domainId int64, queueId, id uint32) (*model.QueueSkill, model.AppError) {
	qs, err := a.GetQueueSkill(ctx, domainId, queueId, id)

	if err != nil {
		return nil, err
	}

	err = a.Store.QueueSkill().Delete(ctx, domainId, queueId, id)
	if err != nil {
		return nil, err
	}
	return qs, nil
}
