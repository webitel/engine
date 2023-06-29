package sqlstore

import (
	"context"
	"fmt"

	"github.com/lib/pq"
	"github.com/webitel/engine/model"
	"github.com/webitel/engine/store"
)

type SqlQueueResourceStore struct {
	SqlStore
}

func NewSqlQueueResourceStore(sqlStore SqlStore) store.QueueResourceStore {
	us := &SqlQueueResourceStore{sqlStore}
	return us
}

func (s SqlQueueResourceStore) Create(ctx context.Context, queueResource *model.QueueResourceGroup) (*model.QueueResourceGroup, model.AppError) {
	var out *model.QueueResourceGroup
	if err := s.GetMaster().WithContext(ctx).SelectOne(&out, `with q as (
		insert into call_center.cc_queue_resource (queue_id, resource_group_id)
		values (:QueueId, :ResourceGroupId)
		returning *
	)
select q.id, q.queue_id, call_center.cc_get_lookup(g.id, g.name::text) as resource_group,
	call_center.cc_get_lookup(c.id, c.name::text::character varying) AS communication
from q
    inner join call_center.cc_outbound_resource_group g on q.resource_group_id = g.id
    left join call_center.cc_communication c on c.id = g.communication_id
`,
		map[string]interface{}{
			"QueueId":         queueResource.QueueId,
			"ResourceGroupId": queueResource.ResourceGroup.Id,
		}); nil != err {
		return nil, model.NewCustomCodeError("store.sql_queue_resource.save.app_error", fmt.Sprintf("queue_id=%v resource_group_id=%v, %v", queueResource.QueueId, queueResource.ResourceGroup.Id, err.Error()), extractCodeFromErr(err))
	} else {
		return out, nil
	}
}

func (s SqlQueueResourceStore) Get(ctx context.Context, domainId, queueId, id int64) (*model.QueueResourceGroup, model.AppError) {
	var out *model.QueueResourceGroup
	if err := s.GetReplica().WithContext(ctx).SelectOne(&out, `select q.id, q.queue_id, call_center.cc_get_lookup(g.id, g.name::text) as resource_group,
			call_center.cc_get_lookup(c.id, c.name::text::character varying) AS communication
		from call_center.cc_queue_resource q
			inner join call_center.cc_outbound_resource_group g on q.resource_group_id = g.id
			left join call_center.cc_communication c on c.id = g.communication_id
		where q.id = :Id and q.queue_id = :QueueId and g.domain_id = :DomainId`, map[string]interface{}{
		"Id":       id,
		"DomainId": domainId,
		"QueueId":  queueId,
	}); err != nil {
		return nil, model.NewCustomCodeError("store.sql_queue_resource.get.app_error", fmt.Sprintf("Id=%v, %s", id, err.Error()), extractCodeFromErr(err))
	} else {
		return out, nil
	}
}

func (s SqlQueueResourceStore) GetAllPage(ctx context.Context, domainId, queueId int64, search *model.SearchQueueResourceGroup) ([]*model.QueueResourceGroup, model.AppError) {
	var out []*model.QueueResourceGroup

	f := map[string]interface{}{
		"DomainId": domainId,
		"QueueId":  queueId,
		"Q":        search.GetQ(),
		"Ids":      pq.Array(search.Ids),
	}

	err := s.ListQuery(ctx, &out, search.ListRequest,
		`domain_id = :DomainId
				and queue_id = :QueueId
				and (:Ids::int[] isnull or id = any(:Ids))
				and (:Q::varchar isnull or (resource_group_name ilike :Q::varchar ))`,
		model.QueueResourceGroup{}, f)

	if err != nil {
		return nil, model.NewCustomCodeError("store.sql_queue_resource.get_all.app_error", err.Error(), extractCodeFromErr(err))
	} else {
		return out, nil
	}
}

func (s SqlQueueResourceStore) Update(ctx context.Context, domainId int64, queueResourceGroup *model.QueueResourceGroup) (*model.QueueResourceGroup, model.AppError) {
	err := s.GetMaster().WithContext(ctx).SelectOne(&queueResourceGroup, `with q as (
    update call_center.cc_queue_resource q
        set resource_group_id = :ResourceGroupId
        from call_center.cc_queue qq
        where q.id = :Id and q.queue_id = :QueueId and qq.id = q.queue_id and qq.domain_id = :DomainId
        returning q.*
)
select q.id, q.queue_id, call_center.cc_get_lookup(g.id, g.name::text) as resource_group,
	call_center.cc_get_lookup(c.id, c.name::text::character varying) AS communication
from  q
         inner join call_center.cc_outbound_resource_group g on q.resource_group_id = g.id
		 left join call_center.cc_communication c on c.id = g.communication_id
`, map[string]interface{}{
		"ResourceGroupId": queueResourceGroup.ResourceGroup.Id,
		"Id":              queueResourceGroup.Id,
		"QueueId":         queueResourceGroup.QueueId,
		"DomainId":        domainId,
	})
	if err != nil {
		return nil, model.NewCustomCodeError("store.sql_queue_resource.update.app_error", fmt.Sprintf("Id=%v, %s", queueResourceGroup.Id, err.Error()), extractCodeFromErr(err))
	}
	return queueResourceGroup, nil
}

func (s SqlQueueResourceStore) Delete(ctx context.Context, queueId, id int64) model.AppError {
	if _, err := s.GetMaster().WithContext(ctx).Exec(`delete from call_center.cc_queue_resource c where c.id=:Id and c.queue_id = :QueueId`,
		map[string]interface{}{"Id": id, "QueueId": queueId}); err != nil {
		return model.NewCustomCodeError("store.sql_queue_resource.delete.app_error", fmt.Sprintf("Id=%v, %s", id, err.Error()), extractCodeFromErr(err))
	}
	return nil
}
