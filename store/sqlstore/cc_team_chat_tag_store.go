package sqlstore

import (
	"context"
	"fmt"

	"github.com/lib/pq"
	"github.com/webitel/engine/model"
	"github.com/webitel/engine/store"
)

type SqlTeamChatTagStore struct {
	SqlStore
}

func NewSqlTeamChatTagStore(sqlStore SqlStore) store.TeamChatTagStore {
	us := &SqlTeamChatTagStore{sqlStore}
	return us
}

func (s SqlTeamChatTagStore) Create(ctx context.Context, domainId int64, teamId int64, in *model.TeamChatTag) (*model.TeamChatTag, model.AppError) {
	var tct *model.TeamChatTag

	err := s.GetMaster().WithContext(ctx).SelectOne(&tct, `with tct as (
    insert into call_center.cc_team_chat_tag (tag, team_id, updated_by, updated_at, created_by, created_at)
    select :Tag, :TeamId, :UpdatedBy, :UpdatedAt, :CreatedBy, :CreatedAt
    where exists (select 1 from call_center.cc_team a where a.domain_id = :DomainId and a.id = :TeamId)
    returning *
)select tct.id,
       tct.tag,
       call_center.cc_get_lookup(uc.id, coalesce(uc.name, uc.username)) "created_by",
       tct.created_at,
       call_center.cc_get_lookup(uu.id, coalesce(uu.name, uu.username)) "updated_by",
       tct.updated_at
from tct
    left join directory.wbt_user uc on uc.id = tct.created_by
    left join directory.wbt_user uu on uu.id = tct.updated_by`, map[string]interface{}{
		"DomainId":  domainId,
		"Tag":       in.Tag,
		"TeamId":    teamId,
		"UpdatedBy": in.UpdatedBy.GetSafeId(),
		"UpdatedAt": in.UpdatedAt,
		"CreatedBy": in.CreatedBy.GetSafeId(),
		"CreatedAt": in.CreatedAt,
	})

	if err != nil {
		return nil, model.NewCustomCodeError("store.sql_team_chat_tag.create.app_error", fmt.Sprintf("tag=%v, %v", in.Tag, messageFromErr(err)), extractCodeFromErr(err))
	}

	return tct, nil
}

func (s SqlTeamChatTagStore) Get(ctx context.Context, domainId int64, teamId int64, id uint32) (*model.TeamChatTag, model.AppError) {
	var tct *model.TeamChatTag

	err := s.GetReplica().WithContext(ctx).SelectOne(&tct, `select tct.id,
       tct.tag,
       call_center.cc_get_lookup(uc.id, coalesce(uc.name, uc.username)) "created_by",
       tct.created_at,
       call_center.cc_get_lookup(uu.id, coalesce(uu.name, uu.username)) "updated_by",
       tct.updated_at
from call_center.cc_team_chat_tag tct
    left join directory.wbt_user uc on uc.id = tct.created_by
    left join directory.wbt_user uu on uu.id = tct.updated_by
where tct.id = :Id
    and tct.team_id = :TeamId
    and exists (select 1 from call_center.cc_team q where q.id = tct.team_id and q.domain_id = :DomainId)`, map[string]interface{}{
		"TeamId":   teamId,
		"Id":       id,
		"DomainId": domainId,
	})

	if err != nil {
		return nil, model.NewCustomCodeError("store.sql_team_chat_tag.get.app_error", fmt.Sprintf("Id=%v, %v", id, err.Error()), extractCodeFromErr(err))
	}

	return tct, nil
}

func (s SqlTeamChatTagStore) GetAllPage(ctx context.Context, domainId int64, teamId int64, search *model.SearchTeamChatTag) ([]*model.TeamChatTag, model.AppError) {
	var list []*model.TeamChatTag

	f := map[string]interface{}{
		"DomainId": domainId,
		"TeamId":   teamId,
		"Q":        search.GetQ(),
		"Ids":      pq.Array(search.Ids),
	}

	err := s.ListQuery(ctx, &list, search.ListRequest,
		` team_id = :TeamId::int8
                and exists (select 1 from call_center.cc_team q where q.id = team_id and q.domain_id = :DomainId)
				and (:Q::text isnull or tag ilike :Q::varchar)
				and (:Ids::int4[] isnull or id = any(:Ids))
			`,
		model.TeamChatTag{}, f)
	if err != nil {
		return nil, model.NewCustomCodeError("store.sql_team_chat_tag.get_all.app_error", err.Error(), extractCodeFromErr(err))
	}

	return list, nil
}

