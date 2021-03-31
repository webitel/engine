package model

import (
	"encoding/json"
	"time"
)

type Member struct {
	Id             int64                 `json:"id" db:"id"`
	Queue          Lookup                `json:"queue" db:"queue"`
	CreatedAt      time.Time             `json:"created_at" db:"created_at"`
	QueueId        int64                 `json:"queue_id" db:"queue_id"` //FIXME delete attr
	Priority       int                   `json:"priority" db:"priority"`
	ExpireAt       *time.Time            `json:"expire_at" db:"expire_at"`
	MinOfferingAt  *time.Time            `json:"min_offering_at" db:"ready_at"`
	Name           string                `json:"name" db:"name"`
	Variables      StringMap             `json:"variables" db:"variables"`
	LastActivityAt int64                 `json:"last_hangup_at" db:"last_hangup_at"`
	Attempts       int                   `json:"attempts" db:"attempts"`
	Timezone       Lookup                `json:"timezone" db:"timezone"`
	Bucket         *Lookup               `json:"bucket" db:"bucket"`
	Communications []MemberCommunication `json:"communications" db:"communications"`
	StopAt         *time.Time            `json:"stop_at" db:"stop_at"`
	StopCause      *string               `json:"stop_cause" db:"stop_cause"`
	Reserved       bool                  `json:"reserved" db:"reserved"`
	Agent          *Lookup               `json:"agent" db:"agent"`
	Skill          *Lookup               `json:"skill" db:"skill"`
}

type MemberPatch struct {
	Priority       *int                  `json:"priority" db:"priority"`
	ExpireAt       *time.Time            `json:"expire_at" db:"expire_at"`
	MinOfferingAt  *time.Time            `json:"min_offering_at" db:"ready_at"`
	Name           *string               `json:"name" db:"name"`
	Variables      StringMap             `json:"variables" db:"variables"`
	Timezone       *Lookup               `json:"timezone" db:"timezone"`
	Bucket         *Lookup               `json:"bucket" db:"bucket"`
	Communications []MemberCommunication `json:"communications" db:"communications"`
	StopCause      *string               `json:"stop_cause" db:"stop_cause"`
	Agent          *Lookup               `json:"agent" db:"agent"`
	Skill          *Lookup               `json:"skill" db:"skill"`
}

func (m *Member) Patch(p *MemberPatch) {
	if p.Priority != nil {
		m.Priority = *p.Priority
	}

	if p.ExpireAt != nil {
		m.ExpireAt = p.ExpireAt
	}

	if p.MinOfferingAt != nil {
		m.MinOfferingAt = p.MinOfferingAt
	}

	if p.Name != nil {
		m.Name = *p.Name
	}

	if p.Variables != nil {
		m.Variables = p.Variables
	}

	if p.Timezone != nil {
		m.Timezone = *p.Timezone
	}

	if p.Bucket != nil {
		m.Bucket = p.Bucket
	}

	if p.Communications != nil {
		m.Communications = p.Communications
	}

	if p.StopCause != nil {
		m.StopCause = p.StopCause
	}

	if p.Agent != nil {
		//todo
		if p.Agent.Id == 0 {
			m.Agent = nil
		} else {
			m.Agent = p.Agent
		}
	}

	if p.Skill != nil {
		//todo
		if p.Skill.Id == 0 {
			m.Skill = nil
		} else {
			m.Skill = p.Skill
		}
	}
}

type MemberView struct {
}

type SearchMemberRequest struct {
	ListRequest
	Id          *int64
	QueueId     *int64
	Destination *string
	BucketId    *int32
}

type OfflineMember struct {
	Id             int64                 `json:"id" db:"id"`
	Name           string                `json:"name" db:"name"`
	Communications []MemberCommunication `json:"communications" db:"communications"`
	Queue          Lookup                `json:"queue" db:"queue"`
	ExpireAt       *int64                `json:"expire_at" db:"expire_at"`
	CreatedAt      int64                 `json:"created_at" db:"created_at"`
	Variables      StringMap             `json:"variables" db:"variables"`
}

