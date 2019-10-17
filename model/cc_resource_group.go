package model

type OutboundResourceGroup struct {
	Id       int64  `json:"id" db:"id"`
	DomainId int64  `json:"domain_id" db:"domain_id"`
	Name     string `json:"name" db:"name"`
	Strategy string `json:"strategy" db:"strategy"`
}
