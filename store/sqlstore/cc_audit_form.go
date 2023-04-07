package sqlstore

import (
	"context"
	"fmt"
	"github.com/lib/pq"
	"github.com/webitel/engine/auth_manager"
	"github.com/webitel/engine/model"
	"github.com/webitel/engine/store"
)

type SqlAuditFormStore struct {
	SqlStore
}

func (s SqlAuditFormStore) CheckAccess(ctx context.Context, domainId int64, id int32, groups []int, access auth_manager.PermissionAccess) (bool, *model.AppError) {
	res, err := s.GetReplica().WithContext(ctx).SelectNullInt(`select 1
		where exists(
          select 1
          from call_center.cc_audit_form_acl a
          where a.dc = :DomainId
            and a.object = :Id
            and a.subject = any (:Groups::int[])
            and a.access & :Access = :Access
        )`, map[string]interface{}{"DomainId": domainId, "Id": id, "Groups": pq.Array(groups), "Access": access.Value()})

	if err != nil {
		return false, nil
	}

	return res.Valid && res.Int64 == 1, nil
}

func (s SqlAuditFormStore) Create(ctx context.Context, domainId int64, form *model.AuditForm) (*model.AuditForm, *model.AppError) {
	err := s.GetMaster().WithContext(ctx).SelectOne(&form, `with ins as (
    insert into call_center.cc_audit_form (domain_id, name, description, enabled, created_by, created_at, updated_by, updated_at, questions, team_ids)
    values (:DomainId, :Name, :Description, :Enabled, :CreatedBy, :CreatedAt, :UpdatedBy,  :UpdatedAt, :Questions, :TeamIds::int[])
    returning *
)
SELECT i.id,
       i.name,
       i.description,
       i.created_at,
       call_center.cc_get_lookup(uc.id, uc.name::character varying) AS created_by,
       i.updated_at,
       call_center.cc_get_lookup(u.id, u.name::character varying)   AS updated_by,
       i.enabled,
       i.questions,
       call_center.cc_get_lookup(u.id, u.name::character varying)   AS updated_by,
              (SELECT jsonb_agg(call_center.cc_get_lookup(aud.id,
                                                   aud.name::character varying)) AS jsonb_agg
        FROM call_center.cc_team aud
        WHERE aud.id = ANY (i.team_ids))                                                                   AS teams,
       i.editable,
       i.archive
FROM ins i
         LEFT JOIN directory.wbt_user uc ON uc.id = i.created_by
         LEFT JOIN directory.wbt_user u ON u.id = i.updated_by`, map[string]interface{}{
		"DomainId":    domainId,
		"Name":        form.Name,
		"Description": form.Description,
		"Enabled":     form.Enabled,
		"CreatedBy":   form.CreatedBy.GetSafeId(),
		"CreatedAt":   form.CreatedAt,
		"UpdatedBy":   form.UpdatedBy.GetSafeId(),
		"UpdatedAt":   form.UpdatedAt,
		"Questions":   form.Questions.ToJson(),
		"TeamIds":     pq.Array(model.LookupIds(form.Teams)),
	})

	if err != nil {
		return nil, model.NewInternalError("SqlAuditFormStore", "store.sql_audit_form.save.app_error", err.Error())
	}

	return form, nil
}

