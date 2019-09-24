package sqlstore

import (
	"fmt"
	"github.com/webitel/engine/model"
	"github.com/webitel/engine/store"
	"net/http"
)

type SqlRoutingVariableStore struct {
	SqlStore
}

func NewSqlRoutingVariableStore(sqlStore SqlStore) store.RoutingVariableStore {
	us := &SqlRoutingVariableStore{sqlStore}
	return us
}

func (s SqlRoutingVariableStore) Create(variable *model.RoutingVariable) (*model.RoutingVariable, *model.AppError) {
	var out *model.RoutingVariable
	err := s.GetMaster().SelectOne(&out, `insert into acr_routing_variables (domain_id, key, value)
	values (:DomainId, :Key, :Value)
	returning *`, map[string]interface{}{"DomainId": variable.DomainId, "Key": variable.Key, "Value": variable.Value})

	if err != nil {
		return nil, model.NewAppError("SqlRoutingVariableStore.Save", "store.sql_acr_variable.save.app_error", nil,
			fmt.Sprintf("Key=%v, %v", variable.Key, err.Error()), extractCodeFromErr(err))
	}
	return out, nil
}

func (s SqlRoutingVariableStore) GetAllPage(domainId int64, offset, limit int) ([]*model.RoutingVariable, *model.AppError) {
	var vars []*model.RoutingVariable

	if _, err := s.GetReplica().Select(&vars,
		`select id, domain_id, key, value
from acr_routing_variables s
where s.domain_id = :DomainId
order by s.id
limit :Limit
offset :Offset`, map[string]interface{}{"DomainId": domainId, "Limit": limit, "Offset": offset}); err != nil {
		return nil, model.NewAppError("SqlRoutingVariableStore.GetAllPage", "store.sql_acr_variable.get_all.app_error", nil, err.Error(), http.StatusInternalServerError)
	} else {
		return vars, nil
	}
}

func (s SqlRoutingVariableStore) Get(domainId int64, id int64) (*model.RoutingVariable, *model.AppError) {
	var variable *model.RoutingVariable
	if err := s.GetReplica().SelectOne(&variable, `select id, domain_id, key, value
from acr_routing_variables s
where s.domain_id = :DomainId and s.id = :Id	
		`, map[string]interface{}{"Id": id, "DomainId": domainId}); err != nil {
		return nil, model.NewAppError("SqlRoutingVariableStore.Get", "store.sql_acr_variable.get.app_error", nil,
			fmt.Sprintf("Id=%v, %s", id, err.Error()), extractCodeFromErr(err))
	} else {
		return variable, nil
	}
}

func (s SqlRoutingVariableStore) Update(variable *model.RoutingVariable) (*model.RoutingVariable, *model.AppError) {
	err := s.GetMaster().SelectOne(&variable, `update acr_routing_variables
set value = :Value,
    key = :Key
where id = :Id and domain_id = :DomainId
returning *`, map[string]interface{}{
		"Value":    variable.Value,
		"Key":      variable.Key,
		"Id":       variable.Id,
		"DomainId": variable.DomainId,
	})
	if err != nil {
		return nil, model.NewAppError("SqlRoutingVariableStore.Update", "store.sql_acr_variable.update.app_error", nil,
			fmt.Sprintf("Id=%v, %s", variable.Id, err.Error()), extractCodeFromErr(err))
	}
	return variable, nil
}

func (s SqlRoutingVariableStore) Delete(domainId, id int64) *model.AppError {
	if _, err := s.GetMaster().Exec(`delete from acr_routing_variables c where c.id=:Id and c.domain_id = :DomainId`,
		map[string]interface{}{"Id": id, "DomainId": domainId}); err != nil {
		return model.NewAppError("SqlRoutingVariableStore.Delete", "store.sql_acr_variable.delete.app_error", nil,
			fmt.Sprintf("Id=%v, %s", id, err.Error()), http.StatusInternalServerError)
	}
	return nil
}
