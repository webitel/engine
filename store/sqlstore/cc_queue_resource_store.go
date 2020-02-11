package sqlstore

import (
	"fmt"
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

func (s SqlQueueResourceStore) Create(queueResource *model.QueueResourceGroup) (*model.QueueResourceGroup, *model.AppError) {
	var out *model.QueueResourceGroup
	if err := s.GetMaster().SelectOne(&out, `with q as (
		insert into cc_queue_resource (queue_id, resource_group_id)
		values (:QueueId, :ResourceGroupId)
		returning *
	)
select q.id, q.queue_id, cc_get_lookup(g.id, g.name::text) as resource_group
from q
    inner join cc_outbound_resource_group g on q.resource_group_id = g.id`,
		map[string]interface{}{
			"QueueId":         queueResource.QueueId,
			"ResourceGroupId": queueResource.ResourceGroup.Id,
		}); nil != err {
		return nil, model.NewAppError("SqlQueueResourceStore.Save", "store.sql_queue_resource.save.app_error", nil,
			fmt.Sprintf("queue_id=%v resource_group_id=%v, %v", queueResource.QueueId, queueResource.ResourceGroup.Id, err.Error()), extractCodeFromErr(err))
	} else {
		return out, nil
	}
}

func (s SqlQueueResourceStore) Get(domainId, queueId, id int64) (*model.QueueResourceGroup, *model.AppError) {
	var out *model.QueueResourceGroup
	if err := s.GetReplica().SelectOne(&out, `select q.id, q.queue_id, cc_get_lookup(g.id, g.name::text) as resource_group
		from cc_queue_resource q
			inner join cc_outbound_resource_group g on q.resource_group_id = g.id
		where q.id = :Id and q.queue_id = :QueueId and g.domain_id = :DomainId`, map[string]interface{}{
		"Id":       id,
		"DomainId": domainId,
		"QueueId":  queueId,
	}); err != nil {
		return nil, model.NewAppError("SqlQueueResourceStore.Get", "store.sql_queue_resource.get.app_error", nil,
			fmt.Sprintf("Id=%v, %s", id, err.Error()), extractCodeFromErr(err))
	} else {
		return out, nil
	}
}

func (s SqlQueueResourceStore) GetAllPage(domainId, queueId int64, search *model.SearchQueueResourceGroup) ([]*model.QueueResourceGroup, *model.AppError) {
	var out []*model.QueueResourceGroup

	if _, err := s.GetReplica().Select(&out,
		`select q.id, q.queue_id, cc_get_lookup(g.id, g.name::text) as resource_group
			from cc_queue_resource q
				inner join cc_outbound_resource_group g on q.resource_group_id = g.id
			where q.queue_id = :QueueId and g.domain_id = :DomainId
				and ( (:Q::varchar isnull or (g.name ilike :Q::varchar ) ))
			order by q.id
			limit :Limit
			offset :Offset`, map[string]interface{}{
			"DomainId": domainId,
			"Limit":    search.GetLimit(),
			"Offset":   search.GetOffset(),
			"Q":        search.GetQ(),
			"QueueId":  queueId,
		}); err != nil {
		return nil, model.NewAppError("SqlQueueResourceStore.GetAllPage", "store.sql_queue_resource.get_all.app_error",
			nil, err.Error(), extractCodeFromErr(err))
	} else {
		return out, nil
	}
}

func (s SqlQueueResourceStore) Update(domainId int64, queueResourceGroup *model.QueueResourceGroup) (*model.QueueResourceGroup, *model.AppError) {
	err := s.GetMaster().SelectOne(&queueResourceGroup, `with q as (
    update cc_queue_resource q
        set resource_group_id = :ResourceGroupId
        from cc_queue qq
        where q.id = :Id and q.queue_id = :QueueId and qq.id = q.queue_id and qq.domain_id = :DomainId
        returning q.*
)
select q.id, q.queue_id, cc_get_lookup(g.id, g.name::text) as resource_group
from  q
         inner join cc_outbound_resource_group g on q.resource_group_id = g.id`, map[string]interface{}{
		"ResourceGroupId": queueResourceGroup.ResourceGroup.Id,
		"Id":              queueResourceGroup.Id,
		"QueueId":         queueResourceGroup.QueueId,
		"DomainId":        domainId,
	})
	if err != nil {
		return nil, model.NewAppError("SqlQueueResourceStore.Update", "store.sql_queue_resource.update.app_error", nil,
			fmt.Sprintf("Id=%v, %s", queueResourceGroup.Id, err.Error()), extractCodeFromErr(err))
	}
	return queueResourceGroup, nil
}

func (s SqlQueueResourceStore) Delete(queueId, id int64) *model.AppError {
	if _, err := s.GetMaster().Exec(`delete from cc_queue_resource c where c.id=:Id and c.queue_id = :QueueId`,
		map[string]interface{}{"Id": id, "QueueId": queueId}); err != nil {
		return model.NewAppError("SqlQueueResourceStore.Delete", "store.sql_queue_resource.delete.app_error", nil,
			fmt.Sprintf("Id=%v, %s", id, err.Error()), extractCodeFromErr(err))
	}
	return nil
}
