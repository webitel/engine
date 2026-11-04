package app

import (
	"context"
	"net/http"
	"time"

	"github.com/webitel/engine/gen/cc"
	"github.com/webitel/engine/model"
)

const defaultEmailCommandTimeout = 10 * time.Second

// EmailOwnAgent is the agent of the session user; commands act only for it.
func (app *App) EmailOwnAgent(ctx context.Context, domainId, userId int64) (int32, model.AppError) {
	agentId, err := app.Store.Agent().EmailAgentIdByUser(ctx, domainId, userId)
	if err != nil {
		return 0, err
	}
	if agentId == nil {
		return 0, model.NewCustomCodeError("engine.email_attempt.agent_not_found", "the user has no agent", http.StatusBadRequest)
	}

	return *agentId, nil
}

// EmailTransferCandidates lists the agents a direct transfer of the attempt of agentId may offer it to.
func (app *App) EmailTransferCandidates(ctx context.Context, domainId, attemptId int64, agentId int32) ([]*model.EmailTransferCandidate, model.AppError) {
	return app.Store.Member().EmailTransferCandidates(ctx, domainId, attemptId, agentId)
}

// EmailAttemptCommand sends the command of the operator to call_center, which checks the owner,
// version and repeat; userId and agentId come from the session.
func (app *App) EmailAttemptCommand(ctx context.Context, domainId, userId int64, agentId int32, cmd *model.EmailAttemptCommand) (*model.EmailAttemptResult, model.AppError) {
	timeout := defaultEmailCommandTimeout
	if sec := app.Config().EmailCommandTimeout; sec > 0 {
		timeout = time.Duration(sec) * time.Second
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	version := cmd.ExpectedVersion
	op := &cc.EmailAttemptOperator{AttemptId: cmd.AttemptId, AgentId: agentId, ExpectedVersion: &version,
		CommandId: cmd.CommandId, UserId: userId}
	member := app.cc.Member()

	var res *cc.EmailAttemptResponse
	var err error
	switch cmd.Op {
	case model.EmailAttemptTake:
		res, err = member.EmailAttemptTake(ctx, &cc.EmailAttemptTakeRequest{DomainId: domainId, ThreadId: cmd.ThreadId, Operator: op})
	case model.EmailAttemptPostpone:
		in := &cc.EmailAttemptPostponeRequest{DomainId: domainId, ThreadId: cmd.ThreadId,
			Caller: &cc.EmailAttemptPostponeRequest_Operator{Operator: op}}
		if cmd.PostponedUntil > 0 {
			in.Deadline = &cc.EmailAttemptPostponeRequest_PostponedUntil{PostponedUntil: cmd.PostponedUntil}
		} else {
			in.Deadline = &cc.EmailAttemptPostponeRequest_PostponeForSec{PostponeForSec: cmd.PostponeForSec}
		}
		res, err = member.EmailAttemptPostpone(ctx, in)
	case model.EmailAttemptReturn:
		res, err = member.EmailAttemptReturn(ctx, &cc.EmailAttemptReturnRequest{DomainId: domainId, ThreadId: cmd.ThreadId,
			Operator: op, Reason: cmd.Reason})
	case model.EmailAttemptComplete:
		res, err = member.EmailAttemptComplete(ctx, &cc.EmailAttemptCompleteRequest{DomainId: domainId, ThreadId: cmd.ThreadId, Operator: op})
	case model.EmailAttemptResume:
		var out *cc.ResumeAttemptResponse
		out, err = member.EmailAttemptResume(ctx, &cc.ResumeAttemptRequest{DomainId: domainId, AttemptId: cmd.AttemptId,
			ThreadId: cmd.ThreadId, AgentId: agentId, ExpectedVersion: &version, CommandId: cmd.CommandId, UserId: userId})
		res = out.GetEmail()
	case model.EmailAttemptRenew:
		var out *cc.AttemptRenewalResultResponse
		out, err = member.EmailAttemptRenew(ctx, &cc.AttemptRenewalResultRequest{DomainId: domainId, AttemptId: cmd.AttemptId,
			Renewal: cmd.RenewalSec, ThreadId: cmd.ThreadId, AgentId: agentId, ExpectedVersion: &version,
			CommandId: cmd.CommandId, UserId: userId})
		res = out.GetEmail()
	case model.EmailAttemptTransfer:
		var out *cc.TransferResponse
		out, err = member.EmailAttemptTransfer(ctx, &cc.TransferRequest{DomainId: domainId, AttemptId: cmd.AttemptId,
			AgentId: cmd.ToAgentId, QueueId: cmd.ToQueueId, ThreadId: cmd.ThreadId, Operator: op, Comment: cmd.Comment})
		if err == nil {
			if out.GetEmail() == nil || out.GetTransferredTo() == nil {
				return nil, model.NewInternalError("engine.email_attempt.internal", "call_center returned an incomplete transfer")
			}
			return &model.EmailAttemptResult{Email: emailAttemptState(out.GetEmail()),
				TransferredTo: emailAttemptState(out.GetTransferredTo())}, nil
		}
	default:
		return nil, model.NewBadRequestError("engine.email_attempt.invalid_argument", "unknown operation")
	}
	if err != nil {
		return nil, emailGrpcError(ctx, err)
	}
	if res == nil {
		return nil, model.NewInternalError("engine.email_attempt.internal", "call_center returned no Email attempt")
	}

	return &model.EmailAttemptResult{Email: emailAttemptState(res)}, nil
}

func emailAttemptState(r *cc.EmailAttemptResponse) *model.EmailAttemptState {
	return &model.EmailAttemptState{
		AttemptId:         r.GetAttemptId(),
		ThreadId:          r.GetThreadId(),
		State:             r.GetState(),
		AttemptVersion:    r.GetAttemptVersion(),
		ProjectionVersion: r.GetProjectionVersion(),
		Result:            r.GetResult(),
		QueueId:           r.GetQueueId(),
		AgentId:           r.GetAgentId(),
		PostponedUntil:    r.GetPostponedUntil(),
		Replayed:          r.GetReplayed(),
	}
}
