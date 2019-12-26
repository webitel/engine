package model

type List struct {
	DomainRecord
	Name        string `json:"name" db:"name"`
	Description string `json:"description,omitempty"`
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

func (l *ListCommunication) IsValid() *AppError {
	//FIXME
	return nil
}