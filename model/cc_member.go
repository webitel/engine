package model

import (
	"encoding/json"
	"sort"
	"strings"
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
	HookCreated    *int32                `json:"-" db:"hook_created"`
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
	Attempts       *int
}

type MultiDeleteMembers struct {
	QueueId int64 `json:"queue_id" db:"queue_id"`
	SearchMemberRequest

	//Buckets   []int64   `json:"buckets" db:"buckets"` // deprecated
	//Causes    []string  `json:"causes" db:"causes"`   // deprecated
	Numbers   []string  `json:"numbers" db:"numbers"`
	Variables StringMap `json:"variables" db:"variables"`
}

type ResetMembers struct {
	QueueId   int64     `json:"queue_id" db:"queue_id"`
	Ids       []int64   `json:"ids" db:"ids"`
	Buckets   []int64   `json:"buckets" db:"buckets"`
	Causes    []string  `json:"causes" db:"causes"`
	AgentIds  []int32   `json:"agent_ids" db:"agent_ids"`
	Numbers   []string  `json:"numbers" db:"numbers"`
	Variables StringMap `json:"variables" db:"variables"`
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

	if len(p.Variables) != 0 {
		if m.Variables == nil {
			m.Variables = make(StringMap)
		}
		for k, v := range p.Variables {
			if v == "" {
				delete(m.Variables, k)
			} else {
				m.Variables[k] = v
			}
		}
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
		if *m.StopCause == "" {
			m.ResetAttempts()
		}
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

	if p.Attempts != nil {
		m.Attempts = *p.Attempts
	}
}

func (m *Member) ResetAttempts() {
	m.StopCause = nil
	m.StopAt = nil
	m.Attempts = 0
	for i := 0; i < len(m.Communications); i++ {
		m.Communications[i].StopAt = nil
		m.Communications[i].Attempts = 0
	}
}

type MemberView struct {
}

type SearchMemberRequest struct {
	ListRequest
	Ids         []int64
	QueueIds    []int32
	BucketIds   []int32
	Destination *string
	CreatedAt   *FilterBetween
	OfferingAt  *FilterBetween
	StopCauses  []string
	Priority    *FilterBetween
	Name        *string
	Attempts    *FilterBetween
	AgentIds    []int32
	QueueId     *int32
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
	Queue        *Lookup           `json:"queue" db:"queue"`
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
	AmdResult   *string             `json:"amd_result" db:"amd_result"`
	Attempts    *int32              `json:"attempts" db:"attempts"`
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
		"amd_result",
		"attempts",
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
	Queue           *Lookup           `json:"queue" db:"queue"`
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
	Attempts    *int32              `json:"attempts" db:"attempts"`
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
		"resource", "bucket", "list", "display", "destination", "result", "attempts",
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
	Attempts    *int32            `json:"attempts" db:"attempts"`
}

type SearchAttempts struct {
	ListRequest
	JoinedAt  *FilterBetween `json:"joined_at" db:"joined_at"`
	Ids       []int64        `json:"ids" db:"ids"`
	MemberIds []int64        `json:"member_ids" db:"member_ids"`
	//ResourceId  *int32        `json:"resource_id" db:"resource_id" `
	QueueIds  []int64 `json:"queue_ids" db:"queue_ids"`
	BucketIds []int64 `json:"bucket_ids" db:"bucket_ids"`
	//Destination *string       `json:"destination" db:"destination"`
	AgentIds   []int64        `json:"agent_ids" db:"agent_ids"`
	Result     []string       `json:"result" db:"result"`
	LeavingAt  *FilterBetween `json:"leaving_at" db:"leaving_at"`
	OfferingAt *FilterBetween `json:"offering_at" db:"offering_at"`
	Duration   *FilterBetween `json:"duration" db:"duration"`
}

type SearchOfflineQueueMembers struct {
	ListRequest
	AgentId int
}

type MembersAttempt struct {
	Member Lookup
	MemberAttempt
}

func (a *MemberAttempt) IsValid() AppError {
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
	StopAt         *int64  `json:"stop_at,omitempty"`
	Attempts       int     `json:"attempts"`
	LastCause      string  `json:"last_cause"`
	Resource       *Lookup `json:"resource"`
	Display        string  `json:"display"`
	Dtmf           *string `json:"dtmf"`
}

func (m *Member) ToJsonCommunications() string {
	// TODO: fix in lib
	sort.Slice(m.Communications[:], func(i, j int) bool {
		return m.Communications[i].Priority > m.Communications[j].Priority
	})
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

func (m *Member) IsValid(maxCommunication int) AppError {
	//FIXME

	if len(m.Communications) > maxCommunication {
		return NewBadRequestError("model.member.is_valid.communications.app_error", "name="+m.Name)
	}

	for _, v := range m.Communications {
		if v.Type.Id < 1 {
			return NewBadRequestError("model.member.is_valid.communications.type.app_error", "name="+m.Name)
		}
	}

	return nil
}

func (Member) DefaultOrder() string {
	return "-id"
}

func (m Member) AllowFields() []string {
	return m.DefaultFields()
}

func (Member) DefaultFields() []string {
	return []string{
		"id", "communications", "queue", "priority", "expire_at", "created_at", "variables", "name",
		"timezone", "bucket", "ready_at", "stop_cause", "stop_at", "last_hangup_at", "attempts", "agent", "skill", "reserved",
	}
}

func (c Member) EntityName() string {
	return "cc_member"
}

func MemberDeprecatedFields(f []string) []string {
	if f == nil {
		return nil
	}

	res := make([]string, 0, len(f))
	for _, v := range f {
		res = append(res, MemberDeprecatedField(v))
	}
	return res
}

func MemberDeprecatedField(s string) string {
	s = strings.Replace(s, "min_offering_at", "ready_at", -1)
	s = strings.Replace(s, "last_activity_at", "last_hangup_at", -1)
	return s
}
