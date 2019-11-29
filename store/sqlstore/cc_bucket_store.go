package sqlstore

import (
	"fmt"
	"github.com/lib/pq"
	"github.com/webitel/engine/model"
	"github.com/webitel/engine/store"
	"net/http"
)

type SqlBucketStore struct {
	SqlStore
}

func NewSqlBucketStore(sqlStore SqlStore) store.BucketSore {
	us := &SqlBucketStore{sqlStore}
	return us
}

func (s SqlBucketStore) Create(bucket *model.Bucket) (*model.Bucket, *model.AppError) {
	var out *model.Bucket
	if err := s.GetMaster().SelectOne(&out, `insert into cc_bucket (name, domain_id, description)
		values (:Name, :DomainId, :Description)
		returning *`,
		map[string]interface{}{"Name": bucket.Name, "DomainId": bucket.DomainId, "Description": bucket.Description}); nil != err {
		return nil, model.NewAppError("SqlBucketStore.Save", "store.sql_bucket.save.app_error", nil,
			fmt.Sprintf("name=%v, %v", bucket.Name, err.Error()), extractCodeFromErr(err))
	} else {
		return out, nil
	}
}

func (s SqlBucketStore) CheckAccess(domainId, id int64, groups []int, access model.PermissionAccess) (bool, *model.AppError) {

	res, err := s.GetReplica().SelectNullInt(`select 1
		where exists(
          select 1
          from cc_bucket_acl a
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

func (s SqlBucketStore) GetAllPage(domainId int64, offset, limit int) ([]*model.Bucket, *model.AppError) {
	var buckets []*model.Bucket

	if _, err := s.GetReplica().Select(&buckets,
		`select b.id,
       b.name,
       b.description
from cc_bucket b
where b.domain_id = :DomainId
order by b.id
limit :Limit
offset :Offset`, map[string]interface{}{"DomainId": domainId, "Limit": limit, "Offset": offset}); err != nil {
		return nil, model.NewAppError("SqlBucketStore.GetAllPage", "store.sql_bucket.get_all.app_error", nil, err.Error(), http.StatusInternalServerError)
	} else {
		return buckets, nil
	}
}

func (s SqlBucketStore) GetAllPageByGroups(domainId int64, groups []int, offset, limit int) ([]*model.Bucket, *model.AppError) {
	var buckets []*model.Bucket

	if _, err := s.GetReplica().Select(&buckets,
		`select b.id,
       b.name,
       b.description
from cc_bucket b
where b.domain_id = :DomainId and (
    exists(select 1
      from cc_bucket_acl a
      where a.dc = b.domain_id and a.object = b.id and a.subject = any(:Groups::int[]) and a.access&:Access = :Access)
  )
order by b.id
limit :Limit
offset :Offset`, map[string]interface{}{"DomainId": domainId, "Limit": limit, "Offset": offset, "Groups": pq.Array(groups), "Access": model.PERMISSION_ACCESS_READ.Value()}); err != nil {
		return nil, model.NewAppError("SqlBucketStore.GetAllPage", "store.sql_bucket.get_all.app_error", nil, err.Error(), http.StatusInternalServerError)
	} else {
		return buckets, nil
	}
}

func (s SqlBucketStore) Get(domainId int64, id int64) (*model.Bucket, *model.AppError) {
	var bucket *model.Bucket
	if err := s.GetReplica().SelectOne(&bucket, `select *
		from cc_bucket b
		where b.id = :Id and b.domain_id = :DomainId`, map[string]interface{}{"Id": id, "DomainId": domainId}); err != nil {
		return nil, model.NewAppError("SqlBucketStore.Get", "store.sql_bucket.get.app_error", nil,
			fmt.Sprintf("Id=%v, %s", id, err.Error()), extractCodeFromErr(err))
	} else {
		return bucket, nil
	}
}

func (s SqlBucketStore) Update(bucket *model.Bucket) (*model.Bucket, *model.AppError) {
	err := s.GetMaster().SelectOne(&bucket, `update cc_bucket
	set name = :Name,
    description = :Description
		where id = :Id and domain_id = :DomainId returning *`, map[string]interface{}{
		"Id":          bucket.Id,
		"Name":        bucket.Name,
		"Description": bucket.Description,
		"DomainId":    bucket.DomainId,
	})
	if err != nil {
		return nil, model.NewAppError("SqlBucketStore.Update", "store.sql_bucket.update.app_error", nil,
			fmt.Sprintf("Id=%v, %s", bucket.Id, err.Error()), extractCodeFromErr(err))
	}
	return bucket, nil
}

func (s SqlBucketStore) Delete(domainId int64, id int64) *model.AppError {
	if _, err := s.GetMaster().Exec(`delete from cc_bucket c where c.id=:Id and c.domain_id = :DomainId`,
		map[string]interface{}{"Id": id, "DomainId": domainId}); err != nil {
		return model.NewAppError("SqlBucketStore.Delete", "store.sql_bucket.delete.app_error", nil,
			fmt.Sprintf("Id=%v, %s", id, err.Error()), extractCodeFromErr(err))
	}
	return nil
}
