package controller

import (
	"context"
	"time"

	"github.com/webitel/engine/model"
	"github.com/webitel/engine/pkg/wbt/auth_manager"
)

func (c *Controller) CreateSkillGroup(ctx context.Context, session *auth_manager.Session, g *model.SkillGroup) (*model.SkillGroup, model.AppError) {
	permission := session.GetPermission(model.PermissionSkillGroup)
	if !permission.CanCreate() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_CREATE)
	}

	t := time.Now()
	g.CreatedAt = &t
	g.UpdatedAt = g.CreatedAt
	g.CreatedBy = &model.Lookup{Id: int(session.UserId)}
	g.UpdatedBy = g.CreatedBy

	if err := g.IsValid(); err != nil {
		return nil, err
	}

	return c.app.CreateSkillGroup(ctx, g)
}

func (c *Controller) SearchSkillGroup(ctx context.Context, session *auth_manager.Session, search *model.SearchSkillGroup) ([]*model.SkillGroup, bool, model.AppError) {
	permission := session.GetPermission(model.PermissionSkillGroup)
	if !permission.CanRead() {
		return nil, false, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if session.UseRBAC(auth_manager.PERMISSION_ACCESS_READ, permission) {
		return c.app.GetSkillGroupsPageByGroups(ctx, session.Domain(search.DomainId), session.GetAclRoles(), search)
	}
	return c.app.GetSkillGroupsPage(ctx, session.Domain(search.DomainId), search)
}

func (c *Controller) ReadSkillGroup(ctx context.Context, session *auth_manager.Session, id int64) (*model.SkillGroup, model.AppError) {
	permission := session.GetPermission(model.PermissionSkillGroup)
	if !permission.CanRead() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if err := c.skillGroupRecordAccess(ctx, session, id, auth_manager.PERMISSION_ACCESS_READ, permission); err != nil {
		return nil, err
	}

	return c.app.GetSkillGroup(ctx, id, session.Domain(0))
}

func (c *Controller) UpdateSkillGroup(ctx context.Context, session *auth_manager.Session, g *model.SkillGroup) (*model.SkillGroup, model.AppError) {
	permission := session.GetPermission(model.PermissionSkillGroup)
	if !permission.CanUpdate() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
	}

	if err := c.skillGroupRecordAccess(ctx, session, g.Id, auth_manager.PERMISSION_ACCESS_UPDATE, permission); err != nil {
		return nil, err
	}

	t := time.Now()
	g.UpdatedAt = &t
	g.UpdatedBy = &model.Lookup{Id: int(session.UserId)}

	if err := g.IsValid(); err != nil {
		return nil, err
	}

	return c.app.UpdateSkillGroup(ctx, g)
}

func (c *Controller) DeleteSkillGroup(ctx context.Context, session *auth_manager.Session, id int64) (*model.SkillGroup, model.AppError) {
	permission := session.GetPermission(model.PermissionSkillGroup)
	if !permission.CanDelete() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_DELETE)
	}

	if err := c.skillGroupRecordAccess(ctx, session, id, auth_manager.PERMISSION_ACCESS_DELETE, permission); err != nil {
		return nil, err
	}

	return c.app.RemoveSkillGroup(ctx, session.Domain(0), id)
}

func (c *Controller) SearchSkillGroupSkills(ctx context.Context, session *auth_manager.Session, search *model.SearchSkillGroupSkill) ([]*model.SkillGroupSkill, bool, model.AppError) {
	permission := session.GetPermission(model.PermissionSkillGroup)
	if !permission.CanRead() {
		return nil, false, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}
	if err := c.skillGroupRecordAccess(ctx, session, search.SkillGroupId, auth_manager.PERMISSION_ACCESS_READ, permission); err != nil {
		return nil, false, err
	}
	return c.app.SkillGroupSkillsSearch(ctx, session.Domain(0), search)
}

func (c *Controller) CreateSkillGroupSkills(ctx context.Context, session *auth_manager.Session, groupId int64, skills []*model.SkillGroupSkill) ([]*model.SkillGroupSkill, model.AppError) {
	if err := c.skillGroupUpdateGuard(ctx, session, groupId); err != nil {
		return nil, err
	}
	return c.app.SkillGroupSkillsCreate(ctx, session.Domain(0), session.UserId, groupId, skills)
}

func (c *Controller) UpdateSkillGroupSkill(ctx context.Context, session *auth_manager.Session, groupId int64, skill *model.SkillGroupSkill) (*model.SkillGroupSkill, model.AppError) {
	if err := c.skillGroupUpdateGuard(ctx, session, groupId); err != nil {
		return nil, err
	}
	return c.app.SkillGroupSkillUpdate(ctx, session.Domain(0), session.UserId, groupId, skill)
}

func (c *Controller) DeleteSkillGroupSkills(ctx context.Context, session *auth_manager.Session, groupId int64, ids, skillIds []int64) ([]*model.SkillGroupSkill, model.AppError) {
	if err := c.skillGroupUpdateGuard(ctx, session, groupId); err != nil {
		return nil, err
	}
	return c.app.SkillGroupSkillsDelete(ctx, session.Domain(0), groupId, ids, skillIds)
}

