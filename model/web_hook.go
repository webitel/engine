package model

import (
	"time"
)

type WebHook struct {
	Id  int32  `json:"id" db:"id"`
	Key string `json:"key" db:"key"`

	CreatedAt *time.Time `json:"created_at" db:"created_at"`
	UpdatedAt *time.Time `json:"updated_at" db:"updated_at"`
	CreatedBy *Lookup    `json:"created_by" db:"created_by"`
	UpdatedBy *Lookup    `json:"updated_by" db:"updated_by"`

	Name          string      `json:"name" db:"name"`
	Description   string      `json:"description" db:"description"`
	Origin        StringArray `json:"origin" db:"origin"`
	Schema        *Lookup     `json:"schema" db:"schema"`
	Enabled       bool        `json:"enabled" db:"enabled"`
	Authorization string      `json:"authorization" db:"authorization"`
}

type SearchWebHook struct {
	ListRequest
	Ids []int32
}

type WebHookPatch struct {
	UpdatedAt *time.Time `json:"updated_at" db:"updated_at"`
	UpdatedBy Lookup     `json:"updated_by" db:"updated_by"`

	Name          *string  `json:"name" db:"name"`
	Description   *string  `json:"description" db:"description"`
	Origin        []string `json:"origin" db:"origin"`
	Schema        *Lookup  `json:"schema" db:"schema"`
	Enabled       *bool    `json:"enabled" db:"enabled"`
	Authorization *string  `json:"authorization" db:"authorization"`
}

func (t WebHook) DefaultOrder() string {
	return "name"
}

func (t WebHook) AllowFields() []string {
	return []string{"id", "key", "created_at", "updated_at", "created_by", "updated_by", "name", "description", "origin",
		"schema", "enabled", "authorization"}
}

func (t WebHook) DefaultFields() []string {
	return []string{"id", "key", "name", "enabled", "schema"}
}

func (t WebHook) EntityName() string {
	return "web_hook_list"
}

func (t *WebHook) IsValid() AppError {
	/// TODO
	return nil
}

func (w *WebHook) Patch(p *WebHookPatch) {
	w.UpdatedBy = &p.UpdatedBy
	w.UpdatedAt = p.UpdatedAt

	if p.Name != nil {
		w.Name = *p.Name
	}
	if p.Description != nil {
		w.Description = *p.Description
	}
	if p.Origin != nil {
		w.Origin = p.Origin
	}
	if p.Schema != nil {
		w.Schema = p.Schema
	}
	if p.Enabled != nil {
		w.Enabled = *p.Enabled
	}
	if p.Authorization != nil {
		w.Authorization = *p.Authorization
	}
}
