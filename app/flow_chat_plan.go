package app

import "github.com/webitel/engine/model"

func (app *App) CreateChatPlan(domainId int64, plan *model.ChatPlan) (*model.ChatPlan, *model.AppError) {
	return app.Store.ChatPlan().Create(domainId, plan)
}

func (a *App) GetChatPlanPage(domainId int64, search *model.SearchChatPlan) ([]*model.ChatPlan, bool, *model.AppError) {
	list, err := a.Store.ChatPlan().GetAllPage(domainId, search)
	if err != nil {
		return nil, false, err
	}
	search.RemoveLastElemIfNeed(&list)
	return list, search.EndOfList(), nil
}

func (a *App) GetChatPlan(domainId int64, id int32) (*model.ChatPlan, *model.AppError) {
	return a.Store.ChatPlan().Get(domainId, id)
}

func (a *App) UpdateChatPlan(domainId int64, plan *model.ChatPlan) (*model.ChatPlan, *model.AppError) {
	oldPlan, err := a.GetChatPlan(domainId, plan.Id)
	if err != nil {
		return nil, err
	}

	oldPlan.Name = plan.Name
	oldPlan.Description = plan.Description
	oldPlan.Enabled = plan.Enabled
	oldPlan.Schema = plan.Schema

	oldPlan, err = a.Store.ChatPlan().Update(domainId, oldPlan)
	if err != nil {
		return nil, err
	}

	return oldPlan, nil
}

func (a *App) PatchChatPlan(domainId int64, id int32, patch *model.PatchChatPlan) (*model.ChatPlan, *model.AppError) {
	oldPlan, err := a.GetChatPlan(domainId, id)
	if err != nil {
		return nil, err
	}

	oldPlan.Patch(patch)

	if err = oldPlan.IsValid(); err != nil {
		return nil, err
	}

	oldPlan, err = a.Store.ChatPlan().Update(domainId, oldPlan)
	if err != nil {
		return nil, err
	}

	return oldPlan, nil
}

func (a *App) RemoveChatPlan(domainId int64, id int32) (*model.ChatPlan, *model.AppError) {
	plan, err := a.GetChatPlan(domainId, id)

	if err != nil {
		return nil, err
	}

	err = a.Store.ChatPlan().Delete(domainId, id)
	if err != nil {
		return nil, err
	}
	return plan, nil
}