func (s SqlAuditFormStore) GetAllPage(ctx context.Context, domainId int64, search *model.SearchAuditForm) ([]*model.AuditForm, *model.AppError) {
	var list []*model.AuditForm

	f := map[string]interface{}{
		"DomainId": domainId,
		"Ids":      pq.Array(search.Ids),
		"Q":        search.GetQ(),
		"TeamIds":  pq.Array(search.TeamIds),
		"Archive":  search.Archive,
		"Editable": search.Editable,
		"Enabled":  search.Enabled,
		"Question": model.ReplaceWebSearch(search.Question),
	}

	err := s.ListQuery(ctx, &list, search.ListRequest,
		`domain_id = :DomainId
				and (:Q::varchar isnull or (name ilike :Q::varchar or description ilike :Q::varchar))
				and (:Ids::int[] isnull or id = any(:Ids))
				and (:TeamIds::int[] isnull or team_ids && :TeamIds)
				and (:Archive::bool isnull or archive = :Archive)
			    and (:Editable::bool isnull or editable = :Editable)
			    and (:Enabled::bool isnull or enabled = :Enabled)
			    and (:Question::varchar isnull or (exists(select 1 from jsonb_array_elements(questions) q where q->>'question' ilike :Question::varchar)))
`,
		model.AuditForm{}, f)
	if err != nil {
		return nil, model.NewAppError("SqlAuditFormStore.GetAllPage", "store.sql_audit_form.get_all.app_error", nil, err.Error(), extractCodeFromErr(err))
	}

	return list, nil
}

func (s SqlAuditFormStore) GetAllPageByGroup(ctx context.Context, domainId int64, groups []int, search *model.SearchAuditForm) ([]*model.AuditForm, *model.AppError) {
	var list []*model.AuditForm

	f := map[string]interface{}{
		"Groups":   pq.Array(groups),
		"Access":   auth_manager.PERMISSION_ACCESS_READ.Value(),
		"DomainId": domainId,
		"Ids":      pq.Array(search.Ids),
		"Q":        search.GetQ(),
		"TeamIds":  search.TeamIds,
		"Archive":  search.Archive,
		"Editable": search.Editable,
		"Question": model.ReplaceWebSearch(search.Question),
	}

	err := s.ListQuery(ctx, &list, search.ListRequest,
		`domain_id = :DomainId
				and (:Ids::int[] isnull or id = any(:Ids))
				and (:TeamIds::int[] isnull or team_ids = any(:TeamIds))
				and (:Archive::bool isnull or archive = :Archive)
			    and (:Editable::bool isnull or editable = :Editable)
				and (:Question::varchar isnull or (exists(select 1 from jsonb_array_elements(questions) q where q->>'question' ilike :Question::varchar)))
				and (
					exists(select 1
					  from call_center.cc_audit_form_acl
					  where call_center.cc_audit_form_acl.dc = t.domain_id and call_center.cc_audit_form_acl.object = t.id 
						and call_center.cc_audit_form_acl.subject = any(:Groups::int[]) and call_center.cc_audit_form_acl.access&:Access = :Access)
		  		)`,
		model.AuditForm{}, f)
	if err != nil {
		return nil, model.NewAppError("SqlAuditFormStore.GetAllPageByGroups", "store.sql_audit_form.get_all.app_error", nil, err.Error(), extractCodeFromErr(err))
	}

	return list, nil
}

func (s SqlAuditFormStore) Get(ctx context.Context, domainId int64, id int32) (*model.AuditForm, *model.AppError) {
	var form *model.AuditForm
	err := s.GetReplica().WithContext(ctx).SelectOne(&form, `SELECT i.id,
       i.name,
       i.description,
       i.created_at,
       call_center.cc_get_lookup(uc.id, uc.name::character varying) AS created_by,
       i.updated_at,
       call_center.cc_get_lookup(u.id, u.name::character varying)   AS updated_by,
       i.enabled,
       i.questions,
       call_center.cc_get_lookup(u.id, u.name::character varying)   AS updated_by,
              (SELECT jsonb_agg(call_center.cc_get_lookup(aud.id,
                                                   aud.name::character varying)) AS jsonb_agg
        FROM call_center.cc_team aud
        WHERE aud.id = ANY (i.team_ids))                                                                   AS teams,
       i.editable,
       i.archive
FROM call_center.cc_audit_form i
         LEFT JOIN directory.wbt_user uc ON uc.id = i.created_by
         LEFT JOIN directory.wbt_user u ON u.id = i.updated_by
where i.domain_id = :DomainId and i.id = :Id`, map[string]interface{}{
		"DomainId": domainId,
		"Id":       id,
	})

	if err != nil {
		return nil, model.NewAppError("SqlAuditFormStore.Get", "store.sql_audit_form.get.app_error", nil,
			fmt.Sprintf("Id=%v, %s", id, err.Error()), extractCodeFromErr(err))
	}

	return form, nil
}

