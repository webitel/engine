package model

type RoutingScheme struct {
	DomainRecord
	Name        string `json:"name" db:"name"`
	Type        int8   `json:"type" db:"type"`
	Scheme      []byte `json:"scheme" db:"scheme"`
	Payload     []byte `json:"payload" db:"payload"`
	Description string `json:"description" db:"description"`
	Debug       bool   `json:"debug" db:"debug"`
}

func (s *RoutingScheme) IsValid() *AppError {
	//FIXME
	return nil
}

type RoutingInboundCall struct {
	DomainRecord
	Name        string      `json:"name" db:"name"`
	Description string      `json:"description" db:"description"`
	StartScheme Lookup      `json:"start_scheme" db:"start_scheme"`
	StopScheme  *Lookup     `json:"stop_scheme" db:"stop_scheme"`
	Numbers     StringArray `json:"numbers" db:"numbers"`
	Host        string      `json:"host" db:"host"`
	Timezone    Lookup      `json:"timezone" db:"timezone"`
	Debug       bool        `json:"debug" db:"debug"`
	Disabled    bool        `json:"disabled" db:"disabled"`
}

func (r *RoutingInboundCall) GetStopSchemeId() *int {
	if r.StopScheme == nil {
		return nil
	}
	return &r.StopScheme.Id
}

func (r *RoutingInboundCall) GetStartSchemeId() *int {
	if r.StartScheme.Id == 0 {
		return nil
	}
	return &r.StartScheme.Id
}

func (s *RoutingInboundCall) IsValid() *AppError {
	//FIXME
	return nil
}

type RoutingOutboundCall struct {
	DomainRecord
	Name        string `json:"name" db:"name"`
	Description string `json:"description" db:"description"`
	Scheme      Lookup `json:"scheme" db:"scheme"`
	Priority    int    `json:"priority" db:"priority"`
	Pattern     string `json:"pattern" db:"pattern"`
	Disabled    bool   `json:"disabled" db:"disabled"`
}

func (r *RoutingOutboundCall) GetSchemeId() *int {
	if r.Scheme.Id == 0 {
		return nil
	}
	return &r.Scheme.Id
}

func (s *RoutingOutboundCall) IsValid() *AppError {
	//FIXME
	return nil
}
