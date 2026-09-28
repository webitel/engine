package sqlstore

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/lib/pq"
	"github.com/webitel/engine/model"
	"github.com/webitel/engine/pkg/wbt/auth_manager"
	"github.com/webitel/engine/store"
)

type SqlSkillGroupStore struct {
	SqlStore
}

func NewSqlSkillGroupStore(sqlStore SqlStore) store.SkillGroupStore {
	return &SqlSkillGroupStore{sqlStore}
}

func (s *SqlSkillGroupStore) CheckAccess(ctx context.Context, domainId, id int64, groups []int, access auth_manager.PermissionAccess) (bool, model.AppError) {
	res, err := s.GetReplica().WithContext(ctx).SelectNullInt(`select 1
		where exists(
          select 1
          from call_center.cc_skill_group_acl a
          where a.dc = :DomainId
            and a.object = :Id
            and a.subject = any (:Groups::int[])
            and a.access & :Access = :Access
        )`, map[string]interface{}{
		"DomainId": domainId,
		"Id":       id,
		"Groups":   pq.Array(groups),
		"Access":   access.Value(),
	})
	if err != nil {
		return false, model.NewInternalError("store.sql_skill_group.access.app_error", fmt.Sprintf("id=%v, domain_id=%v %v", id, domainId, err.Error()))
	}
	return res.Valid && res.Int64 == 1, nil
}

func (s *SqlSkillGroupStore) NameExists(ctx context.Context, domainId int64, name string, excludeId int64) (bool, model.AppError) {
	res, err := s.GetReplica().WithContext(ctx).SelectNullInt(`select 1
		where exists(
		  select 1
		  from call_center.cc_skill_group g
		  where g.domain_id = :DomainId
		    and lower(g.name) = lower(:Name)
		    and g.id != :ExcludeId
		)`, map[string]interface{}{
		"DomainId":  domainId,
		"Name":      name,
		"ExcludeId": excludeId,
	})
	if err != nil {
		return false, model.NewInternalError("store.sql_skill_group.name_exists.app_error", err.Error())
	}
	return res.Valid && res.Int64 == 1, nil
}

func (s *SqlSkillGroupStore) Create(ctx context.Context, group *model.SkillGroup) (*model.SkillGroup, model.AppError) {
	var id int64
	err := s.GetMaster().WithContext(ctx).SelectOne(&id, `insert into call_center.cc_skill_group (name, domain_id, description,
                                          created_at, created_by, updated_at, updated_by)
        values (:Name, :DomainId, :Description, :CreatedAt, :CreatedBy, :UpdatedAt, :UpdatedBy)
        returning id`,
		map[string]interface{}{
			"Name":        group.Name,
			"DomainId":    group.DomainId,
			"Description": group.Description,
			"CreatedAt":   group.CreatedAt,
			"CreatedBy":   group.CreatedBy.GetSafeId(),
			"UpdatedAt":   group.UpdatedAt,
			"UpdatedBy":   group.UpdatedBy.GetSafeId(),
		})
	if err != nil {
		return nil, model.NewInternalError("store.sql_skill_group.save.app_error", fmt.Sprintf("name=%v, %v", group.Name, err.Error()))
	}

	return s.Get(ctx, group.DomainId, id)
}

func (s *SqlSkillGroupStore) Get(ctx context.Context, domainId int64, id int64) (*model.SkillGroup, model.AppError) {
	var group *model.SkillGroup
	err := s.GetReplica().WithContext(ctx).SelectOne(&group, `select id, domain_id, name, description, skills, active_agents, total_agents,
       created_at, created_by, updated_at, updated_by
from call_center.cc_skill_group_view
where id = :Id and domain_id = :DomainId`, map[string]interface{}{
		"Id":       id,
		"DomainId": domainId,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, model.NewNotFoundError("store.sql_skill_group.get.app_error", fmt.Sprintf("Id=%v, %s", id, err.Error()))
		}
		return nil, model.NewInternalError("store.sql_skill_group.get.app_error", fmt.Sprintf("Id=%v, %s", id, err.Error()))
	}
	return group, nil
}

