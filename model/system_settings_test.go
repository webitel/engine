package model_test

import (
	"testing"
	"time"

	"github.com/webitel/engine/model"
)

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

func TestSystemSettingValueEquals(t *testing.T) {
	tests := []struct {
		name string
		a, b *model.SystemSetting
		want bool
	}{
		{"equal bool", &model.SystemSetting{Value: []byte(`true`)}, &model.SystemSetting{Value: []byte(`true`)}, true},
		{"different bool", &model.SystemSetting{Value: []byte(`true`)}, &model.SystemSetting{Value: []byte(`false`)}, false},
		{"equal int", &model.SystemSetting{Value: []byte(`30`)}, &model.SystemSetting{Value: []byte(`30`)}, true},
		{"different int", &model.SystemSetting{Value: []byte(`30`)}, &model.SystemSetting{Value: []byte(`31`)}, false},
		{"equal string", &model.SystemSetting{Value: []byte(`"this day"`)}, &model.SystemSetting{Value: []byte(`"this day"`)}, true},
		{"different string", &model.SystemSetting{Value: []byte(`"this day"`)}, &model.SystemSetting{Value: []byte(`"this week"`)}, false},
		{"object key order", &model.SystemSetting{Value: []byte(`{"a":1,"b":2}`)}, &model.SystemSetting{Value: []byte(`{"b":2, "a":1}`)}, true},
		{"different object", &model.SystemSetting{Value: []byte(`{"a":1}`)}, &model.SystemSetting{Value: []byte(`{"a":2}`)}, false},
		{"invalid json equal bytes", &model.SystemSetting{Value: []byte(`{`)}, &model.SystemSetting{Value: []byte(`{`)}, true},
		{"invalid json different bytes", &model.SystemSetting{Value: []byte(`{`)}, &model.SystemSetting{Value: []byte(`[`)}, false},
		{"both nil", nil, nil, true},
		{"one nil", &model.SystemSetting{Value: []byte(`true`)}, nil, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.a.ValueEquals(tc.b); got != tc.want {
				t.Fatalf("ValueEquals() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestNewSystemSettingEventFromRoutingKey(t *testing.T) {
	tests := []struct {
		name    string
		rk      string
		want    *model.SystemSettingEvent
		wantErr bool
	}{
		{"update", "system_settings.period_to_playback_records.update.1.10", &model.SystemSettingEvent{Name: "period_to_playback_records", DomainID: 1}, false},
		{"delete", "system_settings.enable_2fa.delete.25.3", &model.SystemSettingEvent{Name: "enable_2fa", DomainID: 25}, false},
		{"without user", "system_settings.enable_2fa.create.7", &model.SystemSettingEvent{Name: "enable_2fa", DomainID: 7}, false},
		{"too short", "system_settings.enable_2fa.update", nil, true},
		{"wrong object", "domains.enable_2fa.update.1.1", nil, true},
		{"empty name", "system_settings..update.1.1", nil, true},
		{"domain not a number", "system_settings.enable_2fa.update.abc.1", nil, true},
		{"domain zero", "system_settings.enable_2fa.update.0.1", nil, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := model.NewSystemSettingEventFromRoutingKey(tc.rk)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got %+v", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if *got != *tc.want {
				t.Fatalf("got %+v, want %+v", got, tc.want)
			}
		})
	}
}
