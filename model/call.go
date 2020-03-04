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
	Id     int
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
	To          EndpointRequest  `json:"to"`
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
	Id    string  `json:"id" db:"id"`
	AppId *string `json:"app_id" db:"app_id"`
	State string  `json:"state" db:"state"`
}

type Call struct {
	CallInstance
	CreatedAt int64   `json:"created_at" db:"created_at"`
	User      *Lookup `json:"created_by" db:"created_by"`

	Timestamp int64   `json:"timestamp" db:"timestamp"`
	ParentId  *string `json:"parent_id" db:"parent_id"`

	Direction string    `json:"direction" db:"direction"`
	From      Endpoint  `json:"from" db:"from"`
	To        *Endpoint `json:"to" db:"to"`
}

type SearchCall struct {
	ListRequest
}

type CallEvent struct {
	Id        string      `json:"id"`
	Event     string      `json:"event"`
	Timestamp float64     `json:"timestamp,string"`
	DomainId  string      `json:"domain_id"`
	UserId    string      `json:"user_id,omitempty"`
	AppId     string      `json:"app_id,omitempty"`
	CCAppId   string      `json:"cc_app_id,omitempty"`
	Body      CallPayload `json:"data,string,omitempty"`
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
