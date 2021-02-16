package sqlstore

import (
	"fmt"
	"github.com/lib/pq"
	"github.com/webitel/engine/model"
	"github.com/webitel/engine/store"
)

type SqlPauseCauseStore struct {
	SqlStore
}

func NewSqlPauseCauseStore(sqlStore SqlStore) store.PauseCauseStore {
	us := &SqlPauseCauseStore{sqlStore}
	return us
}

func (s SqlPauseCauseStore) Create(domainId int64, cause *model.AgentPauseCause) (*model.AgentPauseCause, *model.AppError) {
	err := s.GetMaster().SelectOne(&cause, `with s as (
    insert into cc_pause_cause (domain_id, created_at, updated_at, created_by, updated_by,
                                      name, limit_per_day, allow_supervisor, allow_agent, description)
    values (:DomainId, :CreatedAt, :UpdatedAt, :CreatedBy, :UpdatedBy,
            :Name, :LimitPerDay, :AllowSupervisor, :AllowAgent, :Description)
    returning *
)
select s.id,
       s.created_at,
       cc_get_lookup(uc.id, coalesce(uc.name, uc.username)) as created_by,
       s.updated_at,
       cc_get_lookup(uc.id, coalesce(uc.name, uc.username)) as updated_by,
       s.name,
       s.description,
       s.limit_per_day,
       s.allow_agent,
       s.allow_supervisor
from s
         left join directory.wbt_user uc on uc.id = s.created_by
         left join directory.wbt_user uu on uu.id = s.updated_by`, map[string]interface{}{
		"DomainId":        domainId,
		"CreatedAt":       cause.CreatedAt,
		"UpdatedAt":       cause.UpdatedAt,
		"CreatedBy":       cause.CreatedBy.Id,
		"UpdatedBy":       cause.UpdatedBy.Id,
		"Name":            cause.Name,
		"LimitPerDay":     cause.LimitPerDay,
		"AllowSupervisor": cause.AllowSupervisor,
		"AllowAgent":      cause.AllowAgent,
		"Description":     cause.Description,
	})

	if err != nil {
		return nil, model.NewAppError("SqlPauseCauseStore.Create", "store.sql_pause_cause.create.app_error", nil,
			fmt.Sprintf("name=%v, %v", cause.Name, err.Error()), extractCodeFromErr(err))
	}

	return cause, nil
}

func (s SqlPauseCauseStore) GetAllPage(domainId int64, search *model.SearchAgentPauseCause) ([]*model.AgentPauseCause, *model.AppError) {
	var causes []*model.AgentPauseCause

	f := map[string]interface{}{
		"DomainId": domainId,
		"Q":        search.GetRegExpQ(),
		"Ids":      pq.Array(search.Ids),
		"Name":     search.Name,
	}

	err := s.ListQuery(&causes, search.ListRequest,
		`domain_id = :DomainId
				and (:Q::text isnull or description ~ :Q  or  name ~ :Q)
				and (:Ids::int4[] isnull or id = any(:Ids))
			`,
		model.AgentPauseCause{}, f)
	if err != nil {
		return nil, model.NewAppError("SqlPauseCauseStore.GetAllPage", "store.sql_pause_cause.get_all.app_error", nil, err.Error(), extractCodeFromErr(err))
	}

	return causes, nil
}

func (s SqlPauseCauseStore) Get(domainId int64, id uint32) (*model.AgentPauseCause, *model.AppError) {
	var cause *model.AgentPauseCause
	err := s.GetReplica().SelectOne(&cause, `select id, 
       created_at,
       created_by,
       updated_at,
       updated_by,
       name,
       description,
       allow_agent,
       allow_supervisor,
       limit_per_day
from cc_pause_cause_list
where id = :Id and domain_id = :DomainId`, map[string]interface{}{
		"DomainId": domainId,
		"Id":       id,
	})

	if err != nil {
		return nil, model.NewAppError("SqlPauseCauseStore.Get", "store.sql_pause_cause.get.app_error", nil, err.Error(), extractCodeFromErr(err))
	}

	return cause, nil
}

func (s SqlPauseCauseStore) Update(domainId int64, cause *model.AgentPauseCause) (*model.AgentPauseCause, *model.AppError) {
	err := s.GetMaster().SelectOne(&cause, `with s as (
    update cc_pause_cause
        set updated_at = :UpdatedAt,
            updated_by = :UpdatedBy,
            name = :Name,
            description = :Description,
            limit_per_day = :LimitPerDay,
            allow_supervisor = :AllowSupervisor,
            allow_agent = :AllowAgent
        where id = :Id and domain_id = :DomainId
    returning *
)
select s.id,
       s.created_at,
       cc_get_lookup(uc.id, coalesce(uc.name, uc.username)) as created_by,
       s.updated_at,
       cc_get_lookup(uc.id, coalesce(uc.name, uc.username)) as updated_by,
       s.name,
       s.description,
       s.limit_per_day,
       s.allow_agent,
       s.allow_supervisor
from s
         left join directory.wbt_user uc on uc.id = s.created_by
         left join directory.wbt_user uu on uu.id = s.updated_by;`, map[string]interface{}{
		"DomainId":        domainId,
		"Id":              cause.Id,
		"Name":            cause.Name,
		"Description":     cause.Description,
		"LimitPerDay":     cause.LimitPerDay,
		"AllowSupervisor": cause.AllowSupervisor,
		"AllowAgent":      cause.AllowAgent,
		"UpdatedAt":       cause.UpdatedAt,
		"UpdatedBy":       cause.UpdatedBy.Id,
	})

	if err != nil {
		return nil, model.NewAppError("SqlPauseCauseStore.Update", "store.sql_pause_cause.update.app_error", nil, err.Error(), extractCodeFromErr(err))
	}

	return cause, nil
}

func (s SqlPauseCauseStore) Delete(domainId int64, id uint32) *model.AppError {
	if _, err := s.GetMaster().Exec(`delete from cc_pause_cause c where c.id=:Id and c.domain_id = :DomainId`,
		map[string]interface{}{"Id": id, "DomainId": domainId}); err != nil {
		return model.NewAppError("SqlPauseCauseStore.Delete", "store.sql_pause_cause.delete.app_error", nil,
			fmt.Sprintf("Id=%v, %s", id, err.Error()), extractCodeFromErr(err))
	}
	return nil
}
