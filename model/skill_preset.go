package model

import (
	"cmp"
	"strings"
	"time"
)

const StandartSkillPreset string = "Standart Online"

var StandartSkillPresetValue = &SkillPreset{
	Name: StandartSkillPreset,
}

type SkillPreset struct {
	ID          int64     `json:"id" db:"id"`
	DomainID    int64     `json:"domain_id" db:"domain_id"`
	CreatedBy   *Lookup   `json:"created_by" db:"created_by"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedBy   *Lookup   `json:"updated_by" db:"updated_by"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
	Name        string    `json:"name" db:"name"`
	Description *string   `json:"description" db:"description"`

	Skills []*Lookup `json:"skills" db:"skills"`
}

func (s *SkillPreset) GetDescription() string {
	return *(cmp.Or(s.Description, new(string)))
}

func (s *SkillPreset) UpdatedAtUnix() int64 {
	if s == nil {
		return 0
	}

	return max(s.UpdatedAt.UTC().UnixMilli(), 0)
}

func (s *SkillPreset) CreatedAtUnix() int64 {
	if s == nil {
		return 0
	}

	return max(s.CreatedAt.UTC().UnixMilli(), 0)
}

func (s *SkillPreset) ReduceSkillsIDs() []int64 {
	skillIDs := make([]int64, 0, len(s.Skills))

	for _, s := range s.Skills {
		skillIDs = append(skillIDs, int64(s.Id))
	}

	return skillIDs
}

func (s SkillPreset) AllowFields() []string {
	return []string{"id", "name", "description", "created_by", "created_at", "updated_by", "updated_at", "skills"}
}

func (s SkillPreset) DefaultFields() []string {
	return []string{"id", "name"}
}
func (s SkillPreset) EntityName() string   { return "cc_skill_preset_view" }
func (s SkillPreset) DefaultOrder() string { return "+name" }

func (s *SkillPreset) Validate() AppError {
	if s == nil {
		return NewBadRequestError("model.skill_preset.validate.nil_pointer_receiver", "Received empty skill preset call")
	}

	trimmedName := strings.TrimSpace(s.Name)

	if trimmedName == "" {
		return NewBadRequestError("model.skill_preset.validate.empty_name", "Name cannot be empty or contain only whitespaces")
	}

	if strings.EqualFold(trimmedName, StandartSkillPreset) {
		return NewBadRequestError("model.skill_preset.validate.reserved_name", `Name "Standart Online" is reserved`)
	}

	return nil
}

func (s *SkillPreset) PreSave() {
	s.Name = strings.TrimSpace(s.Name)
}

type DeleteSkillPresetCmd struct {
	IDs      []int64
	DomainID int64
}

type GetSkillPresetQuery struct {
	ID       int64
	DomainID int64
}

func (q *GetSkillPresetQuery) Validate() AppError {
	if q == nil {
		return NewBadRequestError("model.skill_preset.validate.empty_get_skill_preset_query", "Received empty get skill preset parameters")
	}

	if q.ID <= 0 {
		return NewBadRequestError("model.skill_preset.validate.id_required", "Skill preset ID is required during get request")
	}

	if q.DomainID <= 0 {
		return NewBadRequestError("model.skill_preset.validate.domain_id_required", "Skill preset Domain ID is required during get request")
	}
	return nil
}

type SearchSkillPresetQuery struct {
	ListRequest

	IDs         []int64
	SkillIDs    []int64
	SkipDefault bool
}

func (s *SearchSkillPresetQuery) OrderBy() string {
	if s.Sort == "" {
		return "is_system desc, name asc"
	}

	return s.Sort
}

type PatchSkillPresetCmd struct {
	Fields      []string
	DomainID    int64
	ID          int64
	Name        *string
	Description *string
	Skills      []*Lookup
	UpdatedBy   Lookup
}

func (p *PatchSkillPresetCmd) ReduceSkillsIDs() []int64 {
	ids := make([]int64, 0, len(p.Skills))

	for _, s := range p.Skills {
		if s != nil {
			ids = append(ids, int64(s.Id))
		}
	}

	return ids
}

func (p *PatchSkillPresetCmd) Validate() AppError {
	if p == nil {
		return NewBadRequestError("model.skill_preset.validate.nil_pointer_patch", "Received empty patch skill preset request")
	}

	if name := p.Name; name != nil {
		trimmedName := strings.TrimSpace(*name)

		if trimmedName == "" {
			return NewBadRequestError("model.skill_preset.validate.patch_empty_name", "Name field cannot be empty string")
		}

		if strings.EqualFold(trimmedName, StandartSkillPreset) {
			return NewBadRequestError("model.skill_preset.validate.patch_reserved_name", `Name "Standart Online" is reserved`)
		}
	}

	return nil
}

func (p *PatchSkillPresetCmd) PrePatch() {
	*p.Name = strings.TrimSpace(*p.Name)
}
