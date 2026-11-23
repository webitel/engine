package app

import (
	"context"
	"net/http"
	"time"

	ccclient "github.com/webitel/engine/app/cc"
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

// emailOperation is the session and operator data shared by the Email attempt commands.
type emailOperation struct {
	member   ccclient.MemberApi
	domainId int64
	userId   int64
	agentId  int32
	operator *cc.EmailAttemptOperator
	cmd      *model.EmailAttemptCommand
}

// emailOperationHandler sends one Email attempt command to its call_center RPC.
type emailOperationHandler func(ctx context.Context, o *emailOperation) (*model.EmailAttemptResult, model.AppError)

var emailOperationHandlers = map[string]emailOperationHandler{
	model.EmailAttemptTake:     emailAttemptTake,
	model.EmailAttemptPostpone: emailAttemptPostpone,
	model.EmailAttemptReturn:   emailAttemptReturn,
	model.EmailAttemptComplete: emailAttemptComplete,
	model.EmailAttemptResume:   emailAttemptResume,
	model.EmailAttemptRenew:    emailAttemptRenew,
	model.EmailAttemptTransfer: emailAttemptTransfer,
}

// EmailAttemptCommand sends the command of the operator to call_center, which checks the owner,
// version and repeat; userId and agentId come from the session.
func (app *App) EmailAttemptCommand(ctx context.Context, domainId, userId int64, agentId int32, cmd *model.EmailAttemptCommand) (*model.EmailAttemptResult, model.AppError) {
	handler, ok := emailOperationHandlers[cmd.Op]
	if !ok {
		return nil, model.NewBadRequestError("engine.email_attempt.invalid_argument", "unknown operation")
	}

	timeout := defaultEmailCommandTimeout
	if sec := app.Config().EmailSettings.EmailCommandTimeout; sec > 0 {
		timeout = time.Duration(sec) * time.Second
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return handler(ctx, &emailOperation{
		member:   app.cc.Member(),
		domainId: domainId,
		userId:   userId,
		agentId:  agentId,
		operator: &cc.EmailAttemptOperator{AttemptId: cmd.AttemptId, AgentId: agentId,
			ExpectedVersion: model.NewInt64(cmd.ExpectedVersion), CommandId: cmd.CommandId, UserId: userId},
		cmd: cmd,
	})
}

func emailAttemptTake(ctx context.Context, o *emailOperation) (*model.EmailAttemptResult, model.AppError) {
	res, err := o.member.EmailAttemptTake(ctx, &cc.EmailAttemptTakeRequest{DomainId: o.domainId, ThreadId: o.cmd.ThreadId,
		Operator: o.operator})
	return emailAttemptResult(res, err)
}

func emailAttemptPostpone(ctx context.Context, o *emailOperation) (*model.EmailAttemptResult, model.AppError) {
	in := &cc.EmailAttemptPostponeRequest{DomainId: o.domainId, ThreadId: o.cmd.ThreadId,
		Caller: &cc.EmailAttemptPostponeRequest_Operator{Operator: o.operator}}
	if o.cmd.PostponedUntil > 0 {
		in.Deadline = &cc.EmailAttemptPostponeRequest_PostponedUntil{PostponedUntil: o.cmd.PostponedUntil}
	} else {
		in.Deadline = &cc.EmailAttemptPostponeRequest_PostponeForSec{PostponeForSec: o.cmd.PostponeForSec}
	}
	res, err := o.member.EmailAttemptPostpone(ctx, in)
	return emailAttemptResult(res, err)
}

func emailAttemptReturn(ctx context.Context, o *emailOperation) (*model.EmailAttemptResult, model.AppError) {
	res, err := o.member.EmailAttemptReturn(ctx, &cc.EmailAttemptReturnRequest{DomainId: o.domainId, ThreadId: o.cmd.ThreadId,
		Operator: o.operator, Reason: o.cmd.Reason})
	return emailAttemptResult(res, err)
}

func emailAttemptComplete(ctx context.Context, o *emailOperation) (*model.EmailAttemptResult, model.AppError) {
	res, err := o.member.EmailAttemptComplete(ctx, &cc.EmailAttemptCompleteRequest{DomainId: o.domainId, ThreadId: o.cmd.ThreadId,
		Operator: o.operator})
	return emailAttemptResult(res, err)
}

func emailAttemptResume(ctx context.Context, o *emailOperation) (*model.EmailAttemptResult, model.AppError) {
	out, err := o.member.EmailAttemptResume(ctx, &cc.ResumeAttemptRequest{DomainId: o.domainId, AttemptId: o.cmd.AttemptId,
		ThreadId: o.cmd.ThreadId, AgentId: o.agentId, ExpectedVersion: model.NewInt64(o.cmd.ExpectedVersion),
		CommandId: o.cmd.CommandId, UserId: o.userId})
	return emailAttemptResult(out.GetEmail(), err)
}

func emailAttemptRenew(ctx context.Context, o *emailOperation) (*model.EmailAttemptResult, model.AppError) {
	out, err := o.member.EmailAttemptRenew(ctx, &cc.AttemptRenewalResultRequest{DomainId: o.domainId, AttemptId: o.cmd.AttemptId,
		Renewal: o.cmd.RenewalSec, ThreadId: o.cmd.ThreadId, AgentId: o.agentId, ExpectedVersion: model.NewInt64(o.cmd.ExpectedVersion),
		CommandId: o.cmd.CommandId, UserId: o.userId})
	return emailAttemptResult(out.GetEmail(), err)
}

// emailAttemptTransfer returns both the transferred attempt and the attempt it ended up in.
func emailAttemptTransfer(ctx context.Context, o *emailOperation) (*model.EmailAttemptResult, model.AppError) {
	out, err := o.member.EmailAttemptTransfer(ctx, &cc.TransferRequest{DomainId: o.domainId, AttemptId: o.cmd.AttemptId,
		AgentId: o.cmd.ToAgentId, QueueId: o.cmd.ToQueueId, ThreadId: o.cmd.ThreadId, Operator: o.operator, Comment: o.cmd.Comment})
	if err != nil {
		return nil, err
	}
	if out.GetEmail() == nil || out.GetTransferredTo() == nil {
		return nil, model.NewInternalError("engine.email_attempt.internal", "call_center returned an incomplete transfer")
	}

	return &model.EmailAttemptResult{Email: emailAttemptState(out.GetEmail()),
		TransferredTo: emailAttemptState(out.GetTransferredTo())}, nil
}

// emailAttemptResult is the result of a command that returns one Email attempt.
func emailAttemptResult(res *cc.EmailAttemptResponse, err model.AppError) (*model.EmailAttemptResult, model.AppError) {
	if err != nil {
		return nil, err
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
