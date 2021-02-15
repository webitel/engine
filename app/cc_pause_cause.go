package app

import "github.com/webitel/engine/model"

func (a *App) GetPauseCausePage(domainId int64, search *model.SearchAgentPauseCause) ([]*model.AgentPauseCause, bool, *model.AppError) {
	list, err := a.Store.PauseCause().GetAllPage(domainId, search)
	if err != nil {
		return nil, false, err
	}
	search.RemoveLastElemIfNeed(&list)
	return list, search.EndOfList(), nil
}

func (a *App) CreatePauseCause(domainId int64, cause *model.AgentPauseCause) (*model.AgentPauseCause, *model.AppError) {
	return a.Store.PauseCause().Create(domainId, cause)
}

func (a *App) GetPauseCause(domainId int64, id uint32) (*model.AgentPauseCause, *model.AppError) {
	return a.Store.PauseCause().Get(domainId, id)
}

func (a *App) UpdatePauseCause(domainId int64, cause *model.AgentPauseCause) (*model.AgentPauseCause, *model.AppError) {
	oldCause, err := a.GetPauseCause(domainId, uint32(cause.Id))
	if err != nil {
		return nil, err
	}

	oldCause.UpdatedBy = cause.UpdatedBy
	oldCause.UpdatedAt = cause.UpdatedAt

	oldCause.Name = cause.Name
	oldCause.Description = cause.Description
	oldCause.AllowAgent = cause.AllowAgent
	oldCause.AllowSupervisor = cause.AllowSupervisor
	oldCause.LimitPerDay = cause.LimitPerDay

	oldCause, err = a.Store.PauseCause().Update(domainId, oldCause)
	if err != nil {
		return nil, err
	}

	return oldCause, nil
}

func (a *App) PatchPauseCause(domainId int64, id uint32, patch *model.AgentPauseCausePatch) (*model.AgentPauseCause, *model.AppError) {
	oldCause, err := a.GetPauseCause(domainId, id)
	if err != nil {
		return nil, err
	}

	oldCause.Patch(patch)

	if err = oldCause.IsValid(); err != nil {
		return nil, err
	}

	oldCause, err = a.Store.PauseCause().Update(domainId, oldCause)
	if err != nil {
		return nil, err
	}

	return oldCause, nil
}

func (a *App) RemovePauseCause(domainId int64, id uint32) (*model.AgentPauseCause, *model.AppError) {
	cause, err := a.GetPauseCause(domainId, id)

	if err != nil {
		return nil, err
	}

	err = a.Store.PauseCause().Delete(domainId, id)
	if err != nil {
		return nil, err
	}
	return cause, nil
}
