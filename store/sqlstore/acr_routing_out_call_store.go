package sqlstore

import (
	"database/sql"
	"fmt"
	"github.com/lib/pq"
	"github.com/webitel/engine/model"
	"github.com/webitel/engine/store"
	"net/http"
)

type SqlRoutingOutboundCallStore struct {
	SqlStore
}

func NewSqlRoutingOutboundCallStore(sqlStore SqlStore) store.RoutingOutboundCallStore {
	us := &SqlRoutingOutboundCallStore{sqlStore}
	return us
}

func (s SqlRoutingOutboundCallStore) Create(routing *model.RoutingOutboundCall) (*model.RoutingOutboundCall, *model.AppError) {
	var out *model.RoutingOutboundCall
	err := s.GetMaster().SelectOne(&out, `with tmp as (
    insert into flow.acr_routing_outbound_call (domain_id, name, description, created_at, created_by, updated_at, updated_by,
                                      scheme_id, pattern, disabled)
	values (:DomainId, :Name, :Description, :CreatedAt, :CreatedBy, :UpdatedAt, :UpdatedBy, :SchemeId, :Pattern, :Disabled)
	returning *
)
select tmp.id, tmp.domain_id, tmp.name, tmp.description, tmp.created_at, call_center.cc_get_lookup(c.id,c.name) as created_by,
       tmp.created_at,  call_center.cc_get_lookup(u.id, u.name) as updated_by, tmp.updated_at, call_center.cc_get_lookup(arst.id, arst.name) as schema,
	   tmp.pattern, disabled
from tmp
    left join directory.wbt_user c on c.id = tmp.created_by
    left join directory.wbt_user u on u.id = tmp.updated_by
    inner join flow.acr_routing_scheme arst on tmp.scheme_id = arst.id`,
		map[string]interface{}{
			"DomainId":    routing.DomainId,
			"Name":        routing.Name,
			"Description": routing.Description,
			"CreatedAt":   routing.CreatedAt,
			"CreatedBy":   routing.CreatedBy.GetSafeId(),
			"UpdatedAt":   routing.UpdatedAt,
			"UpdatedBy":   routing.UpdatedBy.GetSafeId(),
			"SchemeId":    routing.Schema.Id,
			"Pattern":     routing.Pattern,
			"Disabled":    routing.Disabled,
		})

	if err != nil {
		code := http.StatusInternalServerError
		switch err.(type) {
		case *pq.Error:
			if err.(*pq.Error).Code == ForeignKeyViolationErrorCode {
				code = http.StatusBadRequest
			}
		}
		return nil, model.NewAppError("SqlRoutingOutboundCallStore.Create", "store.sql_routing_out_call.create.app_error", nil,
			fmt.Sprintf("Id=%v, %s", routing.Id, err.Error()), code)
	}
	return out, nil
}

func (s SqlRoutingOutboundCallStore) GetAllPage(domainId int64, search *model.SearchRoutingOutboundCall) ([]*model.RoutingOutboundCall, *model.AppError) {
	var routing []*model.RoutingOutboundCall

	f := map[string]interface{}{
		"DomainId":    domainId,
		"Q":           search.GetQ(),
		"Ids":         pq.Array(search.Ids),
		"Name":        search.Name,
		"Description": search.Description,
		"SchemaIds":   pq.Array(search.SchemaIds),
		"Pattern":     search.Pattern,
	}

	err := s.ListQueryFromSchema(&routing, "flow", search.ListRequest,
		`domain_id = :DomainId
				and (:Q::text isnull or ( name ilike :Q::varchar or description ilike :Q::varchar ))
				and (:Ids::int4[] isnull or id = any(:Ids))
				and (:SchemaIds::int4[] isnull or schema_id = any(:SchemaIds))
				and (:Name::text isnull or name = :Name)
				and (:Description::text isnull or description = :Description)
				and (:Pattern::text isnull or pattern = :Pattern)
			`,
		model.RoutingOutboundCall{}, f)
	if err != nil {
		return nil, model.NewAppError("SqlRoutingOutboundCallStore.GetAllPage", "store.sql_routing_out_call.get_all.app_error", nil, err.Error(), extractCodeFromErr(err))
	} else {
		return routing, nil
	}
}

func (s SqlRoutingOutboundCallStore) Get(domainId, id int64) (*model.RoutingOutboundCall, *model.AppError) {
	var routing *model.RoutingOutboundCall

	if err := s.GetReplica().SelectOne(&routing,
		`select tmp.id, tmp.domain_id, tmp.name, tmp.description, tmp.created_at, call_center.cc_get_lookup(c.id, c.name) as created_by,
       tmp.created_at,  call_center.cc_get_lookup(u.id, u.name) as updated_by, call_center.cc_get_lookup(arst.id, arst.name) as schema, 
		tmp.pattern, disabled
from flow.acr_routing_outbound_call tmp
    left join directory.wbt_user c on c.id = tmp.created_by
    left join directory.wbt_user u on u.id = tmp.updated_by
    inner join flow.acr_routing_scheme arst on tmp.scheme_id = arst.id
where tmp.id = :Id and tmp.domain_id = :DomainId`, map[string]interface{}{"DomainId": domainId, "Id": id}); err != nil {
		if err == sql.ErrNoRows {
			return nil, model.NewAppError("SqlRoutingOutboundCallStore.Get", "store.sql_routing_out_call.get.app_error", nil, err.Error(), http.StatusNotFound)
		} else {
			return nil, model.NewAppError("SqlRoutingOutboundCallStore.Get", "store.sql_routing_out_call.get.app_error", nil, err.Error(), http.StatusInternalServerError)
		}
	} else {
		return routing, nil
	}
}

