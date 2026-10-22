package model

import "time"

type SkillGroup struct {
	Id           int64      `json:"id" db:"id"`
	DomainId     int64      `json:"domain_id" db:"domain_id"`
	Name         string     `json:"name" db:"name"`
	Description  string     `json:"description" db:"description"`
	Skills       []*Lookup  `json:"skills" db:"skills"`
	TotalAgents  *int32     `json:"total_agents" db:"total_agents"`
	ActiveAgents *int32     `json:"active_agents" db:"active_agents"`
	CreatedAt    *time.Time `json:"created_at" db:"created_at"`
	CreatedBy    *Lookup    `json:"created_by" db:"created_by"`
	UpdatedAt    *time.Time `json:"updated_at" db:"updated_at"`
	UpdatedBy    *Lookup    `json:"updated_by" db:"updated_by"`
}

type SearchSkillGroup struct {
	ListRequest
	Ids []uint32
}

func (SkillGroup) DefaultOrder() string {
	return "name"
}

func (SkillGroup) AllowFields() []string {
	return []string{
		"id", "domain_id", "name", "description", "skills", "active_agents", "total_agents",
		"updated_at", "created_at", "created_by", "updated_by",
	}
}

func (SkillGroup) DefaultFields() []string {
	return []string{"id", "name", "description", "skills", "active_agents", "total_agents", "created_by", "updated_by"}
}

func (SkillGroup) EntityName() string {
	return "cc_skill_group_view"
}

func (s SkillGroup) IsValid() AppError {
	if s.Name == "" {
		return NewBadRequestError("model.skill_group.is_valid.name.app_error", "name is required")
	}
	return nil
}

type SkillGroupSkill struct {
	Id        int64   `json:"id" db:"id"`
	Skill     *Lookup `json:"skill" db:"skill"`
	Capacity  *int    `json:"capacity" db:"capacity"`
	CreatedAt int64   `json:"created_at" db:"created_at"`
	CreatedBy *Lookup `json:"created_by" db:"created_by"`
	UpdatedAt int64   `json:"updated_at" db:"updated_at"`
	UpdatedBy *Lookup `json:"updated_by" db:"updated_by"`
}

func (s *SkillGroupSkill) IsValid() AppError {
	if s.Skill.GetSafeId() == nil {
		return NewBadRequestError("model.skill_group_skill.is_valid.skill.app_error", "skill is required")
	}
	if s.Capacity == nil || *s.Capacity < 0 || *s.Capacity > 100 {
		return NewBadRequestError("model.skill_group_skill.is_valid.capacity.app_error", "capacity must be between 0 and 100")
	}
	return nil
}

type SearchSkillGroupSkill struct {
	ListRequest
	SkillGroupId int64
	Ids          []int64
}

type SkillGroupAgent struct {
	Id        int64   `json:"id" db:"id"`
	Agent     *Lookup `json:"agent" db:"agent"`
	Team      *Lookup `json:"team" db:"team"`
	CreatedAt int64   `json:"created_at" db:"created_at"`
	CreatedBy *Lookup `json:"created_by" db:"created_by"`
}

type SearchSkillGroupAgent struct {
	ListRequest
	SkillGroupId int64
	Ids          []int64
	AgentIds     []int64
}

type AgentSkillGroup struct {
	Id          int64     `json:"id" db:"id"`
	SkillGroup  *Lookup   `json:"skill_group" db:"skill_group"`
	Description string    `json:"description" db:"description"`
	Skills      []*Lookup `json:"skills" db:"skills"`
}

type SearchAgentSkillGroup struct {
	ListRequest
	AgentId int64
	Ids     []int64
}
