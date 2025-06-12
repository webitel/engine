package sqlstore

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/lib/pq"
	"github.com/webitel/engine/model"
	"github.com/webitel/engine/pkg/wbt/auth_manager"
	"github.com/webitel/engine/store"
)

type SqlCallStore struct {
	SqlStore
}

func NewSqlCallStore(sqlStore SqlStore) store.CallStore {
	us := &SqlCallStore{sqlStore}
	return us
}

func (s SqlCallStore) GetActive(ctx context.Context, domainId int64, search *model.SearchCall) ([]*model.Call, model.AppError) {
	var out []*model.Call

	f := map[string]interface{}{
		"Domain":        domainId,
		"Limit":         search.GetLimit(),
		"Offset":        search.GetOffset(),
		"From":          model.GetBetweenFromTime(search.CreatedAt),
		"To":            model.GetBetweenToTime(search.CreatedAt),
		"Q":             search.GetQ(),
		"UserIds":       pq.Array(search.UserIds),
		"QueueIds":      pq.Array(search.QueueIds),
		"TeamIds":       pq.Array(search.TeamIds),
		"AgentIds":      pq.Array(search.AgentIds),
		"MemberIds":     pq.Array(search.MemberIds),
		"GatewayIds":    pq.Array(search.GatewayIds),
		"SkipParent":    search.SkipParent,
		"ParentId":      search.ParentId,
		"Number":        search.Number,
		"Direction":     pq.Array(search.Direction),
		"Missed":        search.Missed,
		"AnsweredFrom":  model.GetBetweenFromTime(search.AnsweredAt),
		"AnsweredTo":    model.GetBetweenToTime(search.AnsweredAt),
		"DurationFrom":  model.GetBetweenFrom(search.Duration),
		"DurationTo":    model.GetBetweenTo(search.Duration),
		"SupervisorIds": pq.Array(search.SupervisorIds),
		"State":         pq.Array(search.State),
	}

	err := s.ListQueryMaster(ctx, &out, search.ListRequest,
		`domain_id = :Domain and direction notnull
	and (:Q::text isnull or destination ~ :Q  or  from_number ~ :Q or  to_number ~ :Q)
	and ( (:From::timestamptz isnull or :To::timestamptz isnull) or created_at between :From and :To )
	and (:UserIds::int8[] isnull or user_id = any(:UserIds) or exists(select 1 from call_center.cc_calls cc where cc.parent_id = t.id and cc.user_id = any(:UserIds)) )
	and (:QueueIds::int[] isnull or queue_id = any(:QueueIds) )
	and (:SupervisorIds::int[] isnull or supervisor_ids && :SupervisorIds or exists(select 1 from call_center.cc_call_active_list cc where cc.parent_id = t.id and cc.supervisor_ids && :SupervisorIds))
	and (:TeamIds::int[] isnull or team_id = any(:TeamIds) or exists(select 1 from call_center.cc_calls cc where cc.parent_id = t.id and cc.team_id = any(:TeamIds)))
	and (:AgentIds::int[] isnull or agent_id = any(:AgentIds) or exists(select 1 from call_center.cc_calls cc where cc.parent_id = t.id and cc.agent_id = any(:AgentIds)))
	and (:MemberIds::int8[] isnull or member_id = any(:MemberIds) )
	and (:GatewayIds::int8[] isnull or gateway_id = any(:GatewayIds) )
	and (:Number::varchar isnull or from_number ilike :Number::varchar or to_number ilike :Number::varchar or destination ilike :Number::varchar)
	and ( (:SkipParent::bool isnull or not :SkipParent::bool is true ) or parent_id isnull)
	and (:ParentId::uuid isnull or parent_id = :ParentId )
	and ( (:AnsweredFrom::timestamptz isnull or :AnsweredTo::timestamptz isnull) or answered_at between :AnsweredFrom and :AnsweredTo )
	and ( (:DurationFrom::int8 isnull or :DurationFrom::int8 = 0 or duration >= :DurationFrom ))
	and ( (:DurationTo::int8 isnull or :DurationTo::int8 = 0 or duration <= :DurationTo ))
	and (:Direction::varchar[] isnull or direction = any(:Direction) )
	and (:Missed::bool isnull or (:Missed and answered_at isnull))
	and (:State::varchar[] isnull or state = any(:State) )
`,
		model.Call{}, f)
	if err != nil {
		return nil, model.NewInternalError("store.sql_call.get_active.app_error", err.Error())
	}

	return out, nil
}

func (s SqlCallStore) GetActiveByGroups(ctx context.Context, domainId int64, userSupervisorId int64, groups []int, search *model.SearchCall) ([]*model.Call, model.AppError) {
	var out []*model.Call

	f := map[string]interface{}{
		"Domain":           domainId,
		"Limit":            search.GetLimit(),
		"Offset":           search.GetOffset(),
		"From":             model.GetBetweenFromTime(search.CreatedAt),
		"To":               model.GetBetweenToTime(search.CreatedAt),
		"Q":                search.GetQ(),
		"UserIds":          pq.Array(search.UserIds),
		"QueueIds":         pq.Array(search.QueueIds),
		"TeamIds":          pq.Array(search.TeamIds),
		"AgentIds":         pq.Array(search.AgentIds),
		"MemberIds":        pq.Array(search.MemberIds),
		"GatewayIds":       pq.Array(search.GatewayIds),
		"SkipParent":       search.SkipParent,
		"ParentId":         search.ParentId,
		"Number":           search.Number,
		"Direction":        pq.Array(search.Direction),
		"Missed":           search.Missed,
		"AnsweredFrom":     model.GetBetweenFromTime(search.AnsweredAt),
		"AnsweredTo":       model.GetBetweenToTime(search.AnsweredAt),
		"DurationFrom":     model.GetBetweenFrom(search.Duration),
		"DurationTo":       model.GetBetweenTo(search.Duration),
		"SupervisorIds":    pq.Array(search.SupervisorIds),
		"Groups":           pq.Array(groups),
		"Access":           auth_manager.PERMISSION_ACCESS_READ.Value(),
		"UserSupervisorId": userSupervisorId,
	}

	err := s.ListQueryMaster(ctx, &out, search.ListRequest,
		`domain_id = :Domain and direction notnull
	and (:Q::text isnull or destination ~ :Q  or  from_number ~ :Q or  to_number ~ :Q)
	and ( (:From::timestamptz isnull or :To::timestamptz isnull) or created_at between :From and :To )
	and (:UserIds::int8[] isnull or user_id = any(:UserIds))
	and (:QueueIds::int[] isnull or queue_id = any(:QueueIds) )
	and (:SupervisorIds::int[] isnull or supervisor_ids && :SupervisorIds )
	and (:TeamIds::int[] isnull or team_id = any(:TeamIds) )  
	and (:AgentIds::int[] isnull or agent_id = any(:AgentIds) )
	and (:MemberIds::int8[] isnull or member_id = any(:MemberIds) )
	and (:GatewayIds::int8[] isnull or gateway_id = any(:GatewayIds) )
	and (:Number::varchar isnull or from_number ilike :Number::varchar or to_number ilike :Number::varchar or destination ilike :Number::varchar)
	and ( (:SkipParent::bool isnull or not :SkipParent::bool is true ) or parent_id isnull)
	and (:ParentId::uuid isnull or parent_id = :ParentId )
	and ( (:AnsweredFrom::timestamptz isnull or :AnsweredTo::timestamptz isnull) or answered_at between :AnsweredFrom and :AnsweredTo )
	and ( (:DurationFrom::int8 isnull or :DurationFrom::int8 = 0 or duration >= :DurationFrom ))
	and ( (:DurationTo::int8 isnull or :DurationTo::int8 = 0 or duration <= :DurationTo ))
	and (:Direction::varchar[] isnull or direction = any(:Direction) )
	and (:Missed::bool isnull or (:Missed and answered_at isnull))
	and (
        (t.user_id in (
            with x as (
                select a.user_id, a.id agent_id, a.supervisor, a.domain_id
                from directory.wbt_user u
                         inner join call_center.cc_agent a on a.user_id = u.id
                where u.id = :UserSupervisorId
                  and u.dc = :Domain
            )
            select distinct a.user_id
            from x
                     left join lateral (
                select a.user_id, a.auditor_ids && array [x.user_id] aud
                from call_center.cc_agent a
                where a.domain_id = x.domain_id
                  and (a.user_id = x.user_id or (a.supervisor_ids && array [x.agent_id] and a.supervisor) or
                       a.auditor_ids && array [x.user_id])

                union
                distinct

                select a.user_id, a.auditor_ids && array [x.user_id] aud
                from call_center.cc_team t
                         inner join call_center.cc_agent a on a.team_id = t.id
                where t.admin_ids && array [x.agent_id]
                  and x.domain_id = t.domain_id
                ) a on true
        ))
        or (t.queue_id in (
        with x as (
            select a.user_id, a.id agent_id, a.supervisor, a.domain_id
            from directory.wbt_user u
                     inner join call_center.cc_agent a on a.user_id = u.id and a.domain_id = u.dc
            where u.id = :UserSupervisorId
              and u.dc = :Domain
        )
        select distinct qs.queue_id
        from x
                 left join lateral (
            select a.id, a.auditor_ids && array [x.user_id] aud
            from call_center.cc_agent a
            where (a.user_id = x.user_id or (a.supervisor_ids && array [x.agent_id] and a.supervisor))
            union
            distinct
            select a.id, a.auditor_ids && array [x.user_id] aud
            from call_center.cc_team t
                     inner join call_center.cc_agent a on a.team_id = t.id
            where t.admin_ids && array [x.agent_id]
            ) a on true
                 inner join call_center.cc_skill_in_agent sa on sa.agent_id = a.id
                 inner join call_center.cc_queue_skill qs
                            on qs.skill_id = sa.skill_id and sa.capacity between qs.min_capacity and qs.max_capacity
        where sa.enabled
          and qs.enabled
        union
        select q.id
        from call_center.cc_queue q
        where q.domain_id = :Domain
          and q.grantee_id = any (:Groups)
          and q.enabled
    ))
      or exists(
		select acl.*
		from (
			select a.*
			  from directory.wbt_default_acl a
			  join directory.wbt_class c on c.dc = t.domain_id and c.name = 'calls' and a.object = c.id
			 where (a.grantor = t.user_id or a.grantor = t.grantee_id)
				or exists(select r.role_id
						   from directory.wbt_auth_member r
						  where (r.member_id = t.user_id or r.member_id = t.grantee_id)
							and r.role_id = a.grantor
				   )
			union all
			values(t.domain_id, 0, t.user_id, t.user_id, 255)
		) acl
		where acl.subject = any(:Groups::int[]) and acl.access&:Access = :Access
	)
    )
`,
		model.Call{}, f)
	if err != nil {
		return nil, model.NewInternalError("store.sql_call.get_active.app_error", err.Error())
	}

	return out, nil
}

