package model

type Queue struct {
	DomainRecord
	Strategy       string    `json:"strategy" db:"strategy"`
	Enabled        bool      `json:"enabled" db:"enabled"`
	Payload        []byte    `json:"payload" db:"payload"`
	Calendar       Lookup    `json:"calendar" db:"calendar"`
	Priority       int       `json:"priority" db:"priority"`
	Name           string    `json:"name" db:"name"`
	Variables      StringMap `json:"variables" db:"variables"`
	Timeout        int       `json:"timeout" db:"timeout"`
	DncList        *Lookup   `json:"dnc_list" db:"dnc_list"`
	SecLocateAgent int       `json:"sec_locate_agent" db:"sec_locate_agent"`
	Type           int8      `json:"type" db:"type"`
	Team           *Lookup   `json:"team" db:"team"`
	Schema         *Lookup   `json:"schema" db:"schema"`
	Description    string    `json:"description" db:"description"`
	Count          int       `json:"count" db:"count"`
	Waiting        int       `json:"waiting" db:"waiting"`
	Active         int       `json:"active" db:"active"`
	Ringtone       *Lookup   `json:"ringtone" db:"ringtone"`
}

func (q Queue) AllowFields() []string {
	return q.DefaultFields()
}

func (q Queue) DefaultFields() []string {
	return []string{"id", "strategy", "enabled", "payload", "priority", "updated_at", "name", "variables", "timeout",
		"domain_id", "sec_locate_agent", "type", "created_at", "created_by", "updated_by", "calendar", "dnc_list", "team", "description",
		"schema", "count", "waiting", "active", "ringtone"}
}

func (q Queue) EntityName() string {
	return "cc_queue_list"
}

type SearchQueue struct {
	ListRequest
	Ids []string
}

type SearchQueueReportGeneral struct {
	ListRequest
	JoinedAt FilterBetween
	QueueIds []int32
	TeamIds  []int32
	Types    []int32
}

type QueueReportGeneral struct {
	Queue      Lookup  `json:"queue" db:"queue"`
	Team       *Lookup `json:"team" db:"team"`
	Online     int32   `json:"online" db:"online"`
	Pause      int32   `json:"pause" db:"pause"`
	Bridged    float32 `json:"bridged" db:"bridged"`
	Waiting    int64   `json:"waiting" db:"waiting"`
	Processed  int64   `json:"processed" db:"processed"`
	Count      int64   `json:"count" db:"count"`
	Abandoned  float32 `json:"abandoned" db:"abandoned"`
	SumBillSec float32 `json:"sum_bill_sec" db:"sum_bill_sec"`
	AvgWrapSec float32 `json:"avg_wrap_sec" db:"avg_wrap_sec"`
	AvgAwtSec  float32 `json:"avg_awt_sec" db:"avg_awt_sec"`
	MaxAwtSec  float32 `json:"max_awt_sec" db:"max_awt_sec"`
	AvgAsaSec  float32 `json:"avg_asa_sec" db:"avg_asa_sec"`
	AvgAhtSec  float32 `json:"avg_aht_sec" db:"avg_aht_sec"`
}

type QueuePatch struct {
	Strategy       *string   `json:"strategy" db:"strategy"`
	Enabled        *bool     `json:"enabled" db:"enabled"`
	Payload        []byte    `json:"payload" db:"payload"`
	Calendar       *Lookup   `json:"calendar" db:"calendar"`
	Priority       *int      `json:"priority" db:"priority"`
	Name           *string   `json:"name" db:"name"`
	Variables      StringMap `json:"variables" db:"variables"`
	Timeout        *int      `json:"timeout" db:"timeout"`
	DncList        *Lookup   `json:"dnc_list" db:"dnc_list"`
	SecLocateAgent *int      `json:"sec_locate_agent" db:"sec_locate_agent"`
	Team           *Lookup   `json:"team" db:"team"`
	Schema         *Lookup   `json:"schema" db:"schema"`
	Ringtone       *Lookup   `json:"ringtone" db:"ringtone"`
	Description    *string   `json:"description" db:"description"`
}

func (q *Queue) Patch(p *QueuePatch) {
	if p.Strategy != nil {
		q.Strategy = *p.Strategy
	}

	if p.Enabled != nil {
		q.Enabled = *p.Enabled
	}

	if p.Payload != nil {
		q.Payload = p.Payload
	}

	if p.Calendar != nil {
		q.Calendar = *p.Calendar
	}

	if p.Priority != nil {
		q.Priority = *p.Priority
	}

	if p.Name != nil {
		q.Name = *p.Name
	}

	if p.Variables != nil {
		q.Variables = p.Variables
	}

	if p.Timeout != nil {
		q.Timeout = *p.Timeout
	}

	if p.DncList != nil {
		q.DncList = p.DncList
	}

	if p.SecLocateAgent != nil {
		q.SecLocateAgent = *p.SecLocateAgent
	}

	if p.Schema != nil {
		q.Schema = p.Schema
	}

	if p.Team != nil {
		q.Team = p.Team
	}

	if p.Description != nil {
		q.Description = *p.Description
	}

	if p.Ringtone != nil {
		q.Ringtone = p.Ringtone
	}
}

func (q *Queue) IsValid() *AppError {
	//FIXME
	return nil
}

func (q *Queue) DncListId() *int64 {
	if q.DncList != nil {
		return NewInt64(int64(q.DncList.Id))
	}
	return nil
}

func (q *Queue) RingtoneId() *int {
	if q.Ringtone != nil {
		return &q.Ringtone.Id
	}

	return nil
}

func (q *Queue) TeamId() *int64 {
	if q.Team != nil {
		return NewInt64(int64(q.Team.Id))
	}
	return nil
}

func (q *Queue) SchemaId() *int64 {
	if q.Schema != nil {
		return NewInt64(int64(q.Schema.Id))
	}
	return nil
}
