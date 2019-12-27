package sqlstore

import (
	"database/sql"
	"fmt"
	"github.com/lib/pq"
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

func (s SqlAgentStore) CheckAccess(domainId, id int64, groups []int, access model.PermissionAccess) (bool, *model.AppError) {

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
			insert into cc_agent ( user_id, description, domain_id)
			values (:UserId, :Description, :DomainId)
			returning *
		)
		select i.id, i.status, i.state, i.description, i.domain_id, json_build_object('id', ct.id, 'name', ct.name)::jsonb as user
		from i
		  inner join directory.wbt_user ct on ct.id = i.user_id`,
		map[string]interface{}{"UserId": agent.User.Id, "Description": agent.Description,
			"DomainId": agent.DomainId}); err != nil {
		return nil, model.NewAppError("SqlAgentStore.Save", "store.sql_agent.save.app_error", nil,
			fmt.Sprintf("record=%v, %v", agent, err.Error()), http.StatusInternalServerError)
	} else {
		return out, nil
	}
}

func (s SqlAgentStore) GetAllPage(domainId int64, offset, limit int) ([]*model.Agent, *model.AppError) {
	var agents []*model.Agent

	if _, err := s.GetReplica().Select(&agents,
		`select a.id, a.status, state, description, json_build_object('id', ct.id, 'name', ct.name)::jsonb as user
				from cc_agent a
					inner join directory.wbt_user ct on ct.id = a.user_id
				where domain_id = :DomainId
				order by a.id
			limit :Limit
			offset :Offset`, map[string]interface{}{"DomainId": domainId, "Limit": limit, "Offset": offset}); err != nil {
		return nil, model.NewAppError("SqlAgentStore.GetAllPage", "store.sql_agent.get_all.app_error", nil, err.Error(), http.StatusInternalServerError)
	} else {
		return agents, nil
	}
}

func (s SqlAgentStore) GetAllPageByGroups(domainId int64, groups []int, offset, limit int) ([]*model.Agent, *model.AppError) {
	var agents []*model.Agent

	if _, err := s.GetReplica().Select(&agents,
		`select a.id, a.status, state, description, json_build_object('id', ct.id, 'name', ct.name)::jsonb as user
				from cc_agent a
					inner join directory.wbt_user ct on ct.id = a.user_id
				where domain_id = :DomainId and (
					exists(select 1
					  from cc_agent_acl acl
					  where acl.dc = a.domain_id and acl.object = a.id and acl.subject = any(:Groups::int[]) and acl.access&:Access = :Access)
				  )
				order by a.id
			limit :Limit
			offset :Offset`, map[string]interface{}{"DomainId": domainId, "Limit": limit, "Offset": offset, "Groups": pq.Array(groups), "Access": model.PERMISSION_ACCESS_READ.Value()}); err != nil {
		return nil, model.NewAppError("SqlAgentStore.GetAllPage", "store.sql_agent.get_all.app_error", nil, err.Error(), http.StatusInternalServerError)
	} else {
		return agents, nil
	}
}

func (s SqlAgentStore) Get(domainId int64, id int64) (*model.Agent, *model.AppError) {
	var agent *model.Agent
	if err := s.GetReplica().SelectOne(&agent, `
			select a.id, a.status, state, domain_id, description, json_build_object('id', ct.id, 'name', ct.name)::jsonb as user
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
				description = :Description
			where id = :Id and domain_id = :DomainId
			returning *
		)
		select u.id, u.status, u.state, u.domain_id, u.description, json_build_object('id', ct.id, 'name', ct.name)::jsonb as user
		from u
			inner join directory.wbt_user ct on ct.id = u.user_id
		order by id`, map[string]interface{}{
		"UserId":      agent.User.Id,
		"Description": agent.Description,
		"Id":          agent.Id,
		"DomainId":    agent.DomainId,
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

func (s SqlAgentStore) InTeam(domainId, id int64, offset, limit int) ([]*model.AgentInTeam, *model.AppError) {
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
        where s.agent_id = :AgentId
    )
) and t.domain_id = :DomainId
order by t.id
			limit :Limit
			offset :Offset`, map[string]interface{}{
		"AgentId":  id,
		"DomainId": domainId,
		"Limit":    limit,
		"Offset":   offset,
	})

	if err != nil {
		return nil, model.NewAppError("SqlAgentStore.InTeam", "store.sql_agent.get_team.app_error", nil,
			fmt.Sprintf("Id=%v, %s", id, err.Error()), extractCodeFromErr(err))
	}

	return res, nil
}

func (s SqlAgentStore) InQueue(domainId, id int64, offset, limit int) ([]*model.AgentInQueue, *model.AppError) {
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
            where s.agent_id = :AgentId
        )
    ) and t.domain_id = :DomainId
)
group by q.id, q.priority
order by q.priority desc
limit :Limit
offset :Offset`, map[string]interface{}{
		"AgentId":  id,
		"DomainId": domainId,
		"Limit":    limit,
		"Offset":   offset,
	})

	if err != nil {
		return nil, model.NewAppError("SqlAgentStore.InQueue", "store.sql_agent.get_queue.app_error", nil,
			fmt.Sprintf("Id=%v, %s", id, err.Error()), extractCodeFromErr(err))
	}

	return res, nil
}