func (s *SqlSkillGroupStore) GetAllPage(ctx context.Context, domainId int64, search *model.SearchSkillGroup) ([]*model.SkillGroup, model.AppError) {
	var groups []*model.SkillGroup

	f := map[string]interface{}{
		"DomainId": domainId,
		"Ids":      pq.Array(search.Ids),
		"Q":        search.GetQ(),
	}

	err := s.ListQuery(ctx, &groups, search.ListRequest,
		`domain_id = :DomainId
				and (:Ids::int[] isnull or id = any(:Ids))
				and (:Q::varchar isnull or (name ilike :Q::varchar or description ilike :Q::varchar))`,
		model.SkillGroup{}, f)
	if err != nil {
		return nil, model.NewInternalError("store.sql_skill_group.get_all.app_error", err.Error())
	}
	return groups, nil
}

func (s *SqlSkillGroupStore) GetAllPageByGroups(ctx context.Context, domainId int64, groups []int, search *model.SearchSkillGroup) ([]*model.SkillGroup, model.AppError) {
	var list []*model.SkillGroup

	f := map[string]interface{}{
		"DomainId": domainId,
		"Ids":      pq.Array(search.Ids),
		"Q":        search.GetQ(),
		"Groups":   pq.Array(groups),
		"Access":   auth_manager.PERMISSION_ACCESS_READ.Value(),
	}

	err := s.ListQuery(ctx, &list, search.ListRequest,
		`domain_id = :DomainId
				and (exists(select 1
				  from call_center.cc_skill_group_acl a
				  where a.dc = t.domain_id and a.object = t.id and a.subject = any(:Groups::int[]) and a.access&:Access = :Access)
				  )
				and (:Ids::int[] isnull or id = any(:Ids))
				and (:Q::varchar isnull or (name ilike :Q::varchar or description ilike :Q::varchar))`,
		model.SkillGroup{}, f)
	if err != nil {
		return nil, model.NewInternalError("store.sql_skill_group.get_all.app_error", err.Error())
	}
	return list, nil
}

func (s *SqlSkillGroupStore) Update(ctx context.Context, group *model.SkillGroup) (*model.SkillGroup, model.AppError) {
	_, err := s.GetMaster().WithContext(ctx).Exec(`update call_center.cc_skill_group
        set name = :Name,
            description = :Description,
            updated_at = :UpdatedAt,
            updated_by = :UpdatedBy
        where id = :Id and domain_id = :DomainId`, map[string]interface{}{
		"Id":          group.Id,
		"Name":        group.Name,
		"Description": group.Description,
		"DomainId":    group.DomainId,
		"UpdatedAt":   group.UpdatedAt,
		"UpdatedBy":   group.UpdatedBy.GetSafeId(),
	})
	if err != nil {
		return nil, model.NewInternalError("store.sql_skill_group.update.app_error", fmt.Sprintf("Id=%v, %s", group.Id, err.Error()))
	}
	return s.Get(ctx, group.DomainId, group.Id)
}

func (s *SqlSkillGroupStore) Delete(ctx context.Context, domainId int64, id int64) model.AppError {
	agentIds, appErr := s.groupAgentIds(ctx, id)
	if appErr != nil {
		return appErr
	}

	if _, err := s.GetMaster().WithContext(ctx).Exec(`delete from call_center.cc_skill_group c where c.id=:Id and c.domain_id = :DomainId`,
		map[string]interface{}{"Id": id, "DomainId": domainId}); err != nil {
		return model.NewInternalError("store.sql_skill_group.delete.app_error", fmt.Sprintf("Id=%v, %s", id, err.Error()))
	}

	return s.reconcileAgents(ctx, agentIds)
}

