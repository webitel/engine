package controller

import (
	"context"

	"github.com/webitel/engine/auth_manager"
	"github.com/webitel/engine/model"
)

func (c *Controller) CreatePresetQuery(ctx context.Context, session *auth_manager.Session, preset *model.PresetQuery) (*model.PresetQuery, model.AppError) {
	preset.CreatedAt = model.GetTime()
	preset.UpdatedAt = preset.CreatedAt

	return c.app.CreatePresetQuery(ctx, session.Domain(0), session.UserId, preset)
}

func (c *Controller) SearchPresetQuery(ctx context.Context, session *auth_manager.Session, search *model.SearchPresetQuery) ([]*model.PresetQuery, bool, model.AppError) {
	return c.app.GetPresetQueryPage(ctx, session.Domain(0), session.UserId, search)
}

func (c *Controller) ReadPresetQuery(ctx context.Context, session *auth_manager.Session, id int32) (*model.PresetQuery, model.AppError) {
	return c.app.GetPresetQuery(ctx, session.Domain(0), session.UserId, id)
}

func (c *Controller) UpdatePresetQuery(ctx context.Context, session *auth_manager.Session, preset *model.PresetQuery) (*model.PresetQuery, model.AppError) {
	preset.UpdatedAt = model.GetTime()
	return c.app.UpdatePresetQuery(ctx, session.Domain(0), session.UserId, preset)
}

func (c *Controller) PatchPresetQuery(ctx context.Context, session *auth_manager.Session, id int32, patch *model.PresetQueryPatch) (*model.PresetQuery, model.AppError) {
	patch.UpdatedAt = *model.GetTime()
	return c.app.PatchPresetQuery(ctx, session.Domain(0), session.UserId, id, patch)
}

func (c *Controller) RemovePresetQuery(ctx context.Context, session *auth_manager.Session, id int32) (*model.PresetQuery, model.AppError) {
	return c.app.RemovePresetQuery(ctx, session.Domain(0), session.UserId, id)
}
