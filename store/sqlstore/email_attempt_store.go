package sqlstore

import (
	"context"
	"net/http"

	"github.com/lib/pq"

	"github.com/webitel/engine/model"
)

// EmailAgentIdByUser finds the agent of the user without its channels; nil means none.
func (s SqlAgentStore) EmailAgentIdByUser(ctx context.Context, domainId, userId int64) (*int32, model.AppError) {
	id, err := s.GetMaster().WithContext(ctx).SelectNullInt(`select a.id
from call_center.cc_agent a
where a.domain_id = :DomainId::int8 and a.user_id = :UserId::int8`, map[string]any{
		"DomainId": domainId,
		"UserId":   userId,
	})
	if err != nil {
		return nil, model.NewCustomCodeError("store.sql_agent.email_agent_by_user.app_error", err.Error(), extractCodeFromErr(err))
	}
	if !id.Valid {
		return nil, nil
	}
	agentId := int32(id.Int64)

	return &agentId, nil
}

// EmailTransferCandidates reads cc_email_transfer_candidates of call_center, which keeps its rules
// in one place; the transfer checks them again under its locks.
func (s SqlMemberStore) EmailTransferCandidates(ctx context.Context, domainId, attemptId int64, agentId int32) ([]*model.EmailTransferCandidate, model.AppError) {
	var res []*model.EmailTransferCandidate
	_, err := s.GetMaster().WithContext(ctx).Select(&res, `select c.agent_id, c.user_id,
    coalesce(nullif(u.name, ''), u.username::text, '') as name, c.online, c.slot_type, c.available
from call_center.cc_email_transfer_candidates(:DomainId::int8, :AttemptId::int8, :AgentId::int4) c
    left join directory.wbt_user u on u.id = c.user_id and u.dc = :DomainId::int8
order by c.available desc, name, c.agent_id`, map[string]any{
		"DomainId":  domainId,
		"AttemptId": attemptId,
		"AgentId":   agentId,
	})
	if err != nil {
		if e, ok := err.(*pq.Error); ok {
			switch e.Code {
			case "EMANF":
				return nil, model.NewNotFoundError("cc.email_attempt.not_found", "active attempt not found")
			case "EMAOW":
				return nil, model.NewBadRequestError("cc.email_attempt.owner_mismatch", "the attempt is not assigned to the agent")
			}
		}
		return nil, model.NewCustomCodeError("store.sql_member.email_transfer_candidates.app_error", err.Error(), http.StatusInternalServerError)
	}

	return res, nil
}