/*
|id|state|last_state_change|timeout|channel|queue|member|variables|agent|position|resource|bucket|list|display|destination|joined_at|offering_at|bridged_at|reporting_at|leaving_at|active|
*/

type AttemptHistory struct {
	Id           int64             `json:"id" db:"id"`
	JoinedAt     *time.Time        `json:"joined_at" db:"joined_at"`
	OfferingAt   *time.Time        `json:"offering_at" db:"offering_at"`
	BridgedAt    *time.Time        `json:"bridged_at" db:"bridged_at"`
	ReportingAt  *time.Time        `json:"reporting_at" db:"reporting_at"`
	LeavingAt    *time.Time        `json:"leaving_at" db:"leaving_at"`
	Channel      string            `json:"channel" db:"channel"`
	Queue        Lookup            `json:"queue" db:"queue"`
	Member       *Lookup           `json:"member" db:"member"`
	MemberCallId *string           `json:"member_call_id" db:"member_call_id"`
	Variables    map[string]string `json:"variables" db:"variables"`

	Agent       *Lookup             `json:"agent" db:"agent"`
	AgentCallId *string             `json:"agent_call_id" db:"agent_call_id"`
	Position    int                 `json:"position" db:"position"`
	Resource    *Lookup             `json:"resource" db:"resource"`
	Bucket      *Lookup             `json:"bucket" db:"bucket"`
	List        *Lookup             `json:"list" db:"list"`
	Display     string              `json:"display" db:"display"`
	Destination MemberCommunication `json:"destination" db:"destination"`
	Active      bool                `json:"active" db:"active"` // FIXME delete me
	Result      string              `json:"result" db:"result"`
}

func (c AttemptHistory) DefaultOrder() string {
	return "-joined_at"
}

func (c AttemptHistory) AllowFields() []string {
	return c.DefaultFields()
}

func (c AttemptHistory) DefaultFields() []string {
	return []string{
		"id",
		"channel",
		"queue",
		"member",
		"variables",
		"agent",
		"position",
		"resource",
		"bucket",
		"list",
		"display",
		"destination",
		"joined_at",
		"offering_at",
		"bridged_at",
		"reporting_at",
		"leaving_at",
		"result",
	}
}

func (c AttemptHistory) EntityName() string {
	return "cc_member_view_attempt_history"
}

type Attempt struct {
	Id              int64             `json:"id" db:"id"`
	State           string            `json:"state" db:"state"`
	LastStateChange int64             `json:"last_state_change" db:"last_state_change"`
	JoinedAt        int64             `json:"joined_at" db:"joined_at"`
	OfferingAt      int64             `json:"offering_at" db:"offering_at"`
	BridgedAt       int64             `json:"bridged_at" db:"bridged_at"`
	ReportingAt     int64             `json:"reporting_at" db:"reporting_at"`
	Timeout         int64             `json:"timeout" db:"timeout"`
	LeavingAt       int64             `json:"leaving_at" db:"leaving_at"`
	Channel         string            `json:"channel" db:"channel"`
	Queue           Lookup            `json:"queue" db:"queue"`
	Member          *Lookup           `json:"member" db:"member"`
	MemberCallId    *string           `json:"member_call_id" db:"member_call_id"`
	Variables       map[string]string `json:"variables" db:"variables"`

	Agent       *Lookup             `json:"agent" db:"agent"`
	AgentCallId *string             `json:"agent_call_id" db:"agent_call_id"`
	Position    int                 `json:"position" db:"position"`
	Resource    *Lookup             `json:"resource" db:"resource"`
	Bucket      *Lookup             `json:"bucket" db:"bucket"`
	List        *Lookup             `json:"list" db:"list"`
	Display     string              `json:"display" db:"display"`
	Destination MemberCommunication `json:"destination" db:"destination"`
	Result      *string             `json:"result" db:"result"`
}

