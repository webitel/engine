package sqlstore

import (
	"fmt"
	"github.com/webitel/engine/model"
	"github.com/webitel/engine/store"
	"net/http"
)

type SqlEmailProfileStore struct {
	SqlStore
}

func NewSqlEmailProfileStore(sqlStore SqlStore) store.EmailProfileStore {
	us := &SqlEmailProfileStore{sqlStore}
	return us
}

func (s SqlEmailProfileStore) Create(domainId int64, p *model.EmailProfile) (*model.EmailProfile, *model.AppError) {
	var profile *model.EmailProfile
	err := s.GetMaster().SelectOne(&profile, `with t as (
    insert into call_center.cc_email_profile ( domain_id, name, description, enabled, updated_at, flow_id, imap_host, mailbox, imap_port, smtp_port,
                              login, password, created_at, created_by, updated_by, last_activity_at, smtp_host, fetch_interval)
values (:DomainId, :Name, :Description, :Enabled, now(), :FlowId, :ImapHost, :Mailbox, :Imap, :Smtp, :Login, :Pass,
        now(), :CreatedBy, :UpdatedBy, now(), :SmtpHost, :FetchInterval)
	returning *
)
SELECT t.id,
       t.domain_id,
       call_center.cc_view_timestamp(t.created_at)                         AS created_at,
       call_center.cc_get_lookup(t.created_by, cc.name::character varying) AS created_by,
       call_center.cc_view_timestamp(t.updated_at)                         AS updated_at,
       call_center.cc_get_lookup(t.updated_by, cu.name::character varying) AS updated_by,
       call_center.cc_view_timestamp(t.last_activity_at)                         AS activity_at,
       t.name,
       t.imap_host,
       t.smtp_host,
       t.login,
       t.mailbox,
       t.smtp_port,
       t.imap_port,
       t.fetch_err as fetch_error,
       t.fetch_interval,
       t.state,
       call_center.cc_get_lookup(t.flow_id::bigint, s.name)                AS schema,
       t.description,
       t.enabled,
       t.password
FROM t
         LEFT JOIN directory.wbt_user cc ON cc.id = t.created_by
         LEFT JOIN directory.wbt_user cu ON cu.id = t.updated_by
         LEFT JOIN flow.acr_routing_scheme s ON s.id = t.flow_id`, map[string]interface{}{
		"DomainId":      domainId,
		"Name":          p.Name,
		"Description":   p.Description,
		"Enabled":       p.Enabled,
		"FlowId":        p.Schema.Id,
		"ImapHost":      p.ImapHost,
		"SmtpHost":      p.SmtpHost,
		"FetchInterval": p.FetchInterval,
		"Mailbox":       p.Mailbox,
		"Imap":          p.ImapPort,
		"Smtp":          p.SmtpPort,
		"Login":         p.Login,
		"Pass":          p.Password,
		"CreatedBy":     p.CreatedBy.GetSafeId(),
		"UpdatedBy":     p.UpdatedBy.GetSafeId(),
	})

	if err != nil {
		return nil, model.NewAppError("SqlEmailProfileStore.Create", "store.sql_email_profile.create.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	return profile, nil
}

func (s SqlEmailProfileStore) GetAllPage(domainId int64, search *model.SearchEmailProfile) ([]*model.EmailProfile, *model.AppError) {
	var profiles []*model.EmailProfile

	f := map[string]interface{}{
		"DomainId": domainId,
		"Q":        search.GetQ(),
	}

	err := s.ListQuery(&profiles, search.ListRequest,
		`domain_id = :DomainId and (  (:Q::varchar isnull or (description ilike :Q::varchar or name ilike :Q::varchar ) ))`,
		model.EmailProfile{}, f)
	if err != nil {
		return nil, model.NewAppError("SqlEmailProfileStore.GetAllPage", "store.sql_email_profile.get_all.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	return profiles, nil
}

func (s SqlEmailProfileStore) Get(domainId int64, id int) (*model.EmailProfile, *model.AppError) {
	var profile *model.EmailProfile
	err := s.GetReplica().SelectOne(&profile, `
	SELECT t.id,
		   t.domain_id,
		   call_center.cc_view_timestamp(t.created_at)                         AS created_at,
		   call_center.cc_get_lookup(t.created_by, cc.name::character varying) AS created_by,
		   call_center.cc_view_timestamp(t.updated_at)                         AS updated_at,
		   call_center.cc_get_lookup(t.updated_by, cu.name::character varying) AS updated_by,
		   call_center.cc_view_timestamp(t.last_activity_at)                         AS activity_at,
		   t.name,
		   t.imap_host,
		   t.smtp_host,
		   t.login,
		   t.mailbox,
		   t.smtp_port,
		   t.imap_port,
		   t.fetch_err as fetch_error,
		   t.fetch_interval,
		   t.state,
		   call_center.cc_get_lookup(t.flow_id::bigint, s.name)                AS schema,
		   t.description,
		   t.enabled,
		   t.password
	FROM call_center.cc_email_profile t
			 LEFT JOIN directory.wbt_user cc ON cc.id = t.created_by
			 LEFT JOIN directory.wbt_user cu ON cu.id = t.updated_by
			 LEFT JOIN flow.acr_routing_scheme s ON s.id = t.flow_id
	where t.id = :Id and t.domain_id = :DomainId`, map[string]interface{}{
		"Id":       id,
		"DomainId": domainId,
	})

	if err != nil {
		return nil, model.NewAppError("SqlEmailProfileStore.Get", "store.sql_email_profile.get.app_error", nil,
			fmt.Sprintf("Id = %d, error: %s", id, err.Error()), extractCodeFromErr(err))
	}

	return profile, nil
}

func (s SqlEmailProfileStore) Update(domainId int64, p *model.EmailProfile) (*model.EmailProfile, *model.AppError) {
	var profile *model.EmailProfile
	err := s.GetMaster().SelectOne(&profile, `with t as (
    update call_center.cc_email_profile
        set name = :Name,
			description= :Description,
			flow_id = :FlowId,
            imap_host = :ImapHost,
            login = :Login,
            password = :Pass,
            mailbox = :Mailbox,
            smtp_port = :Smtp,
            imap_port = :Imap,
		    enabled = :Enabled,
            updated_by = :UpdatedBy,
            updated_at = now(),
			smtp_host = :SmtpHost,
			fetch_interval = :FetchInterval
        where id = :Id and domain_id = :DomainId
        returning *
)
SELECT t.id,
       t.domain_id,
       call_center.cc_view_timestamp(t.created_at)                         AS created_at,
       call_center.cc_get_lookup(t.created_by, cc.name::character varying) AS created_by,
       call_center.cc_view_timestamp(t.updated_at)                         AS updated_at,
       call_center.cc_get_lookup(t.updated_by, cu.name::character varying) AS updated_by,
       call_center.cc_view_timestamp(t.last_activity_at)                         AS activity_at,
       t.name,
       t.imap_host,
       t.smtp_host,
       t.login,
       t.mailbox,
       t.smtp_port,
       t.imap_port,
       t.fetch_err as fetch_error,
       t.fetch_interval,
       t.state,
       call_center.cc_get_lookup(t.flow_id::bigint, s.name)                AS schema,
       t.description,
       t.enabled,
       t.password
FROM t
         LEFT JOIN directory.wbt_user cc ON cc.id = t.created_by
         LEFT JOIN directory.wbt_user cu ON cu.id = t.updated_by
         LEFT JOIN flow.acr_routing_scheme s ON s.id = t.flow_id`, map[string]interface{}{
		"DomainId":      domainId,
		"Name":          p.Name,
		"Description":   p.Description,
		"FlowId":        p.Schema.Id,
		"ImapHost":      p.ImapHost,
		"Login":         p.Login,
		"Pass":          p.Password,
		"Mailbox":       p.Mailbox,
		"Smtp":          p.SmtpPort,
		"Imap":          p.ImapPort,
		"UpdatedBy":     p.UpdatedBy.GetSafeId(),
		"Id":            p.Id,
		"Enabled":       p.Enabled,
		"SmtpHost":      p.SmtpHost,
		"FetchInterval": p.FetchInterval,
	})

	if err != nil {
		return nil, model.NewAppError("SqlEmailProfileStore.Update", "store.sql_email_profile.update.app_error", nil, err.Error(),
			extractCodeFromErr(err))
	}

	return profile, nil
}

func (s SqlEmailProfileStore) Delete(domainId int64, id int) *model.AppError {
	if _, err := s.GetMaster().Exec(`delete from call_center.cc_email_profile c where c.id=:Id and c.domain_id = :DomainId`,
		map[string]interface{}{"Id": id, "DomainId": domainId}); err != nil {
		return model.NewAppError("SqlEmailProfileStore.Delete", "store.sql_email_profile.delete.app_error", nil,
			fmt.Sprintf("Id=%v, %s", id, err.Error()), extractCodeFromErr(err))
	}
	return nil
}

func (s SqlEmailProfileStore) SetupOAuth2(id int, params *model.MailProfileParams) *model.AppError {
	_, err := s.GetMaster().Exec(`update call_center.cc_email_profile
set params = :Params
where id = :Id;`, map[string]interface{}{
		"Id":     id,
		"Params": params.Json(),
	})

	if err != nil {
		return model.NewAppError("SqlEmailProfileStore.SetupOAuth2", "store.sql_email_profile.oauth.app_error", nil,
			fmt.Sprintf("Id=%v, %s", id, err.Error()), extractCodeFromErr(err))
	}

	return nil
}
