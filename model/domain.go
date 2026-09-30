package model

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type DomainEvent struct {
	ID int64
}

func NewDomainEventFromRoutingKey(rk string) (*DomainEvent, AppError) {
	splitted := strings.Split(rk, ".")
	if len(splitted) < 3 {
		return nil, NewBadRequestError(
			"model.domain.new_domain_event.invalid_rk_len",
			"received roting key with len less than 3",
		)
	}

	const domainIdIndex = 2

	parsedDomainId, err := strconv.ParseInt(splitted[domainIdIndex], 10, 64)
	if err != nil {
		return nil, NewBadRequestError(
			"model.domain.new_domain_event.parsing_id",
			fmt.Sprintf("parsing routing key id to integer: %+v", err),
		)
	}

	if parsedDomainId <= 0 {
		return nil, NewBadRequestError(
			"model.domain.new_domain_event.domain_id_less_or_equal_zero",
			fmt.Sprintf("received domain id less or equal zero: %d", parsedDomainId),
		)
	}

	return &DomainEvent{ID: parsedDomainId}, nil
}

type DomainProvider interface {
	Domain(int64) int64
}

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
	CreatedAt *time.Time `json:"created_at" db:"created_at"`
	CreatedBy *Lookup    `json:"created_by" db:"created_by"`
	UpdatedAt *time.Time `json:"updated_at" db:"updated_at"`
	UpdatedBy *Lookup    `json:"updated_by" db:"updated_by"`
}
