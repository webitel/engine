package model

import (
	"fmt"
	"time"
)

const (
	TriggerTypeCron  = "cron"
	TriggerTypeEvent = "event"
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
	Object    string     `json:"object" db:"object"`
	Event     string     `json:"event" db:"event"`
}

type TriggerWithDomainID struct {
	Trigger
	DomainId int64 `json:"domain_id" db:"domain_id"`
}

type TriggerJob struct {
	Id         int64      `json:"id" db:"id"`
	Trigger    Lookup     `json:"trigger" db:"trigger"`
	State      int        `json:"state" db:"state"`
	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
	StartedAt  *time.Time `json:"started_at" db:"started_at"`
	StoppedAt  *time.Time `json:"stopped_at" db:"stopped_at"`
	Parameters []byte     `json:"parameters" db:"parameters"`
	Error      *string    `json:"error" db:"error"`
	Result     []byte     `json:"result" db:"result"`
}

type SearchTrigger struct {
	ListRequest
	Ids []int32
}

type SearchTriggerJob struct {
	ListRequest
	CreatedAt *FilterBetween
	StartedAt *FilterBetween
	State     []int
	Duration  *FilterBetween
}

func (t Trigger) DefaultOrder() string {
	return "id"
}

func (t Trigger) AllowFields() []string {
	return []string{"id", "name", "enabled", "type", "schema", "variables", "description", "expression",
		"timezone", "timeout", "created_at", "updated_at", "created_by", "updated_by", "object", "event"}
}

func (t Trigger) AllowFieldsWithDomainId() []string {
	return append(t.AllowFields(), "domain_id")
}

func (t Trigger) DefaultFields() []string {
	return []string{"id", "name", "type", "enabled", "schema", "expression"}
}

func (t Trigger) EntityName() string {
	return "cc_trigger_list"
}

func (t *Trigger) IsValid() AppError {
	switch t.Type {
	case TriggerTypeCron:
		if len(t.Expression) == 0 {
			return NewBadRequestError("trigger.validation.expression", "expression is required")
		}
		t.Object = ""
		t.Event = ""
	case TriggerTypeEvent:
		if len(t.Object) == 0 {
			return NewBadRequestError("trigger.validation.object", "object is required")
		}
		if len(t.Event) == 0 {
			return NewBadRequestError("trigger.validation.event", "event is required")
		}
		t.Expression = ""
	default:
		return newAppError("trigger.validation.invalid_type", fmt.Sprintf("invalid trigger type: %s", t.Type))
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
	Object      *string    `json:"object" db:"object"`
	Event       *string    `json:"event" db:"event"`
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
	if p.Object != nil {
		t.Object = *p.Object
	}
	if p.Event != nil {
		t.Event = *p.Event
	}
}

func (t TriggerJob) DefaultOrder() string {
	return "started_at"
}

func (t TriggerJob) AllowFields() []string {
	return t.DefaultFields()
}

func (t TriggerJob) DefaultFields() []string {
	return []string{"id", "trigger", "state", "created_at", "started_at", "stopped_at", "parameters", "error", "result"}
}

func (t TriggerJob) EntityName() string {
	return "cc_trigger_job_list"
}