// fixme
func (s SqlCallStore) GetUserActiveCall(ctx context.Context, domainId, userId int64) ([]*model.Call, model.AppError) {
	var res []*model.Call
	_, err := s.GetMaster().WithContext(ctx).Select(&res, `select row_to_json(at)                                                    task,
       c."id",
       c."app_id",
       c."state",
       c."timestamp",
       c."parent_id",
       c."direction",
       c."destination",
       json_build_object('type', COALESCE(c.from_type, ''::character varying), 'number',
                         COALESCE(c.from_number, ''::character varying), 'id',
                         COALESCE(c.from_id, ''::character varying), 'name',
                         COALESCE(c.from_name, ''::character varying)) AS "from",
       CASE
           WHEN c.to_number::text <> ''::text THEN json_build_object('type', COALESCE(c.to_type, ''::character varying),
                                                                     'number',
                                                                     COALESCE(c.to_number, ''::character varying), 'id',
                                                                     COALESCE(c.to_id, ''::character varying), 'name',
                                                                     COALESCE(c.to_name, ''::character varying))
           ELSE NULL::json
           END                                                         AS "to",
       CASE
           WHEN c.payload IS NULL THEN '{}'::jsonb
           ELSE c.payload
           END                                                         AS variables,
       c."created_at",
       c."answered_at",
       c."bridged_at",
       c."hangup_at",
       c."hold_sec",
       call_center.cc_get_lookup(
               case when at.attempt_id::bigint notnull then coalesce(at.queue_id::bigint, 0) end, at.queue_name)   AS queue,
       c.contact_id,
       to_timestamp(at.leaving_at::double precision / 1000)            as leaving_at --todo
from call_center.cc_calls c
         left join lateral (
    select a.id                                                       as attempt_id,
           a.channel,
           a.node_id                                                  as app_id,
           a.queue_id,
           coalesce(a.queue_params ->> 'queue_name', '')               as queue_name,
           a.member_id,
           a.member_call_id                                           as member_channel_id,
           a.agent_call_id                                            as agent_channel_id,
           a.destination,
           a.state,
           call_center.cc_view_timestamp(a.leaving_at)                   leaving_at,
           coalesce((a.queue_params -> 'has_reporting')::bool, false) as has_reporting,
           coalesce((a.queue_params -> 'has_form')::bool, false)      as has_form,
           (a.queue_params -> 'processing_sec')::int                  as processing_sec,
           (a.queue_params -> 'processing_renewal_sec')::int          as processing_renewal_sec,
           call_center.cc_view_timestamp(a.timeout)                   as processing_timeout_at,
           a.form_view                                                as form
    from call_center.cc_member_attempt a
    where a.id = c.attempt_id
      and a.agent_call_id = c.id::text
    ) at on true
where c.user_id = :UserId
  and c.domain_id = :DomainId
  and ((at.state != 'leaving') or c.hangup_at isnull)`, map[string]interface{}{
		"UserId":   userId,
		"DomainId": domainId,
	})

	if err != nil {
		return nil, model.NewCustomCodeError("store.sql_call.get_user_active.app_error", err.Error(), extractCodeFromErr(err))
	}

	return res, nil
}

func (s SqlCallStore) Get(ctx context.Context, domainId int64, id string) (*model.Call, model.AppError) {
	var out *model.Call

	err := s.GetMaster().WithContext(ctx).SelectOne(&out, `
select c.id, c.app_id, c.state, c."timestamp", c.direction, c.destination, c.parent_id, c.created_at,
   json_build_object('type', coalesce(c.from_type, ''), 'number', coalesce(c.from_number, ''), 'id', coalesce(c.from_id, ''), 'name', coalesce(c.from_name, '')) "from",
   json_build_object('type', coalesce(c.to_type, ''), 'number', coalesce(c.to_number, ''), 'id', coalesce(c.to_id, ''), 'name', coalesce(c.to_name, '')) "to",
   (extract(epoch from now() -  c.created_at))::int8 duration, c.bridged_id	
from call_center.cc_calls c
where c.domain_id = :Domain and c.id = :Id::uuid`, map[string]interface{}{
		"Domain": domainId,
		"Id":     id,
	})

	if err != nil {
		return nil, model.NewCustomCodeError("store.sql_call.get.app_error", err.Error(), extractCodeFromErr(err))
	}

	return out, nil
}

func (s SqlCallStore) GetInstance(ctx context.Context, domainId int64, id string) (*model.CallInstance, model.AppError) {
	var inst *model.CallInstance
	err := s.GetMaster().WithContext(ctx).SelectOne(&inst, `select c.id, c.app_id, c.state
from call_center.cc_calls c
where c.id = :Id::uuid and c.domain_id = :Domain`, map[string]interface{}{
		"Id":     id,
		"Domain": domainId,
	})
	if err != nil {
		return nil, model.NewCustomCodeError("store.sql_call.get_instance.app_error", err.Error(), extractCodeFromErr(err))
	}

	return inst, nil
}

