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

func u32(v uint32) *uint32 { return &v }

func sameLimit(a, b *uint32) bool {
	if a == nil || b == nil {
		return a == b
	}
	return *a == *b
}

func (s *emailLimitsAgentStore) Get(context.Context, int64, int64) (*model.Agent, model.AppError) {
	return &model.Agent{ChatCount: 1, TaskCount: 1, EmailCount: 2, ExtraEmailCount: 3, EmailPostponedCount: u32(4)}, nil
}

func (s *emailLimitsAgentStore) Update(_ context.Context, a *model.Agent, limits model.AgentEmailLimits) (*model.Agent, model.AppError) {
	s.t.Helper()
	if limits != s.limits {
		s.t.Fatal("Update lost Email-limit presence")
	}
	// Model a newer database value; this checks forwarding, not PostgreSQL locking.
	wantBase, wantExtra, wantPostponed := uint32(2), uint32(3), u32(4)
	currentBase, currentExtra, currentPostponed := uint32(9), uint32(8), u32(7)
	if limits.Base != nil {
		wantBase, currentBase = *limits.Base, *limits.Base
	}
	if limits.Extra != nil {
		wantExtra, currentExtra = *limits.Extra, *limits.Extra
	}
	if limits.NoPostponedLimit {
		wantPostponed, currentPostponed = nil, nil
	} else if limits.Postponed != nil {
		wantPostponed, currentPostponed = limits.Postponed, u32(*limits.Postponed)
	}
	if a.EmailCount != wantBase || a.ExtraEmailCount != wantExtra || !sameLimit(a.EmailPostponedCount, wantPostponed) {
		s.t.Fatal("Email limits were not applied before validation and persistence")
	}
	a.EmailCount, a.ExtraEmailCount, a.EmailPostponedCount = currentBase, currentExtra, currentPostponed
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
			name                   string
			base, extra, postponed *uint32
			clearPostponed         bool
			wantBase               uint32
			wantExtra              uint32
			wantPostponed          *uint32
		}{
			{name: "omitted", wantBase: 9, wantExtra: 8, wantPostponed: u32(7)},
			{name: "zero base", base: &zero, wantBase: 0, wantExtra: 8, wantPostponed: u32(7)},
			{name: "zero Extra", extra: &zero, wantBase: 9, wantExtra: 0, wantPostponed: u32(7)},
			{name: "zero postponed", postponed: &zero, wantBase: 9, wantExtra: 8, wantPostponed: u32(0)},
			{name: "clear postponed", clearPostponed: true, wantBase: 9, wantExtra: 8, wantPostponed: nil},
			{name: "all", base: &five, extra: &zero, postponed: &five, wantBase: 5, wantExtra: 0, wantPostponed: u32(5)},
		} {
			t.Run(method+"/"+tc.name, func(t *testing.T) {
				limits := model.AgentEmailLimits{Base: tc.base, Extra: tc.extra, Postponed: tc.postponed, NoPostponedLimit: tc.clearPostponed}
				agents := &emailLimitsAgentStore{t: t, limits: limits}
				a := &App{Store: emailLimitsStore{agent: agents}}
				var got *model.Agent
				var err model.AppError
				if patch {
					got, err = a.PatchAgent(context.Background(), 1, 2, &model.AgentPatch{
						EmailCount: tc.base, ExtraEmailCount: tc.extra, EmailPostponedCount: tc.postponed,
						NoPostponedLimit: tc.clearPostponed,
					})
				} else {
					got, err = a.UpdateAgent(context.Background(), &model.Agent{
						ChatCount: 1, TaskCount: 1,
					}, limits)
				}
				if err != nil || got == nil || !agents.updated {
					t.Fatalf("update failed: %v", err)
				}
				if got.EmailCount != tc.wantBase || got.ExtraEmailCount != tc.wantExtra || !sameLimit(got.EmailPostponedCount, tc.wantPostponed) {
					t.Fatalf("Email limits = %d/%d/%v, want %d/%d/%v", got.EmailCount, got.ExtraEmailCount, got.EmailPostponedCount,
						tc.wantBase, tc.wantExtra, tc.wantPostponed)
				}
			})
		}
	}
}
