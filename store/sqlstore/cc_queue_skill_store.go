package sqlstore

import (
	"context"
	"fmt"

	"github.com/lib/pq"
	"github.com/webitel/engine/model"
	"github.com/webitel/engine/store"
)

type SqlQueueSkillStore struct {
	SqlStore
}

func NewSqlQueueSkillStore(sqlStore SqlStore) store.QueueSkillStore {
	us := &SqlQueueSkillStore{sqlStore}
	return us
}

func (s SqlQueueSkillStore) Create(ctx context.Context, domainId int64, in *model.QueueSkill) (*model.QueueSkill, model.AppError) {
	var qs *model.QueueSkill

	err := s.GetMaster().WithContext(ctx).SelectOne(&qs, `with s as (
    insert into call_center.cc_queue_skill (queue_id, skill_id, bucket_ids, lvl, min_capacity, max_capacity, enabled)
    select :QueueId, :SkillId, :BucketIds, :Lvl, :MinCapacity, :MaxCapacity, :Enabled
    where exists(select 1 from call_center.cc_queue q where q.domain_id = :DomainId)
    returning *
)
select s.id,
       call_center.cc_get_lookup(cs.id, cs.name) skill,
       (select jsonb_agg(call_center.cc_get_lookup(b.id, b.name::varchar))
        from call_center.cc_bucket b
        where b.id = any (s.bucket_ids)
       )                             buckets,
       s.lvl,
       s.min_capacity,
       s.max_capacity,
       s.enabled
from s
         inner join call_center.cc_skill cs on s.skill_id = cs.id`, map[string]interface{}{
		"DomainId":    domainId,
		"QueueId":     in.QueueId,
		"SkillId":     in.Skill.Id,
		"BucketIds":   pq.Array(in.BucketIds()),
		"Lvl":         in.Lvl,
		"MinCapacity": in.MinCapacity,
		"MaxCapacity": in.MaxCapacity,
		"Enabled":     in.Enabled,
	})

	if err != nil {
		return nil, model.NewCustomCodeError("store.sql_queue_skill.create.app_error", fmt.Sprintf("name=%v, %v", in.QueueId, err.Error()), extractCodeFromErr(err))
	}

	return qs, nil
}

func (s SqlQueueSkillStore) Get(ctx context.Context, domainId int64, queueId, id uint32) (*model.QueueSkill, model.AppError) {
	var qs *model.QueueSkill

	err := s.GetReplica().WithContext(ctx).SelectOne(&qs, `select "id", "skill", "buckets", "lvl", "min_capacity", "max_capacity", "enabled"
		from call_center.cc_queue_skill_list
		where id = :Id and queue_id = :QueueId and domain_id = :DomainId
	`, map[string]interface{}{
		"Id":       id,
		"DomainId": domainId,
		"QueueId":  queueId,
	})

	if err != nil {
		return nil, model.NewCustomCodeError("store.sql_queue_skill.get.app_error", fmt.Sprintf("name=%v, %v", queueId, err.Error()), extractCodeFromErr(err))
	}

	return qs, nil
}

func (s SqlQueueSkillStore) GetAllPage(ctx context.Context, domainId int64, search *model.SearchQueueSkill) ([]*model.QueueSkill, model.AppError) {
	var qs []*model.QueueSkill

	f := map[string]interface{}{
		"DomainId":    domainId,
		"QueueId":     search.QueueId,
		"Q":           search.GetQ(),
		"Ids":         pq.Array(search.Ids),
		"SkillIds":    pq.Array(search.SkillIds),
		"BucketIds":   pq.Array(search.BucketIds),
		"Lvl":         pq.Array(search.Lvl),
		"MinCapacity": pq.Array(search.MinCapacity),
		"MaxCapacity": pq.Array(search.MaxCapacity),
		"Enabled":     search.Enabled,
	}

	err := s.ListQuery(ctx, &qs, search.ListRequest,
		`queue_id = :QueueId and domain_id = :DomainId
				and (:Ids::int4[] isnull or id = any(:Ids))
				and (:Q::text isnull or skill->>'name' ilike :Q::text)
				and (:SkillIds::int4[] isnull or skill_id = any(:SkillIds))
				and (:BucketIds::int4[] isnull or bucket_ids && :BucketIds)
				and (:Lvl::int4[] isnull or lvl = any(:Lvl))
				and (:MinCapacity::int4[] isnull or min_capacity = any(:MinCapacity))
				and (:MaxCapacity::int4[] isnull or max_capacity = any(:MaxCapacity))
				and (:Enabled::bool isnull or enabled = :Enabled)
			`,
		model.QueueSkill{}, f)
	if err != nil {
		return nil, model.NewCustomCodeError("store.sql_queue_skill.get_all.app_error", err.Error(), extractCodeFromErr(err))
	}

	return qs, nil
}

func (s SqlQueueSkillStore) Update(ctx context.Context, domainId int64, skill *model.QueueSkill) (*model.QueueSkill, model.AppError) {
	var qs *model.QueueSkill
	err := s.GetMaster().WithContext(ctx).SelectOne(&qs, `with s as (
    update call_center.cc_queue_skill s
    set skill_id = :SkillId,
        bucket_ids = :BucketIds,
        lvl = :Lvl,
        min_capacity = :MinCapacity,
        max_capacity = :MaxCapacity,
        enabled = :Enabled
    where s.id = :Id and exists(select 1 from call_center.cc_queue q where q.id = s.queue_id and q.domain_id = :DomainId)
    returning *
)
select s.id,
       call_center.cc_get_lookup(cs.id, cs.name) skill,
       (select jsonb_agg(call_center.cc_get_lookup(b.id, b.name::varchar))
        from call_center.cc_bucket b
        where b.id = any (s.bucket_ids)
       )                             buckets,
       s.lvl,
       s.min_capacity,
       s.max_capacity,
       s.enabled
from s
         inner join call_center.cc_skill cs on s.skill_id = cs.id`, map[string]interface{}{
		"Id":          skill.Id,
		"DomainId":    domainId,
		"SkillId":     skill.Skill.Id,
		"BucketIds":   pq.Array(skill.BucketIds()),
		"Lvl":         skill.Lvl,
		"MinCapacity": skill.MinCapacity,
		"MaxCapacity": skill.MaxCapacity,
		"Enabled":     skill.Enabled,
	})

	if err != nil {
		return nil, model.NewCustomCodeError("store.sql_queue_skill.update.app_error", err.Error(), extractCodeFromErr(err))
	}

	return qs, nil
}

func (s SqlQueueSkillStore) Delete(ctx context.Context, domainId int64, queueId, id uint32) model.AppError {
	if _, err := s.GetMaster().WithContext(ctx).Exec(`delete from call_center.cc_queue_skill s
where s.id = :Id and s.queue_id = :QueueId and exists(select 1 from call_center.cc_queue q where q.id = s.queue_id and q.domain_id = :DomainId)`,
		map[string]interface{}{
			"Id":       id,
			"DomainId": domainId,
			"QueueId":  queueId,
		}); err != nil {
		return model.NewCustomCodeError("store.sql_queue_skill.delete.app_error", fmt.Sprintf("Id=%v, %s", id, err.Error()), extractCodeFromErr(err))
	}
	return nil
}
