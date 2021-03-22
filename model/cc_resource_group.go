package model

import "encoding/json"

type OutboundResourceGroup struct {
	DomainRecord
	Name          string                      `json:"name" db:"name"`
	Strategy      string                      `json:"strategy" db:"strategy"`
	Description   string                      `json:"description" db:"description"`
	Communication Lookup                      `json:"communication" db:"communication"`
	Time          []OutboundResourceGroupTime `json:"time" db:"time"`
}

type SearchOutboundResourceGroup struct {
	ListRequest
	Ids []uint32
}

func (OutboundResourceGroup) DefaultOrder() string {
	return "id"
}

func (a OutboundResourceGroup) AllowFields() []string {
	return []string{"id", "domain_id", "name", "description", "communication", "time", "communication_id",
		"created_at", "created_by", "updated_at", "updated_by"}
}

func (a OutboundResourceGroup) DefaultFields() []string {
	return []string{"id", "name", "description", "communication"}
}

func (a OutboundResourceGroup) EntityName() string {
	return "cc_outbound_resource_group_view"
}

type OutboundResourceGroupTime struct {
	StartTimeOfDay int16 `json:"start_time_of_day" db:"start_time_of_day"`
	EndTimeOfDay   int16 `json:"end_time_of_day" db:"end_time_of_day"`
}

func OutboundResourceGroupTimesToJson(times []OutboundResourceGroupTime) string {
	data, _ := json.Marshal(times)
	return string(data)
}

func (g *OutboundResourceGroup) IsValid() *AppError {
	//FIXME
	return nil
}

type OutboundResourceInGroup struct {
	Id       int64  `json:"id" db:"id"`
	GroupId  int64  `json:"group_id" db:"group_id"`
	Resource Lookup `json:"resource" db:"resource"`
}

type SearchOutboundResourceInGroup struct {
	ListRequest
	Ids []uint32
}

func (OutboundResourceInGroup) DefaultOrder() string {
	return "resource_name"
}

func (a OutboundResourceInGroup) AllowFields() []string {
	return []string{"id", "resource", "group_id", "resource_id", "resource_name", "domain_id"}
}

func (a OutboundResourceInGroup) DefaultFields() []string {
	return []string{"id", "resource"}
}

func (a OutboundResourceInGroup) EntityName() string {
	return "cc_outbound_resource_in_group_view"
}

func (r *OutboundResourceInGroup) IsValid() *AppError {
	///FIXME
	return nil
}
