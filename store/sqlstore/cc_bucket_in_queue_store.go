package sqlstore

import (
	"context"
	"fmt"

	"github.com/lib/pq"
	"github.com/webitel/engine/model"
	"github.com/webitel/engine/store"
)

type SqlBucketInQueueStore struct {
	SqlStore
}

func NewSqlBucketInQueueStore(sqlStore SqlStore) store.BucketInQueueStore {
	us := &SqlBucketInQueueStore{sqlStore}
	return us
}

func (s SqlBucketInQueueStore) Create(ctx context.Context, queueBucket *model.QueueBucket) (*model.QueueBucket, model.AppError) {
	var out *model.QueueBucket
	if err := s.GetMaster().WithContext(ctx).SelectOne(&out, `with q as (
		insert into call_center.cc_bucket_in_queue (queue_id, ratio, bucket_id, priority, disabled)
		values (:QueueId, :Ratio, :BucketId, :Priority, :Disabled)
		returning *
	)
	select q.id, q.ratio, call_center.cc_get_lookup(cb.id, cb.name::text) as bucket, q.priority, q.disabled
	from q
		inner join call_center.cc_bucket cb on q.bucket_id = cb.id`,
		map[string]interface{}{
			"QueueId":  queueBucket.QueueId,
			"Ratio":    queueBucket.Ratio,
			"BucketId": queueBucket.Bucket.Id,
			"Priority": queueBucket.Priority,
			"Disabled": queueBucket.Disabled,
		}); nil != err {
		return nil, model.NewCustomCodeError("store.sql_queue_bucket.save.app_error", fmt.Sprintf("queue_id=%v bucket_id=%v, %v", queueBucket.QueueId, queueBucket.Bucket.Id, err.Error()), extractCodeFromErr(err))
	} else {
		return out, nil
	}
}

func (s SqlBucketInQueueStore) Get(ctx context.Context, domainId, queueId, id int64) (*model.QueueBucket, model.AppError) {
	var queueBucket *model.QueueBucket
	if err := s.GetReplica().WithContext(ctx).SelectOne(&queueBucket, `select q.id, 
			q.queue_id, q.ratio, call_center.cc_get_lookup(cb.id, cb.name::text) as bucket, q.priority, q.disabled
		from call_center.cc_bucket_in_queue q
			inner join call_center.cc_bucket cb on q.bucket_id = cb.id
		where q.id = :Id and q.queue_id = :QueueId and cb.domain_id = :DomainId`, map[string]interface{}{
		"Id":       id,
		"DomainId": domainId,
		"QueueId":  queueId,
	}); err != nil {
		return nil, model.NewCustomCodeError("store.sql_queue_bucket.get.app_error", fmt.Sprintf("Id=%v, %s", id, err.Error()), extractCodeFromErr(err))
	} else {
		return queueBucket, nil
	}
}

func (s SqlBucketInQueueStore) GetAllPage(ctx context.Context, domainId, queueId int64, search *model.SearchQueueBucket) ([]*model.QueueBucket, model.AppError) {
	var out []*model.QueueBucket

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
				and (:Q::varchar isnull or (bucket_name ilike :Q::varchar ))`,
		model.QueueBucket{}, f)

	if err != nil {
		return nil, model.NewCustomCodeError("store.sql_queue_bucket.get_all.app_error", err.Error(), extractCodeFromErr(err))
	} else {
		return out, nil
	}
}

func (s SqlBucketInQueueStore) Update(ctx context.Context, domainId int64, queueBucket *model.QueueBucket) (*model.QueueBucket, model.AppError) {
	err := s.GetMaster().WithContext(ctx).SelectOne(&queueBucket, `with q as (
		update call_center.cc_bucket_in_queue bq
			set ratio = :Ratio,
				bucket_id = :BucketId,
				priority = :Priority,
				disabled = :Disabled
		from call_center.cc_queue cq
		where bq.id = :Id and cq.id = :QueueId and cq.domain_id = :DomainId
		returning bq.*
	)
	select q.id, q.ratio, call_center.cc_get_lookup(cb.id, cb.name::text) as bucket, q.priority, q.disabled
	from q
		inner join call_center.cc_bucket cb on q.bucket_id = cb.id`, map[string]interface{}{
		"Ratio":    queueBucket.Ratio,
		"BucketId": queueBucket.Bucket.Id,
		"Id":       queueBucket.Id,
		"QueueId":  queueBucket.QueueId,
		"DomainId": domainId,
		"Priority": queueBucket.Priority,
		"Disabled": queueBucket.Disabled,
	})
	if err != nil {
		return nil, model.NewCustomCodeError("store.sql_queue_bucket.update.app_error", fmt.Sprintf("Id=%v, %s", queueBucket.Id, err.Error()), extractCodeFromErr(err))
	}
	return queueBucket, nil
}

func (s SqlBucketInQueueStore) Delete(ctx context.Context, queueId, id int64) model.AppError {
	if _, err := s.GetMaster().WithContext(ctx).Exec(`delete from call_center.cc_bucket_in_queue c where c.id=:Id and c.queue_id = :QueueId`,
		map[string]interface{}{"Id": id, "QueueId": queueId}); err != nil {
		return model.NewCustomCodeError("store.sql_queue_bucket.delete.app_error", fmt.Sprintf("Id=%v, %s", id, err.Error()), extractCodeFromErr(err))
	}
	return nil
}