func (s SqlCallStore) GetHistory(ctx context.Context, domainId int64, search *model.SearchHistoryCall) ([]*model.HistoryCall, model.AppError) {
	var out []*model.HistoryCall

	f := map[string]interface{}{
		"Domain":           domainId,
		"Limit":            search.GetLimit(),
		"Offset":           search.GetOffset(),
		"From":             model.GetBetweenFromTime(search.CreatedAt),
		"To":               model.GetBetweenToTime(search.CreatedAt),
		"Q":                search.GetQ(),
		"UserIds":          pq.Array(search.UserIds),
		"QueueIds":         pq.Array(search.QueueIds),
		"TeamIds":          pq.Array(search.TeamIds),
		"AgentIds":         pq.Array(search.AgentIds),
		"MemberIds":        pq.Array(search.MemberIds),
		"GatewayIds":       pq.Array(search.GatewayIds),
		"SkipParent":       search.SkipParent,
		"ParentId":         search.ParentId,
		"Number":           model.GetRegExpQ(search.Number),
		"CauseArr":         pq.Array(search.CauseArr),
		"HasFile":          search.HasFile,
		"Direction":        search.Direction,
		"Missed":           search.Missed,
		"AnsweredFrom":     model.GetBetweenFromTime(search.AnsweredAt),
		"AnsweredTo":       model.GetBetweenToTime(search.AnsweredAt),
		"DurationFrom":     model.GetBetweenFrom(search.Duration),
		"DurationTo":       model.GetBetweenTo(search.Duration),
		"StoredAtFrom":     model.GetBetweenFromTime(search.StoredAt),
		"StoredAtTo":       model.GetBetweenToTime(search.StoredAt),
		"Ids":              pq.Array(search.Ids),
		"TransferFromIds":  pq.Array(search.TransferFromIds),
		"TransferToIds":    pq.Array(search.TransferToIds),
		"DependencyIds":    pq.Array(search.DependencyIds),
		"Tags":             pq.Array(search.Tags),
		"AmdResult":        pq.Array(search.AmdResult),
		"Variables":        search.Variables.ToSafeJson(),
		"HasTranscript":    search.HasTranscript,
		"Fts":              search.Fts,
		"AgentDescription": model.ReplaceWebSearch(search.AgentDescription),
		"OwnerIds":         pq.Array(search.OwnerIds),
		"GranteeIds":       pq.Array(search.GranteeIds),
		"ContactIds":       pq.Array(search.ContactIds),
		"AmdAiResult":      pq.Array(search.AmdAiResult),

		"TalkFrom": model.GetBetweenFrom(search.Talk),
		"TalkTo":   model.GetBetweenTo(search.Talk),

		"RatedUserIds":      pq.Array(search.RatedUserIds),
		"RatedByIds":        pq.Array(search.RatedByIds),
		"ScoreOptionalFrom": model.GetBetweenFrom(search.ScoreOptional),
		"ScoreOptionalTo":   model.GetBetweenTo(search.ScoreOptional),
		"ScoreRequiredFrom": model.GetBetweenFrom(search.ScoreRequired),
		"ScoreRequiredTo":   model.GetBetweenTo(search.ScoreRequired),
		"Rated":             search.Rated,
		"SchemaIds":         pq.Array(search.SchemaIds),
		"HasTransfer":       search.HasTransfer,
		"Timeline":          search.Timeline,
	}

	err := s.ListQueryTimeout(ctx, &out, search.ListRequest,
		`domain_id = :Domain::int8 
	and (:Q::text isnull or destination ilike :Q::text  or  from_number ilike :Q::text or  to_number ilike :Q::text or id::text = :Q::text)
	and (:Number::text isnull or coalesce(search_number, call_center.cc_array_to_string(array[destination, from_number, to_number], '|')) ~ :Number::text)
	and (:Variables::jsonb isnull or variables @> :Variables::jsonb)
	and ( :From::timestamptz isnull or created_at >= :From::timestamptz )
	and ( :To::timestamptz isnull or created_at <= :To::timestamptz )
	and ( (:StoredAtFrom::timestamptz isnull or :StoredAtTo::timestamptz isnull) or stored_at between :StoredAtFrom and :StoredAtTo )
	and (:UserIds::int8[] isnull or (user_id = any(:UserIds::int8[]) or user_ids::int[] && :UserIds::int[]))
	and (:OwnerIds::int8[] isnull or user_id = any(:OwnerIds::int8[]))
	and (:GranteeIds::int8[] isnull or grantee_id = any(:GranteeIds::int8[]))
	and (:ContactIds::int8[] isnull or contact_id = any(:ContactIds::int8[]))
	and (:Ids::uuid[] isnull or id = any(:Ids))
	and (:TransferFromIds::uuid[] isnull or transfer_from = any(:TransferFromIds))
	and (:TransferToIds::uuid[] isnull or transfer_to = any(:TransferToIds))
	and (:AmdResult::varchar[] isnull or amd_result = any(upper(:AmdResult::text)::varchar[]) or amd_ai_result = any(lower(:AmdResult::text)::varchar[]))
	and (:QueueIds::int[] isnull or (queue_id = any(:QueueIds) or queue_ids && :QueueIds::int[]) )
	and (:TeamIds::int[] isnull or (team_id = any(:TeamIds) or team_ids && :TeamIds::int[]) )  
	and (:AgentIds::int[] isnull or (agent_id = any(:AgentIds) or agent_ids && :AgentIds::int[]) )
	and (:MemberIds::int8[] isnull or member_id = any(:MemberIds) )
	and (:GatewayIds::int8[] isnull or (gateway_id = any(:GatewayIds) or gateway_ids::int[] && :GatewayIds::int4[]) )
	and ( (:SkipParent::bool isnull or not :SkipParent::bool is true ) or parent_id isnull)
	and ( (:Timeline::bool isnull or (:Timeline::bool and :DependencyIds::uuid[] notnull and parent_id notnull and ((transfer_from notnull and user_id notnull) or blind_transfer notnull  or parent_id != (:DependencyIds::uuid[])[1]))))
	and (:ParentId::uuid isnull or parent_id = :ParentId::uuid )
	and (:HasFile::bool isnull or (case :HasFile::bool when true then files notnull else files isnull end))
	and (:CauseArr::varchar[] isnull or cause = any(:CauseArr) )
	and ( (:AnsweredFrom::timestamptz isnull or :AnsweredTo::timestamptz isnull) or answered_at between :AnsweredFrom and :AnsweredTo )
	and ( (:DurationFrom::int8 isnull or :DurationFrom::int8 = 0 or duration >= :DurationFrom ))
	and ( (:DurationTo::int8 isnull or :DurationTo::int8 = 0 or duration <= :DurationTo ))
	and (:Direction::varchar isnull or direction = :Direction )
	and (:Missed::bool isnull or (:Missed and bridged_at isnull and (direction = 'inbound') and not hide_missed is true and not parent_bridged))
	and (:Tags::varchar[] isnull or (tags && :Tags))
	and (:AmdAiResult::varchar[] isnull or amd_ai_result = any(lower(:AmdAiResult::varchar[]::text)::varchar[]))
  	and ( :TalkFrom::int isnull or talk_sec >= :TalkFrom::int )
	and ( :TalkTo::int isnull or talk_sec <= :TalkTo::int )
	and ( :SchemaIds::int[] isnull or schema_ids && :SchemaIds::int[] )
	and (:HasTransfer::bool isnull or (case :HasTransfer::bool when true then (blind_transfer notnull or coalesce(transfer_to, transfer_from) notnull ) else (blind_transfer isnull and coalesce(transfer_to, transfer_from) isnull ) end))
	and (:AgentDescription::varchar isnull or
         (attempt_id notnull and exists(select 1 from call_center.cc_member_attempt_history cma where cma.id = attempt_id and cma.description ilike :AgentDescription::varchar))
         or (exists(select 1 from call_center.cc_calls_annotation ca where ca.call_id = t.id::text and ca.note ilike :AgentDescription::varchar))
    )
    and ((:HasTranscript::bool isnull and :Fts::varchar isnull) or (
        case :HasTranscript::bool when false
         then not exists(select 1 from storage.file_transcript ft where ft.uuid = t.id::text )
         else exists(select  1 from storage.file_transcript ft where ft.uuid = t.id::text and (:Fts::varchar isnull or to_tsvector(ft.transcript) @@ to_tsquery(:Fts::varchar)))
        end

    ))
	and (:DependencyIds::uuid[] isnull or id::uuid = any (
			array(with recursive a as (
                select d.id::uuid
                from call_center.cc_calls_history d
                where d.id::uuid = any(:DependencyIds::uuid[]) and d.domain_id = :Domain
                union all
                select d.id::uuid
                from call_center.cc_calls_history d, a
                where (d.parent_id::uuid = a.id::uuid or d.transfer_from::uuid = a.id::uuid)
            )
            select id::uuid ids
            from a
            where not a.id = any(:DependencyIds::uuid[]))
	))
	and ( (:Rated::bool isnull and :RatedUserIds::int8[] isnull and :RatedByIds::int8[] isnull and :ScoreOptionalFrom::numeric isnull
				and :ScoreOptionalTo::numeric isnull and :ScoreRequiredFrom::numeric isnull and :ScoreRequiredTo::numeric isnull ) or 
			case when not :Rated::bool then not exists(
					select 1
					from call_center.cc_audit_rate ar
					where ar.call_id = t.id::text) 
 				else exists(
					select 1
					from call_center.cc_audit_rate ar
					where ar.call_id = t.id::text
						and (:RatedUserIds::int8[] isnull or ar.rated_user_id = any(:RatedUserIds::int8[]))
						and (:RatedByIds::int8[] isnull or ar.created_by = any(:RatedByIds::int8[]))
						and ( :ScoreOptionalFrom::numeric isnull or score_optional >= :ScoreOptionalFrom::numeric )
						and ( :ScoreOptionalTo::numeric isnull or score_optional <= :ScoreOptionalTo::numeric )
						and ( :ScoreRequiredFrom::numeric isnull or score_required >= :ScoreRequiredFrom::numeric )
						and ( :ScoreRequiredTo::numeric isnull or score_required <= :ScoreRequiredTo::numeric )
				)
			 end
		)
`,
		model.HistoryCall{}, f)
	if err != nil {
		return nil, model.NewInternalError("store.sql_call.get_history.app_error", err.Error())
	}

	return out, nil
}

