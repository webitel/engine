package model

import "time"

const (
	TriggerTypeCron = "cron"
)

type Trigger struct {
	Id int32 `json:"id" db:"id"`

	Name        string    `json:"name" json:"name"`
	Enabled     bool      `json:"enabled" db:"enabled"`
	Type        string    `json:"type" db:"type"`
	Schema      *Lookup   `json:"schema" db:"schema"`
	Variables   StringMap `json:"variables" db:"variables"`
	Description string    `json:"description" db:"description"`
	Expression  string    `json:"expression" db:"expression"`
	Timezone    *Lookup   `json:"timezone" db:"timezone"`
	Timeout     int32     `json:"timeout" db:"timeout"`

	CreatedAt *time.Time `json:"created_at" db:"created_at"`
	UpdatedAt *time.Time `json:"updated_at" db:"updated_at"`
	CreatedBy *Lookup    `json:"created_by" db:"created_by"`
	UpdatedBy *Lookup    `json:"updated_by" db:"updated_by"`
}

type SearchTrigger struct {
	ListRequest
	Ids []int32
}

func (t Trigger) DefaultOrder() string {
	return "id"
}

func (t Trigger) AllowFields() []string {
	return []string{"id", "name", "enabled", "type", "schema", "variables", "description", "expression",
		"timezone", "timeout", "created_at", "updated_at", "created_by", "updated_by"}
}

func (t Trigger) DefaultFields() []string {
	return []string{"id", "name", "enabled", "schema", "expression"}
}

func (t Trigger) EntityName() string {
	return "cc_trigger_list"
}

func (t *Trigger) IsValid() *AppError {
	if t.Type != TriggerTypeCron {
		//error
	}
	return nil
}

type TriggerPatch struct {
	UpdatedAt   *time.Time `json:"updated_at" db:"updated_at"`
	UpdatedBy   *Lookup    `json:"updated_by" db:"updated_by"`
	Name        *string    `json:"name" json:"name"`
	Enabled     *bool      `json:"enabled" db:"enabled"`
	Schema      *Lookup    `json:"schema" db:"schema"`
	Variables   StringMap  `json:"variables" db:"variables"`
	Description *string    `json:"description" db:"description"`
	Expression  *string    `json:"expression" db:"expression"`
	Timezone    *Lookup    `json:"timezone" db:"timezone"`
	Timeout     *int32     `json:"timeout" db:"timeout"`
}

func (t *Trigger) Patch(p *TriggerPatch) {
	t.UpdatedBy = p.UpdatedBy
	t.UpdatedAt = p.UpdatedAt

	if p.Name != nil {
		t.Name = *p.Name
	}
	if p.Enabled != nil {
		t.Enabled = *p.Enabled
	}
	if p.Schema != nil {
		t.Schema = p.Schema
	}
	if p.Variables != nil {
		t.Variables = p.Variables
	}
	if p.Description != nil {
		t.Description = *p.Description
	}
	if p.Expression != nil {
		t.Expression = *p.Expression
	}
	if p.Timezone != nil {
		t.Timezone = p.Timezone
	}
	if p.Timeout != nil {
		t.Timeout = *p.Timeout
	}
}
