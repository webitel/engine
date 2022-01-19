package model

import "time"

type QueueHook struct {
	Id        uint32    `json:"id" db:"id"`
	Schema    Lookup    `json:"schema" db:"schema"`
	Event     string    `json:"event" db:"event"`
	Enabled   bool      `json:"enabled" db:"enabled"`
	UpdatedBy *Lookup   `json:"updated_by" db:"updated_by"`
	CreatedBy *Lookup   `json:"created_by" db:"created_by"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type QueueHookPatch struct {
	UpdatedBy Lookup
	UpdatedAt time.Time

	Schema  *Lookup
	Event   *string
	Enabled *bool
}

type SearchQueueHook struct {
	ListRequest
	Ids       []uint32
	SchemaIds []uint32
	Events    []string
}

func (qh QueueHook) AllowFields() []string {
	return qh.DefaultFields()
}

func (qh QueueHook) DefaultOrder() string {
	return "+event"
}

func (QueueHook) DefaultFields() []string {
	return []string{"id", "schema", "event", "enabled"}
}

func (QueueHook) EntityName() string {
	return "cc_queue_events_list"
}

func (qh *QueueHook) IsValid() *AppError {
	//todo
	return nil
}

func (qh *QueueHook) Patch(patch *QueueHookPatch) {

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
