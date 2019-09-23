package app

import "github.com/webitel/engine/model"

func (app *App) CreateAgentSkill(as *model.AgentSkill) (*model.AgentSkill, *model.AppError) {
	return app.Store.AgentSkill().Create(as)
}

func (app *App) GetAgentsSkillPage(domainId, agentId int64, page, perPage int) ([]*model.AgentSkill, *model.AppError) {
	return app.Store.AgentSkill().GetAllPage(domainId, agentId, page*perPage, perPage)
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

	oldAgentSkill.UpdatedBy.Id = agentSkill.UpdatedBy.Id
	oldAgentSkill.UpdatedAt = model.GetMillis()

	oldAgentSkill, err = app.Store.AgentSkill().Update(oldAgentSkill)
	if err != nil {
		return nil, err
	}

	return oldAgentSkill, nil
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
