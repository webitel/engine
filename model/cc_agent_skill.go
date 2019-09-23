package model

type AgentSkill struct {
	DomainRecord
	//Id       int64  `json:"id" db:"id"`
	Agent    Lookup `json:"agent" json:"agent"`
	Skill    Lookup `json:"skill" db:"skill"`
	Capacity int    `json:"capacity" db:"capacity"`
}

func (as *AgentSkill) IsValid() *AppError {
	//FIXME
	return nil
}