func (s SqlTeamChatTagStore) GetAllPageByUser(ctx context.Context, domainId int64, userId int64, search *model.SearchTeamChatTag) ([]*model.TeamChatTag, model.AppError) {
	var list []*model.TeamChatTag

	f := map[string]interface{}{
		"DomainId": domainId,
		"UserId":   userId,
		"Q":        search.GetQ(),
		"Ids":      pq.Array(search.Ids),
		"Limit":    search.GetLimit(),
		"Offset":   search.GetOffset(),
	}

	_, err := s.GetReplica().WithContext(ctx).Select(&list, `select tct.id,
		   tct.tag,
		   call_center.cc_get_lookup(uc.id, coalesce(uc.name, uc.username)) "created_by",
		   tct.created_at,
		   call_center.cc_get_lookup(uu.id, coalesce(uu.name, uu.username)) "updated_by",
		   tct.updated_at
	from call_center.cc_team_chat_tag tct
		left join directory.wbt_user uc on uc.id = tct.created_by
		left join directory.wbt_user uu on uu.id = tct.updated_by
	where
		tct.team_id = any(select a.team_id from call_center.cc_agent a where a.user_id = :UserId)
		and exists (select 1 from call_center.cc_team q where q.id = tct.team_id and q.domain_id = :DomainId)
		and (:Q::text isnull or tag ilike :Q::varchar)
		and (:Ids::int4[] isnull or id = any(:Ids))
	order by tct.tag
	limit :Limit
	offset :Offset`, f)
	if err != nil {
		return nil, model.NewCustomCodeError("store.sql_team_chat_tag.get_all_page_by_user.execute.error", err.Error(), extractCodeFromErr(err))
	}

	return list, nil
}

func (s SqlTeamChatTagStore) Update(ctx context.Context, domainId int64, teamId int64, tct *model.TeamChatTag) (*model.TeamChatTag, model.AppError) {

	err := s.GetMaster().WithContext(ctx).SelectOne(&tct, `with tct as (
    update call_center.cc_team_chat_tag
    set tag = :Tag,
        updated_by = :UpdatedBy,
        updated_at = :UpdatedAt
    where id = :Id
		and team_id = :TeamId
        and exists(select 1 from call_center.cc_team q where q.id = team_id and q.domain_id = :DomainId)
    returning *
)
select tct.id,
       tct.tag,
       call_center.cc_get_lookup(uc.id, coalesce(uc.name, uc.username)) "created_by",
       tct.created_at,
       call_center.cc_get_lookup(uu.id, coalesce(uu.name, uu.username)) "updated_by",
       tct.updated_at
from tct
    left join directory.wbt_user uc on uc.id = tct.created_by
    left join directory.wbt_user uu on uu.id = tct.updated_by`, map[string]interface{}{
		"Id":        tct.Id,
		"Tag":       tct.Tag,
		"UpdatedBy": tct.UpdatedBy.GetSafeId(),
		"UpdatedAt": tct.UpdatedAt,
		"TeamId":    teamId,
		"DomainId":  domainId,
	})

	if err != nil {
		return nil, model.NewCustomCodeError("store.sql_team_chat_tag.update.app_error", err.Error(), extractCodeFromErr(err))
	}

	return tct, nil
}

func (s SqlTeamChatTagStore) Delete(ctx context.Context, domainId int64, teamId int64, id uint32) model.AppError {
	if _, err := s.GetMaster().WithContext(ctx).Exec(`delete from call_center.cc_team_chat_tag tct where tct.id=:Id and tct.team_id = :TeamId
			and exists(select 1 from call_center.cc_team q where q.id = :TeamId and q.domain_id = :DomainId)`,
		map[string]interface{}{"Id": id, "DomainId": domainId, "TeamId": teamId}); err != nil {
		return model.NewCustomCodeError("store.sql_team_chat_tag.delete.app_error", fmt.Sprintf("Id=%v, %s", id, err.Error()), extractCodeFromErr(err))
	}
	return nil
}
