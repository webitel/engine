package model

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

func (l *List) IsValid() *AppError {
	//FIXME
	return nil
}

type ListCommunication struct {
	Id          int64  `json:"id" db:"id"`
	ListId      int64  `json:"list_id" db:"list_id"`
	Number      string `json:"number" db:"number"`
	Description string `json:"description" db:"description"`
}

type SearchListCommunication struct {
	ListRequest
}

func (l *ListCommunication) IsValid() *AppError {
	//FIXME
	return nil
}
