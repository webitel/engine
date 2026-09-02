package model_test

import (
	"testing"

	"github.com/webitel/engine/model"
)

func TestMutateHistoryAttempt_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		mutate        *model.MutateHistoryAttempt
		expectError   bool
		expectedErrID string
	}{
		{
			name: "Success by ID",
			mutate: &model.MutateHistoryAttempt{
				ID:       100,
				DomainID: 1,
				Fields:   []string{"description"},
			},
			expectError: false,
		},
		{
			name: "Success by MemberCallID",
			mutate: &model.MutateHistoryAttempt{
				MemberCallID: "member-call-uuid",
				DomainID:     1,
				Fields:       []string{"variables"},
			},
			expectError: false,
		},
		{
			name: "Success by AgentCallID",
			mutate: &model.MutateHistoryAttempt{
				AgentCallID: "agent-call-uuid",
				DomainID:    1,
				Fields:      []string{"description", "variables"},
			},
			expectError: false,
		},
		{
			name:          "Error: empty filter",
			mutate:        model.NewMutateHistoryAttempt(0, "", "", "test", nil, "description"),
			expectError:   true,
			expectedErrID: "model.cc_member.mutate_history_attempt.validate.empty_filter",
		},
		{
			name: "Error: empty fields slice",
			mutate: &model.MutateHistoryAttempt{
				ID:       100,
				DomainID: 1,
				Fields:   []string{},
			},
			expectError:   true,
			expectedErrID: "model.cc_member.mutate_history_attempt.validate.empty_fields",
		},
		{
			name: "Error: nil fields slice",
			mutate: &model.MutateHistoryAttempt{
				ID:       100,
				DomainID: 1,
				Fields:   nil,
			},
			expectError:   true,
			expectedErrID: "model.cc_member.mutate_history_attempt.validate.empty_fields",
		},
		{
			name: "Error: domain ID is zero",
			mutate: &model.MutateHistoryAttempt{
				ID:       100,
				DomainID: 0,
				Fields:   []string{"description"},
			},
			expectError:   true,
			expectedErrID: "model.cc_member.mutate_history_attempt.validate.empty_domain",
		},
		{
			name: "Error: negative domain ID",
			mutate: &model.MutateHistoryAttempt{
				ID:       100,
				DomainID: -5,
				Fields:   []string{"description"},
			},
			expectError:   true,
			expectedErrID: "model.cc_member.mutate_history_attempt.validate.empty_domain",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.mutate.Validate()

			if tt.expectError {
				if err == nil {
					t.Fatalf("expected validation error, got nil")
				}
				if err.GetId() != tt.expectedErrID {
					t.Errorf("error id mismatch: got %q, want %q", err.GetId(), tt.expectedErrID)
				}
			} else if err != nil {
				t.Fatalf("expected no validation error, got: %v", err)
			}
		})
	}
}
