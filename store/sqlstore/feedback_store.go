package sqlstore

import (
	"context"
	"github.com/webitel/engine/model"
	"github.com/webitel/engine/store"
)

type SqlFeedbackStore struct {
	SqlStore
}

func NewSqlFeedbackStore(sqlStore SqlStore) store.FeedbackStore {
	us := &SqlFeedbackStore{sqlStore}
	return us
}

func (s *SqlFeedbackStore) Get(ctx context.Context, key model.FeedbackKey) (model.Feedback, model.AppError) {
	var f model.Feedback
	err := s.GetReplica().WithContext(ctx).SelectOne(&f, `select f.id,
       f.source_id,
       f.source,
       f.payload,
       call_center.cc_view_timestamp(f.created_at) created_at,
       f.rating,
       f.description
from call_center.feedback f
where f.domain_id = :DomainId
    and f.source = :Source
    and f.source_id = :SourceId`, map[string]any{
		"DomainId": key.DomainId,
		"Source":   key.Source,
		"SourceId": key.SourceId,
	})

	if err != nil {
		return model.Feedback{}, model.NewCustomCodeError("store.sql_feedback.get", err.Error(), extractCodeFromErr(err))
	}

	return f, nil
}

func (s *SqlFeedbackStore) Create(ctx context.Context, key model.FeedbackKey, rating float32, description string) (model.Feedback, model.AppError) {
	var f model.Feedback
	err := s.GetReplica().WithContext(ctx).SelectOne(&f, `with f as (
    insert into call_center.feedback (domain_id, source, source_id, rating, description, payload)
    values (:DomainId, :Source, :SourceId, :Rating, :Description, :Payload)
    returning *
)
select f.id,
       f.source_id,
       f.source,
       f.payload,
       call_center.cc_view_timestamp(f.created_at) created_at,
       f.rating,
	   f.description
from  f`, map[string]any{
		"DomainId":    key.DomainId,
		"Source":      key.Source,
		"SourceId":    key.SourceId,
		"Payload":     key.Payload.ToSafeBytes(),
		"Rating":      rating,
		"Description": description,
	})

	if err != nil {
		return model.Feedback{}, model.NewCustomCodeError("store.sql_feedback.create", err.Error(), extractCodeFromErr(err))
	}

	return f, nil
}
