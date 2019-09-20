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
    insert into acr_routing_outbound_call (domain_id, name, description, created_at, created_by, updated_at, updated_by,
                                      scheme_id, pattern, priority, debug, disabled)
	values (:DomainId, :Name, :Description, :CreatedAt, :CreatedBy, :UpdatedAt, :UpdatedBy, :SchemeId, :Pattern, :Priority, :Debug, :Disabled)
	returning *
)
select tmp.id, tmp.domain_id, tmp.name, tmp.description, tmp.created_at, cc_get_lookup(c.id,c.name) as created_by,
       tmp.created_at,  cc_get_lookup(u.id, u.name) as updated_by, tmp.updated_at, cc_get_lookup(arst.id, arst.name) as scheme,
	   tmp.pattern, tmp.priority, debug, disabled
from tmp
    left join wbt_user c on c.id = tmp.created_by
    left join wbt_user u on u.id = tmp.updated_by
    inner join acr_routing_scheme arst on tmp.scheme_id = arst.id`,
		map[string]interface{}{
			"DomainId":    routing.DomainId,
			"Name":        routing.Name,
			"Description": routing.Description,
			"CreatedAt":   routing.CreatedAt,
			"CreatedBy":   routing.CreatedBy.Id,
			"UpdatedAt":   routing.UpdatedAt,
			"UpdatedBy":   routing.UpdatedBy.Id,
			"SchemeId":    routing.Scheme.Id,
			"Pattern":     routing.Pattern,
			"Priority":    routing.Priority,
			"Debug":       routing.Debug,
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

func (s SqlRoutingOutboundCallStore) GetAllPage(domainId int64, offset, limit int) ([]*model.RoutingOutboundCall, *model.AppError) {
	var routing []*model.RoutingOutboundCall

	if _, err := s.GetReplica().Select(&routing,
		`select tmp.id, tmp.domain_id, tmp.name, tmp.description, tmp.created_at, cc_get_lookup(c.id, c.name)::jsonb as created_by,
       tmp.created_at,  cc_get_lookup(u.id, u.name) as updated_by, tmp.pattern, tmp.priority, debug, disabled, cc_get_lookup(arst.id, arst.name) as scheme
from acr_routing_outbound_call tmp
	inner join acr_routing_scheme arst on tmp.scheme_id = arst.id
    left join wbt_user c on c.id = tmp.created_by
    left join wbt_user u on u.id = tmp.updated_by
where tmp.domain_id = :DomainId
order by tmp.id
limit :Limit
offset :Offset`, map[string]interface{}{"DomainId": domainId, "Limit": limit, "Offset": offset}); err != nil {
		return nil, model.NewAppError("SqlRoutingOutboundCallStore.GetAllPage", "store.sql_routing_out_call.get_all.app_error", nil, err.Error(), http.StatusInternalServerError)
	} else {
		return routing, nil
	}
}

func (s SqlRoutingOutboundCallStore) Get(domainId, id int64) (*model.RoutingOutboundCall, *model.AppError) {
	var routing *model.RoutingOutboundCall

	if err := s.GetReplica().SelectOne(&routing,
		`select tmp.id, tmp.domain_id, tmp.name, tmp.description, tmp.created_at, cc_get_lookup(c.id, c.name) as created_by,
       tmp.created_at,  cc_get_lookup(u.id, u.name) as updated_by, cc_get_lookup(arst.id, arst.name) as scheme, 
		tmp.pattern, tmp.priority, debug, disabled
from acr_routing_outbound_call tmp
    left join wbt_user c on c.id = tmp.created_by
    left join wbt_user u on u.id = tmp.updated_by
    inner join acr_routing_scheme arst on tmp.scheme_id = arst.id
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
    update acr_routing_outbound_call r
    set name = :Name,
        description = :Description,
        updated_at = :UpdatedAt,
        updated_by = :UpdatedBy,
        scheme_id = :SchemeId,
        pattern = :Pattern,
        priority = :Priority,
        debug = :Debug,
        disabled = :Disabled
    where r.id = :Id and r.domain_id = :Domain
    returning *
)
select tmp.id, tmp.domain_id, tmp.name, tmp.description, tmp.created_at, cc_get_lookup(c.id, c.name) as created_by,
       tmp.created_at,  cc_get_lookup(u.id, u.name) as updated_by, tmp.updated_at, cc_get_lookup(arst.id, arst.name) as scheme, 
		tmp.pattern, tmp.priority, debug, disabled
from tmp
    left join wbt_user c on c.id = tmp.created_by
    left join wbt_user u on u.id = tmp.updated_by
    inner join acr_routing_scheme arst on tmp.scheme_id = arst.id`,
		map[string]interface{}{
			"Id":          routing.Id,
			"Domain":      routing.DomainId,
			"Name":        routing.Name,
			"Description": routing.Description,
			"UpdatedAt":   routing.UpdatedAt,
			"UpdatedBy":   routing.UpdatedBy.Id,
			"SchemeId":    routing.Scheme.Id,
			"Pattern":     routing.Pattern,
			"Priority":    routing.Priority,
			"Debug":       routing.Debug,
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

func (s SqlRoutingOutboundCallStore) Delete(domainId, id int64) *model.AppError {
	if _, err := s.GetMaster().Exec(`delete from acr_routing_outbound_call c where c.id=:Id and c.domain_id = :DomainId`,
		map[string]interface{}{"Id": id, "DomainId": domainId}); err != nil {
		return model.NewAppError("SqlRoutingOutboundCallStore.Delete", "store.sql_routing_out_call.delete.app_error", nil,
			fmt.Sprintf("Id=%v, %s", id, err.Error()), http.StatusInternalServerError)
	}
	return nil
}
