package model

import (
	"encoding/json"
	"fmt"
)

const (
	CALL_DIRECTION_INTERNAL = "internal"
	CALL_DIRECTION_INBOUND  = "inbound"
	CALL_DIRECTION_OUTBOUND = "outbound"
)

const (
	CALL_VARIABLE_DIRECTION         = "sip_h_X-Webitel-Direction"
	CALL_VARIABLE_DISPLAY_DIRECTION = "sip_h_X-Webitel-Display-Direction"
	CALL_VARIABLE_USER_ID           = "sip_h_X-Webitel-User-Id"
	CALL_VARIABLE_DOMAIN_ID         = "sip_h_X-Webitel-Domain-Id"
	CALL_VARIABLE_SOCK_ID           = "sip_h_X-Webitel-Sock-Id"
	CALL_VARIABLE_ID                = "sip_h_X-Webitel-Uuid"
	CALL_VARIABLE_USE_VIDEO         = "wbt_video"
	CALL_VARIABLE_USE_SCREEN        = "wbt_screen"
)
const (
	CALL_STRATEGY_DEFAULT = iota
	CALL_STRATEGY_FAILOVER
	CALL_STRATEGY_MULTIPLE
)

type CallRequestApplication struct {
	AppName string
	Args    string
}

const (
	EndpointTypeUser        = "user"
	EndpointTypeDestination = "destination"
)

type Endpoint struct {
	Type   string
	Number string
	Id     string
	Name   string
}

type EndpointRequest struct {
	AppId       *string
	UserId      *int64
	SchemaId    *int
	Destination *string
}

type CallRequest struct {
	Endpoints    []string
	Strategy     uint8
	Destination  string
	Variables    map[string]string
	Timeout      uint16
	CallerName   string
	CallerNumber string
	Dialplan     string
	Context      string
	Applications []*CallRequestApplication
}

type OutboundCallRequest struct {
	CreatedAt   int64            `json:"created_at"`
	CreatedById int64            `json:"created_by_id"`
	From        *EndpointRequest `json:"from"`
	To          *EndpointRequest `json:"to"`
	Destination string           `json:"destination"`
	Params      CallParameters   `json:"params"`
}

type UserCallRequest struct {
	Id    string  `json:"id"`
	AppId *string `json:"app_id"`
}

type HangupCall struct {
	UserCallRequest
	Cause *string `json:"cause"`
}

type DtmfCall struct {
	UserCallRequest
	Digit rune
}

type BlindTransferCall struct {
	UserCallRequest
	Destination string
}

type BridgeCall struct {
	FromId string `json:"from_id" db:"from_id"`
	ToId   string `json:"to_id" db:"to_id"`
	AppId  string `json:"app_id" db:"app_id"`
}

type EavesdropCall struct {
	UserCallRequest
	//Group       string //TODO https://freeswitch.org/confluence/display/FREESWITCH/mod_dptools%3A+eavesdrop
	Dtmf        bool
	ALeg        bool
	BLeg        bool
	WhisperALeg bool
	WhisperBLeg bool
}

type CallParameters struct {
	Timeout int
	Audio   bool
	Video   bool
	Screen  bool

	Record    bool
	Variables map[string]string
}

func (r *OutboundCallRequest) IsValid() *AppError {
	return nil
}

type CallInstance struct {
	Id        string  `json:"id" db:"id"`
	AppId     *string `json:"app_id" db:"app_id"`
	State     string  `json:"state" db:"state"`
	Timestamp int64   `json:"timestamp" db:"timestamp"`
}

type Call struct {
	CallInstance
	Direction   string  `json:"direction" db:"direction"`
	Destination string  `json:"destination" db:"destination"`
	ParentId    *string `json:"parent_id" db:"parent_id"`

	From Endpoint `json:"from" db:"from"`
	To   Endpoint `json:"to" db:"to"`
}

type CallFile struct {
	Id       int64  `json:"id"`
	Name     string `json:"name"`
	Size     int64  `json:"size"`
	MimeType string `json:"mime_type"`
}