func (s SqlRoutingOutboundCallStore) Update(routing *model.RoutingOutboundCall) (*model.RoutingOutboundCall, *model.AppError) {
	var out *model.RoutingOutboundCall
	err := s.GetMaster().SelectOne(&out, `with tmp as (
    update flow.acr_routing_outbound_call r
    set name = :Name,
        description = :Description,
        updated_at = :UpdatedAt,
        updated_by = :UpdatedBy,
        scheme_id = :SchemeId,
        pattern = :Pattern,
        disabled = :Disabled
    where r.id = :Id and r.domain_id = :Domain
    returning *
)
select tmp.id, tmp.domain_id, tmp.name, tmp.description, tmp.created_at, call_center.cc_get_lookup(c.id, c.name) as created_by,
       tmp.created_at,  call_center.cc_get_lookup(u.id, u.name) as updated_by, tmp.updated_at, call_center.cc_get_lookup(arst.id, arst.name) as schema, 
		tmp.pattern, disabled
from tmp
    left join directory.wbt_user c on c.id = tmp.created_by
    left join directory.wbt_user u on u.id = tmp.updated_by
    inner join flow.acr_routing_scheme arst on tmp.scheme_id = arst.id`,
		map[string]interface{}{
			"Id":          routing.Id,
			"Domain":      routing.DomainId,
			"Name":        routing.Name,
			"Description": routing.Description,
			"UpdatedAt":   routing.UpdatedAt,
			"UpdatedBy":   routing.UpdatedBy.GetSafeId(),
			"SchemeId":    routing.Schema.Id,
			"Pattern":     routing.Pattern,
			"Disabled":    routing.Disabled,
		})

	if err != nil {
		code := http.StatusInternalServerError
		switch err.(type) {
		case *pq.Error:
			if err.(*pq.Error).Code == ForeignKeyViolationErrorCode {
				code = http.StatusBadRequest
			}
		}
		return nil, model.NewAppError("SqlRoutingOutboundCallStore.Update", "store.sql_routing_out_call.update.app_error", nil,
			fmt.Sprintf("Id=%v, %s", routing.Id, err.Error()), code)
	}
	return out, nil
}

func (s SqlRoutingOutboundCallStore) ChangePosition(domainId, fromId, toId int64) *model.AppError {
	i, err := s.GetMaster().SelectInt(`with t as (
		select f.id,
           case when f.pos > lead(f.pos) over () then lead(f.pos) over () else lag(f.pos) over (order by f.pos desc) end as new_pos,
           count(*) over () cnt
        from flow.acr_routing_outbound_call f
        where f.id in (:FromId, :ToId) and f.domain_id = :DomainId
        order by f.pos desc
	),
	u as (
		update flow.acr_routing_outbound_call u
		set pos = t.new_pos
		from t
		where t.id = u.id and t.cnt = 2 and  :FromId != :ToId
		returning u.id
	)
	select count(*)
	from u`, map[string]interface{}{
		"FromId":   fromId,
		"ToId":     toId,
		"DomainId": domainId,
	})

	if err != nil {
		return model.NewAppError("SqlRoutingOutboundCallStore.ChangePosition", "store.sql_routing_out_call.change_position.app_error", nil,
			fmt.Sprintf("FromId=%v, ToId=%v %s", fromId, toId, err.Error()), extractCodeFromErr(err))
	}

	if i == 0 {
		return model.NewAppError("SqlRoutingOutboundCallStore.ChangePosition", "store.sql_routing_out_call.change_position.not_found", nil,
			fmt.Sprintf("FromId=%v, ToId=%v", fromId, toId), http.StatusNotFound)
	}

	return nil
}

func (s SqlRoutingOutboundCallStore) Delete(domainId, id int64) *model.AppError {
	if _, err := s.GetMaster().Exec(`delete from flow.acr_routing_outbound_call c where c.id=:Id and c.domain_id = :DomainId`,
		map[string]interface{}{"Id": id, "DomainId": domainId}); err != nil {
		return model.NewAppError("SqlRoutingOutboundCallStore.Delete", "store.sql_routing_out_call.delete.app_error", nil,
			fmt.Sprintf("Id=%v, %s", id, err.Error()), http.StatusInternalServerError)
	}
	return nil
}
