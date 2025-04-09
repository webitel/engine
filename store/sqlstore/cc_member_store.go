package sqlstore

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/go-gorp/gorp"
	"github.com/lib/pq"
	"github.com/webitel/engine/model"
	"github.com/webitel/engine/store"
	"github.com/webitel/wlog"
	"net/http"
	"strconv"
	"strings"
)

type SqlMemberStore struct {
	SqlStore
}

func NewSqlMemberStore(sqlStore SqlStore) store.MemberStore {
	us := &SqlMemberStore{
		SqlStore: sqlStore,
	}
	return us
}

func (s SqlMemberStore) Create(ctx context.Context, domainId int64, member *model.Member) (*model.Member, model.AppError) {
	var out *model.Member
	if err := s.GetMaster().WithContext(ctx).SelectOne(&out, `with m as (
			insert into call_center.cc_member (queue_id, priority, expire_at, variables, name, timezone_id, communications, bucket_id, ready_at, domain_id, agent_id, skill_id)
			values (:QueueId, :Priority, :ExpireAt, :Variables, :Name, :TimezoneId, :Communications, :BucketId, :MinOfferingAt, :DomainId, :AgentId, :SkillId)
			returning *
		)
		select m.id,  m.stop_at, m.stop_cause, m.attempts, m.last_hangup_at, m.created_at, m.queue_id, m.priority, m.expire_at, m.variables, m.name, call_center.cc_get_lookup(ct.id, ct.name) as "timezone",
			   call_center.cc_member_communications(m.communications) as communications,  call_center.cc_get_lookup(qb.id, qb.name::text) as bucket, ready_at,
               call_center.cc_get_lookup(agn.id, agn.name::text) as agent, call_center.cc_get_lookup(cs.id, cs.name::text) as skill,
			   (select e.schema_id
				from call_center.cc_queue_events e
				where e.queue_id = m.queue_id and e.enabled and e.event = 'add_member'
				limit 1) as hook_created
		from m
			left join flow.calendar_timezones ct on m.timezone_id = ct.id
			left join call_center.cc_bucket qb on m.bucket_id = qb.id
			left join call_center.cc_skill cs on m.skill_id = cs.id
			left join call_center.cc_agent_list agn on m.agent_id = agn.id`,
		map[string]interface{}{
			"DomainId":       domainId,
			"QueueId":        member.QueueId,
			"Priority":       member.Priority,
			"ExpireAt":       member.ExpireAt,
			"Variables":      member.Variables.ToJson(),
			"Name":           member.Name,
			"TimezoneId":     member.Timezone.Id,
			"Communications": member.ToJsonCommunications(),
			"BucketId":       member.Bucket.GetSafeId(),
			"MinOfferingAt":  member.MinOfferingAt,
			"AgentId":        member.Agent.GetSafeId(),
			"SkillId":        member.Skill.GetSafeId(),
		}); nil != err {
		return nil, model.NewCustomCodeError("store.sql_member.save.app_error", fmt.Sprintf("name=%v, %v", member.Name, err.Error()), extractCodeFromErr(err))
	} else {
		return out, nil
	}
}

func (s SqlMemberStore) BulkCreate(ctx context.Context, domainId, queueId int64, fileName string, members []*model.Member) ([]int64, model.AppError) {
	var err error
	var stmp *sql.Stmt
	var tx *gorp.Transaction
	var bulkCount = 5000

	if v, appE := s.SystemSettings().ValueByName(ctx, domainId, model.SysNameMemberInsertChunkSize); appE == nil && v != nil && v.Int() != nil {
		bulkCount = *v.Int()
	}

	tx, err = s.GetMaster().Begin()
	if err != nil {
		if tx != nil {
			tx.Rollback()
		}
		return nil, model.NewInternalError("store.sql_member.bulk_save.app_error", err.Error())
	}

	tableName := fmt.Sprintf("cc_member_tmp_%d", model.GetMillis())
	_, err = tx.WithContext(ctx).Exec("CREATE temp table " + tableName + " ON COMMIT DROP as table call_center.cc_member with no data")

	if err != nil {
		if tx != nil {
			tx.Rollback()
		}
		return nil, model.NewInternalError("store.sql_member.bulk_save.app_error", err.Error())
	}

	if fileName == "" {
		fileName = model.NewId()
	}

	result := make([]int64, 0, len(members))

	stmp, err = tx.Prepare(pq.CopyIn(tableName, "id", "queue_id", "priority", "expire_at", "variables", "name",
		"timezone_id", "communications", "bucket_id", "ready_at", "agent_id", "skill_id", "import_id"))
	if err != nil {
		goto _error
	}

	defer stmp.Close()
	for k, v := range members {
		_, err = stmp.Exec(k, queueId, v.Priority, v.ExpireAt, v.Variables.ToJson(), v.Name, v.Timezone.Id, v.ToJsonCommunications(),
			v.Bucket.GetSafeId(), v.MinOfferingAt, v.Agent.GetSafeId(), v.Skill.GetSafeId(), fileName)
		if err != nil {
			goto _error
		}
	}

	_, err = stmp.Exec()
	if err != nil {
		goto _error
	} else {

		for i := 0; ; i += bulkCount {
			chunk := make([]int64, 0, bulkCount)

			_, err = tx.Select(&chunk, `with i as (
			insert into call_center.cc_member(queue_id, priority, expire_at, variables, name, timezone_id, communications, bucket_id, ready_at, domain_id, agent_id, skill_id, import_id)
			select queue_id, priority, expire_at, variables, name, timezone_id, communications, bucket_id, ready_at, :DomainId, agent_id, skill_id, import_id
			from `+tableName+`
			order by `+tableName+`.id
            limit `+strconv.Itoa(bulkCount)+` 
            offset `+strconv.Itoa(i)+` 
			returning id
		)
		select id from i`, map[string]interface{}{
				"DomainId": domainId,
			})

			if err != nil {
				goto _error
			}

			result = append(result, chunk...)

			if len(chunk) != bulkCount {
				break
			}
		}
	}

	err = tx.Commit()
	if err != nil {
		return nil, model.NewCustomCodeError("store.sql_member.bulk_save.app_error", err.Error(), extractCodeFromErr(err))
	}

	return result, nil

_error:
	tx.Rollback()
	if err == nil {
		return nil, model.NewInternalError("store.sql_member.bulk_save.app_error", "Unknown error")
	}

	wlog.Error(fmt.Sprintf("CreateMemberBulk: sql error, %s", err.Error()))
	return nil, model.NewCustomCodeError("store.sql_member.bulk_save.app_error", err.Error(), extractCodeFromErr(err))
}

