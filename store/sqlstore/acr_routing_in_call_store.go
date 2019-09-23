package sqlstore

import (
	"database/sql"
	"fmt"
	"github.com/lib/pq"
	"github.com/webitel/engine/model"
	"github.com/webitel/engine/store"
	"net/http"
)

type SqlRoutingInboundCallStore struct {
	SqlStore
}

func NewSqlRoutingInboundCallStore(sqlStore SqlStore) store.RoutingInboundCallStore {
	us := &SqlRoutingInboundCallStore{sqlStore}
	return us
}

func (s SqlRoutingInboundCallStore) Create(routing *model.RoutingInboundCall) (*model.RoutingInboundCall, *model.AppError) {
	var out *model.RoutingInboundCall
	err := s.GetMaster().SelectOne(&out, `with tmp as (
    insert into acr_routing_inbound_call (domain_id, name, description, created_at, created_by, updated_at, updated_by,
                                      start_scheme_id, stop_scheme_id, numbers, host, timezone_id, debug, disabled)
	values (:DomainId, :Name, :Description, :CreatedAt, :CreatedBy, :UpdatedAt, :UpdatedBy, :StartSchemeId, :StopSchemeId,
        :Numbers, :Host, :TimezoneId, :Debug, :Disabled)
	returning *
)
select tmp.id, tmp.domain_id, tmp.name, tmp.description, tmp.created_at, cc_get_lookup(c.id,c.name) as created_by,
       tmp.created_at,  cc_get_lookup(u.id, u.name) as updated_by, tmp.updated_at, cc_get_lookup(arst.id, arst.name) as start_scheme,
      cc_get_lookup(arsp.id, arsp.name) as stop_scheme, tmp.numbers, tmp.host, cc_get_lookup(ct.id, ct.name) as timezone,
       debug, disabled
from tmp
    left join directory.wbt_user c on c.id = tmp.created_by
    left join directory.wbt_user u on u.id = tmp.updated_by
    inner join acr_routing_scheme arst on tmp.start_scheme_id = arst.id
    left join acr_routing_scheme arsp on tmp.stop_scheme_id = arsp.id
    inner join calendar_timezones ct on tmp.timezone_id = ct.id`,
		map[string]interface{}{
			"DomainId":      routing.DomainId,
			"Name":          routing.Name,
			"Description":   routing.Description,
			"CreatedAt":     routing.CreatedAt,
			"CreatedBy":     routing.CreatedBy.Id,
			"UpdatedAt":     routing.UpdatedAt,
			"UpdatedBy":     routing.UpdatedBy.Id,
			"StartSchemeId": routing.StartScheme.Id,
			"StopSchemeId":  routing.GetStopSchemeId(),
			"Numbers":       pq.StringArray(routing.Numbers),
			"Host":          routing.Host,
			"TimezoneId":    routing.Timezone.Id,
			"Debug":         routing.Debug,
			"Disabled":      routing.Disabled,
		})

	if err != nil {
		code := http.StatusInternalServerError
		switch err.(type) {
		case *pq.Error:
			if err.(*pq.Error).Code == ForeignKeyViolationErrorCode {
				code = http.StatusBadRequest
			}
		}
		return nil, model.NewAppError("SqlRoutingInboundCallStore.Create", "store.sql_routing_in_call.create.app_error", nil,
			fmt.Sprintf("Id=%v, %s", routing.Id, err.Error()), code)
	}
	return out, nil
}

func (s SqlRoutingInboundCallStore) GetAllPage(domainId int64, offset, limit int) ([]*model.RoutingInboundCall, *model.AppError) {
	var routing []*model.RoutingInboundCall

	if _, err := s.GetReplica().Select(&routing,
		`select tmp.id, tmp.domain_id, tmp.name, tmp.description, tmp.created_at, cc_get_lookup(c.id, c.name)::jsonb as created_by,
       tmp.created_at,  cc_get_lookup(u.id, u.name) as updated_by, tmp.numbers, tmp.host, cc_get_lookup(ct.id, ct.name) as timezone,
       debug, disabled
from acr_routing_inbound_call tmp
    left join directory.wbt_user c on c.id = tmp.created_by
    left join directory.wbt_user u on u.id = tmp.updated_by
inner join calendar_timezones ct on tmp.timezone_id = ct.id 
where tmp.domain_id = :DomainId
order by tmp.id
limit :Limit
offset :Offset`, map[string]interface{}{"DomainId": domainId, "Limit": limit, "Offset": offset}); err != nil {
		return nil, model.NewAppError("SqlRoutingInboundCallStore.GetAllPage", "store.sql_routing_in_call.get_all.app_error", nil, err.Error(), http.StatusInternalServerError)
	} else {
		return routing, nil
	}
}

