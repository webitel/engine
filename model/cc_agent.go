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
	Id              int64  `json:"id" db:"id"`
	DomainId        int64  `json:"domain_id" db:"domain_id"`
	User            Lookup `json:"user" db:"user"`
	Status          string `json:"status" db:"status"`
	State           string `json:"state" db:"state"`
	LastStateChange int64  `json:"last_state_change" db:"last_state_change"`
	StateTimeout    *int64 `json:"state_timeout" db:"state_timeout"`
	Description     string `json:"description" db:"description"`
}

type AgentState struct {
	Id        int64  `json:"id" db:"id"`
	JoinedAt  int64  `json:"joined_at" db:"joined_at"`
	State     string `json:"state" db:"state"`
	TimeoutAt *int64 `json:"timeout_at" db:"timeout_at"`
	QueueId   *int64 `json:"queue_id" db:"queue_id"`
}

func (a *Agent) IsValid() *AppError {
	return nil //TODO
}

type AgentInTeam struct {
	Team     Lookup `json:"team" db:"team"`
	Strategy string `json:"strategy" json:"strategy"`
}

type AgentInQueue struct {
	Queue          Lookup `json:"queue" db:"queue"`
	Priority       int    `json:"priority" db:"priority"`
	Type           int    `json:"type" db:"type"`
	Strategy       string `json:"strategy" db:"strategy"`
	Enabled        bool   `json:"enabled" db:"enabled"`
	CountMembers   int    `json:"count_members" db:"count_members"`
	WaitingMembers int    `json:"waiting_members" db:"waiting_members"`
	ActiveMembers  int    `json:"active_members" db:"active_members"`
}
