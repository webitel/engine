package model

type UserState struct {
	Id        string `json:"id"`
	AppId     string `json:"app_id"`
	Timestamp string `json:"timestamp"`
	Channels  string `json:"channels"`
	State     string `json:"state"`
}

func NewWebSocketUserStateEvent(state *UserState) *WebSocketEvent {
	e := NewWebSocketEvent(WEBSOCKET_EVENT_USER_STATE)
	e.Add("state", state)
	return e
}