func (s *SqlSkillGroupStore) SkillsSearch(ctx context.Context, domainId int64, search *model.SearchSkillGroupSkill) ([]*model.SkillGroupSkill, model.AppError) {
	var list []*model.SkillGroupSkill
	_, err := s.GetReplica().WithContext(ctx).Select(&list, `select sig.id,
       call_center.cc_get_lookup(cs.id::bigint, cs.name) as skill,
       sig.capacity
from call_center.cc_skill_in_group sig
    join call_center.cc_skill cs on cs.id = sig.skill_id
    join call_center.cc_skill_group g on g.id = sig.skill_group_id
where sig.skill_group_id = :GroupId
  and g.domain_id = :DomainId
  and (:Ids::int8[] isnull or sig.id = any(:Ids))
  and (:Q::varchar isnull or cs.name ilike :Q::varchar)
order by cs.name
limit :Limit offset :Offset`, map[string]interface{}{
		"GroupId":  search.SkillGroupId,
		"DomainId": domainId,
		"Ids":      pq.Array(search.Ids),
		"Q":        search.GetQ(),
		"Limit":    search.GetLimit(),
		"Offset":   search.GetOffset(),
	})
	if err != nil {
		return nil, model.NewInternalError("store.sql_skill_group.skills.search.app_error", err.Error())
	}
	return list, nil
}

func (s *SqlSkillGroupStore) SkillsCreate(ctx context.Context, domainId int64, userId int64, groupId int64, skills []*model.SkillGroupSkill) ([]*model.SkillGroupSkill, model.AppError) {
	skillIds := make([]int64, 0, len(skills))
	capacities := make([]int64, 0, len(skills))

	for _, sk := range skills {
		skillIds = append(skillIds, int64(sk.Skill.Id))
		capacities = append(capacities, int64(safeCapacity(sk.Capacity)))
	}

	var list []*model.SkillGroupSkill
	_, err := s.GetMaster().WithContext(ctx).Select(&list, `with ins as (
    insert into call_center.cc_skill_in_group (skill_group_id, skill_id, capacity, created_by, updated_by)
        select :GroupId, x.skill_id, x.capacity, :UserId, :UserId
        from unnest(:SkillIds::int[], :Capacities::int[]) as x(skill_id, capacity)
                 join call_center.cc_skill cs on cs.id = x.skill_id and cs.domain_id = :DomainId
                 join call_center.cc_skill_group g on g.id = :GroupId and g.domain_id = :DomainId
        on conflict (skill_group_id, skill_id) do update set capacity   = excluded.capacity,
                                                             updated_by = excluded.updated_by,
                                                             updated_at = now()
        returning id, skill_id, capacity)
select ins.id, call_center.cc_get_lookup(cs.id::bigint, cs.name) as skill, ins.capacity
from ins
         join call_center.cc_skill cs on cs.id = ins.skill_id`, map[string]interface{}{
		"GroupId":    groupId,
		"DomainId":   domainId,
		"UserId":     userId,
		"SkillIds":   pq.Array(skillIds),
		"Capacities": pq.Array(capacities),
	})
	if err != nil {
		return nil, model.NewInternalError("store.sql_skill_group.skills.create.app_error", err.Error())
	}

	if appErr := s.reconcileGroup(ctx, groupId); appErr != nil {
		return nil, appErr
	}
	return list, nil
}

func (s *SqlSkillGroupStore) SkillUpdate(ctx context.Context, domainId int64, userId int64, groupId int64, skill *model.SkillGroupSkill) (*model.SkillGroupSkill, model.AppError) {
	var out *model.SkillGroupSkill
	err := s.GetMaster().WithContext(ctx).SelectOne(&out, `with upd as (
    update call_center.cc_skill_in_group sig
        set capacity = :Capacity, updated_by = :UserId, updated_at = now()
        from call_center.cc_skill_group g
        where sig.id = :Id and sig.skill_group_id = :GroupId and g.id = sig.skill_group_id and g.domain_id = :DomainId
        returning sig.id, sig.skill_id, sig.capacity)
select upd.id, call_center.cc_get_lookup(cs.id::bigint, cs.name) as skill, upd.capacity
from upd join call_center.cc_skill cs on cs.id = upd.skill_id`, map[string]interface{}{
		"Id":       skill.Id,
		"GroupId":  groupId,
		"DomainId": domainId,
		"UserId":   userId,
		"Capacity": safeCapacity(skill.Capacity),
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, model.NewNotFoundError("store.sql_skill_group.skills.update.not_found", fmt.Sprintf("Id=%v", skill.Id))
		}
		return nil, model.NewInternalError("store.sql_skill_group.skills.update.app_error", err.Error())
	}

	if appErr := s.reconcileGroup(ctx, groupId); appErr != nil {
		return nil, appErr
	}
	return out, nil
}