func (s SqlCallStore) GetHistoryByGroups(ctx context.Context, domainId int64, userSupervisorId int64, groups []int, search *model.SearchHistoryCall) ([]*model.HistoryCall, model.AppError) {
	var out []*model.HistoryCall

	f := map[string]interface{}{
		"Domain":           domainId,
		"Limit":            search.GetLimit(),
		"Offset":           search.GetOffset(),
		"From":             model.GetBetweenFromTime(search.CreatedAt),
		"To":               model.GetBetweenToTime(search.CreatedAt),
		"Q":                search.GetQ(),
		"UserIds":          pq.Array(search.UserIds),
		"QueueIds":         pq.Array(search.QueueIds),
		"TeamIds":          pq.Array(search.TeamIds),
		"AgentIds":         pq.Array(search.AgentIds),
		"MemberIds":        pq.Array(search.MemberIds),
		"GatewayIds":       pq.Array(search.GatewayIds),
		"SkipParent":       search.SkipParent,
		"ParentId":         search.ParentId,
		"Number":           model.GetRegExpQ(search.Number),
		"CauseArr":         pq.Array(search.CauseArr),
		"HasFile":          search.HasFile,
		"Direction":        search.Direction,
		"Missed":           search.Missed,
		"AnsweredFrom":     model.GetBetweenFromTime(search.AnsweredAt),
		"AnsweredTo":       model.GetBetweenToTime(search.AnsweredAt),
		"DurationFrom":     model.GetBetweenFrom(search.Duration),
		"DurationTo":       model.GetBetweenTo(search.Duration),
		"StoredAtFrom":     model.GetBetweenFromTime(search.StoredAt),
		"StoredAtTo":       model.GetBetweenToTime(search.StoredAt),
		"Ids":              pq.Array(search.Ids),
		"TransferFromIds":  pq.Array(search.TransferFromIds),
		"TransferToIds":    pq.Array(search.TransferToIds),
		"DependencyIds":    pq.Array(search.DependencyIds),
		"Tags":             pq.Array(search.Tags),
		"AmdResult":        pq.Array(search.AmdResult),
		"Groups":           pq.Array(groups),
		"Access":           auth_manager.PERMISSION_ACCESS_READ.Value(),
		"UserSupervisorId": userSupervisorId,
		"Variables":        search.Variables.ToSafeJson(),
		"HasTranscript":    search.HasTranscript,
		"Fts":              search.Fts,
		"AgentDescription": model.ReplaceWebSearch(search.AgentDescription),
		"OwnerIds":         pq.Array(search.OwnerIds),
		"GranteeIds":       pq.Array(search.GranteeIds),
		"ContactIds":       pq.Array(search.ContactIds),
		"AmdAiResult":      pq.Array(search.AmdAiResult),

		"TalkFrom": model.GetBetweenFrom(search.Talk),
		"TalkTo":   model.GetBetweenTo(search.Talk),

		"RatedUserIds":      pq.Array(search.RatedUserIds),
		"RatedByIds":        pq.Array(search.RatedByIds),
		"ScoreOptionalFrom": model.GetBetweenFrom(search.ScoreOptional),
		"ScoreOptionalTo":   model.GetBetweenTo(search.ScoreOptional),
		"ScoreRequiredFrom": model.GetBetweenFrom(search.ScoreRequired),
		"ScoreRequiredTo":   model.GetBetweenTo(search.ScoreRequired),
		"Rated":             search.Rated,
		"SchemaIds":         pq.Array(search.SchemaIds),
		"HasTransfer":       search.HasTransfer,
		"Timeline":          search.Timeline,

		"ClassName": model.PERMISSION_SCOPE_CALL,
	}

	err := s.ListQueryTimeout(ctx, &out, search.ListRequest,
		`domain_id = :Domain::int8 
	and (:Q::text isnull or destination ilike :Q::text  or  from_number ilike :Q::text or  to_number ilike :Q::text or id::text = :Q::text)
	and (:Number::text isnull or coalesce(search_number, call_center.cc_array_to_string(array[destination, from_number, to_number], '|')) ~ :Number::text)
	and (:Variables::jsonb isnull or variables @> :Variables::jsonb)
	and ( :From::timestamptz isnull or created_at >= :From::timestamptz )
	and ( :To::timestamptz isnull or created_at <= :To::timestamptz )
	and ( (:StoredAtFrom::timestamptz isnull or :StoredAtTo::timestamptz isnull) or stored_at between :StoredAtFrom and :StoredAtTo )
	and (:UserIds::int8[] isnull or (user_id = any(:UserIds::int8[]) or user_ids::int[] && :UserIds::int[]))
	and (:OwnerIds::int8[] isnull or user_id = any(:OwnerIds::int8[]))
	and (:GranteeIds::int8[] isnull or grantee_id = any(:GranteeIds::int8[]))
	and (:ContactIds::int8[] isnull or contact_id = any(:ContactIds::int8[]))
	and (:Ids::uuid[] isnull or id = any(:Ids))
	and (:TransferFromIds::uuid[] isnull or transfer_from = any(:TransferFromIds))
	and (:TransferToIds::uuid[] isnull or transfer_to = any(:TransferToIds))
	and (:AmdResult::varchar[] isnull or amd_result = any(upper(:AmdResult::text)::varchar[]) or amd_ai_result = any(lower(:AmdResult::text)::varchar[]))
	and (:QueueIds::int[] isnull or (queue_id = any(:QueueIds) or queue_ids && :QueueIds::int[]) )
	and (:TeamIds::int[] isnull or (team_id = any(:TeamIds) or team_ids && :TeamIds::int[]) )  
	and (:AgentIds::int[] isnull or (agent_id = any(:AgentIds) or agent_ids && :AgentIds::int[]) )
	and (:MemberIds::int8[] isnull or member_id = any(:MemberIds) )
	and (:GatewayIds::int8[] isnull or (gateway_id = any(:GatewayIds) or gateway_ids::int[] && :GatewayIds::int4[]) )
	and ( (:SkipParent::bool isnull or not :SkipParent::bool is true ) or parent_id isnull)
	and (:ParentId::uuid isnull or parent_id = :ParentId::uuid )
	and ( (:Timeline::bool isnull or (:Timeline::bool and :DependencyIds::uuid[] notnull and parent_id notnull and ((transfer_from notnull and user_id notnull) or blind_transfer notnull  or parent_id != (:DependencyIds::uuid[])[1]))))
	and (:HasFile::bool isnull or (case :HasFile::bool when true then files notnull else files isnull end))
	and (:CauseArr::varchar[] isnull or cause = any(:CauseArr) )
	and ( (:AnsweredFrom::timestamptz isnull or :AnsweredTo::timestamptz isnull) or answered_at between :AnsweredFrom and :AnsweredTo )
	and ( (:DurationFrom::int8 isnull or :DurationFrom::int8 = 0 or duration >= :DurationFrom ))
	and ( (:DurationTo::int8 isnull or :DurationTo::int8 = 0 or duration <= :DurationTo ))
	and (:Direction::varchar isnull or direction = :Direction )
	and (:Missed::bool isnull or (:Missed and bridged_at isnull and (direction = 'inbound') and not hide_missed is true and not parent_bridged))
	and (:Tags::varchar[] isnull or (tags && :Tags))
	and (:AmdAiResult::varchar[] isnull or amd_ai_result = any(lower(:AmdAiResult::varchar[]::text)::varchar[]))
  	and ( :TalkFrom::int isnull or talk_sec >= :TalkFrom::int )
	and ( :TalkTo::int isnull or talk_sec <= :TalkTo::int )
	and ( :SchemaIds::int[] isnull or schema_ids && :SchemaIds::int[] )
	and (:HasTransfer::bool isnull or (case :HasTransfer::bool when true then (blind_transfer notnull or coalesce(transfer_to, transfer_from) notnull ) else (blind_transfer isnull and coalesce(transfer_to, transfer_from) isnull ) end))
	and (:AgentDescription::varchar isnull or
         (attempt_id notnull and exists(select 1 from call_center.cc_member_attempt_history cma where cma.id = attempt_id and cma.description ilike :AgentDescription::varchar))
         or (exists(select 1 from call_center.cc_calls_annotation ca where ca.call_id = t.id::text and ca.note ilike :AgentDescription::varchar))
    )
    and ((:HasTranscript::bool isnull and :Fts::varchar isnull) or (
        case :HasTranscript::bool when false
         then not exists(select 1 from storage.file_transcript ft where ft.uuid = t.id::text )
         else exists(select  1 from storage.file_transcript ft where ft.uuid = t.id::text and (:Fts::varchar isnull or to_tsvector(ft.transcript) @@ to_tsquery(:Fts::varchar)))
        end

    ))
	and (:DependencyIds::uuid[] isnull or id::uuid = any (
			array(with recursive a as (
                select d.id::uuid
                from call_center.cc_calls_history d
                where d.id::uuid = any(:DependencyIds::uuid[]) and d.domain_id = :Domain
                union all
                select d.id::uuid
                from call_center.cc_calls_history d, a
                where (d.parent_id::uuid = a.id::uuid or d.transfer_from::uuid = a.id::uuid)
            )
            select id::uuid ids
            from a
            where not a.id = any(:DependencyIds::uuid[]))
	))
	and ( (:Rated::bool isnull and :RatedUserIds::int8[] isnull and :RatedByIds::int8[] isnull and :ScoreOptionalFrom::numeric isnull
				and :ScoreOptionalTo::numeric isnull and :ScoreRequiredFrom::numeric isnull and :ScoreRequiredTo::numeric isnull ) or 
			case when not :Rated::bool then not exists(
					select 1
					from call_center.cc_audit_rate ar
					where ar.call_id = t.id::text) 
 				else exists(
					select 1
					from call_center.cc_audit_rate ar
					where ar.call_id = t.id::text
						and (:RatedUserIds::int8[] isnull or ar.rated_user_id = any(:RatedUserIds::int8[]))
						and (:RatedByIds::int8[] isnull or ar.created_by = any(:RatedByIds::int8[]))
						and ( :ScoreOptionalFrom::numeric isnull or score_optional >= :ScoreOptionalFrom::numeric )
						and ( :ScoreOptionalTo::numeric isnull or score_optional <= :ScoreOptionalTo::numeric )
						and ( :ScoreRequiredFrom::numeric isnull or score_required >= :ScoreRequiredFrom::numeric )
						and ( :ScoreRequiredTo::numeric isnull or score_required <= :ScoreRequiredTo::numeric )
				)
			 end
		)
	and (
		(t.user_id = any (call_center.cc_calls_rbac_users(:Domain::int8, :UserSupervisorId::int8) || :Groups::int[])
			or t.queue_id = any (call_center.cc_calls_rbac_queues(:Domain::int8, :UserSupervisorId::int8, :Groups::int[]))
			or (t.user_ids notnull and t.user_ids::int[] && call_center.rbac_users_from_group(:ClassName::varchar, :Domain::int8, :Access::int2, :Groups::int[]))
			or (t.grantee_id = any(:Groups::int[]))
		)
	)
`,
		model.HistoryCall{}, f)
	if err != nil {
		return nil, model.NewInternalError("store.sql_call.get_history.app_error", err.Error())
	}

	return out, nil
}

