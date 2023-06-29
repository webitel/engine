package app

import (
	"context"

	"github.com/webitel/engine/model"
)

func (app *App) CreateChatPlan(ctx context.Context, domainId int64, plan *model.ChatPlan) (*model.ChatPlan, model.AppError) {
	return app.Store.ChatPlan().Create(ctx, domainId, plan)
}

func (a *App) GetChatPlanPage(ctx context.Context, domainId int64, search *model.SearchChatPlan) ([]*model.ChatPlan, bool, model.AppError) {
	list, err := a.Store.ChatPlan().GetAllPage(ctx, domainId, search)
	if err != nil {
		return nil, false, err
	}
	search.RemoveLastElemIfNeed(&list)
	return list, search.EndOfList(), nil
}

func (a *App) GetChatPlan(ctx context.Context, domainId int64, id int32) (*model.ChatPlan, model.AppError) {
	return a.Store.ChatPlan().Get(ctx, domainId, id)
}

func (a *App) UpdateChatPlan(ctx context.Context, domainId int64, plan *model.ChatPlan) (*model.ChatPlan, model.AppError) {
	oldPlan, err := a.GetChatPlan(ctx, domainId, plan.Id)
	if err != nil {
		return nil, err
	}

	oldPlan.Name = plan.Name
	oldPlan.Description = plan.Description
	oldPlan.Enabled = plan.Enabled
	oldPlan.Schema = plan.Schema

	oldPlan, err = a.Store.ChatPlan().Update(ctx, domainId, oldPlan)
	if err != nil {
		return nil, err
	}

	return oldPlan, nil
}

func (a *App) PatchChatPlan(ctx context.Context, domainId int64, id int32, patch *model.PatchChatPlan) (*model.ChatPlan, model.AppError) {
	oldPlan, err := a.GetChatPlan(ctx, domainId, id)
	if err != nil {
		return nil, err
	}

	oldPlan.Patch(patch)

	if err = oldPlan.IsValid(); err != nil {
		return nil, err
	}

	oldPlan, err = a.Store.ChatPlan().Update(ctx, domainId, oldPlan)
	if err != nil {
		return nil, err
	}

	return oldPlan, nil
}

func (a *App) RemoveChatPlan(ctx context.Context, domainId int64, id int32) (*model.ChatPlan, model.AppError) {
	plan, err := a.GetChatPlan(ctx, domainId, id)

	if err != nil {
		return nil, err
	}

	err = a.Store.ChatPlan().Delete(ctx, domainId, id)
	if err != nil {
		return nil, err
	}
	return plan, nil
}