func (c *Controller) SearchSkillGroupAgents(ctx context.Context, session *auth_manager.Session, search *model.SearchSkillGroupAgent) ([]*model.SkillGroupAgent, bool, model.AppError) {
	permission := session.GetPermission(model.PermissionSkillGroup)
	if !permission.CanRead() {
		return nil, false, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}
	if err := c.skillGroupRecordAccess(ctx, session, search.SkillGroupId, auth_manager.PERMISSION_ACCESS_READ, permission); err != nil {
		return nil, false, err
	}
	return c.app.SkillGroupAgentsSearch(ctx, session.Domain(0), search)
}

func (c *Controller) CreateSkillGroupAgents(ctx context.Context, session *auth_manager.Session, groupIds, agentIds []int64) ([]*model.SkillGroupAgent, model.AppError) {
	permission := session.GetPermission(model.PermissionSkillGroup)
	if !permission.CanUpdate() {
		return nil, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
	}
	if session.UseRBAC(auth_manager.PERMISSION_ACCESS_UPDATE, permission) {
		for _, id := range groupIds {
			if err := c.skillGroupRecordAccess(ctx, session, id, auth_manager.PERMISSION_ACCESS_UPDATE, permission); err != nil {
				return nil, err
			}
		}
	}
	return c.app.SkillGroupAgentsCreate(ctx, session.Domain(0), session.UserId, groupIds, agentIds)
}

func (c *Controller) DeleteSkillGroupAgents(ctx context.Context, session *auth_manager.Session, groupId int64, ids, agentIds []int64) ([]*model.SkillGroupAgent, model.AppError) {
	if err := c.skillGroupUpdateGuard(ctx, session, groupId); err != nil {
		return nil, err
	}
	return c.app.SkillGroupAgentsDelete(ctx, session.Domain(0), groupId, ids, agentIds)
}

func (c *Controller) SearchAgentSkillGroups(ctx context.Context, session *auth_manager.Session, search *model.SearchAgentSkillGroup) ([]*model.AgentSkillGroup, bool, model.AppError) {
	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_AGENT)
	if !permission.CanRead() {
		return nil, false, c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}
	if err := c.agentRecordAccess(ctx, session, search.AgentId, auth_manager.PERMISSION_ACCESS_READ, permission); err != nil {
		return nil, false, err
	}
	return c.app.AgentSkillGroupsSearch(ctx, session.Domain(0), search)
}

func (c *Controller) CreateAgentSkillGroups(ctx context.Context, session *auth_manager.Session, agentId int64, groupIds []int64) ([]*model.AgentSkillGroup, model.AppError) {
	if err := c.agentUpdateGuard(ctx, session, agentId); err != nil {
		return nil, err
	}
	if _, err := c.app.SkillGroupAgentsCreate(ctx, session.Domain(0), session.UserId, groupIds, []int64{agentId}); err != nil {
		return nil, err
	}
	list, _, err := c.app.AgentSkillGroupsSearch(ctx, session.Domain(0), &model.SearchAgentSkillGroup{AgentId: agentId})
	return list, err
}

func (c *Controller) DeleteAgentSkillGroups(ctx context.Context, session *auth_manager.Session, agentId int64, ids, groupIds []int64) ([]*model.AgentSkillGroup, model.AppError) {
	if err := c.agentUpdateGuard(ctx, session, agentId); err != nil {
		return nil, err
	}
	return c.app.AgentSkillGroupsDelete(ctx, session.Domain(0), agentId, ids, groupIds)
}

func (c *Controller) skillGroupRecordAccess(ctx context.Context, session *auth_manager.Session, id int64, access auth_manager.PermissionAccess, permission auth_manager.SessionPermission) model.AppError {
	if !session.UseRBAC(access, permission) {
		return nil
	}
	if perm, err := c.app.SkillGroupCheckAccess(ctx, session.Domain(0), id, session.GetAclRoles(), access); err != nil {
		return err
	} else if !perm {
		return c.app.MakeResourcePermissionError(session, id, permission, access)
	}
	return nil
}

func (c *Controller) skillGroupUpdateGuard(ctx context.Context, session *auth_manager.Session, id int64) model.AppError {
	permission := session.GetPermission(model.PermissionSkillGroup)
	if !permission.CanUpdate() {
		return c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
	}
	return c.skillGroupRecordAccess(ctx, session, id, auth_manager.PERMISSION_ACCESS_UPDATE, permission)
}

func (c *Controller) agentRecordAccess(ctx context.Context, session *auth_manager.Session, agentId int64, access auth_manager.PermissionAccess, permission auth_manager.SessionPermission) model.AppError {
	if !session.UseRBAC(access, permission) {
		return nil
	}
	if perm, err := c.app.AgentCheckAccess(ctx, session.Domain(0), agentId, session.GetAclRoles(), access); err != nil {
		return err
	} else if !perm {
		return c.app.MakeResourcePermissionError(session, agentId, permission, access)
	}
	return nil
}

func (c *Controller) agentUpdateGuard(ctx context.Context, session *auth_manager.Session, agentId int64) model.AppError {
	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_AGENT)
	if !permission.CanUpdate() {
		return c.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_UPDATE)
	}
	return c.agentRecordAccess(ctx, session, agentId, auth_manager.PERMISSION_ACCESS_UPDATE, permission)
}
