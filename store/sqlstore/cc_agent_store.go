package sqlstore

import (
	"database/sql"
	"fmt"
	"github.com/lib/pq"
	"github.com/webitel/engine/auth_manager"
	"github.com/webitel/engine/model"
	"github.com/webitel/engine/store"
	"net/http"
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

func (s SqlAgentStore) Create(agent *model.Agent) (*model.Agent, *model.AppError) {
	var out *model.Agent
	if err := s.GetMaster().SelectOne(&out, `with i as (
			insert into cc_agent ( user_id, description, domain_id, created_at, created_by, updated_at, updated_by, progressive_count)
			values (:UserId, :Description, :DomainId, :CreatedAt, :CreatedBy, :UpdatedAt, :UpdatedBy, :ProgressiveCount)
			returning *
		)
		select i.id, i.status, i.state, i.description,  (extract(EPOCH from i.last_state_change) * 1000)::int8 last_state_change, i.state_timeout, i.domain_id,  progressive_count,
			json_build_object('id', ct.id, 'name', coalesce( (ct.name)::varchar, ct.username))::jsonb as user
		from i
		  inner join directory.wbt_user ct on ct.id = i.user_id`,
		map[string]interface{}{
			"UserId":           agent.User.Id,
			"Description":      agent.Description,
			"DomainId":         agent.DomainId,
			"CreatedAt":        agent.CreatedAt,
			"CreatedBy":        agent.CreatedBy.Id,
			"UpdatedAt":        agent.UpdatedAt,
			"UpdatedBy":        agent.UpdatedBy.Id,
			"ProgressiveCount": agent.ProgressiveCount,
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
		"DomainId": domainId,
		"Ids":      pq.Array(search.Ids),
		"Q":        search.GetQ(),
	}

	err := s.ListQuery(&agents, search.ListRequest,
		`domain_id = :DomainId and ( (:Ids::int[] isnull or id = any(:Ids) ) and  (:Q::varchar isnull or (description ilike :Q::varchar or status ilike :Q::varchar ) ))`,
		model.Agent{}, f)
	if err != nil {
		return nil, model.NewAppError("SqlAgentStore.GetAllPage", "store.sql_agent.get_all.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	return agents, nil
}

func (s SqlAgentStore) GetAllPageByGroups(domainId int64, groups []int, search *model.SearchAgent) ([]*model.Agent, *model.AppError) {
	var agents []*model.Agent

	f := map[string]interface{}{
		"DomainId": domainId,
		"Groups":   pq.Array(groups),
		"Access":   auth_manager.PERMISSION_ACCESS_READ.Value(),
		"Ids":      pq.Array(search.Ids),
		"Q":        search.GetQ(),
	}

	err := s.ListQuery(&agents, search.ListRequest,
		`domain_id = :DomainId and ( (:Ids::int[] isnull or id = any(:Ids) ) and  (:Q::varchar isnull or (description ilike :Q::varchar or status ilike :Q::varchar ) )) and
			(
					exists(select 1
					  from cc_agent_acl
					  where cc_agent_acl.dc = t.domain_id and cc_agent_acl.object = t.id and cc_agent_acl.subject = any(:Groups::int[]) and cc_agent_acl.access&:Access = :Access)
		  	) `,
		model.Agent{}, f)
	if err != nil {
		return nil, model.NewAppError("SqlAgentStore.GetAllPageByGroups", "store.sql_agent.get_all.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	return agents, nil
}

func (s SqlAgentStore) Get(domainId int64, id int64) (*model.Agent, *model.AppError) {
	var agent *model.Agent
	if err := s.GetReplica().SelectOne(&agent, `
			select a.id, a.status, a.domain_id, a.description, (extract(EPOCH from a.last_state_change) * 1000)::int8 last_status_change, progressive_count, 
				json_build_object('id', ct.id, 'name', coalesce( (ct.name)::varchar, ct.username))::jsonb as user
				from cc_agent a
					inner join directory.wbt_user ct on ct.id = a.user_id
				where domain_id = :DomainId and a.id = :Id 	
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
	err := s.GetMaster().SelectOne(&agent, `with u as (
			update cc_agent
			set user_id = :UserId,
				description = :Description,
				updated_at = :UpdatedAt,
				updated_by = :UpdatedBy,
			    progressive_count = :ProgressiveCount
			where id = :Id and domain_id = :DomainId
			returning *
		)
		select u.id, u.status, u.domain_id, u.description, (extract(EPOCH from u.last_state_change) * 1000)::int8 last_status_change, progressive_count, 
			json_build_object('id', ct.id, 'name', ct.name)::jsonb as user
		from u
			inner join directory.wbt_user ct on ct.id = u.user_id
		order by id`, map[string]interface{}{
		"UserId":           agent.User.Id,
		"Description":      agent.Description,
		"ProgressiveCount": agent.ProgressiveCount,
		"Id":               agent.Id,
		"DomainId":         agent.DomainId,
		"UpdatedAt":        agent.UpdatedAt,
		"UpdatedBy":        agent.UpdatedBy.Id,
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

func (s SqlAgentStore) InTeam(domainId, id int64, search *model.SearchAgentInTeam) ([]*model.AgentInTeam, *model.AppError) {
	var res []*model.AgentInTeam

	_, err := s.GetReplica().Select(&res, `select cc_get_lookup(t.id, t.name) as team, t.strategy
from cc_team t
where t.id in (
    select a.team_id
    from cc_agent_in_team a
    where a.agent_id = :AgentId
       or a.skill_id in (
        select s.skill_id
        from cc_skill_in_agent s
        where s.agent_id = :AgentId and s.capacity between a.min_capacity and a.max_capacity
    )
) and t.domain_id = :DomainId
  and ( (:Q::varchar isnull or (t.name ilike :Q::varchar ) ))
order by t.id
			limit :Limit
			offset :Offset`, map[string]interface{}{
		"AgentId":  id,
		"DomainId": domainId,
		"Limit":    search.GetLimit(),
		"Offset":   search.GetOffset(),
		"Q":        search.GetQ(),
	})

	if err != nil {
		return nil, model.NewAppError("SqlAgentStore.InTeam", "store.sql_agent.get_team.app_error", nil,
			fmt.Sprintf("Id=%v, %s", id, err.Error()), extractCodeFromErr(err))
	}

	return res, nil
}

func (s SqlAgentStore) InQueue(domainId, id int64, search *model.SearchAgentInQueue) ([]*model.AgentInQueue, *model.AppError) {
	var res []*model.AgentInQueue

	_, err := s.GetReplica().Select(&res, `select cc_get_lookup(q.id, q.name) as queue,
       q.priority,
       q.type,
       q.strategy,
       q.enabled,
       coalesce(sum(cqs.member_count), 0) count_members,
       coalesce(sum(cqs.member_waiting), 0) waiting_members,
       (select count(*) from cc_member_attempt a where a.queue_id = q.id) active_members
from cc_queue q
    left join cc_queue_statistics cqs on q.id = cqs.queue_id
where q.domain_id = :DomainId and q.team_id in (
    select t.id
    from cc_team t
    where t.id in (
        select a.team_id
        from cc_agent_in_team a
        where a.agent_id = :AgentId
           or a.skill_id in (
            select s.skill_id
            from cc_skill_in_agent s
            where s.agent_id = :AgentId and s.capacity between a.min_capacity and a.max_capacity
        )
    ) and t.domain_id = :DomainId
) and ( (:Q::varchar isnull or (q.name ilike :Q::varchar ) ))
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

func (s SqlAgentStore) HistoryState(agentId int64, search *model.SearchAgentState) ([]*model.AgentState, *model.AppError) {
	var res []*model.AgentState
	_, err := s.GetReplica().Select(&res, `select h.id, (extract(EPOCH from h.joined_at) * 1000)::int8 as joined_at, h.state, h.timeout_at, cc_get_lookup(h.queue_id, q.name) queue
from cc_agent_state_history h
    left join cc_queue q on q.id = h.queue_id
where h.agent_id = :AgentId and h.joined_at between to_timestamp( (:From::int8)/1000)::timestamp and to_timestamp(:To::int8/1000)::timestamp
	and ( (:Q::varchar isnull or (q.name ilike :Q::varchar ) ))
order by h.joined_at desc
limit :Limit
offset :Offset`, map[string]interface{}{
		"AgentId": agentId,
		"From":    search.From,
		"To":      search.To,
		"Limit":   search.GetLimit(),
		"Offset":  search.GetOffset(),
		"Q":       search.GetQ(),
	})

	if err != nil {
		return nil, model.NewAppError("SqlAgentStore.HistoryState", "store.sql_agent.get_state_history.app_error", nil,
			fmt.Sprintf("AgentId=%v, %s", agentId, err.Error()), extractCodeFromErr(err))
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
  and   ( (:Q::varchar isnull or (u.name ilike :Q::varchar ) ))
order by u.id
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
       ch.x as channels
from cc_agent a
     LEFT JOIN LATERAL ( SELECT json_agg(json_build_object('channel', c.channel, 'online', c.online, 'state',
                                                       c.state, 'joined_at',
                                                       (date_part('epoch'::text, c.joined_at) * 1000::double precision)::bigint)) AS x
                     FROM call_center.cc_agent_channels c
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
