package model_test

import (
	"testing"

	"github.com/webitel/engine/model"
)

func TestQueue_IsValid_ProgressiveCountValidation(t *testing.T) {
	t.Parallel()

	validCalendar := &model.Lookup{Id: 1}

	tests := []struct {
		name    string
		queue   model.Queue
		wantErr bool
		errID   string
	}{
		{
			name: "predictive queue valid progressive count int",
			queue: model.Queue{
				Type:     model.QueueTypePredictCall,
				Calendar: validCalendar,
				Payload: model.StringInterface{
					model.QueuePayloadProgressiveCountKey: 5,
				},
			},
			wantErr: false,
		},
		{
			name: "progressive queue valid progressive count string",
			queue: model.Queue{
				Type:     model.QueueTypeProgressiveCall,
				Calendar: validCalendar,
				Payload: model.StringInterface{
					model.QueuePayloadProgressiveCountKey: "10",
				},
			},
			wantErr: false,
		},
		{
			name: "progressive count does not exist",
			queue: model.Queue{
				Type:     model.QueueTypePredictCall,
				Calendar: validCalendar,
				Payload:  model.StringInterface{},
			},
			wantErr: true,
			errID:   "model.queue.valid.progressive_count_not_exist",
		},
		{
			name: "progressive count invalid string",
			queue: model.Queue{
				Type:     model.QueueTypeProgressiveCall,
				Calendar: validCalendar,
				Payload: model.StringInterface{
					model.QueuePayloadProgressiveCountKey: "abc",
				},
			},
			wantErr: true,
			errID:   "model.queue.valid.progressive_count_unconvertable_from_string",
		},
		{
			name: "progressive count unsupported type",
			queue: model.Queue{
				Type:     model.QueueTypePredictCall,
				Calendar: validCalendar,
				Payload: model.StringInterface{
					model.QueuePayloadProgressiveCountKey: true,
				},
			},
			wantErr: true,
			errID:   "model.queue.valid.unsupported_type",
		},
		{
			name: "progressive count zero",
			queue: model.Queue{
				Type:     model.QueueTypeProgressiveCall,
				Calendar: validCalendar,
				Payload: model.StringInterface{
					model.QueuePayloadProgressiveCountKey: 0,
				},
			},
			wantErr: true,
			errID:   "model.queue.valid.progressive_count_must_be_gt_zero",
		},
		{
			name: "progressive count negative",
			queue: model.Queue{
				Type:     model.QueueTypePredictCall,
				Calendar: validCalendar,
				Payload: model.StringInterface{
					model.QueuePayloadProgressiveCountKey: -5,
				},
			},
			wantErr: true,
			errID:   "model.queue.valid.progressive_count_must_be_gt_zero",
		},
		{
			name: "non progressive queue skips validation",
			queue: model.Queue{
				Type:     model.QueueTypeInboundCall,
				Calendar: nil,
				Payload: model.StringInterface{
					model.QueuePayloadProgressiveCountKey: "invalid",
				},
			},
			wantErr: false,
		},
		{
			name: "nil payload for predictive queue",
			queue: model.Queue{
				Type:     model.QueueTypePredictCall,
				Calendar: validCalendar,
				Payload:  nil,
			},
			wantErr: true,
			errID:   "model.queue.valid.progressive_count_not_exist",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := tt.queue.IsValid()

			if tt.wantErr {
				if got == nil {
					t.Fatalf("expected error, got nil")
				}

				if got.GetId() != tt.errID {
					t.Fatalf("unexpected error id: got %s, want %s", got.GetId(), tt.errID)
				}

				return
			}

			if got != nil {
				t.Fatalf("expected nil error, got %v", got)
			}
		})
	}
}
