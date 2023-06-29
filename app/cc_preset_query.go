package app

import (
	"context"

	"github.com/webitel/engine/model"
)

func (app *App) CreatePresetQuery(ctx context.Context, domainId, userId int64, preset *model.PresetQuery) (*model.PresetQuery, model.AppError) {
	return app.Store.PresetQuery().Create(ctx, domainId, userId, preset)
}

func (app *App) GetPresetQueryPage(ctx context.Context, domainId, userId int64, search *model.SearchPresetQuery) ([]*model.PresetQuery, bool, model.AppError) {
	list, err := app.Store.PresetQuery().GetAllPage(ctx, domainId, userId, search)
	if err != nil {
		return nil, false, err
	}
	search.RemoveLastElemIfNeed(&list)
	return list, search.EndOfList(), nil
}

func (app *App) GetPresetQuery(ctx context.Context, domainId, userId int64, id int32) (*model.PresetQuery, model.AppError) {
	return app.Store.PresetQuery().Get(ctx, domainId, userId, id)
}

func (app *App) UpdatePresetQuery(ctx context.Context, domainId, userId int64, preset *model.PresetQuery) (*model.PresetQuery, model.AppError) {
	oldPreset, err := app.GetPresetQuery(ctx, domainId, userId, preset.Id)
	if err != nil {
		return nil, err
	}

	oldPreset.Name = preset.Name
	oldPreset.Section = preset.Section
	oldPreset.Preset = preset.Preset
	oldPreset.Description = preset.Description
	oldPreset.UpdatedAt = preset.UpdatedAt

	if err = oldPreset.IsValid(); err != nil {
		return nil, err
	}

	oldPreset, err = app.Store.PresetQuery().Update(ctx, domainId, userId, oldPreset)
	if err != nil {
		return nil, err
	}

	return oldPreset, nil
}

func (app *App) PatchPresetQuery(ctx context.Context, domainId, userId int64, id int32, patch *model.PresetQueryPatch) (*model.PresetQuery, model.AppError) {
	oldPreset, err := app.GetPresetQuery(ctx, domainId, userId, id)
	if err != nil {
		return nil, err
	}

	oldPreset.Patch(patch)

	if err = oldPreset.IsValid(); err != nil {
		return nil, err
	}

	oldPreset, err = app.Store.PresetQuery().Update(ctx, domainId, userId, oldPreset)
	if err != nil {
		return nil, err
	}

	return oldPreset, nil
}

func (app *App) RemovePresetQuery(ctx context.Context, domainId, userId int64, id int32) (*model.PresetQuery, model.AppError) {
	preset, err := app.Store.PresetQuery().Get(ctx, domainId, userId, id)

	if err != nil {
		return nil, err
	}

	err = app.Store.PresetQuery().Delete(ctx, domainId, userId, id)
	if err != nil {
		return nil, err
	}
	return preset, nil
}
