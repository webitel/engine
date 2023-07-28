package model

type Skill struct {
	Id           int64  `json:"id" db:"id"`
	DomainId     int64  `json:"domain_id" db:"domain_id"`
	Name         string `json:"name" db:"name"`
	Description  string `json:"description" db:"description"`
	TotalAgents  *int32 `json:"total_agents" db:"total_agents"`
	ActiveAgents *int32 `json:"active_agents" db:"active_agents"`
}

type SearchSkill struct {
	ListRequest
	Ids []uint32
}

func (Skill) DefaultOrder() string {
	return "id"
}

func (a Skill) AllowFields() []string {
	return []string{"id", "domain_id", "name", "description", "active_agents", "total_agents"}
}

func (a Skill) DefaultFields() []string {
	return []string{"id", "name", "description", "active_agents", "total_agents"}
}

func (a Skill) EntityName() string {
	return "cc_skill_view"
}

func (s *Skill) IsValid() AppError {
	return nil
}
