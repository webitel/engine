package model

import (
	"encoding/json"
	"testing"
)

func TestEmailWebSocketEvent(t *testing.T) {
	big := int64(9007199254740993) // above 2^53: a JSON number would round it
	pv, msg := int64(905), int64(31)
	for name, c := range map[string]struct {
		e    EmailEvent
		want string
	}{
		"thread_changed": {EmailEvent{EventType: EmailWorkspaceThreadChanged, EventId: "e", ThreadId: big, Access: "current",
			WorkspaceVersion: 17, ProjectionVersion: &pv, Changes: []string{"projection"}},
			`{"access":"current","changes":["projection"],"event_id":"e","schema_version":1,"thread_id":"9007199254740993","type":"thread_changed","workspace_version":"17"}`},
		"message_created": {EmailEvent{EventType: EmailWorkspaceMessageCreated, EventId: "e", ThreadId: 1, Access: "current",
			WorkspaceVersion: 17, MessageId: &msg, Changes: []string{}},
			`{"access":"current","event_id":"e","message_id":"31","schema_version":1,"thread_id":"1","type":"message_created","workspace_version":"17"}`},
	} {
		ev := c.e.WebSocketEvent()
		got, _ := json.Marshal(ev.Data)
		if ev.Event != "email" || string(got) != c.want {
			t.Errorf("%s: %s %s", name, ev.Event, got)
		}
	}

	got, _ := json.Marshal(NewEmailResyncEvent(EmailResyncOverflow).Data)
	if string(got) != `{"reason":"overflow","schema_version":1,"type":"resync_required"}` {
		t.Errorf("resync: %s", got)
	}
}
