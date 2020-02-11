package model

type SupervisorInTeam struct {
	Id     int64  `json:"id" db:"id"`
	TeamId int64  `json:"team_id" db:"team_id"`
	Agent  Lookup `json:"agent" db:"agent"`
}

type SearchSupervisorInTeam struct {
	ListRequest
}

func (s *SupervisorInTeam) IsValid() *AppError {
	//FIXME
	return nil
}
