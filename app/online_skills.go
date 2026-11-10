package app

import (
	"context"

	"github.com/webitel/engine/model"
)

func (app *App) CreateOnlineSkills(ctx context.Context, preset *model.OnlineSkills) (*model.OnlineSkills, model.AppError) {
	preset.PreSave()

	return app.Store.OnlineSkills().Create(ctx, preset)
}

func (app *App) GetOnlineSkills(ctx context.Context, query *model.GetSkillPresetQuery) (*model.OnlineSkills, model.AppError) {
	return app.Store.OnlineSkills().Get(ctx, query)
}

func (app *App) SearchOnlineSkills(ctx context.Context, query *model.SearchOnlineSkillsQuery) ([]*model.OnlineSkills, model.AppError) {
	return app.Store.OnlineSkills().Search(ctx, query)
}

func (app *App) UpdateOnlineSkills(ctx context.Context, cmd *model.OnlineSkills) (*model.OnlineSkills, model.AppError) {
	cmd.PreSave()

	return app.Store.OnlineSkills().Update(ctx, cmd)
}

func (app *App) PatchOnlineSkills(ctx context.Context, cmd *model.PatchOnlineSkillsCmd) (*model.OnlineSkills, model.AppError) {
	cmd.PrePatch()

	return app.Store.OnlineSkills().Patch(ctx, cmd)
}

func (app *App) DeleteOnlineSkills(ctx context.Context, cmd *model.DeleteSkillPresetCmd) model.AppError {
	return app.Store.OnlineSkills().Delete(ctx, cmd)
}
