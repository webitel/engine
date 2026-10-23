package app

import (
	"context"
	"testing"

	"github.com/webitel/engine/model"
	"github.com/webitel/engine/store"
)

type emailLimitsStore struct {
	store.Store
	agent store.AgentStore
}

func (s emailLimitsStore) Agent() store.AgentStore { return s.agent }

type emailLimitsAgentStore struct {
	store.AgentStore
	t       *testing.T
	limits  model.AgentEmailLimits
	updated bool
}

func (s *emailLimitsAgentStore) Get(context.Context, int64, int64) (*model.Agent, model.AppError) {
	return &model.Agent{ChatCount: 1, TaskCount: 1, EmailCount: 2, ExtraEmailCount: 3}, nil
}

func (s *emailLimitsAgentStore) Update(_ context.Context, a *model.Agent, limits model.AgentEmailLimits) (*model.Agent, model.AppError) {
	s.t.Helper()
	if limits != s.limits {
		s.t.Fatal("Update lost Email-limit presence")
	}
	// Model a newer database value; this checks forwarding, not PostgreSQL locking.
	wantBase, wantExtra := uint32(2), uint32(3)
	currentBase, currentExtra := uint32(9), uint32(8)
	if limits.Base != nil {
		wantBase, currentBase = *limits.Base, *limits.Base
	}
	if limits.Extra != nil {
		wantExtra, currentExtra = *limits.Extra, *limits.Extra
	}
	if a.EmailCount != wantBase || a.ExtraEmailCount != wantExtra {
		s.t.Fatal("Email limits were not applied before validation and persistence")
	}
	a.EmailCount, a.ExtraEmailCount = currentBase, currentExtra
	s.updated = true
	return a, nil
}

func TestAgentUpdateAndPatchPreserveEmailLimitPresence(t *testing.T) {
	zero, five := uint32(0), uint32(5)
	for _, patch := range []bool{false, true} {
		method := "Update"
		if patch {
			method = "Patch"
		}
		for _, tc := range []struct {
			name        string
			base, extra *uint32
			wantBase    uint32
			wantExtra   uint32
		}{
			{name: "omitted", wantBase: 9, wantExtra: 8},
			{name: "zero base", base: &zero, wantBase: 0, wantExtra: 8},
			{name: "zero Extra", extra: &zero, wantBase: 9, wantExtra: 0},
			{name: "both", base: &five, extra: &zero, wantBase: 5, wantExtra: 0},
		} {
			t.Run(method+"/"+tc.name, func(t *testing.T) {
				limits := model.AgentEmailLimits{Base: tc.base, Extra: tc.extra}
				agents := &emailLimitsAgentStore{t: t, limits: limits}
				a := &App{Store: emailLimitsStore{agent: agents}}
				var got *model.Agent
				var err model.AppError
				if patch {
					got, err = a.PatchAgent(context.Background(), 1, 2, &model.AgentPatch{
						EmailCount: tc.base, ExtraEmailCount: tc.extra,
					})
				} else {
					got, err = a.UpdateAgent(context.Background(), &model.Agent{
						ChatCount: 1, TaskCount: 1,
					}, limits)
				}
				if err != nil || got == nil || !agents.updated {
					t.Fatalf("update failed: %v", err)
				}
				if got.EmailCount != tc.wantBase || got.ExtraEmailCount != tc.wantExtra {
					t.Fatalf("Email limits = %d/%d, want %d/%d", got.EmailCount, got.ExtraEmailCount, tc.wantBase, tc.wantExtra)
				}
			})
		}
	}
}
