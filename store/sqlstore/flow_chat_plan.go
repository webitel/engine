package sqlstore

import (
	"context"
	"fmt"

	"github.com/lib/pq"
	"github.com/webitel/engine/model"
	"github.com/webitel/engine/store"
)

type SqlChatPlanStore struct {
	SqlStore
}

func NewSqlChatPlanStore(sqlStore SqlStore) store.ChatPlanStore {
	us := &SqlChatPlanStore{sqlStore}
	return us
}

func (s SqlChatPlanStore) Create(ctx context.Context, domainId int64, plan *model.ChatPlan) (*model.ChatPlan, model.AppError) {
	err := s.GetMaster().WithContext(ctx).SelectOne(&plan, `with c as (
    insert into flow.acr_chat_plan (domain_id, enabled, name, schema_id, description)
    values (:DomainId::int8, :Enabled::bool, :Name::varchar, :SchemaId::int4, :Description::text)
    returning *
)
select c.id,
    c.enabled,
    c.name,
    flow.get_lookup(s.id, s.name) as schema,
    c.description,
    c.domain_id
from c
    left join flow.acr_routing_scheme s on s.id = c.schema_id`, map[string]interface{}{
		"DomainId":    domainId,
		"Enabled":     plan.Enabled,
		"Name":        plan.Name,
		"SchemaId":    plan.Schema.Id,
		"Description": plan.Description,
	})

	if err != nil {
		return nil, model.NewCustomCodeError("store.sql_chat_plan.create.app_error", fmt.Sprintf("Name=%v, %s", plan.Name, err.Error()), extractCodeFromErr(err))
	}

	return plan, nil
}

func (s SqlChatPlanStore) GetAllPage(ctx context.Context, domainId int64, search *model.SearchChatPlan) ([]*model.ChatPlan, model.AppError) {
	var plans []*model.ChatPlan

	f := map[string]interface{}{
		"DomainId": domainId,
		"Q":        search.GetQ(),
		"Ids":      pq.Array(search.Ids),
		"Name":     search.Name,
		"Enabled":  search.Enabled,
	}

	err := s.ListQueryFromSchema(ctx, &plans, "flow", search.ListRequest,
		`domain_id = :DomainId
				and (:Q::text isnull or ( name ilike :Q::varchar or description ilike :Q::varchar ))
				and (:Ids::int4[] isnull or id = any(:Ids))
				and (:Name::text isnull or name = :Name)
				and (:Enabled::bool isnull or enabled)
			`,
		model.ChatPlan{}, f)
	if err != nil {
		return nil, model.NewCustomCodeError("store.sql_chat_plan.get_all.app_error", err.Error(), extractCodeFromErr(err))
	} else {
		return plans, nil
	}
}

func (s SqlChatPlanStore) Get(ctx context.Context, domainId int64, id int32) (*model.ChatPlan, model.AppError) {
	var plan *model.ChatPlan

	err := s.GetMaster().WithContext(ctx).SelectOne(&plan, `
select c.id,
    c.enabled,
    c.name,
    flow.get_lookup(s.id, s.name) as schema,
    c.description,
    c.domain_id
from flow.acr_chat_plan c
    left join flow.acr_routing_scheme s on s.id = c.schema_id
where c.id = :Id and c.domain_id = :DomainId`, map[string]interface{}{
		"DomainId": domainId,
		"Id":       id,
	})

	if err != nil {
		return nil, model.NewCustomCodeError("store.sql_chat_plan.get.app_error", fmt.Sprintf("Id=%v, %s", id, err.Error()), extractCodeFromErr(err))
	}

	return plan, nil
}

func (s SqlChatPlanStore) Update(ctx context.Context, domainId int64, plan *model.ChatPlan) (*model.ChatPlan, model.AppError) {
	err := s.GetMaster().WithContext(ctx).SelectOne(&plan, `with c as (
    update flow.acr_chat_plan
        set name = :Name,
            enabled = :Enabled,
            schema_id = :SchemaId,
            description = :Description
    where domain_id = :DomainId and id = :Id
    returning *
)
select c.id,
    c.enabled,
    c.name,
    flow.get_lookup(s.id, s.name) as schema,
    c.description,
    c.domain_id
from c
    left join flow.acr_routing_scheme s on s.id = c.schema_id`, map[string]interface{}{
		"DomainId":    domainId,
		"Id":          plan.Id,
		"Enabled":     plan.Enabled,
		"Name":        plan.Name,
		"SchemaId":    plan.Schema.Id,
		"Description": plan.Description,
	})

	if err != nil {
		return nil, model.NewCustomCodeError("store.sql_chat_plan.update.app_error", fmt.Sprintf("Id=%v, %s", plan.Id, err.Error()), extractCodeFromErr(err))
	}

	return plan, nil
}

func (s SqlChatPlanStore) Delete(ctx context.Context, domainId int64, id int32) model.AppError {
	if _, err := s.GetMaster().WithContext(ctx).Exec(`delete from flow.acr_chat_plan c where c.id=:Id and c.domain_id = :DomainId`,
		map[string]interface{}{"Id": id, "DomainId": domainId}); err != nil {
		return model.NewCustomCodeError("store.sql_chat_plan.delete.app_error", fmt.Sprintf("Id=%v, %s", id, err.Error()), extractCodeFromErr(err))
	}
	return nil
}

func (s SqlChatPlanStore) GetSchemaId(ctx context.Context, domainId int64, id int32) (int, model.AppError) {
	schemaId, err := s.GetReplica().WithContext(ctx).SelectInt(`select p.schema_id
from flow.acr_chat_plan p
where p.domain_id = :DomainId
    and p.enabled
    and p.id = :Id`, map[string]interface{}{
		"DomainId": domainId,
		"Id":       id,
	})

	if err != nil {
		return 0, model.NewCustomCodeError("store.sql_chat_plan.get.schema_id.app_error", fmt.Sprintf("Id=%v, %s", id, err.Error()), extractCodeFromErr(err))
	}

	return int(schemaId), nil
}
