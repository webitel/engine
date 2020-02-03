package model

type Skill struct {
	Id          int64  `json:"id" db:"id"`
	DomainId    int64  `json:"domain_id" db:"domain_id"`
	Name        string `json:"name" db:"name"`
	Description string `json:"description" db:"description"`
}

type SearchSkill struct {
	ListRequest
}

func (s *Skill) IsValid() *AppError {
	return nil
}
