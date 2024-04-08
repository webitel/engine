package app

import (
	"context"
	"github.com/webitel/engine/model"
)

func (app *App) CreateTeamTrigger(ctx context.Context, domainId int64, teamId int64, trigger *model.TeamTrigger) (*model.TeamTrigger, model.AppError) {
	return app.Store.TeamTrigger().Create(ctx, domainId, teamId, trigger)
}

func (app *App) SearchTeamTrigger(ctx context.Context, domainId int64, teamId int64, search *model.SearchTeamTrigger) ([]*model.TeamTrigger, bool, model.AppError) {
	list, err := app.Store.TeamTrigger().GetAllPage(ctx, domainId, teamId, search)
	if err != nil {
		return nil, false, err
	}
	search.RemoveLastElemIfNeed(&list)
	return list, search.EndOfList(), nil
}

func (app *App) GetTeamTrigger(ctx context.Context, domainId int64, teamId int64, id uint32) (*model.TeamTrigger, model.AppError) {
	return app.Store.TeamTrigger().Get(ctx, domainId, teamId, id)
}

func (app *App) UpdateTeamTrigger(ctx context.Context, domainId int64, teamId int64, trigger *model.TeamTrigger) (*model.TeamTrigger, model.AppError) {
	oldTrigger, err := app.GetTeamTrigger(ctx, domainId, teamId, trigger.Id)
	if err != nil {
		return nil, err
	}

	oldTrigger.Schema = trigger.Schema
	oldTrigger.Enabled = trigger.Enabled
	oldTrigger.Name = trigger.Name
	oldTrigger.Description = trigger.Description

	oldTrigger.UpdatedAt = trigger.UpdatedAt
	oldTrigger.UpdatedBy = trigger.UpdatedBy

	oldTrigger, err = app.Store.TeamTrigger().Update(ctx, domainId, teamId, oldTrigger)
	if err != nil {
		return nil, err
	}

	return oldTrigger, nil
}

func (app *App) PatchTeamTrigger(ctx context.Context, domainId int64, teamId int64, id uint32, patch *model.TeamTriggerPatch) (*model.TeamTrigger, model.AppError) {
	oldTrigger, err := app.GetTeamTrigger(ctx, domainId, teamId, id)
	if err != nil {
		return nil, err
	}

	oldTrigger.Patch(patch)
	oldTrigger.UpdatedBy = &patch.UpdatedBy
	oldTrigger.UpdatedAt = &patch.UpdatedAt

	if err = oldTrigger.IsValid(); err != nil {
		return nil, err
	}

	oldTrigger, err = app.Store.TeamTrigger().Update(ctx, domainId, teamId, oldTrigger)
	if err != nil {
		return nil, err
	}

	return oldTrigger, nil
}

func (app *App) RemoveTeamTrigger(ctx context.Context, domainId int64, teamId int64, id uint32) (*model.TeamTrigger, model.AppError) {
	qb, err := app.GetTeamTrigger(ctx, domainId, teamId, id)

	if err != nil {
		return nil, err
	}

	err = app.Store.TeamTrigger().Delete(ctx, domainId, teamId, id)
	if err != nil {
		return nil, err
	}
	return qb, nil
}
