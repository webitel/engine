package sqlstore

import (
	"fmt"
	"github.com/lib/pq"
	"github.com/webitel/engine/auth_manager"
	"github.com/webitel/engine/model"
	"github.com/webitel/engine/store"
	"net/http"
)

type SqlQueueStore struct {
	SqlStore
}

func NewSqlQueueStore(sqlStore SqlStore) store.QueueStore {
	us := &SqlQueueStore{sqlStore}
	return us
}

func (s SqlQueueStore) CheckAccess(domainId, id int64, groups []int, access auth_manager.PermissionAccess) (bool, *model.AppError) {

	res, err := s.GetReplica().SelectNullInt(`select 1
		where exists(
          select 1
          from cc_queue_acl a
          where a.dc = :DomainId
            and a.object = :Id
            and a.subject = any (:Groups::int[])
            and a.access & :Access = :Access
        )`, map[string]interface{}{"DomainId": domainId, "Id": id, "Groups": pq.Array(groups), "Access": access.Value()})

	if err != nil {
		return false, nil
	}

	return (res.Valid && res.Int64 == 1), nil
}

func (s SqlQueueStore) Create(queue *model.Queue) (*model.Queue, *model.AppError) {
	var out *model.Queue
	if err := s.GetMaster().SelectOne(&out, `with q as (
    insert into cc_queue (strategy, enabled, payload, calendar_id, priority, updated_at,
                      name, variables, timeout, domain_id, dnc_list_id, sec_locate_agent, type, team_id,
                      created_at, created_by, updated_by, description, schema_id)
values (:Strategy, :Enabled, :Payload, :CalendarId, :Priority, :UpdatedAt, :Name,
        :Variables, :Timeout, :DomainId, :DncListId, :SecLocateAgent, :Type, :TeamId, :CreatedAt, :CreatedBy, :UpdatedBy, :Description, :SchemaId)
    returning *
)
select q.id, q.strategy, q.enabled, q.payload,  q.priority, q.updated_at,
          q.name, q.variables, q.timeout, q.domain_id,  q.sec_locate_agent, q.type,
          q.created_at, cc_get_lookup(uc.id, uc.name) as created_by, cc_get_lookup(u.id, u.name) as updated_by,
          cc_get_lookup(c.id, c.name) as calendar, cc_get_lookup(cl.id, cl.name) as dnc_list, cc_get_lookup(ct.id, ct.name) as team, q.description,
		  cc_get_lookup(s.id, s.name) as schema 
from q
    inner join calendar c on q.calendar_id = c.id
    left join directory.wbt_user uc on uc.id = q.created_by
	left join directory.wbt_user u on u.id = q.updated_by
    left join cc_list cl on q.dnc_list_id = cl.id
	left join acr_routing_scheme s on q.schema_id = s.id
    left join cc_team ct on q.team_id = ct.id`,
		map[string]interface{}{
			"Strategy":       queue.Strategy,
			"Enabled":        queue.Enabled,
			"Payload":        queue.Payload,
			"CalendarId":     queue.Calendar.Id,
			"Priority":       queue.Priority,
			"UpdatedAt":      queue.UpdatedAt,
			"Name":           queue.Name,
			"Variables":      queue.Variables.ToJson(),
			"Timeout":        queue.Timeout,
			"DomainId":       queue.DomainId,
			"DncListId":      queue.DncListId(),
			"SecLocateAgent": queue.SecLocateAgent,
			"Type":           queue.Type,
			"TeamId":         queue.TeamId(),
			"CreatedAt":      queue.CreatedAt,
			"CreatedBy":      queue.CreatedBy.Id,
			"UpdatedBy":      queue.UpdatedBy.Id,
			"Description":    queue.Description,
			"SchemaId":       queue.SchemaId(),
		}); nil != err {
		return nil, model.NewAppError("SqlQueueStore.Save", "store.sql_queue.save.app_error", nil,
			fmt.Sprintf("name=%v, %v", queue.Name, err.Error()), extractCodeFromErr(err))
	} else {
		return out, nil
	}
}

