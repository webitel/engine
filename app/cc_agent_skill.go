package app

import (
	"context"
	"github.com/webitel/engine/model"
)

func (app *App) CreateAgentSkill(ctx context.Context, as *model.AgentSkill) (*model.AgentSkill, *model.AppError) {
	return app.Store.AgentSkill().Create(ctx, as)
}

func (app *App) GetAgentsSkillPage(ctx context.Context, domainId, agentId int64, search *model.SearchAgentSkill) ([]*model.AgentSkill, bool, *model.AppError) {
	list, err := app.Store.AgentSkill().GetAllPage(ctx, domainId, agentId, search)
	if err != nil {
		return nil, false, err
	}
	search.RemoveLastElemIfNeed(&list)
	return list, search.EndOfList(), nil
}

func (app *App) GetAgentsSkillById(ctx context.Context, domainId, agentId, id int64) (*model.AgentSkill, *model.AppError) {
	return app.Store.AgentSkill().GetById(ctx, domainId, agentId, id)
}

func (app *App) UpdateAgentsSkill(ctx context.Context, agentSkill *model.AgentSkill) (*model.AgentSkill, *model.AppError) {
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

func (a *App) PatchAgentSkill(ctx context.Context, domainId int64, agentId, id int64, patch *model.AgentSkillPatch) (*model.AgentSkill, *model.AppError) {
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

func (a *App) RemoveAgentSkill(ctx context.Context, domainId, agentId, id int64) (*model.AgentSkill, *model.AppError) {
	agentSkill, err := a.Store.AgentSkill().GetById(ctx, domainId, agentId, id)

	if err != nil {
		return nil, err
	}

	err = a.Store.AgentSkill().Delete(ctx, agentId, id)
	if err != nil {
		return nil, err
	}
	return agentSkill, nil
}

func (app *App) LookupSkillIfNotExistsAgent(ctx context.Context, domainId, agentId int64, search *model.SearchAgentSkill) ([]*model.Skill, bool, *model.AppError) {
	list, err := app.Store.AgentSkill().LookupNotExistsAgent(ctx, domainId, agentId, search)
	if err != nil {
		return nil, false, err
	}
	search.RemoveLastElemIfNeed(&list)
	return list, search.EndOfList(), nil
}