func (s *SqlSkillGroupStore) SkillsDelete(ctx context.Context, domainId int64, groupId int64, ids []int64, skillIds []int64) ([]*model.SkillGroupSkill, model.AppError) {
	var list []*model.SkillGroupSkill
	_, err := s.GetMaster().WithContext(ctx).Select(&list, `with del as (
    delete from call_center.cc_skill_in_group sig
        using call_center.cc_skill_group g
        where sig.skill_group_id = :GroupId and g.id = sig.skill_group_id and g.domain_id = :DomainId
            and ((:Ids::int8[] notnull and sig.id = any(:Ids)) or (:SkillIds::int[] notnull and sig.skill_id = any(:SkillIds)))
        returning sig.id, sig.skill_id, sig.capacity)
select del.id, call_center.cc_get_lookup(cs.id::bigint, cs.name) as skill, del.capacity
from del join call_center.cc_skill cs on cs.id = del.skill_id`, map[string]interface{}{
		"GroupId":  groupId,
		"DomainId": domainId,
		"Ids":      pq.Array(ids),
		"SkillIds": pq.Array(skillIds),
	})
	if err != nil {
		return nil, model.NewInternalError("store.sql_skill_group.skills.delete.app_error", err.Error())
	}

	if appErr := s.reconcileGroup(ctx, groupId); appErr != nil {
		return nil, appErr
	}
	return list, nil
}

func (s *SqlSkillGroupStore) AgentsSearch(ctx context.Context, domainId int64, search *model.SearchSkillGroupAgent) ([]*model.SkillGroupAgent, model.AppError) {
	var list []*model.SkillGroupAgent
	_, err := s.GetReplica().WithContext(ctx).Select(&list, `select aig.id,
       call_center.cc_get_lookup(a.id::bigint, coalesce(u.name, u.username)::varchar) as agent,
       call_center.cc_get_lookup(t.id, t.name) as team
from call_center.cc_agent_in_skill_group aig
    join call_center.cc_agent a on a.id = aig.agent_id
    left join call_center.cc_team t on t.id = a.team_id
    left join directory.wbt_user u on u.id = a.user_id
where aig.skill_group_id = :GroupId
  and a.domain_id = :DomainId
  and (:Ids::int8[] isnull or aig.id = any(:Ids))
  and (:AgentIds::int8[] isnull or aig.agent_id = any(:AgentIds))
  and (:Q::varchar isnull or coalesce(u.name, u.username) ilike :Q::varchar)
order by agent ->> 'name'
limit :Limit offset :Offset`, map[string]interface{}{
		"GroupId":  search.SkillGroupId,
		"DomainId": domainId,
		"Ids":      pq.Array(search.Ids),
		"AgentIds": pq.Array(search.AgentIds),
		"Q":        search.GetQ(),
		"Limit":    search.GetLimit(),
		"Offset":   search.GetOffset(),
	})
	if err != nil {
		return nil, model.NewInternalError("store.sql_skill_group.agents.search.app_error", err.Error())
	}
	return list, nil
}

