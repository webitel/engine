package sqlstore

import (
	"context"
	"encoding/json"
	"fmt"
	"golang.org/x/oauth2"

	"github.com/webitel/engine/model"
	"github.com/webitel/engine/store"
)

type SqlEmailProfileStore struct {
	SqlStore
}

func NewSqlEmailProfileStore(sqlStore SqlStore) store.EmailProfileStore {
	us := &SqlEmailProfileStore{sqlStore}
	return us
}

func (s SqlEmailProfileStore) Create(ctx context.Context, domainId int64, p *model.EmailProfile) (*model.EmailProfile, model.AppError) {
	var profile *model.EmailProfile
	err := s.GetMaster().WithContext(ctx).SelectOne(&profile, `with t as (
    insert into call_center.cc_email_profile ( domain_id, name, description, enabled, updated_at, flow_id, imap_host, mailbox, imap_port, smtp_port,
                              login, password, created_at, created_by, updated_by, last_activity_at, smtp_host, fetch_interval, auth_type, "listen", params)
values (:DomainId, :Name, :Description, :Enabled, now(), :FlowId, :ImapHost, :Mailbox, :Imap, :Smtp, :Login, :Pass,
        now(), :CreatedBy, :UpdatedBy, now(), :SmtpHost, :FetchInterval, :AuthType, :Listen, :Params)
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
       t.password,
	   t.auth_type,
	   t.params,
	   t.listen
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
		"AuthType":      p.AuthType,
		"Listen":        p.Listen,
		"Params":        p.Params.Json(),
	})

	if err != nil {
		return nil, model.NewInternalError("store.sql_email_profile.create.app_error", err.Error())
	}

	return profile, nil
}

func (s SqlEmailProfileStore) GetAllPage(ctx context.Context, domainId int64, search *model.SearchEmailProfile) ([]*model.EmailProfile, model.AppError) {
	var profiles []*model.EmailProfile

	f := map[string]interface{}{
		"DomainId": domainId,
		"Q":        search.GetQ(),
	}

	err := s.ListQuery(ctx, &profiles, search.ListRequest,
		`domain_id = :DomainId and (  (:Q::varchar isnull or (description ilike :Q::varchar or name ilike :Q::varchar ) ))`,
		model.EmailProfile{}, f)
	if err != nil {
		return nil, model.NewInternalError("store.sql_email_profile.get_all.app_error", err.Error())
	}

	return profiles, nil
}

func (s SqlEmailProfileStore) Get(ctx context.Context, domainId int64, id int) (*model.EmailProfile, model.AppError) {
	var profile *model.EmailProfile
	err := s.GetReplica().WithContext(ctx).SelectOne(&profile, `
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
		   t.password,
           t.auth_type,
		   t.listen,
		   t.params,
		   t.token->>'expiry' notnull and t.token->>'access_token' notnull as logged
	FROM call_center.cc_email_profile t
			 LEFT JOIN directory.wbt_user cc ON cc.id = t.created_by
			 LEFT JOIN directory.wbt_user cu ON cu.id = t.updated_by
			 LEFT JOIN flow.acr_routing_scheme s ON s.id = t.flow_id
	where t.id = :Id and t.domain_id = :DomainId`, map[string]interface{}{
		"Id":       id,
		"DomainId": domainId,
	})

	if err != nil {
		return nil, model.NewCustomCodeError("store.sql_email_profile.get.app_error", fmt.Sprintf("Id = %d, error: %s", id, err.Error()), extractCodeFromErr(err))
	}

	return profile, nil
}

func (s SqlEmailProfileStore) Update(ctx context.Context, domainId int64, p *model.EmailProfile) (*model.EmailProfile, model.AppError) {
	var profile *model.EmailProfile
	err := s.GetMaster().WithContext(ctx).SelectOne(&profile, `with t as (
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
			fetch_interval = :FetchInterval,
            auth_type = :AuthType,
			"listen" = :Listen,
            params = case when not :Params::jsonb isnull then :Params::jsonb end 
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
       t.password,
       t.auth_type,
	   t.listen,
	   t.params,
	   t.token->>'expiry' notnull and t.token->>'access_token' notnull as logged
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
		"AuthType":      p.AuthType,
		"Listen":        p.Listen,
		"Params":        p.Params.Json(),
	})

	if err != nil {
		return nil, model.NewCustomCodeError("store.sql_email_profile.update.app_error", err.Error(), extractCodeFromErr(err))
	}

	return profile, nil
}

func (s SqlEmailProfileStore) Delete(ctx context.Context, domainId int64, id int) model.AppError {
	if _, err := s.GetMaster().WithContext(ctx).Exec(`delete from call_center.cc_email_profile c where c.id=:Id and c.domain_id = :DomainId`,
		map[string]interface{}{"Id": id, "DomainId": domainId}); err != nil {
		return model.NewCustomCodeError("store.sql_email_profile.delete.app_error", fmt.Sprintf("Id=%v, %s", id, err.Error()), extractCodeFromErr(err))
	}
	return nil
}

func (s SqlEmailProfileStore) SetupOAuth2(ctx context.Context, id int, token *oauth2.Token) model.AppError {

	data, _ := json.Marshal(token)

	_, err := s.GetMaster().WithContext(ctx).Exec(`update call_center.cc_email_profile
set token = :Token
where id = :Id;`, map[string]interface{}{
		"Id":    id,
		"Token": data,
	})

	if err != nil {
		return model.NewCustomCodeError("store.sql_email_profile.oauth.app_error", fmt.Sprintf("Id=%v, %s", id, err.Error()), extractCodeFromErr(err))
	}

	return nil
}

func (s SqlEmailProfileStore) CountEnabledByDomain(ctx context.Context, domainId int64) (int, model.AppError) {
	count, err := s.GetReplica().WithContext(ctx).SelectInt(`select count(*)
from call_center.cc_email_profile p
where p.domain_id in (select distinct d.dc
    from directory.wbt_domain d
    left join directory.wbt_domain d2 on d2.customer_id = d.customer_id
    where d2.dc = :DomainId
)
and p.enabled`, map[string]interface{}{
		"DomainId": domainId,
	})

	if err != nil {
		return 0, model.NewCustomCodeError("store.sql_email_profile.count_enabled.app_error", fmt.Sprintf("DomainId=%v, %s", domainId, err.Error()), extractCodeFromErr(err))
	}

	return int(count), nil
}
