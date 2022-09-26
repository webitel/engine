package sqlstore

import (
	"database/sql"
	"fmt"
	"github.com/go-gorp/gorp"
	"github.com/lib/pq"
	"github.com/webitel/engine/model"
	"github.com/webitel/engine/store"
	"net/http"
	"strings"
)

type SqlMemberStore struct {
	SqlStore
}

func NewSqlMemberStore(sqlStore SqlStore) store.MemberStore {
	us := &SqlMemberStore{sqlStore}
	return us
}

func (s SqlMemberStore) Create(domainId int64, member *model.Member) (*model.Member, *model.AppError) {
	var out *model.Member
	if err := s.GetMaster().SelectOne(&out, `with m as (
			insert into call_center.cc_member (queue_id, priority, expire_at, variables, name, timezone_id, communications, bucket_id, ready_at, domain_id, agent_id, skill_id)
			values (:QueueId, :Priority, :ExpireAt, :Variables, :Name, :TimezoneId, :Communications, :BucketId, :MinOfferingAt, :DomainId, :AgentId, :SkillId)
			returning *
		)
		select m.id,  m.stop_at, m.stop_cause, m.attempts, m.last_hangup_at, m.created_at, m.queue_id, m.priority, m.expire_at, m.variables, m.name, call_center.cc_get_lookup(ct.id, ct.name) as "timezone",
			   call_center.cc_member_communications(m.communications) as communications,  call_center.cc_get_lookup(qb.id, qb.name::text) as bucket, ready_at,
               call_center.cc_get_lookup(agn.id, agn.name::text) as agent, call_center.cc_get_lookup(cs.id, cs.name::text) as skill
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
		return nil, model.NewAppError("SqlMemberStore.Save", "store.sql_member.save.app_error", nil,
			fmt.Sprintf("name=%v, %v", member.Name, err.Error()), extractCodeFromErr(err))
	} else {
		return out, nil
	}
}

func (s SqlMemberStore) BulkCreate(domainId, queueId int64, fileName string, members []*model.Member) ([]int64, *model.AppError) {
	var err error
	var stmp *sql.Stmt
	var tx *gorp.Transaction
	tx, err = s.GetMaster().Begin()
	if err != nil {
		return nil, model.NewAppError("SqlMemberStore.Save", "store.sql_member.bulk_save.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	_, err = tx.Exec("CREATE temp table cc_member_tmp ON COMMIT DROP as table call_center.cc_member with no data")
	if err != nil {
		return nil, model.NewAppError("SqlMemberStore.Save", "store.sql_member.bulk_save.app_error", nil, err.Error(), http.StatusInternalServerError)
	}
	if fileName == "" {
		fileName = model.NewId()
	}

	stmp, err = tx.Prepare(pq.CopyIn("cc_member_tmp", "id", "queue_id", "priority", "expire_at", "variables", "name",
		"timezone_id", "communications", "bucket_id", "ready_at", "agent_id", "skill_id", "import_id"))
	if err != nil {
		return nil, model.NewAppError("SqlMemberStore.Save", "store.sql_member.bulk_save.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	defer stmp.Close()
	result := make([]int64, 0, len(members))
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

		_, err = tx.Select(&result, `with i as (
			insert into call_center.cc_member(queue_id, priority, expire_at, variables, name, timezone_id, communications, bucket_id, ready_at, domain_id, agent_id, skill_id, import_id)
			select queue_id, priority, expire_at, variables, name, timezone_id, communications, bucket_id, ready_at, :DomainId, agent_id, skill_id, import_id
			from cc_member_tmp
			returning id
		)
		select id from i`, map[string]interface{}{
			"DomainId": domainId,
		})
		if err != nil {
			goto _error
		}
	}

	err = tx.Commit()
	if err != nil {
		return nil, model.NewAppError("SqlMemberStore.Save", "store.sql_member.bulk_save.app_error", nil, err.Error(), extractCodeFromErr(err))
	}

	return result, nil

_error:
	tx.Rollback()
	return nil, model.NewAppError("SqlMemberStore.Save", "store.sql_member.bulk_save.app_error", nil, err.Error(), extractCodeFromErr(err))
}

// todo fix deprecated fields

func (s SqlMemberStore) SearchMembers(domainId int64, search *model.SearchMemberRequest) ([]*model.Member, *model.AppError) {
	var members []*model.Member

	order := GetOrderBy("cc_member", model.MemberDeprecatedField(search.Sort))
	if order == "" {
		order = "order by id desc"
	}

	fields := GetFields(model.MemberDeprecatedFields(search.Fields), model.Member{})

	if _, err := s.GetReplica().Select(&members,
		`with comm as (select c.id, json_build_object('id', c.id, 'name', c.name)::jsonb j
              from call_center.cc_communication c
              where c.domain_id = :Domain)
   , resources as (select r.id, json_build_object('id', r.id, 'name', r.name)::jsonb j
                   from call_center.cc_outbound_resource r
                   where r.domain_id = :Domain)
   , result as (select m.id
                from call_center.cc_member m
                where m.domain_id = :Domain
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

				  and (:AgentIds::int4[] isnull or m.agent_id = any(:AgentIds::int4[]))

                  and (:StopCauses::varchar[] isnull or m.stop_cause = any (:StopCauses::varchar[]))
                  and (:Name::varchar isnull or m.name ilike :Name::varchar)
                  and (:Q::varchar isnull or
                       (m.name ~~ :Q::varchar or m.search_destinations && array [rtrim(:Q::varchar, '%')]::varchar[]))
				`+order+`
                limit :Limit offset :Offset)
	, list as (
		select m.id,
			   call_center.cc_member_destination_views_to_json(array(select (xid::int2, x ->> 'destination',
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
			   call_center.cc_get_lookup(cq.id, cq.name::varchar)                                                                                    queue,
			   m.priority,
			   m.expire_at,
			   m.created_at,
			   m.variables,
			   m.name,
			   call_center.cc_get_lookup(m.timezone_id::bigint,
										 ct.name::varchar)                                                                                           "timezone",
			   call_center.cc_get_lookup(m.bucket_id, cb.name::varchar)                                                                              bucket,
			   m.ready_at as ready_at,
			   m.stop_cause,
			   m.stop_at,
			   m.last_hangup_at as last_hangup_at,
			   m.attempts,
			   call_center.cc_get_lookup(agn.id, agn.name::varchar)                                                                                  agent,
			   call_center.cc_get_lookup(cs.id, cs.name::varchar)                                                                                    skill,
			   exists(select 1 from call_center.cc_member_attempt a where a.member_id = m.id) as                                                     reserved
		from call_center.cc_member m
				 inner join result on m.id = result.id
				 inner join call_center.cc_queue cq on m.queue_id = cq.id
				 left join flow.calendar_timezones ct on ct.id = m.timezone_id
				 left join call_center.cc_agent_list agn on m.agent_id = agn.id
				 left join call_center.cc_bucket cb on m.bucket_id = cb.id
				 left join call_center.cc_skill cs on m.skill_id = cs.id
	)
	select `+strings.Join(fields, " ,")+` from list`, map[string]interface{}{
			"Domain": domainId,
			"Limit":  search.GetLimit(),
			"Offset": search.GetOffset(),
			"Q":      search.GetRegExpQ(),

			"Ids":         pq.Array(search.Ids),
			"QueueIds":    pq.Array(search.QueueIds),
			"BucketIds":   pq.Array(search.BucketIds),
			"AgentIds":    pq.Array(search.AgentIds),
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
		}); err != nil {
		return nil, model.NewAppError("SqlMemberStore.GetAllPage", "store.sql_member.get_all.app_error", nil, err.Error(), http.StatusInternalServerError)
	} else {
		return members, nil
	}
}

func (s SqlMemberStore) Get(domainId, queueId, id int64) (*model.Member, *model.AppError) {
	var member *model.Member
	if err := s.GetReplica().SelectOne(&member, `select m.id,  m.stop_at, m.stop_cause, m.attempts, m.last_hangup_at, m.created_at, m.queue_id, m.priority, m.expire_at, m.variables, m.name, call_center.cc_get_lookup(ct.id, ct.name) as "timezone",
			   call_center.cc_member_communications(m.communications) as communications,  call_center.cc_get_lookup(qb.id, qb.name::text) as bucket, ready_at,
               call_center.cc_get_lookup(cs.id, cs.name::text) as skill, call_center.cc_get_lookup(agn.id, agn.name::varchar) agent
		from call_center.cc_member m
			left join flow.calendar_timezones ct on m.timezone_id = ct.id
			left join call_center.cc_bucket qb on m.bucket_id = qb.id
			left join call_center.cc_agent_list agn on m.agent_id = agn.id
		    left join call_center.cc_skill cs on m.skill_id = cs.id
	where m.id = :Id and m.queue_id = :QueueId and exists(select 1 from call_center.cc_queue q where q.id = :QueueId and q.domain_id = :DomainId)`, map[string]interface{}{
		"Id":       id,
		"DomainId": domainId,
		"QueueId":  queueId,
	}); err != nil {
		return nil, model.NewAppError("SqlMemberStore.Get", "store.sql_member.get.app_error", nil,
			fmt.Sprintf("Id=%v, %s", id, err.Error()), extractCodeFromErr(err))
	} else {
		return member, nil
	}
}

func (s SqlMemberStore) Update(domainId int64, member *model.Member) (*model.Member, *model.AppError) {
	err := s.GetMaster().SelectOne(&member, `with m as (
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
			stop_at = case when :StopCause::varchar notnull then now() else stop_at end
    where m1.id = :Id and m1.queue_id = :QueueId
    returning *
)
select m.id,  m.stop_at, m.stop_cause, m.attempts, m.last_hangup_at, m.created_at, m.queue_id, m.priority, m.expire_at, m.variables, m.name, call_center.cc_get_lookup(ct.id, ct.name) as "timezone",
			   call_center.cc_member_communications(m.communications) as communications,  call_center.cc_get_lookup(qb.id, qb.name::text) as bucket, ready_at,
               call_center.cc_get_lookup(cs.id, cs.name::text) as skill, call_center.cc_get_lookup(agn.id, agn.name::varchar) agent
		from m
			left join flow.calendar_timezones ct on m.timezone_id = ct.id
			left join call_center.cc_bucket qb on m.bucket_id = qb.id
			left join call_center.cc_skill cs on m.skill_id = cs.id
			left join call_center.cc_agent_list agn on m.agent_id = agn.id`, map[string]interface{}{
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
	})
	if err != nil {
		code := extractCodeFromErr(err)
		if code == http.StatusNotFound { //todo
			return nil, model.NewAppError("SqlMemberStore.Update", "store.sql_member.update.lock", nil,
				fmt.Sprintf("Id=%v, %s", member.Id, err.Error()), http.StatusBadRequest)
		}

		return nil, model.NewAppError("SqlMemberStore.Update", "store.sql_member.update.app_error", nil,
			fmt.Sprintf("Id=%v, %s", member.Id, err.Error()), code)
	}
	return member, nil
}

//TODO add force
func (s SqlMemberStore) Delete(queueId, id int64) *model.AppError {
	var cnt int64
	res, err := s.GetMaster().Exec(`delete
from call_center.cc_member c
where c.id = :Id
  and c.queue_id = :QueueId
  and not exists(select 1 from call_center.cc_member_attempt a where a.member_id = c.id and a.state != 'leaving' for update)`,
		map[string]interface{}{"Id": id, "QueueId": queueId})

	if err != nil {
		return model.NewAppError("SqlMemberStore.Delete", "store.sql_member.delete.app_error", nil,
			fmt.Sprintf("Id=%v, %s", id, err.Error()), extractCodeFromErr(err))
	}

	cnt, err = res.RowsAffected()
	if err != nil {
		return model.NewAppError("SqlMemberStore.Delete", "store.sql_member.delete.app_error", nil,
			fmt.Sprintf("Id=%v, %s", id, err.Error()), extractCodeFromErr(err))
	}

	if cnt == 0 {
		return model.NewAppError("SqlMemberStore.Delete", "store.sql_member.delete.app_error", nil,
			fmt.Sprintf("Id=%v, not found", id), http.StatusNotFound)
	}

	return nil
}

func (s SqlMemberStore) MultiDelete(search *model.MultiDeleteMembers) ([]*model.Member, *model.AppError) {
	var res []*model.Member

	_, err := s.GetMaster().Select(&res, `with m as (
    delete from call_center.cc_member m
    where (:Ids::int8[] isnull or m.id = any (:Ids::int8[]))
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
    returning *
)
select m.id,  m.stop_at, m.stop_cause, m.attempts, m.last_hangup_at, m.created_at, m.queue_id, m.priority, m.expire_at, m.variables, m.name, call_center.cc_get_lookup(ct.id, ct.name) as "timezone",
			   call_center.cc_member_communications(m.communications) as communications,  call_center.cc_get_lookup(qb.id, qb.name::text) as bucket, ready_at,
               call_center.cc_get_lookup(cs.id, cs.name::text) as skill, call_center.cc_get_lookup(agn.id, agn.name::varchar) agent
		from m
			left join flow.calendar_timezones ct on m.timezone_id = ct.id
			left join call_center.cc_bucket qb on m.bucket_id = qb.id
			left join call_center.cc_skill cs on m.skill_id = cs.id
			left join call_center.cc_agent_list agn on m.agent_id = agn.id`, map[string]interface{}{
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
	})

	if err != nil {
		return nil, model.NewAppError("SqlMemberStore.MultiDelete", "store.sql_member.multi_delete.app_error", nil,
			fmt.Sprintf("Ids=%v, %s", search.Ids, err.Error()), extractCodeFromErr(err))
	}

	return res, nil
}

func (s SqlMemberStore) ResetMembers(domainId int64, req *model.ResetMembers) (int64, *model.AppError) {
	cnt, err := s.GetMaster().SelectInt(`with upd as (
    update call_center.cc_member m
    set stop_cause = null,
        stop_at = null,
        attempts = 0
    where m.domain_id = :DomainId
        and m.queue_id = :QueueId
        and (stop_at notnull and not stop_cause in ('success', 'cancel', 'terminate', 'no_communications') )
        and (:Ids::int8[] isnull or m.id = any(:Ids::int8[]))
        and (:Numbers::varchar[] isnull or search_destinations && :Numbers::varchar[])
        and (:Variables::jsonb isnull or variables @> :Variables::jsonb)
        and (:Buckets::int8[] isnull or m.bucket_id = any(:Buckets::int8[]))
        and (:AgentIds::int4[] isnull or m.agent_id = any(:AgentIds::int4[]))
returning m.id
)
select count(*) cnt
from upd`, map[string]interface{}{
		"DomainId": domainId,
		"Ids":      pq.Array(req.Ids),
		"Buckets":  pq.Array(req.Buckets),
		//"Cause":     pq.Array(req.Causes),
		"AgentIds":  pq.Array(req.AgentIds),
		"Numbers":   pq.Array(req.Numbers),
		"Variables": req.Variables.ToSafeJson(),
		"QueueId":   req.QueueId,
	})

	if err != nil {
		return 0, model.NewAppError("SqlMemberStore.ResetMembers", "store.sql_member.reset.app_error", nil,
			fmt.Sprintf("QueueId=%v, %s", req.QueueId, err.Error()), extractCodeFromErr(err))
	}

	return cnt, nil
}

func (s SqlMemberStore) AttemptsList(memberId int64) ([]*model.MemberAttempt, *model.AppError) {
	var attempts []*model.MemberAttempt
	//FIXME
	if _, err := s.GetReplica().Select(&attempts, `with active as (
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
		return nil, model.NewAppError("SqlMemberStore.AttemptsList", "store.sql_member.get_attempts_all.app_error", nil,
			fmt.Sprintf("MemberId=%v, %s", memberId, err.Error()), extractCodeFromErr(err))
	}

	return attempts, nil
}

func (s SqlMemberStore) SearchAttemptsHistory(domainId int64, search *model.SearchAttempts) ([]*model.AttemptHistory, *model.AppError) {
	var att []*model.AttemptHistory

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
		"OfferingFrom": model.GetBetweenFromTime(search.OfferingAt),
		"OfferingTo":   model.GetBetweenToTime(search.OfferingAt),
		"LeavingFrom":  model.GetBetweenFromTime(search.LeavingAt),
		"LeavingTo":    model.GetBetweenToTime(search.LeavingAt),
		"DurationFrom": model.GetBetweenFrom(search.Duration),
		"DurationTo":   model.GetBetweenTo(search.Duration),
	}

	err := s.ListQuery(&att, search.ListRequest,
		`domain_id = :Domain
	and joined_at between :From::timestamptz and :To::timestamptz
	and (:Ids::int8[] isnull or id = any(:Ids))
	and (:QueueIds::int[] isnull or queue_id = any(:QueueIds) )
	and (:BucketIds::int8[] isnull or bucket_id = any(:Ids))
	and (:MemberIds::int8[] isnull or member_id = any(:MemberIds) )
	and (:AgentIds::int[] isnull or agent_id = any(:AgentIds) )
	
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
		return nil, model.NewAppError("SqlMemberStore.SearchAttemptsHistory", "store.sql_member.attempts_history.app_error", nil,
			err.Error(), extractCodeFromErr(err))
	}

	return att, nil
}

func (s SqlMemberStore) SearchAttempts(domainId int64, search *model.SearchAttempts) ([]*model.Attempt, *model.AppError) {
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

	err := s.ListQuery(&att, search.ListRequest,
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
		return nil, model.NewAppError("SqlMemberStore.SearchAttempts", "store.sql_member.attempts.app_error", nil,
			err.Error(), extractCodeFromErr(err))
	}

	return att, nil
}

func (s SqlMemberStore) ListOfflineQueueForAgent(domainId int64, search *model.SearchOfflineQueueMembers) ([]*model.OfflineMember, *model.AppError) {
	var att []*model.OfflineMember
	_, err := s.GetReplica().Select(&att, `with comm as (
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
		return nil, model.NewAppError("SqlMemberStore.ListOfflineQueueForAgent", "store.sql_member.list_offline_queue.app_error", nil,
			err.Error(), extractCodeFromErr(err))
	}

	return att, nil
}

func (s SqlMemberStore) GetAppointmentWidget(id int) (*model.AppointmentWidget, *model.AppError) {
	var widget *model.AppointmentWidget
	err := s.GetReplica().SelectOne(&widget, `with profile as (
    select 
	   (config['queue']->>'id')::int as queue_id,
       (config['communication_type']->>'id')::int as communication_type,
       (config->>'duration')::interval as duration,
       (config->>'days')::int as days,
       (config->>'available_agents')::int as available_agents,
       string_to_array((b.metadata->>'allow_origin'), ',') as allow_origins,
       q.calendar_id,
	   b.id,
	   b.dc as domain_id,
	   c.timezone_id,
	   tz.sys_name as timezone
    from chat.bot b
		inner join lateral (select (b.metadata->>'appointment')::jsonb as config) as cfx on true
        inner join call_center.cc_queue q on q.id = (config['queue']->>'id')::int
		inner join flow.calendar c on c.id = q.calendar_id
		inner join flow.calendar_timezones tz on tz.id = c.timezone_id
    where b.id = :Id
    limit 1
), d as materialized (
    select  q.queue_id,
            q.duration,
            q.available_agents,
            x,
           (extract(isodow from x::timestamp)  ) - 1 as day,
           dy.*
    from profile  q ,
        flow.calendar_day_range(q.calendar_id, least(q.days, 7)) x
        left join lateral (
            select t.*, tz.sys_name, c.excepts
            from flow.calendar c
                inner join flow.calendar_timezones tz on tz.id = c.timezone_id
                inner join lateral unnest(c.accepts::flow.calendar_accept_time[]) t on true
            where c.id = q.calendar_id
                and not t.disabled
            order by 1 asc
    ) y on y.day = (extract(isodow from x)  ) - 1
    left join lateral (
        select (x + (y.start_time_of_day || 'm')::interval)::timestamp as ss,
            case when date_bin(q.duration, (x + (y.end_time_of_day || 'm')::interval)::timestamp, x::timestamp) < (x + (y.end_time_of_day || 'm')::interval)::timestamp
                then date_bin(q.duration, (x + (y.end_time_of_day || 'm')::interval)::timestamp, x::timestamp) + q.duration
                else date_bin(q.duration, (x + (y.end_time_of_day || 'm')::interval)::timestamp, x::timestamp) end as se
    ) dy on true
)
, min_max as materialized (
    select
        queue_id,
        x,
        duration,
        min(ss)  min_ss,
        max(se)  max_se
    from d
    group by 1, 2, 3
)
,res as materialized (
    select
    mem.*
    from min_max
        left join lateral (
            select
                date_bin(min_max.duration, coalesce(ready_at, created_at), coalesce(ready_at, created_at)::date)::timestamp d,
                count(*) cnt
            from call_center.cc_member m
            where m.stop_at isnull
                and m.queue_id = min_max.queue_id
                and coalesce(ready_at, created_at) between min_max.min_ss and min_max.max_se
            group by 1
        ) mem on true
    where mem notnull
)
, list as (
    select
        d.*,
        res.*,
        xx,
        case when xx < now() or coalesce(res.cnt, 0) >= d.available_agents then false
            else true end as reserved
    from d
        left join generate_series(d.ss, d.se, d.duration) xx on true
        left join res on res.d = xx
    limit 10080
)
, ranges AS (
    select
        to_char(list.x::date,'YYYY-MM-DD')::text as date,
        jsonb_agg(jsonb_build_object('time', to_char(list.xx::time, 'HH24:MI'), 'reserved', list.reserved) order by list.x, list.xx) as times
    from list
    group by 1
)
select
    row_to_json(p) as profile,
    jsonb_agg(row_to_json(r)) as list
from profile p
    left join lateral (
        select *
        from ranges
    ) r on true
group by p`, map[string]interface{}{
		"Id": id,
	})

	if err != nil {
		return nil, model.NewAppError("SqlMemberStore.GetAppointmentWidget", "store.sql_member.appointment.widget.app_error", nil,
			err.Error(), extractCodeFromErr(err))
	}

	widget.InitOrigin()

	return widget, nil
}

func (s SqlMemberStore) GetAppointment(memberId int64) (*model.Appointment, *model.AppError) {
	var res *model.Appointment

	err := s.GetReplica().SelectOne(&res, `select
    m.id,
    coalesce(m.ready_at, m.created_at)::date::text as schedule_date,
    to_char(coalesce(m.ready_at, m.created_at), 'HH24:MI') as schedule_time,
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
		return nil, model.NewAppError("SqlMemberStore.GetAppointment", "store.sql_member.appointment.get.app_error", nil,
			err.Error(), extractCodeFromErr(err))
	}

	return res, nil
}

func (s SqlMemberStore) CreateAppointment(profile *model.AppointmentProfile, app *model.Appointment) (*model.Appointment, *model.AppError) {
	err := s.GetMaster().SelectOne(&app, `
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
    coalesce(call_center.cc_member.ready_at, call_center.cc_member.created_at)::date::text as schedule_date,
    DATE_TRUNC('second', coalesce(call_center.cc_member.ready_at, call_center.cc_member.created_at))::time::text as schedule_time,
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
		return nil, model.NewAppError("SqlMemberStore.CreateAppointment", "store.sql_member.appointment.create.app_error", nil,
			err.Error(), extractCodeFromErr(err))
	}

	return app, nil
}

func (s SqlMemberStore) CancelAppointment(memberId int64, reason string) *model.AppError {
	_, err := s.GetMaster().Exec(`update call_center.cc_member 
set stop_at = now(),
    stop_cause = :Reason
where id = :Id`, map[string]interface{}{
		"Id":     memberId,
		"Reason": reason,
	})

	if err != nil {
		return model.NewAppError("SqlMemberStore.CancelAppointment", "store.sql_member.appointment.cancel.app_error", nil,
			err.Error(), extractCodeFromErr(err))
	}

	return nil
}
