package app

import (
	"context"

	"github.com/webitel/engine/model"
)

func (app *App) CreateSkillPreset(ctx context.Context, preset *model.SkillPreset) (*model.SkillPreset, model.AppError) {
	preset.PreSave()

	return app.Store.SkillPreset().Create(ctx, preset)
}

func (app *App) GetSkillPreset(ctx context.Context, query *model.GetSkillPresetQuery) (*model.SkillPreset, model.AppError) {
	return app.Store.SkillPreset().Get(ctx, query)
}

func (app *App) SearchSkillPreset(ctx context.Context, query *model.SearchSkillPresetQuery) ([]*model.SkillPreset, model.AppError) {
	return app.Store.SkillPreset().Search(ctx, query)
}

func (app *App) UpdateSkillPreset(ctx context.Context, cmd *model.SkillPreset) (*model.SkillPreset, model.AppError) {
	cmd.PreSave()

	return app.Store.SkillPreset().Update(ctx, cmd)
}

func (app *App) PatchSkillPreset(ctx context.Context, cmd *model.PatchSkillPresetCmd) (*model.SkillPreset, model.AppError) {
	cmd.PrePatch()

	return app.Store.SkillPreset().Patch(ctx, cmd)
}

func (app *App) DeleteSkillPreset(ctx context.Context, cmd *model.DeleteSkillPresetCmd) ([]*model.SkillPreset, model.AppError) {
	return app.Store.SkillPreset().Delete(ctx, cmd)
}
