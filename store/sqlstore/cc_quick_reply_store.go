package sqlstore

import (
	"context"
	"fmt"
	"strings"

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

	args := map[string]any{
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
	args := map[string]any{
		"DomainId": domainId,
		"Q":        search.GetQ(),
		"Ids":      pq.Array(search.Ids),
		"Name":     search.Name,
		"UserId":   userId,
		"Queue":    pq.Array(search.Queue),
		"RestrictToAgent": search.RestrictToAgent,
	}

	where := `
		domain_id = :DomainId
		AND (:Q::varchar isnull OR t.name ILIKE :Q::varchar)
		AND (:Ids::int8[] isnull OR t.id = ANY(:Ids::bigint[]))
		and (
			:RestrictToAgent = false 
    		or (
				(t.team_ids is null and t.queue_ids is null)
        		or (
        		    select ca.team_id
        		    from call_center.cc_agent ca
        		    where ca.user_id = :UserId and ca.domain_id = :DomainId
        		    limit 1
        		) = any (t.team_ids)
        		or (
        		    call_center.cc_get_agent_queues(:DomainId, :UserId) && t.queue_ids
        		)
    		)
		)
	`
	
	var replies []*model.QuickReply
	if err := s.ListQuery(ctx, &replies, search.ListRequest, where, model.QuickReply{}, args); err != nil {
		return nil, model.NewCustomCodeError("store.sql_quick_reply.get_all.app_error", err.Error(), extractCodeFromErr(err))
	}
	return replies, nil
}

func (s SqlQuickReplyStore) GetAllPageByAgentPriority(ctx context.Context, domainId, userId int64, search *model.SearchQuickReply) ([]*model.QuickReply, model.AppError) {
	args := map[string]any {
			"DomainId": domainId,
			"UserId": userId,
			"Q": search.GetQ(),
			"Ids": pq.Array(search.Ids),
			"Name": search.Name,
			"Queue": pq.Array(search.Queue),
			"RestrictToAgent": search.RestrictToAgent,
	}
	fields := GetFields(search.Fields, &model.QuickReply{})
	parsedFields := strings.Join(fields, ", ")
	
	query := fmt.Sprintf(`
		with agent_info_cte as (
			select (
				select ca.team_id
				from call_center.cc_agent ca
				where ca.user_id = :UserId
				and ca.domain_id = :DomainId
				limit 1
			) as team_id,
			call_center.cc_get_agent_queues(:DomainId, :UserId) as queues_ids
		),
		filtered_quick_replies_cte as (
			select
				%s,
				case 
					when t.queue_ids && agent_info.queues_ids
					then 1
					when agent_info.team_id = any(t.team_ids)
					then 2
					else 3
				end as agent_priority
			from call_center.cc_quick_reply_list t, agent_info_cte agent_info
			where t.domain_id = :DomainId
				and (:Q::varchar is null or t.name ilike :Q::varchar)
				and (:Ids::int8[] is null or t.id = any(:Ids::bigint[]))
				and (
					:RestrictToAgent = false
					or (
						(t.team_ids is null and t.queue_ids is null)
						or agent_info.team_id = any (t.team_ids)
						or agent_info.queues_ids && t.queue_ids
					)
				)
		)
		select
			%s
		from filtered_quick_replies_cte t
		%s
		offset %d
		limit %d
	`, parsedFields, parsedFields, GetOrderBy(model.QuickReply{}.EntityName(), search.Sort), search.ListRequest.GetOffset(), search.ListRequest.GetLimit())

	var quickRepliesList []*model.QuickReply
	if _, err := s.GetReplica().WithContext(ctx).Select(&quickRepliesList, query, args); err != nil {
		return nil, model.NewCustomCodeError("store.sql_quick_reply.get_all_by_agent_priority.app_error", err.Error(), extractCodeFromErr(err))
	}

	return quickRepliesList, nil
}

func (s SqlQuickReplyStore) Get(ctx context.Context, domainId int64, id uint32) (*model.QuickReply, model.AppError) {
	var reply *model.QuickReply

	args := map[string]any{
		"DomainId": domainId,
		"Id":       id,
	}

	if err := s.One(ctx, &reply, `domain_id = :DomainId and id = :Id`, model.QuickReply{}, args); err != nil {
		return nil, model.NewCustomCodeError("store.sql_quick_reply.get.app_error", err.Error(), extractCodeFromErr(err))
	}

	return reply, nil
}

func (s SqlQuickReplyStore) Update(ctx context.Context, domainId int64, reply *model.QuickReply) (*model.QuickReply, model.AppError) {
	args := map[string]any{
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
		map[string]any{"Id": id, "DomainId": domainId}); err != nil {
		return model.NewCustomCodeError("store.sql_quick_reply.delete.app_error", fmt.Sprintf("Id=%v, %s", id, err.Error()), extractCodeFromErr(err))
	}

	return nil
}
