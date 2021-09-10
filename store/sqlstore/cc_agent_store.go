package sqlstore

import (
	"database/sql"
	"fmt"
	"github.com/lib/pq"
	"github.com/webitel/engine/auth_manager"
	"github.com/webitel/engine/model"
	"github.com/webitel/engine/store"
	"net/http"
	"strings"
)

type SqlAgentStore struct {
	SqlStore
}

func NewSqlAgentStore(sqlStore SqlStore) store.AgentStore {
	us := &SqlAgentStore{sqlStore}
	return us
}

func (s SqlAgentStore) CheckAccess(domainId, id int64, groups []int, access auth_manager.PermissionAccess) (bool, *model.AppError) {

	res, err := s.GetReplica().SelectNullInt(`select 1
		where exists(
          select 1
          from call_center.cc_agent_acl a
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

// FIXME
func (s SqlAgentStore) Create(agent *model.Agent) (*model.Agent, *model.AppError) {
	var out *model.Agent
	if err := s.GetMaster().SelectOne(&out, `with a as (
			insert into call_center.cc_agent ( user_id, description, domain_id, created_at, created_by, updated_at, updated_by, progressive_count, greeting_media_id,
				allow_channels, chat_count, supervisor_ids, team_id, region_id, supervisor, auditor_ids)
			values (:UserId, :Description, :DomainId, :CreatedAt, :CreatedBy, :UpdatedAt, :UpdatedBy, :ProgressiveCount, :GreetingMedia,
					:AllowChannels, :ChatCount, :SupervisorIds, :TeamId, :RegionId, :Supervisor, :AuditorIds)
			returning *
		)
	SELECT a.domain_id,
       a.id,
       COALESCE(ct.name::character varying::name, ct.username)::character varying                             AS name,
       a.status,
       a.description,
       (date_part('epoch'::text, a.last_state_change) *
        1000::double precision)::bigint                                                                       AS last_status_change,
       date_part('epoch'::text, now() - a.last_state_change)::bigint                                          AS status_duration,
       a.progressive_count,
       ch.x                                                                                                   AS channel,
       json_build_object('id', ct.id, 'name', COALESCE(ct.name::character varying::name, ct.username))::jsonb AS "user",
       call_center.cc_get_lookup(a.greeting_media_id::bigint, g.name)                                         AS greeting_media,
	   a.allow_channels,
       a.chat_count,
       (SELECT jsonb_agg(sag."user") AS jsonb_agg
        FROM call_center.cc_agent_with_user sag
        WHERE sag.id = any(a.supervisor_ids)) as supervisor,
       (SELECT jsonb_agg(call_center.cc_get_lookup(aud.id, coalesce(aud.name, aud.username))) AS jsonb_agg
        FROM directory.wbt_user aud
WHERE aud.id = any(a.auditor_ids)) as auditor,
	   call_center.cc_get_lookup(t.id, t.name) as team,
	   call_center.cc_get_lookup(r.id, r.name) as region,
       a.supervisor as is_supervisor
FROM a
         LEFT JOIN directory.wbt_user ct ON ct.id = a.user_id
         LEFT JOIN storage.media_files g ON g.id = a.greeting_media_id
         left join call_center.cc_team t on t.id = a.team_id
         left join flow.region r on r.id = a.region_id
         LEFT JOIN LATERAL ( SELECT json_build_object('channel', c.channel, 'online', true, 'state', c.state,
                                                      'joined_at',
                                                      (date_part('epoch'::text, c.joined_at) * 1000::double precision)::bigint) AS x
                             FROM call_center.cc_agent_channel c
                             WHERE c.agent_id = a.id) ch ON true`,
		map[string]interface{}{
			"UserId":           agent.User.Id,
			"Description":      agent.Description,
			"DomainId":         agent.DomainId,
			"CreatedAt":        agent.CreatedAt,
			"CreatedBy":        agent.CreatedBy.Id,
			"UpdatedAt":        agent.UpdatedAt,
			"UpdatedBy":        agent.UpdatedBy.Id,
			"ProgressiveCount": agent.ProgressiveCount,
			"GreetingMedia":    agent.GreetingMediaId(),
			"AllowChannels":    pq.Array(agent.AllowChannels),
			"ChatCount":        agent.ChatCount,
			"SupervisorIds":    pq.Array(model.LookupIds(agent.Supervisor)),
			"TeamId":           agent.Team.GetSafeId(),
			"RegionId":         agent.Region.GetSafeId(),
			"AuditorIds":       pq.Array(model.LookupIds(agent.Auditor)),
			"Supervisor":       agent.IsSupervisor,
		}); err != nil {
		return nil, model.NewAppError("SqlAgentStore.Save", "store.sql_agent.save.app_error", nil,
			fmt.Sprintf("record=%v, %v", agent, err.Error()), http.StatusInternalServerError)
	} else {
		return out, nil
	}
}

func (s SqlAgentStore) GetAllPage(domainId int64, search *model.SearchAgent) ([]*model.Agent, *model.AppError) {
	var agents []*model.Agent

	f := map[string]interface{}{
		"DomainId":      domainId,
		"Ids":           pq.Array(search.Ids),
		"Q":             search.GetQ(),
		"TeamIds":       pq.Array(search.TeamIds),
		"AllowChannels": pq.Array(search.AllowChannels),
		"SupervisorIds": pq.Array(search.SupervisorIds),
		"RegionIds":     pq.Array(search.RegionIds),
		"AuditorIds":    pq.Array(search.AuditorIds),
		"SkillIds":      pq.Array(search.SkillIds),
		"QueueIds":      pq.Array(search.QueueIds),
		"IsSupervisor":  search.IsSupervisor,
		"NotSupervisor": search.NotSupervisor,
		"Extensions":    pq.Array(search.Extensions),
		"UserIds":       pq.Array(search.UserIds),
	}

	err := s.ListQuery(&agents, search.ListRequest,
		`domain_id = :DomainId
				and (:Ids::int[] isnull or id = any(:Ids))
				and (:TeamIds::int[] isnull or team_id = any(:TeamIds))
				and (:AllowChannels::varchar[] isnull or allow_channels && :AllowChannels )
				and (:SupervisorIds::int[] isnull or supervisor_ids && :SupervisorIds)
				and (:RegionIds::int[] isnull or region_id = any(:RegionIds))
				and (:Extensions::varchar[] isnull or extension = any(:Extensions))
				and (:UserIds::int8[] isnull or user_id = any(:UserIds))
				and (:AuditorIds::int8[] isnull or auditor_ids && :AuditorIds)
				and (:QueueIds::int[] isnull or id in (
					select distinct a.id
					from call_center.cc_queue q
						inner join call_center.cc_agent a on a.domain_id = q.domain_id
						inner join call_center.cc_queue_skill qs on qs.queue_id = q.id and qs.enabled
						inner join call_center.cc_skill_in_agent sia on sia.agent_id = a.id and sia.enabled
					where q.id = any(:QueueIds)
						and (q.team_id isnull or a.team_id = q.team_id)
						and qs.skill_id = sia.skill_id and sia.capacity between qs.min_capacity and qs.max_capacity
				))
				and (:IsSupervisor::bool isnull or is_supervisor = :IsSupervisor)
				and (:NotSupervisor::bool isnull or not is_supervisor = :NotSupervisor)
				and (:SkillIds::int[] isnull or exists(select 1 from call_center.cc_skill_in_agent sia where sia.agent_id = t.id and sia.skill_id = any(:SkillIds)))
				and (:Q::varchar isnull or (name ilike :Q::varchar or description ilike :Q::varchar or status ilike :Q::varchar ))`,
		model.Agent{}, f)
	if err != nil {
		return nil, model.NewAppError("SqlAgentStore.GetAllPage", "store.sql_agent.get_all.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	return agents, nil
}

func (s SqlAgentStore) GetAllPageByGroups(domainId int64, groups []int, search *model.SearchAgent) ([]*model.Agent, *model.AppError) {
	var agents []*model.Agent

	f := map[string]interface{}{
		"Groups":        pq.Array(groups),
		"Access":        auth_manager.PERMISSION_ACCESS_READ.Value(),
		"DomainId":      domainId,
		"Ids":           pq.Array(search.Ids),
		"Q":             search.GetQ(),
		"TeamIds":       pq.Array(search.TeamIds),
		"AllowChannels": pq.Array(search.AllowChannels),
		"SupervisorIds": pq.Array(search.SupervisorIds),
		"RegionIds":     pq.Array(search.RegionIds),
		"AuditorIds":    pq.Array(search.AuditorIds),
		"SkillIds":      pq.Array(search.SkillIds),
		"QueueIds":      pq.Array(search.QueueIds),
		"IsSupervisor":  search.IsSupervisor,
		"NotSupervisor": search.NotSupervisor,
		"Extensions":    pq.Array(search.Extensions),
		"UserIds":       pq.Array(search.UserIds),
	}

	err := s.ListQuery(&agents, search.ListRequest,
		`domain_id = :DomainId
				and (:Ids::int[] isnull or id = any(:Ids))
				and (:TeamIds::int[] isnull or team_id = any(:TeamIds))
				and (:AllowChannels::varchar[] isnull or allow_channels && :AllowChannels )
				and (:SupervisorIds::int[] isnull or supervisor_ids && :SupervisorIds)
				and (:RegionIds::int[] isnull or region_id = any(:RegionIds))
				and (:Extensions::varchar[] isnull or extension = any(:Extensions))
				and (:UserIds::int8[] isnull or user_id = any(:UserIds))
				and (:AuditorIds::int8[] isnull or auditor_ids && :AuditorIds)
			    and (:IsSupervisor::bool isnull or is_supervisor = :IsSupervisor)
				and (:NotSupervisor::bool isnull or not is_supervisor = :NotSupervisor)
				and (:QueueIds::int[] isnull or id in (
					select distinct a.id
					from call_center.cc_queue q
						inner join call_center.cc_agent a on a.domain_id = q.domain_id
						inner join call_center.cc_queue_skill qs on qs.queue_id = q.id and qs.enabled
						inner join call_center.cc_skill_in_agent sia on sia.agent_id = a.id and sia.enabled
					where q.id = any(:QueueIds)
						and (q.team_id isnull or a.team_id = q.team_id)
						and qs.skill_id = sia.skill_id and sia.capacity between qs.min_capacity and qs.max_capacity
				))
				and (:SkillIds::int[] isnull or exists(select 1 from call_center.cc_skill_in_agent sia where sia.agent_id = t.id and sia.skill_id = any(:SkillIds)))
				and (:Q::varchar isnull or (name ilike :Q::varchar or description ilike :Q::varchar or status ilike :Q::varchar ))
				and (
					exists(select 1
					  from call_center.cc_agent_acl
					  where call_center.cc_agent_acl.dc = t.domain_id and call_center.cc_agent_acl.object = t.id 
						and call_center.cc_agent_acl.subject = any(:Groups::int[]) and call_center.cc_agent_acl.access&:Access = :Access)
		  		)`,
		model.Agent{}, f)
	if err != nil {
		return nil, model.NewAppError("SqlAgentStore.GetAllPageByGroups", "store.sql_agent.get_all.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	return agents, nil
}

func (s SqlAgentStore) GetActiveTask(domainId, id int64) ([]*model.AgentTask, *model.AppError) {
	var res []*model.AgentTask
	_, err := s.GetReplica().Select(&res, `select a.id as attempt_id,
       a.node_id as app_id,
       a.channel,
       a.queue_id,
       a.member_id,
       a.member_call_id as member_channel_id,
       a.agent_call_id as agent_channel_id,
       destination as communication,
       cq.processing as has_reporting,
       a.state,
       a.agent_id,
       call_center.cc_view_timestamp(a.bridged_at) as bridged_at,
	   call_center.cc_view_timestamp(a.leaving_at) as leaving_at,
       extract(epoch from now() - a.last_state_change )::int as duration
from call_center.cc_member_attempt a
    inner join call_center.cc_agent a2 on a2.id = a.agent_id
    inner join call_center.cc_queue cq on a.queue_id = cq.id
where a.agent_id = :AgentId and a2.domain_id  = :DomainId and a.state != 'leaving'`, map[string]interface{}{
		"AgentId":  id,
		"DomainId": domainId,
	})

	if err != nil {
		return nil, model.NewAppError("SqlAgentStore.GetActiveTask", "store.sql_agent.get_tasks.app_error", nil,
			fmt.Sprintf("Id=%v, %s", id, err.Error()), extractCodeFromErr(err))
	}

	return res, nil
}

func (s SqlAgentStore) Get(domainId int64, id int64) (*model.Agent, *model.AppError) {
	var agent *model.Agent
	if err := s.GetReplica().SelectOne(&agent, `
		SELECT a.domain_id,
			   a.id,
			   COALESCE(ct.name::character varying::name, ct.username)::character varying                             AS name,
			   a.status,
			   a.description,
			   (date_part('epoch'::text, a.last_state_change) *
				1000::double precision)::bigint                                                                       AS last_status_change,
			   date_part('epoch'::text, now() - a.last_state_change)::bigint                                          AS status_duration,
			   a.progressive_count,
			   ch.x                                                                                                   AS channel,
			   json_build_object('id', ct.id, 'name', COALESCE(ct.name::character varying::name, ct.username))::jsonb AS "user",
			   call_center.cc_get_lookup(a.greeting_media_id::bigint, g.name)                                         AS greeting_media,
			   a.allow_channels,
			   a.chat_count,
			   (SELECT jsonb_agg(sag."user") AS jsonb_agg
				FROM call_center.cc_agent_with_user sag
				WHERE sag.id = any(a.supervisor_ids)) as supervisor,
			   (SELECT jsonb_agg(call_center.cc_get_lookup(aud.id, coalesce(aud.name, aud.username))) AS jsonb_agg
				FROM directory.wbt_user aud
				WHERE aud.id = any(a.auditor_ids)) as auditor,
			   call_center.cc_get_lookup(t.id, t.name) as team,
			   call_center.cc_get_lookup(r.id, r.name) as region,
			   a.supervisor as is_supervisor
		FROM call_center.cc_agent a
				 LEFT JOIN directory.wbt_user ct ON ct.id = a.user_id
				 LEFT JOIN storage.media_files g ON g.id = a.greeting_media_id
				 left join call_center.cc_team t on t.id = a.team_id
				 left join flow.region r on r.id = a.region_id
				 LEFT JOIN LATERAL ( SELECT json_build_object('channel', c.channel, 'online', true, 'state', c.state,
															  'joined_at',
															  (date_part('epoch'::text, c.joined_at) * 1000::double precision)::bigint) AS x
									 FROM call_center.cc_agent_channel c
									 WHERE c.agent_id = a.id) ch ON true
				where a.domain_id = :DomainId and a.id = :Id 	
		`, map[string]interface{}{"Id": id, "DomainId": domainId}); err != nil {
		if err == sql.ErrNoRows {
			return nil, model.NewAppError("SqlAgentStore.Get", "store.sql_agent.get.app_error", nil,
				fmt.Sprintf("Id=%v, %s", id, err.Error()), http.StatusNotFound)
		} else {
			return nil, model.NewAppError("SqlAgentStore.Get", "store.sql_agent.get.app_error", nil,
				fmt.Sprintf("Id=%v, %s", id, err.Error()), http.StatusInternalServerError)
		}
	} else {
		return agent, nil
	}
}

func (s SqlAgentStore) Update(agent *model.Agent) (*model.Agent, *model.AppError) {
	err := s.GetMaster().SelectOne(&agent, `with a as (
			update call_center.cc_agent
			set user_id = :UserId,
				description = :Description,
				updated_at = :UpdatedAt,
				updated_by = :UpdatedBy,
			    progressive_count = :ProgressiveCount,
			    greeting_media_id = :GreetingMediaId,
				allow_channels = :AllowChannels,
				chat_count = :ChatCount,
				supervisor_ids = :SupervisorIds,
				team_id = :TeamId,
				region_id = :RegionId,
				supervisor = :Supervisor,
				auditor_ids = :AuditorIds
			where id = :Id and domain_id = :DomainId
			returning *
		)
		SELECT a.domain_id,
			   a.id,
			   COALESCE(ct.name::character varying::name, ct.username)::character varying                             AS name,
			   a.status,
			   a.description,
			   (date_part('epoch'::text, a.last_state_change) *
				1000::double precision)::bigint                                                                       AS last_status_change,
			   date_part('epoch'::text, now() - a.last_state_change)::bigint                                          AS status_duration,
			   a.progressive_count,
			   ch.x                                                                                                   AS channel,
			   json_build_object('id', ct.id, 'name', COALESCE(ct.name::character varying::name, ct.username))::jsonb AS "user",
			   call_center.cc_get_lookup(a.greeting_media_id::bigint, g.name)                                         AS greeting_media,
			   a.allow_channels,
			   a.chat_count,
			   (SELECT jsonb_agg(sag."user") AS jsonb_agg
				FROM call_center.cc_agent_with_user sag
				WHERE sag.id = any(a.supervisor_ids)) as supervisor,
			   (SELECT jsonb_agg(call_center.cc_get_lookup(aud.id, coalesce(aud.name, aud.username))) AS jsonb_agg
				FROM directory.wbt_user aud
				WHERE aud.id = any(a.auditor_ids)) as auditor,
			   call_center.cc_get_lookup(t.id, t.name) as team,
			   call_center.cc_get_lookup(r.id, r.name) as region,
			   a.supervisor as is_supervisor
		FROM  a
				 LEFT JOIN directory.wbt_user ct ON ct.id = a.user_id
				 LEFT JOIN storage.media_files g ON g.id = a.greeting_media_id
				 left join call_center.cc_team t on t.id = a.team_id
				 left join flow.region r on r.id = a.region_id
				 LEFT JOIN LATERAL ( SELECT json_build_object('channel', c.channel, 'online', true, 'state', c.state,
															  'joined_at',
															  (date_part('epoch'::text, c.joined_at) * 1000::double precision)::bigint) AS x
									 FROM call_center.cc_agent_channel c
									 WHERE c.agent_id = a.id) ch ON true`, map[string]interface{}{
		"UserId":           agent.User.Id,
		"Description":      agent.Description,
		"ProgressiveCount": agent.ProgressiveCount,
		"Id":               agent.Id,
		"DomainId":         agent.DomainId,
		"UpdatedAt":        agent.UpdatedAt,
		"UpdatedBy":        agent.UpdatedBy.Id,
		"GreetingMediaId":  agent.GreetingMediaId(),
		"AllowChannels":    pq.Array(agent.AllowChannels),
		"ChatCount":        agent.ChatCount,
		"SupervisorIds":    pq.Array(model.LookupIds(agent.Supervisor)),
		"TeamId":           agent.Team.GetSafeId(),
		"RegionId":         agent.Region.GetSafeId(),
		"AuditorIds":       pq.Array(model.LookupIds(agent.Auditor)),
		"Supervisor":       agent.IsSupervisor,
	})
	if err != nil {
		code := http.StatusInternalServerError
		switch err.(type) {
		case *pq.Error:
			if err.(*pq.Error).Code == ForeignKeyViolationErrorCode {
				code = http.StatusBadRequest
			}
		}

		return nil, model.NewAppError("SqlAgentStore.Update", "store.sql_agent.update.app_error", nil,
			fmt.Sprintf("Id=%v, %s", agent.Id, err.Error()), code)
	}
	return agent, nil
}

func (s SqlAgentStore) Delete(domainId, id int64) *model.AppError {
	if _, err := s.GetMaster().Exec(`delete from call_center.cc_agent c where c.id=:Id and c.domain_id = :DomainId`,
		map[string]interface{}{"Id": id, "DomainId": domainId}); err != nil {
		return model.NewAppError("SqlAgentStore.Delete", "store.sql_agent.delete.app_error", nil,
			fmt.Sprintf("Id=%v, %s", id, err.Error()), http.StatusInternalServerError)
	}
	return nil
}

func (s SqlAgentStore) SetStatus(domainId, agentId int64, status string, payload interface{}) (bool, *model.AppError) {
	if r, err := s.GetMaster().Exec(`update call_center.cc_agent
			set status = :Status
  			,status_payload = :Payload
			where id = :AgentId and domain_id = :DomainId and (status <> :Status or status_payload <> :Payload)`, map[string]interface{}{"AgentId": agentId, "Status": status, "Payload": payload, "DomainId": domainId}); err != nil {
		return false, model.NewAppError("SqlAgentStore.SetStatus", "store.sql_agent.set_status.app_error", nil,
			fmt.Sprintf("AgenetId=%v, %s", agentId, err.Error()), http.StatusInternalServerError)
	} else {
		var cnt int64
		if cnt, err = r.RowsAffected(); err != nil {
			return false, model.NewAppError("SqlAgentStore.SetStatus", "store.sql_agent.set_status.app_error", nil,
				fmt.Sprintf("AgenetId=%v, %s", agentId, err.Error()), http.StatusInternalServerError)
		}
		return cnt > 0, nil
	}
}

func (s SqlAgentStore) InQueue(domainId, id int64, search *model.SearchAgentInQueue) ([]*model.AgentInQueue, *model.AppError) {
	var res []*model.AgentInQueue

	f := map[string]interface{}{
		"DomainId": domainId,
		"AgentId":  id,
		"Q":        search.GetQ(),
	}

	err := s.ListQuery(&res, search.ListRequest,
		`domain_id = :DomainId
				and agent_id = :AgentId
				and (:Q::varchar isnull or (queue_name ilike :Q::varchar ))`,
		model.AgentInQueue{}, f)

	if err != nil {
		return nil, model.NewAppError("SqlAgentStore.InQueue", "store.sql_agent.get_queue.app_error", nil,
			fmt.Sprintf("Id=%v, %s", id, err.Error()), extractCodeFromErr(err))
	}

	return res, nil
}

func (s SqlAgentStore) QueueStatistic(domainId, agentId int64) ([]*model.AgentInQueueStatistic, *model.AppError) {
	var stats []*model.AgentInQueueStatistic
	_, err := s.GetReplica().Select(&stats, `select
    call_center.cc_get_lookup(q.id, q.name) queue,
    json_agg(json_build_object(
        'bucket', call_center.cc_get_lookup(m.bucket_id, b.name::text),
        'skill', call_center.cc_get_lookup(s.id, s.name),
        'member_waiting', m.cnt
   ) order by m.bucket_id nulls last, m.skill_id nulls last ) as statistics
from (
    select x.queue_id,
       x.agent_id,
       array_agg(distinct x.b) filter ( where x.b notnull ) buckets,
       array_agg(distinct x.skill_id) filter ( where x.skill_id notnull ) skills
    from (
        SELECT qs.queue_id, csia.agent_id, b, qs.skill_id
        FROM call_center.cc_queue_skill qs
               JOIN call_center.cc_skill_in_agent csia ON csia.skill_id = qs.skill_id
                left join unnest(qs.bucket_ids) b on true
        WHERE qs.enabled
        AND csia.enabled
        AND csia.agent_id = :AgentId
        AND csia.capacity >= qs.min_capacity
        AND csia.capacity <= qs.max_capacity
    ) x
    group by 1, 2
) x
    inner join lateral (
        select m.bucket_id,
               m.skill_id,
               count(*) cnt
        from call_center.cc_member m
        where m.stop_at isnull
            and m.queue_id = x.queue_id
            and (m.ready_at isnull or m.ready_at < now())
			and(m.expire_at isnull or m.expire_at > now())
            and (m.bucket_id isnull or m.bucket_id = any(x.buckets))
            and (m.skill_id isnull or  m.skill_id = any(x.skills))
            and(m.agent_id isnull or m.agent_id = x.agent_id)
        group by 1,2
    ) m on true
    left join call_center.cc_bucket b on b.id = m.bucket_id
    left join call_center.cc_skill s on s.id = m.skill_id
    inner join call_center.cc_queue q on q.id = x.queue_id
where q.domain_id = :DomainId
group by q.id`, map[string]interface{}{
		"AgentId":  agentId,
		"DomainId": domainId,
	})

	if err != nil {
		return nil, model.NewAppError("SqlAgentStore.QueueStatistic", "store.sql_agent.get_queue_stats.app_error", nil,
			fmt.Sprintf("AgentId=%v, %s", agentId, err.Error()), extractCodeFromErr(err))
	}

	return stats, nil
}

func (s SqlAgentStore) HistoryState(domainId int64, search *model.SearchAgentState) ([]*model.AgentState, *model.AppError) {
	var res []*model.AgentState

	//fixme
	order := GetOrderBy("cc_agent_state_history", search.Sort)
	if order == "" {
		order = "order by joined_at desc"
	}

	_, err := s.GetReplica().Select(&res, `with ags as (
 select distinct a.id, call_center.cc_get_lookup(a.id, coalesce(u.name, u.username)) agent
 from call_center.cc_agent a
    inner join directory.wbt_user u on u.id = a.user_id
 where a.domain_id = :DomainId

)
select
    h.id,
    h.channel,
    ags.agent,
    h.joined_at,
    extract(epoch  from duration)::int8 duration,
    h.state,
    h.payload
from call_center.cc_agent_state_history h
    inner join ags on ags.id = h.agent_id
where (:From::timestamp isnull or h.joined_at between :From and :To) 
  and (:AgentIds::int[] isnull or h.agent_id = any(:AgentIds))
  and (:FromId::int8 isnull or h.id > :FromId::int8)
`+order+`
limit :Limit
offset :Offset`, map[string]interface{}{
		"DomainId": domainId,
		"From":     model.GetBetweenFromTime(&search.JoinedAt),
		"To":       model.GetBetweenToTime(&search.JoinedAt),
		"AgentIds": pq.Array(search.AgentIds),
		"FromId":   search.FromId,
		"Limit":    search.GetLimit(),
		"Offset":   search.GetOffset(),
	})

	if err != nil {
		return nil, model.NewAppError("SqlAgentStore.HistoryState", "store.sql_agent.get_state_history.app_error", nil,
			err.Error(), extractCodeFromErr(err))
	}

	return res, nil
}

func (s SqlAgentStore) LookupNotExistsUsers(domainId int64, search *model.SearchAgentUser) ([]*model.AgentUser, *model.AppError) {
	var users []*model.AgentUser

	if _, err := s.GetReplica().Select(&users,
		`select u.id, coalesce( (u.name)::varchar, u.username) as name
from directory.wbt_user u
where u.dc = :DomainId
  and not exists(select 1 from call_center.cc_agent a where a.domain_id = :DomainId and a.user_id = u.id)
  and   ( (:Q::varchar isnull or (coalesce( (u.name)::varchar, u.username) ilike :Q::varchar ) ))
order by coalesce( (u.name)::varchar, u.username) 
limit :Limit
offset :Offset`, map[string]interface{}{
			"DomainId": domainId,
			"Limit":    search.GetLimit(),
			"Offset":   search.GetOffset(),
			"Q":        search.GetQ(),
		}); err != nil {
		return nil, model.NewAppError("SqlAgentStore.LookupNotExistsUsers", "store.sql_agent.lookup.users.app_error", nil, err.Error(), extractCodeFromErr(err))
	} else {
		return users, nil
	}
}

func (s SqlAgentStore) LookupNotExistsUsersByGroups(domainId int64, groups []int, search *model.SearchAgentUser) ([]*model.AgentUser, *model.AppError) {
	var users []*model.AgentUser

	if _, err := s.GetReplica().Select(&users,
		`select u.id, coalesce( (u.name)::varchar, u.username) as name
from directory.wbt_user u
where u.dc = :DomainId
  and not exists(select 1 from call_center.cc_agent a where a.domain_id = :DomainId and a.user_id = u.id)
  and (
	exists(select 1
	  from directory.wbt_auth_acl acl
	  where acl.dc = u.dc and acl.object = u.id and acl.subject = any(:Groups::int[]) and acl.access&:Access = :Access)
  ) 
  and   ( (:Q::varchar isnull or (coalesce( (u.name)::varchar, u.username) ilike :Q::varchar ) ))
order by coalesce( (u.name)::varchar, u.username) 
limit :Limit
offset :Offset`, map[string]interface{}{
			"DomainId": domainId,
			"Limit":    search.GetLimit(),
			"Offset":   search.GetOffset(),
			"Q":        search.GetQ(),
			"Groups":   pq.Array(groups),
			"Access":   auth_manager.PERMISSION_ACCESS_READ.Value(),
		}); err != nil {
		return nil, model.NewAppError("SqlAgentStore.LookupNotExistsUsers", "store.sql_agent.lookup.users.app_error", nil, err.Error(), extractCodeFromErr(err))
	} else {
		return users, nil
	}
}

func (s SqlAgentStore) GetSession(domainId, userId int64) (*model.AgentSession, *model.AppError) {
	var agent *model.AgentSession
	err := s.GetMaster().SelectOne(&agent, `select a.id as agent_id,
       a.status,
       a.status_payload,
       (extract(EPOCH from last_state_change) * 1000)::int8 last_status_change,
       (extract(EPOCH from now() - last_state_change) )::int8 status_duration,
       ch.x as channels,
       a.on_demand,
       call_center.cc_get_lookup(t.id, t.name) team,
       a.supervisor is_supervisor,
       exists(select 1 from call_center.cc_team tm where tm.domain_id = a.domain_id and tm.admin_ids && array[a.id]) is_admin,
       (SELECT jsonb_agg(sag."user") AS jsonb_agg
        FROM call_center.cc_agent_with_user sag
        WHERE sag.id = any(a.supervisor_ids)) supervisor,
       (SELECT jsonb_agg(call_center.cc_get_lookup(aud.id, coalesce(aud.name, aud.username))) AS jsonb_agg
        FROM directory.wbt_user aud
        WHERE aud.id = any(a.auditor_ids)) auditor
from call_center.cc_agent a
	 left join call_center.cc_team t on t.id = a.team_id
     LEFT JOIN LATERAL ( SELECT json_build_array(json_build_object('channel', c.channel, 'state', c.state, 'open', 0, 'max_open', c.max_opened,
                                           'no_answer', c.no_answers,
                                           'wrap_time_ids', (select array_agg(att.id)
                                                from call_center.cc_member_attempt att
                                                where agent_id = a.id
                                                and att.state = 'wrap_time' and att.channel = c.channel),
                                           'joined_at', call_center.cc_view_timestamp(c.joined_at),
                                            'timeout', call_center.cc_view_timestamp(c.timeout))) AS x
                     FROM call_center.cc_agent_channel c
                     WHERE c.agent_id = a.id) ch ON true
where a.user_id = :UserId and a.domain_id = :DomainId`, map[string]interface{}{
		"UserId":   userId,
		"DomainId": domainId,
	})

	if err != nil {
		return nil, model.NewAppError("SqlAgentStore.GetSession", "store.sql_agent.get_session.app_error", nil, err.Error(), extractCodeFromErr(err))
	}
	return agent, nil
}

func (s SqlAgentStore) CallStatistics(domainId int64, search *model.SearchAgentCallStatistics) ([]*model.AgentCallStatistics, *model.AppError) {
	var stats []*model.AgentCallStatistics

	_, err := s.GetReplica().Select(&stats, `select `+strings.Join(GetFields(search.Fields, model.AgentCallStatistics{}), ", ")+`
from (
    select
        coalesce(u.name, u.username) as name,
        coalesce(res.count, 0) as count,
        coalesce(res.abandoned, 0) as abandoned,
        coalesce(res.handles, 0) as handles,
        coalesce(res.sum_talk_sec, 0) as sum_talk_sec,
        coalesce(res.avg_talk_sec, 0) as avg_talk_sec,
        coalesce(res.min_talk_sec, 0) as min_talk_sec,
        coalesce(res.max_talk_sec, 0) as max_talk_sec,
        coalesce(res.sum_hold_sec, 0) as sum_hold_sec,
        coalesce(res.avg_hold_sec, 0) as avg_hold_sec,
        coalesce(res.min_hold_sec, 0) as min_hold_sec,
        coalesce(res.max_hold_sec, 0) as max_hold_sec,
        case when onl_all > 0 then (onl_all / (onl_all + pause_all)) * 100 else 0 end utilization,
        case when onl_all > 0 then (coalesce(work_dur, 0) / (onl_all + pause_all)) else 0 end occupancy
    from (
         select c.agent_id,
               count(*) as count,
               count(*) filter ( where c.answered_at isnull and c.cause in ('NO_ANSWER', 'ORIGINATOR_CANCEL') ) as abandoned, -- todo is missing
               count(*) filter ( where c.answered_at notnull ) as handles,
               extract(epoch from sum(c.hangup_at - c.bridged_at) filter ( where c.bridged_at notnull )) as sum_talk_sec,
               extract(epoch from avg(c.hangup_at - c.bridged_at) filter ( where c.bridged_at notnull )) as avg_talk_sec,
               extract(epoch from min(c.hangup_at - c.bridged_at) filter ( where c.bridged_at notnull )) as min_talk_sec,
               extract(epoch from max(c.hangup_at - c.bridged_at) filter ( where c.bridged_at notnull )) as max_talk_sec,
               sum(c.hold_sec) sum_hold_sec,
               avg(c.hold_sec) avg_hold_sec,
               min(c.hold_sec) min_hold_sec,
               max(c.hold_sec) max_hold_sec
        from call_center.cc_calls_history c
            inner join call_center.cc_member_attempt_history cma on c.attempt_id = cma.id
        where created_at between :From and :To
            and c.domain_id = :DomainId and c.agent_id notnull
			and (:AgentIds::int[] isnull or c.agent_id = any(:AgentIds) )
        group by c.agent_id
    ) res
        inner join call_center.cc_agent a on a.id = res.agent_id
        inner join directory.wbt_user u on u.id = a.user_id
        left join lateral (
             select ares.agent_id,
                    case when l.state = 'online' then l.delta + ares.online else ares.online end    online,
                    case when l.state = 'offline' then l.delta + ares.offline else ares.offline end offline,
                    case when l.state = 'pause' then l.delta + ares.pause else ares.pause end       pause,
                    extract(epoch from (ares.offering + ares.bridged + ares.wrap_time)) as          work_dur,
                    ares.bridged                                                        as          call_time,
                    ares.cnt                                                                        handles,
				    ares.chat_count,
                    ares.missed,
                    ares.max_bridged_at,
                    ares.max_offering_at
             from (
                      select ah.agent_id,
                             coalesce(sum(duration) filter ( where ah.state = 'online' ), interval '0')    online,
                             coalesce(sum(duration) filter ( where ah.state = 'offline' ), interval '0')   offline,
                             coalesce(sum(duration) filter ( where ah.state = 'pause' ), interval '0')     pause,
                             coalesce(sum(duration) filter ( where ah.state = 'bridged' ), interval '0')   bridged,
                             coalesce(sum(duration) filter ( where ah.state = 'offering' ), interval '0')  offering,
                             coalesce(sum(duration) filter ( where ah.state = 'wrap_time' ), interval '0') wrap_time,
                             coalesce(count(*) filter (where ah.state = 'bridged' ), 0)                    cnt,
                             coalesce(count(*) filter (where ah.state = 'chat' ), 0)                    chat_count,
                             coalesce(count(*) filter (where ah.state = 'missed' ), 0)                     missed,
                             max(ah.joined_at) filter ( where ah.state = 'bridged' )                       max_bridged_at,
                             max(ah.joined_at) filter ( where ah.state = 'offering' )                      max_offering_at,
                             min(ah.joined_at)
                      from call_center.cc_agent_state_history ah
                      where ah.joined_at between (:From::timestamptz) and (:To::timestamptz)
                        and ah.agent_id = a.id
                      group by 1
                  ) ares
                      left join lateral (
                 select h2.state,
                        ares.min - (:From::timestamptz) delta
                 from call_center.cc_agent_state_history h2
                 where h2.joined_at < ares.min
                   and h2.agent_id = ares.agent_id
                   and h2.state in ('online', 'offline', 'pause')
                 order by h2.joined_at desc
                 limit 1
                 ) l on true
         ) stat on stat.agent_id = a.id
        left join lateral (select case
                                         when stat isnull or
                                              (now() - a.last_state_change > :To::timestamptz - :From::timestamptz)
                                             then (:To::timestamptz) - (:From::timestamptz)
                                         else now() - a.last_state_change end t) x on true
        left join lateral extract(epoch from coalesce(
         case
             when a.status = 'online' then (x.t + coalesce(stat.online, interval '0'))
             else stat.online end,
         interval '0')) onl_all on true
        left join lateral extract(epoch from coalesce(
         case
             when a.status = 'pause' then (x.t + coalesce(stat.pause, interval '0'))
             else stat.pause end,
         interval '0')) pause_all on true
    where a.domain_id = :DomainId
) agg
limit :Limit
offset :Offset`, map[string]interface{}{
		"DomainId": domainId,
		"Limit":    search.GetLimit(),
		"Offset":   search.GetOffset(),
		"From":     model.GetBetweenFromTime(&search.Time),
		"To":       model.GetBetweenToTime(&search.Time),
		"AgentIds": pq.Array(search.AgentIds),
	})

	if err != nil {
		return nil, model.NewAppError("SqlAgentStore.CallStatistics", "store.sql_agent.get_call_stats.app_error", nil, err.Error(), extractCodeFromErr(err))
	}

	return stats, nil
}

func (s SqlAgentStore) PauseCause(domainId int64, fromUserId, toAgentId int64, allowChange bool) ([]*model.AgentPauseCause, *model.AppError) {
	var res []*model.AgentPauseCause
	_, err := s.GetReplica().Select(&res, `select c.id,
       c.name,
       limit_min,
       (extract(epoch from
           case when a.status = 'pause' and a.status_payload = c.name then
           now() - a.last_state_change + coalesce(tp.duration, interval '0')
           else coalesce(tp.duration, interval '0') end) / 60)::int8 duration_min
from call_center.cc_pause_cause c
         cross join call_center.cc_agent a
         cross join call_center.cc_agent fa
         left join call_center.cc_team ft on ft.id = fa.team_id
         left join call_center.cc_agent_today_pause_cause tp on tp.cause = c.name and tp.id = a.id
where a.id = :ToAgentId and c.domain_id = :DomainId and a.domain_id = c.domain_id
    and fa.user_id = :FromUserId
    and (not :AllowChange::bool   
		 or case when fa.supervisor or fa.id = any(a.supervisor_ids) then c.allow_supervisor else false end
         or (fa.id = a.id and c.allow_agent)
         or (fa.team_id = a.team_id and ft.admin_ids && array[fa.id] and c.allow_admin)
        )
order by c.name;`, map[string]interface{}{
		"DomainId":    domainId,
		"FromUserId":  fromUserId,
		"ToAgentId":   toAgentId,
		"AllowChange": allowChange,
	})

	if err != nil {
		return nil, model.NewAppError("SqlAgentStore.AgentPauseCause", "store.sql_agent.list_pause_causes.app_error", nil, err.Error(), extractCodeFromErr(err))
	}

	return res, nil
}

// FIXME sort, columns
// allow_change
func (s SqlAgentStore) StatusStatistic(domainId int64, supervisorUserId int64, groups []int, access auth_manager.PermissionAccess, search *model.SearchAgentStatusStatistic) ([]*model.AgentStatusStatistics, *model.AppError) {
	var list []*model.AgentStatusStatistics
	_, err := s.GetReplica().Select(&list, `select agent_id,
       name,
       status,
       status_duration,
       "user",
       team,
       online,
       offline,
       pause,
       utilization,
       call_time,
       handles,
       missed,
       max_bridged_at,
       max_offering_at,
       extension,
       queues,
       active_call_id,
       transferred,
       skills,
       supervisor,
       auditor,
       pause_cause,
       chat_count,
       coalesce(occupancy, 0) as occupancy
from (
         select a.id                                                                                  agent_id,
                a.domain_id,
                coalesce(u.name, u.username)                      as                                  name,
                coalesce(u.extension, '')                         as                                  extension,
                a.status,
                extract(epoch from x.t)::int                                                          status_duration,
                coalesce(a.status_payload, '')                                                        pause_cause,
                call_center.cc_get_lookup(u.id, coalesce(u.name, u.username)) as                                  user,

                call_center.cc_get_lookup(team.id, team.name)                                                     team,
                q.queues                                                                              queues,

                onl_all::int                                                                          online,
                extract(epoch from coalesce(
                        case
                            when a.status = 'offline' then (x.t + coalesce(stat.offline, interval '0'))
                            else stat.offline end,
                        interval '0'))::int                                                           offline,
                extract(epoch from coalesce(
                        case
                            when a.status = 'pause' then (x.t + coalesce(stat.pause, interval '0'))
                            else stat.pause end,
                        interval '0'))::int                                                           pause,

                case when onl_all > 0 then (onl_all / (onl_all + pause_all)) * 100 else 0 end         utilization,
                case when onl_all > 0 then (coalesce(work_dur, 0) / (onl_all + pause_all)) else 0 end occupancy,

                coalesce(extract(epoch from call_time)::int8, 0)                                      call_time,
                coalesce(handles, 0)                                                                  handles,
                coalesce(stat.chat_count, 0)                                                          chat_count,
                coalesce(missed, 0)                                                                   missed,
                0::int                                                                                transferred,
                max_bridged_at,
                max_offering_at,
                active_call.id                                    as                                  active_call_id,
                q.skills,
                q.skill_ids,
                (SELECT jsonb_agg(sag."user") AS jsonb_agg
                 FROM call_center.cc_agent_with_user sag
                 WHERE sag.id = any (a.supervisor_ids))                                               supervisor,
                call_center.cc_get_lookup(r.id, r.name)                                                           region,
                (SELECT jsonb_agg(call_center.cc_get_lookup(aud.id, coalesce(aud.name, aud.username))) AS jsonb_agg
                 FROM directory.wbt_user aud
                 WHERE aud.id = any (a.auditor_ids))                                                  auditor,
                queue_ids,
                a.team_id,
                a.auditor_ids,
                a.supervisor_ids,
                a.region_id
         from call_center.cc_agent a
                  inner join directory.wbt_user u on u.id = a.user_id
                  left join call_center.cc_team team on team.id = a.team_id
                  left join flow.region r on r.id = a.region_id
                  left join lateral (
             select array_agg(distinct q.id)                          queue_ids,
                    array_agg(distinct sia.skill_id)                  skill_ids,
                    jsonb_agg(distinct call_center.cc_get_lookup(cs.id, cs.name)) skills,
                    jsonb_agg(distinct call_center.cc_get_lookup(q.id, q.name))   queues
             from call_center.cc_skill_in_agent sia
                      inner join call_center.cc_queue q on sia.agent_id = a.id and sia.enabled
                      inner join call_center.cc_queue_skill qs on qs.queue_id = q.id and qs.enabled
                      inner join call_center.cc_skill cs on sia.skill_id = cs.id
             where q.domain_id = a.domain_id
               and q.enabled
               and (q.team_id isnull or a.team_id = q.team_id)
               and qs.skill_id = sia.skill_id
               and sia.capacity between qs.min_capacity and qs.max_capacity
             ) q on true
                  left join lateral (
             select ares.agent_id,
                    case when l.state = 'online' then l.delta + ares.online else ares.online end    online,
                    case when l.state = 'offline' then l.delta + ares.offline else ares.offline end offline,
                    case when l.state = 'pause' then l.delta + ares.pause else ares.pause end       pause,
                    extract(epoch from (ares.offering + ares.bridged + ares.wrap_time)) as          work_dur,
                    ares.bridged                                                        as          call_time,
                    ares.cnt                                                                        handles,
                    ares.chat_count,
                    ares.missed,
                    ares.max_bridged_at,
                    ares.max_offering_at
             from (
                      select ah.agent_id,
                             coalesce(sum(duration) filter ( where ah.state = 'online' ), interval '0')    online,
                             coalesce(sum(duration) filter ( where ah.state = 'offline' ), interval '0')   offline,
                             coalesce(sum(duration) filter ( where ah.state = 'pause' ), interval '0')     pause,
                             coalesce(sum(duration) filter ( where ah.state = 'bridged' ), interval '0')   bridged,
                             coalesce(sum(duration) filter ( where ah.state = 'offering' ), interval '0')  offering,
                             coalesce(sum(duration) filter ( where ah.state = 'wrap_time' ), interval '0') wrap_time,
                             coalesce(count(*) filter (where ah.state = 'bridged' ), 0)                    cnt,
                             coalesce(count(*) filter (where ah.state = 'chat' ), 0)                       chat_count,
                             coalesce(count(*) filter (where ah.state = 'missed' ), 0)                     missed,
                             max(ah.joined_at) filter ( where ah.state = 'bridged' )                       max_bridged_at,
                             max(ah.joined_at) filter ( where ah.state = 'offering' )                      max_offering_at,
                             min(ah.joined_at)
                      from call_center.cc_agent_state_history ah
                      where ah.joined_at between (:From::timestamptz) and (:To::timestamptz)
                        and ah.agent_id = a.id
                      group by 1
                  ) ares
                      left join lateral (
                 select h2.state,
                        ares.min - (:From::timestamptz) delta
                 from call_center.cc_agent_state_history h2
                 where h2.joined_at < ares.min
                   and h2.agent_id = ares.agent_id
                   and h2.state in ('online', 'offline', 'pause')
                 order by h2.joined_at desc
                 limit 1
                 ) l on true
             ) stat on stat.agent_id = a.id
                  left join lateral (
             select c.id
             from call_center.cc_calls c
             where c.agent_id = a.id
               and c.hangup_at isnull
               and c.direction notnull
             limit 1
             ) active_call on true
                  inner join lateral (select case
                                                 when stat isnull or
                                                      (now() - a.last_state_change > :To::timestamptz - :From::timestamptz)
                                                     then (:To::timestamptz) - (:From::timestamptz)
                                                 else now() - a.last_state_change end t) x on true
                  left join lateral extract(epoch from coalesce(
                 case
                     when a.status = 'online' then (x.t + coalesce(stat.online, interval '0'))
                     else stat.online end,
                 interval '0')) onl_all on true
                  left join lateral extract(epoch from coalesce(
                 case
                     when a.status = 'pause' then (x.t + coalesce(stat.pause, interval '0'))
                     else stat.pause end,
                 interval '0')) pause_all on true
         where a.domain_id = :DomainId
           and a.id in (
                with x as (
                    select a.user_id, a.id agent_id, a.supervisor, a.domain_id
                    from directory.wbt_user u
                             inner join call_center.cc_agent a on a.user_id = u.id
                    where u.id = :UserSupervisorId
                      and u.dc = :DomainId
                )
                select distinct a.id
                from x
                         left join lateral (
                    select a.id, a.auditor_ids && array [x.user_id] aud
                    from call_center.cc_agent a
                    where a.domain_id = x.domain_id
                      and (a.user_id = x.user_id or (a.supervisor_ids && array [x.agent_id]) or
                           a.auditor_ids && array [x.user_id])
                    union
                    distinct
                    select a.id, a.auditor_ids && array [x.user_id] aud
                    from call_center.cc_team t
                             inner join call_center.cc_agent a on a.team_id = t.id
                    where t.admin_ids && array[x.agent_id]
                      and x.domain_id = t.domain_id
                ) a on true
           )
     ) t
where t.domain_id = :DomainId
  and (:AgentIds::int[] isnull or t.agent_id = any (:AgentIds))
  and (:Q::varchar isnull or t.name ilike :Q::varchar)
  and (:Status::varchar[] isnull or (t.status = any (:Status)))
  and ((:UFrom::numeric isnull or t.utilization >= :UFrom::numeric) and
       (:UTo::numeric isnull or t.utilization <= :UTo::numeric))
  and (:QueueIds::int[] isnull or (t.queue_ids notnull and t.queue_ids::int[] && :QueueIds::int[]))
  and (:SkillIds::int[] isnull or (t.skill_ids notnull and t.skill_ids::int[] && :SkillIds::int[]))
  and (:TeamIds::int[] isnull or (t.team_id notnull and t.team_id = any (:TeamIds::int[])))
  and (:RegionIds::int[] isnull or (t.region_id notnull and t.region_id = any (:RegionIds::int[])))
  and (:AuditorIds::int[] isnull or (t.auditor_ids notnull and t.auditor_ids && :AuditorIds::int8[]))
  and (:HasCall::bool isnull or (not :HasCall or active_call_id notnull))
order by case t.status
             when 'break_out' then 0
             when 'pause' then 1
             when 'online' then 2
             else 3 end, t.name
limit :Limit offset :Offset`, map[string]interface{}{
		"DomainId":         domainId,
		"UserSupervisorId": supervisorUserId,
		//"Groups":     pq.Array(groups),
		//"Access":     access.Value(),
		"Q":          search.GetQ(),
		"Limit":      search.GetLimit(),
		"Offset":     search.GetOffset(),
		"From":       model.GetBetweenFromTime(&search.Time),
		"To":         model.GetBetweenToTime(&search.Time),
		"UFrom":      model.GetBetweenFrom(search.Utilization),
		"UTo":        model.GetBetweenTo(search.Utilization),
		"AgentIds":   pq.Array(search.AgentIds),
		"Status":     pq.Array(search.Status),
		"QueueIds":   pq.Array(search.QueueIds),
		"TeamIds":    pq.Array(search.TeamIds),
		"SkillIds":   pq.Array(search.SkillIds),
		"RegionIds":  pq.Array(search.RegionIds),
		"AuditorIds": pq.Array(search.AuditorIds),
		"HasCall":    search.HasCall,
	})

	if err != nil {
		return nil, model.NewAppError("SqlAgentStore.StatusStatistic", "store.sql_agent.get_status_stats.app_error", nil, err.Error(), extractCodeFromErr(err))
	}

	return list, nil
}

func (s SqlAgentStore) SupervisorAgentItem(domainId int64, agentId int64, t *model.FilterBetween) (*model.SupervisorAgentItem, *model.AppError) {
	var item *model.SupervisorAgentItem

	err := s.GetReplica().SelectOne(&item, `select a.id agent_id,
       coalesce(cawu.name, cawu.username) as name,
       call_center.cc_get_lookup(cawu.id, coalesce(cawu.name, cawu.username)) as user,
       coalesce(cawu.extension, '') as extension,
       a.status,
       extract(epoch from x.t)::int status_duration,

       call_center.cc_get_lookup(t.id, t.name) team,
       (SELECT jsonb_agg(sag."user") AS jsonb_agg
        FROM call_center.cc_agent_with_user sag
        WHERE sag.id = any(a.supervisor_ids)) supervisor,
       (SELECT jsonb_agg(call_center.cc_get_lookup(aud.id, coalesce(aud.name, aud.username))) AS jsonb_agg
        FROM directory.wbt_user aud
        WHERE aud.id = any(a.auditor_ids)) as auditor,
       call_center.cc_get_lookup(r.id, r.name) region,
       a.progressive_count,
       a.chat_count,

       (coalesce(extract(epoch from stat.online), 0) + case when a.status = 'online' then  extract(epoch from x.t) else 0 end)::int8 online,
       (coalesce(extract(epoch from stat.offline), 0) + case when a.status = 'offline' then  extract(epoch from x.t) else 0 end)::int8 offline,
       (coalesce(extract(epoch from stat.pause), 0) + case when a.status = 'pause' then  extract(epoch from x.t) else 0 end)::int8 pause,
       coalesce(a.status_payload, '') pause_cause
from call_center.cc_agent a
  left join call_center.cc_team t on t.id = a.team_id
  left join flow.region r on r.id = a.region_id
  left join lateral (
     select
        ah.agent_id,
        coalesce(sum(duration) filter ( where ah.state = 'online' ), interval '0')    online,
        coalesce(sum(duration) filter ( where ah.state = 'offline' ), interval '0')   offline,
        coalesce(sum(duration) filter ( where ah.state = 'pause' ), interval '0')     pause
     from call_center.cc_agent_state_history ah
     where ah.joined_at between (:From::timestamptz) and (:To::timestamptz)
        and ah.agent_id = a.id
     group by 1
  ) stat on true
  inner join lateral (select case
                             when stat isnull or
                                  (now() - a.last_state_change > :To::timestamptz - :From::timestamptz)
                                 then (:To::timestamptz) - (:From::timestamptz)
                             else now() - a.last_state_change end t) x on true
    left join directory.wbt_user cawu on a.user_id = cawu.id
where a.id = :AgentId and a.domain_id = :DomainId`, map[string]interface{}{
		"DomainId": domainId,
		"AgentId":  agentId,
		"From":     model.GetBetweenFromTime(t),
		"To":       model.GetBetweenToTime(t),
	})

	if err != nil {
		return nil, model.NewAppError("SqlAgentStore.SupervisorAgentItem", "store.sql_agent.get_status_stats_item.app_error", nil, err.Error(), extractCodeFromErr(err))
	}

	return item, nil
}
