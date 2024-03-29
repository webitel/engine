package sqlstore

import (
	"context"
	"fmt"

	"github.com/lib/pq"
	"github.com/webitel/engine/model"
	"github.com/webitel/engine/store"
)

type SqlNotificationStore struct {
	SqlStore
}

func NewSqlNotificationStore(sqlStore SqlStore) store.NotificationStore {
	us := &SqlNotificationStore{sqlStore}
	return us
}

func (s SqlNotificationStore) Create(ctx context.Context, notification *model.Notification) (*model.Notification, model.AppError) {
	err := s.GetMaster().WithContext(ctx).SelectOne(&notification, `insert into call_center.cc_notification (domain_id, action, created_by, timeout, for_users, description)
    values (:DomainId, :Action, :CreatedBy, :Timeout, :ForUsers, :Description)
    returning id, domain_id, action, call_center.cc_view_timestamp(created_at) as created_at, created_by, timeout, for_users, description`, map[string]interface{}{
		"DomainId":    notification.DomainId,
		"Timeout":     nil,
		"Action":      notification.Action,
		"CreatedBy":   notification.CreatedBy,
		"ForUsers":    pq.Array(notification.ForUsers),
		"Description": notification.Description,
	})

	if err != nil {
		return nil, model.NewCustomCodeError("store.sql_notification.save.app_error", fmt.Sprintf("name=%v, %v", notification.Action, err.Error()), extractCodeFromErr(err))
	}

	return notification, nil
}

func (s SqlNotificationStore) Close(ctx context.Context, id, userId int64) (*model.Notification, model.AppError) {
	panic("implement me")
}

func (s SqlNotificationStore) Accept(ctx context.Context, id, userId int64) (*model.Notification, model.AppError) {
	panic("implement me")
}
