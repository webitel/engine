package model

type UserState struct {
	Id        int64       `json:"id,string,omitempty"`
	App       string      `json:"app,omitempty"`
	Status    string      `json:"status,omitempty"`
	Note      string      `json:"note,omitempty"`
	Open      int32       `json:"open,omitempty"`
	Closed    bool        `json:"closed,omitempty"`
	Contact   string      `json:"contact,omitempty"`
	Priority  int32       `json:"priority,omitempty"`
	Sequence  int64       `json:"sequence,omitempty"`
	Expires   int64       `json:"expires,omitempty"`
	Timestamp int64       `json:"timestamp,omitempty"`
	DomainId  int64       `json:"domain_id,omitempty"`
	Presence  interface{} `json:"presence,omitempty"`
}

func NewWebSocketUserStateEvent(state *UserState) *WebSocketEvent {
	e := NewWebSocketEvent(WEBSOCKET_EVENT_USER_STATE)
	e.Add("state", state)
	e.UserId = state.Id
	return e
}