func (s *SqlSkillGroupStore) AgentsCreate(ctx context.Context, domainId int64, userId int64, groupIds []int64, agentIds []int64) ([]*model.SkillGroupAgent, model.AppError) {
	_, err := s.GetMaster().WithContext(ctx).Exec(`insert into call_center.cc_agent_in_skill_group (skill_group_id, agent_id, created_by, updated_by)
        select g.id, a.id, :UserId, :UserId
        from unnest(:GroupIds::int8[]) as gg(id)
                 join call_center.cc_skill_group g on g.id = gg.id and g.domain_id = :DomainId
                 join unnest(:AgentIds::int8[]) as aa(id) on true
                 join call_center.cc_agent a on a.id = aa.id and a.domain_id = :DomainId
        on conflict (skill_group_id, agent_id) do nothing`, map[string]interface{}{
		"UserId":   userId,
		"DomainId": domainId,
		"GroupIds": pq.Array(groupIds),
		"AgentIds": pq.Array(agentIds),
	})
	if err != nil {
		return nil, model.NewInternalError("store.sql_skill_group.agents.create.app_error", err.Error())
	}

	if appErr := s.reconcileAgents(ctx, agentIds); appErr != nil {
		return nil, appErr
	}

	var list []*model.SkillGroupAgent
	_, err = s.GetReplica().WithContext(ctx).Select(&list, `select distinct on (a.id) aig.id,
       call_center.cc_get_lookup(a.id::bigint, coalesce(u.name, u.username)::varchar) as agent,
       call_center.cc_get_lookup(t.id, t.name) as team
from call_center.cc_agent_in_skill_group aig
    join call_center.cc_agent a on a.id = aig.agent_id
    left join call_center.cc_team t on t.id = a.team_id
    left join directory.wbt_user u on u.id = a.user_id
where aig.skill_group_id = any(:GroupIds::int8[])
  and a.domain_id = :DomainId
  and aig.agent_id = any(:AgentIds::int8[])`, map[string]interface{}{
		"DomainId": domainId,
		"GroupIds": pq.Array(groupIds),
		"AgentIds": pq.Array(agentIds),
	})
	if err != nil {
		return nil, model.NewInternalError("store.sql_skill_group.agents.create.list.app_error", err.Error())
	}
	return list, nil
}

func (s *SqlSkillGroupStore) AgentsDelete(ctx context.Context, domainId int64, groupId int64, ids []int64, agentIds []int64) ([]*model.SkillGroupAgent, model.AppError) {
	var list []*model.SkillGroupAgent
	_, err := s.GetMaster().WithContext(ctx).Select(&list, `with del as (
    delete from call_center.cc_agent_in_skill_group aig
        using call_center.cc_agent a
        where aig.skill_group_id = :GroupId and a.id = aig.agent_id and a.domain_id = :DomainId
            and ((:Ids::int8[] notnull and aig.id = any(:Ids)) or (:AgentIds::int8[] notnull and aig.agent_id = any(:AgentIds)))
        returning aig.id, aig.agent_id)
select del.id,
       call_center.cc_get_lookup(a.id::bigint, coalesce(u.name, u.username)::varchar) as agent,
       call_center.cc_get_lookup(t.id, t.name) as team
from del
         join call_center.cc_agent a on a.id = del.agent_id
         left join call_center.cc_team t on t.id = a.team_id
         left join directory.wbt_user u on u.id = a.user_id`, map[string]interface{}{
		"GroupId":  groupId,
		"DomainId": domainId,
		"Ids":      pq.Array(ids),
		"AgentIds": pq.Array(agentIds),
	})
	if err != nil {
		return nil, model.NewInternalError("store.sql_skill_group.agents.delete.app_error", err.Error())
	}

	removed := make([]int64, 0, len(list))
	for _, a := range list {
		if a.Agent != nil {
			removed = append(removed, int64(a.Agent.Id))
		}
	}
	if appErr := s.reconcileAgents(ctx, removed); appErr != nil {
		return nil, appErr
	}
	return list, nil
}

func (s *SqlSkillGroupStore) AgentGroupsSearch(ctx context.Context, domainId int64, search *model.SearchAgentSkillGroup) ([]*model.AgentSkillGroup, model.AppError) {
	var list []*model.AgentSkillGroup
	_, err := s.GetReplica().WithContext(ctx).Select(&list, `select aig.id,
       call_center.cc_get_lookup(g.id, g.name) as skill_group,
       g.description,
       coalesce((select jsonb_agg(call_center.cc_get_lookup(cs.id::bigint, cs.name) order by cs.name)
                 from call_center.cc_skill_in_group sig
                          join call_center.cc_skill cs on cs.id = sig.skill_id
                 where sig.skill_group_id = g.id), '[]'::jsonb) as skills
from call_center.cc_agent_in_skill_group aig
    join call_center.cc_skill_group g on g.id = aig.skill_group_id
where aig.agent_id = :AgentId
  and g.domain_id = :DomainId
  and (:Ids::int8[] isnull or aig.id = any(:Ids))
  and (:Q::varchar isnull or (g.name ilike :Q::varchar or g.description ilike :Q::varchar))
order by g.name
limit :Limit offset :Offset`, map[string]interface{}{
		"AgentId":  search.AgentId,
		"DomainId": domainId,
		"Ids":      pq.Array(search.Ids),
		"Q":        search.GetQ(),
		"Limit":    search.GetLimit(),
		"Offset":   search.GetOffset(),
	})
	if err != nil {
		return nil, model.NewInternalError("store.sql_skill_group.agent_groups.search.app_error", err.Error())
	}
	return list, nil
}

