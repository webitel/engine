package model

import "time"

type Skill struct {
	Id           int64      `json:"id" db:"id"`
	DomainId     int64      `json:"domain_id" db:"domain_id"`
	Name         string     `json:"name" db:"name"`
	Description  string     `json:"description" db:"description"`
	TotalAgents  *int32     `json:"total_agents" db:"total_agents"`
	ActiveAgents *int32     `json:"active_agents" db:"active_agents"`
	CreatedAt    *time.Time `json:"created_at" db:"created_at"`
	CreatedBy    *Lookup    `json:"created_by" db:"created_by"`
	UpdatedAt    *time.Time `json:"updated_at" db:"updated_at"`
	UpdatedBy    *Lookup    `json:"updated_by" db:"updated_by"`
}

type SearchSkill struct {
	ListRequest
	Ids            []uint32
	NotExistsAgent *int64
}

func (Skill) DefaultOrder() string {
	return "name"
}

func (a Skill) AllowFields() []string {
	return []string{"id", "domain_id", "name", "description", "active_agents", "total_agents",
		"updated_at", "created_at", "created_by", "updated_by"}
}

func (a Skill) DefaultFields() []string {
	return []string{"id", "name", "description", "active_agents", "total_agents", "created_by", "updated_by"}
}

func (a Skill) EntityName() string {
	return "cc_skill_view"
}

func (s *Skill) IsValid() AppError {
	return nil
}
