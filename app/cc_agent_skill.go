package app

import (
	"context"

	"github.com/webitel/engine/model"
)

func (app *App) CreateAgentSkill(ctx context.Context, as *model.AgentSkill) (*model.AgentSkill, model.AppError) {
	return app.Store.AgentSkill().Create(ctx, as)
}

func (app *App) CreateAgentSkills(ctx context.Context, domainId, agentId int64, skills []*model.AgentSkill) ([]int64, model.AppError) {
	return app.Store.AgentSkill().BulkCreate(ctx, domainId, agentId, skills)
}

func (app *App) GetAgentsSkillPage(ctx context.Context, domainId, agentId int64, search *model.SearchAgentSkillList) ([]*model.AgentSkill, bool, model.AppError) {
	search.AgentIds = []int64{agentId}
	list, err := app.Store.AgentSkill().GetAllPage(ctx, domainId, search)
	if err != nil {
		return nil, false, err
	}
	search.RemoveLastElemIfNeed(&list)
	return list, search.EndOfList(), nil
}

func (app *App) GetAgentsSkillById(ctx context.Context, domainId, agentId, id int64) (*model.AgentSkill, model.AppError) {
	return app.Store.AgentSkill().GetById(ctx, domainId, agentId, id)
}

func (app *App) UpdateAgentsSkill(ctx context.Context, agentSkill *model.AgentSkill) (*model.AgentSkill, model.AppError) {
	oldAgentSkill, err := app.GetAgentsSkillById(ctx, agentSkill.DomainId, int64(agentSkill.Agent.Id), agentSkill.Id)
	if err != nil {
		return nil, err
	}

	oldAgentSkill.Capacity = agentSkill.Capacity
	oldAgentSkill.Skill.Id = agentSkill.Skill.Id
	oldAgentSkill.Enabled = agentSkill.Enabled

	oldAgentSkill.UpdatedBy = agentSkill.UpdatedBy
	oldAgentSkill.UpdatedAt = model.GetMillis()

	oldAgentSkill, err = app.Store.AgentSkill().Update(ctx, oldAgentSkill)
	if err != nil {
		return nil, err
	}

	return oldAgentSkill, nil
}

func (app *App) UpdateAgentsSkills(ctx context.Context, domainId, agentId int64, search model.SearchAgentSkill, path model.AgentSkillPatch) ([]*model.AgentSkill, model.AppError) {
	search.AgentIds = []int64{agentId}
	return app.Store.AgentSkill().UpdateMany(ctx, domainId, search, path)
}

func (a *App) PatchAgentSkill(ctx context.Context, domainId int64, agentId, id int64, patch *model.AgentSkillPatch) (*model.AgentSkill, model.AppError) {
	oldAs, err := a.GetAgentsSkillById(ctx, domainId, agentId, id)
	if err != nil {
		return nil, err
	}

	oldAs.Patch(patch)

	if err = oldAs.IsValid(); err != nil {
		return nil, err
	}
	oldAs.DomainId = domainId
	oldAs, err = a.Store.AgentSkill().Update(ctx, oldAs)
	if err != nil {
		return nil, err
	}

	return oldAs, nil
}

func (a *App) RemoveAgentSkill(ctx context.Context, domainId, agentId, id int64) (*model.AgentSkill, model.AppError) {
	agentSkill, err := a.Store.AgentSkill().GetById(ctx, domainId, agentId, id)

	if err != nil {
		return nil, err
	}

	err = a.Store.AgentSkill().DeleteById(ctx, agentId, id)
	if err != nil {
		return nil, err
	}
	return agentSkill, nil
}

func (a *App) RemoveAgentSkills(ctx context.Context, domainId, agentId int64, search model.SearchAgentSkill) ([]*model.AgentSkill, model.AppError) {
	search.AgentIds = []int64{agentId}
	res, err := a.Store.AgentSkill().Delete(ctx, domainId, search)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (app *App) LookupSkillIfNotExistsAgent(ctx context.Context, domainId, agentId int64, search *model.SearchAgentSkillList) ([]*model.Skill, bool, model.AppError) {
	list, err := app.Store.AgentSkill().LookupNotExistsAgent(ctx, domainId, agentId, search)
	if err != nil {
		return nil, false, err
	}
	search.RemoveLastElemIfNeed(&list)
	return list, search.EndOfList(), nil
}

func (app *App) CreateAgentsSkills(ctx context.Context, domainId int64, items *model.AgentsSkills) ([]*model.AgentSkill, model.AppError) {
	return app.Store.AgentSkill().CreateMany(ctx, domainId, items)
}

func (app *App) GetAgentsSkillBySkill(ctx context.Context, domainId int64, skillId int64, search *model.SearchAgentSkillList) ([]*model.AgentSkill, bool, model.AppError) {
	search.SkillIds = []int64{skillId}
	list, err := app.Store.AgentSkill().GetAllPage(ctx, domainId, search)
	if err != nil {
		return nil, false, err
	}
	search.RemoveLastElemIfNeed(&list)
	return list, search.EndOfList(), nil
}

func (app *App) PatchAgentsSkill(ctx context.Context, domainId int64, skillId int64, search model.SearchAgentSkill, path model.AgentSkillPatch) ([]*model.AgentSkill, model.AppError) {
	search.SkillIds = []int64{skillId}

	err := path.IsValid()
	if err != nil {
		return nil, err
	}

	return app.Store.AgentSkill().UpdateMany(ctx, domainId, search, path)
}

func (app *App) RemoveAgentsSkill(ctx context.Context, domainId int64, skillId int64, search model.SearchAgentSkill) ([]*model.AgentSkill, model.AppError) {
	search.SkillIds = []int64{skillId}
	return app.Store.AgentSkill().Delete(ctx, domainId, search)
}

func (app *App) HasDisabledSkill(ctx context.Context, domainId int64, skillId int64) (bool, model.AppError) {
	return app.Store.AgentSkill().HasDisabledSkill(ctx, domainId, skillId)
}
