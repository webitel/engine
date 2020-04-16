package model

import "encoding/json"

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
	DomainRecord
	User             Lookup `json:"user" db:"user"`
	Status           string `json:"status" db:"status"`
	State            string `json:"state" db:"state"`
	LastStateChange  int64  `json:"last_state_change" db:"last_state_change"`
	StateTimeout     *int64 `json:"state_timeout" db:"state_timeout"`
	Description      string `json:"description" db:"description"`
	ProgressiveCount int    `json:"progressive_count" db:"progressive_count"`
}

func (a Agent) AllowFields() []string {
	return []string{"id", "status", "state", "description", "last_state_change", "state_timeout", "progressive_count", "user"}
}

func (a Agent) DefaultFields() []string {
	return []string{"id", "status", "state", "description", "last_state_change", "state_timeout", "progressive_count", "user"}
}

func (a Agent) EntityName() string {
	return "cc_agent_list"
}

type AgentSession struct {
	AgentId         int64  `json:"agent_id" db:"agent_id"`
	Status          string `json:"status" db:"status"`
	LastStateChange int64  `json:"last_state_change" db:"last_state_change"`
	AttemptId       *int64 `json:"attempt_id" db:"attempt_id"`
	StateDuration   int    `json:"state_duration" db:"state_duration"`
	StateTimeout    *int64 `json:"state_timeout" db:"state_timeout"`
	StatusPayload   []byte `json:"status_payload" db:"status_payload"`
}

func (a AgentSession) ToMap() map[string]interface{} {
	out := make(map[string]interface{})
	data, _ := json.Marshal(a)
	_ = json.Unmarshal(data, &out)
	return out
}

type SearchAgent struct {
	ListRequest
	Ids []string
}

type AgentUser struct {
	Id   int64
	Name string
}

type SearchAgentUser struct {
	ListRequest
}

type AgentState struct {
	Id        int64   `json:"id" db:"id"`
	JoinedAt  int64   `json:"joined_at" db:"joined_at"`
	State     string  `json:"state" db:"state"`
	TimeoutAt *int64  `json:"timeout_at" db:"timeout_at"`
	Queue     *Lookup `json:"queue" db:"queue"`
}

type SearchAgentState struct {
	ListRequest
	From int64
	To   int64
}

func (a *Agent) IsValid() *AppError {
	return nil //TODO
}

type AgentInTeam struct {
	Team     Lookup `json:"team" db:"team"`
	Strategy string `json:"strategy" json:"strategy"`
}

type SearchAgentInTeam struct {
	ListRequest
}

type SearchAgentInQueue struct {
	ListRequest
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

type AgentInQueueStatistic struct {
	Queue      Lookup               `json:"queue" db:"queue"`
	Statistics []*AgentInQueueStats `json:"statistics" db:"statistics"`
}

type AgentInQueueStats struct {
	Bucket        *Lookup `json:"bucket" db:"bucket"`
	Skill         *Lookup `json:"skill" db:"skill"`
	MemberWaiting int     `json:"member_waiting" db:"member_waiting"`
}

type AgentStatusEvent struct {
	UserId        int64  `json:"user_id"`
	AgentId       int    `json:"agent_id"`
	Timestamp     int64  `json:"timestamp"`
	AttemptId     *int64 `json:"attempt_id"`
	Status        string `json:"status"`
	StatusPayload string `json:"status_payload"`
	Timeout       *int   `json:"timeout"`
}

func NewWebSocketAgentStatusEvent(status *AgentStatusEvent) *WebSocketEvent {
	e := NewWebSocketEvent(WebsocketEventAgentStatus)
	e.Add("status", status)

	return e
}
