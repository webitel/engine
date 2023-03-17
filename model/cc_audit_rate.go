package model

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type QuestionAnswers []*QuestionAnswer

type QuestionAnswer struct {
	Score int32 `json:"score"`
}

type Rate struct {
	CallId    *string         `json:"call_id" db:"call_id"`
	RatedUser *Lookup         `json:"rated_user" db:"rated_user"`
	Form      *Lookup         `json:"form" db:"form"`
	Answers   QuestionAnswers `json:"answers" db:"answers"`
	Comment   string          `json:"comment" db:"comment"`
}

type AuditRate struct {
	Id int64 `json:"id" db:"id"`
	AclRecord
	Rate
	ScoreRequired float32   `json:"score_required" db:"score_required"`
	ScoreOptional float32   `json:"score_optional" db:"score_optional"`
	Questions     Questions `json:"questions" db:"questions"`
}

type SearchAuditRate struct {
	ListRequest
	Ids          []int32
	CallIds      []string `json:"call_ids" db:"call_ids"`
	CreatedAt    *FilterBetween
	FormIds      []int32
	RatedUserIds []int64
}

func (a QuestionAnswers) ToJson() []byte {
	data, _ := json.Marshal(a)
	return data
}

func (AuditRate) DefaultOrder() string {
	return "id"
}

func (AuditRate) AllowFields() []string {
	return []string{"id", "created_at", "created_by", "updated_at", "updated_by", "rated_user",
		"form", "answers", "score_required", "score_optional", "comment", "call_id", "questions"}
}

func (AuditRate) DefaultFields() []string {
	return []string{"id", "created_at", "created_by", "rated_user", "form", "score_required", "score_optional"}
}

func (AuditRate) EntityName() string {
	return "cc_audit_rate_view"
}

func (r *AuditRate) IsValid() *AppError {

	return nil
}

// TODO call_id

func (r *AuditRate) SetRate(form *AuditForm, rate Rate) *AppError {
	if len(form.Questions) != len(rate.Answers) {
		return NewAppError("AuditRate", "audit.rate.valid.answers", nil, "Answers not equals questions", http.StatusBadRequest)
	}

	r.Answers = rate.Answers

	r.Form = &Lookup{Id: int(form.Id)}
	r.RatedUser = rate.RatedUser
	r.CallId = rate.CallId
	r.Comment = rate.Comment

	for i, a := range r.Answers {
		if form.Questions[i].Required {
			if a == nil {
				return NewAppError("AuditRate", "audit.rate.valid.question", nil,
					fmt.Sprintf("question \"%s\" is required", form.Questions[i].Question), http.StatusBadRequest)
			}

			if !form.Questions[i].ValidAnswer(*a) {
				return NewAppError("AuditRate", "audit.rate.valid.answer", nil,
					fmt.Sprintf("answer \"%s\" not allowed %d", form.Questions[i].Question, a.Score), http.StatusBadRequest)
			}

			r.ScoreRequired += float32(a.Score)
		} else if a != nil { // skip optional if empty

			if !form.Questions[i].ValidAnswer(*a) {
				return NewAppError("AuditRate", "audit.rate.valid.answer", nil,
					fmt.Sprintf("answer \"%s\" not allowed %d", form.Questions[i].Question, a.Score), http.StatusBadRequest)
			}
			r.ScoreOptional += float32(a.Score)
		}
	}

	if r.ScoreRequired > 0 {
		r.ScoreRequired = (r.ScoreRequired * 100) / float32(form.Questions.SumMax(true))
	}

	if r.ScoreOptional > 0 {
		r.ScoreOptional = (r.ScoreOptional * 100) / float32(form.Questions.SumMax(false))
	}

	return nil
}
