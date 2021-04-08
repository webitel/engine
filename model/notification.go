package model

import (
	"encoding/json"
)

type Notification struct {
	Id        int64  `json:"id" db:"id"`
	DomainId  int64  `json:"-" db:"domain_id"`
	Action    string `json:"action" db:"action"`
	Timeout   *int64 `json:"timeout" db:"timeout"`
	CreatedAt int64  `json:"created_at" db:"created_at"`
	CreatedBy *int64 `json:"created_by" db:"created_by"`

	AcceptedAt *int64 `json:"accepted_at" db:"accepted_at"`
	AcceptedBy *int64 `json:"accepted_by" db:"accepted_by"`

	ClosedAt *int64 `json:"closed_at" db:"closed_at"`

	ForUsers Int64Array `json:"for_users" db:"for_users"`

	Description string `json:"description" db:"description"`
}

func (n *Notification) ToJson() string {
	b, _ := json.Marshal(n)
	return string(b)
}

func NewWebSocketNotificationEvent(n *Notification) *WebSocketEvent {
	e := NewWebSocketEvent(WebsocketNotificationEvent)
	e.Add("notification", n)
	return e
}
