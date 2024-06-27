package sqlstore

import (
	"context"
	"fmt"

	"github.com/lib/pq"
	"github.com/webitel/engine/model"
	"github.com/webitel/engine/store"
)

type SqlTeamTriggerStore struct {
	SqlStore
}

func NewSqlTeamTriggerStore(sqlStore SqlStore) store.TeamTriggerStore {
	us := &SqlTeamTriggerStore{sqlStore}
	return us
}

func (s SqlTeamTriggerStore) Create(ctx context.Context, domainId int64, teamId int64, in *model.TeamTrigger) (*model.TeamTrigger, model.AppError) {
	var qt *model.TeamTrigger

	err := s.GetMaster().WithContext(ctx).SelectOne(&qt, `with qt as (
    insert into call_center.cc_team_trigger (name, description, schema_id, team_id, enabled, updated_by, updated_at, created_by, created_at)
    select :Name, :Description, :SchemaId, :TeamId, :Enabled, :UpdatedBy, :UpdatedAt, :CreatedBy, :CreatedAt
    where exists (select 1 from call_center.cc_team a where a.domain_id = :DomainId and a.id = :TeamId)
    returning *
)select qt.id,
       call_center.cc_get_lookup(qt.schema_id, s.name) "schema",
       qt.name,
       qt.description,
       qt.enabled,
       call_center.cc_get_lookup(uc.id, coalesce(uc.name, uc.username)) "created_by",
       qt.created_at,
       call_center.cc_get_lookup(uu.id, coalesce(uu.name, uu.username)) "updated_by"
from qt
    left join flow.acr_routing_scheme s on s.id = qt.schema_id
    left join directory.wbt_user uc on uc.id = qt.created_by
    left join directory.wbt_user uu on uu.id = qt.updated_by`, map[string]interface{}{
		"DomainId":    domainId,
		"SchemaId":    in.Schema.GetSafeId(),
		"TeamId":      teamId,
		"Enabled":     in.Enabled,
		"Name":        in.Name,
		"Description": in.Description,
		"UpdatedBy":   in.UpdatedBy.GetSafeId(),
		"UpdatedAt":   in.UpdatedAt,
		"CreatedBy":   in.UpdatedBy.GetSafeId(),
		"CreatedAt":   in.UpdatedAt,
	})

	if err != nil {
		return nil, model.NewCustomCodeError("store.sql_team_trigger.create.app_error", fmt.Sprintf("trigger=%v, %v", in.Name, messageFromErr(err)), extractCodeFromErr(err))
	}

	return qt, nil
}

func (s SqlTeamTriggerStore) Get(ctx context.Context, domainId int64, teamId int64, id uint32) (*model.TeamTrigger, model.AppError) {
	var qt *model.TeamTrigger

	err := s.GetReplica().WithContext(ctx).SelectOne(&qt, `select qt.id,
       call_center.cc_get_lookup(qt.schema_id, s.name) "schema",
       qt.name,
       qt.description,
       qt.enabled,
       call_center.cc_get_lookup(uc.id, coalesce(uc.name, uc.username)) "created_by",
       qt.created_at,
       call_center.cc_get_lookup(uu.id, coalesce(uu.name, uu.username)) "updated_by"
from call_center.cc_team_trigger qt
    left join flow.acr_routing_scheme s on s.id = qt.schema_id
    left join directory.wbt_user uc on uc.id = qt.created_by
    left join directory.wbt_user uu on uu.id = qt.updated_by
where qt.id = :Id
    and qt.team_id = :TeamId
    and exists (select 1 from call_center.cc_team q where q.id = qt.team_id and q.domain_id = :DomainId)`, map[string]interface{}{
		"TeamId":   teamId,
		"Id":       id,
		"DomainId": domainId,
	})

	if err != nil {
		return nil, model.NewCustomCodeError("store.sql_team_trigger.get.app_error", fmt.Sprintf("Id=%v, %v", id, err.Error()), extractCodeFromErr(err))
	}

	return qt, nil
}

func (s SqlTeamTriggerStore) GetAllPage(ctx context.Context, domainId int64, teamId int64, search *model.SearchTeamTrigger) ([]*model.TeamTrigger, model.AppError) {
	var list []*model.TeamTrigger

	f := map[string]interface{}{
		"DomainId":  domainId,
		"TeamId":    teamId,
		"Q":         search.GetQ(),
		"Ids":       pq.Array(search.Ids),
		"SchemaIds": pq.Array(search.SchemaIds),
		"Enabled":   search.Enabled,
	}

	err := s.ListQuery(ctx, &list, search.ListRequest,
		` team_id = :TeamId::int8
                and exists (select 1 from call_center.cc_team q where q.id = team_id and q.domain_id = :DomainId)
				and (:Q::text isnull or ( "name" ilike :Q::varchar or "description" ilike :Q::varchar ))
				and (:Ids::int4[] isnull or id = any(:Ids))
				and (:Enabled::bool isnull or enabled = :Enabled)
				and (:SchemaIds::int4[] isnull or schema_id = any(:SchemaIds))
			`,
		model.TeamTrigger{}, f)
	if err != nil {
		return nil, model.NewCustomCodeError("store.sql_team_trigger.get_all.app_error", err.Error(), extractCodeFromErr(err))
	}

	return list, nil
}

