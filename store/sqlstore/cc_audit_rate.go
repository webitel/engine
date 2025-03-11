package sqlstore

import (
	"context"
	"github.com/webitel/engine/auth_manager"
	"net/http"

	"github.com/lib/pq"
	"github.com/webitel/engine/model"
	"github.com/webitel/engine/store"
)

type SqlAuditRateStore struct {
	SqlStore
}

func NewSqlAuditRateStore(sqlStore SqlStore) store.AuditRateStore {
	us := &SqlAuditRateStore{sqlStore}
	return us
}

func (s *SqlAuditRateStore) CheckAccess(ctx context.Context, domainId, rateId int64, groups []int, access auth_manager.PermissionAccess) (bool, model.AppError) {
	n, err := s.GetReplica().WithContext(ctx).SelectNullInt(`select 1
    from directory.wbt_class c
        inner join directory.wbt_default_acl a on a.object = c.id
        inner join call_center.cc_audit_rate ar on ar.id = :RateId and ar.domain_id = :DomainId
        inner join directory.wbt_auth a2 on a2.id = a.grantor
        left join directory.wbt_auth_member am on am.role_id = a2.id
        left join directory.wbt_auth a3 on a3.id = am.member_id and a3.can_login
    where c.name = :ClassName::varchar
      and c.dc = :DomainId::int8
      and a.access&:Access = :Access
      and a.subject = any(:Groups::int[])
      and case when a2.can_login then a2.id else a3.id end = ar.created_by;`, map[string]any{
		"ClassName": model.PermissionAuditRate,
		"DomainId":  domainId,
		"Access":    access.Value(),
		"Groups":    pq.Array(groups),
		"RateId":    rateId,
	})

	if err != nil {
		return false, model.NewCustomCodeError("store.sql_audit_rate.access", err.Error(), extractCodeFromErr(err))
	}

	return n.Int64 == 1, nil
}

func (s *SqlAuditRateStore) Create(ctx context.Context, domainId int64, rate *model.AuditRate) (*model.AuditRate, model.AppError) {
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

func (s *SqlAuditRateStore) GetAllPage(ctx context.Context, domainId int64, search *model.SearchAuditRate) ([]*model.AuditRate, model.AppError) {
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

func (s *SqlAuditRateStore) Get(ctx context.Context, domainId int64, id int64) (*model.AuditRate, model.AppError) {
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

func (s *SqlAuditRateStore) FormId(ctx context.Context, domainId, id int64) (int32, model.AppError) {
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

func (s *SqlAuditRateStore) Update(ctx context.Context, domainId int64, rate *model.AuditRate) (*model.AuditRate, model.AppError) {
	var ar *model.AuditRate
	err := s.GetMaster().WithContext(ctx).SelectOne(&ar, `with r as (
    update call_center.cc_audit_rate ar
    set answers = :Answers,
        comment = :Comment,
        updated_by = :UpdatedBy,
        updated_at = :UpdatedAt,
        score_required = :ScoreRequired,
        score_optional = :ScoreOptional
    where id = :Id and domain_id = :DomainId
    returning *
)
select r.id,
       r.created_at,
       call_center.cc_get_lookup(uc.id, coalesce(uc.name::character varying, uc.username::character varying)) AS created_by,
       r.updated_at,
       call_center.cc_get_lookup(u.id, coalesce(u.name::character varying, u.username::character varying))   AS updated_by,
       call_center.cc_get_lookup(ur.id, ur.name::character varying)   AS rated_user,
       call_center.cc_get_lookup(f.id, f.name::character varying)   AS form,
       ans.v AS answers,
       r.score_required,
       r.score_optional,
       r.comment,
       r.call_id,
       f.questions
from  r
    LEFT JOIN LATERAL ( SELECT jsonb_agg(
                                            CASE
                                                WHEN u_1.id IS NOT NULL THEN x.j || jsonb_build_object('updated_by',
                                                                                                       call_center.cc_get_lookup(
                                                                                                               u_1.id,
                                                                                                               COALESCE(u_1.name, u_1.username::text)::character varying))
                                                ELSE x.j
                                                END ORDER BY x.i) AS v
                             FROM jsonb_array_elements(r.answers) WITH ORDINALITY x(j, i)
                                      LEFT JOIN directory.wbt_user u_1
                                                ON u_1.id = (x.j -> 'updated_by'->'id'::text)::bigint) ans ON true
    left join call_center.cc_audit_form f on f.id = r.form_id
    LEFT JOIN directory.wbt_user uc ON uc.id = r.created_by
    LEFT JOIN directory.wbt_user u ON u.id = r.updated_by
    LEFT JOIN directory.wbt_user ur ON ur.id = r.rated_user_id`, map[string]any{
		"DomainId":      domainId,
		"Id":            rate.Id,
		"UpdatedAt":     rate.UpdatedAt,
		"UpdatedBy":     rate.UpdatedBy.GetSafeId(),
		"Answers":       rate.Answers.ToJson(),
		"ScoreRequired": rate.ScoreRequired,
		"ScoreOptional": rate.ScoreOptional,
		"Comment":       rate.Comment,
	})

	if err != nil {
		return nil, model.NewCustomCodeError("store.sql_audit_rate.update.app_error", err.Error(), extractCodeFromErr(err))
	}

	return ar, nil
}

func (s *SqlAuditRateStore) Delete(ctx context.Context, domainId, id int64) model.AppError {
	var rows int64
	r, err := s.GetMaster().WithContext(ctx).Exec(`delete
from call_center.cc_audit_rate
where domain_id = :DomainId and id = :Id;`, map[string]any{
		"DomainId": domainId,
		"Id":       id,
	})
	if err != nil {
		return model.NewCustomCodeError("store.sql_audit_rate.delete.app_error", err.Error(), extractCodeFromErr(err))
	}
	rows, err = r.RowsAffected()
	if err != nil {
		return model.NewCustomCodeError("store.sql_audit_rate.delete.app_error", err.Error(), extractCodeFromErr(err))
	}

	if rows != 1 {
		return model.NewCustomCodeError("store.sql_audit_rate.delete.not_found", "Not found", http.StatusNotFound)
	}

	return nil
}
