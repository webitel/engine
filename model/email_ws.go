package model

import "strconv"

const (
	WebsocketEmailEvent = "email"

	EmailResyncSubscriptionLost = "subscription_lost"
	EmailResyncOverflow         = "overflow"
)

var emailWsTypes = map[string]string{
	EmailWorkspaceThreadChanged:   "thread_changed",
	EmailWorkspaceMessageCreated:  "message_created",
	EmailWorkspaceMessageDelivery: "message_delivery",
}

// WebSocketEvent is the event "email" of docs/email-events.md: only the fields of its type,
// ids and versions as decimal strings.
func (e *EmailEvent) WebSocketEvent() *WebSocketEvent {
	d := map[string]interface{}{
		"type":              emailWsTypes[e.EventType],
		"schema_version":    EmailSchema,
		"event_id":          e.EventId,
		"thread_id":         strconv.FormatInt(e.ThreadId, 10),
		"access":            e.Access,
		"workspace_version": strconv.FormatInt(e.WorkspaceVersion, 10),
	}
	if e.EventType == EmailWorkspaceThreadChanged {
		d["changes"] = e.Changes
	} else {
		d["message_id"] = strconv.FormatInt(*e.MessageId, 10)
	}

	return &WebSocketEvent{Event: WebsocketEmailEvent, Data: d}
}

// NewEmailResyncEvent tells the client to read everything in view again.
func NewEmailResyncEvent(reason string) *WebSocketEvent {
	return &WebSocketEvent{Event: WebsocketEmailEvent, Data: map[string]interface{}{
		"type": "resync_required", "schema_version": EmailSchema, "reason": reason,
	}}
}
