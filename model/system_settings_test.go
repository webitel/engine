package model_test

import (
	"encoding/json"
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
