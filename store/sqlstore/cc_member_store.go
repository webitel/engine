package sqlstore

import (
	"database/sql"
	"fmt"
	"github.com/go-gorp/gorp"
	"github.com/lib/pq"
	"github.com/webitel/engine/model"
	"github.com/webitel/engine/store"
	"net/http"
)

type SqlMemberStore struct {
	SqlStore
}

func NewSqlMemberStore(sqlStore SqlStore) store.MemberStore {
	us := &SqlMemberStore{sqlStore}
	return us
}

func (s SqlMemberStore) Create(member *model.Member) (*model.Member, *model.AppError) {
	var out *model.Member
	if err := s.GetMaster().SelectOne(&out, `with m as (
			insert into cc_member (queue_id, priority, expire_at, variables, name, timezone_id, communications, bucket_id)
			values (:QueueId, :Priority, :ExpireAt, :Variables, :Name, :TimezoneId, :Communications, :BucketId)
			returning *
		)
		select m.id, m.queue_id, m.priority, m.expire_at, m.variables, m.name, cc_get_lookup(ct.id, ct.name) as "timezone",
			   m.communications,  cc_get_lookup(qb.id, qb.name::text) as bucket
		from m
			left join calendar_timezones ct on m.timezone_id = ct.id
			left join cc_bucket qb on m.bucket_id = qb.id`,
		map[string]interface{}{
			"QueueId":        member.QueueId,
			"Priority":       member.Priority,
			"ExpireAt":       member.GetExpireAt(),
			"Variables":      member.Variables.ToJson(),
			"Name":           member.Name,
			"TimezoneId":     member.Timezone.Id,
			"Communications": member.ToJsonCommunications(),
			"BucketId":       member.GetBucketId(),
		}); nil != err {
		return nil, model.NewAppError("SqlMemberStore.Save", "store.sql_member.save.app_error", nil,
			fmt.Sprintf("name=%v, %v", member.Name, err.Error()), http.StatusInternalServerError)
	} else {
		return out, nil
	}
}

