package model

const (
	CALL_DIRECTION_INBOUND  = "inbound"
	CALL_DIRECTION_OUTBOUND = "outbound"

	CALL_EVENT_HEADER_ID               = "Unique-ID"
	CALL_EVENT_HEADER_DIRECTION        = "Presence-Call-Direction" //"Call-Direction"
	CALL_EVENT_HEADER_DOMAIN_ID        = "variable_sip_h_X-Webitel-Domain-Id"
	CALL_EVENT_HEADER_DOMAIN_NAME      = "variable_sip_h_X-Webitel-Domain"
	CALL_EVENT_HEADER_USER_ID          = "variable_sip_h_X-Webitel-User-Id"
	CALL_EVENT_HEADER_STATE            = "Channel-Call-State"
	CALL_EVENT_HEADER_TO_NUMBER        = "Caller-Callee-ID-Number"
	CALL_EVENT_HEADER_TO_NAME          = "Caller-Callee-ID-Name"
	CALL_EVENT_HEADER_FROM_NUMBER      = "Caller-Caller-ID-Number"
	CALL_EVENT_HEADER_FROM_NAME        = "Caller-Caller-ID-Name"
	CALL_EVENT_HEADER_FROM_DESTINATION = "Caller-Destination-Number"
)

const (
	EVENT_CHANNEL_CREATE  = "CHANNEL_CREATE"
	EVENT_CHANNEL_DESTROY = "CHANNEL_DESTROY"
)

type Call struct {
	Id         string  `json:"id"`
	DomainId   *string `json:"domain_id"`
	UserId     *string `json:"user_id"`
	State      string  `json:"state"`
	ToNumber   string  `json:"to_number"`
	ToName     string  `json:"to_name"`
	FromNumber string  `json:"from_number"`
	FromName   string  `json:"from_name"`
	Direction  string  `json:"direction"`
	ParentId   *string `json:"parent_id"`
	NodeName   string  `json:"node_name"`
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
	e.Add("id", call.Id)
	e.Add("domain_id", call.DomainId)
	e.Add("user_id", call.UserId)
	e.Add("state", call.State)
	e.Add("to_number", call.ToNumber)
	e.Add("to_name", call.ToName)
	e.Add("from_number", call.FromNumber)
	e.Add("from_name", call.FromName)
	e.Add("direction", call.Direction)
	e.Add("parent_id", call.ParentId)
	e.Add("node_name", call.NodeName)

	return e
}
