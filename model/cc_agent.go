package model

import (
	"encoding/json"
	"fmt"
	"net/http"
)

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

type AgentChannel struct {
	Channel  string `json:"channel"`
	State    string `json:"state"`
	JoinedAt int64  `json:"joined_at"`
	Timeout  *int64 `json:"timeout,omitempty"`
	Active   bool   `json:"active"`
	Open     int    `json:"open"`
	MaxOpen  int    `json:"max_open"`
	NoAnswer int    `json:"no_answer"`
}

type Agent struct {
	DomainRecord
	User             Lookup         `json:"user" db:"user"`
	Name             string         `json:"name" db:"name"`
	Status           string         `json:"status" db:"status"`
	LastStatusChange int64          `json:"last_status_change" db:"last_status_change"`
	Description      string         `json:"description" db:"description"`
	ProgressiveCount int            `json:"progressive_count" db:"progressive_count"`
	Channels         []AgentChannel `json:"channels" db:"channels"`
}

func (a Agent) AllowFields() []string {
	return a.DefaultFields()
}

func (a Agent) DefaultFields() []string {
	return []string{"id", "status", "name", "channels", "description", "last_status_change", "progressive_count", "user"}
}

func (a Agent) EntityName() string {
	return "cc_agent_list"
}

type AgentSession struct {
	AgentId          int64          `json:"agent_id" db:"agent_id"`
	Status           string         `json:"status" db:"status"`
	StatusPayload    *string        `json:"status_payload" db:"status_payload"`
	LastStatusChange int64          `json:"last_status_change" db:"last_status_change"`
	StatusDuration   int64          `json:"status_duration" db:"status_duration"`
	OnDemand         bool           `json:"on_demand" db:"on_demand"`
	Channels         []AgentChannel `json:"channels" db:"channels"`
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

type AgentCallStatistics struct {
	Name      string `json:"name" db:"name"`
	Count     int32  `json:"count" db:"count"`
	Abandoned int32  `json:"abandoned" db:"abandoned"`
	Handles   int32  `json:"handles" db:"handles"`

	SumTalkSec float32 `json:"sum_talk_sec" db:"sum_talk_sec"`
	AvgTalkSec float32 `json:"avg_talk_sec" db:"avg_talk_sec"`
	MinTalkSec float32 `json:"min_talk_sec" db:"min_talk_sec"`
	MaxTalkSec float32 `json:"max_talk_sec" db:"max_talk_sec"`

	SumHoldSec float32 `json:"sum_hold_sec" db:"sum_hold_sec"`
	AvgHoldSec float32 `json:"avg_hold_sec" db:"avg_hold_sec"`
	MinHoldSec float32 `json:"min_hold_sec" db:"min_hold_sec"`
	MaxHoldSec float32 `json:"max_hold_sec" db:"max_hold_sec"`
}

type SearchAgentCallStatistics struct {
	ListRequest
	Time     FilterBetween
	AgentIds []int32
}

func (c AgentCallStatistics) AllowFields() []string {
	return []string{"name", "count", "abandoned", "handles", "sum_talk_sec", "avg_talk_sec", "min_talk_sec", "max_talk_sec",
		"sum_hold_sec", "avg_hold_sec", "min_hold_sec", "max_hold_sec",
	}
}

func (c AgentCallStatistics) DefaultFields() []string {
	return []string{"name", "count", "abandoned", "handles", "sum_talk_sec", "sum_hold_sec"}
}

func (c AgentCallStatistics) EntityName() string {
	return ""
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

type CallCenterPayload map[string]interface{}

type CallCenterEvent struct {
	Event  string            `json:"event"`
	UserId int64             `json:"user_id"`
	Body   CallCenterPayload `json:"data,string,omitempty"`
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

func NewWebSocketCallCenterEvent(ev *CallCenterEvent) (*WebSocketEvent, *AppError) {
	var e *WebSocketEvent

	switch ev.Event {
	case WebsocketCCEventAgentStatus, WebsocketCCEventChannelStatus:
		e = NewWebSocketEvent(ev.Event)
	default:
		return nil, NewAppError("Event", "event.cc.valid.event", nil,
			fmt.Sprintf("unknown event \"%s\"", ev.Event), http.StatusInternalServerError)
	}
	e.UserId = ev.UserId
	e.SetData(ev.Body)

	return e, nil
}
