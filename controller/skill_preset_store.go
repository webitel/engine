package controller

import (
	"context"

	"github.com/webitel/engine/model"
	"github.com/webitel/engine/pkg/wbt/auth_manager"
)

func (c *Controller) CreateSkillPreset(ctx context.Context, preset *model.SkillPreset) (*model.SkillPreset, model.AppError) {
	session, err := c.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	permission := session.GetPermission(model.PermissionSkill)
	if !permission.CanCreate() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_CREATE)
	}

	preset.CreatedBy = &model.Lookup{Id: int(session.UserId)}
	preset.UpdatedBy = &model.Lookup{Id: int(session.UserId)}
	preset.DomainID = session.Domain(0)

	if err := preset.Validate(); err != nil {
		return nil, err
	}

	return c.app.CreateSkillPreset(ctx, preset)
}

func (c *Controller) GetSkillPreset(ctx context.Context, query *model.GetSkillPresetQuery) (*model.SkillPreset, model.AppError) {
	session, err := c.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	permission := session.GetPermission(model.PermissionSkill)
	if !permission.CanRead() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	query.DomainID = session.Domain(0)

	if err := query.Validate(); err != nil {
		return nil, err
	}

	return c.app.GetSkillPreset(ctx, query)
}

func (c *Controller) SearchSkillPreset(ctx context.Context, query *model.SearchSkillPresetQuery) ([]*model.SkillPreset, model.AppError) {
	session, err := c.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	permission := session.GetPermission(model.PermissionSkill)
	if !permission.CanRead() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	query.DomainId = session.Domain(0)

	return c.app.SearchSkillPreset(ctx, query)
}

func (c *Controller) UpdateSkillPreset(ctx context.Context, cmd *model.SkillPreset) (*model.SkillPreset, model.AppError) {
	session, err := c.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	permission := session.GetPermission(model.PermissionSkill)
	if !permission.CanUpdate() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
	}

	cmd.UpdatedBy = &model.Lookup{Id: int(session.UserId)}
	cmd.DomainID = session.Domain(0)

	if err := cmd.Validate(); err != nil {
		return nil, err
	}

	return c.app.Store.SkillPreset().Update(ctx, cmd)
}

func (c *Controller) PatchSkillPreset(ctx context.Context, cmd *model.PatchSkillPresetCmd) (*model.SkillPreset, model.AppError) {
	session, err := c.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	permission := session.GetPermission(model.PermissionSkill)
	if !permission.CanUpdate() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
	}

	cmd.UpdatedBy = model.Lookup{Id: int(session.UserId)}
	cmd.DomainID = session.Domain(0)

	if err := cmd.Validate(); err != nil {
		return nil, err
	}

	return c.app.Store.SkillPreset().Patch(ctx, cmd)
}

func (c *Controller) DeleteSkillPreset(ctx context.Context, cmd *model.DeleteSkillPresetCmd) ([]*model.SkillPreset, model.AppError) {
	session, err := c.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	permission := session.GetPermission(model.PermissionSkill)
	if !permission.CanDelete() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_DELETE)
	}

	cmd.DomainID = session.Domain(0)

	return c.app.DeleteSkillPreset(ctx, cmd)
}