func (c Attempt) DefaultOrder() string {
	return "-joined_at"
}

func (c Attempt) AllowFields() []string {
	return c.DefaultFields()
}

func (c Attempt) DefaultFields() []string {
	return []string{
		"id", "state", "last_state_change", "joined_at", "offering_at", "bridged_at", "reporting_at", "leaving_at",
		"timeout", "channel", "queue", "member", "member_call_id", "variables", "agent", "agent_call_id", "position",
		"resource", "bucket", "list", "display", "destination", "result",
	}
}

func (c Attempt) EntityName() string {
	return "cc_member_view_attempt"
}

type MemberAttempt struct {
	Id          int64             `json:"id" db:"id"`
	CreatedAt   int64             `json:"created_at" db:"created_at"`
	Destination string            `json:"destination" db:"destination"`
	Weight      int               `json:"weight" db:"weight"`
	OriginateAt int64             `json:"originate_at" db:"originate_at"`
	AnsweredAt  int64             `json:"answered_at" db:"answered_at"`
	BridgedAt   int64             `json:"bridged_at" db:"bridged_at"`
	HangupAt    int64             `json:"hangup_at" db:"hangup_at"`
	Resource    Lookup            `json:"resource" db:"resource"`
	LegAId      *string           `json:"leg_a_id" db:"leg_a_id"`
	LegBId      *string           `json:"leg_b_id" db:"leg_b_id"`
	Node        *string           `json:"node" json:"node"`
	Result      *string           `json:"result" db:"result"`
	Agent       *Lookup           `json:"agent" db:"agent"`
	Bucket      *Lookup           `json:"bucket" db:"bucket"`
	Logs        []byte            `json:"logs" db:"logs"`
	Active      bool              `json:"active" db:"active"`
	Variables   map[string]string `json:"variables" db:"variables"`
}

type SearchAttempts struct {
	ListRequest
	JoinedAt  FilterBetween `json:"joined_at" db:"joined_at"`
	Ids       []int64       `json:"ids" db:"ids"`
	MemberIds []int64       `json:"member_ids" db:"member_ids"`
	//ResourceId  *int32        `json:"resource_id" db:"resource_id" `
	QueueIds  []int64 `json:"queue_ids" db:"queue_ids"`
	BucketIds []int64 `json:"bucket_ids" db:"bucket_ids"`
	//Destination *string       `json:"destination" db:"destination"`
	AgentIds []int64 `json:"agent_ids" db:"agent_ids"`
	Result   *string `json:"result" db:"result"`
}

type SearchOfflineQueueMembers struct {
	ListRequest
	AgentId int
}

type MembersAttempt struct {
	Member Lookup
	MemberAttempt
}

func (a *MemberAttempt) IsValid() *AppError {
	//FIXME
	return nil
}

type MemberCommunication struct {
	Id             int64   `json:"id"`
	Destination    string  `json:"destination"`
	Type           Lookup  `json:"type"`
	Priority       int     `json:"priority"`
	State          int     `json:"state"`
	Description    string  `json:"description"`
	LastActivityAt int64   `json:"last_activity_at"`
	Attempts       int     `json:"attempts"`
	LastCause      string  `json:"last_cause"`
	Resource       *Lookup `json:"resource"`
	Display        string  `json:"display"`
}

func (m *Member) ToJsonCommunications() string {
	data, _ := json.Marshal(m.Communications)
	return string(data)
}

func (m *Member) GetBucketId() *int64 {
	if m.Bucket != nil {
		return NewInt64(int64(m.Bucket.Id))
	}
	return nil
}

func (m *Member) GetAgentId() *int {
	if m.Agent != nil {
		return &m.Agent.Id
	}

	return nil
}

func (m *Member) IsValid() *AppError {
	//FIXME
	return nil
}