// todo fix deprecated fields

func (s SqlMemberStore) SearchMembers(ctx context.Context, domainId int64, search *model.SearchMemberRequest) ([]*model.Member, model.AppError) {
	var members []*model.Member

	order := GetOrderBy("cc_member", model.MemberDeprecatedField(search.Sort))
	if order == "" {
		order = "order by created_at desc"
	}

	fields := GetFields(model.MemberDeprecatedFields(search.Fields), model.Member{})

	query := `with result as materialized (select m.id, row_number() over ( ` + order + ` ) rn
                from call_center.cc_member m
                where m.domain_id = :Domain::int8
                  and (:Ids::int8[] isnull or m.id = any (:Ids::int8[]))
				  and (:Variables::jsonb isnull or variables @> :Variables::jsonb)
                  and (:QueueIds::int4[] isnull or m.queue_id = any (:QueueIds::int4[]))
                  and (:QueueId::int4 isnull or m.queue_id = :QueueId::int4)
                  and (:BucketIds::int4[] isnull or m.bucket_id = any (:BucketIds::int4[]))
                  and (:Destination::varchar isnull or
                       m.search_destinations && array [:Destination::varchar]::varchar[])

                  and (:CreatedFrom::timestamptz isnull or m.created_at >= :CreatedFrom::timestamptz)
                  and (:CreatedTo::timestamptz isnull or created_at <= :CreatedTo::timestamptz)

                  and (:OfferingFrom::timestamptz isnull or m.ready_at >= :OfferingFrom::timestamptz)
                  and (:OfferingTo::timestamptz isnull or m.ready_at <= :OfferingTo::timestamptz)

                  and (:PriorityFrom::int isnull or m.priority >= :PriorityFrom::int)
                  and (:PriorityTo::int isnull or m.priority <= :PriorityTo::int)
                  and (:AttemptsFrom::int isnull or m.attempts >= :AttemptsFrom::int)
                  and (:AttemptsTo::int isnull or m.attempts <= :AttemptsTo::int)

                  and (:AgentIds::int4[] isnull or m.agent_id = any (:AgentIds::int4[]))

                  and (:StopCauses::varchar[] isnull or m.stop_cause = any (:StopCauses::varchar[]))
                  and (:Name::varchar isnull or m.name ilike :Name::varchar)
                  and (:Q::varchar isnull or
                       (m.name ~~ :Q::varchar or
                        m.search_destinations && array [replace(rtrim(:Q::varchar, '%'), '\', '')]::varchar[]))
                ` + order + `
                limit :Limit offset :Offset )
   , comm as materialized(select c.id, json_build_object('id', c.id, 'name', c.name)::jsonb j
              from call_center.cc_communication c
              where c.domain_id = :Domain)
   , resources as materialized(select r.id, json_build_object('id', r.id, 'name', r.name)::jsonb j
                   from call_center.cc_outbound_resource r
                   where r.domain_id = :Domain)
   , list as materialized (select m.id,
                                  call_center.cc_member_destination_views_to_json(array(select (xid::int2,
                                                                                                x ->> 'destination',
                                                                                                resources.j,
                                                                                                comm.j,
                                                                                                (x -> 'priority')::int,
                                                                                                (x -> 'state')::int,
                                                                                                x -> 'description',
                                                                                                (x -> 'last_activity_at')::int8,
                                                                                                (x -> 'attempts')::int,
                                                                                                x ->> 'last_cause',
                                                                                                x ->>
                                                                                                'display')::call_center.cc_member_destination_view
                                                                                        from jsonb_array_elements(m.communications) with ordinality as x (x, xid)
                                                                                                 left join comm on comm.id = (x -> 'type' -> 'id')::int
                                                                                                 left join resources on resources.id = (x -> 'resource' -> 'id')::int)) communications,
                                  call_center.cc_get_lookup(cq.id::int8, cq.name::varchar)                                                                                    queue,
                                  m.priority,
                                  m.expire_at,
                                  m.created_at,
                                  m.variables,
                                  m.name,
                                  call_center.cc_get_lookup(m.timezone_id::bigint,
                                                            ct.name::varchar)                                                                                           "timezone",
                                  call_center.cc_get_lookup(m.bucket_id, cb.name::varchar)                                                                              bucket,
                                  m.ready_at                                                                     as                                                     ready_at,
                                  m.stop_cause,
                                  m.stop_at,
                                  m.last_hangup_at                                                               as                                                     last_hangup_at,
                                  m.attempts,
                                  call_center.cc_get_lookup(a.id, coalesce(agn.name::varchar, agn.username::varchar)::varchar) agent,
                                  call_center.cc_get_lookup(cs.id, cs.name::varchar)                                                                                    skill,
                                  exists(select 1 from call_center.cc_member_attempt a where a.member_id = m.id for update of a skip locked ) as                                                     reserved
                           from result 
									inner join call_center.cc_member m on result.id = m.id
                                    left join call_center.cc_queue cq on cq.id = m.queue_id
                                    left join flow.calendar_timezones ct on ct.id = m.timezone_id
                                    left join call_center.cc_agent a on m.agent_id = a.id
                                    left join directory.wbt_user agn on agn.id = a.user_id
                                    left join call_center.cc_bucket cb on m.bucket_id = cb.id
                                    left join call_center.cc_skill cs on m.skill_id = cs.id
                           order by result.rn)
	select ` + strings.Join(fields, " ,") + ` from list`

	if _, err := s.GetMaster().WithContext(ctx).Select(&members, query, map[string]interface{}{
		"Domain":    domainId,
		"Limit":     search.GetLimit(),
		"Offset":    search.GetOffset(),
		"Q":         search.GetQ(),
		"Variables": search.Variables.ToSafeJson(),

		"Ids":         pq.Array(search.Ids),
		"QueueIds":    pq.Array(search.QueueIds),
		"BucketIds":   pq.Array(search.BucketIds),
		"AgentIds":    pq.Array(search.AgentIds),
		"Destination": search.Destination,
		"QueueId":     search.QueueId,

		"CreatedFrom":  model.GetBetweenFromTime(search.CreatedAt),
		"CreatedTo":    model.GetBetweenToTime(search.CreatedAt),
		"OfferingFrom": model.GetBetweenFromTime(search.OfferingAt),
		"OfferingTo":   model.GetBetweenToTime(search.OfferingAt),

		"PriorityFrom": model.GetBetweenFrom(search.Priority),
		"PriorityTo":   model.GetBetweenTo(search.Priority),
		"AttemptsFrom": model.GetBetweenFrom(search.Attempts),
		"AttemptsTo":   model.GetBetweenTo(search.Attempts),

		"StopCauses": pq.Array(search.StopCauses),
		"Name":       search.Name,
	}); err != nil {
		return nil, model.NewInternalError("store.sql_member.get_all.app_error", err.Error())
	} else {
		return members, nil
	}
}

