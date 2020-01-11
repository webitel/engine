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
	Description    string    `json:"description" db:"description"`
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

	if p.Team != nil {
		q.Team = p.Team
	}

	if p.Description != nil {
		q.Description = *p.Description
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

func (q *Queue) TeamId() *int64 {
	if q.Team != nil {
		return NewInt64(int64(q.Team.Id))
	}
	return nil
}
