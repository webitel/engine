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

func (s SqlCallStore) GetAllPage(domainId int64, search *model.SearchCall) ([]*model.Call, *model.AppError) {
	var out []*model.Call

	_, err := s.GetMaster().Select(&out, `
select c.uuid                                                       as id,
       (extract(EPOCH from c.created_at) * 1000)                    as created_at,
       null::int                                                    as created_by,
       null::int8                                                   as timestamp,
       null::varchar                                                as parent_id,
       c.hostname                                                   as app_id,
       case c.call_state
           when 1 then 'ringing'
           when 2 then 'bla'
           else 'unknown' end                                       as state,
       case when c.direction = 1 then 'inbound' else 'outbound' end as direction,

       null::jsonb                                                  as from,
       null::jsonb                                                  as to
from directory.voip_channel c
         left join directory.wbt_user u on u.id = c.user_id
         left join directory.sip_gateway g on g.id = c.gateway_id
where c.domain_id = :DomainId
  and (:Q::varchar isnull or
       (u.name ilike :Q or g.name ilike :Q or c.caller_number ilike :Q or c.caller_name ilike :Q
           or c.callee_number ilike :Q or c.callee_name ilike :Q)
    )
order by c.created_at asc
limit :Limit
offset :Offset`, map[string]interface{}{
		"DomainId": domainId,
		"Limit":    search.GetLimit(),
		"Offset":   search.GetOffset(),
		"Q":        search.GetQ(),
	})

	if err != nil {
		return nil, model.NewAppError("SqlCallStore.GetAllPage", "store.sql_call.get_all.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	return out, nil
}

func (s SqlCallStore) Get(domainId int64, id string) (*model.Call, *model.AppError) {
	var out *model.Call

	err := s.GetMaster().SelectOne(&out, `
select c.uuid                                                       as id,
       (extract(EPOCH from c.created_at) * 1000)                    as created_at,
       null::int                                                    as created_by,
       null::int8                                                   as timestamp,
       null::varchar                                                as parent_id,
       c.hostname                                                   as app_id,
       case c.call_state
           when 1 then 'ringing'
           when 2 then 'bla'
           else 'unknown' end                                       as state,
       case when c.direction = 1 then 'inbound' else 'outbound' end as direction,

       null::jsonb                                                  as from,
       null::jsonb                                                  as to
from directory.voip_channel c
         left join directory.wbt_user u on u.id = c.user_id
         left join directory.sip_gateway g on g.id = c.gateway_id
where c.domain_id = :DomainId and c.uuid = :Id`, map[string]interface{}{
		"DomainId": domainId,
		"Id":       id,
	})

	if err != nil {
		return nil, model.NewAppError("SqlCallStore.Get", "store.sql_call.get.app_error", nil, err.Error(), extractCodeFromErr(err))
	}

	return out, nil
}

func (s SqlCallStore) GetInstance(domainId int64, id string) (*model.CallInstance, *model.AppError) {
	var inst *model.CallInstance
	err := s.GetMaster().SelectOne(&inst, `select c.uuid as id, c.hostname as app_id, 'tood' as state
from directory.voip_channel c
where c.uuid = :Id and c.domain_id = :DomainId`, map[string]interface{}{
		"Id":       id,
		"DomainId": domainId,
	})
	if err != nil {
		return nil, model.NewAppError("SqlCallStore.GetInstance", "store.sql_call.get_instance.app_error", nil, err.Error(), extractCodeFromErr(err))
	}

	return inst, nil
}
