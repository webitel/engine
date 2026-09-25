package model_test

import (
	"testing"

	"github.com/webitel/engine/model"
)

func TestMember_IsValid_CommunicationDestination(t *testing.T) {
	t.Parallel()

	const destinationFormatErrId = "model.member.is_valid.communications.destination.format.app_error"
	const destinationEmptyErrId = "model.member.is_valid.communications.destination.app_error"

	tests := []struct {
		name        string
		destination string
		wantErrId   string
	}{
		{
			name:        "valid_destination",
			destination: "+380-99_111.22!33",
		},
		{
			name:        "empty",
			destination: "",
			wantErrId:   destinationEmptyErrId,
		},
		{
			name:        "only_whitespace",
			destination: "   ",
			wantErrId:   destinationEmptyErrId,
		},
		{
			name:        "invalid_format",
			destination: "380 99 111 22 33",
			wantErrId:   destinationFormatErrId,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			m := &model.Member{
				Name: "test",
				Communications: []*model.MemberCommunication{
					{
						Destination: tt.destination,
						Type:        model.Lookup{Id: 1},
					},
				},
			}

			err := m.IsValid(10)

			if tt.wantErrId == "" {
				if err != nil {
					t.Fatalf("expected nil error, got %v", err)
				}
				return
			}

			if err == nil {
				t.Fatalf("expected error id %q, got nil", tt.wantErrId)
			}
			if err.GetId() != tt.wantErrId {
				t.Fatalf("expected error id %q, got %q (%v)", tt.wantErrId, err.GetId(), err)
			}
		})
	}
}
