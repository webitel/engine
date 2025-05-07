package sqlstore

import (
	"context"
	"fmt"

	"github.com/lib/pq"
	"github.com/webitel/engine/model"
	"github.com/webitel/engine/pkg/wbt/auth_manager"
	"github.com/webitel/engine/store"
)

type SqlBucketStore struct {
	SqlStore
}

func NewSqlBucketStore(sqlStore SqlStore) store.BucketStore {
	us := &SqlBucketStore{sqlStore}
	return us
}

func (s SqlBucketStore) Create(ctx context.Context, bucket *model.Bucket) (*model.Bucket, model.AppError) {
	var out *model.Bucket
	if err := s.GetMaster().WithContext(ctx).SelectOne(&out, `insert into call_center.cc_bucket (name, domain_id, description, created_at, created_by, updated_at, updated_by)
		values (:Name, :DomainId, :Description, :CreatedAt, :CreatedBy, :UpdatedAt, :UpdatedBy)
		returning id, name, description, domain_id`,
		map[string]interface{}{
			"Name":        bucket.Name,
			"DomainId":    bucket.DomainId,
			"Description": bucket.Description,
			"CreatedAt":   bucket.CreatedAt,
			"CreatedBy":   bucket.CreatedBy.GetSafeId(),
			"UpdatedAt":   bucket.UpdatedAt,
			"UpdatedBy":   bucket.UpdatedBy.GetSafeId(),
		}); nil != err {
		return nil, model.NewCustomCodeError("store.sql_bucket.save.app_error", fmt.Sprintf("name=%v, %v", bucket.Name, err.Error()), extractCodeFromErr(err))
	} else {
		return out, nil
	}
}

func (s SqlBucketStore) CheckAccess(ctx context.Context, domainId, id int64, groups []int, access auth_manager.PermissionAccess) (bool, model.AppError) {

	res, err := s.GetReplica().WithContext(ctx).SelectNullInt(`select 1
		where exists(
          select 1
          from call_center.cc_bucket_acl a
          where a.dc = :DomainId
            and a.object = :Id
            and a.subject = any (:Groups::int[])
            and a.access & :Access = :Access
        )`, map[string]interface{}{"DomainId": domainId, "Id": id, "Groups": pq.Array(groups), "Access": access.Value()})

	if err != nil {
		return false, nil
	}

	return res.Valid && res.Int64 == 1, nil
}

func (s SqlBucketStore) GetAllPage(ctx context.Context, domainId int64, search *model.SearchBucket) ([]*model.Bucket, model.AppError) {
	var buckets []*model.Bucket

	f := map[string]interface{}{
		"DomainId": domainId,
		"Ids":      pq.Array(search.Ids),
		"Q":        search.GetQ(),
	}

	err := s.ListQuery(ctx, &buckets, search.ListRequest,
		`domain_id = :DomainId
				and (:Ids::int[] isnull or id = any(:Ids))
				and (:Q::varchar isnull or (name ilike :Q::varchar or description ilike :Q::varchar))`,
		model.Bucket{}, f)

	if err != nil {
		return nil, model.NewInternalError("store.sql_bucket.get_all.app_error", err.Error())
	} else {
		return buckets, nil
	}
}

func (s SqlBucketStore) Get(ctx context.Context, domainId int64, id int64) (*model.Bucket, model.AppError) {
	var bucket *model.Bucket
	if err := s.GetReplica().WithContext(ctx).SelectOne(&bucket, `select id, name, description, domain_id
		from call_center.cc_bucket b
		where b.id = :Id and b.domain_id = :DomainId`, map[string]interface{}{"Id": id, "DomainId": domainId}); err != nil {
		return nil, model.NewCustomCodeError("store.sql_bucket.get.app_error", fmt.Sprintf("Id=%v, %s", id, err.Error()), extractCodeFromErr(err))
	} else {
		return bucket, nil
	}
}

func (s SqlBucketStore) Update(ctx context.Context, bucket *model.Bucket) (*model.Bucket, model.AppError) {
	err := s.GetMaster().WithContext(ctx).SelectOne(&bucket, `update call_center.cc_bucket
	set name = :Name,
    description = :Description,
	updated_at = :UpdatedAt,
	updated_by = :UpdatedBy
		where id = :Id and domain_id = :DomainId returning id, name, description, domain_id`, map[string]interface{}{
		"Id":          bucket.Id,
		"Name":        bucket.Name,
		"Description": bucket.Description,
		"DomainId":    bucket.DomainId,
		"UpdatedAt":   bucket.UpdatedAt,
		"UpdatedBy":   bucket.UpdatedBy.GetSafeId(),
	})
	if err != nil {
		return nil, model.NewCustomCodeError("store.sql_bucket.update.app_error", fmt.Sprintf("Id=%v, %s", bucket.Id, err.Error()), extractCodeFromErr(err))
	}
	return bucket, nil
}

func (s SqlBucketStore) Delete(ctx context.Context, domainId int64, id int64) model.AppError {
	if _, err := s.GetMaster().WithContext(ctx).Exec(`delete from call_center.cc_bucket c where c.id=:Id and c.domain_id = :DomainId`,
		map[string]interface{}{"Id": id, "DomainId": domainId}); err != nil {
		return model.NewCustomCodeError("store.sql_bucket.delete.app_error", fmt.Sprintf("Id=%v, %s", id, err.Error()), extractCodeFromErr(err))
	}
	return nil
}
