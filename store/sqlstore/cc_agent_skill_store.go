package sqlstore

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/go-gorp/gorp"
	"github.com/lib/pq"
	"github.com/webitel/engine/model"
	"github.com/webitel/engine/store"
)

type SqlAgentSkillStore struct {
	SqlStore
}

func NewSqlAgentSkillStore(sqlStore SqlStore) store.AgentSkillStore {
	us := &SqlAgentSkillStore{sqlStore}
	return us
}

func (s SqlAgentSkillStore) Create(ctx context.Context, in *model.AgentSkill) (*model.AgentSkill, model.AppError) {
	var out *model.AgentSkill
	if err := s.GetMaster().WithContext(ctx).SelectOne(&out, `with tmp as (
    insert into call_center.cc_skill_in_agent (skill_id, agent_id, capacity, created_at, created_by, updated_at, updated_by, enabled)
    values (:SkillId, :AgentId, :Capacity, :CreatedAt, :CreatedBy, :UpdatedAt, :UpdatedBy, :Enabled)
    returning *
)
select tmp.id, call_center.cc_get_lookup(s.id, s.name) as skill, call_center.cc_get_lookup(a.id, wu.name) as agent, tmp.capacity, tmp.created_at,
	call_center.cc_get_lookup(c.id, c.name) as created_by, tmp.updated_at, call_center.cc_get_lookup(u.id, u.name) as updated_by, tmp.enabled
from tmp
    inner join call_center.cc_skill s on s.id = tmp.skill_id
    inner join call_center.cc_agent a on a.id = tmp.agent_id
    inner join directory.wbt_user wu on a.user_id = wu.id
    left join directory.wbt_user c on c.id = tmp.created_by
    left join directory.wbt_user u on u.id = tmp.updated_by`,
		map[string]interface{}{
			"SkillId":   in.Skill.Id,
			"AgentId":   in.Agent.Id,
			"Capacity":  in.Capacity,
			"CreatedAt": in.CreatedAt,
			"CreatedBy": in.CreatedBy.GetSafeId(),
			"UpdatedAt": in.UpdatedAt,
			"UpdatedBy": in.UpdatedBy.GetSafeId(),
			"Enabled":   in.Enabled,
		}); err != nil {
		return nil, model.NewCustomCodeError("store.sql_skill_in_agent.create.app_error", fmt.Sprintf("AgentId=%v, SkillId=%v %s", in.Agent.Id, in.Skill.Id, err.Error()), extractCodeFromErr(err))
	} else {
		return out, nil
	}
}

