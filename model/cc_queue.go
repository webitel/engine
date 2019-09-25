package model

type Queue struct {
	DomainRecord
	Strategy          string    `json:"strategy" db:"strategy"`
	Enabled           bool      `json:"enabled" db:"enabled"`
	Payload           []byte    `json:"payload" db:"payload"`
	Calendar          Lookup    `json:"calendar" db:"calendar"`
	Priority          int       `json:"priority" db:"priority"`
	MaxCalls          int       `json:"max_calls" db:"max_calls"`
	SecBetweenRetries int       `json:"sec_between_retries" db:"sec_between_retries"`
	Name              string    `json:"name" db:"name"`
	MaxOfRetry        int       `json:"max_of_retry" db:"max_of_retry"`
	Variables         StringMap `json:"variables" db:"variables"`
	Timeout           int       `json:"timeout" db:"timeout"`
	DncList           *Lookup   `json:"dnc_list" db:"dnc_list"`
	SecLocateAgent    int       `json:"sec_locate_agent" db:"sec_locate_agent"`
	Type              int8      `json:"type" db:"type"`
	Team              *Lookup   `json:"team" db:"team"`
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
