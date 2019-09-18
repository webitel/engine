package model

type AgentStatus struct {
	Name string
}

var (
	AgentStatusOnline  = AgentStatus{"online"}
	AgentStatusOffline = AgentStatus{"offline"}
	AgentStatusPause   = AgentStatus{"pause"}
)

func (status AgentStatus) String() string {
	return status.Name
}

type Agent struct {
	Id          int64  `json:"id" db:"id"`
	DomainId    int64  `json:"domain_id" db:"domain_id"`
	User        Lookup `json:"user" db:"user"`
	Status      string `json:"status" db:"status"`
	State       string `json:"state" db:"state"`
	Description string `json:"description" db:"description"`
}

func (a *Agent) IsValid() *AppError {
	return nil //TODO
}
