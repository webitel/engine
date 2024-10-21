package sqlstore

import (
	"context"

	"github.com/lib/pq"
	"github.com/webitel/engine/model"
	"github.com/webitel/engine/store"
)

type SqlAuditRateStore struct {
	SqlStore
}

func (s SqlAuditRateStore) Create(ctx context.Context, domainId int64, rate *model.AuditRate) (*model.AuditRate, model.AppError) {
	err := s.GetMaster().WithContext(ctx).SelectOne(&rate, `with r as (
    insert into call_center.cc_audit_rate (domain_id, form_id, created_at, created_by, updated_at, updated_by, answers, score_required, score_optional, 
		comment, call_id, call_created_at, rated_user_id)
    values (:DomainId, :FormId, :CreatedAt, :CreatedBy, :UpdatedAt, :UpdatedBy, :Answers, :ScoreRequired, :ScoreOptional, 
		:Comment, :CallId, :CallCreatedAt, :RatedUserId)
    returning *
)
select r.id,
       r.created_at,
       call_center.cc_get_lookup(uc.id, coalesce(uc.name::character varying, uc.username::character varying)) AS created_by,
       r.updated_at,
       call_center.cc_get_lookup(u.id, coalesce(u.name::character varying, u.username::character varying))   AS updated_by,
       call_center.cc_get_lookup(ur.id, ur.name::character varying)   AS rated_user,
       call_center.cc_get_lookup(f.id, f.name::character varying)   AS form,
       r.answers,
       r.score_required,
       r.score_optional,
       r.comment,
       r.call_id,
       f.questions
from  r
    left join call_center.cc_audit_form f on f.id = r.form_id
    LEFT JOIN directory.wbt_user uc ON uc.id = r.created_by
    LEFT JOIN directory.wbt_user u ON u.id = r.updated_by
    LEFT JOIN directory.wbt_user ur ON ur.id = r.rated_user_id`, map[string]interface{}{
		"DomainId":      domainId,
		"FormId":        rate.Form.Id,
		"CreatedAt":     rate.CreatedAt,
		"CreatedBy":     rate.CreatedBy.GetSafeId(),
		"UpdatedAt":     rate.UpdatedAt,
		"UpdatedBy":     rate.UpdatedBy.GetSafeId(),
		"Answers":       rate.Answers.ToJson(),
		"ScoreRequired": rate.ScoreRequired,
		"ScoreOptional": rate.ScoreOptional,
		"Comment":       rate.Comment,
		"CallId":        rate.CallId,
		"CallCreatedAt": rate.CallCreatedAt,
		"RatedUserId":   rate.RatedUser.GetSafeId(),
	})

	if err != nil {
		return nil, model.NewInternalError("store.sql_audit_rate.save.app_error", err.Error())
	}

	return rate, nil
}

func (s SqlAuditRateStore) GetAllPage(ctx context.Context, domainId int64, search *model.SearchAuditRate) ([]*model.AuditRate, model.AppError) {
	var list []*model.AuditRate

	f := map[string]interface{}{
		"DomainId":     domainId,
		"Ids":          pq.Array(search.Ids),
		"Q":            search.GetQ(),
		"CallIds":      pq.Array(search.CallIds),
		"FormIds":      pq.Array(search.FormIds),
		"RatedUserIds": pq.Array(search.RatedUserIds),
		"From":         model.GetBetweenFromTime(search.CreatedAt),
		"To":           model.GetBetweenToTime(search.CreatedAt),
	}

	err := s.ListQuery(ctx, &list, search.ListRequest,
		`domain_id = :DomainId
    and (:Ids::int[] isnull or id = any(:Ids))
    and (:CallIds::varchar[] isnull or call_id = any(:CallIds))
    and (:FormIds::int[] isnull or form_id = any(:FormIds))
    and (:RatedUserIds::int8[] isnull or rated_user_id = any(:RatedUserIds))
	and ( :From::timestamptz isnull or created_at >= :From::timestamptz )
	and ( :To::timestamptz isnull or created_at <= :To::timestamptz )
	and (:Q::varchar isnull or ("comment" ilike :Q::varchar))
`,
		model.AuditRate{}, f)
	if err != nil {
		return nil, model.NewCustomCodeError("store.sql_audit_rate.get_all.app_error", err.Error(), extractCodeFromErr(err))
	}

	return list, nil
}

func (s SqlAuditRateStore) Get(ctx context.Context, domainId int64, id int64) (*model.AuditRate, model.AppError) {
	var rate *model.AuditRate
	f := map[string]interface{}{
		"DomainId": domainId,
		"Id":       id,
	}

	err := s.One(ctx, &rate,
		`domain_id = :DomainId and id = :Id`,
		model.AuditRate{}, f)
	if err != nil {
		return nil, model.NewCustomCodeError("store.sql_audit_rate.get.app_error", err.Error(), extractCodeFromErr(err))
	}

	return rate, nil
}

func (s SqlAuditRateStore) FormId(ctx context.Context, domainId, id int64) (int32, model.AppError) {
	res, err := s.GetReplica().WithContext(ctx).SelectNullInt(`select r.form_id
from call_center.cc_audit_rate r
where r.id = :Id and r.domain_id = :DomainId`, map[string]interface{}{
		"Id":       id,
		"DomainId": domainId,
	})

	if err != nil {
		return 0, model.NewCustomCodeError("store.sql_audit_rate.get_form.app_error", err.Error(), extractCodeFromErr(err))
	}

	return int32(res.Int64), nil
}

func NewSqlAuditRateStore(sqlStore SqlStore) store.AuditRateStore {
	us := &SqlAuditRateStore{sqlStore}
	return us
}