func (s SqlMemberStore) Get(ctx context.Context, domainId, queueId, id int64) (*model.Member, model.AppError) {
	var member *model.Member
	if err := s.GetReplica().WithContext(ctx).SelectOne(&member, `select m.id,  m.stop_at, m.stop_cause, m.attempts, m.last_hangup_at, m.created_at, m.queue_id, m.priority, m.expire_at, m.variables, m.name, call_center.cc_get_lookup(ct.id, ct.name) as "timezone",
			   call_center.cc_member_communications(m.communications) as communications,  call_center.cc_get_lookup(qb.id, qb.name::text) as bucket, ready_at,
               call_center.cc_get_lookup(cs.id, cs.name::text) as skill, call_center.cc_get_lookup(a.id, (coalesce(agn.name, agn.username))::varchar) agent
		from call_center.cc_member m
			left join flow.calendar_timezones ct on m.timezone_id = ct.id
			left join call_center.cc_bucket qb on m.bucket_id = qb.id
			left join call_center.cc_agent a on m.agent_id = a.id
			left join directory.wbt_user agn on agn.id = a.user_id
		    left join call_center.cc_skill cs on m.skill_id = cs.id
	where m.id = :Id and m.queue_id = :QueueId and exists(select 1 from call_center.cc_queue q where q.id = :QueueId and q.domain_id = :DomainId)`, map[string]interface{}{
		"Id":       id,
		"DomainId": domainId,
		"QueueId":  queueId,
	}); err != nil {
		return nil, model.NewCustomCodeError("store.sql_member.get.app_error", fmt.Sprintf("Id=%v, %s", id, err.Error()), extractCodeFromErr(err))
	} else {
		return member, nil
	}
}

func (s SqlMemberStore) Update(ctx context.Context, domainId int64, member *model.Member) (*model.Member, model.AppError) {
	err := s.GetMaster().WithContext(ctx).SelectOne(&member, `with m as (
    update call_center.cc_member m1
        set priority = :Priority,
            expire_at = :ExpireAt,
            variables = :Variables,
            name = :Name,
            timezone_id = :TimezoneId,
            communications = :Communications,
            bucket_id = :BucketId,
            ready_at = :MinOfferingAt,
            stop_cause = :StopCause::varchar,
            agent_id = :AgentId,
            skill_id = :SkillId,
            stop_at = case
                          when :StopCause::varchar notnull and stop_at isnull then now()
                          when :StopCause::varchar isnull and stop_at notnull then null
                          else stop_at end,
            attempts = :Attempts
        where m1.id = :Id and m1.queue_id = :QueueId
        returning *)
select m.id,
       m.stop_at,
       m.stop_cause,
       m.attempts,
       m.last_hangup_at,
       m.created_at,
       m.queue_id,
       m.priority,
       m.expire_at,
       m.variables,
       m.name,
       call_center.cc_get_lookup(ct.id, ct.name)              as                    "timezone",
       call_center.cc_member_communications(m.communications) as                    communications,
       call_center.cc_get_lookup(qb.id, qb.name::text)        as                    bucket,
       ready_at,
       call_center.cc_get_lookup(cs.id, cs.name::text)        as                    skill,
       call_center.cc_get_lookup(a.id, (coalesce(agn.name, agn.username))::varchar) agent,
       ac.id active_attempt_id,
       ac.node_id active_app_id
from m
         left join flow.calendar_timezones ct on m.timezone_id = ct.id
         left join call_center.cc_bucket qb on m.bucket_id = qb.id
         left join call_center.cc_skill cs on m.skill_id = cs.id
         left join call_center.cc_agent a on m.agent_id = a.id
         left join directory.wbt_user agn on agn.id = a.user_id
         left join lateral (
              select a.id, a.node_id
              from call_center.cc_member_attempt a 
              where a.member_id = m.id and a.leaving_at isnull 
              limit 1
    ) ac on true`, map[string]interface{}{
		"Priority":       member.Priority,
		"ExpireAt":       member.ExpireAt,
		"Variables":      member.Variables.ToJson(),
		"Name":           member.Name,
		"TimezoneId":     member.Timezone.Id,
		"Communications": member.ToJsonCommunications(),
		"BucketId":       member.Bucket.GetSafeId(),
		"Id":             member.Id,
		"QueueId":        member.QueueId,
		"DomainId":       domainId,
		"MinOfferingAt":  member.MinOfferingAt,
		"StopCause":      member.StopCause,
		"AgentId":        member.Agent.GetSafeId(),
		"SkillId":        member.Skill.GetSafeId(),
		"Attempts":       member.Attempts,
	})
	if err != nil {
		code := extractCodeFromErr(err)
		if code == http.StatusNotFound { //todo
			return nil, model.NewBadRequestError("store.sql_member.update.lock", fmt.Sprintf("Id=%v, %s", member.Id, err.Error()))
		}

		return nil, model.NewCustomCodeError("store.sql_member.update.app_error", fmt.Sprintf("Id=%v, %s", member.Id, err.Error()), code)
	}
	return member, nil
}

