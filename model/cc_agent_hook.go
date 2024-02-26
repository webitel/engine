package model

import "time"

type AgentHook struct {
	Id        int32     `json:"id" db:"id"`
	Schema    Lookup    `json:"schema" db:"schema"`
	Event     string    `json:"event" db:"event"`
	Enabled   bool      `json:"enabled" db:"enabled"`
	UpdatedBy *Lookup   `json:"updated_by" db:"updated_by"`
	CreatedBy *Lookup   `json:"created_by" db:"created_by"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type AgentHookPatch struct {
	UpdatedBy Lookup
	UpdatedAt time.Time

	Schema  *Lookup
	Event   *string
	Enabled *bool
}

type SearchAgentHook struct {
	ListRequest
	Ids       []uint32
	SchemaIds []uint32
	Events    []string
}

func (qh AgentHook) AllowFields() []string {
	return qh.DefaultFields()
}

func (qh AgentHook) DefaultOrder() string {
	return "+event"
}

func (AgentHook) DefaultFields() []string {
	return []string{"id", "schema", "event", "enabled"}
}

func (AgentHook) EntityName() string {
	return "cc_agent_events_list"
}

func (qh *AgentHook) IsValid() AppError {
	//todo
	return nil
}

func (qh *AgentHook) Patch(patch *AgentHookPatch) {

	if patch.Event != nil {
		qh.Event = *patch.Event
	}

	if patch.Schema != nil {
		qh.Schema = *patch.Schema
	}

	if patch.Enabled != nil {
		qh.Enabled = *patch.Enabled
	}
}
