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
          from cc_agent_acl a
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
			insert into cc_agent ( user_id, description, domain_id, created_at, created_by, updated_at, updated_by, progressive_count, greeting_media_id,
				allow_channels, chat_count, supervisor_id, team_id, region_id, supervisor, auditor_id)
			values (:UserId, :Description, :DomainId, :CreatedAt, :CreatedBy, :UpdatedAt, :UpdatedBy, :ProgressiveCount, :GreetingMedia,
					:AllowChannels, :ChatCount, :SupervisorId, :TeamId, :RegionId, :Supervisor, :AuditorId)
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
       sup.user as supervisor,
       cc_get_lookup(aud.id, coalesce(aud.name, aud.username)) as auditor,
	   cc_get_lookup(t.id, t.name) as team,
	   cc_get_lookup(r.id, r.name) as region,
       a.supervisor as is_supervisor
FROM a
         LEFT JOIN directory.wbt_user ct ON ct.id = a.user_id
         LEFT JOIN storage.media_files g ON g.id = a.greeting_media_id
         left join cc_agent_with_user sup on sup.id = a.supervisor_id
         left join directory.wbt_user aud on aud.id = a.auditor_id
         left join cc_team t on t.id = a.team_id
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
			"SupervisorId":     agent.Supervisor.GetSafeId(),
			"TeamId":           agent.Team.GetSafeId(),
			"RegionId":         agent.Region.GetSafeId(),
			"AuditorId":        agent.Auditor.GetSafeId(),
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
	}

	err := s.ListQuery(&agents, search.ListRequest,
		`domain_id = :DomainId
				and (:Ids::int[] isnull or id = any(:Ids))
				and (:TeamIds::int[] isnull or team_id = any(:TeamIds))
				and (:AllowChannels::varchar[] isnull or allow_channels && :AllowChannels )
				and (:SupervisorIds::int[] isnull or supervisor_id = any(:SupervisorIds))
				and (:RegionIds::int[] isnull or region_id = any(:RegionIds))
				and (:AuditorIds::int[] isnull or auditor_id = any(:AuditorIds))
				and (:QueueIds::int[] isnull or id in (
					select distinct a.id
					from cc_queue q
						inner join cc_agent a on a.team_id = q.team_id
						inner join cc_queue_skill qs on qs.queue_id = q.id and qs.enabled
						inner join cc_skill_in_agent sia on sia.agent_id = a.id and sia.enabled
					where q.id = any(:QueueIds) and qs.skill_id = sia.skill_id and sia.capacity between qs.min_capacity and qs.max_capacity
				))
				and (:IsSupervisor::bool isnull or is_supervisor = :IsSupervisor)
				and (:SkillIds::int[] isnull or exists(select 1 from cc_skill_in_agent sia where sia.agent_id = t.id and sia.skill_id = any(:SkillIds)))
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
	}

	err := s.ListQuery(&agents, search.ListRequest,
		`domain_id = :DomainId
				and (:Ids::int[] isnull or id = any(:Ids))
				and (:TeamIds::int[] isnull or team_id = any(:TeamIds))
				and (:AllowChannels::varchar[] isnull or allow_channels && :AllowChannels )
				and (:SupervisorIds::int[] isnull or supervisor_id = any(:SupervisorIds))
				and (:RegionIds::int[] isnull or region_id = any(:RegionIds))
				and (:AuditorIds::int[] isnull or auditor_id = any(:AuditorIds))
			    and (:IsSupervisor::bool isnull or is_supervisor = :IsSupervisor)
				and (:QueueIds::int[] isnull or id in (
					select distinct a.id
					from cc_queue q
						inner join cc_agent a on a.team_id = q.team_id
						inner join cc_queue_skill qs on qs.queue_id = q.id and qs.enabled
						inner join cc_skill_in_agent sia on sia.agent_id = a.id and sia.enabled
					where q.id = any(:QueueIds) and qs.skill_id = sia.skill_id and sia.capacity between qs.min_capacity and qs.max_capacity
				))
				and (:SkillIds::int[] isnull or exists(select 1 from cc_skill_in_agent sia where sia.agent_id = t.id and sia.skill_id = any(:SkillIds)))
				and (:Q::varchar isnull or (name ilike :Q::varchar or description ilike :Q::varchar or status ilike :Q::varchar ))
				and (
					exists(select 1
					  from cc_agent_acl
					  where cc_agent_acl.dc = t.domain_id and cc_agent_acl.object = t.id and cc_agent_acl.subject = any(:Groups::int[]) and cc_agent_acl.access&:Access = :Access)
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
       cc_view_timestamp(a.bridged_at) as bridged_at,
	   cc_view_timestamp(a.leaving_at) as leaving_at,
       extract(epoch from now() - a.last_state_change )::int as duration
from cc_member_attempt a
    inner join cc_agent a2 on a2.id = a.agent_id
    inner join cc_queue cq on a.queue_id = cq.id
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
			   sup.user as supervisor,
			   cc_get_lookup(aud.id, coalesce(aud.name, aud.username)) as auditor,
			   cc_get_lookup(t.id, t.name) as team,
			   cc_get_lookup(r.id, r.name) as region,
			   a.supervisor as is_supervisor
		FROM call_center.cc_agent a
				 LEFT JOIN directory.wbt_user ct ON ct.id = a.user_id
				 LEFT JOIN storage.media_files g ON g.id = a.greeting_media_id
				 left join cc_agent_with_user sup on sup.id = a.supervisor_id
				 left join directory.wbt_user aud on aud.id = a.auditor_id
				 left join cc_team t on t.id = a.team_id
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
			update cc_agent
			set user_id = :UserId,
				description = :Description,
				updated_at = :UpdatedAt,
				updated_by = :UpdatedBy,
			    progressive_count = :ProgressiveCount,
			    greeting_media_id = :GreetingMediaId,
				allow_channels = :AllowChannels,
				chat_count = :ChatCount,
				supervisor_id = :SupervisorId,
				team_id = :TeamId,
				region_id = :RegionId,
				supervisor = :Supervisor,
				auditor_id = :AuditorId
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
			   sup.user as supervisor,
			   cc_get_lookup(aud.id, coalesce(aud.name, aud.username)) as auditor,
			   cc_get_lookup(t.id, t.name) as team,
			   cc_get_lookup(r.id, r.name) as region,
			   a.supervisor as is_supervisor
		FROM  a
				 LEFT JOIN directory.wbt_user ct ON ct.id = a.user_id
				 LEFT JOIN storage.media_files g ON g.id = a.greeting_media_id
				 left join cc_agent_with_user sup on sup.id = a.supervisor_id
				 left join directory.wbt_user aud on aud.id = a.auditor_id
				 left join cc_team t on t.id = a.team_id
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
		"SupervisorId":     agent.Supervisor.GetSafeId(),
		"TeamId":           agent.Team.GetSafeId(),
		"RegionId":         agent.Region.GetSafeId(),
		"AuditorId":        agent.Auditor.GetSafeId(),
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
	if _, err := s.GetMaster().Exec(`delete from cc_agent c where c.id=:Id and c.domain_id = :DomainId`,
		map[string]interface{}{"Id": id, "DomainId": domainId}); err != nil {
		return model.NewAppError("SqlAgentStore.Delete", "store.sql_agent.delete.app_error", nil,
			fmt.Sprintf("Id=%v, %s", id, err.Error()), http.StatusInternalServerError)
	}
	return nil
}

func (s SqlAgentStore) SetStatus(domainId, agentId int64, status string, payload interface{}) (bool, *model.AppError) {
	if r, err := s.GetMaster().Exec(`update cc_agent
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

	_, err := s.GetReplica().Select(&res, `select cc_get_lookup(q.id, q.name) as                           queue,
       q.priority,
       q.type,
       q.strategy,
       q.enabled,
       coalesce(sum(cqs.member_count), 0)                                 count_members,
       coalesce(sum(cqs.member_waiting), 0)                               waiting_members,
       (select count(*) from cc_member_attempt a where a.queue_id = q.id) active_members
from cc_agent a
         inner join cc_queue q on q.team_id = a.team_id
         left join cc_queue_statistics cqs on q.id = cqs.queue_id
where a.id = :AgentId
  and a.domain_id = :DomainId
  and ((:Q::varchar isnull or (q.name ilike :Q::varchar)))
  and exists(select qs.queue_id
             from cc_queue_skill qs
                      inner join cc_skill_in_agent csia on csia.skill_id = qs.skill_id
             where qs.enabled
               and csia.enabled
               and csia.agent_id = a.id
               and qs.queue_id = q.id
               and csia.capacity between qs.min_capacity and qs.max_capacity)
group by q.id, q.priority
order by q.priority desc
limit :Limit
offset :Offset`, map[string]interface{}{
		"AgentId":  id,
		"DomainId": domainId,
		"Limit":    search.GetLimit(),
		"Offset":   search.GetOffset(),
		"Q":        search.GetQ(),
	})

	if err != nil {
		return nil, model.NewAppError("SqlAgentStore.InQueue", "store.sql_agent.get_queue.app_error", nil,
			fmt.Sprintf("Id=%v, %s", id, err.Error()), extractCodeFromErr(err))
	}

	return res, nil
}

func (s SqlAgentStore) QueueStatistic(domainId, agentId int64) ([]*model.AgentInQueueStatistic, *model.AppError) {
	var stats []*model.AgentInQueueStatistic
	_, err := s.GetReplica().Select(&stats, `select cc_get_lookup(q.id, q.name) queue,
		   json_agg(json_build_object(
				'bucket', cc_get_lookup(t.bucket_id, b.name::text),
				'skill', cc_get_lookup(cqs.skill_id, s.name),
				'member_waiting', cqs.member_waiting
		   ) order by t.bucket_id nulls last, cqs.skill_id nulls last ) as statistics
	from (
			 select at.team_id, x bucket_id
			 from cc_agent_in_team at
					  left join lateral unnest(at.bucket_ids) x on true
			 where at.agent_id = :AgentId
	
			 union all
	
			 select at.team_id, x bucket_id
			 from cc_agent_in_team at
					  left join lateral unnest(at.bucket_ids) x on true
					  inner join cc_skill_in_agent csia on at.skill_id = csia.skill_id
			 where csia.agent_id = :AgentId
			   and csia.capacity between at.min_capacity and at.max_capacity
	) t
		 inner join cc_queue q on q.team_id = t.team_id
		 inner join cc_queue_statistics cqs on
			(cqs.queue_id, coalesce(cqs.bucket_id, 0::bigint)) = (q.id, coalesce(t.bucket_id::bigint, 0::bigint))
		 left join cc_bucket b on b.id = t.bucket_id
		 left join cc_skill s on s.id = cqs.skill_id
	where q.enabled and q.domain_id = :DomainId
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
	order := GetOrderBy(search.Sort)
	if order == "" {
		order = "order by joined_at desc"
	}

	_, err := s.GetReplica().Select(&res, `with ags as (
 select distinct a.id, cc_get_lookup(a.id, coalesce(u.name, u.username)) agent
 from cc_agent a
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
from cc_agent_state_history h
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
  and not exists(select 1 from cc_agent a where a.domain_id = :DomainId and a.user_id = u.id)
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
  and not exists(select 1 from cc_agent a where a.domain_id = :DomainId and a.user_id = u.id)
  and (
	exists(select 1
	  from directory.wbt_auth_acl acl
	  where acl.dc = u.dc and acl.object = u.id and acl.subject = any(:Groups::int[]) and acl.access&:Access = :Access)
  ) 
  and   ( (:Q::varchar isnull or (u.name ilike :Q::varchar ) ))
order by u.id
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
       a.on_demand
from cc_agent a
     LEFT JOIN LATERAL ( SELECT json_build_array(json_build_object('channel', c.channel, 'state', c.state, 'open', 0, 'max_open', c.max_opened,
                                           'no_answer', c.no_answers,
                                           'wrap_time_ids', (select array_agg(att.id)
                                                from cc_member_attempt att
                                                where agent_id = a.id
                                                and att.state = 'wrap_time' and att.channel = c.channel),
                                           'joined_at', cc_view_timestamp(c.joined_at),
                                            'timeout', cc_view_timestamp(c.timeout))) AS x
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
        coalesce(res.max_hold_sec, 0) as max_hold_sec
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
        from cc_calls_history c
            inner join cc_member_attempt_history cma on c.attempt_id = cma.id
        where created_at between :From and :To
            and c.domain_id = :DomainId and c.agent_id notnull
			and (:AgentIds::int[] isnull or c.agent_id = any(:AgentIds) )
        group by c.agent_id
    ) res
        inner join cc_agent a on a.id = res.agent_id
        inner join directory.wbt_user u on u.id = a.user_id
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

// FIXME add RBAC & sort, columns
func (s SqlAgentStore) StatusStatistic(domainId int64, search *model.SearchAgentStatusStatistic) ([]*model.AgentStatusStatistics, *model.AppError) {
	var list []*model.AgentStatusStatistics
	_, err := s.GetReplica().Select(&list, `select     agent_id, name, status, status_duration, "user", teams, online, offline, pause, utilization, call_time, handles, missed,
       max_bridged_at, max_offering_at, extension, queues, active_call_id
from (
    select a.id                                                 agent_id,
       a.domain_id,
       coalesce(u.name, u.username)                      as name,
       coalesce(u.extension, '')                         as extension,
       a.status,
       extract(epoch from x.t)::int                         status_duration,
       cc_get_lookup(u.id, coalesce(u.name, u.username)) as user,

       teams.v                                              teams,
       teams.a                                              queues,

       extract(epoch from coalesce(
               case when a.status = 'online' then (x.t + coalesce(stat.online, interval '0')) else stat.online end,
               interval '0'))::int                          online,
       extract(epoch from coalesce(
               case when a.status = 'offline' then (x.t + coalesce(stat.offline, interval '0')) else stat.offline end,
               interval '0'))::int                          offline,
       extract(epoch from coalesce(
               case when a.status = 'pause' then (x.t + coalesce(stat.pause, interval '0')) else stat.pause end,
               interval '0'))::int                          pause,
       coalesce(utilization, 0)                             utilization,
       coalesce(extract(epoch from call_time)::int8, 0)     call_time,
       coalesce(handles, 0)                                 handles,
       coalesce(missed, 0)                                  missed,
       max_bridged_at,
       max_offering_at,
       active_call.id as active_call_id,
       teams.queue_ids,
       teams.team_ids
from cc_agent a
         inner join directory.wbt_user u on u.id = a.user_id
         LEFT JOIN LATERAL ( select array_agg(distinct t.id)                                                        tt,
                                    json_agg(distinct cc_get_lookup(t.id, t.name))                                  v,
                                    array_agg(distinct t.id) filter ( where t.id notnull ) team_ids,
                                    array_agg(distinct cq.id) filter ( where cq.id notnull ) queue_ids,
                                    json_agg(distinct cc_get_lookup(cq.id, cq.name)) filter ( where cq.id notnull ) a
                             from cc_team t
                                      left join cc_queue cq on t.id = cq.team_id
                             where t.id in (
                                 select distinct ait.team_id
                                 from cc_agent_in_team ait
                                 where ait.agent_id = a.id
                                    or ait.skill_id in (
                                     select distinct s.skill_id
                                     from cc_skill_in_agent s
                                     where s.agent_id = a.id
                                       and s.capacity between ait.min_capacity and ait.max_capacity
                                 )
                             )
                               and t.domain_id = a.domain_id
                             limit 10) teams ON true

         left join lateral (
    select ares.agent_id,
           case when l.state = 'online' then l.delta + ares.online else ares.online end    online,
           case when l.state = 'offline' then l.delta + ares.offline else ares.offline end offline,
           case when l.state = 'pause' then l.delta + ares.pause else ares.pause end       pause,
           (extract(epoch from (ares.offering + ares.bridged + ares.wrap_time)) /
            extract(epoch from :To::timestamptz - :From::timestamptz)) * 100 as            utilization,
           ares.bridged                                                      as            call_time,
           ares.cnt                                                                        handles,
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
                    coalesce(count(*) filter (where ah.state = 'missed' ), 0)                     missed,
                    max(ah.joined_at) filter ( where ah.state = 'bridged' )                       max_bridged_at,
                    max(ah.joined_at) filter ( where ah.state = 'offering' )                      max_offering_at,
                    min(ah.joined_at)
             from cc_agent_state_history ah
             where ah.joined_at between (:From::timestamptz) and (:To::timestamptz)
               and ah.agent_id = a.id
             group by 1
         ) ares
             left join lateral (
        select h2.state,
               ares.min - (:From::timestamptz) delta
        from cc_agent_state_history h2
        where h2.joined_at < ares.min
          and h2.agent_id = ares.agent_id
          and h2.state in ('online', 'offline', 'pause')
        order by h2.joined_at desc
        limit 1
        ) l on true
    ) stat on stat.agent_id = a.id
         left join lateral (
            select c.id
            from cc_calls c
            where c.agent_id = a.id and c.hangup_at isnull and c.direction notnull
            limit 1
         ) active_call on true
         inner join lateral (select case
                                        when stat isnull or
                                             (now() - a.last_state_change > :To::timestamptz - :From::timestamptz)
                                            then (:To::timestamptz) - (:From::timestamptz)
                                        else now() - a.last_state_change end t) x on true
) t
where t.domain_id = :DomainId
 and (:AgentIds::int[] isnull or t.agent_id = any(:AgentIds))
and (:Q::varchar isnull or t.name ilike :Q::varchar)
and (:Status::varchar[] isnull or (t.status = any(:Status)))
and ( (:UFrom::numeric isnull or :UTo::numeric isnull) or (t.utilization between :UFrom and :UTo) )
and (:QueueIds::int[] isnull  or (t.queue_ids notnull and t.queue_ids::int[] && :QueueIds::int[]))
and (:TeamIds::int[] isnull  or (t.team_ids notnull and t.team_ids::int[] && :TeamIds::int[]))
and (:HasCall::bool isnull or (not :HasCall or active_call_id notnull ))
limit :Limit
offset :Offset`, map[string]interface{}{
		"DomainId": domainId,
		"Q":        search.GetQ(),
		"Limit":    search.GetLimit(),
		"Offset":   search.GetOffset(),
		"From":     model.GetBetweenFromTime(&search.Time),
		"To":       model.GetBetweenToTime(&search.Time),
		"UFrom":    model.GetBetweenFrom(search.Utilization),
		"UTo":      model.GetBetweenFrom(search.Utilization),
		"AgentIds": pq.Array(search.AgentIds),
		"Status":   pq.Array(search.Status),
		"QueueIds": pq.Array(search.QueueIds),
		"TeamIds":  pq.Array(search.TeamIds),
		"HasCall":  search.HasCall,
	})

	if err != nil {
		return nil, model.NewAppError("SqlAgentStore.StatusStatistic", "store.sql_agent.get_status_stats.app_error", nil, err.Error(), extractCodeFromErr(err))
	}

	return list, nil
}