func (s SqlTeamTriggerStore) GetAllPageByUser(ctx context.Context, domainId int64, userId int64, search *model.SearchTeamTrigger) ([]*model.TeamTrigger, model.AppError) {
	var list []*model.TeamTrigger

	f := map[string]interface{}{
		"DomainId":  domainId,
		"UserId":    userId,
		"Q":         search.GetQ(),
		"Ids":       pq.Array(search.Ids),
		"SchemaIds": pq.Array(search.SchemaIds),
		"Enabled":   search.Enabled,
	}

	_, err := s.GetReplica().WithContext(ctx).Select(&list, `select qt.id,
		   call_center.cc_get_lookup(qt.schema_id, s.name) "schema",
		   qt.name,
		   qt.description,
		   qt.enabled,
		   call_center.cc_get_lookup(uc.id, coalesce(uc.name, uc.username)) "created_by",
		   qt.created_at,
		   call_center.cc_get_lookup(uu.id, coalesce(uu.name, uu.username)) "updated_by"
	from call_center.cc_team_trigger qt
		left join flow.acr_routing_scheme s on s.id = qt.schema_id
		left join directory.wbt_user uc on uc.id = qt.created_by
		left join directory.wbt_user uu on uu.id = qt.updated_by
	where
		qt.team_id = any(select a.team_id from call_center.cc_agent a where a.user_id = :UserId)
		and exists (select 1 from call_center.cc_team q where q.id = qt.team_id and q.domain_id = :DomainId
		and (:Q::text isnull or ( "name" ilike :Q::varchar or "description" ilike :Q::varchar ))
		and (:Ids::int4[] isnull or id = any(:Ids))
		and (:Enabled::bool isnull or enabled = :Enabled)
		and (:SchemaIds::int4[] isnull or schema_id = any(:SchemaIds)))`, f)
	if err != nil {
		return nil, model.NewCustomCodeError("store.sql_team_trigger.get_all_page_by_user.execute.error", err.Error(), extractCodeFromErr(err))
	}

	return list, nil
}

func (s SqlTeamTriggerStore) Update(ctx context.Context, domainId int64, teamId int64, qt *model.TeamTrigger) (*model.TeamTrigger, model.AppError) {

	err := s.GetMaster().WithContext(ctx).SelectOne(&qt, `with qt as (
    update call_center.cc_team_trigger
    set schema_id = :SchemaId,
        enabled = :Enabled,
        name = :Name,
        description = :Description,
        updated_by = :UpdatedBy,
        updated_at = :UpdatedAt
    where id = :Id
		and team_id = :TeamId
        and exists(select 1 from call_center.cc_team q where q.id = team_id and q.domain_id = :DomainId)
    returning *
)
select qt.id,
       call_center.cc_get_lookup(qt.schema_id, s.name) "schema",
       qt.name,
       qt.description,
       qt.enabled,
       call_center.cc_get_lookup(uc.id, coalesce(uc.name, uc.username)) "created_by",
       qt.created_at,
       call_center.cc_get_lookup(uu.id, coalesce(uu.name, uu.username)) "updated_by"
from qt
    left join flow.acr_routing_scheme s on s.id = qt.schema_id
    left join directory.wbt_user uc on uc.id = qt.created_by
    left join directory.wbt_user uu on uu.id = qt.updated_by`, map[string]interface{}{
		"Id":          qt.Id,
		"SchemaId":    qt.Schema.GetSafeId(),
		"Name":        qt.Name,
		"Description": qt.Description,
		"Enabled":     qt.Enabled,
		"UpdatedBy":   qt.UpdatedBy.GetSafeId(),
		"UpdatedAt":   qt.UpdatedAt,
		"TeamId":      teamId,
		"DomainId":    domainId,
	})

	if err != nil {
		return nil, model.NewCustomCodeError("store.sql_team_trigger.update.app_error", err.Error(), extractCodeFromErr(err))
	}

	return qt, nil
}

func (s SqlTeamTriggerStore) Delete(ctx context.Context, domainId int64, teamId int64, id uint32) model.AppError {
	if _, err := s.GetMaster().WithContext(ctx).Exec(`delete from call_center.cc_team_trigger qt where qt.id=:Id and qt.team_id = :TeamId 
			and exists(select 1 from call_center.cc_team q where q.id = :TeamId and q.domain_id = :DomainId)`,
		map[string]interface{}{"Id": id, "DomainId": domainId, "TeamId": teamId}); err != nil {
		return model.NewCustomCodeError("store.sql_team_trigger.delete.app_error", fmt.Sprintf("Id=%v, %s", id, err.Error()), extractCodeFromErr(err))
	}
	return nil
}
