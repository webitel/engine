package model

type Skill struct {
	Id          int64  `json:"id" db:"id"`
	DomainId    int64  `json:"domain_id" db:"domain_id"`
	Name        string `json:"name" db:"name"`
	Description string `json:"description" db:"description"`
}

type SearchSkill struct {
	ListRequest
	Ids []uint32
}

func (Skill) DefaultOrder() string {
	return "id"
}

func (a Skill) AllowFields() []string {
	return []string{"id", "domain_id", "name", "description"}
}

func (a Skill) DefaultFields() []string {
	return []string{"id", "name", "description"}
}

func (a Skill) EntityName() string {
	return "cc_skill_view"
}

func (s *Skill) IsValid() *AppError {
	return nil
}
