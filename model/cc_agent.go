package model

import (
	"encoding/json"
	"fmt"
	"time"
)

const (
	AgentStatusOnline  = "online"
	AgentStatusOffline = "offline"
	AgentStatusPause   = "pause"
)

type AgentChannel struct {
	Channel     string  `json:"channel"`
	State       string  `json:"state"`
	JoinedAt    int64   `json:"joined_at"`
	Timeout     *int64  `json:"timeout,omitempty"`
	Open        int     `json:"open"`
	MaxOpen     int     `json:"max_open"`
	NoAnswer    int     `json:"no_answer"`
	WrapTimeIds []int64 `json:"wrap_time_ids,omitempty"`
}

type Agent struct {
	DomainRecord
	User                  Lookup         `json:"user" db:"user"`
	Name                  string         `json:"name" db:"name"`
	Status                string         `json:"status" db:"status"`
	LastStatusChange      int64          `json:"last_status_change" db:"last_status_change"`
	StatusDuration        int64          `json:"status_duration" db:"status_duration"`
	Description           string         `json:"description" db:"description"`
	ProgressiveCount      int            `json:"progressive_count" db:"progressive_count"`
	Channel               []AgentChannel `json:"channel" db:"channel"`
	GreetingMedia         *Lookup        `json:"greeting_media" db:"greeting_media"`
	AllowChannels         StringArray    `json:"allow_channels" db:"allow_channels"`
	ChatCount             uint32         `json:"chat_count" db:"chat_count"`
	Supervisor            []*Lookup      `json:"supervisor" db:"supervisor"`
	Team                  *Lookup        `json:"team" db:"team"`
	Region                *Lookup        `json:"region" db:"region"`
	Auditor               []*Lookup      `json:"auditor" db:"auditor"`
	IsSupervisor          bool           `json:"is_supervisor" db:"is_supervisor"`
	Skills                []*Lookup      `json:"skills" db:"skills"`
	Extension             *string        `json:"extension" db:"extension"`
	TaskCount             uint32         `json:"task_count" db:"task_count"`
	ScreenControl         bool           `json:"screen_control" db:"screen_control"`
	AllowSetScreenControl bool           `json:"allow_set_screen_control" db:"allow_set_screen_control"`
}

type AgentPatch struct {
	UpdatedBy        Lookup
	UpdatedAt        int64
	User             *Lookup
	Description      *string
	ProgressiveCount *int
	GreetingMedia    *Lookup
	ChatCount        *uint32
	Supervisor       []*Lookup
	Team             *Lookup
	Region           *Lookup
	Auditor          []*Lookup
	IsSupervisor     *bool
	ScreenControl    *bool
}

type AgentStatusStatistics struct {
	AgentId        int32      `json:"agent_id" db:"agent_id"`
	Name           string     `json:"name" db:"name"`
	Status         string     `json:"status" db:"status"`
	StatusDuration int64      `json:"status_duration" db:"status_duration"`
	StatusComment  *string    `json:"status_comment" db:"status_comment"`
	User           Lookup     `json:"user" json:"user"`
	Extension      string     `json:"extension" db:"extension"`
	Team           *Lookup    `json:"team" db:"team"`
	Queues         []*Lookup  `json:"queues" db:"queues"`
	Online         int64      `json:"online" db:"online"`
	Offline        int64      `json:"offline" db:"offline"`
	Pause          int64      `json:"pause" db:"pause"`
	Utilization    float32    `json:"utilization" db:"utilization"`
	CallTime       int64      `json:"call_time" db:"call_time"`
	ActiveCallId   *string    `json:"active_call_id" db:"active_call_id"`
	Handles        int32      `json:"handles" db:"handles"`
	Missed         int32      `json:"missed" db:"missed"`
	MaxBridgedAt   *time.Time `json:"max_bridged_at" db:"max_bridged_at"`
	MaxOfferingAt  *time.Time `json:"max_offering_at" db:"max_offering_at"`

	Transferred uint32    `json:"transferred" db:"transferred"`
	Skills      []*Lookup `json:"skills" db:"skills"`
	Supervisor  []*Lookup `json:"supervisor" db:"supervisor"`
	Auditor     []*Lookup `json:"auditor" db:"auditor"`
	PauseCause  string    `json:"pause_cause" db:"pause_cause"`
	ChatCount   int32     `json:"chat_count" db:"chat_count"`

	Occupancy float32 `json:"occupancy" db:"occupancy"`
	DescTrack bool    `json:"desc_track" db:"desc_track"`
}

