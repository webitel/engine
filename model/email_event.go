package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
)

// The Workspace events of email-service, addressed to one user by their routing key, as the
// events of calls and chats are; docs/email-events.md describes them.
const (
	EmailServiceExchange = "email.events"
	EmailContentType     = "application/json"
	EmailMaxEventSize    = 16 << 10
	EmailSchema          = 1

	EmailWorkspaceThreadChanged   = "email.workspace.thread.changed"
	EmailWorkspaceMessageCreated  = "email.workspace.message.created"
	EmailWorkspaceMessageDelivery = "email.workspace.message.delivery"
)

// ErrEmailEventInvalid marks an event no node delivers.
var ErrEmailEventInvalid = errors.New("email event: invalid")

func emailInvalid(format string, args ...any) error {
	return fmt.Errorf("%w: %s", ErrEmailEventInvalid, fmt.Sprintf(format, args...))
}

// EmailRoutingKey is the routing key of an event for one user.
func EmailRoutingKey(eventType string, domainId, userId int64) string {
	return eventType + ".v1." + strconv.FormatInt(domainId, 10) + "." + strconv.FormatInt(userId, 10)
}

// EmailBindingKey matches every Workspace event of one user, never the Flow event.
func EmailBindingKey(domainId, userId int64) string {
	return EmailRoutingKey("email.workspace.*.*", domainId, userId)
}

// EmailEvent is a Workspace event of email-service for one user.
type EmailEvent struct {
	EventId           string    `json:"event_id"`
	EventType         string    `json:"event_type"`
	SchemaVersion     int       `json:"schema_version"`
	OccurredAt        time.Time `json:"occurred_at"`
	DomainId          int64     `json:"domain_id"`
	UserId            int64     `json:"user_id"`
	ThreadId          int64     `json:"thread_id"`
	Access            string    `json:"access"`
	WorkspaceVersion  int64     `json:"workspace_version"`
	ProjectionVersion *int64    `json:"projection_version"`
	MessageId         *int64    `json:"message_id"`
	Changes           []string  `json:"changes"`
}

var (
	emailTypes   = map[string]bool{EmailWorkspaceThreadChanged: true, EmailWorkspaceMessageCreated: true, EmailWorkspaceMessageDelivery: true}
	emailAccess  = map[string]bool{"current": true, "history": true, "none": true}
	emailChanges = map[string]bool{"projection": true, "actions": true, "participants": true, "contact": true}
)

// ParseEmailEvent accepts an event of the queue of a node: JSON within the size limit, the
// envelope of webitel-emails docs/events.md, and its type, domain and user equal to those of
// the routing key. Unknown fields are ignored.
func ParseEmailEvent(routingKey, contentType string, body []byte) (*EmailEvent, error) {
	if contentType != EmailContentType {
		return nil, emailInvalid("content type %q", contentType)
	}
	if len(body) > EmailMaxEventSize {
		return nil, emailInvalid("%d bytes", len(body))
	}
	var e EmailEvent
	if err := json.Unmarshal(body, &e); err != nil {
		return nil, emailInvalid("json: %v", err)
	}
	if err := e.validate(); err != nil {
		return nil, err
	}
	if routingKey != EmailRoutingKey(e.EventType, e.DomainId, e.UserId) {
		return nil, emailInvalid("routing key %q of another type, domain or user", routingKey)
	}

	return &e, nil
}

func (e *EmailEvent) validate() error {
	switch {
	case !emailEventId(e.EventId):
		return emailInvalid("event_id")
	case !emailTypes[e.EventType] || e.SchemaVersion != EmailSchema:
		return emailInvalid("type %q schema %d", e.EventType, e.SchemaVersion)
	case e.OccurredAt.IsZero():
		return emailInvalid("occurred_at")
	case e.DomainId <= 0 || e.UserId <= 0 || e.ThreadId <= 0 || e.WorkspaceVersion <= 0:
		return emailInvalid("ids and version must be positive")
	case e.ProjectionVersion != nil && *e.ProjectionVersion <= 0:
		return emailInvalid("projection_version")
	case !emailAccess[e.Access]:
		return emailInvalid("access %q", e.Access)
	case (e.EventType == EmailWorkspaceThreadChanged) == (e.MessageId != nil) || (e.MessageId != nil && *e.MessageId <= 0):
		return emailInvalid("message_id")
	case e.Changes == nil:
		return emailInvalid("changes must be a list")
	case e.Access == "none" && (e.EventType != EmailWorkspaceThreadChanged || len(e.Changes) > 0):
		return emailInvalid("access none is a bare invalidation")
	case e.EventType != EmailWorkspaceThreadChanged && len(e.Changes) > 0:
		return emailInvalid("changes belong to thread events")
	}
	for _, c := range e.Changes {
		if !emailChanges[c] {
			return emailInvalid("change %q", c)
		}
	}

	return nil
}

// emailEventId accepts a canonical UUID other than nil, as email-service does.
func emailEventId(s string) bool {
	u, err := uuid.Parse(s)

	return err == nil && u != uuid.Nil && u.String() == s
}
