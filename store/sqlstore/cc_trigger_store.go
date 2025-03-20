package sqlstore

import (
	"context"
	"fmt"
	"github.com/webitel/engine/utils"
	"strings"

	"github.com/lib/pq"
	"github.com/webitel/engine/auth_manager"
	"github.com/webitel/engine/model"
	"github.com/webitel/engine/store"
)

type SqlTriggerStore struct {
	SqlStore
}

func NewSqlTriggerStore(sqlStore SqlStore) store.TriggerStore {
	us := &SqlTriggerStore{sqlStore}
	return us
}

func (s SqlTriggerStore) CheckAccess(ctx context.Context, domainId int64, id int32, groups []int, access auth_manager.PermissionAccess) (bool, model.AppError) {

	res, err := s.GetReplica().WithContext(ctx).SelectNullInt(`select 1
		where exists(
          select 1
          from call_center.cc_trigger_acl a
          where a.dc = :DomainId
            and a.object = :Id
            and a.subject = any (:Groups::int[])
            and a.access & :Access = :Access
        )`, map[string]interface{}{"DomainId": domainId, "Id": id, "Groups": pq.Array(groups), "Access": access.Value()})

	if err != nil {
		return false, nil
	}

	return res.Valid && res.Int64 == 1, nil
}

func (s SqlTriggerStore) Create(ctx context.Context, domainId int64, trigger *model.Trigger) (*model.Trigger, model.AppError) {
	if err := s.GetMaster().WithContext(ctx).SelectOne(&trigger, `with t as (
    insert into call_center.cc_trigger (domain_id, name, enabled, type, schema_id, variables, description, expression,
                                    timezone_id, created_by, updated_by, created_at, updated_at, timeout_sec)
    values (:DomainId, :Name, :Enabled, :Type, :SchemaId, :Variables, :Description, :Expression,
                :TimezoneId, :CreatedBy, :UpdatedBy, :CreatedAt, :UpdatedAt, :TimeoutSec)
    returning *
)
select
    t.id,
    t.name,
    t.enabled,
    t.type,
    call_center.cc_get_lookup(s.id, s.name) as schema,
    coalesce(t.variables, '{}') as variables,
    t.description,
    t.expression,
    call_center.cc_get_lookup(tz.id, tz.name) as timezone,
    t.timeout_sec as timeout,
    call_center.cc_get_lookup(uc.id, coalesce(uc.name, uc.username)::text) as created_by,
    call_center.cc_get_lookup(uu.id, coalesce(uu.name, uu.username)::text) as updated_by,
    t.created_at,
    t.updated_at
from t
    left join flow.acr_routing_scheme s on s.id = t.schema_id
    left join flow.calendar_timezones tz on tz.id = t.timezone_id
    left join directory.wbt_user uc on uc.id = t.created_by
    left join directory.wbt_user uu on uu.id = t.updated_by`,
		map[string]interface{}{
			"DomainId":    domainId,
			"Name":        trigger.Name,
			"Enabled":     trigger.Enabled,
			"Type":        trigger.Type,
			"SchemaId":    trigger.Schema.GetSafeId(),
			"Variables":   trigger.Variables.ToSafeJson(),
			"Description": trigger.Description,
			"Expression":  trigger.Expression,
			"TimezoneId":  trigger.Timezone.GetSafeId(),
			"CreatedBy":   trigger.CreatedBy.GetSafeId(),
			"UpdatedBy":   trigger.UpdatedBy.GetSafeId(),
			"CreatedAt":   trigger.CreatedAt,
			"UpdatedAt":   trigger.UpdatedAt,
			"TimeoutSec":  trigger.Timeout,
		}); nil != err {
		return nil, model.NewCustomCodeError("store.sql_trigger.save.app_error", fmt.Sprintf("name=%v, %v", trigger.Name, err.Error()), extractCodeFromErr(err))
	} else {
		return trigger, nil
	}
}

func (s SqlTriggerStore) GetAllByType(ctx context.Context, type_ string) ([]*model.TriggerWithDomainID, model.AppError) {
	var triggers []*model.TriggerWithDomainID
	fields := strings.Join(utils.MapFn(pq.QuoteIdentifier, model.Trigger{}.AllowFieldsWithDomainId()), ", ")
	tableName := fmt.Sprintf("call_center.%s", pq.QuoteIdentifier(model.Trigger{}.EntityName())) // TODO :: do not hardcode scheme
	query := fmt.Sprintf(`select %s from %s WHERE "type" =:Type and enabled`, fields, tableName)
	args := map[string]interface{}{
		"Type": type_,
	}

	_, err := s.GetReplica().WithContext(ctx).Select(&triggers, query, args)
	if err != nil {
		return nil, model.NewCustomCodeError("store.sql_trigger.get_by_type.app_error", err.Error(), extractCodeFromErr(err))
	}
	return triggers, nil
}

