package model

import (
	"encoding/json"
	"time"
)

type OutboundResourceParameters struct {
	CidType          string `json:"cid_type"`
	IgnoreEarlyMedia string `json:"ignore_early_media"`
}

type OutboundCallResource struct {
	DomainRecord
	Limit                 int                        `json:"limit" db:"limit"`
	Enabled               bool                       `json:"enabled" db:"enabled"`
	RPS                   int                        `json:"rps" db:"rps"`
	Reserve               bool                       `json:"reserve" db:"reserve"`
	Variables             StringMap                  `json:"variables" db:"variables"`
	Number                string                     `json:"number" db:"number"`
	MaxSuccessivelyErrors int                        `json:"max_successively_errors" db:"max_successively_errors"`
	Name                  string                     `json:"name" db:"name"`
	ErrorIds              StringArray                `json:"error_ids" db:"error_ids"`
	LastErrorId           *string                    `json:"last_error_id" db:"last_error_id"`
	SuccessivelyErrors    int                        `json:"successively_errors" db:"successively_errors"`
	LastErrorAt           *time.Time                 `json:"last_error_at" db:"last_error_at"`
	Gateway               *Lookup                    `json:"gateway" db:"gateway"`
	Description           *string                    `json:"description" db:"description"`
	Patterns              StringArray                `json:"patterns" db:"patterns"`
	FailureDialDelay      uint32                     `json:"failure_dial_delay" db:"failure_dial_delay"`
	Parameters            OutboundResourceParameters `json:"parameters" db:"parameters"`
}

type SearchOutboundCallResource struct {
	ListRequest
	Ids []uint32
}

func (params OutboundResourceParameters) ToJson() string {
	data, _ := json.Marshal(params)
	return string(data)
}

func (OutboundCallResource) DefaultOrder() string {
	return "id"
}

func (a OutboundCallResource) AllowFields() []string {
	return []string{"id", "name", "gateway", "enabled", "reserve", "limit",
		"domain_id", "rps", "variables", "number", "max_successively_errors", "error_ids", "last_error_id", "successively_errors", "last_error_at",
		"created_at", "created_by", "updated_at", "updated_by", "description", "patterns", "failure_dial_delay"}
}

func (a OutboundCallResource) DefaultFields() []string {
	return []string{"id", "name", "gateway", "enabled", "reserve", "limit", "description"}
}

func (a OutboundCallResource) EntityName() string {
	return "cc_outbound_resource_view"
}

type ResourceDisplay struct {
	Id         int64  `json:"id" db:"id"`
	Display    string `json:"display" db:"display"`
	ResourceId int64  `json:"resource_id" db:"resource_id"`
}

type SearchResourceDisplay struct {
	ListRequest
	Ids []uint32
}

func (ResourceDisplay) DefaultOrder() string {
	return "display"
}

func (a ResourceDisplay) AllowFields() []string {
	return []string{"id", "display", "resource_id", "domain_id"}
}

func (a ResourceDisplay) DefaultFields() []string {
	return []string{"id", "display"}
}

func (a ResourceDisplay) EntityName() string {
	return "cc_outbound_resource_display_view"
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
	Description           *string      `json:"description" db:"description"`
	FailureDialDelay      *uint32      `json:"failure_dial_delay" db:"failure_dial_delay"`
}

func (r *OutboundCallResource) GetGatewayId() *int {
	if r.Gateway != nil {
		return NewInt(r.Gateway.Id)
	}
	return nil
}

func (d *ResourceDisplay) IsValid() *AppError {
	//FIXME
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

	if p.Description != nil {
		r.Description = p.Description
	}

	if p.FailureDialDelay != nil {
		r.FailureDialDelay = *p.FailureDialDelay
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
