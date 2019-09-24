package model

type OutboundCallResource struct {
	DomainRecord
	Limit                 int         `json:"limit" db:"limit"`
	Enabled               bool        `json:"enabled" db:"enabled"`
	RPS                   int         `json:"rps" db:"rps"`
	Reserve               bool        `json:"reserve" db:"reserve"`
	Variables             StringMap   `json:"variables" db:"variables"`
	Number                string      `json:"number" db:"number"`
	MaxSuccessivelyErrors int         `json:"max_successively_errors" db:"max_successively_errors"`
	Name                  string      `json:"name" db:"name"`
	DialString            string      `json:"dial_string" db:"dial_string"`
	ErrorIds              StringArray `json:"error_ids" db:"error_ids"`
	LastErrorId           *string     `json:"last_error_id" db:"last_error_id"`
	SuccessivelyErrors    int         `json:"successively_errors" db:"successively_errors"`
	LastErrorAt           int64       `json:"last_error_at" db:"last_error_at"`
}

func (r *OutboundCallResource) LastError() string {
	if r.LastErrorId == nil {
		return ""
	}
	return *r.LastErrorId
}

func (r *OutboundCallResource) IsValid() *AppError {
	//FIXME
	return nil
}
