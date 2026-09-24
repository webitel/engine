package model_test

import (
	"testing"

	"github.com/webitel/engine/model"
)

func TestQueue_IsValid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		queue   *model.Queue
		wantErr bool
		errCode string
	}{
		{
			name:    "nil queue",
			queue:   nil,
			wantErr: true,
			errCode: "model.cc_queue.validate.empty",
		},
		{
			name: "nil payload",
			queue: &model.Queue{
				Payload: nil,
			},
			wantErr: true,
			errCode: "model.cc_queue.validate.payload.empty",
		},
		{
			name: "valid inbound queue without calendar",
			queue: &model.Queue{
				Type:    model.QueueTypeInboundCall,
				Payload: model.StringInterface{},
			},
			wantErr: false,
		},
		{
			name: "valid progressive queue with valid payload and calendar",
			queue: &model.Queue{
				Type: model.QueueTypeProgressiveCall,
				Payload: model.StringInterface{
					model.QueuePayloadProgressiveCountKey: 2,
					model.QueuePayloadMaxAgentLineKey:     5,
				},
				Calendar: &model.Lookup{Id: 1, Name: "Default"},
			},
			wantErr: false,
		},
		{
			name: "progressive queue sets default progressive count when missing",
			queue: &model.Queue{
				Type: model.QueueTypeProgressiveCall,
				Payload: model.StringInterface{
					model.QueuePayloadMaxAgentLineKey: 3,
				},
				Calendar: &model.Lookup{Id: 1},
			},
			wantErr: false,
		},
		{
			name: "progressive queue sets default progressive count when non-positive",
			queue: &model.Queue{
				Type: model.QueueTypeProgressiveCall,
				Payload: model.StringInterface{
					model.QueuePayloadProgressiveCountKey: 0,
					model.QueuePayloadMaxAgentLineKey:     3,
				},
				Calendar: &model.Lookup{Id: 1},
			},
			wantErr: false,
		},
		{
			name: "progressive queue float decimal rejected in progressive count",
			queue: &model.Queue{
				Type: model.QueueTypeProgressiveCall,
				Payload: model.StringInterface{
					model.QueuePayloadProgressiveCountKey: 2.5,
					model.QueuePayloadMaxAgentLineKey:     3,
				},
				Calendar: &model.Lookup{Id: 1},
			},
			wantErr: true,
			errCode: "model.cc_queue.validate.config.decimal_not_allowed",
		},
		{
			name: "progressive queue float integer accepted in progressive count",
			queue: &model.Queue{
				Type: model.QueueTypeProgressiveCall,
				Payload: model.StringInterface{
					model.QueuePayloadProgressiveCountKey: float64(2),
					model.QueuePayloadMaxAgentLineKey:     3,
				},
				Calendar: &model.Lookup{Id: 1},
			},
			wantErr: false,
		},
		{
			name: "progressive queue string number accepted in progressive count",
			queue: &model.Queue{
				Type: model.QueueTypeProgressiveCall,
				Payload: model.StringInterface{
					model.QueuePayloadProgressiveCountKey: "3",
					model.QueuePayloadMaxAgentLineKey:     3,
				},
				Calendar: &model.Lookup{Id: 1},
			},
			wantErr: false,
		},
		{
			name: "progressive queue invalid string rejected in progressive count",
			queue: &model.Queue{
				Type: model.QueueTypeProgressiveCall,
				Payload: model.StringInterface{
					model.QueuePayloadProgressiveCountKey: "invalid",
					model.QueuePayloadMaxAgentLineKey:     3,
				},
				Calendar: &model.Lookup{Id: 1},
			},
			wantErr: true,
			errCode: "model.cc_queue.validate.config.invalid_string_number",
		},
		{
			name: "progressive queue unsupported type rejected in progressive count",
			queue: &model.Queue{
				Type: model.QueueTypeProgressiveCall,
				Payload: model.StringInterface{
					model.QueuePayloadProgressiveCountKey: true,
					model.QueuePayloadMaxAgentLineKey:     3,
				},
				Calendar: &model.Lookup{Id: 1},
			},
			wantErr: true,
			errCode: "model.cc_queue.validate.config.unsupported_type",
		},
		{
			name: "progressive queue missing max agent line",
			queue: &model.Queue{
				Type:     model.QueueTypeProgressiveCall,
				Payload:  model.StringInterface{},
				Calendar: &model.Lookup{Id: 1},
			},
			wantErr: false,
		},
		{
			name: "progressive queue non-positive max agent line",
			queue: &model.Queue{
				Type: model.QueueTypeProgressiveCall,
				Payload: model.StringInterface{
					model.QueuePayloadMaxAgentLineKey: 0,
				},
				Calendar: &model.Lookup{Id: 1},
			},
			wantErr: false,
		},
		{
			name: "predictive queue requires calendar",
			queue: &model.Queue{
				Type: model.QueueTypePredictCall,
				Payload: model.StringInterface{
					model.QueuePayloadMaxAgentLineKey: 1,
				},
				Calendar: nil,
			},
			wantErr: true,
			errCode: "model.cc_queue.validate.calendar.required",
		},
		{
			name: "predictive queue empty calendar struct",
			queue: &model.Queue{
				Type: model.QueueTypePredictCall,
				Payload: model.StringInterface{
					model.QueuePayloadMaxAgentLineKey: 1,
				},
				Calendar: &model.Lookup{Id: 0, Name: ""},
			},
			wantErr: true,
			errCode: "model.cc_queue.validate.calendar.required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.queue.IsValid()

			if (err != nil) != tt.wantErr {
				t.Fatalf("IsValid() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr && err != nil && tt.errCode != "" {
				if err.GetId() != tt.errCode {
					t.Errorf("IsValid() err.Id = %v, want %v", err.GetId(), tt.errCode)
				}
			}
		})
	}
}
