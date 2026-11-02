package controller

import (
	"context"

	"github.com/webitel/engine/model"
	"github.com/webitel/engine/pkg/wbt/auth_manager"
)

func (c *Controller) CreateOnlineSkills(ctx context.Context, preset *model.OnlineSkills) (*model.OnlineSkills, model.AppError) {
	session, err := c.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	permission := session.GetPermission(model.PermissionSkill)
	if !permission.CanCreate() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_CREATE)
	}

	preset.InitializeCreateMetadata(session.Domain(0), session.UserId)

	if err := preset.Validate(); err != nil {
		return nil, err
	}

	return c.app.CreateOnlineSkills(ctx, preset)
}

func (c *Controller) GetOnlineSkills(ctx context.Context, query *model.GetSkillPresetQuery) (*model.OnlineSkills, model.AppError) {
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

	return c.app.GetOnlineSkills(ctx, query)
}

func (c *Controller) SearchOnlineSkills(ctx context.Context, query *model.SearchOnlineSkillsQuery) ([]*model.OnlineSkills, model.AppError) {
	session, err := c.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	permission := session.GetPermission(model.PermissionSkill)
	if !permission.CanRead() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	query.DomainId = session.Domain(0)

	return c.app.SearchOnlineSkills(ctx, query)
}

func (c *Controller) UpdateOnlineSkills(ctx context.Context, cmd *model.OnlineSkills) (*model.OnlineSkills, model.AppError) {
	session, err := c.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	permission := session.GetPermission(model.PermissionSkill)
	if !permission.CanUpdate() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
	}

	cmd.ActualizeUpdatorInfo(session.Domain(0), session.UserId)

	if err := cmd.Validate(); err != nil {
		return nil, err
	}

	return c.app.UpdateOnlineSkills(ctx, cmd)
}

func (c *Controller) PatchOnlineSkills(ctx context.Context, cmd *model.PatchOnlineSkillsCmd) (*model.OnlineSkills, model.AppError) {
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

	return c.app.PatchOnlineSkills(ctx, cmd)
}

func (c *Controller) DeleteOnlineSkills(ctx context.Context, cmd *model.DeleteSkillPresetCmd) model.AppError {
	session, err := c.app.GetSessionFromCtx(ctx)
	if err != nil {
		return err
	}

	permission := session.GetPermission(model.PermissionSkill)
	if !permission.CanDelete() {
		return c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_DELETE)
	}

	cmd.DomainID = session.Domain(0)

	return c.app.DeleteOnlineSkills(ctx, cmd)
}
