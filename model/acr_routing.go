package model

type RoutingSchema struct {
	DomainRecord
	Name        string `json:"name" db:"name"`
	Type        int8   `json:"type" db:"type"`
	Schema      []byte `json:"schema" db:"scheme"`
	Payload     []byte `json:"payload" db:"payload"`
	Description string `json:"description" db:"description"`
	Debug       bool   `json:"debug" db:"debug"`
}

type RoutingSchemaPath struct {
	UpdatedById int
	Name        *string `json:"name" db:"name"`
	Type        *int8   `json:"type" db:"type"`
	Schema      []byte  `json:"schema" db:"scheme"`
	Payload     []byte  `json:"payload" db:"payload"`
	Description *string `json:"description" db:"description"`
	Debug       *bool   `json:"debug" db:"debug"`
}

func (s *RoutingSchema) IsValid() *AppError {
	//FIXME
	return nil
}

func (s *RoutingSchema) Patch(in *RoutingSchemaPath) {
	if in.Name != nil {
		s.Name = *in.Name
	}

	if in.Type != nil {
		s.Type = *in.Type
	}

	if in.Schema != nil {
		s.Schema = in.Schema
	}

	if in.Payload != nil {
		s.Payload = in.Payload
	}

	if in.Description != nil {
		s.Description = *in.Description
	}

	if in.Debug != nil {
		s.Debug = *in.Debug
	}
}

type RoutingInboundCall struct {
	DomainRecord
	Name        string      `json:"name" db:"name"`
	Description string      `json:"description" db:"description"`
	StartSchema Lookup      `json:"start_schema" db:"start_scheme"`
	StopSchema  *Lookup     `json:"stop_schema" db:"stop_scheme"`
	Numbers     StringArray `json:"numbers" db:"numbers"`
	Host        string      `json:"host" db:"host"`
	Timezone    Lookup      `json:"timezone" db:"timezone"`
	Debug       bool        `json:"debug" db:"debug"`
	Disabled    bool        `json:"disabled" db:"disabled"`
}

type RoutingVariable struct {
	Id       int64  `json:"id" db:"id"`
	DomainId int64  `json:"domain_id" db:"domain_id"`
	Key      string `json:"key" db:"key"`
	Value    string `json:"value" db:"value"`
}

func (r *RoutingVariable) IsValid() *AppError {
	//FIXME
	return nil
}

func (r *RoutingInboundCall) GetStopSchemaId() *int {
	if r.StopSchema == nil {
		return nil
	}
	return &r.StopSchema.Id
}

func (r *RoutingInboundCall) GetStartSchemaId() *int {
	if r.StartSchema.Id == 0 {
		return nil
	}
	return &r.StartSchema.Id
}

func (s *RoutingInboundCall) IsValid() *AppError {
	//FIXME
	return nil
}

type RoutingOutboundCall struct {
	DomainRecord
	Name        string `json:"name" db:"name"`
	Description string `json:"description" db:"description"`
	Schema      Lookup `json:"schema" db:"scheme"`
	Priority    int    `json:"priority" db:"priority"`
	Pattern     string `json:"pattern" db:"pattern"`
	Disabled    bool   `json:"disabled" db:"disabled"`
}

type RoutingOutboundCallPatch struct {
	UpdatedById int
	Name        *string `json:"name" db:"name"`
	Description *string `json:"description" db:"description"`
	Schema      *Lookup `json:"schema" db:"scheme"`
	Priority    *int    `json:"priority" db:"priority"`
	Pattern     *string `json:"pattern" db:"pattern"`
	Disabled    *bool   `json:"disabled" db:"disabled"`
}

func (r *RoutingOutboundCall) Patch(patch *RoutingOutboundCallPatch) {
	if patch.Name != nil {
		r.Name = *patch.Name
	}

	if patch.Description != nil {
		r.Description = *patch.Description
	}

	if patch.Schema != nil {
		r.Schema = *patch.Schema
	}

	if patch.Priority != nil {
		r.Priority = *patch.Priority
	}

	if patch.Pattern != nil {
		r.Pattern = *patch.Pattern
	}

	if patch.Disabled != nil {
		r.Disabled = *patch.Disabled
	}
}

func (r *RoutingOutboundCall) GetSchemaId() *int {
	if r.Schema.Id == 0 {
		return nil
	}
	return &r.Schema.Id
}

func (s *RoutingOutboundCall) IsValid() *AppError {
	//FIXME
	return nil
}
