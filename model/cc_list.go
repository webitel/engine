package model

import "time"

type List struct {
	DomainRecord
	Name        string `json:"name" db:"name"`
	Description string `json:"description,omitempty"`
	Count       int64  `json:"count" db:"count"`
}

type SearchList struct {
	ListRequest
	Ids []uint32
}

func (List) DefaultOrder() string {
	return "id"
}

func (a List) AllowFields() []string {
	return []string{"id", "name", "description", "count", "domain_id", "created_at", "created_by", "updated_at", "updated_by"}
}

func (a List) DefaultFields() []string {
	return []string{"id", "name", "description", "count"}
}

func (a List) EntityName() string {
	return "cc_list_view"
}

func (l *List) IsValid() AppError {
	//FIXME
	return nil
}

type ListCommunication struct {
	Id          int64      `json:"id" db:"id"`
	ListId      int64      `json:"list_id" db:"list_id"`
	Number      string     `json:"number" db:"number"`
	Description string     `json:"description" db:"description"`
	ExpireAt    *time.Time `json:"expire_at" db:"expire_at"`
}

type SearchListCommunication struct {
	ListRequest
	Ids      []uint32
	ExpireAt *FilterBetween
}

func (ListCommunication) DefaultOrder() string {
	return "number"
}

func (a ListCommunication) AllowFields() []string {
	return []string{"id", "number", "description", "list_id", "domain_id", "expire_at"}
}

func (a ListCommunication) DefaultFields() []string {
	return []string{"id", "number", "description", "expire_at"}
}

func (a ListCommunication) EntityName() string {
	return "cc_list_communications_view"
}

func (l *ListCommunication) IsValid() AppError {
	//FIXME
	return nil
}