func (s SqlQueueStore) GetAllPage(domainId int64, search *model.SearchQueue) ([]*model.Queue, *model.AppError) {
	var queues []*model.Queue

	if _, err := s.GetReplica().Select(&queues,
		`select q.id, q.strategy, q.enabled, q.payload,  q.priority, q.updated_at,
          q.name, q.variables, q.timeout, q.domain_id,  q.sec_locate_agent, q.type,
          q.created_at, cc_get_lookup(uc.id, uc.name) as created_by, cc_get_lookup(u.id, u.name) as updated_by,
          cc_get_lookup(c.id, c.name) as calendar, cc_get_lookup(cl.id, cl.name) as dnc_list, cc_get_lookup(ct.id, ct.name) as team, q.description,
		  cc_get_lookup(s.id, s.name) as schema, coalesce(ss.member_count, 0) count, coalesce(ss.member_waiting, 0) waiting, coalesce(act.cnt, 0) active
from cc_queue q
    inner join calendar c on q.calendar_id = c.id
    left join directory.wbt_user uc on uc.id = q.created_by
	left join directory.wbt_user u on u.id = q.updated_by
	left join acr_routing_scheme s on q.schema_id = s.id
    left join cc_list cl on q.dnc_list_id = cl.id
    left join cc_team ct on q.team_id = ct.id
    left join lateral (
        select sum(s.member_waiting) member_waiting, sum(s.member_count) member_count
        from cc_queue_statistics s
        where s.queue_id = q.id
    ) ss on true
    left join lateral (
        select count(*) cnt
        from cc_member_attempt a
        where a.queue_id = q.id and a.hangup_at = 0
    ) act on true
where q.domain_id = :DomainId and ( (:Q::varchar isnull or (q.name ilike :Q::varchar or q.description ilike :Q::varchar ) )) 
order by q.priority DESC
limit :Limit
offset :Offset`, map[string]interface{}{
			"DomainId": domainId,
			"Limit":    search.GetLimit(),
			"Offset":   search.GetOffset(),
			"Q":        search.GetQ(),
		}); err != nil {
		return nil, model.NewAppError("SqlQueueStore.GetAllPage", "store.sql_queue.get_all.app_error", nil, err.Error(), http.StatusInternalServerError)
	} else {
		return queues, nil
	}
}

func (s SqlQueueStore) GetAllPageByGroups(domainId int64, groups []int, search *model.SearchQueue) ([]*model.Queue, *model.AppError) {
	var queues []*model.Queue

	if _, err := s.GetReplica().Select(&queues,
		`select q.id, q.strategy, q.enabled, q.payload,  q.priority, q.updated_at,
          q.name, q.variables, q.timeout, q.domain_id,  q.sec_locate_agent, q.type,
          q.created_at, cc_get_lookup(uc.id, uc.name) as created_by, cc_get_lookup(u.id, u.name) as updated_by,
          cc_get_lookup(c.id, c.name) as calendar, cc_get_lookup(cl.id, cl.name) as dnc_list, cc_get_lookup(ct.id, ct.name) as team, q.description,
		  cc_get_lookup(s.id, s.name) as schema, coalesce(ss.member_count, 0) count, coalesce(ss.member_waiting, 0) waiting, coalesce(act.cnt, 0) active
from cc_queue q
    inner join calendar c on q.calendar_id = c.id
    left join directory.wbt_user uc on uc.id = q.created_by
	left join directory.wbt_user u on u.id = q.updated_by
	left join acr_routing_scheme s on q.schema_id = s.id
    left join cc_list cl on q.dnc_list_id = cl.id
    left join cc_team ct on q.team_id = ct.id
    left join lateral (
        select sum(s.member_waiting) member_waiting, sum(s.member_count) member_count
        from cc_queue_statistics s
        where s.queue_id = q.id
    ) ss on true
    left join lateral (
        select count(*) cnt
        from cc_member_attempt a
        where a.queue_id = q.id and a.hangup_at = 0
    ) act on true
where q.domain_id = :DomainId  and (
    exists(select 1
      from cc_queue_acl a
      where a.dc = c.domain_id and a.object = q.id and a.subject = any(:Groups::int[]) and a.access&:Access = :Access)
  ) and ( (:Q::varchar isnull or (q.name ilike :Q::varchar or q.description ilike :Q::varchar ) )) 
order by q.priority DESC
limit :Limit
offset :Offset`, map[string]interface{}{
			"DomainId": domainId,
			"Limit":    search.GetLimit(),
			"Offset":   search.GetOffset(),
			"Q":        search.GetQ(),
			"Groups":   pq.Array(groups),
			"Access":   auth_manager.PERMISSION_ACCESS_READ.Value(),
		}); err != nil {
		return nil, model.NewAppError("SqlQueueStore.GetAllPage", "store.sql_queue.get_all.app_error", nil, err.Error(), http.StatusInternalServerError)
	} else {
		return queues, nil
	}
}

