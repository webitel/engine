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
	Gateway               *Lookup     `json:"gateway" db:"gateway"`
}

type OutboundCallResourcePath struct {
	Limit                 *int         `json:"limit" db:"limit"`
	Enabled               *bool        `json:"enabled" db:"enabled"`
	RPS                   *int         `json:"rps" db:"rps"`
	Reserve               *bool        `json:"reserve" db:"reserve"`
	MaxSuccessivelyErrors *int         `json:"max_successively_errors" db:"max_successively_errors"`
	Name                  *string      `json:"name" db:"name"`
	ErrorIds              *StringArray `json:"error_ids" db:"error_ids"`
	Gateway               *Lookup      `json:"gateway" db:"gateway"`
}

func (r *OutboundCallResource) GetGatewayId() *int {
	if r.Gateway != nil {
		return NewInt(r.Gateway.Id)
	}
	return nil
}

func (r *OutboundCallResource) Path(p *OutboundCallResourcePath) {
	if p.Limit != nil {
		r.Limit = *p.Limit
	}
	if p.Enabled != nil {
		r.Enabled = *p.Enabled
	}
	if p.RPS != nil {
		r.RPS = *p.RPS
	}
	if p.Reserve != nil {
		r.Reserve = *p.Reserve
	}
	if p.MaxSuccessivelyErrors != nil {
		r.MaxSuccessivelyErrors = *p.MaxSuccessivelyErrors
	}
	if p.Name != nil {
		r.Name = *p.Name
	}
	if p.ErrorIds != nil {
		r.ErrorIds = *p.ErrorIds
	}

	if p.Gateway != nil {
		r.Gateway = p.Gateway
	}
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
