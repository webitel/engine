package app

import (
	"context"
	"github.com/webitel/engine/model"
)

func (app *App) CreateAgentSkill(as *model.AgentSkill) (*model.AgentSkill, *model.AppError) {
	return app.Store.AgentSkill().Create(as)
}

func (app *App) CreateAgentSkills(ctx context.Context, domainId, agentId int64, skills []*model.AgentSkill) ([]int64, *model.AppError) {
	return app.Store.AgentSkill().BulkCreate(ctx, domainId, agentId, skills)
}

func (app *App) GetAgentsSkillPage(domainId, agentId int64, search *model.SearchAgentSkillList) ([]*model.AgentSkill, bool, *model.AppError) {
	list, err := app.Store.AgentSkill().GetAllPage(domainId, agentId, search)
	if err != nil {
		return nil, false, err
	}
	search.RemoveLastElemIfNeed(&list)
	return list, search.EndOfList(), nil
}

func (app *App) GetAgentsSkillById(domainId, agentId, id int64) (*model.AgentSkill, *model.AppError) {
	return app.Store.AgentSkill().GetById(domainId, agentId, id)
}

func (app *App) UpdateAgentsSkill(agentSkill *model.AgentSkill) (*model.AgentSkill, *model.AppError) {
	oldAgentSkill, err := app.GetAgentsSkillById(agentSkill.DomainId, int64(agentSkill.Agent.Id), agentSkill.Id)
	if err != nil {
		return nil, err
	}

	oldAgentSkill.Capacity = agentSkill.Capacity
	oldAgentSkill.Skill.Id = agentSkill.Skill.Id
	oldAgentSkill.Enabled = agentSkill.Enabled

	oldAgentSkill.UpdatedBy = agentSkill.UpdatedBy
	oldAgentSkill.UpdatedAt = model.GetMillis()

	oldAgentSkill, err = app.Store.AgentSkill().Update(oldAgentSkill)
	if err != nil {
		return nil, err
	}

	return oldAgentSkill, nil
}

func (app *App) UpdateAgentsSkills(ctx context.Context, domainId, agentId int64, search model.SearchAgentSkill, path model.AgentSkillPatch) ([]*model.AgentSkill, *model.AppError) {
	return app.Store.AgentSkill().UpdateMany(ctx, domainId, agentId, search, path)
}

func (a *App) PatchAgentSkill(domainId int64, agentId, id int64, patch *model.AgentSkillPatch) (*model.AgentSkill, *model.AppError) {
	oldAs, err := a.GetAgentsSkillById(domainId, agentId, id)
	if err != nil {
		return nil, err
	}

	oldAs.Patch(patch)

	if err = oldAs.IsValid(); err != nil {
		return nil, err
	}
	oldAs.DomainId = domainId
	oldAs, err = a.Store.AgentSkill().Update(oldAs)
	if err != nil {
		return nil, err
	}

	return oldAs, nil
}

func (a *App) RemoveAgentSkill(domainId, agentId, id int64) (*model.AgentSkill, *model.AppError) {
	agentSkill, err := a.Store.AgentSkill().GetById(domainId, agentId, id)

	if err != nil {
		return nil, err
	}

	err = a.Store.AgentSkill().Delete(agentId, id)
	if err != nil {
		return nil, err
	}
	return agentSkill, nil
}

func (a *App) RemoveAgentSkills(ctx context.Context, domainId, agentId int64, search model.SearchAgentSkill) ([]*model.AgentSkill, *model.AppError) {
	res, err := a.Store.AgentSkill().DeleteMany(ctx, domainId, agentId, search)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (app *App) LookupSkillIfNotExistsAgent(domainId, agentId int64, search *model.SearchAgentSkillList) ([]*model.Skill, bool, *model.AppError) {
	list, err := app.Store.AgentSkill().LookupNotExistsAgent(domainId, agentId, search)
	if err != nil {
		return nil, false, err
	}
	search.RemoveLastElemIfNeed(&list)
	return list, search.EndOfList(), nil
}
