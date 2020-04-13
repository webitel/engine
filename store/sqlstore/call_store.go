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
	_, err := s.GetReplica().Select(&out, `
select c.id, c.app_id, c.direction, c.destination, c.parent_id,
   json_build_object('type', coalesce(c.from_type, ''), 'number', coalesce(c.from_number, ''), 'id', coalesce(c.from_id, ''), 'name', coalesce(c.from_name, '')) "from",
   json_build_object('type', coalesce(c.to_type, ''), 'number', coalesce(c.to_number, ''), 'id', coalesce(c.to_id, ''), 'name', coalesce(c.to_name, '')) "to",
   c.payload, c.created_at as created_at, c.answered_at, c.bridged_at, c.hangup_at, c.hold_sec, c.cause, c.sip_code
   ,cc_get_lookup(cq.id, cq.name) as queue, cc_get_lookup(ct.id, ct.name) team, cc_get_lookup(ca.id, coalesce(ag.name, ag.username)) agent
   ,cc_get_lookup(cm.id, cm.name) member, f.files, c.duration
from cc_calls_history c
    left join lateral (
        select json_agg(jsonb_build_object('id', f.id, 'name', f.name, 'size', f.size, 'mime_type', f.mime_type)) files
        from storage.files f
        where f.domain_id = c.domain_id and f.uuid = c.id
    ) f on true
    left join cc_queue cq on c.queue_id = cq.id
    left join cc_team ct on c.team_id = ct.id
    left join cc_agent ca on c.agent_id = ca.id
    left join directory.wbt_user ag on ag.id = ca.user_id
    left join cc_member cm on c.member_id = cm.id
where c.domain_id = :Domain and c.created_at between :From::int8 and :To::int8 and (:UserIds::int8[] isnull or c.user_id = any(:UserIds))
	and (:QueueIds::int[] isnull or c.queue_id = any(:QueueIds) ) and (:TeamIds::int[] isnull or c.team_id = any(:TeamIds) )  and (:AgentIds::int[] isnull or c.agent_id = any(:AgentIds) )
	and (:MemberIds::int8[] isnull or c.member_id = any(:MemberIds) )
	and (:GatewayIds::int8[] isnull or c.gateway_id = any(:GatewayIds) )
	and (:Number::varchar isnull or c.from_number ilike :Number::varchar or c.to_number ilike :Number::varchar)
	and ( (:SkipParent::bool isnull or not :SkipParent::bool is true ) or c.parent_id isnull)
	and (:ParentId::varchar isnull or c.parent_id = :ParentId )
order by c.created_at desc
limit :Limit
offset :Offset`, map[string]interface{}{
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
	})

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
