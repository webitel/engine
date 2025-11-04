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

func toInt64Slice(a any) []int64 {
	switch v := a.(type) {
	case nil:
		return []int64{}
	case []int64:
		return v
	case []int:
		out := make([]int64, len(v))
		for i, x := range v { out[i] = int64(x) }
		return out
	case int64:
		return []int64{v}
	case int:
		return []int64{int64(v)}
	default:
		return []int64{}
	}
}

func toInt32Slice(a any) []int32 {
	switch v := a.(type) {
	case nil:
		return []int32{}
	case []int32:
		return v
	case []int:
		out := make([]int32, len(v))
		for i, x := range v { out[i] = int32(x) }
		return out
	case []int64:
		out := make([]int32, len(v))
		for i, x := range v { out[i] = int32(x) }
		return out
	case int32:
		return []int32{v}
	case int64:
		return []int32{int32(v)}
	case int:
		return []int32{int32(v)}
	default:
		return []int32{}
	}
}


func (s SqlQuickReplyStore) GetAllPage(ctx context.Context, domainId int64, search *model.SearchQuickReply, userId int64) ([]*model.QuickReply, model.AppError) {
	args := map[string]interface{}{
		"DomainId": domainId,
		"Q":        search.GetQ(),
		"Ids":      pq.Array(toInt64Slice(search.Ids)),   // -> :Ids::bigint[]
		"Name":     search.Name,
		"UserId":   userId,
		"Queue":    pq.Array(toInt32Slice(search.Queue)), // -> :Queue::int4[]
		"RestrictToAgent": search.RestrictToAgent,
	}

	if len(search.Ids) > 0 {
    args["Ids"] = pq.Array(search.Ids) // []int64
	} else {
		args["Ids"] = nil                  // => NULL у SQL
	}

	if len(search.Queue) > 0 {
		args["Queue"] = pq.Array(search.Queue) // []int32
	} else {
		args["Queue"] = nil                    // теж можна NULL
	}

	where := `
		domain_id = :DomainId
		AND (:Q::varchar isnull OR t.name ILIKE :Q::varchar)
		AND (:Ids::int8[] isnull OR t.id = ANY(:Ids::bigint[]))
		and (
			:RestrictToAgent = false 
    		or (
        		(
        		    t.team_ids is null or (
        		        select ca.team_id
        		        from call_center.cc_agent ca
        		        where ca.user_id = :UserId and ca.domain_id = :DomainId
        		        limit 1
        		    ) = any (t.team_ids)
        		)
        		or (
        		    t.queue_ids is null
        		    or call_center.cc_get_agent_queues(:DomainId, :UserId) && t.queue_ids
        		)
    		)
		)
	`

	if search.ListRequest.Sort == "" {
		search.ListRequest.Sort = "+sort_priority"
	} 
	
	var replies []*model.QuickReply
	if err := s.ListQuery(ctx, &replies, search.ListRequest, where, model.QuickReply{}, args); err != nil {
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
