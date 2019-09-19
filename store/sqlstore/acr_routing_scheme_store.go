package sqlstore

import (
	"database/sql"
	"fmt"
	"github.com/lib/pq"
	"github.com/webitel/engine/model"
	"github.com/webitel/engine/store"
	"net/http"
)

type SqlRoutingSchemeStore struct {
	SqlStore
}

func NewSqlRoutingSchemeStore(sqlStore SqlStore) store.RoutingSchemeStore {
	us := &SqlRoutingSchemeStore{sqlStore}
	return us
}

func (s SqlRoutingSchemeStore) Create(scheme *model.RoutingScheme) (*model.RoutingScheme, *model.AppError) {
	var out *model.RoutingScheme
	if err := s.GetMaster().SelectOne(&out, `with s as (
    insert into acr_routing_scheme (domain_id, name, scheme, payload, type, created_at, created_by, updated_at, updated_by)
    values (:DomainId, :Name, :Scheme, :Payload, :Type, :CreatedAt, :CreatedBy, :UpdatedAt, :UpdatedBy)
    returning *
)
select s.id, s.domain_id, s.name, s.created_at, cc_get_lookup(c.id, c.name) as created_by,
    s.updated_at, cc_get_lookup(u.id, u.name) as updated_by, s.scheme, s.payload
from s
    left join wbt_user c on c.id = s.created_by
    left join wbt_user u on u.id = s.updated_by`,
		map[string]interface{}{
			"DomainId":  scheme.DomainId,
			"Name":      scheme.Name,
			"Scheme":    scheme.Scheme,
			"Payload":   scheme.Payload,
			"Type":      scheme.Type,
			"CreatedAt": scheme.CreatedAt,
			"CreatedBy": scheme.CreatedBy.Id,
			"UpdatedAt": scheme.UpdatedAt,
			"UpdatedBy": scheme.UpdatedBy.Id,
		}); err != nil {
		return nil, model.NewAppError("SqlRoutingSchemeStore.Save", "store.sql_routing_scheme.save.app_error", nil,
			fmt.Sprintf("name=%v, %v", scheme.Name, err.Error()), http.StatusInternalServerError)
	} else {
		return out, nil
	}
}

func (s SqlRoutingSchemeStore) GetAllPage(domainId int64, offset, limit int) ([]*model.RoutingScheme, *model.AppError) {
	var schemes []*model.RoutingScheme

	if _, err := s.GetReplica().Select(&schemes,
		`select s.id, s.domain_id, s.name, s.created_at, cc_get_lookup(c.id, c.name) as created_by,
    s.updated_at, cc_get_lookup(u.id, u.name) as updated_by
from acr_routing_scheme s
    left join wbt_user c on c.id = s.created_by
    left join wbt_user u on u.id = s.updated_by
where s.domain_id = :DomainId
order by s.id
limit :Limit
offset :Offset`, map[string]interface{}{"DomainId": domainId, "Limit": limit, "Offset": offset}); err != nil {
		return nil, model.NewAppError("SqlRoutingSchemeStore.GetAllPage", "store.sql_routing_scheme.get_all.app_error", nil, err.Error(), http.StatusInternalServerError)
	} else {
		return schemes, nil
	}
}

func (s SqlRoutingSchemeStore) Get(domainId int64, id int64) (*model.RoutingScheme, *model.AppError) {
	var rScheme *model.RoutingScheme
	if err := s.GetReplica().SelectOne(&rScheme, `
			select s.id, s.domain_id, s.name, s.created_at, cc_get_lookup(c.id, c.name) as created_by,
		s.updated_at, cc_get_lookup(u.id, u.name) as updated_by, s.scheme, s.payload
	from acr_routing_scheme s
		left join wbt_user c on c.id = s.created_by
		left join wbt_user u on u.id = s.updated_by
	where s.id = :Id and s.domain_id = :DomainId
	order by s.id	
		`, map[string]interface{}{"Id": id, "DomainId": domainId}); err != nil {
		if err == sql.ErrNoRows {
			return nil, model.NewAppError("SqlRoutingSchemeStore.Get", "store.sql_routing_scheme.get.app_error", nil,
				fmt.Sprintf("Id=%v, %s", id, err.Error()), http.StatusNotFound)
		} else {
			return nil, model.NewAppError("SqlRoutingSchemeStore.Get", "store.sql_routing_scheme.get.app_error", nil,
				fmt.Sprintf("Id=%v, %s", id, err.Error()), http.StatusInternalServerError)
		}
	} else {
		return rScheme, nil
	}
}

func (s SqlRoutingSchemeStore) Update(scheme *model.RoutingScheme) (*model.RoutingScheme, *model.AppError) {
	err := s.GetMaster().SelectOne(&scheme, `with s as (
    update acr_routing_scheme s
    set name = :Name,
        scheme = :Scheme,
        payload = :Payload,
        type = :Type,
        updated_at = :UpdatedAt,
        updated_by = :UpdatedBy,
		description = :Description
    where s.id = :Id and s.domain_id = :Domain
    returning *
)
select s.id, s.domain_id, s.description, s.name, s.created_at, cc_get_lookup(c.id, c.name) as created_by,
    s.updated_at, cc_get_lookup(u.id, u.name) as updated_by, s.scheme, s.payload
from s
    left join wbt_user c on c.id = s.created_by
    left join wbt_user u on u.id = s.updated_by`, map[string]interface{}{
		"Name":        scheme.Name,
		"Scheme":      scheme.Scheme,
		"Payload":     scheme.Payload,
		"Type":        scheme.Type,
		"UpdatedAt":   scheme.UpdatedAt,
		"UpdatedBy":   scheme.UpdatedBy.Id,
		"Id":          scheme.Id,
		"Domain":      scheme.DomainId,
		"Description": scheme.Description,
	})
	if err != nil {
		code := http.StatusInternalServerError
		switch err.(type) {
		case *pq.Error:
			if err.(*pq.Error).Code == ForeignKeyViolationErrorCode {
				code = http.StatusBadRequest
			}
		}
		return nil, model.NewAppError("SqlRoutingSchemeStore.Update", "store.sql_routing_scheme.update.app_error", nil,
			fmt.Sprintf("Id=%v, %s", scheme.Id, err.Error()), code)
	}
	return scheme, nil
}

func (s SqlRoutingSchemeStore) Delete(domainId, id int64) *model.AppError {
	if _, err := s.GetMaster().Exec(`delete from acr_routing_scheme c where c.id=:Id and c.domain_id = :DomainId`,
		map[string]interface{}{"Id": id, "DomainId": domainId}); err != nil {
		return model.NewAppError("SqlRoutingSchemeStore.Delete", "store.sql_routing_scheme.delete.app_error", nil,
			fmt.Sprintf("Id=%v, %s", id, err.Error()), http.StatusInternalServerError)
	}
	return nil
}