func (s *SqlSkillGroupStore) AgentGroupsDelete(ctx context.Context, domainId int64, agentId int64, ids []int64, groupIds []int64) ([]*model.AgentSkillGroup, model.AppError) {
	var list []*model.AgentSkillGroup
	_, err := s.GetMaster().WithContext(ctx).Select(&list, `with del as (
    delete from call_center.cc_agent_in_skill_group aig
        using call_center.cc_agent a
        where aig.agent_id = :AgentId and a.id = aig.agent_id and a.domain_id = :DomainId
            and ((:Ids::int8[] notnull and aig.id = any(:Ids)) or (:GroupIds::int8[] notnull and aig.skill_group_id = any(:GroupIds)))
        returning aig.id, aig.skill_group_id)
select del.id, call_center.cc_get_lookup(g.id, g.name) as skill_group, g.description,
       coalesce((select jsonb_agg(call_center.cc_get_lookup(cs.id::bigint, cs.name) order by cs.name)
                 from call_center.cc_skill_in_group sig
                          join call_center.cc_skill cs on cs.id = sig.skill_id
                 where sig.skill_group_id = g.id), '[]'::jsonb) as skills
from del join call_center.cc_skill_group g on g.id = del.skill_group_id`, map[string]interface{}{
		"AgentId":  agentId,
		"DomainId": domainId,
		"Ids":      pq.Array(ids),
		"GroupIds": pq.Array(groupIds),
	})
	if err != nil {
		return nil, model.NewInternalError("store.sql_skill_group.agent_groups.delete.app_error", err.Error())
	}

	if appErr := s.reconcileAgents(ctx, []int64{agentId}); appErr != nil {
		return nil, appErr
	}
	return list, nil
}

func (s *SqlSkillGroupStore) groupAgentIds(ctx context.Context, groupId int64) ([]int64, model.AppError) {
	var ids []int64
	_, err := s.GetReplica().WithContext(ctx).Select(&ids,
		`select agent_id from call_center.cc_agent_in_skill_group where skill_group_id = :GroupId`,
		map[string]interface{}{"GroupId": groupId})
	if err != nil {
		return nil, model.NewInternalError("store.sql_skill_group.group_agents.app_error", err.Error())
	}
	return ids, nil
}

func (s *SqlSkillGroupStore) reconcileAgents(ctx context.Context, agentIds []int64) model.AppError {
	if len(agentIds) == 0 {
		return nil
	}
	_, err := s.GetMaster().WithContext(ctx).Exec(
		`select call_center.cc_skill_group_reconcile_agent(a::int) from unnest(:AgentIds::int8[]) a`,
		map[string]interface{}{"AgentIds": pq.Array(agentIds)})
	if err != nil {
		return model.NewInternalError("store.sql_skill_group.reconcile_agent.app_error", err.Error())
	}
	return nil
}

func (s *SqlSkillGroupStore) reconcileGroup(ctx context.Context, groupId int64) model.AppError {
	_, err := s.GetMaster().WithContext(ctx).Exec(
		`select call_center.cc_skill_group_reconcile_group(:GroupId)`,
		map[string]interface{}{"GroupId": groupId})
	if err != nil {
		return model.NewInternalError("store.sql_skill_group.reconcile_group.app_error", err.Error())
	}
	return nil
}

func safeCapacity(c *int) int {
	if c == nil {
		return 0
	}
	return *c
}
