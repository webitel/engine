package sqlstore

import (
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

func (s SqlQueueSkillStore) Create(domainId int64, in *model.QueueSkill) (*model.QueueSkill, *model.AppError) {
	var qs *model.QueueSkill

	err := s.GetMaster().SelectOne(&qs, `with s as (
    insert into cc_queue_skill (queue_id, skill_id, bucket_ids, lvl, min_capacity, max_capacity, disabled)
    select :QueueId, :SkillId, :BucketIds, :Lvl, :MinCapacity, :MaxCapacity, :Disabled
    where exists(select 1 from cc_queue q where q.domain_id = :DomainId)
    returning *
)
select s.id,
       cc_get_lookup(cs.id, cs.name) skill,
       (select jsonb_agg(cc_get_lookup(b.id, b.name::varchar))
        from cc_bucket b
        where b.id = any (s.bucket_ids)
       )                             buckets,
       s.lvl,
       s.min_capacity,
       s.max_capacity,
       s.disabled
from s
         inner join cc_skill cs on s.skill_id = cs.id`, map[string]interface{}{
		"DomainId":    domainId,
		"QueueId":     in.QueueId,
		"SkillId":     in.Skill.Id,
		"BucketIds":   pq.Array(in.BucketIds()),
		"Lvl":         in.Lvl,
		"MinCapacity": in.MinCapacity,
		"MaxCapacity": in.MaxCapacity,
		"Disabled":    in.Disabled,
	})

	if err != nil {
		return nil, model.NewAppError("SqlQueueSkillStore.Create", "store.sql_queue_skill.create.app_error", nil,
			fmt.Sprintf("name=%v, %v", in.QueueId, err.Error()), extractCodeFromErr(err))
	}

	return qs, nil
}

func (s SqlQueueSkillStore) Get(domainId int64, queueId, id uint32) (*model.QueueSkill, *model.AppError) {
	var qs *model.QueueSkill

	err := s.GetReplica().SelectOne(&qs, `select "id", "skill", "buckets", "lvl", "min_capacity", "max_capacity", "disabled"
		from cc_queue_skill_list
		where id = :Id and queue_id = :QueueId and domain_id = :DomainId
	`, map[string]interface{}{
		"Id":       id,
		"DomainId": domainId,
		"QueueId":  queueId,
	})

	if err != nil {
		return nil, model.NewAppError("SqlQueueSkillStore.Get", "store.sql_queue_skill.get.app_error", nil,
			fmt.Sprintf("name=%v, %v", queueId, err.Error()), extractCodeFromErr(err))
	}

	return qs, nil
}

func (s SqlQueueSkillStore) GetAllPage(domainId int64, search *model.SearchQueueSkill) ([]*model.QueueSkill, *model.AppError) {
	var qs []*model.QueueSkill

	f := map[string]interface{}{
		"DomainId":    domainId,
		"QueueId":     search.QueueId,
		"Ids":         pq.Array(search.Ids),
		"SkillIds":    pq.Array(search.SkillIds),
		"BucketIds":   pq.Array(search.BucketIds),
		"Lvl":         pq.Array(search.Lvl),
		"MinCapacity": pq.Array(search.MinCapacity),
		"MaxCapacity": pq.Array(search.MaxCapacity),
		"Disabled":    search.Disabled,
	}

	err := s.ListQuery(&qs, search.ListRequest,
		`queue_id = :QueueId and domain_id = :DomainId
				and (:Ids::int4[] isnull or id = any(:Ids))
				and (:SkillIds::int4[] isnull or skill_id = any(:SkillIds))
				and (:BucketIds::int4[] isnull or bucket_ids && :BucketIds)
				and (:Lvl::int4[] isnull or lvl = any(:Lvl))
				and (:MinCapacity::int4[] isnull or min_capacity = any(:MinCapacity))
				and (:MaxCapacity::int4[] isnull or max_capacity = any(:MaxCapacity))
				and (:Disabled::bool isnull or disabled = :Disabled)
			`,
		model.QueueSkill{}, f)
	if err != nil {
		return nil, model.NewAppError("SqlQueueSkillStore.GetAllPage", "store.sql_queue_skill.get_all.app_error", nil, err.Error(), extractCodeFromErr(err))
	}

	return qs, nil
}

func (s SqlQueueSkillStore) Update(domainId int64, skill *model.QueueSkill) (*model.QueueSkill, *model.AppError) {
	var qs *model.QueueSkill
	err := s.GetMaster().SelectOne(&qs, `with s as (
    update cc_queue_skill s
    set skill_id = :SkillId,
        bucket_ids = :BucketIds,
        lvl = :Lvl,
        min_capacity = :MinCapacity,
        max_capacity = :MaxCapacity,
        disabled = :Disabled
    where s.id = :Id and exists(select 1 from cc_queue q where q.id = s.queue_id and q.domain_id = :DomainId)
    returning *
)
select s.id,
       cc_get_lookup(cs.id, cs.name) skill,
       (select jsonb_agg(cc_get_lookup(b.id, b.name::varchar))
        from cc_bucket b
        where b.id = any (s.bucket_ids)
       )                             buckets,
       s.lvl,
       s.min_capacity,
       s.max_capacity,
       s.disabled
from s
         inner join cc_skill cs on s.skill_id = cs.id`, map[string]interface{}{
		"Id":          skill.Id,
		"DomainId":    domainId,
		"SkillId":     skill.Skill.Id,
		"BucketIds":   pq.Array(skill.BucketIds()),
		"Lvl":         skill.Lvl,
		"MinCapacity": skill.MinCapacity,
		"MaxCapacity": skill.MaxCapacity,
		"Disabled":    skill.Disabled,
	})

	if err != nil {
		return nil, model.NewAppError("SqlQueueSkillStore.Update", "store.sql_queue_skill.update.app_error", nil, err.Error(), extractCodeFromErr(err))
	}

	return qs, nil
}

func (s SqlQueueSkillStore) Delete(domainId int64, queueId, id uint32) *model.AppError {
	if _, err := s.GetMaster().Exec(`delete from cc_queue_skill s
where s.id = :Id and s.queue_id = :QueueId and exists(select 1 from cc_queue q where q.id = s.queue_id and q.domain_id = :DomainId)`,
		map[string]interface{}{
			"Id":       id,
			"DomainId": domainId,
			"QueueId":  queueId,
		}); err != nil {
		return model.NewAppError("SqlQueueSkillStore.Delete", "store.sql_queue_skill.delete.app_error", nil,
			fmt.Sprintf("Id=%v, %s", id, err.Error()), extractCodeFromErr(err))
	}
	return nil
}
