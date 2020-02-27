package model

//{"app_id":"6ab22cfb-c873-4dda-848f-a40606faa487","channels":"1","domain_id":"1","event":"user_state","state":"REGED","timestamp":"1582808603935","user_id":"3"}
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
