package model

import "time"

type PauseCause struct {
	AclRecord
	Name            string `json:"name" db:"name"`
	Description     string `json:"description" db:"description"`
	LimitMin        uint32 `json:"limit_min" db:"limit_min"`
	AllowSupervisor bool   `json:"allow_supervisor" db:"allow_supervisor"`
	AllowAdmin      bool   `json:"allow_admin" db:"allow_admin"`
	AllowAgent      bool   `json:"allow_agent" db:"allow_agent"`
}

type SearchPauseCause struct {
	ListRequest
	Ids  []uint32
	Name *string
}

type PauseCausePatch struct {
	UpdatedAt       *time.Time `json:"updated_at"`
	UpdatedBy       Lookup     `json:"updated_by"`
	Name            *string    `json:"name"`
	Description     *string    `json:"description"`
	LimitMin        *uint32    `json:"limit_min"`
	AllowSupervisor *bool      `json:"allow_supervisor"`
	AllowAgent      *bool      `json:"allow_agent"`
	AllowAdmin      *bool      `json:"allow_admin"`
}

func (p PauseCause) AllowFields() []string {
	return []string{"id", "created_by", "created_at", "updated_by", "updated_at", "name", "description", "limit_min", "allow_agent", "allow_supervisor", "allow_admin"}
}

func (PauseCause) DefaultOrder() string {
	return "-name"
}

func (PauseCause) DefaultFields() []string {
	return []string{"id", "name", "description", "limit_min", "allow_agent", "allow_supervisor", "allow_admin"}
}

func (PauseCause) EntityName() string {
	return "cc_pause_cause_list"
}

func (p *PauseCause) Patch(patch *PauseCausePatch) {
	p.UpdatedAt = patch.UpdatedAt
	p.UpdatedBy = &patch.UpdatedBy

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

	if patch.AllowAdmin != nil {
		p.AllowAdmin = *patch.AllowAdmin
	}

	if patch.LimitMin != nil {
		p.LimitMin = *patch.LimitMin
	}
}

// Todo
func (r *PauseCause) IsValid() *AppError {
	return nil
}
