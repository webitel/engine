package model

type Queue struct {
	DomainRecord
	Strategy             string          `json:"strategy" db:"strategy"`
	Enabled              bool            `json:"enabled" db:"enabled"`
	Payload              StringInterface `json:"payload" db:"payload"`
	Calendar             *Lookup         `json:"calendar" db:"calendar"`
	Priority             int             `json:"priority" db:"priority"`
	Name                 string          `json:"name" db:"name"`
	Variables            StringMap       `json:"variables" db:"variables"`
	Timeout              int             `json:"timeout" db:"-"`          //todo del me
	SecLocateAgent       int             `json:"sec_locate_agent" db:"-"` //todo del me
	DncList              *Lookup         `json:"dnc_list" db:"dnc_list"`
	Type                 int8            `json:"type" db:"type"`
	Team                 *Lookup         `json:"team" db:"team"`
	Schema               *Lookup         `json:"schema" db:"schema"`
	Description          string          `json:"description" db:"description"`
	Count                int             `json:"count" db:"count"`
	Waiting              int             `json:"waiting" db:"waiting"`
	Active               int             `json:"active" db:"active"`
	Ringtone             *Lookup         `json:"ringtone" db:"ringtone"`
	DoSchema             *Lookup         `json:"do_schema" db:"do_schema"`
	AfterSchema          *Lookup         `json:"after_schema" db:"after_schema"`
	StickyAgent          bool            `json:"sticky_agent" db:"sticky_agent"`
	Processing           bool            `json:"processing" db:"processing"`
	ProcessingSec        uint32          `json:"processing_sec" db:"processing_sec"`
	ProcessingRenewalSec uint32          `json:"processing_renewal_sec" db:"processing_renewal_sec"`
	FormSchema           *Lookup         `json:"form_schema" db:"form_schema"`

	TaskProcessing *QueueTaskProcessing `json:"task_processing" db:"task_processing"`
	Grantee        *Lookup              `json:"grantee" db:"grantee"`
}

type QueueTaskProcessing struct {
	Enabled    bool    `json:"enabled"`
	Sec        uint32  `json:"sec"`
	RenewalSec uint32  `json:"renewal_sec"`
	FormSchema *Lookup `json:"form_schema"`
}

func (q Queue) AllowFields() []string {
	return q.DefaultFields()
}

func (q Queue) DefaultOrder() string {
	return "-priority"
}

func (q Queue) DefaultFields() []string {
	return []string{"id", "strategy", "enabled", "payload", "priority", "updated_at", "name", "variables",
		"domain_id", "type", "created_at", "created_by", "updated_by", "calendar", "dnc_list", "team", "description",
		"schema", "count", "waiting", "active", "ringtone", "do_schema", "after_schema", "sticky_agent",
		"processing", "processing_sec", "processing_renewal_sec", "form_schema", "task_processing", "grantee"}
}

func (q Queue) EntityName() string {
	return "cc_queue_list"
}

type SearchQueue struct {
	ListRequest
	Ids   []string
	Types []uint32
}

type SearchQueueReportGeneral struct {
	ListRequest
	JoinedAt FilterBetween
	QueueIds []int32
	TeamIds  []int32
	Types    []int32
}

type QueueAgentAgg struct {
	Online  uint32 `json:"online" db:"online"`
	Pause   uint32 `json:"pause" db:"pause"`
	Offline uint32 `json:"offline" db:"offline"`
	Free    uint32 `json:"free" db:"free"`
	Total   uint32 `json:"total" db:"total"`
}