func (s SqlMemberStore) BulkCreate(queueId int64, members []*model.Member) ([]int64, *model.AppError) {
	var err error
	var stmp *sql.Stmt
	var tx *gorp.Transaction
	tx, err = s.GetMaster().Begin()
	if err != nil {
		return nil, model.NewAppError("SqlMemberStore.Save", "store.sql_member.bulk_save.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	_, err = tx.Exec("CREATE temp table cc_member_tmp ON COMMIT DROP as table cc_member with no data")
	if err != nil {
		return nil, model.NewAppError("SqlMemberStore.Save", "store.sql_member.bulk_save.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	stmp, err = tx.Prepare(pq.CopyIn("cc_member_tmp", "id", "queue_id", "priority", "expire_at", "variables", "name",
		"timezone_id", "communications", "bucket_id"))
	if err != nil {
		return nil, model.NewAppError("SqlMemberStore.Save", "store.sql_member.bulk_save.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	defer stmp.Close()
	result := make([]int64, 0, len(members))
	for k, v := range members {
		_, err = stmp.Exec(k, queueId, v.Priority, v.GetExpireAt(), v.Variables.ToJson(), v.Name, v.Timezone.Id, v.ToJsonCommunications(), v.GetBucketId())
		if err != nil {
			goto _error
		}
	}

	_, err = stmp.Exec()
	if err != nil {
		goto _error
	} else {

		_, err = tx.Select(&result, `with i as (
			insert into cc_member(queue_id, priority, expire_at, variables, name, timezone_id, communications, bucket_id)
			select queue_id, priority, expire_at, variables, name, timezone_id, communications, bucket_id
			from cc_member_tmp
			returning id
		)
		select id from i`)
		if err != nil {
			goto _error
		}
	}

	err = tx.Commit()
	if err != nil {
		return nil, model.NewAppError("SqlMemberStore.Save", "store.sql_member.bulk_save.app_error", nil, err.Error(), extractCodeFromErr(err))
	}

	return result, nil

_error:
	tx.Rollback()
	return nil, model.NewAppError("SqlMemberStore.Save", "store.sql_member.bulk_save.app_error", nil, err.Error(), extractCodeFromErr(err))
}

func (s SqlMemberStore) GetAllPage(domainId, queueId int64, offset, limit int) ([]*model.Member, *model.AppError) {
	var members []*model.Member

	if _, err := s.GetReplica().Select(&members,
		`select m.id, m.queue_id, m.priority, m.expire_at, m.variables, m.name, cc_get_lookup(ct.id, ct.name) as "timezone",
			   m.communications, null as bucket
		from cc_member m
			left join calendar_timezones ct on m.timezone_id = ct.id
where m.queue_id = :QueueId and exists(select 1 from cc_queue q where q.id = :QueueId and q.domain_id = :DomainId)
order by m.id
limit :Limit
offset :Offset`, map[string]interface{}{
			"QueueId":  queueId,
			"DomainId": domainId,
			"Limit":    limit,
			"Offset":   offset,
		}); err != nil {
		return nil, model.NewAppError("SqlMemberStore.GetAllPage", "store.sql_member.get_all.app_error", nil, err.Error(), http.StatusInternalServerError)
	} else {
		return members, nil
	}
}

func (s SqlMemberStore) Get(domainId, queueId, id int64) (*model.Member, *model.AppError) {
	var member *model.Member
	if err := s.GetReplica().SelectOne(&member, `select m.id, m.queue_id, m.priority, m.expire_at, m.variables, m.name, cc_get_lookup(ct.id, ct.name) as "timezone",
				   m.communications, null as bucket
			from cc_member m
				left join calendar_timezones ct on m.timezone_id = ct.id
	where m.id = :Id and m.queue_id = :QueueId and exists(select 1 from cc_queue q where q.id = :QueueId and q.domain_id = :DomainId)`, map[string]interface{}{
		"Id":       id,
		"DomainId": domainId,
		"QueueId":  queueId,
	}); err != nil {
		return nil, model.NewAppError("SqlMemberStore.Get", "store.sql_member.get.app_error", nil,
			fmt.Sprintf("Id=%v, %s", id, err.Error()), extractCodeFromErr(err))
	} else {
		return member, nil
	}
}

func (s SqlMemberStore) Update(domainId int64, member *model.Member) (*model.Member, *model.AppError) {
	err := s.GetMaster().SelectOne(&member, `with m as (
    update cc_member m1
        set priority = :Priority,
            expire_at = :ExpireAt,
            variables = :Variables,
            name = :Name,
            timezone_id = :TimezoneId,
            communications = :Communications,
            bucket_id = :BucketId
    where m1.id = :Id and m1.queue_id = :QueueId
    returning *
)
select m.id, m.queue_id, m.priority, m.expire_at, m.variables, m.name, cc_get_lookup(ct.id, ct.name) as "timezone",
       m.communications,  cc_get_lookup(qb.id, qb.name::text) as bucket
from m
    left join calendar_timezones ct on m.timezone_id = ct.id
    left join cc_bucket qb on m.bucket_id = qb.id`, map[string]interface{}{
		"Priority":       member.Priority,
		"ExpireAt":       member.GetExpireAt(),
		"Variables":      member.Variables.ToJson(),
		"Name":           member.Name,
		"TimezoneId":     member.Timezone.Id,
		"Communications": member.ToJsonCommunications(),
		"BucketId":       member.GetBucketId(),
		"Id":             member.Id,
		"QueueId":        member.QueueId,
		"DomainId":       domainId,
	})
	if err != nil {
		return nil, model.NewAppError("SqlMemberStore.Update", "store.sql_member.update.app_error", nil,
			fmt.Sprintf("Id=%v, %s", member.Id, err.Error()), extractCodeFromErr(err))
	}
	return member, nil
}

func (s SqlMemberStore) Delete(queueId, id int64) *model.AppError {
	if _, err := s.GetMaster().Exec(`delete from cc_member c where c.id=:Id and c.queue_id = :QueueId`,
		map[string]interface{}{"Id": id, "QueueId": queueId}); err != nil {
		return model.NewAppError("SqlMemberStore.Delete", "store.sql_member.delete.app_error", nil,
			fmt.Sprintf("Id=%v, %s", id, err.Error()), extractCodeFromErr(err))
	}
	return nil
}
