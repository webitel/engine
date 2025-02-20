package sqlstore

import (
	"context"
	"fmt"

	"github.com/lib/pq"
	"github.com/webitel/engine/model"
	"github.com/webitel/engine/store"
)

type SqlQuickReplyStore struct {
	SqlStore
}

func NewSqlQuickReplyStore(sqlStore SqlStore) store.QuickReplyStore {
	us := &SqlQuickReplyStore{sqlStore}
	return us
}

func (s SqlQuickReplyStore) Create(ctx context.Context, domainId int64, reply *model.QuickReply) (*model.QuickReply, model.AppError) {
	resp := &model.QuickReply{}
	err := s.GetMaster().WithContext(ctx).SelectOne(&resp, `with s as (
    insert into call_center.cc_quick_reply (domain_id, created_at, updated_at, created_by, updated_by,
                                      name, text, article, team, queue)
    values (:DomainId, :CreatedAt, :UpdatedAt, :CreatedBy, :UpdatedBy,
            :Name, :Text, :Article, :Team, :Queue)
    returning *
)
select s.id,
       s.created_at,
       call_center.cc_get_lookup(uc.id, coalesce(uc.name, uc.username)) as created_by,
       s.updated_at,
       call_center.cc_get_lookup(uc.id, coalesce(uc.name, uc.username)) as updated_by,
       s.name,
       s.text
from s
         left join directory.wbt_user uc on uc.id = s.created_by
         left join directory.wbt_user uu on uu.id = s.updated_by`, map[string]interface{}{
		"DomainId":  domainId,
		"CreatedAt": reply.CreatedAt,
		"UpdatedAt": reply.UpdatedAt,
		"CreatedBy": reply.CreatedBy.GetSafeId(),
		"UpdatedBy": reply.UpdatedBy.GetSafeId(),
		"Name":      reply.Name,
		"Text":      reply.Text,
		"Article":   reply.Article.Id,
		"Team":      reply.Team.GetSafeId(),
		"Queue":     reply.Queue.GetSafeId(),
	})

	if err != nil {
		return nil, model.NewCustomCodeError("store.sql_quick_reply.create.app_error", fmt.Sprintf("name=%v, %v", reply.Name, err.Error()), extractCodeFromErr(err))
	}
	resp, err = s.Get(ctx, domainId, uint32(resp.Id))
	return resp, nil
}

func (s SqlQuickReplyStore) GetAllPage(ctx context.Context, domainId int64, search *model.SearchQuickReply) ([]*model.QuickReply, model.AppError) {
	var replies []*model.QuickReply

	f := map[string]interface{}{
		"DomainId": domainId,
		"Q":        search.GetQ(),
		"Ids":      pq.Array(search.Ids),
		"Name":     search.Name,
	}

	err := s.ListQuery(ctx, &replies, search.ListRequest,
		`domain_id = :DomainId
				and (:Q::varchar isnull or (name ilike :Q::varchar or text ilike :Q::varchar))
				and (:Ids::int4[] isnull or id = any(:Ids))
			`,
		model.QuickReply{}, f)
	if err != nil {
		return nil, model.NewCustomCodeError("store.sql_quick_reply.get_all.app_error", err.Error(), extractCodeFromErr(err))
	}

	return replies, nil
}

func (s SqlQuickReplyStore) Get(ctx context.Context, domainId int64, id uint32) (*model.QuickReply, model.AppError) {
	var reply *model.QuickReply
	err := s.GetReplica().WithContext(ctx).SelectOne(&reply, `select q.id, 
       q.created_at,
       q.updated_at,
	   call_center.cc_get_lookup(uc.id, uc.name)         as created_by,
       call_center.cc_get_lookup(u.id, u.name)           as updated_by,
       q.name,
       q.text,
	   call_center.cc_get_lookup(a.id, a.title)         as article,
	   call_center.cc_get_lookup(t.id, t.name)         as team,
	   call_center.cc_get_lookup(cq.id, cq.name)         as queue
from call_center.cc_quick_reply q
	left join directory.wbt_user uc on uc.id = q.created_by
	left join directory.wbt_user u on u.id = q.updated_by
	left join call_center.cc_team t on t.id = q.team
	left join knowledge_base.article a on a.id = q.article
	left join call_center.cc_queue cq on cq.id = q.queue
where q.id = :Id and q.domain_id = :DomainId`, map[string]interface{}{
		"DomainId": domainId,
		"Id":       id,
	})

	if err != nil {
		return nil, model.NewCustomCodeError("store.sql_quick_reply.get.app_error", err.Error(), extractCodeFromErr(err))
	}

	return reply, nil
}

func (s SqlQuickReplyStore) Update(ctx context.Context, domainId int64, reply *model.QuickReply) (*model.QuickReply, model.AppError) {
	err := s.GetMaster().WithContext(ctx).SelectOne(&reply, `with s as (
    update call_center.cc_quick_reply
        set updated_at = :UpdatedAt,
            updated_by = :UpdatedBy,
            name = :Name,
            text = :Text,
            article = :Article,
			team = :Team,
			queue = :Queue
        where id = :Id and domain_id = :DomainId
    returning *
)
select s.id,
       s.created_at,
       call_center.cc_get_lookup(uc.id, coalesce(uc.name, uc.username)) as created_by,
       s.updated_at,
       call_center.cc_get_lookup(uc.id, coalesce(uc.name, uc.username)) as updated_by,
       s.name,
       s.text,
	   call_center.cc_get_lookup(a.id, a.title)         as article,
	   call_center.cc_get_lookup(t.id, t.name)         as team,
	   call_center.cc_get_lookup(cq.id, cq.name)         as queue
from s
         left join knowledge_base.article a on a.id = s.article
		 left join call_center.cc_team t on t.id = s.team
		 left join call_center.cc_queue cq on cq.id = s.queue
		 left join directory.wbt_user uc on uc.id = s.created_by
         left join directory.wbt_user uu on uu.id = s.updated_by;`, map[string]interface{}{
		"DomainId":  domainId,
		"Id":        reply.Id,
		"Name":      reply.Name,
		"Text":      reply.Text,
		"UpdatedAt": reply.UpdatedAt,
		"UpdatedBy": reply.UpdatedBy.GetSafeId(),
		"Article":   reply.Article.GetSafeId(),
		"Team":      reply.Team.GetSafeId(),
		"Queue":     reply.Queue.GetSafeId(),
	})

	if err != nil {
		return nil, model.NewCustomCodeError("store.sql_quick_reply.update.app_error", err.Error(), extractCodeFromErr(err))
	}

	return reply, nil
}

func (s SqlQuickReplyStore) Delete(ctx context.Context, domainId int64, id uint32) model.AppError {
	if _, err := s.GetMaster().WithContext(ctx).Exec(`delete from call_center.cc_quick_reply c where c.id=:Id and c.domain_id = :DomainId`,
		map[string]interface{}{"Id": id, "DomainId": domainId}); err != nil {
		return model.NewCustomCodeError("store.sql_quick_reply.delete.app_error", fmt.Sprintf("Id=%v, %s", id, err.Error()), extractCodeFromErr(err))
	}
	return nil
}
