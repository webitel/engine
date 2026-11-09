package wsapi

import (
	"strings"
	"testing"

	"github.com/webitel/engine/model"
)

func emailRequest(extra map[string]interface{}) *model.WebSocketRequest {
	d := map[string]interface{}{
		"thread_id": float64(101), "attempt_id": float64(7), "expected_version": float64(0),
		"command_id": "0199b2a4-3c4e-7a10-8f2e-1b2c3d4e5f60",
	}
	for k, v := range extra {
		if v == nil {
			delete(d, k)
		} else {
			d[k] = v
		}
	}
	return &model.WebSocketRequest{Action: "cc_email_test", Data: d}
}

func TestEmailCommandFromRequest(t *testing.T) {
	cmd, err := emailCommandFromRequest(emailRequest(map[string]interface{}{"queue_id": float64(5), "comment": "  to sales "}), model.EmailAttemptTransfer)
	if err != nil || cmd.ToQueueId != 5 || cmd.ToAgentId != 0 || cmd.Comment != "to sales" || cmd.ExpectedVersion != 0 {
		t.Fatalf("transfer: %+v, %v", cmd, err)
	}
	cmd, err = emailCommandFromRequest(emailRequest(map[string]interface{}{"postpone_for_sec": float64(60)}), model.EmailAttemptPostpone)
	if err != nil || cmd.PostponeForSec != 60 || cmd.PostponedUntil != 0 {
		t.Fatalf("postpone: %+v, %v", cmd, err)
	}

	for name, tc := range map[string]struct {
		op    string
		extra map[string]interface{}
	}{
		"no expected version":   {model.EmailAttemptTake, map[string]interface{}{"expected_version": nil}},
		"fractional attempt":    {model.EmailAttemptTake, map[string]interface{}{"attempt_id": 7.5}},
		"not a UUID":            {model.EmailAttemptTake, map[string]interface{}{"command_id": "c-1"}},
		"two deadlines":         {model.EmailAttemptPostpone, map[string]interface{}{"postpone_for_sec": float64(60), "postponed_until": float64(1)}},
		"no deadline":           {model.EmailAttemptPostpone, nil},
		"no renewal":            {model.EmailAttemptRenew, nil},
		"return without reason": {model.EmailAttemptReturn, map[string]interface{}{"reason": "   "}},
		"two targets":           {model.EmailAttemptTransfer, map[string]interface{}{"queue_id": float64(5), "agent_id": float64(4)}},
		"no target":             {model.EmailAttemptTransfer, nil},
		"long comment":          {model.EmailAttemptTransfer, map[string]interface{}{"queue_id": float64(5), "comment": strings.Repeat("ї", 501)}},
		"queue beyond int32":    {model.EmailAttemptTransfer, map[string]interface{}{"queue_id": float64(1 << 32)}},
	} {
		if _, err := emailCommandFromRequest(emailRequest(tc.extra), tc.op); err == nil {
			t.Errorf("%s: accepted", name)
		}
	}
}