// TODO add force
func (s SqlMemberStore) Delete(ctx context.Context, queueId, id int64) model.AppError {
	var cnt int64
	res, err := s.GetMaster().WithContext(ctx).Exec(`delete
from call_center.cc_member c
where c.id = :Id
  and c.queue_id = :QueueId
  and not exists(select 1 from call_center.cc_member_attempt a where a.member_id = c.id and a.state != 'leaving' for update)`,
		map[string]interface{}{"Id": id, "QueueId": queueId})

	if err != nil {
		return model.NewCustomCodeError("store.sql_member.delete.app_error", fmt.Sprintf("Id=%v, %s", id, err.Error()), extractCodeFromErr(err))
	}

	cnt, err = res.RowsAffected()
	if err != nil {
		return model.NewCustomCodeError("store.sql_member.delete.app_error", fmt.Sprintf("Id=%v, %s", id, err.Error()), extractCodeFromErr(err))
	}

	if cnt == 0 {
		return model.NewNotFoundError("store.sql_member.delete.app_error", fmt.Sprintf("Id=%v, not found", id))
	}

	return nil
}

func (s SqlMemberStore) multiDelete(ctx context.Context, sort, limit string, filters map[string]interface{}) ([]*model.Member, error) {
	var res []*model.Member

	_, err := s.GetMaster().WithContext(ctx).Select(&res, `with m as (
    delete from call_center.cc_member m
    where m.id in (
 		select m.id
		from call_center.cc_member m
			where m.domain_id = :DomainId::int8
				  and (:Ids::int8[] isnull or m.id = any (:Ids::int8[]))
				  and (:QueueIds::int4[] isnull or m.queue_id = any (:QueueIds::int4[]))
				  and (:BucketIds::int4[] isnull or m.bucket_id = any (:BucketIds::int4[]))
				  and (:Destination::varchar isnull or
					   m.search_destinations && array [:Destination::varchar]::varchar[])
		
				  and (:CreatedFrom::timestamptz isnull or m.created_at >= :CreatedFrom::timestamptz)
				  and (:CreatedTo::timestamptz isnull or created_at <= :CreatedTo::timestamptz)
		
				  and (:OfferingFrom::timestamptz isnull or m.ready_at >= :OfferingFrom::timestamptz)
				  and (:OfferingTo::timestamptz isnull or m.ready_at <= :OfferingTo::timestamptz)
		
				  and (:PriorityFrom::int isnull or m.priority >= :PriorityFrom::int)
				  and (:PriorityTo::int isnull or m.priority <= :PriorityTo::int)
				  and (:AttemptsFrom::int isnull or m.attempts >= :AttemptsFrom::int)
				  and (:AttemptsTo::int isnull or m.attempts <= :AttemptsTo::int)
		
				  and (:StopCauses::varchar[] isnull or m.stop_cause = any (:StopCauses::varchar[]))
				  and (:Name::varchar isnull or m.name ilike :Name::varchar)
				  and (:Q::varchar isnull or
					   (m.name ~~ :Q::varchar or m.search_destinations && array [rtrim(:Q::varchar, '%')]::varchar[]))
		
				and (:Numbers::varchar[] isnull or search_destinations && :Numbers::varchar[])
				and (:Variables::jsonb isnull or variables @> :Variables::jsonb)
				and (:AgentIds::int4[] isnull or m.agent_id = any(:AgentIds::int4[]))
				and not exists(select 1 from call_center.cc_member_attempt a where a.member_id = m.id and a.state != 'leaving' for update)
		`+sort+`
		`+limit+`
    )
    returning *
)
select m.id,  m.stop_at, m.stop_cause, m.attempts, m.last_hangup_at, m.created_at, m.queue_id, m.priority, m.expire_at, m.variables, m.name, call_center.cc_get_lookup(ct.id, ct.name) as "timezone",
			   call_center.cc_member_communications(m.communications) as communications,  call_center.cc_get_lookup(qb.id, qb.name::text) as bucket, ready_at,
               call_center.cc_get_lookup(cs.id, cs.name::text) as skill, call_center.cc_get_lookup(agn.id, agn.name::varchar) agent
		from m
			left join flow.calendar_timezones ct on m.timezone_id = ct.id
			left join call_center.cc_bucket qb on m.bucket_id = qb.id
			left join call_center.cc_skill cs on m.skill_id = cs.id
			left join call_center.cc_agent_list agn on m.agent_id = agn.id`, filters)

	return res, err
}

func (s SqlMemberStore) multiDeleteWithoutResult(ctx context.Context, sort, limit string, filters map[string]interface{}) error {
	_, err := s.GetMaster().WithContext(ctx).Exec(`
delete from call_center.cc_member m
    where m.id in (
 		select m.id
		from call_center.cc_member m
			where m.domain_id = :DomainId::int8
				  and (:Ids::int8[] isnull or m.id = any (:Ids::int8[]))
				  and (:QueueIds::int4[] isnull or m.queue_id = any (:QueueIds::int4[]))
				  and (:BucketIds::int4[] isnull or m.bucket_id = any (:BucketIds::int4[]))
				  and (:Destination::varchar isnull or
					   m.search_destinations && array [:Destination::varchar]::varchar[])
		
				  and (:CreatedFrom::timestamptz isnull or m.created_at >= :CreatedFrom::timestamptz)
				  and (:CreatedTo::timestamptz isnull or created_at <= :CreatedTo::timestamptz)
		
				  and (:OfferingFrom::timestamptz isnull or m.ready_at >= :OfferingFrom::timestamptz)
				  and (:OfferingTo::timestamptz isnull or m.ready_at <= :OfferingTo::timestamptz)
		
				  and (:PriorityFrom::int isnull or m.priority >= :PriorityFrom::int)
				  and (:PriorityTo::int isnull or m.priority <= :PriorityTo::int)
				  and (:AttemptsFrom::int isnull or m.attempts >= :AttemptsFrom::int)
				  and (:AttemptsTo::int isnull or m.attempts <= :AttemptsTo::int)
		
				  and (:StopCauses::varchar[] isnull or m.stop_cause = any (:StopCauses::varchar[]))
				  and (:Name::varchar isnull or m.name ilike :Name::varchar)
				  and (:Q::varchar isnull or
					   (m.name ~~ :Q::varchar or m.search_destinations && array [rtrim(:Q::varchar, '%')]::varchar[]))
		
				and (:Numbers::varchar[] isnull or search_destinations && :Numbers::varchar[])
				and (:Variables::jsonb isnull or variables @> :Variables::jsonb)
				and (:AgentIds::int4[] isnull or m.agent_id = any(:AgentIds::int4[]))
				and not exists(select 1 from call_center.cc_member_attempt a where a.member_id = m.id and a.state != 'leaving' for update)
		`+sort+`
		`+limit+`
    )`, filters)

	return err
}