type SupervisorAgentItem struct {
	AgentId        int32  `json:"agent_id" db:"agent_id"`
	Name           string `json:"name" db:"name"`
	Status         string `json:"status" db:"status"`
	StatusDuration int64  `json:"status_duration" db:"status_duration"`
	User           Lookup `json:"user" json:"user"`
	Extension      string `json:"extension" db:"extension"`

	Team             *Lookup   `json:"team" db:"team"`
	Supervisor       []*Lookup `json:"supervisor" dlb:"supervisor"`
	Auditor          []*Lookup `json:"auditor" db:"auditor"`
	Region           *Lookup   `json:"region" db:"region"`
	ProgressiveCount uint32    `json:"progressive_count" db:"progressive_count"`
	ChatCount        uint32    `json:"chat_count" db:"chat_count"`

	PauseCause    string `json:"pause_cause" db:"pause_cause"`
	StatusComment string `json:"status_comment" db:"status_comment"`

	Online           int64   `json:"online" db:"online"`
	Offline          int64   `json:"offline" db:"offline"`
	Pause            int64   `json:"pause" db:"pause"`
	ScoreRequiredAvg float32 `json:"score_required_avg" db:"score_required_avg"`
	ScoreOptionalAvg float32 `json:"score_optional_avg" db:"score_optional_avg"`
	ScoreCount       int64   `json:"score_count" db:"score_count"`
	DescTrack        bool    `json:"desc_track" db:"desc_track"`
}

type SearchAgentStatusStatistic struct {
	ListRequest
	Time          FilterBetween
	Utilization   *FilterBetween
	AgentIds      []int64
	Status        []string
	TeamIds       []int32
	QueueIds      []int32
	SkillIds      []uint32
	RegionIds     []uint32
	SupervisorIds []uint32
	AuditorIds    []int64

	HasCall bool
}

type AgentPauseCause struct {
	Id          uint32 `json:"id" db:"id"`
	Name        string `json:"name" db:"name"`
	LimitMin    uint32 `json:"limit_min" db:"limit_min"`
	DurationMin uint32 `json:"duration_min" db:"duration_min"`
}

type DistributeAgentInfo struct {
	AgentId    int32 `json:"agent_id" db:"agent_id"`
	Distribute bool  `json:"distribute" db:"distribute"`
	Busy       bool  `json:"busy" db:"busy"`
}

func (a *Agent) Patch(patch *AgentPatch) {
	a.UpdatedBy = &patch.UpdatedBy
	a.UpdatedAt = patch.UpdatedAt

	if patch.User != nil {
		a.User = *patch.User
	}

	if patch.Description != nil {
		a.Description = *patch.Description
	}

	if patch.ProgressiveCount != nil {
		a.ProgressiveCount = *patch.ProgressiveCount
	}

	if patch.GreetingMedia != nil {
		a.GreetingMedia = patch.GreetingMedia
	}

	if patch.ChatCount != nil {
		a.ChatCount = *patch.ChatCount
	}

	if patch.Supervisor != nil {
		a.Supervisor = patch.Supervisor
	}

	if patch.Team != nil {
		a.Team = patch.Team
	}

	if patch.Region != nil {
		a.Region = patch.Region
	}

	if patch.Auditor != nil {
		a.Auditor = patch.Auditor
	}

	if patch.IsSupervisor != nil {
		a.IsSupervisor = *patch.IsSupervisor
	}

	if patch.ScreenControl != nil {
		a.ScreenControl = *patch.ScreenControl
	}
}

func (a Agent) DefaultOrder() string {
	return "id"
}

func (a Agent) AllowFields() []string {
	return []string{"id", "status", "name", "channel", "description", "status_duration", "last_status_change",
		"progressive_count", "user", "greeting_media", "allow_channels", "chat_count", "supervisor", "team", "region",
		"auditor", "is_supervisor", "skills", "extension", "task_count", "screen_control", "allow_set_screen_control"}
}

func (a Agent) DefaultFields() []string {
	return []string{"id", "status", "name", "channel", "description", "status_duration", "last_status_change",
		"progressive_count", "user", "greeting_media", "allow_channels", "chat_count", "supervisor", "team", "region",
		"auditor", "is_supervisor", "extension"}
}

func (a Agent) EntityName() string {
	return "cc_agent_list"
}

