package model

import "time"

type AgentPauseCause struct {
	AclRecord
	Name            string `json:"name" db:"name"`
	Description     string `json:"description" db:"description"`
	LimitPerDay     uint32 `json:"limit_per_day" db:"limit_per_day"`
	AllowSupervisor bool   `json:"allow_supervisor" db:"allow_supervisor"`
	AllowAgent      bool   `json:"allow_agent" db:"allow_agent"`
}

type SearchAgentPauseCause struct {
	ListRequest
	Ids  []uint32
	Name *string
}

type AgentPauseCausePatch struct {
	UpdatedAt       *time.Time `json:"updated_at"`
	UpdatedBy       Lookup     `json:"updated_by"`
	Name            *string    `json:"name"`
	Description     *string    `json:"description"`
	LimitPerDay     *uint32    `json:"limit_per_day"`
	AllowSupervisor *bool      `json:"allow_supervisor"`
	AllowAgent      *bool      `json:"allow_agent"`
}

func (p AgentPauseCause) AllowFields() []string {
	return []string{"id", "created_by", "created_at", "updated_by", "updated_at", "name", "description", "limit_per_day", "allow_agent", "allow_supervisor"}
}

func (AgentPauseCause) DefaultOrder() string {
	return "-name"
}

func (AgentPauseCause) DefaultFields() []string {
	return []string{"id", "name", "description", "limit_per_day", "allow_agent", "allow_supervisor"}
}

func (AgentPauseCause) EntityName() string {
	return "cc_pause_cause_list"
}

func (p *AgentPauseCause) Patch(patch *AgentPauseCausePatch) {
	p.UpdatedAt = patch.UpdatedAt
	p.UpdatedBy = patch.UpdatedBy

	if patch.Name != nil {
		p.Name = *patch.Name
	}

	if patch.Description != nil {
		p.Description = *patch.Description
	}

	if patch.AllowAgent != nil {
		p.AllowAgent = *patch.AllowAgent
	}

	if patch.AllowSupervisor != nil {
		p.AllowSupervisor = *patch.AllowSupervisor
	}

	if patch.LimitPerDay != nil {
		p.LimitPerDay = *patch.LimitPerDay
	}
}

// Todo
func (r *AgentPauseCause) IsValid() *AppError {
	return nil
}
