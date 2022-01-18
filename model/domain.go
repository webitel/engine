package model

import "time"

// todo deprecated
type DomainRecord struct {
	Id        int64   `json:"id" db:"id"`
	DomainId  int64   `json:"domain_id" db:"domain_id"`
	CreatedAt int64   `json:"created_at" db:"created_at"`
	CreatedBy *Lookup `json:"created_by" db:"created_by"`
	UpdatedAt int64   `json:"updated_at" db:"updated_at"`
	UpdatedBy *Lookup `json:"updated_by" db:"updated_by"`
}

type AclRecord struct {
	Id int `json:"id" db:"id"`
	//DomainId  int64      `json:"domain_id" db:"domain_id"`
	CreatedAt *time.Time `json:"created_at" db:"created_at"`
	CreatedBy Lookup     `json:"created_by" db:"created_by"`
	UpdatedAt *time.Time `json:"updated_at" db:"updated_at"`
	UpdatedBy Lookup     `json:"updated_by" db:"updated_by"`
}