func (s SqlCallStore) SetVariables(ctx context.Context, domainId int64, id string, vars model.StringMap) (*model.CallDomain, model.AppError) {
	var res *model.CallDomain
	err := s.GetMaster().WithContext(ctx).SelectOne(&res, `with a as (
    update call_center.cc_calls c
        set payload = coalesce(payload, '{}') || :Vars
    where c.id = :Id::uuid and c.domain_id = :DomainId
    returning c.id, c.app_id
), h as (
    update call_center.cc_calls_history c
        set payload = coalesce(payload, '{}') || :Vars
    where c.id = :Id::uuid and c.domain_id = :DomainId
    returning c.id
)
select *
from (
    select id, app_id
    from a
    union all
    select id,  null
    from h
 ) as t
limit 1`, map[string]interface{}{
		"DomainId": domainId,
		"Id":       id,
		"Vars":     vars.ToJson(),
	})

	if err != nil {
		return nil, model.NewCustomCodeError("store.sql_call.set_vars.app_error", err.Error(), extractCodeFromErr(err))
	}

	return res, nil
}

func (s SqlCallStore) CreateAnnotation(ctx context.Context, annotation *model.CallAnnotation) (*model.CallAnnotation, model.AppError) {
	err := s.GetMaster().WithContext(ctx).SelectOne(&annotation, `
		with a as (
			insert into call_center.cc_calls_annotation (call_id, created_by, created_at, note, start_sec, end_sec, updated_by, updated_at)
			values (:CallId, :CreatedBy, :CreatedAt, :Note, :StartSec, :EndSec, :UpdatedBy, :UpdatedAt)
			returning *
		)
		select
			a.id,
			a.call_id,
			a.created_at,
			call_center.cc_get_lookup(cc.id, coalesce(cc.name, cc.username)) created_by,
			a.updated_at,
			call_center.cc_get_lookup(uc.id, coalesce(uc.name, uc.username)) updated_by,
			a.note,
			a.start_sec,
			a.end_sec
		from a
			left join directory.wbt_user cc on cc.id = a.created_by
			left join directory.wbt_user uc on uc.id = a.updated_by;
`, map[string]interface{}{
		"CallId":    annotation.CallId,
		"CreatedBy": annotation.CreatedBy.GetSafeId(),
		"CreatedAt": annotation.CreatedAt,
		"Note":      annotation.Note,
		"StartSec":  annotation.StartSec,
		"EndSec":    annotation.EndSec,
		"UpdatedBy": annotation.UpdatedBy.GetSafeId(),
		"UpdatedAt": annotation.UpdatedAt,
	})

	if err != nil {
		return nil, model.NewCustomCodeError("store.sql_call.annotation.create.app_error", err.Error(), extractCodeFromErr(err))
	}

	return annotation, nil
}

func (s SqlCallStore) GetAnnotation(ctx context.Context, id int64) (*model.CallAnnotation, model.AppError) {
	var annotation *model.CallAnnotation
	err := s.GetReplica().WithContext(ctx).SelectOne(&annotation, `
select
    a.id,
    a.call_id,
    a.created_at,
    call_center.cc_get_lookup(cc.id, coalesce(cc.name, cc.username)) created_by,
    a.updated_at,
    call_center.cc_get_lookup(uc.id, coalesce(uc.name, uc.username)) updated_by,
    a.note,
    a.start_sec,
    a.end_sec
from call_center.cc_calls_annotation a
    left join directory.wbt_user cc on cc.id = a.created_by
    left join directory.wbt_user uc on uc.id = a.updated_by
where a.id = :Id`, map[string]interface{}{
		"Id": id,
	})

	if err != nil {
		return nil, model.NewCustomCodeError("store.sql_call.annotation.get.app_error", err.Error(), extractCodeFromErr(err))
	}

	return annotation, nil
}

func (s SqlCallStore) UpdateAnnotation(ctx context.Context, domainId int64, annotation *model.CallAnnotation) (*model.CallAnnotation, model.AppError) {
	err := s.GetMaster().WithContext(ctx).SelectOne(&annotation, `
		with a as (
			update call_center.cc_calls_annotation
				set updated_at = :UpdatedAt,
					updated_by = :UpdatedBy,
					note = :Note,
					start_sec = :StartSec,
					end_sec = :EndSec
			where id = :Id
			returning *
		)
		select
			a.id,
			a.call_id,
			a.created_at,
			call_center.cc_get_lookup(cc.id, coalesce(cc.name, cc.username)) created_by,
			a.updated_at,
			call_center.cc_get_lookup(uc.id, coalesce(uc.name, uc.username)) updated_by,
			a.note,
			a.start_sec,
			a.end_sec
		from a
			left join directory.wbt_user cc on cc.id = a.created_by
			left join directory.wbt_user uc on uc.id = a.updated_by
`, map[string]interface{}{
		"Id":        annotation.Id,
		"UpdatedAt": annotation.UpdatedAt,
		"UpdatedBy": annotation.UpdatedBy.GetSafeId(),
		"Note":      annotation.Note,
		"StartSec":  annotation.StartSec,
		"EndSec":    annotation.EndSec,
	})

	if err != nil {
		return nil, model.NewCustomCodeError("store.sql_call.annotation.update.app_error", err.Error(), extractCodeFromErr(err))
	}

	return annotation, nil
}

func (s SqlCallStore) DeleteAnnotation(ctx context.Context, id int64) model.AppError {
	_, err := s.GetMaster().WithContext(ctx).Exec(`delete from call_center.cc_calls_annotation where id = :Id`, map[string]interface{}{
		"Id": id,
	})

	if err != nil {
		return model.NewCustomCodeError("store.sql_call.annotation.delete.app_error", err.Error(), extractCodeFromErr(err))
	}

	return nil
}

func AggregateField(group *model.AggregateGroup) string {
	if group.Interval != "" {
		return fmt.Sprintf("to_timestamp(extract(epoch from (date_trunc('seconds', (%s - timestamptz 'epoch') / EXTRACT(EPOCH FROM INTERVAL %s)) * EXTRACT(EPOCH FROM INTERVAL %s) + timestamptz 'epoch')))",
			QuoteIdentifier(group.Id), QuoteLiteral(group.Interval), QuoteLiteral(group.Interval))
	}

	if strings.HasPrefix(group.Id, "variables.") {
		return "payload->>" + QuoteLiteral(group.Id[10:])
	}

	return QuoteIdentifier(group.Id)
}

// TODO
func GroupData(groups []model.AggregateGroup) string {
	if len(groups) < 1 {
		return ""
	}
	sql := "group by "
	for i, v := range groups {
		if i > 0 {
			sql += ", "
		}
		sql += AggregateField(&v)
	}

	return sql
}

