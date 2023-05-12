package model

import (
	"net/http"
	"time"
)

type AuditForm struct {
	Id int32 `json:"id" db:"id"`
	AclRecord
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	Enabled     bool      `json:"enabled" db:"enabled"`
	Questions   Questions `json:"questions" db:"questions"`
	Teams       []*Lookup `json:"teams" db:"teams"`
	Archive     bool      `json:"archive"`
	Editable    bool      `json:"editable"`
}

type AuditFormPatch struct {
	UpdatedAt   time.Time
	UpdatedBy   Lookup
	Name        *string   `json:"name" db:"name"`
	Description *string   `json:"description" db:"description"`
	Enabled     *bool     `json:"enabled" db:"enabled"`
	Questions   Questions `json:"questions" db:"questions"`
	Teams       []*Lookup `json:"teams" db:"teams"`
	Archive     *bool     `json:"archive"`
}

type SearchAuditForm struct {
	ListRequest
	Ids      []int32
	TeamIds  []int32 `json:"team_ids" db:"team_ids"`
	Archive  *bool   `json:"archive"`
	Editable *bool   `json:"editable"`
	Enabled  *bool   `json:"enabled"`
	Question string  `json:"question"`
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
		q.Questions = p.Questions
	}

	if p.Teams != nil {
		q.Teams = p.Teams
	}

	if p.Archive != nil {
		q.Archive = *p.Archive
	}
}

func (AuditForm) DefaultOrder() string {
	return "name"
}

func (AuditForm) AllowFields() []string {
	return []string{"id", "name", "description", "domain_id", "created_at", "created_by", "updated_at", "updated_by",
		"enabled", "questions", "teams", "archive", "editable"}
}

func (AuditForm) DefaultFields() []string {
	return []string{"id", "name", "description", "teams", "archive", "editable", "enabled", "created_at", "created_by", "updated_at", "updated_by"}
}

func (AuditForm) EntityName() string {
	return "cc_audit_form_view"
}

func (af *AuditForm) IsValid() *AppError {
	if len(af.Name) < 3 || len(af.Name) > 256 {
		return NewAppError("AuditForm.IsValid", "app.audit_form.is_valid.name", nil, "Name should not be less than 3 characters or greater than 256 characters", http.StatusBadRequest)
	}

	if len(af.Description) > 516 {
		return NewAppError("AuditForm.IsValid", "app.audit_form.is_valid.description", nil, "Value should not be greater than 516 characters", http.StatusBadRequest)
	}

	for _, v := range af.Questions {
		switch v.Type {
		case QuestionTypeScore:
			if v.Max == 0 {
				return NewAppError("AuditForm.IsValid", "app.audit_form.is_valid.question.max", nil, "", http.StatusBadRequest)
			}

			if v.Min > v.Max {
				return NewAppError("AuditForm.IsValid", "app.audit_form.is_valid.question.min_max", nil, "", http.StatusBadRequest)
			}

		case QuestionTypeOptions:
			if len(v.Options) == 0 {
				return NewAppError("AuditForm.IsValid", "app.audit_form.is_valid.option.options", nil, "", http.StatusBadRequest)
			}
		default:
			return NewAppError("AuditForm.IsValid", "app.audit_form.is_valid.question.type", nil, "", http.StatusBadRequest)
		}
	}

	return nil
}