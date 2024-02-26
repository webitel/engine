package sqlstore

import (
	"context"
	"fmt"

	"github.com/lib/pq"
	"github.com/webitel/engine/model"
	"github.com/webitel/engine/store"
)

type SqlAgentHookStore struct {
	SqlStore
}

func NewSqlAgentHookStore(sqlStore SqlStore) store.AgentHookStore {
	us := &SqlAgentHookStore{sqlStore}
	return us
}

func (s SqlAgentHookStore) Create(ctx context.Context, domainId int64, agentId int64, in *model.AgentHook) (*model.AgentHook, model.AppError) {
	var qh *model.AgentHook

	err := s.GetMaster().WithContext(ctx).SelectOne(&qh, `with qe as (
    insert into call_center.cc_agent_events (schema_id, event, agent_id, enabled, updated_by, updated_at)
    select :SchemaId, :Event, :AgentId, :Enabled, :UpdatedBy, :UpdatedAt
    where exists (select 1 from call_center.cc_agent a where a.domain_id = :DomainId and a.id = :AgentId)
    returning *
)
select qe.id,
       call_center.cc_get_lookup(qe.schema_id, s.name) "schema",
       qe.event,
       qe.enabled
from qe
    left join flow.acr_routing_scheme s on s.id = qe.schema_id`, map[string]interface{}{
		"DomainId":  domainId,
		"SchemaId":  in.Schema.Id,
		"Event":     in.Event,
		"AgentId":   agentId,
		"Enabled":   in.Enabled,
		"UpdatedBy": in.UpdatedBy.GetSafeId(),
		"UpdatedAt": in.UpdatedAt,
	})

	if err != nil {
		return nil, model.NewCustomCodeError("store.sql_agent_hook.create.app_error", fmt.Sprintf("event=%v, %v", in.Event, err.Error()), extractCodeFromErr(err))
	}

	return qh, nil
}

func (s SqlAgentHookStore) Get(ctx context.Context, domainId int64, agentId int64, id int32) (*model.AgentHook, model.AppError) {
	var qh *model.AgentHook

	err := s.GetReplica().WithContext(ctx).SelectOne(&qh, `select
    id,
    schema,
    event,
    enabled
from call_center.cc_agent_events_list qe
where qe.agent_id = :AgentId
     and qe.id = :Id
     and exists (select 1 from call_center.cc_agent q where q.id = qe.agent_id and q.domain_id = :DomainId)`, map[string]interface{}{
		"AgentId":  agentId,
		"Id":       id,
		"DomainId": domainId,
	})

	if err != nil {
		return nil, model.NewCustomCodeError("store.sql_agent_hook.get.app_error", fmt.Sprintf("Id=%v, %v", id, err.Error()), extractCodeFromErr(err))
	}

	return qh, nil
}

func (s SqlAgentHookStore) GetAllPage(ctx context.Context, domainId int64, agentId int64, search *model.SearchAgentHook) ([]*model.AgentHook, model.AppError) {
	var list []*model.AgentHook

	f := map[string]interface{}{
		"DomainId":  domainId,
		"AgentId":   agentId,
		"Q":         search.GetQ(),
		"Ids":       pq.Array(search.Ids),
		"SchemaIds": pq.Array(search.SchemaIds),
		"Events":    pq.Array(search.Events),
	}

	err := s.ListQuery(ctx, &list, search.ListRequest,
		` agent_id = :AgentId::int
                and exists (select 1 from call_center.cc_agent q where q.id = agent_id and q.domain_id = :DomainId)
				and (:Q::text isnull or ( "event" ilike :Q::varchar ))
				and (:Ids::int4[] isnull or id = any(:Ids))
				and (:SchemaIds::int4[] isnull or schema_id = any(:SchemaIds))
				and (:Events::varchar[] isnull or "event" = any(:Events))
			`,
		model.AgentHook{}, f)
	if err != nil {
		return nil, model.NewCustomCodeError("store.sql_agent_hook.get_all.app_error", err.Error(), extractCodeFromErr(err))
	}

	return list, nil
}

func (s SqlAgentHookStore) Update(ctx context.Context, domainId int64, agentId int64, qh *model.AgentHook) (*model.AgentHook, model.AppError) {

	err := s.GetMaster().WithContext(ctx).SelectOne(&qh, `with qe as (
    update call_center.cc_agent_events
    set schema_id = :SchemaId,
        event = :Event,
        enabled = :Enabled,
        updated_by = :UpdatedBy,
        updated_at = :UpdatedAt
    where id = :Id
		and agent_id = :AgentId
        and exists(select 1 from call_center.cc_agent q where q.id = agent_id and q.domain_id = :DomainId)
    returning *
)
select qe.id,
       call_center.cc_get_lookup(qe.schema_id, s.name) "schema",
       qe.event,
       qe.enabled
from qe
    left join flow.acr_routing_scheme s on s.id = qe.schema_id`, map[string]interface{}{
		"Id":        qh.Id,
		"SchemaId":  qh.Schema.Id,
		"Event":     qh.Event,
		"Enabled":   qh.Enabled,
		"UpdatedBy": qh.UpdatedBy.GetSafeId(),
		"UpdatedAt": qh.UpdatedAt,
		"AgentId":   agentId,
		"DomainId":  domainId,
	})

	if err != nil {
		return nil, model.NewCustomCodeError("store.sql_agent_hook.update.app_error", err.Error(), extractCodeFromErr(err))
	}

	return qh, nil
}

func (s SqlAgentHookStore) Delete(ctx context.Context, domainId int64, agentId int64, id int32) model.AppError {
	if _, err := s.GetMaster().WithContext(ctx).Exec(`delete from call_center.cc_agent_events qe where qe.id=:Id and qe.agent_id = :AgentId 
			and exists(select 1 from call_center.cc_agent q where q.id = :AgentId and q.domain_id = :DomainId)`,
		map[string]interface{}{"Id": id, "DomainId": domainId, "AgentId": agentId}); err != nil {
		return model.NewCustomCodeError("store.sql_agent_hook.delete.app_error", fmt.Sprintf("Id=%v, %s", id, err.Error()), extractCodeFromErr(err))
	}
	return nil
}