func GroupWhere(table string, groups []model.AggregateGroup) string {
	if len(groups) < 1 {
		return ""
	}

	where := make([]string, 0, 1)
	for _, v := range groups {
		id := ""
		if strings.HasPrefix(v.Id, "variables.") {
			id = "payload->>" + QuoteLiteral(v.Id[10:])
		} else {
			id = QuoteIdentifier(v.Id)
		}

		order := ""

		switch v.Aggregate {
		case "min":
			order = fmt.Sprintf("min(%s)", QuoteIdentifier(v.Field))
		case "max":
			order = fmt.Sprintf("max(%s)", QuoteIdentifier(v.Field))
		case "avg":
			order = fmt.Sprintf("avg(%s)", QuoteIdentifier(v.Field))
		case "sum":
			order = fmt.Sprintf("sum(%s)", QuoteIdentifier(v.Field))
		case "count":
			if v.Field == "" {
				order = "count(*)"
			} else {
				order = fmt.Sprintf("count(%s)", QuoteIdentifier(v.Field))
			}
		default:
			continue
		}

		if v.Desc {
			order += " desc"
		}

		where = append(where, fmt.Sprintf(`%s in (select
				%s
			from %s
			where %s notnull
			group by 1
			order by %s
			limit %d)`, id, id, QuoteIdentifier(table), id, order, v.Top))
	}

	if len(where) == 0 {
		return ""
	}

	return "where " + strings.Join(where, " and")
}

func TimeHistogram(dateRange *model.FilterBetween, group *model.AggregateGroup) string {
	if dateRange == nil || group == nil {
		return ""
	}

	return fmt.Sprintf("right join generate_series(%s::timestamptz, %s::timestamptz, interval %s) x on (l.created_at between x and (x + interval %s - interval '1 sec'))",
		QuoteLiteral(model.GetBetweenFromTime(dateRange).Format("2006-01-02 15:04:05")), QuoteLiteral(model.GetBetweenToTime(dateRange).Format("2006-01-02 15:04:05")),
		QuoteLiteral(group.Interval), QuoteLiteral(group.Interval))
}

func (s SqlCallStore) ParseAgg(histogramRange *model.FilterBetween, table string, agg *model.Aggregate) string {
	fields := []string{}
	results := []string{}

	var sql string
	var histogramField *model.AggregateGroup

	for _, v := range agg.Group {
		fields = append(fields, fmt.Sprintf("%s as %s", AggregateField(&v), QuoteIdentifier(v.Id)))

		if v.Interval != "" && histogramRange != nil {
			histogramField = new(model.AggregateGroup)
			*histogramField = v
			results = append(results, fmt.Sprintf("x as %s", QuoteIdentifier(v.Id)))
		} else {
			results = append(results, QuoteIdentifier(v.Id))
		}
	}

	for _, v := range agg.Sum {
		fields = append(fields, "sum("+QuoteIdentifier(v)+") as "+QuoteIdentifier("sum_"+v))
		results = append(results, QuoteIdentifier("sum_"+v))
	}
	for _, v := range agg.Avg {
		fields = append(fields, "avg("+QuoteIdentifier(v)+") as "+QuoteIdentifier("avg_"+v))
		results = append(results, QuoteIdentifier("avg_"+v))
	}
	for _, v := range agg.Max {
		fields = append(fields, "max("+QuoteIdentifier(v)+") as "+QuoteIdentifier("max_"+v))
		results = append(results, QuoteIdentifier("max_"+v))
	}
	for _, v := range agg.Min {
		fields = append(fields, "min("+QuoteIdentifier(v)+") as "+QuoteIdentifier("min_"+v))
		results = append(results, QuoteIdentifier("min_"+v))
	}

	for _, v := range agg.Count {
		if v == "*" {
			fields = append(fields, "count(*) as count")
			results = append(results, "count")
		} else {
			fields = append(fields, "count("+QuoteIdentifier(v)+") as "+QuoteIdentifier("count_"+v))
			results = append(results, QuoteIdentifier("count_"+v))
		}
	}

	if len(fields) < 1 {
		//todo error
	}

	sql = `select json_agg(row_to_json(t)) as data
    from (
		select *
		from (
			select ` + strings.Join(results, ", ") + `
			from (
          		select ` + strings.Join(fields, ", ") + `
          		from ` + table + `
				` + GroupWhere(table, agg.Group) + `	
		  		` + GroupData(agg.Group) + `
			) l
			` + TimeHistogram(histogramRange, histogramField) + `
		) t
		` + GetOrderArrayBy(agg.Sort) + `
        limit %d 
    ) t`

	return fmt.Sprintf(sql, model.GetLimit(agg.Limit))
}

func GetOrderArrayBy(s []string) string {
	if len(s) == 0 {
		return ""
	}

	order := make([]string, 0, len(s))

	for _, v := range s {
		switch v[0] {
		case '+':
			order = append(order, QuoteIdentifier(v[1:])+" asc")
		case '-':
			order = append(order, QuoteIdentifier(v[1:])+" desc")
		default:
			order = append(order, QuoteIdentifier(v))
		}
	}

	return "order by " + strings.Join(order, ",")
}

func (s SqlCallStore) Aggregate(ctx context.Context, domainId int64, aggs *model.CallAggregate) ([]*model.AggregateResult, model.AppError) {

	/*
		todo materialized ??
	*/
	sql := `with calls as materialized (
    select h.id,
		   h.hold_sec,
		   h.agent_id,
		   extract(EPOCH from h.hangup_at - h.created_at)::int duration,
		   case when h.answered_at notnull then extract(EPOCH from h.hangup_at - h.created_at)::int end answer_sec,
		   case when h.bridged_at notnull then extract(EPOCH from h.hangup_at - h.bridged_at)::int else 0 end bill,
		   case when h.bridged_at notnull then true else false end bridged,
		   h.created_at,
		   h.answered_at,
		   h.bridged_at,
		   h.hangup_at,
		   h.hangup_by,
		   h.user_id,
		   h.payload,
		   coalesce(u.name, u.username) as user,
		   h.direction,
		   h.gateway_id,
		   g.name as gateway,
		   h.team_id,
		   t.name team,
		   coalesce(ua.name, ua.username) agent,
		   h.cause,
		   h.sip_code,
		   h.queue_id,
		   q.name as queue,
		   h.tags	
	from call_center.cc_calls_history h
		left join call_center.cc_agent ca on h.agent_id = ca.id
		left join directory.wbt_user ua on ua.id = ca.user_id
		left join directory.wbt_user u on u.id = h.user_id
		left join directory.sip_gateway g on g.id = h.gateway_id
		left join call_center.cc_queue q on q.id = h.queue_id
		left join call_center.cc_team t on t.id = h.team_id
	where h.domain_id = :Domain 
		and (:Q::text isnull or h.destination ilike :Q::text  or  h.from_number ilike :Q::text or  h.to_number ilike :Q::text or h.id::text = :Q::text)
		and (:Number::text isnull or h.from_number ~ :Number::text or h.to_number ~ :Number::text or h.destination ~ :Number::text)
		and ( (:From::timestamptz isnull or :To::timestamptz isnull) or h.created_at between :From and :To )
		and ( (:StoredAtFrom::timestamptz isnull or :StoredAtTo::timestamptz isnull) or h.stored_at between :StoredAtFrom and :StoredAtTo )
		and (:UserIds::int8[] isnull or h.user_id = any(:UserIds))
		and (:Ids::uuid[] isnull or h.id = any(:Ids))
		and (:TransferFromIds::uuid[] isnull or h.transfer_from = any(:TransferFromIds))
		and (:TransferToIds::uuid[] isnull or h.transfer_to = any(:TransferToIds))
		and (:QueueIds::int[] isnull or h.queue_id = any(:QueueIds) )
		and (:ContactIds::int8[] isnull or h.contact_id = any(:ContactIds::int8[]))
		and (:TeamIds::int[] isnull or h.team_id = any(:TeamIds) )  
		and (:AgentIds::int[] isnull or h.agent_id = any(:AgentIds) )
		and (:MemberIds::int8[] isnull or h.member_id = any(:MemberIds) )
		and (:GatewayIds::int8[] isnull or h.gateway_id = any(:GatewayIds) )
		and ( (:SkipParent::bool isnull or not :SkipParent::bool is true ) or h.parent_id isnull)
		and (:ParentId::uuid isnull or h.parent_id = :ParentId )
		and (:CauseArr::varchar[] isnull or h.cause = any(:CauseArr) )
		and ( (:AnsweredFrom::timestamptz isnull or :AnsweredTo::timestamptz isnull) or h.answered_at between :AnsweredFrom and :AnsweredTo )
		and (:Directions::varchar[] isnull or h.direction = any(:Directions) )
		and (:Direction::varchar isnull or h.direction = :Direction )
		and (:Missed::bool isnull or (:Missed and h.answered_at isnull))
		and (:Tags::varchar[] isnull or (h.tags && :Tags))
		and (:AgentDescription::varchar isnull or (attempt_id notnull and exists(select 1 from call_center.cc_member_attempt_history cma where cma.id = attempt_id and cma.description ilike :AgentDescription::varchar)))
		and (:AmdResult::varchar[] isnull or h.amd_result = any(:AmdResult))
		and (:HasFile::bool isnull or (case :HasFile::bool when true then exists(select 1 from storage.files ft where ft.uuid = h.id::text ) else not exists(select 1 from storage.files ft where ft.uuid = h.id::text ) end))
		and ((:HasTranscript::bool isnull and :Fts::varchar isnull) or (
				case :HasTranscript::bool when false
				 then not exists(select 1 from storage.file_transcript ft where ft.uuid = h.id::text )
				 else exists(select  1 from storage.file_transcript ft where ft.uuid = h.id::text and (:Fts::varchar isnull or to_tsvector(ft.transcript) @@ to_tsquery(:Fts::varchar)))
				end
		
			))
		and (:DependencyIds::uuid[] isnull or h.id in (
			with recursive a as (
				select t.id
				from call_center.cc_calls_history t
				where t.id = any(:DependencyIds::uuid[])
				union all
				select t.id
				from call_center.cc_calls_history t, a
				where t.parent_id = a.id or t.transfer_from = a.id
			)
			select a.id
			from a
			where not a.id = any(:DependencyIds::uuid[])
		))
)
`

	for _, v := range aggs.Aggs {
		sql += `, ` + QuoteIdentifier(v.Name) + ` as (` + s.ParseAgg(aggs.CreatedAt, "calls", &v) + `) `
	}

	f := map[string]interface{}{
		"Domain":          domainId,
		"Limit":           aggs.GetLimit(),
		"Offset":          aggs.GetOffset(),
		"From":            model.GetBetweenFromTime(aggs.CreatedAt),
		"To":              model.GetBetweenToTime(aggs.CreatedAt),
		"Q":               aggs.GetQ(),
		"UserIds":         pq.Array(aggs.UserIds),
		"QueueIds":        pq.Array(aggs.QueueIds),
		"TeamIds":         pq.Array(aggs.TeamIds),
		"AgentIds":        pq.Array(aggs.AgentIds),
		"MemberIds":       pq.Array(aggs.MemberIds),
		"GatewayIds":      pq.Array(aggs.GatewayIds),
		"SkipParent":      aggs.SkipParent,
		"ParentId":        aggs.ParentId,
		"Number":          model.GetRegExpQ(aggs.Number),
		"CauseArr":        pq.Array(aggs.CauseArr),
		"Directions":      pq.Array(aggs.Directions),
		"Direction":       aggs.Direction,
		"Missed":          aggs.Missed,
		"AnsweredFrom":    model.GetBetweenFromTime(aggs.AnsweredAt),
		"AnsweredTo":      model.GetBetweenToTime(aggs.AnsweredAt),
		"DurationFrom":    model.GetBetweenFrom(aggs.Duration),
		"DurationTo":      model.GetBetweenTo(aggs.Duration),
		"StoredAtFrom":    model.GetBetweenFromTime(aggs.StoredAt),
		"StoredAtTo":      model.GetBetweenToTime(aggs.StoredAt),
		"Ids":             pq.Array(aggs.Ids),
		"TransferFromIds": pq.Array(aggs.TransferFromIds),
		"TransferToIds":   pq.Array(aggs.TransferToIds),
		"DependencyIds":   pq.Array(aggs.DependencyIds),
		"Tags":            pq.Array(aggs.Tags),
		"ContactIds":      pq.Array(aggs.ContactIds),

		"AmdResult":        pq.Array(aggs.AmdResult),
		"HasFile":          aggs.HasFile,
		"HasTranscript":    aggs.HasTranscript,
		"Fts":              aggs.Fts,
		"AgentDescription": model.ReplaceWebSearch(aggs.AgentDescription),
	}

	for i, v := range aggs.Aggs {
		if i > 0 {
			sql += "union all "
		}
		sql += "select " + QuoteLiteral(v.Name) + " as name, (select data from " + QuoteIdentifier(v.Name) + ") as data "
	}

	var res []*model.AggregateResult

	_, err := s.GetReplica().WithContext(ctx).Select(&res, sql, f)
	if err != nil {
		return nil,
			model.NewCustomCodeError("store.sql_call.aggregate.app_error", err.Error(), extractCodeFromErr(err))
	}

	return res, nil
}

