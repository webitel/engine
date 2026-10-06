package sqlstore

import (
	"context"

	"github.com/lib/pq"

	"github.com/webitel/engine/model"
	"github.com/webitel/engine/store"
)

const userNotificationFilter = `t.domain_id = :DomainId
	and t.user_id = :UserId
	and t.created_at > now() - make_interval(days => coalesce((select ss.value::int
		from call_center.system_settings ss
		where ss.domain_id = :DomainId and ss.name = 'message_ttl'), 30))
	and (:Q::varchar isnull or t.message ilike :Q::varchar)
	and (:From::timestamptz isnull or t.created_at >= :From::timestamptz)
	and (:To::timestamptz isnull or t.created_at <= :To::timestamptz)
	and (:Read::bool isnull or (t.read_at notnull) = :Read::bool)`

type SqlUserNotificationStore struct {
	SqlStore
}

func NewSqlUserNotificationStore(sqlStore SqlStore) store.UserNotificationStore {
	us := &SqlUserNotificationStore{sqlStore}

	return us
}

func userNotificationFilterParams(domainId, userId int64, search *model.SearchUserNotification) map[string]any {
	return map[string]any{
		"DomainId": domainId,
		"UserId":   userId,
		"Q":        search.GetQ(),
		"From":     model.GetBetweenFromTime(search.CreatedAt),
		"To":       model.GetBetweenToTime(search.CreatedAt),
		"Read":     search.Read,
	}
}

func (s SqlUserNotificationStore) GetAllPage(ctx context.Context, domainId, userId int64, search *model.SearchUserNotification) ([]*model.UserNotification, model.AppError) {
	var list []*model.UserNotification

	err := s.ListQuery(ctx, &list, search.ListRequest, userNotificationFilter, model.UserNotification{},
		userNotificationFilterParams(domainId, userId, search))
	if err != nil {
		return nil, model.NewCustomCodeError("store.sql_user_notification.get_all.app_error", err.Error(), extractCodeFromErr(err))
	}

	return list, nil
}

func (s SqlUserNotificationStore) Count(ctx context.Context, domainId, userId int64, search *model.SearchUserNotification) (int64, model.AppError) {
	count, err := s.GetReplica().WithContext(ctx).SelectInt(`select count(*)
from call_center.cc_user_notification t
where `+userNotificationFilter, userNotificationFilterParams(domainId, userId, search))
	if err != nil {
		return 0, model.NewCustomCodeError("store.sql_user_notification.count.app_error", err.Error(), extractCodeFromErr(err))
	}

	return count, nil
}

func (s SqlUserNotificationStore) MarkRead(ctx context.Context, domainId, userId int64, ids []int64, all bool) (int64, model.AppError) {
	res, err := s.GetMaster().WithContext(ctx).Exec(`update call_center.cc_user_notification t
set read_at = now()
where t.domain_id = :DomainId
	and t.user_id = :UserId
	and t.read_at isnull
	and (:All::bool or t.id = any(:Ids::int8[]))`, map[string]any{
		"DomainId": domainId,
		"UserId":   userId,
		"Ids":      pq.Array(ids),
		"All":      all,
	})
	if err != nil {
		return 0, model.NewCustomCodeError("store.sql_user_notification.mark_read.app_error", err.Error(), extractCodeFromErr(err))
	}

	cnt, err := res.RowsAffected()
	if err != nil {
		return 0, model.NewCustomCodeError("store.sql_user_notification.mark_read.app_error", err.Error(), extractCodeFromErr(err))
	}

	return cnt, nil
}
