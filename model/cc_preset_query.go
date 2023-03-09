package model

import "time"

type PresetQuery struct {
	Id          int32           `json:"id" db:"id"`
	Name        string          `json:"name" db:"name"`
	Description string          `json:"description" db:"description"`
	Section     string          `json:"section" db:"section"`
	Preset      StringInterface `json:"preset" db:"preset"`
	CreatedAt   *time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt   *time.Time      `json:"updated_at" db:"updated_at"`
}

type PresetQueryPatch struct {
	UpdatedAt time.Time `json:"updated_at"`

	Name        *string         `json:"name" db:"name"`
	Description *string         `json:"description" db:"description"`
	Section     *string         `json:"section" db:"section"`
	Preset      StringInterface `json:"preset" db:"preset"`
}

type SearchPresetQuery struct {
	ListRequest
	Ids     []int32
	Section []string
}

func (p *PresetQuery) Patch(patch *PresetQueryPatch) {
	p.UpdatedAt = &patch.UpdatedAt

	if patch.Preset != nil {
		p.Preset = patch.Preset
	}

	if patch.Name != nil {
		p.Name = *patch.Name
	}

	if patch.Description != nil {
		p.Description = *patch.Description
	}
	if patch.Section != nil {
		p.Section = *patch.Section
	}
}

func (p PresetQuery) IsValid() *AppError {
	return nil
}

func (PresetQuery) DefaultOrder() string {
	return "name"
}

func (PresetQuery) AllowFields() []string {
	return []string{"id", "name", "description", "created_at", "updated_at", "section", "preset"}
}

func (PresetQuery) DefaultFields() []string {
	return []string{"id", "name", "preset"}
}

func (PresetQuery) EntityName() string {
	return "cc_preset_query_list"
}