func (s SqlTriggerStore) GetAllPage(ctx context.Context, domainId int64, search *model.SearchTrigger) ([]*model.Trigger, model.AppError) {
	var triggers []*model.Trigger

	f := map[string]interface{}{
		"DomainId": domainId,
		"Ids":      pq.Array(search.Ids),
		"Q":        search.GetQ(),
	}

	err := s.ListQuery(ctx, &triggers, search.ListRequest,
		`domain_id = :DomainId 
			and ( (:Ids::int[] isnull or id = any(:Ids) )
			and (:Q::varchar isnull or (name ilike :Q::varchar or description ilike :Q::varchar ) ))`,
		model.Trigger{}, f)
	if err != nil {
		return nil, model.NewCustomCodeError("store.sql_trigger.get_all.app_error", err.Error(), extractCodeFromErr(err))
	}

	return triggers, nil
}

func (s SqlTriggerStore) GetAllPageByGroup(ctx context.Context, domainId int64, groups []int, search *model.SearchTrigger) ([]*model.Trigger, model.AppError) {
	var triggers []*model.Trigger

	f := map[string]interface{}{
		"DomainId": domainId,
		"Ids":      pq.Array(search.Ids),
		"Q":        search.GetQ(),
		"Groups":   pq.Array(groups),
		"Access":   auth_manager.PERMISSION_ACCESS_READ.Value(),
	}

	err := s.ListQuery(ctx, &triggers, search.ListRequest,
		`domain_id = :DomainId 
			and ( (:Ids::int[] isnull or id = any(:Ids) )
			and (:Q::varchar isnull or (name ilike :Q::varchar or description ilike :Q::varchar ) ))
			and  (
					exists(select 1
					  from call_center.cc_trigger_acl acl
					  where acl.dc = t.domain_id and acl.object = t.id and acl.subject = any(:Groups::int[]) and acl.access&:Access = :Access)
		  	)
`,
		model.Trigger{}, f)
	if err != nil {
		return nil, model.NewCustomCodeError("store.sql_trigger.get_all.app_error", err.Error(), extractCodeFromErr(err))
	}

	return triggers, nil
}

func (s SqlTriggerStore) Get(ctx context.Context, domainId int64, id int32) (*model.Trigger, model.AppError) {
	var trigger *model.Trigger
	f := map[string]interface{}{
		"DomainId": domainId,
		"Id":       id,
	}

	err := s.One(ctx, &trigger,
		`domain_id = :DomainId and id = :Id`,
		model.Trigger{}, f)
	if err != nil {
		return nil, model.NewCustomCodeError("store.sql_trigger.get.app_error", err.Error(), extractCodeFromErr(err))
	}

	return trigger, nil
}

func (s SqlTriggerStore) Update(ctx context.Context, domainId int64, trigger *model.Trigger) (*model.Trigger, model.AppError) {
	err := s.GetMaster().WithContext(ctx).SelectOne(&trigger, `with t as (
    update call_center.cc_trigger
        set name = :Name,
            enabled = :Enabled,
            schema_id = :SchemaId,
            variables = :Variables,
            description = :Description,
            expression = :Expression,
            timezone_id = :TimezoneId,
            timeout_sec = :Timeout,
            updated_by = :UpdatedBy,
            updated_at = :UpdatedAt,
			type = :Type
        where domain_id = :DomainId and id = :Id
        returning *)
select t.id,
       t.name,
       t.enabled,
       t.type,
       call_center.cc_get_lookup(s.id, s.name)                                as schema,
       coalesce(t.variables, '{}') as variables,
       t.description,
       t.expression,
       call_center.cc_get_lookup(tz.id, tz.name)                              as timezone,
       t.timeout_sec                                                          as timeout,
       call_center.cc_get_lookup(uc.id, coalesce(uc.name, uc.username)::text) as created_by,
       call_center.cc_get_lookup(uu.id, coalesce(uu.name, uu.username)::text) as updated_by,
       t.created_at,
       t.updated_at
from t
         left join flow.acr_routing_scheme s on s.id = t.schema_id
         left join flow.calendar_timezones tz on tz.id = t.timezone_id
         left join directory.wbt_user uc on uc.id = t.created_by
         left join directory.wbt_user uu on uu.id = t.updated_by;`, map[string]interface{}{
		"DomainId":    domainId,
		"Id":          trigger.Id,
		"Name":        trigger.Name,
		"Enabled":     trigger.Enabled,
		"SchemaId":    trigger.Schema.GetSafeId(),
		"Variables":   trigger.Variables.ToSafeJson(),
		"Description": trigger.Description,
		"Expression":  trigger.Expression,
		"TimezoneId":  trigger.Timezone.GetSafeId(),
		"UpdatedBy":   trigger.UpdatedBy.GetSafeId(),
		"UpdatedAt":   trigger.UpdatedAt,
		"Timeout":     trigger.Timeout,
		"Type":        trigger.Type,
	})
	if err != nil {
		return nil, model.NewCustomCodeError("store.sql_trigger.update.app_error", fmt.Sprintf("Id=%v, %s", trigger.Id, err.Error()), extractCodeFromErr(err))
	}
	return trigger, nil
}

