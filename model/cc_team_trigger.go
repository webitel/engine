package model

import "time"

type TeamTrigger struct {
	Id          uint32     `json:"id" db:"id"`
	Schema      *Lookup    `json:"schema" db:"schema"`
	Enabled     bool       `json:"enabled" db:"enabled"`
	Name        string     `json:"name" db:"name"`
	Description string     `json:"description" db:"description"`
	UpdatedBy   *Lookup    `json:"updated_by" db:"updated_by"`
	CreatedBy   *Lookup    `json:"created_by" db:"created_by"`
	UpdatedAt   *time.Time `json:"updated_at" db:"updated_at"`
	CreatedAt   *time.Time `json:"created_at" db:"created_at"`
}

type TeamTriggerPatch struct {
	UpdatedBy Lookup
	UpdatedAt time.Time

	Schema      *Lookup
	Enabled     *bool
	Name        *string
	Description *string
}

type SearchTeamTrigger struct {
	ListRequest
	Ids       []uint32
	SchemaIds []uint32
	Enabled   *bool
}

func (qt TeamTrigger) AllowFields() []string {
	return qt.DefaultFields()
}

func (qt TeamTrigger) DefaultOrder() string {
	return "+name"
}

func (TeamTrigger) DefaultFields() []string {
	return []string{"id", "schema", "name", "enabled"}
}

func (TeamTrigger) EntityName() string {
	return "cc_team_trigger_list"
}

func (qt *TeamTrigger) IsValid() AppError {
	//todo
	return nil
}

func (qt *TeamTrigger) Patch(patch *TeamTriggerPatch) {

	if patch.Enabled != nil {
		qt.Enabled = *patch.Enabled
	}

	if patch.Name != nil {
		qt.Name = *patch.Name
	}

	if patch.Description != nil {
		qt.Description = *patch.Description
	}

	if patch.Schema != nil {
		qt.Schema = patch.Schema
	}
}
