package app

import (
	"context"
	"errors"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestEmailGrpcError(t *testing.T) {
	cases := []struct {
		err  error
		id   string
		code int
	}{
		{status.Error(codes.ResourceExhausted, "cc.email_attempt.slot_exhausted"), "cc.email_attempt.slot_exhausted", 429},
		{status.Error(codes.FailedPrecondition, "cc.email_attempt.agent_ineligible"), "cc.email_attempt.agent_ineligible", 400},
		{status.Error(codes.Aborted, "cc.email_attempt.version_conflict"), "cc.email_attempt.version_conflict", 409},
		{status.Error(codes.AlreadyExists, "cc.email_attempt.command_id_conflict"), "cc.email_attempt.command_id_conflict", 409},
		{status.Error(codes.NotFound, "cc.email_attempt.not_found"), "cc.email_attempt.not_found", 404},
		{status.Error(codes.Unavailable, "cc.email_attempt.lock_timeout"), "cc.email_attempt.lock_timeout", 503},
		{status.Error(codes.InvalidArgument, "postponed_until must be positive Unix milliseconds"), "engine.email_attempt.invalid_argument", 400},
		{status.Error(codes.DataLoss, "x"), "engine.email_attempt.internal", 500},
		{errors.New("plain"), "engine.email_attempt.internal", 500},
	}
	for _, tc := range cases {
		got := emailGrpcError(context.Background(), tc.err)
		if got.GetId() != tc.id || got.GetStatusCode() != tc.code {
			t.Errorf("%v: got %s %d, want %s %d", tc.err, got.GetId(), got.GetStatusCode(), tc.id, tc.code)
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if got := emailGrpcError(ctx, status.Error(codes.Canceled, "context canceled")); got.GetStatusCode() != 499 ||
		got.GetId() != "engine.email_attempt.canceled" {
		t.Errorf("canceled: %s %d", got.GetId(), got.GetStatusCode())
	}
}
