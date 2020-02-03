package model

type CommunicationType struct {
	Id          int64  `json:"id" db:"id"`
	DomainId    int64  `json:"domain_id" db:"domain_id"`
	Name        string `json:"name" db:"name"`
	Code        string `json:"code" db:"code"`
	Type        string `json:"type" db:"type"`
	Description string `json:"description" db:"description"`
}

type SearchCommunicationType struct {
	ListRequest
}

func (s *CommunicationType) IsValid() *AppError {
	return nil
}