func (s SqlMemberStore) MultiDelete(ctx context.Context, domainId int64, search *model.MultiDeleteMembers, withoutMembers bool) ([]*model.Member, model.AppError) {
	var res []*model.Member
	var err error

	sort := ""
	limit := ""

	if search.PerPage > 0 {
		limit = fmt.Sprintf("limit %d", search.PerPage)
	}

	if search.Sort != "" {
		sort = GetOrderBy(model.Member{}.EntityName(), search.Sort)
	}

	filters := map[string]interface{}{
		"DomainId":    domainId,
		"Q":           search.GetQ(),
		"QueueIds":    pq.Array(search.QueueIds),
		"Ids":         pq.Array(search.Ids),
		"BucketIds":   pq.Array(search.BucketIds),
		"Destination": search.Destination,

		"CreatedFrom":  model.GetBetweenFromTime(search.CreatedAt),
		"CreatedTo":    model.GetBetweenToTime(search.CreatedAt),
		"OfferingFrom": model.GetBetweenFromTime(search.OfferingAt),
		"OfferingTo":   model.GetBetweenToTime(search.OfferingAt),

		"PriorityFrom": model.GetBetweenFrom(search.Priority),
		"PriorityTo":   model.GetBetweenTo(search.Priority),
		"AttemptsFrom": model.GetBetweenFrom(search.Attempts),
		"AttemptsTo":   model.GetBetweenTo(search.Attempts),

		"StopCauses": pq.Array(search.StopCauses),
		"Name":       search.Name,

		"AgentIds":  pq.Array(search.AgentIds),
		"Numbers":   pq.Array(search.Numbers),
		"Variables": search.Variables.ToSafeJson(),
	}

	if withoutMembers {
		err = s.multiDeleteWithoutResult(ctx, sort, limit, filters)
	} else {
		res, err = s.multiDelete(ctx, sort, limit, filters)
	}

	if err != nil {
		return nil, model.NewCustomCodeError("store.sql_member.multi_delete.app_error", fmt.Sprintf("Ids=%v, %s", search.Ids, err.Error()), extractCodeFromErr(err))
	}

	return res, nil
}

func (s SqlMemberStore) ResetMembers(ctx context.Context, domainId int64, req *model.ResetMembers) (int64, model.AppError) {
	res, err := s.GetMaster().WithContext(ctx).Exec(`update call_center.cc_member m2
    set stop_cause = null,
        stop_at = null,
        attempts = 0,
        communications = (select jsonb_agg((xx - 'stop_at' - 'attempts')::jsonb) as e
			from jsonb_array_elements(m2.communications) xx)
from (
    select m.id
    from call_center.cc_member m
    where m.domain_id = :DomainId
      and m.queue_id = :QueueId
      and (stop_at IS NOT NULL)
      AND ((stop_cause)::text <> ALL ('{success,expired,cancel,terminate,no_communications}'::text[]))
      and (:Ids::int8[] isnull or m.id = any (:Ids::int8[]))
      and (:Numbers::varchar[] isnull or m.search_destinations && :Numbers::varchar[])
      and (:Variables::jsonb isnull or m.variables @> :Variables::jsonb)
      and (:Buckets::int8[] isnull or m.bucket_id = any (:Buckets::int8[]))
      and (:AgentIds::int4[] isnull or m.agent_id = any (:AgentIds::int4[]))
      and (:Cause::text[] isnull or m.stop_cause = any (:Cause::text[]))
	  and ( (:PriorityFrom::smallint isnull or :PriorityFrom::smallint = 0 or priority >= :PriorityFrom ))
	  and ( (:PriorityTo::smallint isnull or :PriorityTo::smallint = 0 or priority <= :PriorityTo ))
	  and ( :CreatedAtFrom::timestamptz isnull or created_at >= :CreatedAtFrom::timestamptz )
	  and ( :CreatedAtTo::timestamptz isnull or created_at <= :CreatedAtTo::timestamptz )
 ) x
where m2.id = x.id`, map[string]interface{}{
		"DomainId":      domainId,
		"Ids":           pq.Array(req.Ids),
		"Buckets":       pq.Array(req.Buckets),
		"Cause":         pq.Array(req.Causes),
		"AgentIds":      pq.Array(req.AgentIds),
		"Numbers":       pq.Array(req.Numbers),
		"Variables":     req.Variables.ToSafeJson(),
		"QueueId":       req.QueueId,
		"PriorityFrom":  model.GetBetweenFrom(req.Priority),
		"PriorityTo":    model.GetBetweenTo(req.Priority),
		"CreatedAtFrom": model.GetBetweenFromTime(req.CreatedAt),
		"CreatedAtTo":   model.GetBetweenToTime(req.CreatedAt),
	})

	if err != nil {
		return 0, model.NewCustomCodeError("store.sql_member.reset.app_error", fmt.Sprintf("QueueId=%v, %s", req.QueueId, err.Error()), extractCodeFromErr(err))
	}
	var cnt int64
	cnt, err = res.RowsAffected()
	if err != nil {
		return 0, model.NewCustomCodeError("store.sql_member.reset.app_error", fmt.Sprintf("QueueId=%v, %s", req.QueueId, err.Error()), extractCodeFromErr(err))
	}

	return cnt, nil
}