func (a *Agent) GreetingMediaId() *int {
	if a.GreetingMedia != nil && a.GreetingMedia.Id > 0 {
		return &a.GreetingMedia.Id
	}

	return nil
}

type AgentSession struct {
	AgentId          int64          `json:"agent_id" db:"agent_id"`
	Status           string         `json:"status" db:"status"`
	StatusPayload    *string        `json:"status_payload" db:"status_payload"`
	StatusComment    *string        `json:"status_comment" db:"status_comment"`
	LastStatusChange int64          `json:"last_status_change" db:"last_status_change"`
	StatusDuration   int64          `json:"status_duration" db:"status_duration"`
	OnDemand         bool           `json:"on_demand" db:"on_demand"`
	Team             *Lookup        `json:"team" db:"team"`
	IsSupervisor     bool           `json:"is_supervisor" db:"is_supervisor"`
	IsAdmin          bool           `json:"is_admin" db:"is_admin"`
	Channels         []AgentChannel `json:"channels" db:"channels"`

	Supervisor    []*Lookup `json:"supervisor" db:"supervisor"`
	Auditor       []*Lookup `json:"auditor" db:"auditor"`
	ScreenControl bool      `json:"screen_control" db:"screen_control"`
}

type AgentCC struct {
	HasAgent     bool   `json:"has_agent" db:"has_agent"`
	HasExtension bool   `json:"has_extension" db:"has_extension"`
	AgentId      *int64 `json:"agent_id" db:"agent_id"`
}

func (a *AgentCC) Valid() AppError {
	if !a.HasAgent {
		return NewNotFoundError("User.valid.agent_id", "")
	}
	if !a.HasExtension {
		//return NewNotFoundError("User.valid.extension", "")
	}

	return nil
}

func (a AgentSession) ToMap() map[string]interface{} {
	out := make(map[string]interface{})
	data, _ := json.Marshal(a)
	_ = json.Unmarshal(data, &out)
	return out
}

type SearchAgent struct {
	ListRequest
	Ids           []string
	AllowChannels []string
	SupervisorIds []uint32
	TeamIds       []uint32
	RegionIds     []uint32
	AuditorIds    []uint32
	SkillIds      []uint32
	QueueIds      []uint32
	IsSupervisor  *bool
	NotSupervisor *bool `json:"not_supervisor"`
	Extensions    []string
	UserIds       []int64
	NotTeamIds    []uint32
	NotSkillIds   []uint32
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

	SumHoldSec  float32 `json:"sum_hold_sec" db:"sum_hold_sec"`
	AvgHoldSec  float32 `json:"avg_hold_sec" db:"avg_hold_sec"`
	MinHoldSec  float32 `json:"min_hold_sec" db:"min_hold_sec"`
	MaxHoldSec  float32 `json:"max_hold_sec" db:"max_hold_sec"`
	Utilization float32 `json:"utilization" db:"utilization"`
	Occupancy   float32 `json:"occupancy" db:"occupancy"`
	ChatAccepts int32   `json:"chat_accepts" db:"chat_accepts"`
	ChatAHT     int32   `json:"chat_aht" db:"chat_aht"`
}

type SearchAgentCallStatistics struct {
	ListRequest
	Time     FilterBetween
	AgentIds []int32
}

func (c AgentCallStatistics) DefaultOrder() string {
	return "name"
}

func (c AgentCallStatistics) AllowFields() []string {
	return []string{"name", "count", "abandoned", "handles", "sum_talk_sec", "avg_talk_sec", "min_talk_sec", "max_talk_sec",
		"sum_hold_sec", "avg_hold_sec", "min_hold_sec", "max_hold_sec", "utilization", "occupancy", "chat_accepts", "chat_aht",
	}
}

func (c AgentCallStatistics) DefaultFields() []string {
	return []string{"name", "count", "abandoned", "handles", "sum_talk_sec", "sum_hold_sec", "utilization", "occupancy", "chat_accepts", "chat_aht"}
}

func (c AgentCallStatistics) EntityName() string {
	return ""
}

type SearchAgentUser struct {
	ListRequest
}

type UserStatus struct {
	Id        int64       `json:"id" db:"id"`
	Name      string      `json:"name" db:"name"`
	Extension string      `json:"extension" db:"extension"`
	Presence  StringArray `json:"presence" db:"presence"`
	Status    string      `json:"status" db:"status"`
}

type SearchUserStatus struct {
	ListRequest
}

