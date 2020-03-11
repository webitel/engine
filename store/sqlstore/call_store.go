package sqlstore

import (
	"github.com/webitel/engine/model"
	"github.com/webitel/engine/store"
	"net/http"
)

type SqlCallStore struct {
	SqlStore
}

func NewSqlCallStore(sqlStore SqlStore) store.CallStore {
	us := &SqlCallStore{sqlStore}
	return us
}

func (s SqlCallStore) GetActive(domainId int64, search *model.SearchCall) ([]*model.Call, *model.AppError) {
	var out []*model.Call

	_, err := s.GetMaster().Select(&out, `
select c.id, c.app_id, c.state, c."timestamp", c.direction, c.destination, c.parent_id,
   json_build_object('type', coalesce(c.from_type, ''), 'number', coalesce(c.from_number, ''), 'id', coalesce(c.from_id, ''), 'name', coalesce(c.from_name, '')) "from",
   json_build_object('type', coalesce(c.to_type, ''), 'number', coalesce(c.to_number, ''), 'id', coalesce(c.to_id, ''), 'name', coalesce(c.to_name, '')) "to"
from cc_calls c
where c.domain_id = :Domain and c.state != 'hangup'
limit :Limit
offset :Offset`, map[string]interface{}{
		"Domain": domainId,
		"Limit":  search.GetLimit(),
		"Offset": search.GetOffset(),
	})

	if err != nil {
		return nil, model.NewAppError("SqlCallStore.GetAllPage", "store.sql_call.get_all.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	return out, nil
}

func (s SqlCallStore) Get(domainId int64, id string) (*model.Call, *model.AppError) {
	var out *model.Call

	err := s.GetMaster().SelectOne(&out, `
select c.id, c.app_id, c.state, c."timestamp", c.direction, c.destination, c.parent_id,
   json_build_object('type', coalesce(c.from_type, ''), 'number', coalesce(c.from_number, ''), 'id', coalesce(c.from_id, ''), 'name', coalesce(c.from_name, '')) "from",
   json_build_object('type', coalesce(c.to_type, ''), 'number', coalesce(c.to_number, ''), 'id', coalesce(c.to_id, ''), 'name', coalesce(c.to_name, '')) "to"
from cc_calls c
where c.domain_id = :Domain and c.id = :Id`, map[string]interface{}{
		"Domain": domainId,
		"Id":     id,
	})

	if err != nil {
		return nil, model.NewAppError("SqlCallStore.Get", "store.sql_call.get.app_error", nil, err.Error(), extractCodeFromErr(err))
	}

	return out, nil
}

func (s SqlCallStore) GetInstance(domainId int64, id string) (*model.CallInstance, *model.AppError) {
	var inst *model.CallInstance
	err := s.GetMaster().SelectOne(&inst, `select c.id, c.app_id, c.state
from cc_calls c
where c.id = :Id and c.domain_id = :Domain`, map[string]interface{}{
		"Id":     id,
		"Domain": domainId,
	})
	if err != nil {
		return nil, model.NewAppError("SqlCallStore.GetInstance", "store.sql_call.get_instance.app_error", nil, err.Error(), extractCodeFromErr(err))
	}

	return inst, nil
}

func (s SqlCallStore) GetHistory(domainId int64, search *model.SearchHistoryCall) ([]*model.HistoryCall, *model.AppError) {
	var out []*model.HistoryCall
	_, err := s.GetReplica().Select(&out, `
select c.id, c.app_id, c.direction, c.destination, c.parent_id,
   json_build_object('type', coalesce(c.from_type, ''), 'number', coalesce(c.from_number, ''), 'id', coalesce(c.from_id, ''), 'name', coalesce(c.from_name, '')) "from",
   json_build_object('type', coalesce(c.to_type, ''), 'number', coalesce(c.to_number, ''), 'id', coalesce(c.to_id, ''), 'name', coalesce(c.to_name, '')) "to",
   c.payload, c.created_at as created_at, c.answered_at, c.bridged_at, c.hangup_at, c.hold_sec, c.cause, c.sip_code
from cc_calls_history c
where c.domain_id = :Domain and c.created_at between :From::int8 and :To::int8 and (:UserId::int8 isnull or c.user_id = :UserId) 
order by c.created_at desc
limit :Limit
offset :Offset`, map[string]interface{}{
		"Domain": domainId,
		"Limit":  search.GetLimit(),
		"Offset": search.GetOffset(),
		"From":   search.CreatedAt.From,
		"To":     search.CreatedAt.To,
		"UserId": search.UserId,
	})

	if err != nil {
		return nil, model.NewAppError("SqlCallStore.GetHistory", "store.sql_call.get_history.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	return out, nil
}