func (s SqlQueueStore) Get(domainId int64, id int64) (*model.Queue, *model.AppError) {
	var queue *model.Queue
	if err := s.GetReplica().SelectOne(&queue, `
			select q.id, q.strategy, q.enabled, q.payload,  q.priority, q.updated_at,
          q.name, q.variables, q.timeout, q.domain_id,  q.sec_locate_agent, q.type,
          q.created_at, cc_get_lookup(uc.id, uc.name) as created_by, cc_get_lookup(u.id, u.name) as updated_by,
          cc_get_lookup(c.id, c.name) as calendar, cc_get_lookup(cl.id, cl.name) as dnc_list, cc_get_lookup(ct.id, ct.name) as team, q.description,
		  cc_get_lookup(s.id, s.name) as schema
from cc_queue q
    inner join calendar c on q.calendar_id = c.id
    left join directory.wbt_user uc on uc.id = q.created_by
	left join directory.wbt_user u on u.id = q.updated_by
	left join acr_routing_scheme s on q.schema_id = s.id
    left join cc_list cl on q.dnc_list_id = cl.id
    left join cc_team ct on q.team_id = ct.id
where q.domain_id = :DomainId and q.id = :Id 	
		`, map[string]interface{}{"Id": id, "DomainId": domainId}); err != nil {
		return nil, model.NewAppError("SqlQueueStore.Get", "store.sql_queue.get.app_error", nil,
			fmt.Sprintf("Id=%v, %s", id, err.Error()), extractCodeFromErr(err))
	} else {
		return queue, nil
	}
}

func (s SqlQueueStore) Update(queue *model.Queue) (*model.Queue, *model.AppError) {
	err := s.GetMaster().SelectOne(&queue, `with q as (
    update cc_queue q
set updated_at = :UpdatedAt,
    updated_by = :UpdatedBy,
    strategy = :Strategy,
    enabled = :Enabled,
    payload = :Payload,
    calendar_id = :CalendarId,
    priority = :Priority,
    name = :Name,
    variables = :Variables,
    timeout = :Timeout,
    dnc_list_id = :DncListId,
    sec_locate_agent = :SecLocateAgent,
    type = :Type,
    team_id = :TeamId,
	description = :Description,
	schema_id = :SchemaId
where q.id = :Id and q.domain_id = :DomainId
    returning *
)
select q.id, q.strategy, q.enabled, q.payload,  q.priority, q.updated_at,
          q.name, q.variables, q.timeout, q.domain_id,  q.sec_locate_agent, q.type,
          q.created_at, cc_get_lookup(uc.id, uc.name) as created_by, cc_get_lookup(u.id, u.name) as updated_by,
          cc_get_lookup(c.id, c.name) as calendar, cc_get_lookup(cl.id, cl.name) as dnc_list, cc_get_lookup(ct.id, ct.name) as team, q.description,
		  cc_get_lookup(s.id, s.name) as schema
from q
    inner join calendar c on q.calendar_id = c.id
    left join directory.wbt_user uc on uc.id = q.created_by
	left join directory.wbt_user u on u.id = q.updated_by
	left join acr_routing_scheme s on q.schema_id = s.id
    left join cc_list cl on q.dnc_list_id = cl.id
    left join cc_team ct on q.team_id = ct.id;`, map[string]interface{}{
		"UpdatedAt":      queue.UpdatedAt,
		"UpdatedBy":      queue.UpdatedBy.Id,
		"Strategy":       queue.Strategy,
		"Enabled":        queue.Enabled,
		"Payload":        queue.Payload,
		"CalendarId":     queue.Calendar.Id,
		"Priority":       queue.Priority,
		"Name":           queue.Name,
		"Variables":      queue.Variables.ToJson(),
		"Timeout":        queue.Timeout,
		"DncListId":      queue.DncListId(),
		"SecLocateAgent": queue.SecLocateAgent,
		"Type":           queue.Type,
		"TeamId":         queue.TeamId(),
		"SchemaId":       queue.SchemaId(),
		"Id":             queue.Id,
		"DomainId":       queue.DomainId,
		"Description":    queue.Description,
	})
	if err != nil {
		return nil, model.NewAppError("SqlQueueStore.Update", "store.sql_queue.update.app_error", nil,
			fmt.Sprintf("Id=%v, %s", queue.Id, err.Error()), extractCodeFromErr(err))
	}
	return queue, nil
}

func (s SqlQueueStore) Delete(domainId, id int64) *model.AppError {
	if _, err := s.GetMaster().Exec(`delete from cc_queue c where c.id=:Id and c.domain_id = :DomainId`,
		map[string]interface{}{"Id": id, "DomainId": domainId}); err != nil {
		return model.NewAppError("SqlQueueStore.Delete", "store.sql_queue.delete.app_error", nil,
			fmt.Sprintf("Id=%v, %s", id, err.Error()), http.StatusInternalServerError)
	}
	return nil
}
