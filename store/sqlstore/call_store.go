package sqlstore

import (
	"github.com/lib/pq"
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

	f := map[string]interface{}{
		"Domain":     domainId,
		"Limit":      search.GetLimit(),
		"Offset":     search.GetOffset(),
		"From":       search.CreatedAt.From,
		"To":         search.CreatedAt.To,
		"UserIds":    pq.Array(search.UserIds),
		"QueueIds":   pq.Array(search.QueueIds),
		"TeamIds":    pq.Array(search.TeamIds),
		"AgentIds":   pq.Array(search.AgentIds),
		"MemberIds":  pq.Array(search.MemberIds),
		"GatewayIds": pq.Array(search.GatewayIds),
		"SkipParent": search.SkipParent,
		"ParentId":   search.ParentId,
		"Number":     search.Number,
		"Cause":      search.Cause,
	}

	err := s.ListQuery(&out, search.ListRequest,
		`domain_id = :Domain and created_at between :From::int8 and :To::int8 
	and (:UserIds::int8[] isnull or user_id = any(:UserIds))
	and (:QueueIds::int[] isnull or queue_id = any(:QueueIds) )
	and (:TeamIds::int[] isnull or team_id = any(:TeamIds) )  
	and (:AgentIds::int[] isnull or agent_id = any(:AgentIds) )
	and (:MemberIds::int8[] isnull or member_id = any(:MemberIds) )
	and (:GatewayIds::int8[] isnull or gateway_id = any(:GatewayIds) )
	and (:Number::varchar isnull or from_number ilike :Number::varchar or to_number ilike :Number::varchar)
	and ( (:SkipParent::bool isnull or not :SkipParent::bool is true ) or parent_id isnull)
	and (:ParentId::varchar isnull or parent_id = :ParentId )
	and (:Cause::varchar isnull or cause = :Cause )`,
		model.HistoryCall{}, f)
	if err != nil {
		return nil, model.NewAppError("SqlCallStore.GetHistory", "store.sql_call.get_history.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	return out, nil
}

func (s SqlCallStore) BridgeInfo(domainId int64, fromId, toId string) (*model.BridgeCall, *model.AppError) {
	var res *model.BridgeCall
	err := s.GetMaster().SelectOne(&res, `select coalesce(c.parent_id, c.id) from_id, coalesce(c2.parent_id, c2.id) to_id, c.app_id
from cc_calls c,
     cc_calls c2
where c.id = :FromId and c2.id = :ToId and c.domain_id = :DomainId and c2.domain_id = :DomainId`, map[string]interface{}{
		"DomainId": domainId,
		"FromId":   fromId,
		"ToId":     toId,
	})
	if err != nil {
		return nil, model.NewAppError("SqlCallStore.GetBridgeInfo", "store.sql_call.get_bridge_info.app_error", nil, err.Error(), extractCodeFromErr(err))
	} else {
		return res, nil
	}
}