func (s SqlAuditFormStore) Update(ctx context.Context, domainId int64, form *model.AuditForm) (*model.AuditForm, *model.AppError) {
	err := s.GetMaster().WithContext(ctx).SelectOne(&form, `with ins as (
    update call_center.cc_audit_form
		set updated_by = :UpdatedBy,
			updated_at = :UpdatedAt,
			name = :Name,
			description = :Description,
			enabled = :Enabled,
			questions = :Questions,
			archive = :Archive,
			team_ids = :TeamIds::int[]
		where domain_id = :DomainId and id = :Id
		returning *
)
SELECT i.id,
       i.name,
       i.description,
       i.created_at,
       call_center.cc_get_lookup(uc.id, uc.name::character varying) AS created_by,
       i.updated_at,
       call_center.cc_get_lookup(u.id, u.name::character varying)   AS updated_by,
       i.enabled,
       i.questions,
       call_center.cc_get_lookup(u.id, u.name::character varying)   AS updated_by,
              (SELECT jsonb_agg(call_center.cc_get_lookup(aud.id,
                                                   aud.name::character varying)) AS jsonb_agg
        FROM call_center.cc_team aud
        WHERE aud.id = ANY (i.team_ids))                                                                   AS teams,
       i.editable,
       i.archive
FROM ins i
         LEFT JOIN directory.wbt_user uc ON uc.id = i.created_by
         LEFT JOIN directory.wbt_user u ON u.id = i.updated_by`, map[string]interface{}{
		"Id":          form.Id,
		"DomainId":    domainId,
		"Name":        form.Name,
		"Description": form.Description,
		"Enabled":     form.Enabled,
		"UpdatedBy":   form.UpdatedBy.GetSafeId(),
		"UpdatedAt":   form.UpdatedAt,
		"Questions":   form.Questions.ToJson(),
		"Archive":     form.Archive,
		"TeamIds":     pq.Array(model.LookupIds(form.Teams)),
	})

	if err != nil {
		return nil, model.NewAppError("SqlAuditFormStore", "store.sql_audit_form.update.app_error", nil, err.Error(), extractCodeFromErr(err))
	}

	return form, nil
}

func (s SqlAuditFormStore) Delete(ctx context.Context, domainId int64, id int32) *model.AppError {
	if _, err := s.GetMaster().WithContext(ctx).Exec(`delete from call_center.cc_audit_form c where c.id=:Id and c.domain_id = :DomainId`,
		map[string]interface{}{"Id": id, "DomainId": domainId}); err != nil {
		return model.NewAppError("SqlAuditFormStore.Delete", "store.sql_audit_form.delete.app_error", nil,
			fmt.Sprintf("Id=%v, %s", id, err.Error()), extractCodeFromErr(err))
	}
	return nil
}

func (s SqlAuditFormStore) SetEditable(ctx context.Context, id int32, editable bool) *model.AppError {
	_, err := s.GetMaster().WithContext(ctx).Exec(`update call_center.cc_audit_form
set editable = :Editable 
where id = :Id`, map[string]interface{}{
		"Editable": editable,
		"Id":       id,
	})
	if err != nil {
		return model.NewAppError("SqlAuditFormStore.SetEditable", "store.sql_audit_form.set_editable.app_error", nil,
			fmt.Sprintf("Id=%v, %s", id, err.Error()), extractCodeFromErr(err))
	}

	return nil
}

func NewSqlAuditFormStore(sqlStore SqlStore) store.AuditFormStore {
	us := &SqlAuditFormStore{sqlStore}
	return us
}