func (s SqlCallStore) BridgeInfo(ctx context.Context, domainId int64, fromId, toId string) (*model.BridgeCall, model.AppError) {
	var res *model.BridgeCall
	err := s.GetMaster().WithContext(ctx).SelectOne(&res, `select coalesce(c.bridged_id, c.id) from_id, coalesce(c2.bridged_id, c2.id) to_id, 
       c.app_id, c.contact_id
from call_center.cc_calls c,
     call_center.cc_calls c2
where c.id = :FromId::uuid and c2.id = :ToId::uuid and c.domain_id = :DomainId and c2.domain_id = :DomainId`, map[string]interface{}{
		"DomainId": domainId,
		"FromId":   fromId,
		"ToId":     toId,
	})
	if err != nil {
		return nil, model.NewCustomCodeError("store.sql_call.get_bridge_info.app_error", err.Error(), extractCodeFromErr(err))
	} else {
		return res, nil
	}
}

func (s SqlCallStore) LastFile(ctx context.Context, domainId int64, id string) (int64, model.AppError) {
	fileId, err := s.GetReplica().WithContext(ctx).SelectInt(`select f.id
from storage.files f
where f.domain_id = :DomainId and f.uuid = (
    select coalesce(c.parent_id, c.id)::text
    from call_center.cc_calls_history c
    where c.id = :Id::uuid and c.domain_id = :DomainId
    limit 1
)`, map[string]interface{}{
		"DomainId": domainId,
		"Id":       id,
	})

	if err != nil {
		return 0, model.NewCustomCodeError("store.sql_call.get_last_file.app_error", err.Error(), extractCodeFromErr(err))
	}

	return fileId, nil
}

func (s SqlCallStore) BridgedId(ctx context.Context, id string) (string, model.AppError) {
	res, err := s.GetMaster().WithContext(ctx).SelectStr(`select coalesce(c.bridged_id, c.parent_id, c.id)
from call_center.cc_calls c
where id = :Id::uuid`, map[string]string{
		"Id": id,
	})

	if err != nil {
		return "", model.NewCustomCodeError("store.sql_call.get_bridge_id.app_error", err.Error(), extractCodeFromErr(err))
	} else {
		return res, nil
	}
}

func (s SqlCallStore) BlindTransferInfo(ctx context.Context, id string) (*model.BlindTransferInfo, model.AppError) {
	var res *model.BlindTransferInfo
	err := s.GetMaster().WithContext(ctx).SelectOne(&res, `select coalesce(c.bridged_id, c.parent_id, c.id) as id, c.contact_id,  
       	 (c.answered_at isnull and c.queue_id notnull ) queue_unanswered
from call_center.cc_calls c
where id = :Id::uuid`, map[string]string{
		"Id": id,
	})

	if err != nil {
		return nil, model.NewCustomCodeError("store.sql_call.blind_transfer.app_error", err.Error(), extractCodeFromErr(err))
	} else {
		return res, nil
	}
}

func (s SqlCallStore) TransferInfo(ctx context.Context, id string, domainId int64, queueId *int, agentId *int) (*model.TransferInfo, model.AppError) {
	var res *model.TransferInfo
	err := s.GetMaster().WithContext(ctx).SelectOne(&res, `select coalesce(c.bridged_id, c.parent_id, c.id) as  id,
       c.contact_id,
       (c.answered_at isnull and c.queue_id notnull) queue_unanswered,
       c.app_id,
       (select q.name from call_center.cc_queue q where q.id = :QueueId::int and q.enabled) queue_name,
       a.agent_name,
       a.extension as agent_extension
from call_center.cc_calls c
    left join (
        select coalesce(u.name::text, u.username) agent_name, u.extension
        from call_center.cc_agent a
            inner join directory.wbt_user u on u.id = a.user_id
        where a.id = :AgentId::int
    ) a on :AgentId::int notnull
where id = :Id::uuid and c.domain_id = :DomainId::int8;`, map[string]any{
		"Id":       id,
		"DomainId": domainId,
		"QueueId":  queueId,
		"AgentId":  agentId,
	})

	if err != nil {
		return nil, model.NewCustomCodeError("store.sql_call.transfer_info.app_error", err.Error(), extractCodeFromErr(err))
	} else {
		return res, nil
	}
}

func (s SqlCallStore) SetEmptySeverCall(ctx context.Context, domainId int64, id string) (*model.CallServiceHangup, model.AppError) {
	var e *model.CallServiceHangup
	err := s.GetMaster().WithContext(ctx).SelectOne(&e, `with c as (
    select
        c.id,
       call_center.cc_view_timestamp(now())::text as "timestamp",
       c.domain_id::text,
       coalesce(c.user_id::text, '') as user_id,
       c.app_id,
       coalesce(cma.node_id, '') as cc_app_id
    from  call_center.cc_calls c
        left join call_center.cc_member_attempt cma on c.attempt_id = cma.id
    where c.id = :Id::uuid and c.domain_id = :DomainId and c.hangup_at isnull
    and c.timestamp < now() - interval '15 sec' and c.hangup_by isnull
)
update call_center.cc_calls c1
set hangup_by = 'service'
from c
where c.id = c1.id
returning c.*;`, map[string]interface{}{
		"Id":       id,
		"DomainId": domainId,
	})

	if err != nil {
		return nil, model.NewCustomCodeError("store.sql_call.set.empty_call.app_error", err.Error(), extractCodeFromErr(err))
	} else {
		return e, nil
	}
}

