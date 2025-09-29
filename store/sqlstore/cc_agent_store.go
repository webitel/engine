package sqlstore

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/lib/pq"
	"github.com/webitel/engine/model"
	"github.com/webitel/engine/pkg/wbt/auth_manager"
	"github.com/webitel/engine/store"
)

type SqlAgentStore struct {
	SqlStore
}

func NewSqlAgentStore(sqlStore SqlStore) store.AgentStore {
	us := &SqlAgentStore{sqlStore}
	return us
}

func (s SqlAgentStore) AgentCC(ctx context.Context, domainId int64, userId int64) (*model.AgentCC, model.AppError) {
	var res *model.AgentCC
	err := s.GetReplica().WithContext(ctx).SelectOne(&res, `select length(coalesce(u.extension, '')) > 0 as has_extension,
       a.id notnull as has_agent, a.id as agent_id
from directory.wbt_user u
    left join call_center.cc_agent a on u.id = a.user_id
where u.id = :UserId and u.dc = :DomainId`, map[string]interface{}{
		"UserId":   userId,
		"DomainId": domainId,
	})

	if err != nil {
		return nil, model.NewCustomCodeError("store.sql_agent.has_agent.app_error", fmt.Sprintf("Id=%v, %s", userId, err.Error()), extractCodeFromErr(err))
	}

	return res, nil
}

