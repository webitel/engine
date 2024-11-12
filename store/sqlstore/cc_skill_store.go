package sqlstore

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/webitel/engine/auth_manager"

	"github.com/lib/pq"
	"github.com/webitel/engine/model"
	"github.com/webitel/engine/store"
)

type SqlSkillStore struct {
	SqlStore
}

func NewSqlSkillStore(sqlStore SqlStore) store.SkillStore {
	us := &SqlSkillStore{sqlStore}
	return us
}

func (s *SqlSkillStore) CheckAccess(ctx context.Context, domainId, id int64, groups []int, access auth_manager.PermissionAccess) (bool, model.AppError) {
	res, err := s.GetReplica().WithContext(ctx).SelectNullInt(`select 1
		where exists(
          select 1
          from call_center.cc_skill_acl a
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
		return false, model.NewInternalError("store.sql_skill.access.app_error", fmt.Sprintf("id=%v, domain_id=%v %v", id, domainId, err.Error()))
	}

	return (res.Valid && res.Int64 == 1), nil
}

func (s *SqlSkillStore) Create(ctx context.Context, skill *model.Skill) (*model.Skill, model.AppError) {
	var out *model.Skill
	if err := s.GetMaster().WithContext(ctx).SelectOne(&out, `with s as (
    insert into call_center.cc_skill (name, domain_id, description,
                                      updated_at, created_at, created_by, updated_by)
        values (:Name, :DomainId, :Description, :UpdatedAt, :CreatedAt, :CreatedBy, :UpdatedBy)
        returning *)
SELECT s.id,
       s.created_at,
       call_center.cc_get_lookup(c.id, COALESCE(c.name, c.username::text)::character varying) AS created_by,
       s.updated_at,
       call_center.cc_get_lookup(u.id, COALESCE(u.name, u.username::text)::character varying) AS updated_by,
       s.name,
       s.description,
       s.domain_id,
       agents.active_agents,
       agents.total_agents
FROM s
         LEFT JOIN directory.wbt_user c ON c.id = s.created_by
         LEFT JOIN directory.wbt_user u ON u.id = s.updated_by
         LEFT JOIN LATERAL ( SELECT count(DISTINCT sa.agent_id) FILTER (WHERE sa.enabled) AS active_agents,
                                    count(DISTINCT sa.agent_id)                           AS total_agents
                             FROM call_center.cc_skill_in_agent sa
                             WHERE sa.skill_id = s.id) agents ON true`,
		map[string]interface{}{
			"Name":        skill.Name,
			"DomainId":    skill.DomainId,
			"Description": skill.Description,
			"CreatedAt":   skill.CreatedAt,
			"CreatedBy":   skill.CreatedBy.GetSafeId(),
			"UpdatedAt":   skill.UpdatedAt,
			"UpdatedBy":   skill.UpdatedBy.GetSafeId(),
		}); nil != err {
		return nil, model.NewInternalError("store.sql_skill.save.app_error", fmt.Sprintf("name=%v, %v", skill.Name, err.Error()))
	} else {
		return out, nil
	}

}

func (s *SqlSkillStore) Get(ctx context.Context, domainId int64, id int64) (*model.Skill, model.AppError) {
	var skill *model.Skill
	if err := s.GetReplica().WithContext(ctx).SelectOne(&skill, `SELECT s.id,
       s.created_at,
       call_center.cc_get_lookup(c.id, COALESCE(c.name, c.username::text)::character varying) AS created_by,
       s.updated_at,
       call_center.cc_get_lookup(u.id, COALESCE(u.name, u.username::text)::character varying) AS updated_by,
       s.name,
       s.description,
       s.domain_id,
       agents.active_agents,
       agents.total_agents
FROM call_center.cc_skill s
         LEFT JOIN directory.wbt_user c ON c.id = s.created_by
         LEFT JOIN directory.wbt_user u ON u.id = s.updated_by
         LEFT JOIN LATERAL ( SELECT count(DISTINCT sa.agent_id) FILTER (WHERE sa.enabled) AS active_agents,
                                    count(DISTINCT sa.agent_id)                           AS total_agents
                             FROM call_center.cc_skill_in_agent sa
                             WHERE sa.skill_id = s.id) agents ON true
		where s.id = :Id and s.domain_id = :DomainId`, map[string]interface{}{
		"Id":       id,
		"DomainId": domainId,
	}); err != nil {
		if err == sql.ErrNoRows {
			return nil, model.NewNotFoundError("store.sql_skill.get.app_error", fmt.Sprintf("Id=%v, %s", id, err.Error()))
		} else {
			return nil, model.NewInternalError("store.sql_skill.get.app_error", fmt.Sprintf("Id=%v, %s", id, err.Error()))
		}
	} else {
		return skill, nil
	}
}

func (s *SqlSkillStore) GetAllPage(ctx context.Context, domainId int64, search *model.SearchSkill) ([]*model.Skill, model.AppError) {
	var skills []*model.Skill

	f := map[string]interface{}{
		"NotExistsAgent": search.NotExistsAgent,
		"DomainId":       domainId,
		"Ids":            pq.Array(search.Ids),
		"Q":              search.GetQ(),
	}

	err := s.ListQuery(ctx, &skills, search.ListRequest,
		`domain_id = :DomainId
				and (:Ids::int[] isnull or id = any(:Ids))
				and (:NotExistsAgent::int isnull or not exists(select 1 from call_center.cc_skill_in_agent sa where sa.agent_id = :NotExistsAgent and sa.skill_id = t.id))
				and (:Q::varchar isnull or (name ilike :Q::varchar or description ilike :Q::varchar))`,
		model.Skill{}, f)

	if err != nil {
		return nil, model.NewInternalError("store.sql_skill.get_all.app_error", err.Error())
	} else {
		return skills, nil
	}
}

func (s *SqlSkillStore) GetAllPageByGroups(ctx context.Context, domainId int64, groups []int, search *model.SearchSkill) ([]*model.Skill, model.AppError) {
	var skills []*model.Skill

	f := map[string]interface{}{
		"DomainId":       domainId,
		"NotExistsAgent": search.NotExistsAgent,
		"Ids":            pq.Array(search.Ids),
		"Q":              search.GetQ(),
		"Groups":         pq.Array(groups),
		"Access":         auth_manager.PERMISSION_ACCESS_READ.Value(),
	}

	err := s.ListQuery(ctx, &skills, search.ListRequest,
		`domain_id = :DomainId
				and (exists(select 1
				  from call_center.cc_skill_acl a
				  where a.dc = t.domain_id and a.object = t.id and a.subject = any(:Groups::int[]) and a.access&:Access = :Access)
			  	) 
				and (:NotExistsAgent::int isnull or not exists(select 1 from call_center.cc_skill_in_agent sa where sa.agent_id = :NotExistsAgent and sa.skill_id = t.id))
				and (:Ids::int[] isnull or id = any(:Ids))
				and (:Q::varchar isnull or (name ilike :Q::varchar or description ilike :Q::varchar))`,
		model.Skill{}, f)

	if err != nil {
		return nil, model.NewInternalError("store.sql_skill.get_all.app_error", err.Error())
	} else {
		return skills, nil
	}
}

func (s *SqlSkillStore) Delete(ctx context.Context, domainId int64, id int64) model.AppError {
	if _, err := s.GetMaster().WithContext(ctx).Exec(`delete from call_center.cc_skill c where c.id=:Id and c.domain_id = :DomainId`,
		map[string]interface{}{"Id": id, "DomainId": domainId}); err != nil {
		return model.NewInternalError("store.sql_skill.delete.app_error", fmt.Sprintf("Id=%v, %s", id, err.Error()))
	}
	return nil
}

func (s *SqlSkillStore) Update(ctx context.Context, skill *model.Skill) (*model.Skill, model.AppError) {
	err := s.GetMaster().WithContext(ctx).SelectOne(&skill, `with s as (
    update call_center.cc_skill
        set name = :Name,
            description = :Description,
            updated_at = :UpdatedAt,
            updated_by = :UpdatedBy
        where id = :Id and domain_id = :DomainId returning *)
SELECT s.id,
       s.created_at,
       call_center.cc_get_lookup(c.id, COALESCE(c.name, c.username::text)::character varying) AS created_by,
       s.updated_at,
       call_center.cc_get_lookup(u.id, COALESCE(u.name, u.username::text)::character varying) AS updated_by,
       s.name,
       s.description,
       s.domain_id,
       agents.active_agents,
       agents.total_agents
FROM s
         LEFT JOIN directory.wbt_user c ON c.id = s.created_by
         LEFT JOIN directory.wbt_user u ON u.id = s.updated_by
         LEFT JOIN LATERAL ( SELECT count(DISTINCT sa.agent_id) FILTER (WHERE sa.enabled) AS active_agents,
                                    count(DISTINCT sa.agent_id)                           AS total_agents
                             FROM call_center.cc_skill_in_agent sa
                             WHERE sa.skill_id = s.id) agents ON true`, map[string]interface{}{
		"Id":          skill.Id,
		"Name":        skill.Name,
		"Description": skill.Description,
		"DomainId":    skill.DomainId,
		"UpdatedAt":   skill.UpdatedAt,
		"UpdatedBy":   skill.UpdatedBy.GetSafeId(),
	})
	if err != nil {
		return nil, model.NewInternalError("store.sql_skill.update.app_error", fmt.Sprintf("Id=%v, %s", skill.Id, err.Error()))
	}
	return skill, nil
}
