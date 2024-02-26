package sqlstore

import (
	"context"
	"fmt"

	"github.com/lib/pq"
	"github.com/webitel/engine/model"
	"github.com/webitel/engine/store"
)

type SqlTeamHookStore struct {
	SqlStore
}

func NewSqlTeamHookStore(sqlStore SqlStore) store.TeamHookStore {
	us := &SqlTeamHookStore{sqlStore}
	return us
}

func (s SqlTeamHookStore) Create(ctx context.Context, domainId int64, teamId int64, in *model.TeamHook) (*model.TeamHook, model.AppError) {
	var qh *model.TeamHook

	err := s.GetMaster().WithContext(ctx).SelectOne(&qh, `with qe as (
    insert into call_center.cc_team_events (schema_id, event, team_id, enabled, updated_by, updated_at)
    select :SchemaId, :Event, :TeamId, :Enabled, :UpdatedBy, :UpdatedAt
    where exists (select 1 from call_center.cc_team a where a.domain_id = :DomainId and a.id = :TeamId)
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
		"TeamId":    teamId,
		"Enabled":   in.Enabled,
		"UpdatedBy": in.UpdatedBy.GetSafeId(),
		"UpdatedAt": in.UpdatedAt,
	})

	if err != nil {
		return nil, model.NewCustomCodeError("store.sql_team.create.app_error", fmt.Sprintf("event=%v, %v", in.Event, err.Error()), extractCodeFromErr(err))
	}

	return qh, nil
}

func (s SqlTeamHookStore) Get(ctx context.Context, domainId int64, teamId int64, id uint32) (*model.TeamHook, model.AppError) {
	var qh *model.TeamHook

	err := s.GetReplica().WithContext(ctx).SelectOne(&qh, `select
    id,
    schema,
    event,
    enabled
from call_center.cc_team_events_list qe
where qe.team_id = :TeamId
     and qe.id = :Id
     and exists (select 1 from call_center.cc_team q where q.id = qe.team_id and q.domain_id = :DomainId)`, map[string]interface{}{
		"TeamId":   teamId,
		"Id":       id,
		"DomainId": domainId,
	})

	if err != nil {
		return nil, model.NewCustomCodeError("store.sql_team_hook.get.app_error", fmt.Sprintf("Id=%v, %v", id, err.Error()), extractCodeFromErr(err))
	}

	return qh, nil
}

func (s SqlTeamHookStore) GetAllPage(ctx context.Context, domainId int64, teamId int64, search *model.SearchTeamHook) ([]*model.TeamHook, model.AppError) {
	var list []*model.TeamHook

	f := map[string]interface{}{
		"DomainId":  domainId,
		"TeamId":    teamId,
		"Q":         search.GetQ(),
		"Ids":       pq.Array(search.Ids),
		"SchemaIds": pq.Array(search.SchemaIds),
		"Events":    pq.Array(search.Events),
	}

	err := s.ListQuery(ctx, &list, search.ListRequest,
		` team_id = :TeamId::int8
                and exists (select 1 from call_center.cc_team q where q.id = team_id and q.domain_id = :DomainId)
				and (:Q::text isnull or ( "event" ilike :Q::varchar ))
				and (:Ids::int4[] isnull or id = any(:Ids))
				and (:SchemaIds::int4[] isnull or schema_id = any(:SchemaIds))
				and (:Events::varchar[] isnull or "event" = any(:Events))
			`,
		model.TeamHook{}, f)
	if err != nil {
		return nil, model.NewCustomCodeError("store.sql_team_hook.get_all.app_error", err.Error(), extractCodeFromErr(err))
	}

	return list, nil
}

func (s SqlTeamHookStore) Update(ctx context.Context, domainId int64, teamId int64, qh *model.TeamHook) (*model.TeamHook, model.AppError) {

	err := s.GetMaster().WithContext(ctx).SelectOne(&qh, `with qe as (
    update call_center.cc_team_events
    set schema_id = :SchemaId,
        event = :Event,
        enabled = :Enabled,
        updated_by = :UpdatedBy,
        updated_at = :UpdatedAt
    where id = :Id
		and team_id = :TeamId
        and exists(select 1 from call_center.cc_team q where q.id = team_id and q.domain_id = :DomainId)
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
		"TeamId":    teamId,
		"DomainId":  domainId,
	})

	if err != nil {
		return nil, model.NewCustomCodeError("store.sql_team_hook.update.app_error", err.Error(), extractCodeFromErr(err))
	}

	return qh, nil
}

func (s SqlTeamHookStore) Delete(ctx context.Context, domainId int64, teamId int64, id uint32) model.AppError {
	if _, err := s.GetMaster().WithContext(ctx).Exec(`delete from call_center.cc_team_events qe where qe.id=:Id and qe.team_id = :TeamId 
			and exists(select 1 from call_center.cc_team q where q.id = :TeamId and q.domain_id = :DomainId)`,
		map[string]interface{}{"Id": id, "DomainId": domainId, "TeamId": teamId}); err != nil {
		return model.NewCustomCodeError("store.sql_team_hook.delete.app_error", fmt.Sprintf("Id=%v, %s", id, err.Error()), extractCodeFromErr(err))
	}
	return nil
}
