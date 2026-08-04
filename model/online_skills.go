package model

import (
	"cmp"
	"strings"
	"time"
)

const StandartOnlineSkill string = "Standart Online"

type OnlineSkills struct {
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

func (s *OnlineSkills) InitializeCreateMetadata(domainID, creatorID int64, creatorName ...string) *OnlineSkills {
	s.ActualizeCreatorInfo(domainID, creatorID, creatorName...).
		ActualizeUpdatorInfo(domainID, creatorID, creatorName...)

	return s
}

func (s *OnlineSkills) ActualizeCreatorInfo(domainID, creatorID int64, creatorName ...string) *OnlineSkills {
	s.DomainID = domainID

	creator := Lookup{Id: int(creatorID)}

	if len(creatorName) != 0 && strings.TrimSpace(creatorName[0]) != "" {
		creator.Name = creatorName[0]
	}

	s.CreatedBy = &creator

	return s
}

func (s *OnlineSkills) ActualizeUpdatorInfo(domainID, creatorID int64, updatorName ...string) *OnlineSkills {
	s.DomainID = domainID

	updator := Lookup{Id: int(creatorID)}

	if len(updatorName) != 0 && strings.TrimSpace(updatorName[0]) != "" {
		updator.Name = updatorName[0]
	}

	s.UpdatedBy = &updator

	return s
}

func (s *OnlineSkills) GetDescription() string {
	return *(cmp.Or(s.Description, new(string)))
}

func (s *OnlineSkills) TryUseDescription(desc string) *OnlineSkills {
	if strings.TrimSpace(desc) != "" {
		s.Description = &desc
	}

	return s
}

func (s *OnlineSkills) UpdatedAtUnix() int64 {
	if s == nil {
		return 0
	}

	return max(s.UpdatedAt.UTC().UnixMilli(), 0)
}

func (s *OnlineSkills) CreatedAtUnix() int64 {
	if s == nil {
		return 0
	}

	return max(s.CreatedAt.UTC().UnixMilli(), 0)
}

func (s *OnlineSkills) ReduceSkillsIDs() []int64 {
	skillIDs := make([]int64, 0, len(s.Skills))

	for _, s := range s.Skills {
		skillIDs = append(skillIDs, int64(s.Id))
	}

	return skillIDs
}

func (s OnlineSkills) AllowFields() []string {
	return []string{"id", "name", "description", "created_by", "created_at", "updated_by", "updated_at", "skills"}
}

func (s OnlineSkills) DefaultFields() []string {
	return []string{"id", "name"}
}
func (s OnlineSkills) EntityName() string   { return "cc_online_skills_list" }
func (s OnlineSkills) DefaultOrder() string { return "is_system desc, name asc" }

func (s *OnlineSkills) Validate() AppError {
	if s == nil {
		return NewBadRequestError("model.online_skills.validate.nil_pointer_receiver", "Received empty online skills call")
	}

	trimmedName := strings.TrimSpace(s.Name)

	if trimmedName == "" {
		return NewBadRequestError("model.online_skills.validate.empty_name", "Name cannot be empty or contain only whitespaces")
	}

	if strings.EqualFold(trimmedName, StandartOnlineSkill) {
		return NewBadRequestError("model.online_skills.validate.reserved_name", `Name "Standart Online" is reserved`)
	}

	return nil
}

func (s *OnlineSkills) PreSave() {
	s.Name = strings.TrimSpace(s.Name)
}

type DeleteSkillPresetCmd struct {
	ID       int64
	DomainID int64
}

type GetSkillPresetQuery struct {
	ID       int64
	DomainID int64
}

func (q *GetSkillPresetQuery) Validate() AppError {
	if q == nil {
		return NewBadRequestError("model.online_skills.validate.empty_get_online_skills_query", "Received empty get online skills parameters")
	}

	if q.ID <= 0 {
		return NewBadRequestError("model.online_skills.validate.id_required", "Online skills ID is required during get request")
	}

	if q.DomainID <= 0 {
		return NewBadRequestError("model.online_skills.validate.domain_id_required", "Online skills Domain ID is required during get request")
	}
	return nil
}

type SearchOnlineSkillsQuery struct {
	ListRequest

	IDs         []int64
	SkillIDs    []int64
	SkipDefault bool
}

type PatchOnlineSkillsCmd struct {
	Fields      []string
	DomainID    int64
	ID          int64
	Name        *string
	Description *string
	Skills      []*Lookup
	UpdatedBy   Lookup
}

func (p *PatchOnlineSkillsCmd) ReduceSkillsIDs() []int64 {
	ids := make([]int64, 0, len(p.Skills))

	for _, s := range p.Skills {
		if s != nil {
			ids = append(ids, int64(s.Id))
		}
	}

	return ids
}

func (p *PatchOnlineSkillsCmd) Validate() AppError {
	if p == nil {
		return NewBadRequestError("model.online_skills.validate.nil_pointer_patch", "Received empty patch online skills request")
	}

	if name := p.Name; name != nil {
		trimmedName := strings.TrimSpace(*name)

		if trimmedName == "" {
			return NewBadRequestError("model.online_skills.validate.patch_empty_name", "Name field cannot be empty string")
		}

		if strings.EqualFold(trimmedName, StandartOnlineSkill) {
			return NewBadRequestError("model.online_skills.validate.patch_reserved_name", `Name "Standart Online" is reserved`)
		}
	}

	return nil
}

func (p *PatchOnlineSkillsCmd) PrePatch() {
	*p.Name = strings.TrimSpace(*p.Name)
}
