package model_test

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/webitel/engine/model"
)

func TestNewWebSocketSystemSettingsEvent(t *testing.T) {
	c := &model.SystemSettingsChange{Name: "enable_omnichannel", Value: json.RawMessage(`true`)}

	ev := model.NewWebSocketSystemSettingsEvent(c)

	if ev.EventType() != model.WebsocketSystemSettingsEvent {
		t.Fatalf("expected event type %q, got %q", model.WebsocketSystemSettingsEvent, ev.EventType())
	}

	if got, _ := ev.Data["name"].(string); got != c.Name {
		t.Fatalf("expected name %q, got %v", c.Name, ev.Data["name"])
	}

	if _, ok := ev.Data["value"]; !ok {
		t.Fatalf("expected value in event data")
	}
}

func TestNewWebSocketSystemSettingsEvent_NoValue(t *testing.T) {
	ev := model.NewWebSocketSystemSettingsEvent(&model.SystemSettingsChange{Name: model.SysNameDefaultPassword})

	if _, ok := ev.Data["value"]; ok {
		t.Fatalf("expected no value in event data for a name-only change")
	}
}

func TestNewSystemSettingsChange_Sanitizes(t *testing.T) {
	cases := []struct {
		name      string
		setting   string
		wantValue bool
	}{
		{"regular carries value", model.SysNameOmnichannel, true},
		{"default_password omits value", model.SysNameDefaultPassword, false},
		{"chat_ai_connection omits value", model.SysNameChatAiConnection, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := model.NewSystemSettingsChange(&model.SystemSetting{Name: tc.setting, Value: json.RawMessage(`"x"`)})

			if tc.wantValue && c.Value == nil {
				t.Fatalf("expected value present for %s", tc.setting)
			}

			if !tc.wantValue && c.Value != nil {
				t.Fatalf("expected value omitted for sensitive %s", tc.setting)
			}
		})
	}
}

func TestSystemSettingsChange_JsonRoundTrip(t *testing.T) {
	in := &model.SystemSettingsChange{Name: "enable_2fa", Value: json.RawMessage(`true`)}

	var out model.SystemSettingsChange
	if err := json.Unmarshal([]byte(in.ToJSON()), &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if out.Name != in.Name {
		t.Fatalf("expected name %q, got %q", in.Name, out.Name)
	}

	if !reflect.DeepEqual([]byte(in.Value), []byte(out.Value)) {
		t.Fatalf("expected value %s, got %s", in.Value, out.Value)
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