type HistoryCall struct {
	Id          string            `json:"id" db:"id"`
	AppId       string            `json:"app_id" db:"app_id"`
	Type        string            `json:"type" db:"type"`
	ParentId    *string           `json:"parent_id" db:"parent_id"`
	User        *Lookup           `json:"user" db:"user"`
	Extension   *string           `json:"extension" db:"extension"`
	Gateway     *Lookup           `json:"gateway" db:"gateway"`
	Direction   string            `json:"direction" db:"direction"`
	Destination string            `json:"destination" db:"destination"`
	From        *Endpoint         `json:"from" db:"from"`
	To          *Endpoint         `json:"to" db:"to"`
	Variables   map[string]string `json:"variables" db:"variables"`

	CreatedAt  int64  `json:"created_at" db:"created_at"`
	AnsweredAt int64  `json:"answered_at" db:"answered_at"`
	BridgedAt  int64  `json:"bridged_at" db:"bridged_at"`
	HangupAt   int64  `json:"hangup_at" db:"hangup_at"`
	HangupBy   string `json:"hangup_by" db:"hangup_by"`
	Cause      string `json:"cause" db:"cause"`

	Duration int `json:"duration" db:"duration"`
	HoldSec  int `json:"hold_sec" db:"hold_sec"`
	WaitSec  int `json:"wait_sec" db:"wait_sec"`
	BillSec  int `json:"bill_sec" db:"bill_sec"`

	SipCode *int        `json:"sip_code" db:"sip_code"`
	Files   []*CallFile `json:"files" db:"files"`

	Queue  *Lookup `json:"queue" db:"queue"`
	Member *Lookup `json:"member" db:"member"`
	Team   *Lookup `json:"team" db:"team"`
	Agent  *Lookup `json:"agent" db:"agent"`

	JoinedAt         *int64      `json:"joined_at" db:"joined_at"`
	LeavingAt        *int64      `json:"leaving_at" db:"leaving_at"`
	ReportingAt      *int64      `json:"reporting_at" db:"reporting_at"`
	QueueBridgedAt   *int64      `json:"queue_bridged_at" db:"queue_bridged_at"`
	QueueWaitSec     *int        `json:"queue_wait_sec" db:"queue_wait_sec"`
	QueueDurationSec *int        `json:"queue_duration_sec" db:"queue_duration_sec"`
	ReportingSec     *int        `json:"reporting_sec" db:"reporting_sec"`
	Result           *string     `json:"result" db:"result"`
	Tags             StringArray `json:"tags" db:"tags"`
	Display          *string     `json:"display" db:"display"`
}

func (c HistoryCall) AllowFields() []string {
	return c.DefaultFields()
}

func (c HistoryCall) DefaultFields() []string {
	return []string{"id", "app_id", "parent_id", "user", "extension", "gateway", "direction", "destination", "from", "to", "variables",
		"created_at", "answered_at", "bridged_at", "hangup_at", "hangup_by", "cause", "duration", "hold_sec", "wait_sec", "bill_sec",
		"sip_code", "files", "queue", "member", "team", "agent", "joined_at", "leaving_at", "reporting_at", "queue_bridged_at",
		"queue_wait_sec", "queue_duration_sec", "result", "reporting_sec", "tags", "display",
	}
}

func (c HistoryCall) EntityName() string {
	return "cc_call_history_list"
}

func (c *HistoryCall) GetResult() string {
	if c.Result != nil {
		return *c.Result
	}

	return ""
}

type SearchCall struct {
	ListRequest
	UserId *int64 `json:"user_id"`
}

type SearchHistoryCall struct {
	ListRequest
	CreatedAt  FilterBetween
	Duration   *FilterBetween
	Number     *string
	ParentId   *string
	Cause      *string
	SkipParent bool
	ExistsFile bool
	UserIds    []int64
	QueueIds   []int64
	TeamIds    []int64
	AgentIds   []int64
	MemberIds  []int64
	GatewayIds []int64
}

type CallEvent struct {
	Id        string  `json:"id"`
	Event     string  `json:"event"`
	Timestamp float64 `json:"timestamp,string"`
	DomainId  string  `json:"domain_id"`
	UserId    string  `json:"user_id,omitempty"`
	AppId     string  `json:"app_id,omitempty"`
	//CCAppId   string      `json:"cc_app_id,omitempty"`
	Body CallPayload `json:"data,string,omitempty"`
}

type CallPayload map[string]interface{}

func (cp CallPayload) MarshalJSON() ([]byte, error) {
	return json.Marshal((*map[string]interface{})(&cp))
}

func (cp *CallPayload) UnmarshalText(b []byte) error {
	return json.Unmarshal(b, (*map[string]interface{})(cp))
}

func (cr *CallRequest) AddUserVariable(name, value string) {
	cr.AddVariable(fmt.Sprintf("usr_%s", name), value)
}

func (cr *CallRequest) AddVariable(name, value string) {
	if cr.Variables == nil {
		cr.Variables = make(map[string]string)
	}
	cr.Variables[name] = value
}

func NewWebSocketCallEvent(call *CallEvent) *WebSocketEvent {
	e := NewWebSocketEvent(WEBSOCKET_EVENT_CALL)
	e.Add("call", call)

	return e
}
