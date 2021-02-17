package app

import "github.com/webitel/engine/model"

func (a *App) SearchQueueSkill(domainId int64, search *model.SearchQueueSkill) ([]*model.QueueSkill, bool, *model.AppError) {
	list, err := a.Store.QueueSkill().GetAllPage(domainId, search)
	if err != nil {
		return nil, false, err
	}
	search.RemoveLastElemIfNeed(&list)
	return list, search.EndOfList(), nil
}

func (a *App) CreateQueueSkill(domainId int64, qs *model.QueueSkill) (*model.QueueSkill, *model.AppError) {
	return a.Store.QueueSkill().Create(domainId, qs)
}

func (a *App) GetQueueSkill(domainId int64, queueId, id uint32) (*model.QueueSkill, *model.AppError) {
	return a.Store.QueueSkill().Get(domainId, queueId, id)
}

func (a *App) UpdateQueueSkill(domainId int64, qs *model.QueueSkill) (*model.QueueSkill, *model.AppError) {
	oldQs, err := a.GetQueueSkill(domainId, qs.QueueId, qs.Id)
	if err != nil {
		return nil, err
	}

	oldQs.Lvl = qs.Lvl
	oldQs.MaxCapacity = qs.MaxCapacity
	oldQs.MinCapacity = qs.MinCapacity
	oldQs.Skill = qs.Skill
	oldQs.Buckets = qs.Buckets
	oldQs.Enabled = qs.Enabled

	oldQs, err = a.Store.QueueSkill().Update(domainId, oldQs)
	if err != nil {
		return nil, err
	}

	return oldQs, nil
}

func (a *App) PatchQueueSkill(domainId int64, queueId, id uint32, patch *model.QueueSkillPatch) (*model.QueueSkill, *model.AppError) {
	oldQs, err := a.GetQueueSkill(domainId, queueId, id)
	if err != nil {
		return nil, err
	}

	oldQs.Patch(patch)

	if err = oldQs.IsValid(); err != nil {
		return nil, err
	}

	oldQs, err = a.Store.QueueSkill().Update(domainId, oldQs)
	if err != nil {
		return nil, err
	}

	return oldQs, nil
}

func (a *App) RemoveQueueSkill(domainId int64, queueId, id uint32) (*model.QueueSkill, *model.AppError) {
	qs, err := a.GetQueueSkill(domainId, queueId, id)

	if err != nil {
		return nil, err
	}

	err = a.Store.QueueSkill().Delete(domainId, queueId, id)
	if err != nil {
		return nil, err
	}
	return qs, nil
}
