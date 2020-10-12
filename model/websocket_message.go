package model

import (
	"encoding/json"
	"fmt"
	"io"
)

const (
	WEBSOCKET_EVENT_HELLO = "hello"

	WEBSOCKET_EVENT_RESPONSE   = "response"
	WEBSOCKET_EVENT_CALL       = "call"
	WEBSOCKET_EVENT_CHAT       = "chat"
	WEBSOCKET_EVENT_USER_STATE = "user_state"

	WEBSOCKET_AUTHENTICATION_CHALLENGE = "authentication_challenge"
)

const (
	WebsocketCCEventAgentStatus   = "agent_status"
	WebsocketCCEventChannelStatus = "channel"
)

type WebSocketMessage interface {
	ToJson() string
	IsValid() bool
	EventType() string
}

type precomputedWebSocketEventJSON struct {
	Event json.RawMessage
	Data  json.RawMessage
}

type WebSocketEvent struct {
	Event    string                 `json:"event"`
	Data     map[string]interface{} `json:"data"`
	Sequence int64                  `json:"seq"`
	UserId   int64                  `json:"user_id,omitempty"`

	precomputedJSON *precomputedWebSocketEventJSON
}

// PrecomputeJSON precomputes and stores the serialized JSON for all fields other than Sequence.
// This makes ToJson much more efficient when sending the same event to multiple connections.
func (m *WebSocketEvent) PrecomputeJSON() {
	event, _ := json.Marshal(m.Event)
	data, _ := json.Marshal(m.Data)
	m.precomputedJSON = &precomputedWebSocketEventJSON{
		Event: json.RawMessage(event),
		Data:  json.RawMessage(data),
	}
}

func (m *WebSocketEvent) Add(key string, value interface{}) {
	m.Data[key] = value
}

func (m *WebSocketEvent) SetData(data map[string]interface{}) {
	m.Data = data
}

func (m *WebSocketEvent) GetString(key string) string {
	if v, ok := m.Data[key]; ok {
		return v.(string)
	}
	return ""
}

func NewWebSocketEvent(event string) *WebSocketEvent {
	return &WebSocketEvent{Event: event, Data: make(map[string]interface{})}
}

func (o *WebSocketEvent) IsValid() bool {
	return o.Event != ""
}

func (o *WebSocketEvent) EventType() string {
	return o.Event
}

func (o *WebSocketEvent) ToJson() string {
	if o.precomputedJSON != nil {
		return fmt.Sprintf(`{"event": %s, "data": %s, "seq": %d}`, o.precomputedJSON.Event, o.precomputedJSON.Data, o.Sequence)
	}
	b, _ := json.Marshal(o)
	return string(b)
}

func WebSocketEventFromJson(data io.Reader) *WebSocketEvent {
	var o *WebSocketEvent
	json.NewDecoder(data).Decode(&o)
	return o
}

type WebSocketResponse struct {
	Status   string                 `json:"status"`
	SeqReply int64                  `json:"seq_reply,omitempty"`
	Data     map[string]interface{} `json:"data,omitempty"`
	Error    error                  `json:"error,omitempty"`
}

func (m *WebSocketResponse) Add(key string, value interface{}) {
	m.Data[key] = value
}

func NewWebSocketResponse(status string, seqReply int64, data map[string]interface{}) *WebSocketResponse {
	return &WebSocketResponse{Status: status, SeqReply: seqReply, Data: data}
}

func NewWebSocketError(seqReply int64, err error) *WebSocketResponse {
	return &WebSocketResponse{Status: STATUS_FAIL, SeqReply: seqReply, Error: err}
}

func (o *WebSocketResponse) IsValid() bool {
	return o.Status != ""
}

func (o *WebSocketResponse) EventType() string {
	return WEBSOCKET_EVENT_RESPONSE
}

func (o *WebSocketResponse) ToJson() string {
	b, _ := json.Marshal(o)
	return string(b)
}

func WebSocketResponseFromJson(data io.Reader) *WebSocketResponse {
	var o *WebSocketResponse
	json.NewDecoder(data).Decode(&o)
	return o
}