func (s SqlMemberStore) AttemptsList(ctx context.Context, memberId int64) ([]*model.MemberAttempt, model.AppError) {
	var attempts []*model.MemberAttempt
	//FIXME
	if _, err := s.GetMaster().WithContext(ctx).Select(&attempts, `with active as (
    select a.id,
           --a.member_id,
           (extract(EPOCH from a.created_at) * 1000)::int8 as created_at,
           'TODO' as destination,
           a.weight,
           a.originate_at,
           a.answered_at,
           a.bridged_at,
           a.hangup_at,
           call_center.cc_get_lookup(cor.id, cor.name) as resource,
           leg_a_id,
           leg_b_id,
           node_id as node,
           result,
           call_center.cc_get_lookup(u.id, u.name) as agent,
           call_center.cc_get_lookup(cb.id::int8, cb.name::varchar) as bucket,
           logs,
           false as active
    from call_center.cc_member_attempt a
        left join call_center.cc_outbound_resource cor on a.resource_id = cor.id
        left join call_center.cc_agent ca on a.agent_id = ca.id
        left join directory.wbt_user u on u.id = ca.user_id
        left join call_center.cc_bucket cb on a.bucket_id = cb.id
    where a.member_id = :MemberId
    order by a.created_at
), log as (
    select a.id,
          -- a.member_id,
           (extract(EPOCH from a.created_at) * 1000)::int8 as created_at,
           'TODO' as destination,
           a.weight,
           a.originate_at,
           a.answered_at,
           a.bridged_at,
           a.hangup_at,
           call_center.cc_get_lookup(cor.id, cor.name) as resource,
           leg_a_id,
           leg_b_id,
           node_id as node,
           result,
           call_center.cc_get_lookup(u.id, u.name) as agent,
           call_center.cc_get_lookup(cb.id::int8, cb.name::varchar) as bucket,
           logs,
           false as active
    from call_center.cc_member_attempt_log a
        left join call_center.cc_outbound_resource cor on a.resource_id = cor.id
        left join call_center.cc_agent ca on a.agent_id = ca.id
        left join directory.wbt_user u on u.id = ca.user_id
        left join call_center.cc_bucket cb on a.bucket_id = cb.id
    where a.member_id = :MemberId
    order by a.created_at
)
select *
from active a
union all
select *
from log a`, map[string]interface{}{"MemberId": memberId}); err != nil {
		return nil, model.NewCustomCodeError("store.sql_member.get_attempts_all.app_error", fmt.Sprintf("MemberId=%v, %s", memberId, err.Error()), extractCodeFromErr(err))
	}

	return attempts, nil
}

func (s SqlMemberStore) SearchAttemptsHistory(ctx context.Context, domainId int64, search *model.SearchAttempts) ([]*model.AttemptHistory, model.AppError) {
	var att []*model.AttemptHistory

	f := map[string]interface{}{
		"Domain":          domainId,
		"Q":               search.GetQ(),
		"Limit":           search.GetLimit(),
		"Offset":          search.GetOffset(),
		"From":            model.GetBetweenFromTime(search.JoinedAt),
		"To":              model.GetBetweenToTime(search.JoinedAt),
		"Ids":             pq.Array(search.Ids),
		"QueueIds":        pq.Array(search.QueueIds),
		"BucketIds":       pq.Array(search.BucketIds),
		"MemberIds":       pq.Array(search.MemberIds),
		"AgentIds":        pq.Array(search.AgentIds),
		"OfferedAgentIds": pq.Array(search.OfferedAgentIds),
		"Result":          pq.Array(search.Result),
		"OfferingFrom":    model.GetBetweenFromTime(search.OfferingAt),
		"OfferingTo":      model.GetBetweenToTime(search.OfferingAt),
		"LeavingFrom":     model.GetBetweenFromTime(search.LeavingAt),
		"LeavingTo":       model.GetBetweenToTime(search.LeavingAt),
		"DurationFrom":    model.GetBetweenFrom(search.Duration),
		"DurationTo":      model.GetBetweenTo(search.Duration),
	}

	err := s.ListQuery(ctx, &att, search.ListRequest,
		`domain_id = :Domain
	and joined_at between :From::timestamptz and :To::timestamptz
	and (:Ids::int8[] isnull or id = any(:Ids))
	and (:QueueIds::int[] isnull or queue_id = any(:QueueIds) )
	and (:BucketIds::int8[] isnull or bucket_id = any(:Ids))
	and (:MemberIds::int8[] isnull or member_id = any(:MemberIds) )
	and (:AgentIds::int[] isnull or agent_id = any(:AgentIds) )
	and (:OfferedAgentIds::int[] isnull or offered_agent_ids && :OfferedAgentIds::int[] )
	
	and ( :OfferingFrom::timestamptz isnull or offering_at >= :OfferingFrom::timestamptz )
	and ( :OfferingTo::timestamptz isnull or offering_at <= :OfferingTo::timestamptz )
 
	and ( :LeavingFrom::timestamptz isnull or leaving_at >= :LeavingFrom::timestamptz )
	and ( :LeavingTo::timestamptz isnull or leaving_at <= :LeavingTo::timestamptz )
 
	and ( :LeavingFrom::timestamptz isnull or leaving_at >= :LeavingFrom::timestamptz )
	and ( :LeavingTo::timestamptz isnull or leaving_at <= :LeavingTo::timestamptz )
 
	and ( :DurationFrom::int8 isnull or extract(epoch from coalesce(reporting_at, leaving_at) - joined_at)::int8 >= :DurationFrom::int8 )
	and ( :DurationTo::int8 isnull or extract(epoch from coalesce(reporting_at, leaving_at) - joined_at)::int8 <= :DurationTo::int8 )

	and (:Result::varchar[] isnull or result = any(:Result) )
	and (:Q::varchar isnull or ( destination->>'destination' ilike :Q or destination->>'display' ilike :Q))
`,
		model.AttemptHistory{}, f)
	if err != nil {
		return nil, model.NewCustomCodeError("store.sql_member.attempts_history.app_error", err.Error(), extractCodeFromErr(err))
	}

	return att, nil
}