func (s SqlAgentSkillStore) BulkCreate(ctx context.Context, domainId, agentId int64, skills []*model.AgentSkill) ([]int64, model.AppError) {
	var err error
	var stmp *sql.Stmt
	var tx *gorp.Transaction
	tx, err = s.GetMaster().Begin()
	if err != nil {
		return nil, model.NewInternalError("store.sql_skill_in_agent.bulk_save.app_error", err.Error())
	}

	_, err = tx.WithContext(ctx).Exec("CREATE temp table cc_skill_in_agent_tmp ON COMMIT DROP as table call_center.cc_skill_in_agent with no data")
	if err != nil {
		return nil, model.NewInternalError("store.sql_skill_in_agent.bulk_save.app_error", err.Error())
	}

	stmp, err = tx.Prepare(pq.CopyIn("cc_skill_in_agent_tmp", "id", "skill_id", "agent_id", "capacity", "created_at", "created_by",
		"updated_at", "updated_by", "enabled"))
	if err != nil {
		return nil, model.NewInternalError("store.sql_skill_in_agent.bulk_save.app_error", err.Error())
	}

	defer stmp.Close()
	result := make([]int64, 0, len(skills))
	for k, v := range skills {
		_, err = stmp.Exec(k, v.Skill.GetSafeId(), agentId, v.Capacity, v.CreatedAt, v.CreatedBy.GetSafeId(), v.CreatedAt,
			v.CreatedBy.GetSafeId(), v.Enabled)
		if err != nil {
			goto _error
		}
	}

	_, err = stmp.Exec()
	if err != nil {
		goto _error
	} else {

		_, err = tx.Select(&result, `with i as (
			insert into call_center.cc_skill_in_agent (skill_id, agent_id, capacity, created_at, created_by, updated_at, updated_by, enabled)
			select skill_id, agent_id, capacity, created_at, created_by, updated_at, updated_by, enabled
			from cc_skill_in_agent_tmp t
				inner join call_center.cc_skill s on s.id = t.skill_id
			where s.domain_id = :DomainId
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
		return nil, model.NewCustomCodeError("store.sql_skill_in_agent.bulk_save.app_error", err.Error(), extractCodeFromErr(err))
	}

	return result, nil

_error:
	tx.Rollback()
	return nil, model.NewCustomCodeError("store.sql_skill_in_agent.bulk_save.app_error", err.Error(), extractCodeFromErr(err))
}

func (s SqlAgentSkillStore) GetAllPage(ctx context.Context, domainId int64, search *model.SearchAgentSkillList) ([]*model.AgentSkill, model.AppError) {
	var agentSkill []*model.AgentSkill

	f := map[string]interface{}{
		"DomainId": domainId,
		"AgentIds": pq.Array(search.AgentIds),
		"Ids":      pq.Array(search.Ids),
		"SkillIds": pq.Array(search.SkillIds),
		"Q":        search.GetQ(),
	}

	err := s.ListQuery(ctx, &agentSkill, search.ListRequest,
		`domain_id = :DomainId
				and (:AgentIds::int[] isnull or agent_id = any(:AgentIds))
				and (:Ids::int[] isnull or id = any(:Ids))
				and (:SkillIds::int[] isnull or skill_id = any(:SkillIds))
				and (:Q::varchar isnull or (skill_name ilike :Q::varchar or agent_name ilike :Q::varchar ))`,
		model.AgentSkill{}, f)

	if err != nil {
		return nil, model.NewInternalError("store.sql_skill_in_agent.get_all.app_error", err.Error())
	} else {
		return agentSkill, nil
	}
}

func (s SqlAgentSkillStore) HasDisabledSkill(ctx context.Context, domainId int64, skillId int64) (bool, model.AppError) {
	var exists bool
	err := s.GetReplica().WithContext(ctx).SelectOne(&exists, `select not exists(select 1
from call_center.cc_skill_in_agent_view s
where s.domain_id = :DomainId::int8
    and s.skill_id = :SkillId::int
    and s.enabled) as has_disabled`, map[string]interface{}{
		"DomainId": domainId,
		"SkillId":  skillId,
	})
	if err != nil {
		return false, model.NewCustomCodeError("store.sql_skill_in_agent.has_disabled.app_error", err.Error(), extractCodeFromErr(err))
	}

	return exists, nil
}

func (s SqlAgentSkillStore) GetById(ctx context.Context, domainId, agentId, id int64) (*model.AgentSkill, model.AppError) {
	var agentSkill *model.AgentSkill

	if err := s.GetReplica().WithContext(ctx).SelectOne(&agentSkill,
		`select tmp.id, call_center.cc_get_lookup(s.id, s.name) as skill, call_center.cc_get_lookup(a.id, wu.name) as agent, tmp.capacity, tmp.created_at,
	call_center.cc_get_lookup(c.id, c.name) as created_by, tmp.updated_at, call_center.cc_get_lookup(u.id, u.name) as updated_by, tmp.enabled
from call_center.cc_skill_in_agent tmp
    inner join call_center.cc_skill s on s.id = tmp.skill_id
    inner join call_center.cc_agent a on a.id = tmp.agent_id
    inner join directory.wbt_user wu on a.user_id = wu.id
    left join directory.wbt_user c on c.id = tmp.created_by
    left join directory.wbt_user u on u.id = tmp.updated_by
where tmp.id = :Id and tmp.agent_id = :AgentId and a.domain_id = :DomainId
`, map[string]interface{}{"DomainId": domainId, "Id": id, "AgentId": agentId}); err != nil {
		return nil, model.NewCustomCodeError("store.sql_skill_in_agent.get_all.app_error", err.Error(), extractCodeFromErr(err))
	} else {
		return agentSkill, nil
	}
}

func (s SqlAgentSkillStore) Update(ctx context.Context, agentSkill *model.AgentSkill) (*model.AgentSkill, model.AppError) {
	var out *model.AgentSkill
	err := s.GetMaster().WithContext(ctx).SelectOne(&out, `with tmp as (
    update call_center.cc_skill_in_agent s
        set updated_at = :UpdatedAt,
            updated_by = :UpdatedBy,
            skill_id = :SkillId,
            capacity = :Capacity,
			enabled = :Enabled
    where s.id = :Id and s.agent_id = :AgentId
    returning *
)
select tmp.id, call_center.cc_get_lookup(s.id, s.name) as skill, call_center.cc_get_lookup(a.id, wu.name) as agent, tmp.capacity, tmp.created_at,
	call_center.cc_get_lookup(c.id, c.name) as created_by, tmp.updated_at, call_center.cc_get_lookup(u.id, u.name) as updated_by, tmp.enabled
from tmp
    inner join call_center.cc_skill s on s.id = tmp.skill_id
    inner join call_center.cc_agent a on a.id = tmp.agent_id
    inner join directory.wbt_user wu on a.user_id = wu.id
    left join directory.wbt_user c on c.id = tmp.created_by
    left join directory.wbt_user u on u.id = tmp.updated_by`, map[string]interface{}{
		"UpdatedAt": agentSkill.UpdatedAt,
		"UpdatedBy": agentSkill.UpdatedBy.GetSafeId(),
		"SkillId":   agentSkill.Skill.Id,
		"Capacity":  agentSkill.Capacity,
		"Id":        agentSkill.Id,
		"AgentId":   agentSkill.Agent.Id,
		"Enabled":   agentSkill.Enabled,
	})
	if err != nil {
		return nil, model.NewCustomCodeError("store.sql_skill_in_agent.update.app_error", fmt.Sprintf("Id=%v, %s", agentSkill.Id, err.Error()), extractCodeFromErr(err))
	}
	return out, nil
}

func (s SqlAgentSkillStore) UpdateMany(ctx context.Context, domainId int64, search model.SearchAgentSkill, path model.AgentSkillPatch) ([]*model.AgentSkill, model.AppError) {
	var res []*model.AgentSkill

	_, err := s.GetMaster().WithContext(ctx).Select(&res, `with tmp as (
    update call_center.cc_skill_in_agent
        set capacity = coalesce(:Capacity, capacity),
            enabled = coalesce(:Enabled, enabled),
			skill_id = coalesce(:SkillId, skill_id),
            updated_by = :UpdatedBy,
            updated_at = :UpdatedAt
    where id in (
            select sa.id
            from call_center.cc_skill_in_agent sa
                inner join call_center.cc_skill s on s.id = sa.skill_id
            where s.domain_id = :DomainId
                            and (:AgentIds::int[] isnull or agent_id = any(:AgentIds))
                            and (:Ids::int[] isnull or sa.id = any(:Ids))
                            and (:SkillIds::int[] isnull or sa.skill_id = any(:SkillIds))
    )
   returning *
)
select tmp.id, call_center.cc_get_lookup(s.id, s.name) as skill, call_center.cc_get_lookup(a.id, wu.name) as agent, tmp.capacity, tmp.created_at,
	call_center.cc_get_lookup(c.id, c.name) as created_by, tmp.updated_at, call_center.cc_get_lookup(u.id, u.name) as updated_by, tmp.enabled
from tmp
    inner join call_center.cc_skill s on s.id = tmp.skill_id
    inner join call_center.cc_agent a on a.id = tmp.agent_id
    inner join directory.wbt_user wu on a.user_id = wu.id
    left join directory.wbt_user c on c.id = tmp.created_by
    left join directory.wbt_user u on u.id = tmp.updated_by`, map[string]interface{}{
		"DomainId":  domainId,
		"AgentIds":  pq.Array(search.AgentIds),
		"Ids":       pq.Array(search.Ids),
		"SkillIds":  pq.Array(search.SkillIds),
		"Capacity":  path.Capacity,
		"Enabled":   path.Enabled,
		"UpdatedBy": path.UpdatedBy.GetSafeId(),
		"UpdatedAt": path.UpdatedAt,
		"SkillId":   path.Skill.GetSafeId(),
	})

	if err != nil {
		return nil, model.NewCustomCodeError("store.sql_skill_in_agent.update_many.app_error", fmt.Sprintf("Query=%v, %s", search, err.Error()), extractCodeFromErr(err))
	}

	return res, nil
}

func (s SqlAgentSkillStore) DeleteById(ctx context.Context, agentId, id int64) model.AppError {
	if _, err := s.GetMaster().WithContext(ctx).Exec(`delete from call_center.cc_skill_in_agent a
where a.id = :Id and a.agent_id = :AgentId`,
		map[string]interface{}{"Id": id, "AgentId": agentId}); err != nil {
		return model.NewCustomCodeError("store.sql_skill_in_agent.delete.app_error", fmt.Sprintf("Id=%v, %s", id, err.Error()), extractCodeFromErr(err))
	}
	return nil
}

func (s SqlAgentSkillStore) Delete(ctx context.Context, domainId int64, search model.SearchAgentSkill) ([]*model.AgentSkill, model.AppError) {
	var res []*model.AgentSkill
	_, err := s.GetMaster().WithContext(ctx).Select(&res, `with tmp as (
    delete from call_center.cc_skill_in_agent
    where id in (
            select sa.id
            from call_center.cc_skill_in_agent sa
                inner join call_center.cc_skill s on s.id = sa.skill_id
            where s.domain_id = :DomainId
                            and (:AgentIds::int[] isnull or sa.agent_id = any(:AgentIds))
                            and (:Ids::int[] isnull or sa.id = any(:Ids))
                            and (:SkillIds::int[] isnull or sa.skill_id = any(:SkillIds))
    )
   returning *
)
select tmp.id, call_center.cc_get_lookup(s.id, s.name) as skill, call_center.cc_get_lookup(a.id, wu.name) as agent, tmp.capacity, tmp.created_at,
	call_center.cc_get_lookup(c.id, c.name) as created_by, tmp.updated_at, call_center.cc_get_lookup(u.id, u.name) as updated_by, tmp.enabled
from tmp
    inner join call_center.cc_skill s on s.id = tmp.skill_id
    inner join call_center.cc_agent a on a.id = tmp.agent_id
    inner join directory.wbt_user wu on a.user_id = wu.id
    left join directory.wbt_user c on c.id = tmp.created_by
    left join directory.wbt_user u on u.id = tmp.updated_by`, map[string]interface{}{
		"DomainId": domainId,
		"AgentIds": pq.Array(search.AgentIds),
		"Ids":      pq.Array(search.Ids),
		"SkillIds": pq.Array(search.SkillIds),
	})

	if err != nil {
		return nil, model.NewCustomCodeError("store.sql_skill_in_agent.delete.app_error", fmt.Sprintf("Query=%v, %s", search, err.Error()), extractCodeFromErr(err))
	}

	return res, nil
}

func (s SqlAgentSkillStore) LookupNotExistsAgent(ctx context.Context, domainId, agentId int64, search *model.SearchAgentSkillList) ([]*model.Skill, model.AppError) {
	var skills []*model.Skill

	if _, err := s.GetReplica().WithContext(ctx).Select(&skills,
		`select c.id,
       c.name,
       c.description
from call_center.cc_skill c
where c.domain_id = :DomainId and ( (:Q::varchar isnull or (c.name ilike :Q::varchar or c.description ilike :Q::varchar ) )) 
	and not exists(select 1 from call_center.cc_skill_in_agent sa where sa.agent_id = :AgentId and sa.skill_id = c.id)
order by id
limit :Limit
offset :Offset`, map[string]interface{}{
			"DomainId": domainId,
			"Limit":    search.GetLimit(),
			"Offset":   search.GetOffset(),
			"AgentId":  agentId,
			"Q":        search.GetQ(),
		}); err != nil {
		return nil, model.NewInternalError("store.sql_skill_in_agent.lookup.skill.app_error", err.Error())
	} else {
		return skills, nil
	}
}

func (s SqlAgentSkillStore) CreateMany(ctx context.Context, domainId int64, in *model.AgentsSkills) ([]*model.AgentSkill, model.AppError) {
	var items []*model.AgentSkill
	_, err := s.GetMaster().WithContext(ctx).Select(&items, `with tmp as (
    insert into call_center.cc_skill_in_agent (skill_id, agent_id, capacity, created_at, created_by, updated_at,
                                               updated_by, enabled)
        select s as skill_id,
               a as agent_id,
               :Capacity,
               :CreatedAt,
               :CreatedBy,
               :UpdatedAt,
               :UpdatedBy,
               :Enabled
        from unnest(:Agents::int[]) a,
             unnest(:Skill::int[]) s
        where exists(select 1 from call_center.cc_skill ss where ss.id = s and ss.domain_id = :DomainId)  
            and exists(select 1 from call_center.cc_agent aa where aa.id = a and aa.domain_id = :DomainId)
		returning *
	)
select tmp.id, call_center.cc_get_lookup(s.id, s.name) as skill, call_center.cc_get_lookup(a.id, wu.name) as agent, tmp.capacity, tmp.created_at,
	call_center.cc_get_lookup(c.id, c.name) as created_by, tmp.updated_at, call_center.cc_get_lookup(u.id, u.name) as updated_by, tmp.enabled
from tmp
    inner join call_center.cc_skill s on s.id = tmp.skill_id
    inner join call_center.cc_agent a on a.id = tmp.agent_id
    inner join directory.wbt_user wu on a.user_id = wu.id
    left join directory.wbt_user c on c.id = tmp.created_by
    left join directory.wbt_user u on u.id = tmp.updated_by`, map[string]interface{}{
		"DomainId":  domainId,
		"Agents":    pq.Array(in.AgentIds),
		"Skill":     pq.Array(in.SkillIds),
		"Capacity":  in.Capacity,
		"Enabled":   in.Enabled,
		"CreatedAt": in.CreatedAt,
		"CreatedBy": in.CreatedBy.GetSafeId(),
		"UpdatedAt": in.UpdatedAt,
		"UpdatedBy": in.UpdatedBy.GetSafeId(),
	})

	if err != nil {
		return nil, model.NewCustomCodeError("store.sql_skill_in_agent.create_many.app_error", fmt.Sprintf("AgentIds=%v, SkillIds=%v %s", in.AgentIds, in.SkillIds, err.Error()), extractCodeFromErr(err))
	}

	return items, nil
}
