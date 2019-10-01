package sqlstore

import (
	"fmt"
	"github.com/webitel/engine/model"
	"github.com/webitel/engine/store"
	"net/http"
)

type SqlQueueRoutingStore struct {
	SqlStore
}

func NewSqlQueueRoutingStore(sqlStore SqlStore) store.QueueRoutingStore {
	us := &SqlQueueRoutingStore{sqlStore}
	return us
}

func (s SqlQueueRoutingStore) Create(routing *model.QueueRouting) (*model.QueueRouting, *model.AppError) {
	var out *model.QueueRouting
	err := s.GetMaster().SelectOne(&out, `insert into cc_queue_routing (queue_id, pattern, priority, disabled)
values (:QueueId, :Pattern, :Priority, :Disabled)
returning *`, map[string]interface{}{
		"QueueId":  routing.QueueId,
		"Pattern":  routing.Pattern,
		"Priority": routing.Priority,
		"Disabled": routing.Disabled,
	})

	if err != nil {
		return nil, model.NewAppError("SqlQueueRoutingStore.Save", "store.sql_queue_routing.save.app_error", nil,
			fmt.Sprintf("record=%v, %v", routing, err.Error()), http.StatusInternalServerError)
	}

	return out, nil
}

func (s SqlQueueRoutingStore) GetAllPage(domainId, queueId int64, offset, limit int) ([]*model.QueueRouting, *model.AppError) {
	var out []*model.QueueRouting

	if _, err := s.GetReplica().Select(&out,
		`select t.id, t.queue_id, t.pattern, t.priority, t.disabled
			from cc_queue_routing t
			where t.queue_id = :QueueId
			order by t.priority desc
			limit :Limit
			offset :Offset`, map[string]interface{}{
			"DomainId": domainId,
			"Limit":    limit,
			"Offset":   offset,
			"QueueId":  queueId,
		}); err != nil {
		return nil, model.NewAppError("SqlQueueRoutingStore.GetAllPage", "store.sql_queue_routing.get_all.app_error", nil, err.Error(), http.StatusInternalServerError)
	} else {
		return out, nil
	}
}

func (s SqlQueueRoutingStore) Get(domainId, queueId int64, id int64) (*model.QueueRouting, *model.AppError) {
	var out *model.QueueRouting
	if err := s.GetReplica().SelectOne(&out, `
			select t.id, t.queue_id, t.pattern, t.priority, t.disabled
			from cc_queue_routing t
			where t.queue_id = :QueueId and t.id = :Id
		`, map[string]interface{}{
		"Id":      id,
		"QueueId": queueId,
	}); err != nil {
		return nil, model.NewAppError("SqlQueueRoutingStore.Get", "store.sql_queue_routing.get.app_error", nil,
			fmt.Sprintf("Id=%v, %s", id, err.Error()), extractCodeFromErr(err))
	} else {
		return out, nil
	}
}

func (s SqlQueueRoutingStore) Update(qr *model.QueueRouting) (*model.QueueRouting, *model.AppError) {
	err := s.GetMaster().SelectOne(&qr, `update cc_queue_routing r
		set pattern = :Pattern,
			priority = :Priority,
			disabled = :Disabled
		where r.queue_id = :QueueId and r.id = :Id
		returning *`, map[string]interface{}{
		"Pattern":  qr.Pattern,
		"Priority": qr.Priority,
		"Disabled": qr.Disabled,
		"QueueId":  qr.QueueId,
		"Id":       qr.Id,
	})
	if err != nil {
		return nil, model.NewAppError("SqlQueueRoutingStore.Update", "store.sql_queue_routing.update.app_error", nil,
			fmt.Sprintf("Id=%v, %s", qr.Id, err.Error()), extractCodeFromErr(err))
	}
	return qr, nil
}

func (s SqlQueueRoutingStore) Delete(queueId, id int64) *model.AppError {
	if _, err := s.GetMaster().Exec(`delete from cc_queue_routing c where c.id=:Id and c.queue_id = :QueueId`,
		map[string]interface{}{"Id": id, "QueueId": queueId}); err != nil {
		return model.NewAppError("SqlQueueRoutingStore.Delete", "store.sql_queue_routing.delete.app_error", nil,
			fmt.Sprintf("Id=%v, %s", id, err.Error()), http.StatusInternalServerError)
	}
	return nil
}
