package app

import (
	"context"

	"github.com/webitel/engine/model"
)

func (app *App) CreateTeamHook(ctx context.Context, domainId int64, teamId int64, hook *model.TeamHook) (*model.TeamHook, model.AppError) {
	return app.Store.TeamHook().Create(ctx, domainId, teamId, hook)
}

func (app *App) SearchTeamHook(ctx context.Context, domainId int64, teamId int64, search *model.SearchTeamHook) ([]*model.TeamHook, bool, model.AppError) {
	list, err := app.Store.TeamHook().GetAllPage(ctx, domainId, teamId, search)
	if err != nil {
		return nil, false, err
	}
	search.RemoveLastElemIfNeed(&list)
	return list, search.EndOfList(), nil
}

func (app *App) GetTeamHook(ctx context.Context, domainId int64, teamId int64, id uint32) (*model.TeamHook, model.AppError) {
	return app.Store.TeamHook().Get(ctx, domainId, teamId, id)
}

func (app *App) UpdateTeamHook(ctx context.Context, domainId int64, teamId int64, hook *model.TeamHook) (*model.TeamHook, model.AppError) {
	oldHook, err := app.GetTeamHook(ctx, domainId, teamId, hook.Id)
	if err != nil {
		return nil, err
	}

	oldHook.Schema = hook.Schema
	oldHook.Enabled = hook.Enabled
	oldHook.Event = hook.Event
	oldHook.UpdatedAt = hook.UpdatedAt
	oldHook.UpdatedBy = hook.UpdatedBy

	oldHook, err = app.Store.TeamHook().Update(ctx, domainId, teamId, oldHook)
	if err != nil {
		return nil, err
	}

	return oldHook, nil
}

func (app *App) PatchTeamHook(ctx context.Context, domainId int64, teamId int64, id uint32, patch *model.TeamHookPatch) (*model.TeamHook, model.AppError) {
	oldHook, err := app.GetTeamHook(ctx, domainId, teamId, id)
	if err != nil {
		return nil, err
	}

	oldHook.Patch(patch)
	oldHook.UpdatedBy = &patch.UpdatedBy
	oldHook.UpdatedAt = patch.UpdatedAt

	if err = oldHook.IsValid(); err != nil {
		return nil, err
	}

	oldHook, err = app.Store.TeamHook().Update(ctx, domainId, teamId, oldHook)
	if err != nil {
		return nil, err
	}

	return oldHook, nil
}

func (app *App) RemoveTeamHook(ctx context.Context, domainId int64, teamId int64, id uint32) (*model.TeamHook, model.AppError) {
	qb, err := app.GetTeamHook(ctx, domainId, teamId, id)

	if err != nil {
		return nil, err
	}

	err = app.Store.TeamHook().Delete(ctx, domainId, teamId, id)
	if err != nil {
		return nil, err
	}
	return qb, nil
}
