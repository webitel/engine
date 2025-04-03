package sqlstore

import (
	"context"
	"fmt"

	"github.com/webitel/engine/model"
	"github.com/webitel/engine/store"
)

type SqlChatStore struct {
	SqlStore
}

func NewSqlChatStore(sqlStore SqlStore) store.ChatStore {
	us := &SqlChatStore{sqlStore}
	return us
}

// todo deprecated
func (s SqlChatStore) OpenedConversations(ctx context.Context, domainId, userId int64, hasContact bool) ([]*model.Conversation, model.AppError) {
	var res []*model.Conversation
	_, err := s.GetMaster().WithContext(ctx).Select(&res, `
select c.id,
       ch.invite_id,
       ch.id                                                                channel_id,
       c.title,
       call_center.cc_view_timestamp(coalesce(ch.created_at, c.created_at)) created_at,
       call_center.cc_view_timestamp(c.updated_at)                          updated_at,
       call_center.cc_view_timestamp(ch.joined_at)                          joined_at,
       call_center.cc_view_timestamp(ch.closed_at)                          closed_at,
       m.messages,
       mem.members,
       case when cont.contact_id notnull then jsonb_build_object('wbt_contact_id', cont.contact_id::text) else '{}' end || coalesce(ch.props, '{}')::jsonb as variables,
       row_to_json(at)                                                      task,
       at.leaving_at                   as                                   leaving_at
from (select 1                    pri,
             inv.created_at,
             null::uuid           id,
             inv.id::uuid         invite_id,
             null::timestamptz    joined_at,
             inv.conversation_id,
             inv.user_id,
             inv.created_at       updated_at,
             inv.props,
             null::timestamptz as closed_at
      from chat.invite inv
      where inv.user_id = :UserId::int8
        and inv.closed_at isnull
        and inv.domain_id = :DomainId::int8
      union

      select 2,
             ch.created_at,
             ch.id::uuid,
             null::uuid       invite_id,
             ch.created_at as joined_at,
             ch.conversation_id,
             ch.user_id,
             ch.updated_at,
             ch.props,
             ch.closed_at
      from (select ch.id, ch.created_at, ch.conversation_id, ch.user_id, ch.updated_at, ch.props, ch.closed_at
            from chat.channel ch
            where ch.user_id = :UserId::int8
              and ch.internal
              and (ch.closed_at isnull or exists(select 1
                                                 from call_center.cc_member_attempt mat
                                                          inner join call_center.cc_agent a on a.id = mat.agent_id
                                                 where a.user_id = :UserId::int8
                                                   and mat.agent_call_id = ch.id::text
                                                   and mat.state != 'leaving'))
              and ch.domain_id = :DomainId::int8
            order by ch.created_at desc, ch.updated_at desc
            limit 40) ch) ch
         inner join chat.conversation c on c.id = ch.conversation_id
         left join lateral (
    select json_agg(t) messages
    from (select m.id,
                 call_center.cc_view_timestamp(m.created_at) created_at,
                 call_center.cc_view_timestamp(m.updated_at) updated_at,
                 m.text   as                                 text,
                 (case
                      when (m.file_id isnull and nullif(m.file_url, '') isnull) then null
                      else
                          json_build_object('id', m.file_id, 'size', m.file_size, 'mime', m.file_type, 'name',
                                            m.file_name, 'url', m.file_url)
                     end) as                                 "file",
                 m.type,
                 m.channel_id
          from chat.message m
          where m.conversation_id = ch.conversation_id
          order by m.created_at desc
          limit 250) t
    ) m on true
         left join lateral (
    select json_agg(t) members
    from (select ch2.id,
                 ch2.type,
                 ch2.user_id,
                 ch2.name,
                 ch2.props ->> 'user' as external_id,
          		 case when ch2.internal then null else json_build_object('id', b.id, 'name', b.name, 'type', b.provider) end AS via
          from chat.channel ch2
            left join chat.bot b on connection = b.id::text
          where ch2.conversation_id = c.id::uuid
            and (not ch2.id::uuid = ch.id or ch.id isnull)
          limit 10) t
    ) mem on true
            left join lateral (
            select i.contact_id
            from chat.channel ch
                left join contacts.contact_imclient i on i.user_id = ch.user_id
            where not ch.internal and ch.conversation_id = c.id  and i.protocol = ch.type
            order by c.created_at desc
            limit 1
        ) cont on :HasContact::bool
         left join lateral (
    select a.id                                                       as attempt_id,
           a.channel,
           a.node_id                                                  as app_id,
           a.queue_id,
           coalesce((a.queue_params ->> 'queue_name'), '')            as queue_name,
           a.member_id,
           a.member_call_id                                           as member_channel_id,
           a.agent_call_id                                            as agent_channel_id,
           a.destination,
           a.state,
           call_center.cc_view_timestamp(a.leaving_at)                as leaving_at,
           coalesce((a.queue_params -> 'has_reporting')::bool, false) as has_reporting,
           coalesce((a.queue_params -> 'has_form')::bool, false)      as has_form,
           (a.queue_params -> 'processing_sec')::int                  as processing_sec,
           (a.queue_params -> 'processing_renewal_sec')::int          as processing_renewal_sec,
           call_center.cc_view_timestamp(a.timeout)                   as processing_timeout_at,
           a.form_view                                                as form
    from call_center.cc_member_attempt a
    where a.agent_call_id = ch.id::varchar
    ) at on true
order by ch.pri, ch.updated_at desc`, map[string]interface{}{
		"UserId":     userId,
		"DomainId":   domainId,
		"HasContact": hasContact,
	})

	if err != nil {
		return nil, model.NewCustomCodeError("store.sql_chat.list_opened.app_error", fmt.Sprintf("userId=%v, %v", userId, err.Error()), extractCodeFromErr(err))
	}

	return res, nil
}

func (s SqlChatStore) ValidDomain(ctx context.Context, domainId int64, profileId int64) model.AppError {
	res, err := s.GetReplica().WithContext(ctx).SelectInt(`select 1
from chat.bot p
where p.dc = :DomainId and p.id = :Id`, map[string]interface{}{
		"DomainId": domainId,
		"Id":       profileId,
	})

	if err != nil {
		return model.NewCustomCodeError("store.sql_chat.valid_domain.app_error", fmt.Sprintf("domainId=%v, %v", domainId, err.Error()), extractCodeFromErr(err))
	}

	if res != 1 {
		return model.NewNotFoundError("store.sql_chat.valid_domain.not_found", fmt.Sprintf("domainId=%v", domainId))
	}

	return nil
}