func (s SqlRoutingInboundCallStore) Get(domainId, id int64) (*model.RoutingInboundCall, *model.AppError) {
	var routing *model.RoutingInboundCall

	if err := s.GetReplica().SelectOne(&routing,
		`select tmp.id, tmp.domain_id, tmp.name, tmp.description, tmp.created_at, cc_get_lookup(c.id, c.name) as created_by,
       tmp.created_at,  cc_get_lookup(u.id, u.name) as updated_by, cc_get_lookup(arst.id, arst.name) as start_scheme,
      cc_get_lookup(arsp.id, arsp.name) as stop_scheme, tmp.numbers, tmp.host, cc_get_lookup(ct.id, ct.name) as timezone,
       debug, disabled
from acr_routing_inbound_call tmp
    left join wbt_user c on c.id = tmp.created_by
    left join wbt_user u on u.id = tmp.updated_by
    inner join acr_routing_scheme arst on tmp.start_scheme_id = arst.id
    left join acr_routing_scheme arsp on tmp.stop_scheme_id = arsp.id
inner join calendar_timezones ct on tmp.timezone_id = ct.id 
where tmp.id = :Id and tmp.domain_id = :DomainId`, map[string]interface{}{"DomainId": domainId, "Id": id}); err != nil {
		if err == sql.ErrNoRows {
			return nil, model.NewAppError("SqlRoutingInboundCallStore.Get", "store.sql_routing_in_call.get.app_error", nil, err.Error(), http.StatusNotFound)
		} else {
			return nil, model.NewAppError("SqlRoutingInboundCallStore.Get", "store.sql_routing_in_call.get.app_error", nil, err.Error(), http.StatusInternalServerError)
		}
	} else {
		return routing, nil
	}
}

func (s SqlRoutingInboundCallStore) Update(routing *model.RoutingInboundCall) (*model.RoutingInboundCall, *model.AppError) {
	var out *model.RoutingInboundCall
	err := s.GetMaster().SelectOne(&out, `with tmp as (
    update acr_routing_inbound_call r
    set name = :Name,
        description = :Description,
        updated_at = :UpdatedAt,
        updated_by = :UpdatedBy,
        start_scheme_id = :StartSchemeId,
        stop_scheme_id = :StopSchemeId,
        numbers = :Numbers,
        host = :Host,
        timezone_id = :TimezoneId,
        debug = :Debug,
        disabled = :Disabled
    where r.id = :Id and r.domain_id = :Domain
    returning *
)
select tmp.id, tmp.domain_id, tmp.name, tmp.description, tmp.created_at, cc_get_lookup(c.id, c.name) as created_by,
       tmp.created_at,  cc_get_lookup(u.id, u.name) as updated_by, cc_get_lookup(arst.id, arst.name) as start_scheme,
       cc_get_lookup(arsp.id, arsp.name) as stop_scheme, tmp.numbers, tmp.host, cc_get_lookup(ct.id, ct.name) as timezone,
       debug, disabled
from tmp
    left join wbt_user c on c.id = tmp.created_by
    left join wbt_user u on u.id = tmp.updated_by
    inner join acr_routing_scheme arst on tmp.start_scheme_id = arst.id
    left join acr_routing_scheme arsp on tmp.stop_scheme_id = arsp.id
    inner join calendar_timezones ct on tmp.timezone_id = ct.id`,
		map[string]interface{}{
			"Id":            routing.Id,
			"Domain":        routing.DomainId,
			"Name":          routing.Name,
			"Description":   routing.Description,
			"UpdatedAt":     routing.UpdatedAt,
			"UpdatedBy":     routing.UpdatedBy.Id,
			"StartSchemeId": routing.StartScheme.Id,
			"StopSchemeId":  routing.GetStopSchemeId(),
			"Numbers":       pq.Array(routing.Numbers),
			"Host":          routing.Host,
			"TimezoneId":    routing.Timezone.Id,
			"Debug":         routing.Debug,
			"Disabled":      routing.Disabled,
		})

	if err != nil {
		code := http.StatusInternalServerError
		switch err.(type) {
		case *pq.Error:
			if err.(*pq.Error).Code == ForeignKeyViolationErrorCode {
				code = http.StatusBadRequest
			}
		}
		return nil, model.NewAppError("SqlRoutingInboundCallStore.Update", "store.sql_routing_in_call.update.app_error", nil,
			fmt.Sprintf("Id=%v, %s", routing.Id, err.Error()), code)
	}
	return out, nil
}

func (s SqlRoutingInboundCallStore) Delete(domainId, id int64) *model.AppError {
	if _, err := s.GetMaster().Exec(`delete from acr_routing_inbound_call c where c.id=:Id and c.domain_id = :DomainId`,
		map[string]interface{}{"Id": id, "DomainId": domainId}); err != nil {
		return model.NewAppError("SqlRoutingInboundCallStore.Delete", "store.sql_routing_in_call.delete.app_error", nil,
			fmt.Sprintf("Id=%v, %s", id, err.Error()), http.StatusInternalServerError)
	}
	return nil
}
