package controller

import (
	"context"

	"github.com/webitel/engine/model"
	"github.com/webitel/engine/pkg/wbt/auth_manager"
)

func (c *Controller) CreateAgentsSkills(ctx context.Context, session *auth_manager.Session, items *model.AgentsSkills) ([]*model.AgentSkill, model.AppError) {
	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_AGENT)
	if !permission.CanRead() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}
	if !permission.CanUpdate() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
	}

	if session.UseRBAC(auth_manager.PERMISSION_ACCESS_UPDATE, permission) {
		var err model.AppError
		items.AgentIds, err = c.app.AccessAgentsIds(ctx, session.Domain(0), items.AgentIds, session.RoleIds, auth_manager.PERMISSION_ACCESS_UPDATE)
		if err != nil {
			return nil, err
		}
	}

	items.CreatedBy = &model.Lookup{
		Id: int(session.UserId),
	}
	items.UpdatedBy = items.CreatedBy
	items.CreatedAt = model.GetMillis()
	items.UpdatedAt = items.CreatedAt

	items.DomainId = session.Domain(0)

	return c.app.CreateAgentsSkills(ctx, items.DomainId, items)
}

func (c *Controller) GetAgentsSkillBySkill(ctx context.Context, session *auth_manager.Session, skillId int64, search *model.SearchAgentSkillList) ([]*model.AgentSkill, bool, bool, uint32, model.AppError) {
	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_AGENT)
	if !permission.CanRead() {
		return nil, false, false, 0, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}
	// TODO RBAC AGENTS

	list, next, err := c.app.GetAgentsSkillBySkill(ctx, session.Domain(0), skillId, search)
	if err != nil {
		return nil, false, false, 0, err
	}

	var (
		existsDisabled bool
		potentialRows uint32
	)
	existsDisabled, potentialRows, err = c.app.HasDisabledSkill(ctx, session.Domain(0), skillId, search.GetQ())
	if err != nil {
		return nil, false, false, 0, err
	}

	return list, next, existsDisabled, potentialRows, nil

}

func (c *Controller) PatchAgentsSkillBySkill(ctx context.Context, session *auth_manager.Session, skillId int64, search model.SearchAgentSkill, path model.AgentSkillPatch) ([]*model.AgentSkill, model.AppError) {
	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_AGENT)
	if !permission.CanRead() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}
	if !permission.CanUpdate() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
	}
	// TODO RBAC AGENTS

	path.UpdatedBy = model.Lookup{
		Id: int(session.UserId),
	}
	path.UpdatedAt = model.GetMillis()

	return c.app.PatchAgentsSkill(ctx, session.Domain(0), skillId, search, path)
}

func (c *Controller) DeleteAgentsSkill(ctx context.Context, session *auth_manager.Session, skillId int64, search model.SearchAgentSkill) ([]*model.AgentSkill, model.AppError) {
	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_AGENT)
	if !permission.CanRead() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}
	if !permission.CanUpdate() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
	}

	// TODO RBAC AGENTS

	return c.app.RemoveAgentsSkill(ctx, session.Domain(0), skillId, search)
}
