package model

type AgentSkill struct {
	DomainRecord
	//Id       int64  `json:"id" db:"id"`
	Agent    Lookup `json:"agent" json:"agent"`
	Skill    Lookup `json:"skill" db:"skill"`
	Capacity int    `json:"capacity" db:"capacity"`
	Enabled  bool   `json:"enabled" db:"enabled"`
}

type AgentSkillPatch struct {
	UpdatedAt int64
	UpdatedBy Lookup

	Agent    *Lookup
	Skill    *Lookup
	Capacity *int
	Enabled  *bool
}

type SearchAgentSkill struct {
	ListRequest
}

func (as *AgentSkill) Patch(patch *AgentSkillPatch) {
	as.UpdatedBy = patch.UpdatedBy
	as.UpdatedAt = patch.UpdatedAt

	if patch.Agent != nil {
		as.Agent = *patch.Agent
	}

	if patch.Skill != nil {
		as.Skill = *patch.Skill
	}

	if patch.Capacity != nil {
		as.Capacity = *patch.Capacity
	}

	if patch.Enabled != nil {
		as.Enabled = *patch.Enabled
	}
}

func (as *AgentSkill) IsValid() *AppError {
	//FIXME
	return nil
}