type AgentState struct {
	Id       int64      `json:"id" db:"id"`
	Channel  *string    `json:"channel" db:"channel"`
	Agent    *Lookup    `json:"agent" db:"agent"`
	Queue    *Lookup    `json:"queue" db:"queue"`
	JoinedAt *time.Time `json:"joined_at" db:"joined_at"`
	Duration int64      `json:"duration" db:"duration"`
	State    string     `json:"state" db:"state"`
	Payload  *string    `json:"payload" db:"payload"`
}

type SearchAgentState struct {
	ListRequest
	JoinedAt FilterBetween
	AgentIds []int64
	FromId   *int64
}

func (a *Agent) IsValid() AppError {
	//todo fire error ?
	if a.IsSupervisor && a.Supervisor != nil {
		a.Supervisor = nil
	}
	if a.TaskCount < 1 {
		return NewBadRequestError("model.Agent.valid.TaskCount", "The task count should be more or equal 1")
	}
	if a.ChatCount < 1 {
		return NewBadRequestError("model.Agent.valid.ChatCount", "The chat count should be more or equal 1")
	}
	if a.ProgressiveCount < 1 {
		return NewBadRequestError("model.Agent.valid.ProgressiveCount", "The call count should be more or equal 1")
	}
	return nil //TODO
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
	Queue          Lookup        `json:"queue" db:"queue"`
	Priority       int           `json:"priority" db:"priority"`
	Type           int           `json:"type" db:"type"`
	Strategy       string        `json:"strategy" db:"strategy"`
	Enabled        bool          `json:"enabled" db:"enabled"`
	CountMembers   int           `json:"count_members" db:"count_members"`
	WaitingMembers int           `json:"waiting_members" db:"waiting_members"`
	ActiveMembers  int           `json:"active_members" db:"active_members"`
	MaxMemberLimit int           `json:"max_member_limit" db:"max_member_limit"`
	Agents         QueueAgentAgg `json:"agents" db:"agents"`
}

func (AgentInQueue) DefaultOrder() string {
	return "queue_name"
}

func (a AgentInQueue) AllowFields() []string {
	return []string{"queue", "type", "strategy", "count_members", "waiting_members", "active_members",
		"priority", "enabled", "queue_id", "queue_name", "domain_id", "agent_id", "agents", "max_member_limit"}
}

func (a AgentInQueue) DefaultFields() []string {
	return []string{"queue", "type", "strategy", "count_members", "waiting_members", "active_members", "agents"}
}

func (a AgentInQueue) EntityName() string {
	return "cc_agent_in_queue_view"
}

func (UserStatus) DefaultOrder() string {
	return "position"
}

func (a UserStatus) AllowFields() []string {
	return []string{"id", "name", "extension", "presence", "status"}
}

func (a UserStatus) DefaultFields() []string {
	return a.AllowFields()
}

func (a UserStatus) EntityName() string {
	return "cc_user_status_view"
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

type AgentTask struct {
	AttemptId       int64               `json:"attempt_id" db:"attempt_id"`
	AppId           string              `json:"app_id" db:"app_id"`
	Channel         *string             `json:"channel" db:"channel"`
	QueueId         int                 `json:"queue_id" db:"queue_id"`
	MemberId        *int64              `json:"member_id" db:"member_id"`
	AgentId         int                 `json:"agent_id" db:"agent_id"`
	MemberChannelId *string             `json:"member_channel_id" db:"member_channel_id"`
	AgentChannelId  *string             `json:"agent_channel_id" db:"agent_channel_id"`
	Communication   MemberCommunication `json:"communication" db:"communication"`
	HasReporting    bool                `json:"has_reporting" db:"has_reporting"`
	State           string              `json:"state" db:"state"`
	BridgedAt       *int64              `json:"bridged_at" db:"bridged_at"`
	LeavingAt       *int64              `json:"leaving_at" db:"leaving_at"`
	TimeoutAt       *int64              `json:"timeout_at" db:"timeout_at"`
	Duration        int                 `json:"duration" db:"duration"`
}

func NewWebSocketCallCenterEvent(ev *CallCenterEvent) (*WebSocketEvent, AppError) {
	var e *WebSocketEvent

	switch ev.Event {
	case WebsocketCCEventAgentStatus, WebsocketCCEventChannelStatus:
		e = NewWebSocketEvent(ev.Event)
	default:
		return nil, NewInternalError("event.cc.valid.event", fmt.Sprintf("unknown event \"%s\"", ev.Event))
	}
	e.UserId = ev.UserId
	e.SetData(ev.Body)

	return e, nil
}
