package sqlstore

import (
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
func (s SqlChatStore) OpenedConversations(domainId, userId int64) ([]*model.Conversation, *model.AppError) {
	var res []*model.Conversation
	_, err := s.GetMaster().Select(&res, `
select
    c.id,
    ch.invite_id,
    ch.id channel_id,
    c.title,
    cc_view_timestamp(c.created_at) created_at, 
    cc_view_timestamp(c.updated_at) updated_at,
    cc_view_timestamp(ch.joined_at) joined_at,
    m.messages,
    mem.members
from (
     select 1 pri, null::varchar id, inv.id invite_id, null::timestamptz joined_at, inv.conversation_id, inv.user_id, inv.created_at updated_at
     from chat.invite inv
     where inv.user_id = :UserId::int8 and inv.closed_at isnull
        and inv.domain_id = :DomainId::int8
     union

     select 2, ch.id, null::text invite_id, ch.joined_at, ch.conversation_id, ch.user_id, ch.updated_at
     from (
        select ch.id, ch.joined_at, ch.conversation_id, ch.user_id, ch.updated_at
        from chat.channel ch
        where ch.user_id = :UserId::int8 and ch.closed_at isnull
            and ch.domain_id = :DomainId::int8
        order by ch.created_at desc, ch.updated_at desc
        limit 40
     ) ch
) ch
    inner join chat.conversation c on c.id = ch.conversation_id
    left join lateral (
        select json_agg(t) messages
        from (
            select m.id, cc_view_timestamp(m.created_at) created_at,
                   cc_view_timestamp(m.updated_at) updated_at, m.text, m.type, m.channel_id
            from chat.message m
            where m.conversation_id = ch.conversation_id
            order by m.created_at desc
            limit 20
       ) t
    ) m on true
    left join lateral (
        select json_agg(t) members
        from (
            select
                   ch2.id,
                   ch2.type,
                   ch2.user_id,
                   ch2.name
            from chat.channel ch2
            where ch2.conversation_id = c.id
              and not ch2.id = ch.id
            limit 10
        ) t
    ) mem on true
order by ch.pri, ch.updated_at desc`, map[string]interface{}{
		"UserId":   userId,
		"DomainId": domainId,
	})

	if err != nil {
		return nil, model.NewAppError("ChatStore.OpenedConversations", "store.sql_chat.list_opened.app_error", nil,
			fmt.Sprintf("userId=%v, %v", userId, err.Error()), extractCodeFromErr(err))
	}

	return res, nil
}
