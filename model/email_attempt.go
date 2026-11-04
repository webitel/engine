package model

// PERMISSION_SCOPE_EMAIL_THREAD is the object class of Email Threads; commands need update.
const PERMISSION_SCOPE_EMAIL_THREAD = "email_thread"

// Operations of the Email attempt commands of an operator.
const (
	EmailAttemptTake     = "take"
	EmailAttemptPostpone = "postpone"
	EmailAttemptResume   = "resume"
	EmailAttemptReturn   = "return"
	EmailAttemptComplete = "complete"
	EmailAttemptRenew    = "renew"
	EmailAttemptTransfer = "transfer"
)

// EmailAttemptCommand is a command of the operator; the domain, the user and the agent come from the session.
type EmailAttemptCommand struct {
	Op              string
	ThreadId        int64
	AttemptId       int64
	ExpectedVersion int64
	CommandId       string
	// Postpone: one of them; Unix milliseconds or seconds.
	PostponedUntil int64
	PostponeForSec int32
	RenewalSec     uint32
	Reason         string
	// Transfer: one of them.
	ToQueueId int32
	ToAgentId int32
	Comment   string
}

// EmailAttemptState is an attempt after the command, or after its first execution when Replayed.
type EmailAttemptState struct {
	AttemptId         int64
	ThreadId          int64
	State             string
	AttemptVersion    int64
	ProjectionVersion int64
	Result            string
	QueueId           int32
	AgentId           int32
	PostponedUntil    int64
	Replayed          bool
}

// EmailAttemptResult is one attempt, or for a transfer the transferred one and the new one.
type EmailAttemptResult struct {
	Email         *EmailAttemptState
	TransferredTo *EmailAttemptState
}

// EmailTransferCandidate is an agent a direct transfer may offer the email to; Available repeats the
// check of call_center without reserving a slot.
type EmailTransferCandidate struct {
	AgentId   int32   `json:"agent_id" db:"agent_id"`
	UserId    int64   `json:"user_id" db:"user_id"`
	Name      string  `json:"name" db:"name"`
	Online    bool    `json:"online" db:"online"`
	SlotType  *string `json:"slot_type" db:"slot_type"`
	Available bool    `json:"available" db:"available"`
}

func (s *EmailAttemptState) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"attempt_id":         s.AttemptId,
		"thread_id":          s.ThreadId,
		"state":              s.State,
		"attempt_version":    s.AttemptVersion,
		"projection_version": s.ProjectionVersion,
		"result":             s.Result,
		"queue_id":           s.QueueId,
		"postponed_until":    s.PostponedUntil,
		"replayed":           s.Replayed,
	}
}
