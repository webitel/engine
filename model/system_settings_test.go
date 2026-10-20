package model_test

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/webitel/engine/model"
)

func TestSystemSetting_IsValid_RecordAllCalls(t *testing.T) {
	cases := []struct {
		name    string
		value   json.RawMessage
		wantErr bool
	}{
		{name: "true", value: json.RawMessage(`true`), wantErr: false},
		{name: "false", value: json.RawMessage(`false`), wantErr: false},
		{name: "string", value: json.RawMessage(`"x"`), wantErr: true},
		{name: "number", value: json.RawMessage(`5`), wantErr: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := &model.SystemSetting{Name: model.SysNameRecordAllCalls, Value: tc.value}
			err := s.IsValid()

			if tc.wantErr && err == nil {
				t.Fatalf("expected error for value %s, got nil", tc.value)
			}

			if !tc.wantErr && err != nil {
				t.Fatalf("expected no error for value %s, got %v", tc.value, err)
			}
		})
	}
}

func TestSystemSetting_IsValid_AutoGenerateExtensionNumber(t *testing.T) {
	cases := []struct {
		name    string
		value   json.RawMessage
		wantErr bool
	}{
		{name: "positive", value: json.RawMessage(`205`), wantErr: false},
		{name: "one", value: json.RawMessage(`1`), wantErr: false},
		{name: "zero", value: json.RawMessage(`0`), wantErr: true},
		{name: "negative", value: json.RawMessage(`-5`), wantErr: true},
		{name: "string", value: json.RawMessage(`"205"`), wantErr: true},
		{name: "bool", value: json.RawMessage(`true`), wantErr: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := &model.SystemSetting{Name: model.SysNameAutoGenerateExtensionNumber, Value: tc.value}
			err := s.IsValid()

			if tc.wantErr && err == nil {
				t.Fatalf("expected error for value %s, got nil", tc.value)
			}

			if !tc.wantErr && err != nil {
				t.Fatalf("expected no error for value %s, got %v", tc.value, err)
			}
		})
	}
}

func TestSystemSetting_Sanitized(t *testing.T) {
	cases := []struct {
		name      string
		setting   string
		wantValue bool
	}{
		{"regular keeps value", model.SysNameOmnichannel, true},
		{"default_password drops value", model.SysNameDefaultPassword, false},
		{"chat_ai_connection drops value", model.SysNameChatAiConnection, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			in := &model.SystemSetting{Id: 3, Name: tc.setting, Value: json.RawMessage(`"x"`)}

			out := in.Sanitized()

			if out.Id != in.Id || out.Name != in.Name {
				t.Fatalf("expected id/name preserved, got %+v", out)
			}

			if tc.wantValue != (out.Value != nil) {
				t.Fatalf("value present = %v, want %v", out.Value != nil, tc.wantValue)
			}

			if string(in.Value) != `"x"` {
				t.Fatalf("original setting mutated: %s", in.Value)
			}
		})
	}
}

func TestSystemSetting_Sanitized_JSONHasNoSecret(t *testing.T) {
	s := &model.SystemSetting{Name: model.SysNameDefaultPassword, Value: json.RawMessage(`"s3cret"`)}

	b, err := json.Marshal(s.Sanitized())
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	if strings.Contains(string(b), "s3cret") {
		t.Fatalf("sanitized body leaks the value: %s", b)
	}
}

func TestNewSystemSettingsNotification(t *testing.T) {
	cases := []struct {
		name      string
		setting   *model.SystemSetting
		wantNil   bool
		wantValue bool
	}{
		{"live setting carries value", &model.SystemSetting{Name: model.SysNameCallEndSoundNotification, Value: json.RawMessage(`true`)}, false, true},
		{"live setting without value is name-only", &model.SystemSetting{Name: model.SysNameDefaultWorkspaceTab}, false, false},
		{"screenshot_interval is pushed", &model.SystemSetting{Name: model.SysNameScreenshotInterval, Value: json.RawMessage(`30`)}, false, true},
		{"login_options is not pushed", &model.SystemSetting{Name: model.SysNameLoginOptions, Value: json.RawMessage(`{}`)}, true, false},
		{"amd_cancel_not_human is not pushed", &model.SystemSetting{Name: model.SysNameAmdCancelNotHuman, Value: json.RawMessage(`true`)}, true, false},
		{"scheme_version_limit is not pushed", &model.SystemSetting{Name: model.SysNameSchemeVersionLimit, Value: json.RawMessage(`10`)}, true, false},
		{"sensitive setting is not pushed", &model.SystemSetting{Name: model.SysNameDefaultPassword, Value: json.RawMessage(`"x"`)}, true, false},
		{"unknown setting is not pushed", &model.SystemSetting{Name: "brand_new_setting", Value: json.RawMessage(`1`)}, true, false},
		{"nil setting", nil, true, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			n := model.NewSystemSettingsNotification(7, tc.setting)

			if tc.wantNil {
				if n != nil {
					t.Fatalf("expected no notification, got %+v", n)
				}

				return
			}

			if n == nil {
				t.Fatal("expected notification, got nil")
			}

			if n.Action != model.NotificationSystemSettings || !n.Broadcast || n.DomainId != 7 || len(n.ForUsers) != 0 {
				t.Fatalf("unexpected notification: %+v", n)
			}

			body, _ := n.Body.(map[string]any)
			if body["name"] != tc.setting.Name {
				t.Fatalf("expected name %q in body, got %v", tc.setting.Name, body)
			}

			if _, ok := body["value"]; ok != tc.wantValue {
				t.Fatalf("value present = %v, want %v", ok, tc.wantValue)
			}
		})
	}
}

