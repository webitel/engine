package model

type UserState struct {
	AppId     string `json:"app_id"`
	Timestamp string `json:"timestamp"`
	UserId    string `json:"user_id"`
	Channels  string `json:"channels"`
	State     string `json:"state"`
}

func NewWebSocketUserStateEvent(state *UserState) *WebSocketEvent {
	e := NewWebSocketEvent(WEBSOCKET_EVENT_USER_STATE)
	e.Add("state", state)
	return e
}
