package model_test

import (
	"encoding/json"
	"testing"

	"github.com/webitel/engine/model"
)

func TestCCTask_AcceptTimeoutJSON(t *testing.T) {
	t.Parallel()

	offeringAt := int64(1700000000000)
	timeout := 45

	tests := []struct {
		name           string
		task           model.CCTask
		wantOfferingAt any
		wantTimeout    any
	}{
		{
			name: "task with accept timeout",
			task: model.CCTask{
				Channel:           "task",
				OfferingAt:        &offeringAt,
				TeamAcceptTimeout: &timeout,
			},
			wantOfferingAt: float64(offeringAt),
			wantTimeout:    float64(timeout),
		},
		{
			name: "non-task channel omits accept timeout",
			task: model.CCTask{
				Channel:    "call",
				OfferingAt: &offeringAt,
			},
			wantOfferingAt: float64(offeringAt),
		},
		{
			name: "nothing set omits both",
			task: model.CCTask{Channel: "chat"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			raw, err := json.Marshal(tt.task)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}

			var got map[string]any
			if err := json.Unmarshal(raw, &got); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}

			assertField(t, got, "offering_at", tt.wantOfferingAt)
			assertField(t, got, "team_accept_timeout", tt.wantTimeout)
		})
	}
}

func assertField(t *testing.T, got map[string]any, key string, want any) {
	t.Helper()

	v, ok := got[key]
	if want == nil {
		if ok {
			t.Fatalf("%s = %v, want absent", key, v)
		}

		return
	}

	if v != want {
		t.Fatalf("%s = %v, want %v", key, v, want)
	}
}
