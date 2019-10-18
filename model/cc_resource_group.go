package model

type OutboundResourceGroup struct {
	DomainRecord
	Name          string `json:"name" db:"name"`
	Strategy      string `json:"strategy" db:"strategy"`
	Description   string `json:"description" db:"description"`
	Communication Lookup `json:"communication" db:"communication"`
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

func (r *OutboundResourceInGroup) IsValid() *AppError {
	///FIXME
	return nil
}
