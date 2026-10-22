package app

import (
	"context"

	"github.com/webitel/engine/model"
	"github.com/webitel/engine/pkg/wbt/auth_manager"
)

func (app *App) SkillGroupCheckAccess(ctx context.Context, domainId, id int64, groups []int, access auth_manager.PermissionAccess) (bool, model.AppError) {
	return app.Store.SkillGroup().CheckAccess(ctx, domainId, id, groups, access)
}

func (app *App) CreateSkillGroup(ctx context.Context, group *model.SkillGroup) (*model.SkillGroup, model.AppError) {
	if exists, appErr := app.Store.SkillGroup().NameExists(ctx, group.DomainId, group.Name, 0); appErr != nil {
		return nil, appErr
	} else if exists {
		return nil, model.NewBadRequestError("app.skill_group.save.valid.name", "Skill group with this name already exists.")
	}
	return app.Store.SkillGroup().Create(ctx, group)
}

func (app *App) GetSkillGroup(ctx context.Context, id, domainId int64) (*model.SkillGroup, model.AppError) {
	return app.Store.SkillGroup().Get(ctx, domainId, id)
}

func (app *App) GetSkillGroupsPage(ctx context.Context, domainId int64, search *model.SearchSkillGroup) ([]*model.SkillGroup, bool, model.AppError) {
	list, err := app.Store.SkillGroup().GetAllPage(ctx, domainId, search)
	if err != nil {
		return nil, false, err
	}
	search.RemoveLastElemIfNeed(&list)
	return list, search.EndOfList(), nil
}

func (app *App) GetSkillGroupsPageByGroups(ctx context.Context, domainId int64, groups []int, search *model.SearchSkillGroup) ([]*model.SkillGroup, bool, model.AppError) {
	list, err := app.Store.SkillGroup().GetAllPageByGroups(ctx, domainId, groups, search)
	if err != nil {
		return nil, false, err
	}
	search.RemoveLastElemIfNeed(&list)
	return list, search.EndOfList(), nil
}

func (app *App) RemoveSkillGroup(ctx context.Context, domainId, id int64) (*model.SkillGroup, model.AppError) {
	group, err := app.Store.SkillGroup().Get(ctx, domainId, id)
	if err != nil {
		return nil, err
	}

	if err = app.Store.SkillGroup().Delete(ctx, domainId, id); err != nil {
		return nil, err
	}
	return group, nil
}

func (app *App) UpdateSkillGroup(ctx context.Context, group *model.SkillGroup) (*model.SkillGroup, model.AppError) {
	oldGroup, err := app.Store.SkillGroup().Get(ctx, group.DomainId, group.Id)
	if err != nil {
		return nil, err
	}

	if oldGroup.Name != group.Name {
		if exists, appErr := app.Store.SkillGroup().NameExists(ctx, group.DomainId, group.Name, group.Id); appErr != nil {
			return nil, appErr
		} else if exists {
			return nil, model.NewBadRequestError("app.skill_group.save.valid.name", "Skill group with this name already exists.")
		}
	}

	oldGroup.Name = group.Name
	oldGroup.Description = group.Description
	oldGroup.UpdatedBy = group.UpdatedBy
	oldGroup.UpdatedAt = group.UpdatedAt

	return app.Store.SkillGroup().Update(ctx, oldGroup)
}

func (app *App) SkillGroupSkillsSearch(ctx context.Context, domainId int64, search *model.SearchSkillGroupSkill) ([]*model.SkillGroupSkill, bool, model.AppError) {
	list, err := app.Store.SkillGroup().SkillsSearch(ctx, domainId, search)
	if err != nil {
		return nil, false, err
	}
	search.RemoveLastElemIfNeed(&list)
	return list, search.EndOfList(), nil
}

func (app *App) SkillGroupSkillsCreate(ctx context.Context, domainId, userId, groupId int64, skills []*model.SkillGroupSkill) ([]*model.SkillGroupSkill, model.AppError) {
	for _, sk := range skills {
		if appErr := sk.IsValid(); appErr != nil {
			return nil, appErr
		}
	}
	return app.Store.SkillGroup().SkillsCreate(ctx, domainId, userId, groupId, skills)
}

func (app *App) SkillGroupSkillUpdate(ctx context.Context, domainId, userId, groupId int64, skill *model.SkillGroupSkill) (*model.SkillGroupSkill, model.AppError) {
	if appErr := skill.IsValid(); appErr != nil {
		return nil, appErr
	}
	return app.Store.SkillGroup().SkillUpdate(ctx, domainId, userId, groupId, skill)
}

func (app *App) SkillGroupSkillsDelete(ctx context.Context, domainId, groupId int64, ids, skillIds []int64) ([]*model.SkillGroupSkill, model.AppError) {
	return app.Store.SkillGroup().SkillsDelete(ctx, domainId, groupId, ids, skillIds)
}

func (app *App) SkillGroupAgentsSearch(ctx context.Context, domainId int64, search *model.SearchSkillGroupAgent) ([]*model.SkillGroupAgent, bool, model.AppError) {
	list, err := app.Store.SkillGroup().AgentsSearch(ctx, domainId, search)
	if err != nil {
		return nil, false, err
	}
	search.RemoveLastElemIfNeed(&list)
	return list, search.EndOfList(), nil
}

func (app *App) SkillGroupAgentsCreate(ctx context.Context, domainId, userId int64, groupIds, agentIds []int64) ([]*model.SkillGroupAgent, model.AppError) {
	return app.Store.SkillGroup().AgentsCreate(ctx, domainId, userId, groupIds, agentIds)
}

func (app *App) SkillGroupAgentsDelete(ctx context.Context, domainId, groupId int64, ids, agentIds []int64) ([]*model.SkillGroupAgent, model.AppError) {
	return app.Store.SkillGroup().AgentsDelete(ctx, domainId, groupId, ids, agentIds)
}

func (app *App) AgentSkillGroupsSearch(ctx context.Context, domainId int64, search *model.SearchAgentSkillGroup) ([]*model.AgentSkillGroup, bool, model.AppError) {
	list, err := app.Store.SkillGroup().AgentGroupsSearch(ctx, domainId, search)
	if err != nil {
		return nil, false, err
	}
	search.RemoveLastElemIfNeed(&list)
	return list, search.EndOfList(), nil
}

func (app *App) AgentSkillGroupsDelete(ctx context.Context, domainId, agentId int64, ids, groupIds []int64) ([]*model.AgentSkillGroup, model.AppError) {
	return app.Store.SkillGroup().AgentGroupsDelete(ctx, domainId, agentId, ids, groupIds)
}
