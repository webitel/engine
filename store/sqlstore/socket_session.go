package sqlstore

import (
	"context"
	"github.com/webitel/engine/model"
	"github.com/webitel/engine/store"
	"time"
)

type SqlSocketSessionStore struct {
	SqlStore
}

func NewSqlSocketSessionStore(sqlStore SqlStore) store.SocketSessionStore {
	us := &SqlSocketSessionStore{sqlStore}
	return us
}

func (s *SqlSocketSessionStore) DeleteByApp(ctx context.Context, appId string) model.AppError {
	_, err := s.GetMaster().WithContext(ctx).Exec(`delete from call_center.socket_session
where app_id = :AppId;`, map[string]any{
		"AppId": appId,
	})

	if err != nil {
		return model.NewCustomCodeError("store.sql_socket.del_by_app.app_error", err.Error(), extractCodeFromErr(err))
	}

	return nil
}

func (s *SqlSocketSessionStore) DeleteById(ctx context.Context, id string) model.AppError {
	_, err := s.GetMaster().WithContext(ctx).Exec(`delete from call_center.socket_session
where id = :Id;`, map[string]any{
		"Id": id,
	})

	if err != nil {
		return model.NewCustomCodeError("store.sql_socket.del_by_id.app_error", err.Error(), extractCodeFromErr(err))
	}

	return nil
}

func (s *SqlSocketSessionStore) Create(ctx context.Context, session model.SocketSession) model.AppError {
	_, err := s.GetMaster().WithContext(ctx).Exec(`insert into call_center.socket_session (id, created_at, updated_at, user_agent, user_id, ip, client, app_id, domain_id)
values (:Id, :CreatedAt, :UpdatedAt, :UA, :UserId, :Ip, :Client, :AppId, :DomainId)`, map[string]any{
		"Id":        session.Id,
		"CreatedAt": session.CreatedAt,
		"UpdatedAt": session.UpdatedAt,
		"UA":        session.UserAgent,
		"UserId":    session.UserId,
		"Ip":        session.Ip,
		"Client":    session.Client,
		"AppId":     session.AppId,
		"DomainId":  session.DomainId,
	})

	if err != nil {
		return model.NewCustomCodeError("store.sql_socket.create.app_error", err.Error(), extractCodeFromErr(err))
	}

	return nil
}

func (s *SqlSocketSessionStore) SetUpdatedAt(ctx context.Context, id string, t time.Time) model.AppError {
	_, err := s.GetMaster().WithContext(ctx).Exec(`update call_center.socket_session
set updated_at = :UpdatedAt
where id = :Id`, map[string]any{
		"Id":        id,
		"UpdatedAt": t,
	})

	if err != nil {
		return model.NewCustomCodeError("store.sql_socket.updated_at.app_error", err.Error(), extractCodeFromErr(err))
	}

	return nil
}
