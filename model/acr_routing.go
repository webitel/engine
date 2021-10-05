package model

type RoutingSchema struct {
	DomainRecord
	Name        string `json:"name" db:"name"`
	Type        string `json:"type" db:"type"`
	Schema      []byte `json:"schema" db:"schema"`
	Payload     []byte `json:"payload" db:"payload"`
	Description string `json:"description" db:"description"`
	Debug       bool   `json:"debug" db:"debug"`
	Editor      bool   `json:"editor" db:"editor"`
}

type SearchRoutingSchema struct {
	ListRequest
	Ids  []uint32
	Name *string
}

func (RoutingSchema) DefaultOrder() string {
	return "id"
}

func (a RoutingSchema) AllowFields() []string {
	return []string{"id", "domain_id", "name", "created_at", "created_by", "updated_at", "updated_by",
		"debug", "schema", "payload", "editor", "type"}
}

func (a RoutingSchema) DefaultFields() []string {
	return []string{"id", "name", "editor", "type"}
}

func (a RoutingSchema) EntityName() string {
	return "acr_routing_scheme_view"
}

type RoutingSchemaPath struct {
	UpdatedById int
	Name        *string `json:"name" db:"name"`
	Type        *string `json:"type" db:"type"`
	Schema      []byte  `json:"schema" db:"scheme"`
	Payload     []byte  `json:"payload" db:"payload"`
	Description *string `json:"description" db:"description"`
	Debug       *bool   `json:"debug" db:"debug"`
	Editor      *bool   `json:"editor" db:"editor"`
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

	if in.Editor != nil {
		s.Editor = *in.Editor
	}
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

type RoutingOutboundCall struct {
	DomainRecord
	Name        string `json:"name" db:"name"`
	Description string `json:"description" db:"description"`
	Schema      Lookup `json:"schema" db:"schema"`
	Position    int    `json:"position" db:"position"`
	Pattern     string `json:"pattern" db:"pattern"`
	Disabled    bool   `json:"disabled" db:"disabled"`
}

type SearchRoutingOutboundCall struct {
	ListRequest
	Ids         []uint32
	Name        *string
	SchemaIds   []uint32
	Pattern     *string
	Description *string
}

func (RoutingOutboundCall) DefaultOrder() string {
	return "+position"
}

func (a RoutingOutboundCall) AllowFields() []string {
	return []string{"id", "domain_id", "name", "description", "created_at", "created_by", "updated_at", "updated_by",
		"pattern", "disabled", "schema", "position"}
}

func (a RoutingOutboundCall) DefaultFields() []string {
	return []string{"id", "name", "description",
		"pattern", "disabled", "schema", "position"}
}

func (a RoutingOutboundCall) EntityName() string {
	return "acr_routing_outbound_call_view"
}

type RoutingOutboundCallPatch struct {
	UpdatedById int
	Name        *string `json:"name" db:"name"`
	Description *string `json:"description" db:"description"`
	Schema      *Lookup `json:"schema" db:"scheme"`
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
