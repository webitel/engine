package model

import (
	"encoding/json"
	"fmt"
)

const (
	CALL_DIRECTION_INTERNAL = "internal"
	CALL_DIRECTION_INBOUND  = "inbound"
	CALL_DIRECTION_OUTBOUND = "outbound"

	CALL_STATE_RINGING = "ringing"
	CALL_STATE_ACTIVE  = "active"
	CALL_STATE_HOLD    = "hold"
	CALL_STATE_HANGUP  = "hangup"

	CALL_EVENT_HEADER_NODE_NAME              = "FreeSWITCH-Switchname"
	CALL_EVENT_HEADER_ID                     = "Unique-ID"
	CALL_EVENT_HEADER_DIRECTION              = "Presence-Call-Direction" //"Call-Direction"
	CALL_EVENT_HEADER_CALL_DIRECTION         = "variable_sip_h_X-Webitel-Direction"
	CALL_EVENT_HEADER_CALL_DISPLAY_DIRECTION = "variable_sip_h_X-Webitel-Display-Direction"
	CALL_EVENT_HEADER_DOMAIN_ID              = "variable_sip_h_X-Webitel-Domain-Id"
	CALL_EVENT_HEADER_DOMAIN_NAME            = "variable_sip_h_X-Webitel-Domain"
	CALL_EVENT_HEADER_USER_ID                = "variable_sip_h_X-Webitel-User-Id"
	CALL_EVENT_HEADER_DESTINATION            = "variable_sip_h_X-Webitel-Destination"
	CALL_EVENT_HEADER_STATE                  = "Answer-State"
	CALL_EVENT_HEADER_STATE_NUMBER           = "Channel-State-Number"
	CALL_EVENT_HEADER_TO_NUMBER              = "Caller-Callee-ID-Number"
	CALL_EVENT_HEADER_TO_NAME                = "Caller-Callee-ID-Name"
	CALL_EVENT_HEADER_FROM_NUMBER            = "Caller-Caller-ID-Number"
	CALL_EVENT_HEADER_FROM_NAME              = "Caller-Caller-ID-Name"
	CALL_EVENT_HEADER_FROM_DESTINATION       = "Caller-Destination-Number"
	CALL_EVENT_HEADER_HANGUP_CAUSE           = "variable_hangup_cause"
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
	EVENT_CHANNEL_CREATE          = "CHANNEL_CREATE"
	EVENT_CHANNEL_ANSWER          = "CHANNEL_ANSWER"
	EVENT_CHANNEL_DESTROY         = "CHANNEL_DESTROY"
	EVENT_CHANNEL_UNHOLD          = "CHANNEL_UNHOLD"
	EVENT_CHANNEL_BRIDGE          = "CHANNEL_BRIDGE"
	EVENT_CHANNEL_HOLD            = "CHANNEL_HOLD"
	EVENT_CHANNEL_HANGUP_COMPLETE = "CHANNEL_HANGUP_COMPLETE"
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

type Call struct {
	Action        string `json:"action"`
	Id            string `json:"id"`
	DomainId      string `json:"domain_id"`
	UserId        string `json:"user_id,omitempty"`
	Application   string `json:"application,omitempty"`
	ToNumber      string `json:"to_number,omitempty"`
	ToName        string `json:"to_name,omitempty"`
	FromNumber    string `json:"from_number,omitempty"`
	FromName      string `json:"from_name,omitempty"`
	Destination   string `json:"destination,omitempty"`
	Direction     string `json:"direction,omitempty"`
	ParentId      string `json:"parent_id,omitempty"`
	OwnerId       string `json:"owner_id,omitempty"`
	NodeName      string `json:"node_name,omitempty"`
	HangupCause   string `json:"cause,omitempty"`
	VideoFlow     string `json:"video_flow,omitempty"`
	VideoRequest  bool   `json:"video_request,string,omitempty"`
	ScreenRequest bool   `json:"screen_request,string,omitempty"`
	Digit         string `json:"digit,omitempty"`

	Debug   map[string]interface{} `json:"debug,omitempty"`
	Payload *CallPayload           `json:"payload,string,omitempty"`
}

type CallPayload map[string]interface{}

func (cp CallPayload) MarshalJSON() ([]byte, error) {
	return json.Marshal((*map[string]interface{})(&cp))
}

func (cp *CallPayload) UnmarshalText(b []byte) error {
	return json.Unmarshal(b, (*map[string]interface{})(cp))
}

func (cr *CallRequest) AddUserVariable(name, value string) {
	cr.AddVariable(fmt.Sprintf("wbt_%s", name), value)
}
func (cr *CallRequest) AddVariable(name, value string) {
	if cr.Variables == nil {
		cr.Variables = make(map[string]string)
	}
	cr.Variables[name] = value
}

type CallEvent interface {
	Name() string
	Id() string
	GetVariable(name string) (string, bool)
	GetIntVariable(name string) (int, bool)
	ToMapStringInterface() map[string]interface{}
}

func NewWebSocketCallEvent(call *Call) *WebSocketEvent {
	e := NewWebSocketEvent(WEBSOCKET_EVENT_CALL)
	e.Add("call", call)

	if call.Debug != nil {
		e.Add("debug", call.Debug)
	}

	return e
}
