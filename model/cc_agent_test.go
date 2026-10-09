package model_test

import (
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
		{name: "zero disables email", emailCount: 0},
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