type QueueReportGeneral struct {
	Queue       Lookup         `json:"queue" db:"queue"`
	Team        *Lookup        `json:"team" db:"team"`
	AgentStatus *QueueAgentAgg `json:"agent_status" db:"agent_status"`

	Missed    uint32 `json:"missed" db:"missed"`
	Processed uint32 `json:"processed" db:"processed"`
	Waiting   uint32 `json:"waiting" db:"waiting"`

	Count       uint64  `json:"count" db:"count"`
	Transferred uint32  `json:"transferred" db:"transferred"`
	Abandoned   uint32  `json:"abandoned" db:"abandoned"`
	Attempts    uint32  `json:"attempts" db:"attempts"`
	Bridged     float32 `json:"bridged" db:"bridged"`

	SumBillSec float32 `json:"sum_bill_sec" db:"sum_bill_sec"`
	AvgWrapSec float32 `json:"avg_wrap_sec" db:"avg_wrap_sec"`
	AvgAwtSec  float32 `json:"avg_awt_sec" db:"avg_awt_sec"`
	AvgAsaSec  float32 `json:"avg_asa_sec" db:"avg_asa_sec"`
	AvgAhtSec  float32 `json:"avg_aht_sec" db:"avg_aht_sec"`
	Sl20       float32 `json:"sl20" db:"sl20"`
	Sl30       float32 `json:"sl30" db:"sl30"`
}

type QueueReportGeneralAgg struct {
	Next  bool                  `json:"next"`
	Items []*QueueReportGeneral `json:"items" db:"items"`
	Aggs  QueueAgentAgg         `json:"aggs"`
}

type QueuePatch struct {
	Strategy             *string         `json:"strategy" db:"strategy"`
	Enabled              *bool           `json:"enabled" db:"enabled"`
	Payload              StringInterface `json:"payload" db:"payload"`
	Calendar             *Lookup         `json:"calendar" db:"calendar"`
	Priority             *int            `json:"priority" db:"priority"`
	Name                 *string         `json:"name" db:"name"`
	Variables            StringMap       `json:"variables" db:"variables"`
	DncList              *Lookup         `json:"dnc_list" db:"dnc_list"`
	Team                 *Lookup         `json:"team" db:"team"`
	Schema               *Lookup         `json:"schema" db:"schema"`
	Ringtone             *Lookup         `json:"ringtone" db:"ringtone"`
	DoSchema             *Lookup         `json:"do_schema" db:"do_schema"`
	AfterSchema          *Lookup         `json:"after_schema" db:"after_schema"`
	Description          *string         `json:"description" db:"description"`
	StickyAgent          *bool           `json:"sticky_agent" db:"sticky_agent"`
	Processing           *bool           `json:"processing" db:"processing"`
	ProcessingSec        *uint32         `json:"processing_sec" db:"processing_sec"`
	ProcessingRenewalSec *uint32         `json:"processing_renewal_sec" db:"processing_renewal_sec"`
	FormSchema           *Lookup         `json:"form_schema" db:"form_schema"`
	Grantee              *Lookup         `json:"grantee" db:"grantee"`
}

func (q *Queue) Patch(p *QueuePatch) {
	// TODO
	q.UpdatedAt = GetMillis()

	if p.Strategy != nil {
		q.Strategy = *p.Strategy
	}

	if p.Enabled != nil {
		q.Enabled = *p.Enabled
	}

	if p.Payload != nil {
		if q.Payload == nil {
			q.Payload = StringInterface{}
		}
		for k, v := range p.Payload {
			q.Payload[k] = v
		}
	}

	if p.Calendar != nil {
		q.Calendar = p.Calendar
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

	if p.DncList != nil {
		q.DncList = p.DncList
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

	if p.DoSchema != nil {
		q.DoSchema = p.DoSchema
	}

	if p.AfterSchema != nil {
		q.AfterSchema = p.AfterSchema
	}

	if p.StickyAgent != nil {
		q.StickyAgent = *p.StickyAgent
	}

	if p.Processing != nil {
		q.Processing = *p.Processing
	}

	if p.ProcessingSec != nil {
		q.ProcessingSec = *p.ProcessingSec
	}

	if p.ProcessingRenewalSec != nil {
		q.ProcessingRenewalSec = *p.ProcessingRenewalSec
	}

	if p.FormSchema != nil {
		q.FormSchema = p.FormSchema
	}

	if p.Grantee != nil {
		q.Grantee = p.Grantee
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

func (q *Queue) DoSchemaId() *int64 {
	if q.DoSchema != nil {
		return NewInt64(int64(q.DoSchema.Id))
	}
	return nil
}

func (q *Queue) AfterSchemaId() *int64 {
	if q.AfterSchema != nil {
		return NewInt64(int64(q.AfterSchema.Id))
	}
	return nil
}