func (s SqlMemberStore) SearchAttempts(ctx context.Context, domainId int64, search *model.SearchAttempts) ([]*model.Attempt, model.AppError) {
	var att []*model.Attempt

	f := map[string]interface{}{
		"Domain":       domainId,
		"Q":            search.GetQ(),
		"Limit":        search.GetLimit(),
		"Offset":       search.GetOffset(),
		"From":         model.GetBetweenFromTime(search.JoinedAt),
		"To":           model.GetBetweenToTime(search.JoinedAt),
		"Ids":          pq.Array(search.Ids),
		"QueueIds":     pq.Array(search.QueueIds),
		"BucketIds":    pq.Array(search.BucketIds),
		"MemberIds":    pq.Array(search.MemberIds),
		"AgentIds":     pq.Array(search.AgentIds),
		"Result":       pq.Array(search.Result),
		"OfferingFrom": model.GetBetweenFrom(search.OfferingAt),
		"OfferingTo":   model.GetBetweenTo(search.OfferingAt),
		"LeavingFrom":  model.GetBetweenFrom(search.LeavingAt),
		"LeavingTo":    model.GetBetweenTo(search.LeavingAt),
		"DurationFrom": model.GetBetweenFrom(search.Duration),
		"DurationTo":   model.GetBetweenTo(search.Duration),
	}

	err := s.ListQueryMaster(ctx, &att, search.ListRequest,
		`domain_id = :Domain
	and ( :From::timestamptz isnull or joined_at_timestamp >= :From::timestamptz )
	and ( :To::timestamptz isnull or joined_at_timestamp <= :To::timestamptz )

	and (:Ids::int8[] isnull or id = any(:Ids))
	and (:QueueIds::int[] isnull or queue_id = any(:QueueIds) )
	and (:BucketIds::int8[] isnull or bucket_id = any(:Ids))
	and (:MemberIds::int8[] isnull or member_id = any(:MemberIds) )
	and (:AgentIds::int[] isnull or agent_id = any(:AgentIds) )
 
	and ( :OfferingFrom::int8 isnull or offering_at >= :OfferingFrom::int8 )
	and ( :OfferingTo::int8 isnull or offering_at <= :OfferingTo::int8 )
 
	and ( :LeavingFrom::int8 isnull or leaving_at >= :LeavingFrom::int8 )
	and ( :LeavingTo::int8 isnull or leaving_at <= :LeavingTo::int8 )
 
	and ( :LeavingFrom::int8 isnull or leaving_at >= :LeavingFrom::int8 )
	and ( :LeavingTo::int8 isnull or leaving_at <= :LeavingTo::int8 )
 
	and ( :DurationFrom::int8 isnull or (extract(epoch from now()) - (joined_at/1000))::int8 >= :DurationFrom::int8 )
	and ( :DurationTo::int8 isnull or (extract(epoch from now()) - (joined_at/1000))::int8 <= :DurationTo::int8 )

	and (:Result::varchar[] isnull or result = any(:Result) )
	and (:Q::varchar isnull or ( destination->>'destination' ilike :Q or destination->>'display' ilike :Q))`,
		model.Attempt{}, f)
	if err != nil {
		return nil, model.NewCustomCodeError("store.sql_member.attempts.app_error", err.Error(), extractCodeFromErr(err))
	}

	return att, nil
}

func (s SqlMemberStore) ListOfflineQueueForAgent(ctx context.Context, domainId int64, search *model.SearchOfflineQueueMembers) ([]*model.OfflineMember, model.AppError) {
	var att []*model.OfflineMember
	_, err := s.GetMaster().WithContext(ctx).Select(&att, `with comm as (
    select c.id, json_build_object('id', c.id, 'name',  c.name)::jsonb j
    from call_center.cc_communication c
    where c.domain_id = :Domain
)

,resources as (
    select r.id, json_build_object('id', r.id, 'name',  r.name)::jsonb j
    from call_center.cc_outbound_resource r
    where r.domain_id = :Domain
)
, result as (
	select x as id
	from call_center.cc_offline_members_ids(:Domain::int8, :AgentId::int, :Limit::int) x
)
select m.id, call_center.cc_member_destination_views_to_json(array(select ( xid::int2, x ->> 'destination',
								resources.j,
                                comm.j,
                                (x -> 'priority')::int ,
                                (x -> 'state')::int  ,
                                x -> 'description'  ,
                                (x -> 'last_activity_at')::int8,
                                (x -> 'attempts')::int,
                                x ->> 'last_cause',
                                x ->> 'display'    )::call_center.cc_member_destination_view
                         from jsonb_array_elements(m.communications) with ordinality as x (x, xid)
                            left join comm on comm.id = (x -> 'type' -> 'id')::int
                            left join resources on resources.id = (x -> 'resource' -> 'id')::int)) communications,
       call_center.cc_get_lookup(cq.id, cq.name::varchar) queue, call_center.cc_view_timestamp(m.expire_at) expire_at, call_center.cc_view_timestamp(m.created_at) created_at,
			m.variables, m.name
from call_center.cc_member m
    inner join result on m.id = result.id
    inner join call_center.cc_queue cq on m.queue_id = cq.id`, map[string]interface{}{
		"Domain":  domainId,
		"Limit":   search.GetLimit(),
		"AgentId": search.AgentId,
	})

	if err != nil {
		return nil, model.NewCustomCodeError("store.sql_member.list_offline_queue.app_error", err.Error(), extractCodeFromErr(err))
	}

	return att, nil
}

func (s SqlMemberStore) ResetActiveMemberAttempts(ctx context.Context, payload *model.ResetActiveMemberAttempts) ([]*model.MemberAttempt, model.AppError) {
	var (
		res []*model.MemberAttempt
	)
	if payload == nil {
		return nil, model.NewBadRequestError("store.sql_member.reset_active_call_attempts.check_args.payload.app_error", "payload required")
	}
	if payload.DomainId <= 0 {
		return nil, model.NewBadRequestError("store.sql_member.reset_active_call_attempts.check_args.domain.app_error", "domain id required")
	}
	if payload.Result == "" {
		return nil, model.NewBadRequestError("store.sql_member.reset_active_call_attempts.check_args.reason.app_error", "reason required")
	}
	if payload.IdleForMinutes <= 0 {
		return nil, model.NewBadRequestError("store.sql_member.reset_active_call_attempts.check_args.idle.app_error", "idle for required")
	}
	if len(payload.AttemptTypes) == 0 {
		return nil, model.NewBadRequestError("store.sql_member.reset_active_call_attempts.check_args.idle.app_error", "attempt type required")
	}
	_, err := s.GetMaster().WithContext(ctx).Select(&res,
		`update call_center.cc_member_attempt
				set state = 'leaving',
					leaving_at = now(),
					result = :Result
				where domain_id = :DomainId
					 and channel = ANY(:Types)
					and now() - joined_at > (:Interval || ' min')::interval
					returning id, node_id as node`,
		map[string]interface{}{
			"Result":   payload.Result,
			"DomainId": payload.DomainId,
			"Interval": payload.IdleForMinutes,
			"Types":    pq.Array(payload.AttemptTypes),
		})

	if err != nil {
		return nil, model.NewCustomCodeError("store.sql_member.reset_active_call_attempts.app_error", err.Error(), extractCodeFromErr(err))
	}

	return res, nil
}

