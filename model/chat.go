package model

type ChatEvent struct {
	Event    string `json:"event"`
	UserId   int64  `json:"user_id"`
	DomainId int64  `json:"domain_id"`
	Data     map[string]interface{}
}

func NewWebSocketChatEvent(event *ChatEvent) *WebSocketEvent {
	e := NewWebSocketEvent(WEBSOCKET_EVENT_CHAT)
	e.Add("data", event.Data)
	e.Add("action", event.Event)

	return e
}
