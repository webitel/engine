package sqlstore

import (
	"context"
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

func (s SqlPauseCauseStore) Create(ctx context.Context, domainId int64, cause *model.PauseCause) (*model.PauseCause, model.AppError) {
	err := s.GetMaster().WithContext(ctx).SelectOne(&cause, `with s as (
    insert into call_center.cc_pause_cause (domain_id, created_at, updated_at, created_by, updated_by,
                                      name, limit_min, allow_supervisor, allow_agent, allow_admin, description)
    values (:DomainId, :CreatedAt, :UpdatedAt, :CreatedBy, :UpdatedBy,
            :Name, :LimitMin, :AllowSupervisor, :AllowAgent, :AllowAdmin, :Description)
    returning *
)
select s.id,
       s.created_at,
       call_center.cc_get_lookup(uc.id, coalesce(uc.name, uc.username)) as created_by,
       s.updated_at,
       call_center.cc_get_lookup(uc.id, coalesce(uc.name, uc.username)) as updated_by,
       s.name,
       s.description,
       s.limit_min,
       s.allow_agent,
       s.allow_supervisor,
	   s.allow_admin
from s
         left join directory.wbt_user uc on uc.id = s.created_by
         left join directory.wbt_user uu on uu.id = s.updated_by`, map[string]interface{}{
		"DomainId":        domainId,
		"CreatedAt":       cause.CreatedAt,
		"UpdatedAt":       cause.UpdatedAt,
		"CreatedBy":       cause.CreatedBy.GetSafeId(),
		"UpdatedBy":       cause.UpdatedBy.GetSafeId(),
		"Name":            cause.Name,
		"LimitMin":        cause.LimitMin,
		"AllowSupervisor": cause.AllowSupervisor,
		"AllowAgent":      cause.AllowAgent,
		"AllowAdmin":      cause.AllowAdmin,
		"Description":     cause.Description,
	})

	if err != nil {
		return nil, model.NewCustomCodeError("store.sql_pause_cause.create.app_error", fmt.Sprintf("name=%v, %v", cause.Name, err.Error()), extractCodeFromErr(err))
	}

	return cause, nil
}

func (s SqlPauseCauseStore) GetAllPage(ctx context.Context, domainId int64, search *model.SearchPauseCause) ([]*model.PauseCause, model.AppError) {
	var causes []*model.PauseCause

	f := map[string]interface{}{
		"DomainId": domainId,
		"Q":        search.GetQ(),
		"Ids":      pq.Array(search.Ids),
		"Name":     search.Name,
	}

	err := s.ListQuery(ctx, &causes, search.ListRequest,
		`domain_id = :DomainId
				and (:Q::varchar isnull or (name ilike :Q::varchar or description ilike :Q::varchar))
				and (:Ids::int4[] isnull or id = any(:Ids))
			`,
		model.PauseCause{}, f)
	if err != nil {
		return nil, model.NewCustomCodeError("store.sql_pause_cause.get_all.app_error", err.Error(), extractCodeFromErr(err))
	}

	return causes, nil
}

func (s SqlPauseCauseStore) Get(ctx context.Context, domainId int64, id uint32) (*model.PauseCause, model.AppError) {
	var cause *model.PauseCause
	err := s.GetReplica().WithContext(ctx).SelectOne(&cause, `select id, 
       created_at,
       created_by,
       updated_at,
       updated_by,
       name,
       description,
       allow_agent,
       allow_supervisor,
	   allow_admin,
       limit_min
from call_center.cc_pause_cause_list
where id = :Id and domain_id = :DomainId`, map[string]interface{}{
		"DomainId": domainId,
		"Id":       id,
	})

	if err != nil {
		return nil, model.NewCustomCodeError("store.sql_pause_cause.get.app_error", err.Error(), extractCodeFromErr(err))
	}

	return cause, nil
}

func (s SqlPauseCauseStore) Update(ctx context.Context, domainId int64, cause *model.PauseCause) (*model.PauseCause, model.AppError) {
	err := s.GetMaster().WithContext(ctx).SelectOne(&cause, `with s as (
    update call_center.cc_pause_cause
        set updated_at = :UpdatedAt,
            updated_by = :UpdatedBy,
            name = :Name,
            description = :Description,
            limit_min = :LimitMin,
            allow_supervisor = :AllowSupervisor,
            allow_agent = :AllowAgent,
			allow_admin = :AllowAdmin
        where id = :Id and domain_id = :DomainId
    returning *
)
select s.id,
       s.created_at,
       call_center.cc_get_lookup(uc.id, coalesce(uc.name, uc.username)) as created_by,
       s.updated_at,
       call_center.cc_get_lookup(uc.id, coalesce(uc.name, uc.username)) as updated_by,
       s.name,
       s.description,
       s.limit_min,
       s.allow_agent,
       s.allow_supervisor,
	   s.allow_admin	
from s
         left join directory.wbt_user uc on uc.id = s.created_by
         left join directory.wbt_user uu on uu.id = s.updated_by;`, map[string]interface{}{
		"DomainId":        domainId,
		"Id":              cause.Id,
		"Name":            cause.Name,
		"Description":     cause.Description,
		"LimitMin":        cause.LimitMin,
		"AllowSupervisor": cause.AllowSupervisor,
		"AllowAgent":      cause.AllowAgent,
		"AllowAdmin":      cause.AllowAdmin,
		"UpdatedAt":       cause.UpdatedAt,
		"UpdatedBy":       cause.UpdatedBy.GetSafeId(),
	})

	if err != nil {
		return nil, model.NewCustomCodeError("store.sql_pause_cause.update.app_error", err.Error(), extractCodeFromErr(err))
	}

	return cause, nil
}

func (s SqlPauseCauseStore) Delete(ctx context.Context, domainId int64, id uint32) model.AppError {
	if _, err := s.GetMaster().WithContext(ctx).Exec(`delete from call_center.cc_pause_cause c where c.id=:Id and c.domain_id = :DomainId`,
		map[string]interface{}{"Id": id, "DomainId": domainId}); err != nil {
		return model.NewCustomCodeError("store.sql_pause_cause.delete.app_error", fmt.Sprintf("Id=%v, %s", id, err.Error()), extractCodeFromErr(err))
	}
	return nil
}
