package model

import (
	"errors"
	"strings"
	"testing"
)

const testEventId = "5b247aa1-9433-53ce-b37d-346cec14bff5"

func threadJSON(access, changes string) string {
	return `{"event_id":"` + testEventId + `","event_type":"email.workspace.thread.changed","schema_version":1,
"occurred_at":"2026-11-06T09:00:00Z","domain_id":1,"user_id":42,"thread_id":1024,"access":"` + access + `",
"workspace_version":17,"projection_version":905,"message_id":null,"changes":` + changes + `,"future":1}`
}

func TestParseEmailEvent(t *testing.T) {
	key := "email.workspace.thread.changed.v1.1.42"
	if EmailBindingKey(1, 42) != "email.workspace.*.*.v1.1.42" {
		t.Fatalf("binding %s", EmailBindingKey(1, 42))
	}
	e, err := ParseEmailEvent(key, EmailContentType, []byte(threadJSON("current", `["projection"]`)))
	if err != nil || e.UserId != 42 || e.WorkspaceVersion != 17 || e.Changes[0] != "projection" {
		t.Fatalf("%+v, %v", e, err)
	}

	valid := threadJSON("none", `[]`)
	for name, c := range map[string]struct{ key, contentType, body string }{
		"other user in the key":  {"email.workspace.thread.changed.v1.1.43", EmailContentType, valid},
		"other type in the key":  {"email.workspace.message.created.v1.1.42", EmailContentType, valid},
		"content type":           {key, "text/json", valid},
		"oversized":              {key, EmailContentType, strings.Replace(valid, `"future":1`, `"x":"`+strings.Repeat("x", EmailMaxEventSize)+`"`, 1)},
		"nil event id":           {key, EmailContentType, strings.Replace(valid, testEventId, "00000000-0000-0000-0000-000000000000", 1)},
		"upper case event id":    {key, EmailContentType, strings.Replace(valid, testEventId, strings.ToUpper(testEventId), 1)},
		"none with changes":      {key, EmailContentType, threadJSON("none", `["actions"]`)},
		"changes are not a list": {key, EmailContentType, threadJSON("current", `null`)},
		"schema":                 {key, EmailContentType, strings.Replace(valid, `"schema_version":1`, `"schema_version":2`, 1)},
	} {
		if _, err := ParseEmailEvent(c.key, c.contentType, []byte(c.body)); !errors.Is(err, ErrEmailEventInvalid) {
			t.Errorf("%s: err = %v", name, err)
		}
	}
}