func (s SqlCallStore) GetEavesdropInfo(ctx context.Context, domainId int64, id string) (*model.EavesdropInfo, model.AppError) {
	var res *model.EavesdropInfo

	err := s.GetMaster().WithContext(ctx).SelectOne(&res, `select
    case when owner_agent.v then c.id else c.bridged_id end agent_call_id,
    c.id as parent_id,
    c.app_id,
    case when owner_agent.v then f.f else t.t end agent,
    case when not owner_agent.v then f.f else t.t end client,
	extract(epoch from now() - coalesce(c.bridged_at, c.created_at)  )::int8 as duration
from call_center.cc_calls c
    left join lateral (select not(c.bridged_id notnull and c.user_id isnull) v) owner_agent on true
    left join lateral (select json_build_object('type', coalesce(c.from_type, ''), 'number', coalesce(c.from_number, ''), 'id', coalesce(c.from_id, ''), 'name', coalesce(c.from_name, '')) f) as f on true
    left join lateral (select json_build_object('type', coalesce(c.to_type, ''), 'number', coalesce(c.to_number, ''), 'id', coalesce(c.to_id, ''), 'name', coalesce(c.to_name, '')) t) as t on true
where c.domain_id = :DomainId and c.id = :Id::uuid and c.state in ('active', 'bridge', 'eavesdrop')`, map[string]interface{}{
		"DomainId": domainId,
		"Id":       id,
	})

	if err != nil {
		return nil, model.NewCustomCodeError("store.sql_call.get.eavesdrop_info.app_error", err.Error(), extractCodeFromErr(err))
	}

	return res, nil
}

func (s SqlCallStore) GetOwnerUserCall(ctx context.Context, id string) (*int64, time.Time, model.AppError) {
	var userId *int64
	var createdAt time.Time

	row := s.GetReplica().WithContext(ctx).QueryRow(`select coalesce(c.user_id, p.user_id) as rate_user,
       case when c.user_id notnull or q.type = 2  then c.created_at else p.created_at end as created_at
from call_center.cc_calls_history c
    left join call_center.cc_queue q on q.id = c.queue_id
    left join call_center.cc_calls_history p on p.id = c.bridged_id
where c.id = $1::uuid`, id)

	err := row.Err()
	if err != nil {
		return userId, createdAt, model.NewCustomCodeError("store.sql_call.get.owner.app_error", err.Error(), extractCodeFromErr(err))
	}

	err = row.Scan(&userId, &createdAt)
	if err != nil {
		return userId, createdAt, model.NewCustomCodeError("store.sql_call.get.owner.app_error", err.Error(), extractCodeFromErr(err))
	}

	return userId, createdAt, nil
}

func (s SqlCallStore) UpdateHistoryCall(ctx context.Context, domainId int64, id string, upd *model.HistoryCallPatch) model.AppError {
	res, err := s.GetMaster().WithContext(ctx).Exec(`update call_center.cc_calls_history c
set payload = coalesce(payload, '{}') || :Vars::jsonb,
	hide_missed = coalesce(:HideMissed::bool, hide_missed)
where id = :Id::uuid and domain_id = :DomainId`, map[string]interface{}{
		"Vars":       upd.Variables.ToJson(),
		"Id":         id,
		"HideMissed": upd.HideMissed,
		"DomainId":   domainId,
	})

	if err != nil {
		return model.NewCustomCodeError("store.sql_call.update.app_error", err.Error(), extractCodeFromErr(err))
	}

	var cnt int64
	cnt, err = res.RowsAffected()
	if err != nil {
		return model.NewCustomCodeError("store.sql_call.update.app_error", err.Error(), extractCodeFromErr(err))
	}

	if cnt != 1 {
		return model.NewNotFoundError("store.sql_call.update.not_found", "Not found")
	}

	return nil
}

func (s SqlCallStore) SetContactId(ctx context.Context, domainId int64, id string, contactId int64) model.AppError {
	var info *string
	err := s.GetMaster().WithContext(ctx).SelectOne(&info, `with master as (
    select x.master_id
    from (select coalesce(c.parent_id, c.id) master_id
          from call_center.cc_calls c
          where c.id = :Id
            and domain_id = :DomainId
          union
          select coalesce(c.parent_id, c.id) master_id
          from call_center.cc_calls_history c
          where c.id = :Id
            and domain_id = :DomainId) x
    limit 1
), ua as (
    update call_center.cc_calls c
        set contact_id  = :ContactId
    from master
    where id = master.master_id
    returning id
), uh as (
    update call_center.cc_calls_history
        set contact_id  = :ContactId
    from master
    where id = master.master_id
        and not exists(select 1 from ua)
    returning id
)
select ua.id as id
from ua
union all
select uh.id as id
from uh`, map[string]interface{}{
		"DomainId":  domainId,
		"ContactId": contactId,
		"Id":        id,
	})

	if err != nil {
		return model.NewCustomCodeError("store.sql_call.set_contact.app_error", err.Error(), extractCodeFromErr(err))
	}

	if info == nil {
		return model.NewNotFoundError("store.sql_call.set_contact.not_found", "Not found")
	}

	return nil
}

func (s SqlCallStore) GetSipId(ctx context.Context, domainId int64, userId int64, id string) (string, model.AppError) {
	sipId, err := s.GetMaster().WithContext(ctx).SelectStr(`select coalesce((params ->> 'sip_id')::varchar, case when c.parent_id notnull then c.id::varchar end) sip_id
from call_center.cc_calls c
where c.id = :Id
  and c.domain_id = :Domain
  and c.user_id = :UserId`, map[string]interface{}{
		"Id":     id,
		"Domain": domainId,
		"UserId": userId,
	})

	if err != nil {
		return "", model.NewCustomCodeError("store.sql_call.get_sip_id.app_error", err.Error(), extractCodeFromErr(err))
	}

	return sipId, nil
}

func (s SqlCallStore) SetHideMissedLeg(ctx context.Context, domainId int64, userId int64, id string) model.AppError {
	_, err := s.GetMaster().WithContext(ctx).Exec(`update call_center.cc_calls_history
set hide_missed = true
where domain_id = :DomainId and id = :Id::uuid and user_id = :UserId`, map[string]interface{}{
		"DomainId": domainId,
		"Id":       id,
		"UserId":   userId,
	})

	if err != nil {
		return model.NewCustomCodeError("store.sql_call.set_hide_missed.app_error", err.Error(), extractCodeFromErr(err))
	}

	return nil
}

func (s SqlCallStore) SetHideMissedAllParent(ctx context.Context, domainId int64, userId int64, id string) model.AppError {
	_, err := s.GetMaster().WithContext(ctx).Exec(`update call_center.cc_calls_history
    set hide_missed = true
where domain_id = :DomainId
    and not hide_missed is true
    and parent_id = (select parent_id
from call_center.cc_calls_history
where domain_id = :DomainId and id = :Id::uuid and user_id = :UserId)`, map[string]interface{}{
		"DomainId": domainId,
		"Id":       id,
		"UserId":   userId,
	})

	if err != nil {
		return model.NewCustomCodeError("store.sql_call.set_hide_missed_all.app_error", err.Error(), extractCodeFromErr(err))
	}

	return nil
}

func (s SqlCallStore) FromNumber(ctx context.Context, domainId int64, userId int64, id string) (string, model.AppError) {
	from, err := s.GetReplica().WithContext(ctx).SelectNullStr(`select from_number
from call_center.cc_calls_history
where domain_id = :DomainId and id = :Id::uuid and user_id = :UserId`, map[string]interface{}{
		"DomainId": domainId,
		"Id":       id,
		"UserId":   userId,
	})

	if err != nil {
		return "", model.NewCustomCodeError("store.sql_call.from_number.app_error", err.Error(), extractCodeFromErr(err))
	}

	if !from.Valid {
		return "", model.NewNotFoundError("store.sql_call.from_number.app_error", fmt.Sprintf("callId=%s not found", id))
	}

	return from.String, nil
}

func (s SqlCallStore) FromNumberWithUserIds(ctx context.Context, domainId int64, userId int64, id string) (model.RedialFrom, model.AppError) {
	var f model.RedialFrom
	err := s.GetReplica().WithContext(ctx).SelectOne(&f, `select c.from_number as number, c2.user_ids
from call_center.cc_calls_history c
    left join call_center.cc_calls_history c2 on c2.id = c.parent_id
where c.domain_id = :DomainId and c.id = :Id::uuid and c.user_id = :UserId`, map[string]interface{}{
		"DomainId": domainId,
		"Id":       id,
		"UserId":   userId,
	})

	if err != nil {
		return f, model.NewCustomCodeError("store.sql_call.from_number_users.app_error", err.Error(), extractCodeFromErr(err))
	}

	return f, nil
}
