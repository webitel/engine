package wsapi

import (
	"context"
	"math"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/webitel/engine/app"
	"github.com/webitel/engine/model"
)

const maxEmailCommentLen = 500

var emailCommandIdRe = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

// InitEmail registers the Email attempt commands of an operator.
func (api *API) InitEmail() {
	for _, op := range []string{model.EmailAttemptTake, model.EmailAttemptPostpone, model.EmailAttemptResume,
		model.EmailAttemptReturn, model.EmailAttemptComplete, model.EmailAttemptRenew, model.EmailAttemptTransfer} {
		op := op
		api.Router.Handle("cc_email_"+op, api.ApiWebSocketHandler(func(ctx context.Context, conn *app.WebConn, req *model.WebSocketRequest) (map[string]interface{}, model.AppError) {
			return api.emailAttemptCommand(ctx, conn, req, op)
		}))
	}
	api.Router.Handle("cc_email_transfer_agents", api.ApiWebSocketHandler(api.emailTransferAgents))
}

// The handlers check req.Session, read for this request, never the session cached in the connection.
func (api *API) emailAttemptCommand(ctx context.Context, _ *app.WebConn, req *model.WebSocketRequest, op string) (map[string]interface{}, model.AppError) {
	cmd, err := emailCommandFromRequest(req, op)
	if err != nil {
		return nil, err
	}

	res, err := api.ctrl.EmailAttemptCommand(ctx, &req.Session, cmd)
	if err != nil {
		return nil, err
	}
	if res.TransferredTo != nil {
		return map[string]interface{}{"email": res.Email.ToMap(), "transferred_to": res.TransferredTo.ToMap()}, nil
	}

	return res.Email.ToMap(), nil
}

func (api *API) emailTransferAgents(ctx context.Context, _ *app.WebConn, req *model.WebSocketRequest) (map[string]interface{}, model.AppError) {
	attemptId, ok := emailPositive(req.Data, "attempt_id", math.MaxInt64)
	if !ok {
		return nil, NewInvalidWebSocketParamError(req.Action, "attempt_id")
	}

	agents, err := api.ctrl.EmailTransferCandidates(ctx, &req.Session, attemptId)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{"items": agents}, nil
}

// emailCommandFromRequest validates the parameters of the client; the domain, the user and the agent
// are never taken from it.
func emailCommandFromRequest(req *model.WebSocketRequest, op string) (*model.EmailAttemptCommand, model.AppError) {
	d := req.Data
	invalid := func(name string) (*model.EmailAttemptCommand, model.AppError) {
		return nil, NewInvalidWebSocketParamError(req.Action, name)
	}

	cmd := &model.EmailAttemptCommand{Op: op}
	var ok bool
	if cmd.ThreadId, ok = emailPositive(d, "thread_id", math.MaxInt64); !ok {
		return invalid("thread_id")
	}
	if cmd.AttemptId, ok = emailPositive(d, "attempt_id", math.MaxInt64); !ok {
		return invalid("attempt_id")
	}
	if cmd.ExpectedVersion, ok = emailInteger(d, "expected_version", 0, math.MaxInt64); !ok {
		return invalid("expected_version")
	}
	if cmd.CommandId, ok = d["command_id"].(string); !ok || !emailCommandIdRe.MatchString(cmd.CommandId) {
		return invalid("command_id")
	}

	switch op {
	case model.EmailAttemptPostpone:
		until, hasUntil := emailPositive(d, "postponed_until", math.MaxInt64)
		forSec, hasFor := emailPositive(d, "postpone_for_sec", math.MaxInt32)
		if hasUntil == hasFor || (hasUntil && d["postpone_for_sec"] != nil) || (hasFor && d["postponed_until"] != nil) {
			return invalid("postponed_until")
		}
		cmd.PostponedUntil, cmd.PostponeForSec = until, int32(forSec)
	case model.EmailAttemptRenew:
		sec, ok := emailPositive(d, "renewal_sec", math.MaxInt32)
		if !ok {
			return invalid("renewal_sec")
		}
		cmd.RenewalSec = uint32(sec)
	case model.EmailAttemptReturn:
		if cmd.Reason, ok = emailText(d, "reason", true); !ok {
			return invalid("reason")
		}
	case model.EmailAttemptTransfer:
		queueId, hasQueue := emailPositive(d, "queue_id", math.MaxInt32)
		agentId, hasAgent := emailPositive(d, "agent_id", math.MaxInt32)
		if hasQueue == hasAgent || (hasQueue && d["agent_id"] != nil) || (hasAgent && d["queue_id"] != nil) {
			return invalid("queue_id")
		}
		cmd.ToQueueId, cmd.ToAgentId = int32(queueId), int32(agentId)
		if cmd.Comment, ok = emailText(d, "comment", false); !ok {
			return invalid("comment")
		}
	}

	return cmd, nil
}

// emailInteger reads a whole JSON number within [min, max]; JSON numbers are exact up to 2^53.
func emailInteger(d map[string]interface{}, name string, min, max int64) (int64, bool) {
	v, ok := d[name].(float64)
	if !ok || v != math.Trunc(v) || math.Abs(v) > 1<<53 || v < float64(min) || v > float64(max) {
		return 0, false
	}

	return int64(v), true
}

func emailPositive(d map[string]interface{}, name string, max int64) (int64, bool) {
	return emailInteger(d, name, 1, max)
}

// emailText trims a reason or a comment: up to maxEmailCommentLen characters, present when required.
func emailText(d map[string]interface{}, name string, required bool) (string, bool) {
	raw, present := d[name]
	if !present || raw == nil {
		return "", !required
	}
	s, ok := raw.(string)
	if !ok {
		return "", false
	}
	s = strings.TrimSpace(s)
	n := utf8.RuneCountInString(s)

	return s, n <= maxEmailCommentLen && (n > 0 || !required)
}
