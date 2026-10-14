package model_test

import (
	"encoding/json"
	"math"
	"slices"
	"testing"

	"github.com/webitel/engine/model"
)

func TestAgent_EmailCount(t *testing.T) {
	t.Parallel()

	valid := func(emailCount uint32) model.Agent {
		return model.Agent{ChatCount: 1, TaskCount: 1, EmailCount: emailCount}
	}

	for _, tc := range []struct {
		name       string
		emailCount uint32
		wantErr    bool
	}{
		{name: "zero disables base slots", emailCount: 0},
		{name: "smallint maximum", emailCount: math.MaxInt16},
		{name: "smallint overflow", emailCount: math.MaxInt16 + 1, wantErr: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			a := valid(tc.emailCount)
			err := a.IsValid()
			if tc.wantErr != (err != nil) {
				t.Fatalf("IsValid() = %v, wantErr %v", err, tc.wantErr)
			}
			if err != nil && err.GetId() != "model.Agent.valid.EmailCount" {
				t.Fatalf("error id = %s", err.GetId())
			}
		})
	}

	t.Run("patch changes only the given limit", func(t *testing.T) {
		a := model.Agent{ChatCount: 3, EmailCount: 2}
		a.Patch(&model.AgentPatch{})
		if a.EmailCount != 2 {
			t.Fatalf("patch without email_count changed it to %d", a.EmailCount)
		}

		five := uint32(5)
		a.Patch(&model.AgentPatch{EmailCount: &five})
		if a.EmailCount != 5 || a.ChatCount != 3 {
			t.Fatalf("got email_count %d, chat_count %d", a.EmailCount, a.ChatCount)
		}
	})
}

func TestAgent_EmailCountIsSearchedByDefault(t *testing.T) {
	t.Parallel()

	var a model.Agent
	if !slices.Contains(a.DefaultFields(), "email_count") || !slices.Contains(a.AllowFields(), "email_count") {
		t.Fatal("email_count is not a default and allowed search field")
	}
}

func TestAgent_ExtraEmailCount(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name    string
		count   uint32
		wantErr bool
	}{
		{name: "zero disables Extra slots"},
		{name: "Extra with zero base", count: 1},
		{name: "smallint maximum", count: math.MaxInt16},
		{name: "smallint overflow", count: math.MaxInt16 + 1, wantErr: true},
		{name: "uint32 maximum", count: math.MaxUint32, wantErr: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			a := model.Agent{ChatCount: 1, TaskCount: 1, EmailCount: 0, ExtraEmailCount: tc.count}
			err := a.IsValid()
			if tc.wantErr != (err != nil) {
				t.Fatalf("IsValid() = %v, wantErr %v", err, tc.wantErr)
			}
			if err != nil && err.GetId() != "model.Agent.valid.ExtraEmailCount" {
				t.Fatalf("error id = %s", err.GetId())
			}
		})
	}
}

func TestAgent_EmailLimitsPatchIndependently(t *testing.T) {
	t.Parallel()

	zero, five := uint32(0), uint32(5)
	for _, tc := range []struct {
		name        string
		patch       model.AgentPatch
		base, extra uint32
	}{
		{name: "omitted limits", base: 2, extra: 3},
		{name: "explicit zero base", patch: model.AgentPatch{EmailCount: &zero}, base: 0, extra: 3},
		{name: "explicit zero Extra", patch: model.AgentPatch{ExtraEmailCount: &zero}, base: 2, extra: 0},
		{name: "change Extra only", patch: model.AgentPatch{ExtraEmailCount: &five}, base: 2, extra: 5},
		{name: "both limits", patch: model.AgentPatch{EmailCount: &five, ExtraEmailCount: &zero}, base: 5, extra: 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			a := model.Agent{ChatCount: 4, ExtraChatCount: 6, TaskCount: 7, EmailCount: 2, ExtraEmailCount: 3}
			a.Patch(&tc.patch)
			if a.EmailCount != tc.base || a.ExtraEmailCount != tc.extra {
				t.Fatalf("Email limits = %d/%d, want %d/%d", a.EmailCount, a.ExtraEmailCount, tc.base, tc.extra)
			}
			if a.ChatCount != 4 || a.ExtraChatCount != 6 || a.TaskCount != 7 {
				t.Fatal("Email patch changed other channel limits")
			}
		})
	}
}

func TestAgent_ExtraEmailCountIsSearchedByDefault(t *testing.T) {
	t.Parallel()

	var a model.Agent
	if !slices.Contains(a.DefaultFields(), "extra_email_count") || !slices.Contains(a.AllowFields(), "extra_email_count") {
		t.Fatal("extra_email_count is not a default and allowed search field")
	}
}

func TestAgent_EmailLimitsRejectNegativeJSON(t *testing.T) {
	t.Parallel()

	for _, input := range []string{`{"email_count":-1}`, `{"extra_email_count":-1}`} {
		var a model.Agent
		if err := json.Unmarshal([]byte(input), &a); err == nil {
			t.Fatalf("accepted negative limit: %s", input)
		}
	}
}

func TestAgent_EmailPostponedCount(t *testing.T) {
	t.Parallel()

	a := model.Agent{ChatCount: 1, TaskCount: 1, EmailPostponedCount: math.MaxInt16}
	if err := a.IsValid(); err != nil {
		t.Fatalf("smallint maximum: %v", err)
	}
	a.EmailPostponedCount++
	if err := a.IsValid(); err == nil || err.GetId() != "model.Agent.valid.EmailPostponedCount" {
		t.Fatalf("smallint overflow: %v", err)
	}

	b := model.Agent{EmailCount: 2, EmailPostponedCount: 1}
	three := uint32(3)
	b.Patch(&model.AgentPatch{EmailPostponedCount: &three})
	if b.EmailPostponedCount != 3 || b.EmailCount != 2 {
		t.Fatalf("got email_postponed_count %d, email_count %d", b.EmailPostponedCount, b.EmailCount)
	}

	if !slices.Contains(b.DefaultFields(), "email_postponed_count") || !slices.Contains(b.AllowFields(), "email_postponed_count") {
		t.Fatal("email_postponed_count is not a default and allowed search field")
	}
}