func (s SqlTriggerStore) Delete(ctx context.Context, domainId int64, id int32) model.AppError {
	if _, err := s.GetMaster().WithContext(ctx).Exec(`delete from call_center.cc_trigger c where c.id=:Id and c.domain_id = :DomainId`,
		map[string]interface{}{"Id": id, "DomainId": domainId}); err != nil {
		return model.NewCustomCodeError("store.sql_trigger.delete.app_error", fmt.Sprintf("Id=%v, %s", id, err.Error()), extractCodeFromErr(err))
	}
	return nil
}

func (s SqlTriggerStore) CreateJob(ctx context.Context, domainId int64, triggerId int32, _ map[string]string) (*model.TriggerJob, model.AppError) {
	var job *model.TriggerJob
	err := s.GetMaster().WithContext(ctx).SelectOne(&job, `with j as (
    insert into call_center.cc_trigger_job (trigger_id, state, created_at, parameters, domain_id)
        select t.id,
               0,
               now(),
               jsonb_build_object('variables', t.variables,
                                  'schema_id', t.schema_id,
                                  'timeout', t.timeout_sec
                   ) as params,
               t.domain_id
        from call_center.cc_trigger t
        where t.id = :TriggerId
          and t.domain_id = :DomainId
        returning call_center.cc_trigger_job.*)
select
    j.id,
    call_center.cc_get_lookup(t.id, t.name) as trigger,
    j.state,
    j.created_at,
    j.started_at,
    j.stopped_at,
    j.parameters,
    j.error,
    j.result
from j
    left join call_center.cc_trigger t on t.id = j.trigger_id`, map[string]interface{}{
		"TriggerId": triggerId,
		"DomainId":  domainId,
	})

	if err != nil {
		return nil, model.NewCustomCodeError("store.sql_trigger.job_create.app_error", fmt.Sprintf("TriggerId=%v, %s", triggerId, err.Error()), extractCodeFromErr(err))
	}

	return job, nil
}

func (s SqlTriggerStore) GetAllJobs(ctx context.Context, triggerId int32, search *model.SearchTriggerJob) ([]*model.TriggerJob, model.AppError) {
	var jobs []*model.TriggerJob

	f := map[string]interface{}{
		"TriggerId":    triggerId,
		"From":         model.GetBetweenFromTime(search.CreatedAt),
		"To":           model.GetBetweenToTime(search.CreatedAt),
		"StartedFrom":  model.GetBetweenFromTime(search.StartedAt),
		"StartedTo":    model.GetBetweenToTime(search.StartedAt),
		"State":        pq.Array(search.State),
		"DurationFrom": model.GetBetweenFrom(search.Duration),
		"DurationTo":   model.GetBetweenTo(search.Duration),
	}

	err := s.ListQueryMaster(ctx, &jobs, search.ListRequest,
		`trigger_id = :TriggerId
				and ( :From::timestamptz isnull or created_at >= :From::timestamptz )
				and ( :To::timestamptz isnull or created_at <= :To::timestamptz )
				and ( :StartedFrom::timestamptz isnull or started_at >= :StartedFrom::timestamptz )
				and ( :StartedTo::timestamptz isnull or started_at <= :StartedTo::timestamptz )
				and ( :State::int[] isnull or state = any(:State::int[]) )

				and ( :DurationFrom::int8 isnull or extract(epoch from coalesce(stopped_at, started_at, now()) - created_at)::int8 >= :DurationFrom::int8 )
				and ( :DurationTo::int8 isnull or extract(epoch from coalesce(stopped_at, started_at, now()) - created_at)::int8 <= :DurationTo::int8 )
			`,
		model.TriggerJob{}, f)
	if err != nil {
		return nil, model.NewCustomCodeError("store.sql_trigger.get_all.app_error", err.Error(), extractCodeFromErr(err))
	}

	return jobs, nil
}