func TestNewSystemSettingsNotification_WireShape(t *testing.T) {
	n := model.NewSystemSettingsNotification(7, &model.SystemSetting{Name: model.SysNameShowFullContact, Value: json.RawMessage(`true`)})

	var got struct {
		Event string `json:"event"`
		Data  struct {
			Notification map[string]json.RawMessage `json:"notification"`
		} `json:"data"`
	}
	if err := json.Unmarshal([]byte(model.NewWebSocketNotificationEvent(n).ToJson()), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if got.Event != model.WebsocketNotificationEvent {
		t.Fatalf("expected event %q, got %q", model.WebsocketNotificationEvent, got.Event)
	}

	if string(got.Data.Notification["action"]) != `"system_settings"` {
		t.Fatalf("expected action system_settings, got %s", got.Data.Notification["action"])
	}

	if string(got.Data.Notification["body"]) != `{"name":"show_full_contact","value":true}` {
		t.Fatalf("unexpected body: %s", got.Data.Notification["body"])
	}

	if _, ok := got.Data.Notification["broadcast"]; ok {
		t.Fatal("broadcast flag must not be serialized")
	}
}

func TestSystemSetting_ValueEquals(t *testing.T) {
	cases := []struct {
		name string
		a    string
		b    string
		want bool
	}{
		{"equal ints", `1`, `1`, true},
		{"changed ints", `1`, `2`, false},
		{"equal bools", `true`, `true`, true},
		{"changed bools", `true`, `false`, false},
		{"equal strings", `"hello"`, `"hello"`, true},
		{"changed strings", `"hello"`, `"world"`, false},
		{"object key order", `{"a":1,"b":2}`, `{"b":2,"a":1}`, true},
		{"object whitespace", `{"a":1}`, `{ "a" : 1 }`, true},
		{"changed object", `{"a":1}`, `{"a":2}`, false},
		{"equal arrays", `[1,2,3]`, `[1,2,3]`, true},
		{"changed arrays", `[1,2,3]`, `[1,2]`, false},
		{"int vs float form", `1`, `1.0`, true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			a := &model.SystemSetting{Value: json.RawMessage(tc.a)}
			b := &model.SystemSetting{Value: json.RawMessage(tc.b)}

			if got := a.ValueEquals(b); got != tc.want {
				t.Fatalf("ValueEquals(%s, %s) = %v, want %v", tc.a, tc.b, got, tc.want)
			}

			if got := b.ValueEquals(a); got != tc.want {
				t.Fatalf("ValueEquals(%s, %s) = %v, want %v (symmetry)", tc.b, tc.a, got, tc.want)
			}
		})
	}
}

func TestPrepareDefaultMembersFilter_Default(t *testing.T) {
	before := time.Now().UTC()

	result := model.PrepareDefaultMembersFilter(model.SysValue{})

	after := time.Now().UTC()

	expectedFrom := time.Date(
		before.Year(),
		before.Month()-1,
		1,
		0,
		0,
		0,
		0,
		time.UTC,
	).UnixMilli()

	if result.From != expectedFrom {
		t.Fatalf("expected From=%d, got=%d", expectedFrom, result.From)
	}

	if result.To < before.UnixMilli() || result.To > after.UnixMilli() {
		t.Fatalf(
			"expected To between %d and %d, got=%d",
			before.UnixMilli(),
			after.UnixMilli(),
			result.To,
		)
	}
}

func TestPrepareDefaultMembersFilter_ThisDay(t *testing.T) {
	before := time.Now().UTC()

	result := model.PrepareDefaultMembersFilter(model.SysValue(`"this day"`))

	after := time.Now().UTC()

	expectedFrom := time.Date(
		before.Year(),
		before.Month(),
		before.Day(),
		0,
		0,
		0,
		0,
		time.UTC,
	).UnixMilli()

	if result.From != expectedFrom {
		t.Fatalf("expected From=%d, got=%d", expectedFrom, result.From)
	}

	if result.To < before.UnixMilli() || result.To > after.UnixMilli() {
		t.Fatalf(
			"expected To between %d and %d, got=%d",
			before.UnixMilli(),
			after.UnixMilli(),
			result.To,
		)
	}
}

func TestPrepareDefaultMembersFilter_ThisWeek(t *testing.T) {
	before := time.Now().UTC()

	result := model.PrepareDefaultMembersFilter(model.SysValue(`"this week"`))

	after := time.Now().UTC()

	weekday := int(before.Weekday())
	if weekday == 0 {
		weekday = 7
	}

	monday := before.AddDate(0, 0, -(weekday - 1))

	expectedFrom := time.Date(
		monday.Year(),
		monday.Month(),
		monday.Day(),
		0,
		0,
		0,
		0,
		time.UTC,
	).UnixMilli()

	if result.From != expectedFrom {
		t.Fatalf("expected From=%d, got=%d", expectedFrom, result.From)
	}

	if result.To < before.UnixMilli() || result.To > after.UnixMilli() {
		t.Fatalf(
			"expected To between %d and %d, got=%d",
			before.UnixMilli(),
			after.UnixMilli(),
			result.To,
		)
	}
}

func TestPrepareDefaultMembersFilter_ThisMonth(t *testing.T) {
	before := time.Now().UTC()

	result := model.PrepareDefaultMembersFilter(model.SysValue(`"this month"`))

	after := time.Now().UTC()

	expectedFrom := time.Date(
		before.Year(),
		before.Month(),
		1,
		0,
		0,
		0,
		0,
		time.UTC,
	).UnixMilli()

	if result.From != expectedFrom {
		t.Fatalf("expected From=%d, got=%d", expectedFrom, result.From)
	}

	if result.To < before.UnixMilli() || result.To > after.UnixMilli() {
		t.Fatalf(
			"expected To between %d and %d, got=%d",
			before.UnixMilli(),
			after.UnixMilli(),
			result.To,
		)
	}
}
