package sqlstore

import (
	"context"
	"fmt"
	"github.com/lib/pq"
	"github.com/webitel/engine/model"
	"github.com/webitel/engine/store"
)

type SqlWebHookStore struct {
	SqlStore
}

func NewSqlSqlWebHookStore(sqlStore SqlStore) store.WebHookStore {
	us := &SqlWebHookStore{sqlStore}
	return us
}

func (s SqlWebHookStore) Create(ctx context.Context, domainId int64, hook *model.WebHook) (*model.WebHook, model.AppError) {
	err := s.GetMaster().SelectOne(&hook, `with h as (
    insert into flow.web_hook (name, domain_id, description, origin, schema_id, "authorization", created_by, updated_by,
                           key, enabled)
    values (:Name, :DomainId, :Description, :Origin::varchar[], :SchemaId, :Authorization, :CreatedBy, :UpdatedBy, :Key, :Enabled)
    returning *
)
select h.id,
       h.key,
       h.name,
       h.description,
       h.origin,
       h.enabled,
       h."authorization",
       flow.get_lookup(s.id, s.name)                    AS schema,
       flow.get_lookup(c.id, c.name::character varying) AS created_by,
       flow.get_lookup(u.id, u.name::character varying) AS updated_by,
       h.created_at,
       h.updated_at
from  h
         left join flow.acr_routing_scheme s on s.id = h.schema_id
         LEFT JOIN directory.wbt_user c ON c.id = h.created_by
         LEFT JOIN directory.wbt_user u ON u.id = h.updated_by`, map[string]interface{}{
		"DomainId":      domainId,
		"Name":          hook.Name,
		"Description":   hook.Description,
		"Origin":        pq.Array(hook.Origin),
		"SchemaId":      hook.Schema.GetSafeId(),
		"Authorization": hook.Authorization,
		"CreatedBy":     hook.CreatedBy.GetSafeId(),
		"UpdatedBy":     hook.UpdatedBy.GetSafeId(),
		"Key":           hook.Key,
		"Enabled":       hook.Enabled,
	})

	if err != nil {
		return nil, model.NewCustomCodeError("store.sql_hook.create.app_error", fmt.Sprintf("name=%v, %v", hook.Name, err.Error()), extractCodeFromErr(err))
	}

	return hook, nil
}

func (s SqlWebHookStore) GetAllPage(ctx context.Context, domainId int64, search *model.SearchWebHook) ([]*model.WebHook, model.AppError) {
	var list []*model.WebHook

	f := map[string]interface{}{
		"DomainId": domainId,
		"Q":        search.GetQ(),
		"Ids":      pq.Array(search.Ids),
	}

	err := s.ListQueryFromSchema(ctx, &list, "flow", search.ListRequest,
		`domain_id = :DomainId
				and (:Q::text isnull or ( name ilike :Q::varchar or description ilike :Q::varchar ))
				and (:Ids::int4[] isnull or id = any(:Ids))
			`,
		model.WebHook{}, f)
	if err != nil {
		return nil, model.NewCustomCodeError("store.sql_hook.get_all.app_error", err.Error(), extractCodeFromErr(err))
	}

	return list, nil
}

func (s SqlWebHookStore) Get(ctx context.Context, domainId int64, id int32) (*model.WebHook, model.AppError) {
	var hook *model.WebHook
	err := s.GetMaster().SelectOne(&hook, `
select h.id,
       h.key,
       h.name,
       h.description,
       h.origin,
       h.enabled,
       h."authorization",
       flow.get_lookup(s.id, s.name)                    AS schema,
       flow.get_lookup(c.id, c.name::character varying) AS created_by,
       flow.get_lookup(u.id, u.name::character varying) AS updated_by,
       h.created_at,
       h.updated_at
from  flow.web_hook h
         left join flow.acr_routing_scheme s on s.id = h.schema_id
         LEFT JOIN directory.wbt_user c ON c.id = h.created_by
         LEFT JOIN directory.wbt_user u ON u.id = h.updated_by
where h.domain_id = :DomainId and h.id = :Id`, map[string]interface{}{
		"DomainId": domainId,
		"Id":       id,
	})

	if err != nil {
		return nil, model.NewCustomCodeError("store.sql_hook.get.app_error", fmt.Sprintf("Id=%v, %v", id, err.Error()), extractCodeFromErr(err))
	}

	return hook, nil
}

func (s SqlWebHookStore) Update(ctx context.Context, domainId int64, hook *model.WebHook) (*model.WebHook, model.AppError) {
	err := s.GetMaster().SelectOne(&hook, `with h as (
    update flow.web_hook h
        set name = :Name,
            description = :Description,
            origin = :Origin,
            enabled = :Enabled,
            "authorization" = :Authorization,
            schema_id = :SchemaId,
            updated_at = :UpdatedAt,
            updated_by= :UpdatedBy
    where h.domain_id = :DomainId and h.id = :Id
    returning *
)
select h.id,
       h.key,
       h.name,
       h.description,
       h.origin,
       h.enabled,
       h."authorization",
       flow.get_lookup(s.id, s.name)                    AS schema,
       flow.get_lookup(c.id, c.name::character varying) AS created_by,
       flow.get_lookup(u.id, u.name::character varying) AS updated_by,
       h.created_at,
       h.updated_at
from  h
         left join flow.acr_routing_scheme s on s.id = h.schema_id
         LEFT JOIN directory.wbt_user c ON c.id = h.created_by
         LEFT JOIN directory.wbt_user u ON u.id = h.updated_by`, map[string]interface{}{
		"DomainId":      domainId,
		"Id":            hook.Id,
		"Name":          hook.Name,
		"Description":   hook.Description,
		"Origin":        pq.Array(hook.Origin),
		"SchemaId":      hook.Schema.GetSafeId(),
		"Authorization": hook.Authorization,
		"UpdatedBy":     hook.UpdatedBy.GetSafeId(),
		"UpdatedAt":     hook.UpdatedAt,
		"Enabled":       hook.Enabled,
	})

	if err != nil {
		return nil, model.NewCustomCodeError("store.sql_hook.update.app_error", fmt.Sprintf("name=%v, %v", hook.Name, err.Error()), extractCodeFromErr(err))
	}

	return hook, nil
}

func (s SqlWebHookStore) Delete(ctx context.Context, domainId int64, id int32) model.AppError {
	_, err := s.GetMaster().WithContext(ctx).Exec(`delete from flow.web_hook
where domain_id = :DomainId and id = :Id`, map[string]interface{}{
		"DomainId": domainId,
		"Id":       id,
	})

	if err != nil {
		return model.NewCustomCodeError("store.sql_hook.delete.app_error", fmt.Sprintf("id=%v, %v", id, err.Error()), extractCodeFromErr(err))
	}

	return nil
}
