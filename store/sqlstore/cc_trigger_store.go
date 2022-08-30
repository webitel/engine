package sqlstore

import (
	"fmt"
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

func (s SqlTriggerStore) CheckAccess(domainId int64, id int32, groups []int, access auth_manager.PermissionAccess) (bool, *model.AppError) {

	res, err := s.GetReplica().SelectNullInt(`select 1
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

func (s SqlTriggerStore) Create(domainId int64, trigger *model.Trigger) (*model.Trigger, *model.AppError) {
	if err := s.GetMaster().SelectOne(&trigger, `with t as (
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
    t.variables,
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
		return nil, model.NewAppError("SqlTriggerStore.Save", "store.sql_trigger.save.app_error", nil,
			fmt.Sprintf("name=%v, %v", trigger.Name, err.Error()), extractCodeFromErr(err))
	} else {
		return trigger, nil
	}
}

func (s SqlTriggerStore) GetAllPage(domainId int64, search *model.SearchTrigger) ([]*model.Trigger, *model.AppError) {
	var triggers []*model.Trigger

	f := map[string]interface{}{
		"DomainId": domainId,
		"Ids":      pq.Array(search.Ids),
		"Q":        search.GetQ(),
	}

	err := s.ListQuery(&triggers, search.ListRequest,
		`domain_id = :DomainId 
			and ( (:Ids::int[] isnull or id = any(:Ids) )
			and (:Q::varchar isnull or (name ilike :Q::varchar or description ilike :Q::varchar ) ))`,
		model.Trigger{}, f)
	if err != nil {
		return nil, model.NewAppError("SqlTriggerStore.GetAllPage", "store.sql_trigger.get_all.app_error", nil, err.Error(), extractCodeFromErr(err))
	}

	return triggers, nil
}

func (s SqlTriggerStore) GetAllPageByGroup(domainId int64, groups []int, search *model.SearchTrigger) ([]*model.Trigger, *model.AppError) {
	var triggers []*model.Trigger

	f := map[string]interface{}{
		"DomainId": domainId,
		"Ids":      pq.Array(search.Ids),
		"Q":        search.GetQ(),
		"Groups":   pq.Array(groups),
		"Access":   auth_manager.PERMISSION_ACCESS_READ.Value(),
	}

	err := s.ListQuery(&triggers, search.ListRequest,
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
		return nil, model.NewAppError("SqlTriggerStore.GetAllPage", "store.sql_trigger.get_all.app_error", nil, err.Error(), extractCodeFromErr(err))
	}

	return triggers, nil
}

func (s SqlTriggerStore) Get(domainId int64, id int32) (*model.Trigger, *model.AppError) {
	var trigger *model.Trigger
	f := map[string]interface{}{
		"DomainId": domainId,
		"Id":       id,
	}

	err := s.One(&trigger,
		`domain_id = :DomainId and id = :Id`,
		model.Trigger{}, f)
	if err != nil {
		return nil, model.NewAppError("SqlTriggerStore.Get", "store.sql_trigger.get.app_error", nil, err.Error(), extractCodeFromErr(err))
	}

	return trigger, nil
}

func (s SqlTriggerStore) Update(domainId int64, trigger *model.Trigger) (*model.Trigger, *model.AppError) {
	err := s.GetMaster().SelectOne(&trigger, `with t as (
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
            updated_at = :UpdatedAt
        where domain_id = :DomainId and id = :Id
        returning *)
select t.id,
       t.name,
       t.enabled,
       t.type,
       call_center.cc_get_lookup(s.id, s.name)                                as schema,
       t.variables,
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
	})
	if err != nil {
		return nil, model.NewAppError("SqlTriggerStore.Update", "store.sql_trigger.update.app_error", nil,
			fmt.Sprintf("Id=%v, %s", trigger.Id, err.Error()), extractCodeFromErr(err))
	}
	return trigger, nil
}

func (s SqlTriggerStore) Delete(domainId int64, id int32) *model.AppError {
	if _, err := s.GetMaster().Exec(`delete from call_center.cc_trigger c where c.id=:Id and c.domain_id = :DomainId`,
		map[string]interface{}{"Id": id, "DomainId": domainId}); err != nil {
		return model.NewAppError("SqlTriggerStore.Delete", "store.sql_trigger.delete.app_error", nil,
			fmt.Sprintf("Id=%v, %s", id, err.Error()), extractCodeFromErr(err))
	}
	return nil
}