func (s SqlAgentStore) CheckAccess(ctx context.Context, domainId, id int64, groups []int, access auth_manager.PermissionAccess) (bool, model.AppError) {

	res, err := s.GetReplica().WithContext(ctx).SelectNullInt(`select 1
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

	return res.Valid && res.Int64 == 1, nil
}

func (s SqlAgentStore) AccessAgents(ctx context.Context, domainId int64, agentIds []int64, groups []int, access auth_manager.PermissionAccess) ([]int64, model.AppError) {
	var res []int64
	_, err := s.GetReplica().WithContext(ctx).Select(&res, `select distinct a.object::int agent_id
          from call_center.cc_agent_acl a
          where a.dc = :DomainId
            and a.object = any(:Ids::int[])
            and a.subject = any (:Groups::int[])
            and a.access & :Access = :Access`, map[string]interface{}{
		"DomainId": domainId,
		"Ids":      pq.Array(agentIds),
		"Groups":   pq.Array(groups),
		"Access":   access.Value(),
	})

	if err != nil {
		return nil, model.NewInternalError("store.sql_agent.access.app_error", fmt.Sprintf("record=%v, %v", agentIds, err.Error()))
	}

	return res, nil
}

// FIXME
func (s SqlAgentStore) Create(ctx context.Context, agent *model.Agent) (*model.Agent, model.AppError) {
	var out *model.Agent
	if err := s.GetMaster().WithContext(ctx).SelectOne(&out, `with a as (
			insert into call_center.cc_agent ( user_id, description, domain_id, created_at, created_by, updated_at, updated_by, progressive_count, greeting_media_id,
				allow_channels, chat_count, supervisor_ids, team_id, region_id, supervisor, auditor_ids, task_count, screen_control)
			values (:UserId, :Description, :DomainId, :CreatedAt, :CreatedBy, :UpdatedAt, :UpdatedBy, :ProgressiveCount, :GreetingMedia,
					:AllowChannels, :ChatCount, :SupervisorIds, :TeamId, :RegionId, :Supervisor, :AuditorIds, :TaskCount, :ScreenControl)
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
	   a.task_count,
       (SELECT jsonb_agg(sag."user") AS jsonb_agg
        FROM call_center.cc_agent_with_user sag
        WHERE sag.id = any(a.supervisor_ids)) as supervisor,
       (SELECT jsonb_agg(call_center.cc_get_lookup(aud.id, coalesce(aud.name, aud.username))) AS jsonb_agg
        FROM directory.wbt_user aud
WHERE aud.id = any(a.auditor_ids)) as auditor,
	   call_center.cc_get_lookup(t.id, t.name) as team,
	   call_center.cc_get_lookup(r.id, r.name) as region,
       a.supervisor as is_supervisor,
	   a.screen_control,
	   t.screen_control is false allow_set_screen_control	
FROM a
         LEFT JOIN directory.wbt_user ct ON ct.id = a.user_id
         LEFT JOIN storage.media_files g ON g.id = a.greeting_media_id
         left join call_center.cc_team t on t.id = a.team_id
         left join flow.region r on r.id = a.region_id
         LEFT JOIN LATERAL ( SELECT jsonb_agg(jsonb_build_object('channel', c.channel, 'online', true, 'state', c.state,
                                                      'joined_at',
                                                      (date_part('epoch'::text, c.joined_at) * 1000::double precision)::bigint)) AS x
                             FROM call_center.cc_agent_channel c
                             WHERE c.agent_id = a.id) ch ON true`,
		map[string]interface{}{
			"UserId":           agent.User.Id,
			"Description":      agent.Description,
			"DomainId":         agent.DomainId,
			"CreatedAt":        agent.CreatedAt,
			"CreatedBy":        agent.CreatedBy.GetSafeId(),
			"UpdatedAt":        agent.UpdatedAt,
			"UpdatedBy":        agent.UpdatedBy.GetSafeId(),
			"ProgressiveCount": agent.ProgressiveCount,
			"GreetingMedia":    agent.GreetingMediaId(),
			"AllowChannels":    pq.Array(agent.AllowChannels),
			"ChatCount":        agent.ChatCount,
			"SupervisorIds":    pq.Array(model.LookupIds(agent.Supervisor)),
			"TeamId":           agent.Team.GetSafeId(),
			"RegionId":         agent.Region.GetSafeId(),
			"AuditorIds":       pq.Array(model.LookupIds(agent.Auditor)),
			"Supervisor":       agent.IsSupervisor,
			"TaskCount":        agent.TaskCount,
			"ScreenControl":    agent.ScreenControl,
		}); err != nil {
		return nil, model.NewCustomCodeError("store.sql_agent.save.app_error", fmt.Sprintf("%v", err.Error()), extractCodeFromErr(err))
	} else {
		return out, nil
	}
}

func (s SqlAgentStore) GetAllPage(ctx context.Context, domainId int64, search *model.SearchAgent) ([]*model.Agent, model.AppError) {
	var (
		agents         []*model.Agent
		searchOperator string
	)
	q, regExpFound := model.ParseRegexp(search.Q)
	if regExpFound {
		searchOperator = RegExpComparisonOperator
	} else {
		searchOperator = ILikeComparisonOperator
	}
	f := map[string]interface{}{
		"DomainId":      domainId,
		"Ids":           pq.Array(search.Ids),
		"Q":             q,
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
		"NotTeamIds":    pq.Array(search.NotTeamIds),
		"NotSkillIds":   pq.Array(search.NotSkillIds),
	}
	err := s.ListQuery(ctx, &agents, search.ListRequest,
		fmt.Sprintf(`domain_id = :DomainId
				and (:Ids::int[] isnull or id = any(:Ids))
				and (:TeamIds::int[] isnull or team_id = any(:TeamIds))
				and (:AllowChannels::varchar[] isnull or allow_channels && :AllowChannels )
				and (:SupervisorIds::int[] isnull or supervisor_ids && :SupervisorIds)
				and (:RegionIds::int[] isnull or region_id = any(:RegionIds))
				and (:Extensions::varchar[] isnull or extension = any(:Extensions))
				and (:UserIds::int8[] isnull or user_id = any(:UserIds))
				and (:AuditorIds::int8[] isnull or auditor_ids && :AuditorIds)
				and (:NotTeamIds::int[] isnull or (team_id isnull or not team_id = any(:NotTeamIds::int[])))
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
				and (:NotSkillIds::int[] isnull or not exists(select 1 from call_center.cc_skill_in_agent sia where sia.agent_id = t.id and sia.skill_id = any(:NotSkillIds)))
				and (:Q::varchar isnull or (name %s :Q::varchar or description %[1]s :Q::varchar or status %[1]s :Q::varchar ))`, searchOperator),
		model.Agent{}, f)
	if err != nil {
		return nil, model.NewInternalError("store.sql_agent.get_all.app_error", err.Error())
	}

	return agents, nil
}

func (s SqlAgentStore) GetAllPageByGroups(ctx context.Context, domainId int64, groups []int, search *model.SearchAgent) ([]*model.Agent, model.AppError) {
	var agents []*model.Agent

	f := map[string]interface{}{
		"Groups":        pq.Array(groups),
		"Access":        auth_manager.PERMISSION_ACCESS_READ.Value(),
		"DomainId":      domainId,
		"Ids":           pq.Array(search.Ids),
		"Q":             search.GetRegExpQ(),
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
		"NotTeamIds":    pq.Array(search.NotTeamIds),
		"NotSkillIds":   pq.Array(search.NotSkillIds),
	}

	err := s.ListQuery(ctx, &agents, search.ListRequest,
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
				and (:NotTeamIds::int[] isnull or (team_id isnull or not team_id = any(:NotTeamIds::int[])))
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
				and (:NotSkillIds::int[] isnull or not exists(select 1 from call_center.cc_skill_in_agent sia where sia.agent_id = t.id and sia.skill_id = any(:NotSkillIds)))
				and (:Q::varchar isnull or (name ~ :Q::varchar or description ~ :Q::varchar or status ~ :Q::varchar ))
				and (
					exists(select 1
					  from call_center.cc_agent_acl
					  where call_center.cc_agent_acl.dc = t.domain_id and call_center.cc_agent_acl.object = t.id 
						and call_center.cc_agent_acl.subject = any(:Groups::int[]) and call_center.cc_agent_acl.access&:Access = :Access)
		  		)`,
		model.Agent{}, f)
	if err != nil {
		return nil, model.NewInternalError("store.sql_agent.get_all.app_error", err.Error())
	}

	return agents, nil
}

func (s SqlAgentStore) GetActiveTask(ctx context.Context, domainId, id int64) ([]*model.CCTask, model.AppError) {
	query := `
		select 
		    a.id as attempt_id,
		    a.channel,
		    a.node_id as app_id,
		    a.queue_id,
		    coalesce(a.queue_params->>'queue_name', '') as queue_name,
		    a.member_id,
		    a.member_call_id as member_channel_id,
		    a.agent_call_id as agent_channel_id,
		    a.destination,
		    a.state,
		    call_center.cc_view_timestamp(a.leaving_at) as leaving_at,
		    coalesce((a.queue_params->'has_reporting')::bool, false) as has_reporting,
		    coalesce((a.queue_params->'has_form')::bool, false) as has_form,
		    (a.queue_params->'processing_sec')::int as processing_sec,
		    (a.queue_params->'processing_renewal_sec')::int as processing_renewal_sec,
			coalesce((a.queue_params->'has_prolongation')::bool, false) as has_prolongation,
			(a.queue_params->'remaining_prolongations')::int as remaining_prolongations,
			(a.queue_params->'prolongation_sec')::int as prolongation_sec,
			call_center.cc_view_timestamp(a.timeout) as processing_timeout_at,
		    a.form_view as form,
		    m.variables,
		    m.name as member_name,
		    call_center.cc_view_timestamp(a.bridged_at) as bridged_at,
		    a.agent_id
		from 
			call_center.cc_member_attempt a
		inner join 
			call_center.cc_agent a2 on a2.id = a.agent_id
		left join 
			call_center.cc_member m on a.member_id = m.id
		where 
			a.agent_id = :AgentId
		    and a2.domain_id = :DomainId
		    and a.state != 'leaving'
		    and a.node_id notnull
	`

	args := map[string]any{
		"AgentId":  id,
		"DomainId": domainId,
	}

	cc, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	var res []*model.CCTask
	if _, err := s.GetMaster().WithContext(cc).Select(&res, query, args); err != nil {
		return nil, model.NewCustomCodeError("store.sql_agent.get_tasks.app_error", fmt.Sprintf("Id=%v, %s", id, err.Error()), extractCodeFromErr(err))
	}

	return res, nil
}

func (s SqlAgentStore) Get(ctx context.Context, domainId int64, id int64) (*model.Agent, model.AppError) {
	var agent *model.Agent
	if err := s.GetReplica().WithContext(ctx).SelectOne(&agent, `
		SELECT a.domain_id,
			   a.id,
			   COALESCE(ct.name, ct.username)                             AS name,
			   a.status,
			   a.description,
			   (date_part('epoch'::text, a.last_state_change) *
				1000::double precision)::bigint                                                                       AS last_status_change,
			   date_part('epoch'::text, now() - a.last_state_change)::bigint                                          AS status_duration,
			   a.progressive_count,
			   ch.x                                                                                                   AS channel,
			   json_build_object('id', ct.id, 'name', COALESCE(ct.name, ct.username))::jsonb AS "user",
			   call_center.cc_get_lookup(a.greeting_media_id::bigint, g.name)                                         AS greeting_media,
			   a.allow_channels,
			   a.chat_count,
			   a.task_count,
			   (SELECT jsonb_agg(sag."user") AS jsonb_agg
				FROM call_center.cc_agent_with_user sag
				WHERE sag.id = any(a.supervisor_ids)) as supervisor,
			   (SELECT jsonb_agg(call_center.cc_get_lookup(aud.id, coalesce(aud.name, aud.username))) AS jsonb_agg
				FROM directory.wbt_user aud
				WHERE aud.id = any(a.auditor_ids)) as auditor,
			   call_center.cc_get_lookup(t.id, t.name) as team,
			   call_center.cc_get_lookup(r.id, r.name) as region,
			   a.supervisor as is_supervisor,
			   a.screen_control,
			   t.screen_control is false allow_set_screen_control,
			   ct.extension	
		FROM call_center.cc_agent a
				 LEFT JOIN directory.wbt_user ct ON ct.id = a.user_id
				 LEFT JOIN storage.media_files g ON g.id = a.greeting_media_id
				 left join call_center.cc_team t on t.id = a.team_id
				 left join flow.region r on r.id = a.region_id
				 LEFT JOIN LATERAL ( SELECT jsonb_agg(jsonb_build_object('channel', c.channel, 'online', true, 'state', c.state,
                                                      'joined_at',
                                                      (date_part('epoch'::text, c.joined_at) * 1000::double precision)::bigint)) AS x
                             FROM call_center.cc_agent_channel c
                             WHERE c.agent_id = a.id) ch ON true
				where a.domain_id = :DomainId and a.id = :Id 	
		`, map[string]interface{}{"Id": id, "DomainId": domainId}); err != nil {
		if err == sql.ErrNoRows {
			return nil, model.NewNotFoundError("store.sql_agent.get.app_error", fmt.Sprintf("Id=%v, %s", id, err.Error()))
		} else {
			return nil, model.NewInternalError("store.sql_agent.get.app_error", fmt.Sprintf("Id=%v, %s", id, err.Error()))
		}
	} else {
		return agent, nil
	}
}

func (s SqlAgentStore) Update(ctx context.Context, agent *model.Agent) (*model.Agent, model.AppError) {
	err := s.GetMaster().WithContext(ctx).SelectOne(&agent, `with a as (
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
				auditor_ids = :AuditorIds,
                task_count = :TaskCount,
                screen_control = :ScreenControl
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
			   a.task_count,
			   (SELECT jsonb_agg(sag."user") AS jsonb_agg
				FROM call_center.cc_agent_with_user sag
				WHERE sag.id = any(a.supervisor_ids)) as supervisor,
			   (SELECT jsonb_agg(call_center.cc_get_lookup(aud.id, coalesce(aud.name, aud.username))) AS jsonb_agg
				FROM directory.wbt_user aud
				WHERE aud.id = any(a.auditor_ids)) as auditor,
			   call_center.cc_get_lookup(t.id, t.name) as team,
			   call_center.cc_get_lookup(r.id, r.name) as region,
			   a.supervisor as is_supervisor,
			   a.screen_control,
			   t.screen_control is false allow_set_screen_control
		FROM  a
				 LEFT JOIN directory.wbt_user ct ON ct.id = a.user_id
				 LEFT JOIN storage.media_files g ON g.id = a.greeting_media_id
				 left join call_center.cc_team t on t.id = a.team_id
				 left join flow.region r on r.id = a.region_id
				 LEFT JOIN LATERAL ( SELECT jsonb_agg(jsonb_build_object('channel', c.channel, 'online', true, 'state', c.state,
                                                      'joined_at',
                                                      (date_part('epoch'::text, c.joined_at) * 1000::double precision)::bigint)) AS x
                             FROM call_center.cc_agent_channel c
                             WHERE c.agent_id = a.id) ch ON true`, map[string]interface{}{
		"UserId":           agent.User.Id,
		"Description":      agent.Description,
		"ProgressiveCount": agent.ProgressiveCount,
		"Id":               agent.Id,
		"DomainId":         agent.DomainId,
		"UpdatedAt":        agent.UpdatedAt,
		"UpdatedBy":        agent.UpdatedBy.GetSafeId(),
		"GreetingMediaId":  agent.GreetingMediaId(),
		"AllowChannels":    pq.Array(agent.AllowChannels),
		"ChatCount":        agent.ChatCount,
		"SupervisorIds":    pq.Array(model.LookupIds(agent.Supervisor)),
		"TeamId":           agent.Team.GetSafeId(),
		"RegionId":         agent.Region.GetSafeId(),
		"AuditorIds":       pq.Array(model.LookupIds(agent.Auditor)),
		"Supervisor":       agent.IsSupervisor,
		"TaskCount":        agent.TaskCount,
		"ScreenControl":    agent.ScreenControl,
	})
	if err != nil {
		return nil, model.NewCustomCodeError("store.sql_agent.update.app_error", fmt.Sprintf("Id=%v, %s", agent.Id, err.Error()), extractCodeFromErr(err))
	}
	return agent, nil
}

func (s SqlAgentStore) Delete(ctx context.Context, domainId, id int64) model.AppError {
	if _, err := s.GetMaster().WithContext(ctx).Exec(`delete from call_center.cc_agent c where c.id=:Id and c.domain_id = :DomainId`,
		map[string]interface{}{"Id": id, "DomainId": domainId}); err != nil {
		return model.NewCustomCodeError("store.sql_agent.delete.app_error", fmt.Sprintf("Id=%v, %s", id, err.Error()), extractCodeFromErr(err))
	}
	return nil
}

func (s SqlAgentStore) SetStatus(ctx context.Context, domainId, agentId int64, status string, payload interface{}) (bool, model.AppError) {
	if r, err := s.GetMaster().WithContext(ctx).Exec(`update call_center.cc_agent
			set status = :Status
  			,status_payload = :Payload
			where id = :AgentId and domain_id = :DomainId and (status <> :Status or status_payload <> :Payload)`, map[string]interface{}{"AgentId": agentId, "Status": status, "Payload": payload, "DomainId": domainId}); err != nil {
		return false, model.NewInternalError("store.sql_agent.set_status.app_error", fmt.Sprintf("AgenetId=%v, %s", agentId, err.Error()))
	} else {
		var cnt int64
		if cnt, err = r.RowsAffected(); err != nil {
			return false, model.NewInternalError("store.sql_agent.set_status.app_error", fmt.Sprintf("AgenetId=%v, %s", agentId, err.Error()))
		}
		return cnt > 0, nil
	}
}

func (s SqlAgentStore) InQueue(ctx context.Context, domainId, id int64, search *model.SearchAgentInQueue) ([]*model.AgentInQueue, model.AppError) {
	var res []*model.AgentInQueue

	f := map[string]interface{}{
		"DomainId": domainId,
		"AgentId":  id,
		"Q":        search.GetQ(),
	}

	err := s.ListQueryMaster(ctx, &res, search.ListRequest,
		`domain_id = :DomainId
				and enabled
				and agent_id = :AgentId
				and (:Q::varchar isnull or (queue_name ilike :Q::varchar ))`,
		model.AgentInQueue{}, f)

	if err != nil {
		return nil, model.NewCustomCodeError("store.sql_agent.get_queue.app_error", fmt.Sprintf("Id=%v, %s", id, err.Error()), extractCodeFromErr(err))
	}

	return res, nil
}

func (s SqlAgentStore) QueueStatistic(ctx context.Context, domainId, agentId int64) ([]*model.AgentInQueueStatistic, model.AppError) {
	var stats []*model.AgentInQueueStatistic
	_, err := s.GetReplica().WithContext(ctx).Select(&stats, `select
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
    inner join call_center.cc_queue q on q.id = x.queue_id
    left join (
       SELECT c.id, array_agg(DISTINCT o1.id)::integer[] l
       FROM flow.calendar c
              LEFT JOIN flow.calendar_timezones tz ON tz.id = c.timezone_id
              JOIN LATERAL unnest(c.accepts) a(disabled, day, start_time_of_day, end_time_of_day) ON true
              JOIN flow.calendar_timezone_offsets o1
                   ON (a.day + 1) = date_part('isodow'::text, timezone(o1.names[1], now()))::integer AND
                      (to_char(timezone(o1.names[1], now()), 'SSSS'::text)::integer / 60) >= a.start_time_of_day AND
                      (to_char(timezone(o1.names[1], now()), 'SSSS'::text)::integer / 60) <= a.end_time_of_day
        WHERE NOT a.disabled IS TRUE
        group by 1
    ) y on y.id = q.calendar_id
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
            and m.sys_offset_id = any(y.l)
        group by 1,2
    ) m on true
    left join call_center.cc_bucket b on b.id = m.bucket_id
    left join call_center.cc_skill s on s.id = m.skill_id
where q.domain_id = :DomainId
group by q.id`, map[string]interface{}{
		"AgentId":  agentId,
		"DomainId": domainId,
	})

	if err != nil {
		return nil, model.NewCustomCodeError("store.sql_agent.get_queue_stats.app_error", fmt.Sprintf("AgentId=%v, %s", agentId, err.Error()), extractCodeFromErr(err))
	}

	return stats, nil
}

func (s SqlAgentStore) HistoryState(ctx context.Context, domainId int64, search *model.SearchAgentState) ([]*model.AgentState, model.AppError) {
	var res []*model.AgentState

	//fixme
	order := GetOrderBy("cc_agent_state_history", search.Sort)
	if order == "" {
		order = "order by joined_at desc"
	}

	_, err := s.GetReplica().WithContext(ctx).Select(&res, `with ags as (
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
		return nil, model.NewCustomCodeError("store.sql_agent.get_state_history.app_error", err.Error(), extractCodeFromErr(err))
	}

	return res, nil
}

func (s SqlAgentStore) LookupNotExistsUsers(ctx context.Context, domainId int64, search *model.SearchAgentUser) ([]*model.AgentUser, model.AppError) {
	var users []*model.AgentUser

	if _, err := s.GetReplica().WithContext(ctx).Select(&users,
		`select u.id, COALESCE(u.name::text, u.username) COLLATE "default" as name
from directory.wbt_user u
where u.dc = :DomainId
  and not exists(select 1 from call_center.cc_agent a where a.domain_id = :DomainId and a.user_id = u.id)
  and   ( (:Q::varchar isnull or (COALESCE(u.name::text, u.username) COLLATE "default" ilike :Q::varchar ) ))
order by COALESCE(u.name::text, u.username) COLLATE "default"
limit :Limit
offset :Offset`, map[string]interface{}{
			"DomainId": domainId,
			"Limit":    search.GetLimit(),
			"Offset":   search.GetOffset(),
			"Q":        search.GetQ(),
		}); err != nil {
		return nil, model.NewCustomCodeError("store.sql_agent.lookup.users.app_error", err.Error(), extractCodeFromErr(err))
	} else {
		return users, nil
	}
}

func (s SqlAgentStore) LookupNotExistsUsersByGroups(ctx context.Context, domainId int64, groups []int, search *model.SearchAgentUser) ([]*model.AgentUser, model.AppError) {
	var users []*model.AgentUser

	if _, err := s.GetReplica().WithContext(ctx).Select(&users,
		`select u.id, COALESCE(u.name::text, u.username) COLLATE "default" as name
from directory.wbt_user u
where u.dc = :DomainId
  and not exists(select 1 from call_center.cc_agent a where a.domain_id = :DomainId and a.user_id = u.id)
  and (
	exists(select 1
	  from directory.wbt_auth_acl acl
	  where acl.dc = u.dc and acl.object = u.id and acl.subject = any(:Groups::int[]) and acl.access&:Access = :Access)
  ) 
  and   ( (:Q::varchar isnull or (COALESCE(u.name::text, u.username) COLLATE "default" ilike :Q::varchar ) ))
order by COALESCE(u.name::text, u.username) COLLATE "default"
limit :Limit
offset :Offset`, map[string]interface{}{
			"DomainId": domainId,
			"Limit":    search.GetLimit(),
			"Offset":   search.GetOffset(),
			"Q":        search.GetQ(),
			"Groups":   pq.Array(groups),
			"Access":   auth_manager.PERMISSION_ACCESS_READ.Value(),
		}); err != nil {
		return nil, model.NewCustomCodeError("store.sql_agent.lookup.users.app_error", err.Error(), extractCodeFromErr(err))
	} else {
		return users, nil
	}
}

func (s SqlAgentStore) GetSession(ctx context.Context, domainId, userId int64) (*model.AgentSession, model.AppError) {
	var agent *model.AgentSession
	err := s.GetMaster().WithContext(ctx).SelectOne(&agent, `select a.id as agent_id,
       a.status,
       coalesce(a.status_payload, '') status_payload,
	   coalesce(a.status_comment, '') status_comment, 
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
        WHERE aud.id = any(a.auditor_ids)) auditor,
    	a.screen_control
from call_center.cc_agent a
	 left join call_center.cc_team t on t.id = a.team_id
     LEFT JOIN LATERAL ( SELECT jsonb_agg(json_build_object('channel', c.channel, 'state', c.state, 'open', 0, 'max_open', c.max_opened,
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
		return nil, model.NewCustomCodeError("store.sql_agent.get_session.app_error", err.Error(), extractCodeFromErr(err))
	}
	return agent, nil
}

func (s SqlAgentStore) CallStatistics(ctx context.Context, domainId int64, search *model.SearchAgentCallStatistics) ([]*model.AgentCallStatistics, model.AppError) {
	var stats []*model.AgentCallStatistics

	_, err := s.GetReplica().WithContext(ctx).Select(&stats, `select `+strings.Join(GetFields(search.Fields, model.AgentCallStatistics{}), ", ")+`
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
        case when onl_all > 0 then ((coalesce(work_dur, 0) / (onl_all + pause_all))) * 100 else 0 end occupancy,
		coalesce(res.chat_count, 0) as chat_accepts,
		coalesce(res.chat_aht, 0) as chat_aht
    from (
         select cma.agent_id,
			   count(*) filter ( where cma.channel = 'chat' and cma.bridged_at notnull) as chat_count,
			   (avg(extract(epoch from coalesce(cma.reporting_at, cma.leaving_at) - cma.bridged_at)) filter ( where cma.channel = 'chat' and cma.bridged_at notnull ))::int4 as chat_aht,
			   count(*) filter ( where cma.channel = 'call' ) as count,
			   count(*) filter ( where cma.channel = 'call' and c.answered_at isnull and c.cause in ('NO_ANSWER', 'ORIGINATOR_CANCEL') ) as abandoned, -- todo is missing
			   count(*) filter ( where cma.channel = 'call' and  c.answered_at notnull ) as handles,
			   extract(epoch from sum(c.hangup_at - c.bridged_at) filter ( where cma.channel = 'call' and c.bridged_at notnull )) as sum_talk_sec,
			   extract(epoch from avg(c.hangup_at - c.bridged_at) filter ( where cma.channel = 'call' and c.bridged_at notnull )) as avg_talk_sec,
			   extract(epoch from min(c.hangup_at - c.bridged_at) filter ( where cma.channel = 'call' and c.bridged_at notnull )) as min_talk_sec,
			   extract(epoch from max(c.hangup_at - c.bridged_at) filter ( where cma.channel = 'call' and c.bridged_at notnull )) as max_talk_sec,
			   sum(c.hold_sec) sum_hold_sec,
			   avg(c.hold_sec) avg_hold_sec,
			   min(c.hold_sec) min_hold_sec,
			   max(c.hold_sec) max_hold_sec
		from call_center.cc_member_attempt_history cma
			   left join call_center.cc_calls_history c on c.id = cma.agent_call_id::uuid and cma.channel = 'call'
		where (cma.joined_at between :From::timestamptz and :To::timestamptz)
			and cma.domain_id = :DomainId::int8
			and (:AgentIds::int[] isnull or cma.agent_id = any(:AgentIds) )
        group by cma.agent_id
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
		return nil, model.NewCustomCodeError("store.sql_agent.get_call_stats.app_error", err.Error(), extractCodeFromErr(err))
	}

	return stats, nil
}

func (s SqlAgentStore) PauseCause(ctx context.Context, domainId int64, fromUserId, toAgentId int64, allowChange bool) ([]*model.AgentPauseCause, model.AppError) {
	var res []*model.AgentPauseCause
	_, err := s.GetReplica().WithContext(ctx).Select(&res, `select c.id,
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
		return nil, model.NewCustomCodeError("store.sql_agent.list_pause_causes.app_error", err.Error(), extractCodeFromErr(err))
	}

	return res, nil
}

// FIXME sort, columns
// allow_change
func (s SqlAgentStore) StatusStatistic(ctx context.Context, domainId int64, supervisorUserId int64, groups []int, access auth_manager.PermissionAccess, search *model.SearchAgentStatusStatistic) ([]*model.AgentStatusStatistics, model.AppError) {
	var (
		list           []*model.AgentStatusStatistics
		searchOperator string
	)

	q, found := model.ParseRegexp(search.Q)
	if found {
		searchOperator = RegExpComparisonOperator
	} else {
		searchOperator = ILikeComparisonOperator
	}
	_, err := s.GetMaster().WithContext(ctx).Select(&list, fmt.Sprintf(`select agent_id,
       name,
       status,
       status_duration,
	   status_comment,
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
       coalesce(occupancy, 0) as occupancy,
       desc_track,
       screen_control
from (
         select a.id                                                                                  agent_id,
                a.domain_id,
                coalesce(u.name, u.username)::varchar COLLATE "default"                      as                                  name,
                coalesce(u.extension, '')                         as                                  extension,
                exists(select 1 from call_center.socket_session_view ss where ss.user_id = a.user_id and ss.pong < 65 and application_name = 'desc_track') as desc_track,
                a.status,
				a.status_comment,
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
                case when onl_all > 0 then ((coalesce(work_dur, 0) / (onl_all + pause_all))) * 100 else 0 end occupancy,

                coalesce(extract(epoch from call_time)::int8, 0)                                      call_time,
                coalesce(handles, 0)                                                                  handles,
                coalesce(stat.chat_count, 0)                                                          chat_count,
                coalesce(missed, 0)                                                                   missed,
                0::int                                                                                transferred,
                max_bridged_at,
                max_offering_at,
                active_call.id                                    as                                  active_call_id,
                sa.skills,
                sa.skill_ids,
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
                a.region_id,
                a.screen_control
         from call_center.cc_agent a
                  inner join directory.wbt_user u on u.id = a.user_id
                  left join call_center.cc_team team on team.id = a.team_id
                  left join flow.region r on r.id = a.region_id
             	  left join lateral (
             	    select array_agg(distinct sia.skill_id) skill_ids,
						   jsonb_agg(distinct call_center.cc_get_lookup(cs.id, cs.name)) skills
					from call_center.cc_skill_in_agent sia
						inner join call_center.cc_skill cs on cs.id = sia.skill_id
					where sia.agent_id = a.id
             	  ) sa on true
                  left join lateral (
             select array_agg(distinct q.id)                          queue_ids,
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
                  left join (
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
                                                 when (stat isnull and a.last_state_change < :From::timestamptz) or
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
		   and (:SupervisorIds::int4[] isnull or a.supervisor_ids && :SupervisorIds )
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
  and (:Q::varchar isnull or t.name %s :Q::varchar)
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
limit :Limit offset :Offset`, searchOperator), map[string]interface{}{
		"DomainId":         domainId,
		"UserSupervisorId": supervisorUserId,
		//"Groups":     pq.Array(groups),
		//"Access":     access.Value(),
		"Q":             q,
		"Limit":         search.GetLimit(),
		"Offset":        search.GetOffset(),
		"From":          model.GetBetweenFromTime(&search.Time),
		"To":            model.GetBetweenToTime(&search.Time),
		"UFrom":         model.GetBetweenFrom(search.Utilization),
		"UTo":           model.GetBetweenTo(search.Utilization),
		"AgentIds":      pq.Array(search.AgentIds),
		"Status":        pq.Array(search.Status),
		"QueueIds":      pq.Array(search.QueueIds),
		"TeamIds":       pq.Array(search.TeamIds),
		"SkillIds":      pq.Array(search.SkillIds),
		"RegionIds":     pq.Array(search.RegionIds),
		"AuditorIds":    pq.Array(search.AuditorIds),
		"SupervisorIds": pq.Array(search.SupervisorIds),
		"HasCall":       search.HasCall,
	})

	if err != nil {
		return nil, model.NewCustomCodeError("store.sql_agent.get_status_stats.app_error", err.Error(), extractCodeFromErr(err))
	}

	return list, nil
}

func (s SqlAgentStore) SupervisorAgentItem(ctx context.Context, domainId int64, agentId int64, t *model.FilterBetween) (*model.SupervisorAgentItem, model.AppError) {
	var item *model.SupervisorAgentItem

	err := s.GetReplica().WithContext(ctx).SelectOne(&item, `select a.id                                                                           agent_id,
       coalesce(cawu.name, cawu.username)                                     as      name,
       call_center.cc_get_lookup(cawu.id, coalesce(cawu.name, cawu.username)) as      user,
       coalesce(cawu.extension, '')                                           as      extension,
       a.status,
       extract(epoch from x.t)::int                                                   status_duration,

       call_center.cc_get_lookup(t.id, t.name)                                        team,
       (SELECT jsonb_agg(sag."user") AS jsonb_agg
        FROM call_center.cc_agent_with_user sag
        WHERE sag.id = any (a.supervisor_ids))                                        supervisor,
       (SELECT jsonb_agg(call_center.cc_get_lookup(aud.id, coalesce(aud.name, aud.username))) AS jsonb_agg
        FROM directory.wbt_user aud
        WHERE aud.id = any (a.auditor_ids))                                   as      auditor,
       call_center.cc_get_lookup(r.id, r.name)                                        region,
       a.progressive_count,
       a.chat_count,

       (coalesce(extract(epoch from stat.online), 0) +
        case when a.status = 'online' then extract(epoch from x.t) else 0 end)::int8  online,
       (coalesce(extract(epoch from stat.offline), 0) +
        case when a.status = 'offline' then extract(epoch from x.t) else 0 end)::int8 offline,
       (coalesce(extract(epoch from stat.pause), 0) +
        case when a.status = 'pause' then extract(epoch from x.t) else 0 end)::int8   pause,
       coalesce(a.status_payload, '')                                                 pause_cause,
	   coalesce(a.status_comment, '')												  status_comment,
       coalesce(ts.score_optional_avg, 0.0)                                     as      score_optional_avg,
       coalesce(ts.score_required_avg, 0.0)                                     as      score_required_avg,
       coalesce(ts.score_count, 0)                                                as      score_count,
       exists(select 1 from call_center.socket_session_view ss where ss.user_id = a.user_id and ss.pong < 65 and application_name = 'desc_track') as desc_track,
       a.screen_control
from call_center.cc_agent a
         left join call_center.cc_team t on t.id = a.team_id
         left join flow.region r on r.id = a.region_id
         left join flow.calendar_timezones tz on tz.id = r.timezone_id
         left join lateral (
    select ah.agent_id,
           coalesce(sum(duration) filter ( where ah.state = 'online' ), interval '0')  online,
           coalesce(sum(duration) filter ( where ah.state = 'offline' ), interval '0') offline,
           coalesce(sum(duration) filter ( where ah.state = 'pause' ), interval '0')   pause
    from call_center.cc_agent_state_history ah
    where ah.joined_at between (:From::timestamptz) and (:To::timestamptz)
      and ah.agent_id = a.id
    group by 1
    ) stat on true
         inner join lateral (select case
                                        when (stat isnull and a.last_state_change < :From::timestamptz) or
                                             (now() - a.last_state_change > :To::timestamptz - :From::timestamptz)
                                            then (:To::timestamptz) - (:From::timestamptz)
                                        else now() - a.last_state_change end t) x on true
         left join directory.wbt_user cawu on a.user_id = cawu.id
         left join call_center.cc_agent_today_stats ts on ts.agent_id = a.id
where a.id = :AgentId
  and a.domain_id = :DomainId`, map[string]interface{}{
		"DomainId": domainId,
		"AgentId":  agentId,
		"From":     model.GetBetweenFromTime(t),
		"To":       model.GetBetweenToTime(t),
	})

	if err != nil {
		return nil, model.NewCustomCodeError("store.sql_agent.get_status_stats_item.app_error", err.Error(), extractCodeFromErr(err))
	}

	return item, nil
}

func (s SqlAgentStore) DistributeInfoByUserId(ctx context.Context, domainId, userId int64, channel string) (*model.DistributeAgentInfo, model.AppError) {
	var res *model.DistributeAgentInfo
	err := s.GetMaster().WithContext(ctx).SelectOne(&res, `select a.id as agent_id,
   exists(select 1 from call_center.cc_member_attempt att
    where att.agent_id = a.id and att.agent_call_id isnull ) distribute,
   c.state = any(array ['offering', 'bridged']) busy
from call_center.cc_agent a
    inner join call_center.cc_agent_channel c on c.agent_id = a.id
where a.user_id = :UserId and a.domain_id = :DomainId and c.channel = :Channel::varchar`, map[string]interface{}{
		"UserId":   userId,
		"DomainId": domainId,
		"Channel":  channel,
	})

	if err != nil {
		return nil, model.NewCustomCodeError("store.sql_agent.get_dis_stats_item.app_error", err.Error(), extractCodeFromErr(err))
	}

	return res, nil
}

func (s SqlAgentStore) TodayStatistics(ctx context.Context, domainId int64, agentId *int64, userId *int64) (*model.AgentStatistics, model.AppError) {
	params := map[string]interface{}{
		"DomainId": domainId,
	}

	q := `select
	s.utilization, 
	s.occupancy, 
	s.call_abandoned, 
	s.call_handled, 
	s.call_missed, 
	s.call_inbound, 
	s.avg_talk_sec, 
	s.avg_hold_sec, 
	s.chat_accepts, 
	s.chat_aht,
	s.score_count,
	s.score_optional_avg,
	s.score_optional_sum,
	s.score_required_avg,
	s.score_required_sum,
    s.sum_talk_sec,
    s.voice_mail,
    s.available,
    s.online,
    s.processing,
    s.task_accepts,
	s.queue_talk_sec,
	s.call_queue_missed,
	s.call_inbound_queue,
	s.call_dialer_queue,
	s.call_manual
from call_center.cc_agent_today_stats s
where s.domain_id = :DomainId and `

	if agentId != nil {
		params["Id"] = agentId
		q += `s.agent_id = :Id`
	} else {
		params["Id"] = userId
		q += `s.user_id = :Id`
	}

	var stat *model.AgentStatistics
	err := s.GetReplica().WithContext(ctx).SelectOne(&stat, q, params)

	if err != nil {
		return nil, model.NewCustomCodeError("store.sql_agent.statistic.today", err.Error(), extractCodeFromErr(err))
	}

	return stat, nil

}

func (s SqlAgentStore) UsersStatus(ctx context.Context, domainId int64, search *model.SearchUserStatus) ([]*model.UserStatus, model.AppError) {
	var (
		users          []*model.UserStatus
		searchOperator string
	)
	q, regExpFound := model.ParseRegexp(search.Q)
	if regExpFound {
		searchOperator = RegExpComparisonOperator
	} else {
		searchOperator = ILikeComparisonOperator
	}
	f := map[string]interface{}{
		"DomainId": domainId,
		"Q":        q,
	}

	err := s.ListQuery(ctx, &users, search.ListRequest,
		fmt.Sprintf(`domain_id = :DomainId
				and (:Q::varchar isnull or (name %s :Q::varchar or extension %[1]s :Q::varchar ))`, searchOperator),
		model.UserStatus{}, f)
	if err != nil {
		return nil, model.NewInternalError("store.sql_agent.get_users.app_error", err.Error())
	}

	return users, nil
}

func (s SqlAgentStore) UsersStatusByGroup(ctx context.Context, domainId int64, groups []int, search *model.SearchUserStatus) ([]*model.UserStatus, model.AppError) {
	var (
		users          []*model.UserStatus
		searchOperator string
	)
	q, found := model.ParseRegexp(search.Q)
	if found {
		searchOperator = RegExpComparisonOperator
	} else {
		searchOperator = ILikeComparisonOperator
	}
	f := map[string]interface{}{
		"DomainId": domainId,
		"Q":        q,
		"Groups":   pq.Array(groups),
		"Access":   auth_manager.PERMISSION_ACCESS_READ.Value(),
	}

	err := s.ListQuery(ctx, &users, search.ListRequest,
		fmt.Sprintf(`domain_id = :DomainId
				and (:Q::varchar isnull or (name %s :Q::varchar or extension %[1]s :Q::varchar ))
				and (
					exists(select 1
					  from directory.wbt_auth_acl
					  where directory.wbt_auth_acl.dc = t.domain_id and directory.wbt_auth_acl.object = t.id 
						and directory.wbt_auth_acl.subject = any(:Groups::int[]) and directory.wbt_auth_acl.access&:Access = :Access)
		  		)`, searchOperator),
		model.UserStatus{}, f)
	if err != nil {
		return nil, model.NewInternalError("store.sql_agent.get_users.app_error", err.Error())
	}

	return users, nil
}
