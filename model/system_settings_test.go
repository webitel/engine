package model_test

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/webitel/engine/model"
)

func TestNewWebSocketSystemSettingsEvent(t *testing.T) {
	names := []string{"enable_omnichannel", "password_min_length"}

	ev := model.NewWebSocketSystemSettingsEvent(names)

	if ev.EventType() != model.WebsocketSystemSettingsEvent {
		t.Fatalf("expected event type %q, got %q", model.WebsocketSystemSettingsEvent, ev.EventType())
	}

	got, ok := ev.Data["names"].([]string)
	if !ok {
		t.Fatalf("expected data[names] to be []string, got %T", ev.Data["names"])
	}

	if !reflect.DeepEqual(got, names) {
		t.Fatalf("expected names %v, got %v", names, got)
	}
}

func TestSystemSettingsChange_JsonRoundTrip(t *testing.T) {
	in := &model.SystemSettingsChange{Names: []string{"enable_2fa", "default_workspace_tab"}}

	var out model.SystemSettingsChange
	if err := json.Unmarshal([]byte(in.ToJSON()), &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if !reflect.DeepEqual(in.Names, out.Names) {
		t.Fatalf("expected names %v, got %v", in.Names, out.Names)
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
