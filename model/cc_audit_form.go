package model

import "time"

type Questions string

type AuditForm struct {
	Id int32 `json:"id" db:"id"`
	AclRecord
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	Enabled     bool      `json:"enabled" db:"enabled"`
	Questions   Questions `json:"questions" db:"questions"`
	Teams       []*Lookup `json:"teams" db:"teams"`
}

type AuditFormPatch struct {
	UpdatedAt   time.Time
	UpdatedBy   Lookup
	Name        *string    `json:"name" db:"name"`
	Description *string    `json:"description" db:"description"`
	Enabled     *bool      `json:"enabled" db:"enabled"`
	Questions   *Questions `json:"questions" db:"questions"`
	Teams       []*Lookup  `json:"teams" db:"teams"`
}

type SearchAuditForm struct {
	ListRequest
	Ids []int32
}

func (q *AuditForm) Patch(p *AuditFormPatch) {
	q.UpdatedAt = &p.UpdatedAt
	q.UpdatedBy = &p.UpdatedBy

	if p.Name != nil {
		q.Name = *p.Name
	}

	if p.Description != nil {
		q.Description = *p.Description
	}

	if p.Enabled != nil {
		q.Enabled = *p.Enabled
	}

	if p.Questions != nil {
		q.Questions = *p.Questions
	}

	if p.Teams != nil {
		q.Teams = p.Teams
	}
}
func (AuditForm) DefaultOrder() string {
	return "name"
}

func (AuditForm) AllowFields() []string {
	return []string{"id", "name", "description", "domain_id", "created_at", "created_by", "updated_at", "updated_by",
		"enabled", "questions", "teams"}
}

func (AuditForm) DefaultFields() []string {
	return []string{"id", "name", "description", "teams"}
}

func (AuditForm) EntityName() string {
	return "cc_audit_form_view"
}

func (af *AuditForm) IsValid() *AppError {
	//FIXME
	return nil
}
