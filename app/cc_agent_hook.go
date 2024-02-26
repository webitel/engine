package app

import (
	"context"

	"github.com/webitel/engine/model"
)

func (app *App) CreateAgentHook(ctx context.Context, domainId int64, agentId int64, hook *model.AgentHook) (*model.AgentHook, model.AppError) {
	return app.Store.AgentHook().Create(ctx, domainId, agentId, hook)
}

func (app *App) SearchAgentHook(ctx context.Context, domainId int64, agentId int64, search *model.SearchAgentHook) ([]*model.AgentHook, bool, model.AppError) {
	list, err := app.Store.AgentHook().GetAllPage(ctx, domainId, agentId, search)
	if err != nil {
		return nil, false, err
	}
	search.RemoveLastElemIfNeed(&list)
	return list, search.EndOfList(), nil
}

func (app *App) GetAgentHook(ctx context.Context, domainId int64, agentId int64, id int32) (*model.AgentHook, model.AppError) {
	return app.Store.AgentHook().Get(ctx, domainId, agentId, id)
}

func (app *App) UpdateAgentHook(ctx context.Context, domainId int64, agentId int64, hook *model.AgentHook) (*model.AgentHook, model.AppError) {
	oldHook, err := app.GetAgentHook(ctx, domainId, agentId, hook.Id)
	if err != nil {
		return nil, err
	}

	oldHook.Schema = hook.Schema
	oldHook.Enabled = hook.Enabled
	oldHook.Event = hook.Event
	oldHook.UpdatedAt = hook.UpdatedAt
	oldHook.UpdatedBy = hook.UpdatedBy

	oldHook, err = app.Store.AgentHook().Update(ctx, domainId, agentId, oldHook)
	if err != nil {
		return nil, err
	}

	return oldHook, nil
}

func (app *App) PatchAgentHook(ctx context.Context, domainId int64, agentId int64, id int32, patch *model.AgentHookPatch) (*model.AgentHook, model.AppError) {
	oldHook, err := app.GetAgentHook(ctx, domainId, agentId, id)
	if err != nil {
		return nil, err
	}

	oldHook.Patch(patch)
	oldHook.UpdatedBy = &patch.UpdatedBy
	oldHook.UpdatedAt = patch.UpdatedAt

	if err = oldHook.IsValid(); err != nil {
		return nil, err
	}

	oldHook, err = app.Store.AgentHook().Update(ctx, domainId, agentId, oldHook)
	if err != nil {
		return nil, err
	}

	return oldHook, nil
}

func (app *App) RemoveAgentHook(ctx context.Context, domainId int64, agentId int64, id int32) (*model.AgentHook, model.AppError) {
	qb, err := app.GetAgentHook(ctx, domainId, agentId, id)

	if err != nil {
		return nil, err
	}

	err = app.Store.AgentHook().Delete(ctx, domainId, agentId, id)
	if err != nil {
		return nil, err
	}
	return qb, nil
}
