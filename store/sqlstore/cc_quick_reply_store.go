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
	var resp *model.QuickReply

	args := map[string]interface{}{
		"DomainId":  domainId,
		"CreatedAt": reply.CreatedAt,
		"UpdatedAt": reply.UpdatedAt,
		"CreatedBy": reply.CreatedBy.GetSafeId(),
		"UpdatedBy": reply.UpdatedBy.GetSafeId(),
		"Name":      reply.Name,
		"Text":      reply.Text,
		"Article":   reply.Article.GetSafeId(),
		"Teams":     pq.Array(model.LookupIds(reply.Teams)),
		"Queues":    pq.Array(model.LookupIds(reply.Queues)),
	}

	query := `with s as (
    insert into call_center.cc_quick_reply (domain_id, created_at, updated_at, created_by, updated_by,
                                      name, text, article, teams, queues)
    values (:DomainId, :CreatedAt, :UpdatedAt, :CreatedBy, :UpdatedBy,
            :Name, :Text, :Article, :Teams, :Queues)
    returning *
	)

	select s.id
       	, s.created_at
       	, call_center.cc_get_lookup(uc.id, coalesce(uc.name, uc.username)) as created_by
       	, s.updated_at
       	, call_center.cc_get_lookup(uc.id, coalesce(uc.name, uc.username)) as updated_by
		, s.name
       	, s.text
		, ( SELECT jsonb_agg(call_center.cc_get_lookup(t.id, t.name)) AS jsonb_agg
           	FROM call_center.cc_team t
          	WHERE t.id = ANY (s.teams)) AS teams
	   	, ( SELECT jsonb_agg(call_center.cc_get_lookup(a.id::bigint, a.name)) AS jsonb_agg
			FROM call_center.cc_queue a
          	WHERE a.id = ANY (s.queues)) AS queues
	from s
		left join directory.wbt_user uc on uc.id = s.created_by
		left join directory.wbt_user uu on uu.id = s.updated_by`

	if err := s.GetMaster().WithContext(ctx).SelectOne(&resp, query, args); err != nil {
		return nil, model.NewCustomCodeError("store.sql_quick_reply.create.app_error", fmt.Sprintf("name=%v, %v", reply.Name, err.Error()), extractCodeFromErr(err))
	}

	return resp, nil
}

func (s SqlQuickReplyStore) GetAllPage(ctx context.Context, domainId int64, search *model.SearchQuickReply, userId int64) ([]*model.QuickReply, model.AppError) {
	var replies []*model.QuickReply

	args := map[string]interface{}{
		"DomainId": domainId,
		"Q":        search.GetQ(),
		"Ids":      pq.Array(search.Ids),
		"Name":     search.Name,
		"UserId":   userId,
	}

	err := s.ListQuery(ctx, &replies, search.ListRequest,
		`domain_id = :DomainId
				and (:Q::varchar isnull or (name ilike :Q::varchar))
				and (:Ids::int4[] isnull or id = any(:Ids))
				and (t.team_ids && (
						SELECT array_agg(ca.team_id)::bigint[]
						FROM call_center.cc_agent ca
						WHERE ca.user_id = :UserId
					)
					OR t.team_ids IS NULL
					OR t.team_ids = '{}'
					)
			`,
		model.QuickReply{}, args)
	if err != nil {
		return nil, model.NewCustomCodeError("store.sql_quick_reply.get_all.app_error", err.Error(), extractCodeFromErr(err))
	}

	return replies, nil
}

func (s SqlQuickReplyStore) Get(ctx context.Context, domainId int64, id uint32) (*model.QuickReply, model.AppError) {
	var reply *model.QuickReply

	args := map[string]interface{}{
		"DomainId": domainId,
		"Id":       id,
	}

	if err := s.One(ctx, &reply, `domain_id = :DomainId and id = :Id`, model.QuickReply{}, args); err != nil {
		return nil, model.NewCustomCodeError("store.sql_quick_reply.get.app_error", err.Error(), extractCodeFromErr(err))
	}

	return reply, nil
}

func (s SqlQuickReplyStore) Update(ctx context.Context, domainId int64, reply *model.QuickReply) (*model.QuickReply, model.AppError) {
	args := map[string]interface{}{
		"DomainId":  domainId,
		"Id":        reply.Id,
		"Name":      reply.Name,
		"Text":      reply.Text,
		"UpdatedAt": reply.UpdatedAt,
		"UpdatedBy": reply.UpdatedBy.GetSafeId(),
		"Article":   reply.Article.GetSafeId(),
		"Teams":     pq.Array(model.LookupIds(reply.Teams)),
		"Queues":    pq.Array(model.LookupIds(reply.Queues)),
	}

	query := `with s as (
		update call_center.cc_quick_reply
			set updated_at = :UpdatedAt,
				updated_by = :UpdatedBy,
				name = :Name,
				text = :Text,
				article = :Article,
				teams = :Teams,
				queues = :Queues
			where id = :Id and domain_id = :DomainId
		returning *
	)
	
	select s.id
       	, s.created_at
       	, call_center.cc_get_lookup(uc.id, coalesce(uc.name, uc.username)) as created_by
       	, s.updated_at
       	, call_center.cc_get_lookup(uc.id, coalesce(uc.name, uc.username)) as updated_by
		, s.name
       	, s.text
		, ( SELECT jsonb_agg(call_center.cc_get_lookup(t.id, t.name)) AS jsonb_agg
           	FROM call_center.cc_team t
          	WHERE t.id = ANY (s.teams)) AS teams
	   	, ( SELECT jsonb_agg(call_center.cc_get_lookup(a.id::bigint, a.name)) AS jsonb_agg
			FROM call_center.cc_queue a
          	WHERE a.id = ANY (s.queues)) AS queues
	from s
		left join directory.wbt_user uc on uc.id = s.created_by
		left join directory.wbt_user uu on uu.id = s.updated_by`

	if err := s.GetMaster().WithContext(ctx).SelectOne(&reply, query, args); err != nil {
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