func (s SqlMemberStore) GetAppointmentWidget(ctx context.Context, uri string) (*model.AppointmentWidget, model.AppError) {
	var widget *model.AppointmentWidget
	err := s.GetReplica().WithContext(ctx).SelectOne(&widget, `select profile, list
from call_center.appointment_widget(:Uri::varchar)`, map[string]interface{}{
		"Uri": uri,
	})

	if err != nil {
		return nil, model.NewCustomCodeError("store.sql_member.appointment.widget.app_error", err.Error(), extractCodeFromErr(err))
	}

	widget.InitOrigin()

	return widget, nil
}

func (s SqlMemberStore) GetAppointment(ctx context.Context, memberId int64) (*model.Appointment, model.AppError) {
	var res *model.Appointment

	err := s.GetReplica().WithContext(ctx).SelectOne(&res, `select
    m.id,
    coalesce(m.ready_at at time zone tz.sys_name, m.created_at at time zone tz.sys_name)::date::text as schedule_date,
    to_char(coalesce(m.ready_at at time zone tz.sys_name, m.created_at at time zone tz.sys_name), 'HH24:MI') as schedule_time,
    m.name,
    m.communications[0]->>'destination' as destination,
    m.variables,
    coalesce(m.import_id, '') as import_id,
	tz.sys_name as timezone
from call_center.cc_member m
	left join flow.calendar_timezones tz on tz.id = m.timezone_id
where m.id = :Id and m.stop_at isnull and m.stop_cause isnull`, map[string]interface{}{
		"Id": memberId,
	})

	if err != nil {
		return nil, model.NewCustomCodeError("store.sql_member.appointment.get.app_error", err.Error(), extractCodeFromErr(err))
	}

	return res, nil
}

func (s SqlMemberStore) CreateAppointment(ctx context.Context, profile *model.AppointmentProfile, app *model.Appointment) (*model.Appointment, model.AppError) {
	err := s.GetMaster().WithContext(ctx).SelectOne(&app, `
insert into call_center.cc_member (queue_id, communications, timezone_id, domain_id, variables, ready_at, expire_at, name, import_id)
select :QueueId,
       jsonb_build_array(
           jsonb_build_object('destination', :Destination::varchar, 'type', jsonb_build_object('id', :TypeId::int))
       ),
       :TimezoneId,
       :DomainId,
	   :Vars,	
       (:Date || ' ' || :Time)::timestamp at time zone :TzName,
       ((:Date || ' ' || :Time)::timestamp at time zone :TzName)::date + interval '1d' - interval '1s',
	   :Name,
	   :Ip	
where not exists(select 1 from call_center.cc_member m
          where m.queue_id = :QueueId
            and search_destinations && array[:Destination::varchar]
            and m.stop_at isnull and 1=2
    )
returning call_center.cc_member.id,
    coalesce(call_center.cc_member.ready_at at time zone :TzName, call_center.cc_member.created_at at time zone :TzName)::date::text as schedule_date,
	to_char(coalesce(call_center.cc_member.ready_at at time zone :TzName, call_center.cc_member.created_at), 'HH24:MI') as schedule_time,
    call_center.cc_member.name,
    call_center.cc_member.communications[0]->>'destination' as destination,
	coalesce(call_center.cc_member.import_id, '') as import_id,
    call_center.cc_member.variables`, map[string]interface{}{
		"Destination": app.Destination,
		"TypeId":      profile.CommunicationTypeId,
		"TimezoneId":  profile.TimezoneId,
		"DomainId":    profile.DomainId,
		"Vars":        app.Variables.ToSafeJson(),
		"Date":        app.ScheduleDate,
		"Time":        app.ScheduleTime,
		"Name":        app.Name,
		"QueueId":     profile.QueueId,
		"Ip":          app.Ip,
		"TzName":      profile.Timezone,
	})

	if err != nil {
		return nil, model.NewCustomCodeError("store.sql_member.appointment.create.app_error", err.Error(), extractCodeFromErr(err))
	}

	return app, nil
}

func (s SqlMemberStore) CancelAppointment(ctx context.Context, memberId int64, reason string) model.AppError {
	_, err := s.GetMaster().WithContext(ctx).Exec(`update call_center.cc_member 
set stop_at = now(),
    stop_cause = :Result
where id = :Id`, map[string]interface{}{
		"Id":     memberId,
		"Result": reason,
	})

	if err != nil {
		return model.NewCustomCodeError("store.sql_member.appointment.cancel.app_error", err.Error(), extractCodeFromErr(err))
	}

	return nil
}

func (s SqlMemberStore) QueueId(ctx context.Context, domainId, memberId int64) (int64, model.AppError) {
	queueId, err := s.GetMaster().WithContext(ctx).SelectInt(`select queue_id
from call_center.cc_member
where id = :Id::int8 and domain_id = :DomainId::int8`, map[string]interface{}{
		"Id":       memberId,
		"DomainId": domainId,
	})
	if err != nil {
		return 0, model.NewCustomCodeError("store.sql_member.get_queue.app_error", err.Error(), extractCodeFromErr(err))
	}

	if queueId == 0 {
		return 0, model.NewBadRequestError("store.sql_member.get_queue.not_found", "Not found member or queue")
	}

	return queueId, nil
}
