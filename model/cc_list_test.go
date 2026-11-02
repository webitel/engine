package model_test

import (
	"testing"
	"time"

	"github.com/webitel/engine/model"
)

func TestListCommunication_IsValid(t *testing.T) {
	t.Parallel()

	past := time.Now().Add(-time.Hour)
	future := time.Now().Add(time.Hour)

	tests := []struct {
		name    string
		comm    model.ListCommunication
		wantErr bool
		errID   string
	}{
		{
			name: "valid number",
			comm: model.ListCommunication{Number: "380638140008"},
		},
		{
			name:    "empty number",
			comm:    model.ListCommunication{Number: ""},
			wantErr: true,
			errID:   "list_communication.valid.number",
		},
		{
			name:    "whitespace-only number",
			comm:    model.ListCommunication{Number: "   "},
			wantErr: true,
			errID:   "list_communication.valid.number",
		},
		{
			name:    "empty number with description",
			comm:    model.ListCommunication{Description: "без номера"},
			wantErr: true,
			errID:   "list_communication.valid.number",
		},
		{
			name:    "expired",
			comm:    model.ListCommunication{Number: "380638140008", ExpireAt: &past},
			wantErr: true,
			errID:   "list_communication.valid.expire_at",
		},
		{
			name: "future expire",
			comm: model.ListCommunication{Number: "380638140008", ExpireAt: &future},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.comm.IsValid()
			if (err != nil) != tt.wantErr {
				t.Fatalf("IsValid() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr && err.GetId() != tt.errID {
				t.Errorf("IsValid() error ID = %s, want %s", err.GetId(), tt.errID)
			}
		})
	}
}
