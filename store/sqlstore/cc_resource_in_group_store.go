package sqlstore

import (
	"context"
	"fmt"
	"github.com/lib/pq"
	"github.com/webitel/engine/model"
	"github.com/webitel/engine/store"
	"net/http"
)

type SqlOutboundResourceInGroupStore struct {
	SqlStore
}

func NewSqlOutboundResourceInGroupStore(sqlStore SqlStore) store.OutboundResourceInGroupStore {
	us := &SqlOutboundResourceInGroupStore{sqlStore}
	return us
}

func (s SqlOutboundResourceInGroupStore) Create(ctx context.Context, domainId int64, res *model.OutboundResourceInGroup) (*model.OutboundResourceInGroup, *model.AppError) {
	var out *model.OutboundResourceInGroup
	if err := s.GetMaster().WithContext(ctx).SelectOne(&out, `with s as (
    insert into call_center.cc_outbound_resource_in_group (resource_id, group_id, reserve_resource_id, priority)
    select :ResourceId, :GroupId, :ReserveResourceId, :Priority
    where exists(select 1 from call_center.cc_outbound_resource_group where domain_id = :DomainId)
    returning *
)
select s.id, s.group_id, call_center.cc_get_lookup(cor.id, cor.name) as resource,
	call_center.cc_get_lookup(res.id::bigint, res.name) AS reserve_resource,
	s.priority
from s
    inner join call_center.cc_outbound_resource cor on s.resource_id = cor.id
	left join call_center.cc_outbound_resource res on res.id = s.reserve_resource_id`,
		map[string]interface{}{
			"DomainId":          domainId,
			"ResourceId":        res.Resource.GetSafeId(),
			"ReserveResourceId": res.ReserveResource.GetSafeId(),
			"Priority":          res.Priority,
			"GroupId":           res.GroupId,
		}); nil != err {
		return nil, model.NewAppError("SqlOutboundResourceInGroupStore.Save", "store.sql_out_resource_in_group.save.app_error", nil,
			fmt.Sprintf("GroupId=%v, %v", res.GroupId, err.Error()), extractCodeFromErr(err))
	} else {
		return out, nil
	}
}

func (s SqlOutboundResourceInGroupStore) GetAllPage(ctx context.Context, domainId, groupId int64, search *model.SearchOutboundResourceInGroup) ([]*model.OutboundResourceInGroup, *model.AppError) {
	var groups []*model.OutboundResourceInGroup

	f := map[string]interface{}{
		"DomainId": domainId,
		"GroupId":  groupId,
		"Ids":      pq.Array(search.Ids),
		"Q":        search.GetQ(),
	}

	err := s.ListQuery(ctx, &groups, search.ListRequest,
		`domain_id = :DomainId
				and group_id = :GroupId
				and (:Ids::int[] isnull or id = any(:Ids))
				and (:Q::varchar isnull or (resource_name ilike :Q::varchar ))`,
		model.OutboundResourceInGroup{}, f)

	if err != nil {
		return nil, model.NewAppError("SqlOutboundResourceInGroupStore.GetAllPage", "store.sql_out_resource_in_group.get_all.app_error", nil,
			fmt.Sprintf("DomainId=%v, %s", domainId, err.Error()), extractCodeFromErr(err))
	} else {
		return groups, nil
	}
}

func (s SqlOutboundResourceInGroupStore) Get(ctx context.Context, domainId, groupId, id int64) (*model.OutboundResourceInGroup, *model.AppError) {
	var res *model.OutboundResourceInGroup
	if err := s.GetReplica().WithContext(ctx).SelectOne(&res, `
			select s.id, s.group_id, resource, reserve_resource, priority
			from call_center.cc_outbound_resource_in_group_view s
			where s.group_id = :GroupId and s.domain_id = :DomainId	and s.id = :Id
		`, map[string]interface{}{"Id": id, "DomainId": domainId, "GroupId": groupId}); err != nil {
		return nil, model.NewAppError("SqlOutboundResourceInGroupStore.Get", "store.sql_out_resource_in_group.get.app_error", nil,
			fmt.Sprintf("Id=%v, %s", id, err.Error()), extractCodeFromErr(err))
	} else {
		return res, nil
	}
}

func (s SqlOutboundResourceInGroupStore) Update(ctx context.Context, domainId int64, res *model.OutboundResourceInGroup) (*model.OutboundResourceInGroup, *model.AppError) {

	err := s.GetMaster().WithContext(ctx).SelectOne(&res, `with s as (
    update call_center.cc_outbound_resource_in_group 
        set resource_id  = :ResourceId,
			reserve_resource_id = :ReserveResourceId,
			priority = :Priority
    where id = :Id and group_id = :GroupId and exists(select 1 from call_center.cc_outbound_resource_group where domain_id = :DomainId)
    returning *
)
SELECT s.id,
       s.group_id,
       call_center.cc_get_lookup(cor.id::bigint, cor.name) AS resource,
       call_center.cc_get_lookup(res.id::bigint, res.name) AS reserve_resource,
       s.priority
FROM s
         LEFT JOIN call_center.cc_outbound_resource cor ON s.resource_id = cor.id
         LEFT JOIN call_center.cc_outbound_resource_group corg ON s.group_id = corg.id
         left join call_center.cc_outbound_resource res on res.id = s.reserve_resource_id`, map[string]interface{}{
		"ResourceId":        res.Resource.Id,
		"GroupId":           res.GroupId,
		"Id":                res.Id,
		"DomainId":          domainId,
		"ReserveResourceId": res.ReserveResource.GetSafeId(),
		"Priority":          res.Priority,
	})

	if err != nil {
		return nil, model.NewAppError("SqlOutboundResourceInGroupStore.Update", "store.sql_out_resource_in_group.update.app_error", nil,
			fmt.Sprintf("Id=%v, %s", res.Id, err.Error()), extractCodeFromErr(err))
	}

	return res, nil
}

func (s SqlOutboundResourceInGroupStore) Delete(ctx context.Context, domainId, groupId, id int64) *model.AppError {
	if _, err := s.GetMaster().WithContext(ctx).Exec(`delete from call_center.cc_outbound_resource_in_group c 
			where id = :Id and group_id = :GroupId and exists(select 1 from call_center.cc_outbound_resource_group where domain_id = :DomainId)`,
		map[string]interface{}{"Id": id, "DomainId": domainId, "GroupId": groupId}); err != nil {
		return model.NewAppError("SqlOutboundResourceGroupStore.Delete", "store.sql_out_resource_group.delete.app_error", nil,
			fmt.Sprintf("Id=%v, %s", id, err.Error()), http.StatusInternalServerError)
	}
	return nil
}
