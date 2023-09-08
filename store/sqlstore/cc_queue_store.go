package sqlstore

import (
	"context"
	"fmt"

	"github.com/lib/pq"
	"github.com/webitel/engine/auth_manager"
	"github.com/webitel/engine/model"
	"github.com/webitel/engine/store"
)

type SqlQueueStore struct {
	SqlStore
}

func NewSqlQueueStore(sqlStore SqlStore) store.QueueStore {
	us := &SqlQueueStore{sqlStore}
	return us
}

func (s SqlQueueStore) CheckAccess(ctx context.Context, domainId, id int64, groups []int, access auth_manager.PermissionAccess) (bool, model.AppError) {

	res, err := s.GetReplica().WithContext(ctx).SelectNullInt(`select 1
		where exists(
          select 1
          from call_center.cc_queue_acl a
          where a.dc = :DomainId
            and a.object = :Id
            and a.subject = any (:Groups::int[])
            and a.access & :Access = :Access
        )`, map[string]interface{}{"DomainId": domainId, "Id": id, "Groups": pq.Array(groups), "Access": access.Value()})

	if err != nil {
		return false, nil
	}

	return (res.Valid && res.Int64 == 1), nil
}

func (s SqlQueueStore) Create(ctx context.Context, queue *model.Queue) (*model.Queue, model.AppError) {
	var out *model.Queue
	if err := s.GetMaster().WithContext(ctx).SelectOne(&out, `with q as (
    insert into call_center.cc_queue (strategy, enabled, payload, calendar_id, priority, updated_at,
                      name, variables, domain_id, dnc_list_id, type, team_id,
                      created_at, created_by, updated_by, description, ringtone_id, schema_id, do_schema_id, after_schema_id, sticky_agent,
					  processing, processing_sec, processing_renewal_sec, form_schema_id, grantee_id)
values (:Strategy, :Enabled, :Payload, :CalendarId, :Priority, :UpdatedAt, :Name,
        :Variables, :DomainId, :DncListId, :Type, :TeamId, :CreatedAt, :CreatedBy, :UpdatedBy, :Description, :RingtoneId,
		:SchemaId, :DoSchemaId, :AfterSchemaId, :StickyAgent, :Processing, :ProcessingSec, :ProcessingRenewalSec, :FormSchemaId, :GranteeId)
    returning *
)
select q.id,
       q.strategy,
       q.enabled,
       q.payload,
       q.priority,
       q.updated_at,
       q.name,
       q.variables,
       q.domain_id,
       q.type,
       q.created_at,
       call_center.cc_get_lookup(uc.id, uc.name)         as created_by,
       call_center.cc_get_lookup(u.id, u.name)           as updated_by,
       call_center.cc_get_lookup(c.id, c.name)           as calendar,
       call_center.cc_get_lookup(cl.id, cl.name)         as dnc_list,
       call_center.cc_get_lookup(ct.id, ct.name)         as team,
       q.description,
       call_center.cc_get_lookup(s.id, s.name)           as schema,
       call_center.cc_get_lookup(ds.id, ds.name)                      AS do_schema,
       call_center.cc_get_lookup(afs.id, afs.name)                      AS after_schema,
       call_center.cc_get_lookup(q.ringtone_id, mf.name) as ringtone,
	   q.sticky_agent,
	   q.processing,
	   q.processing_sec,
	   q.processing_renewal_sec,
	   call_center.cc_get_lookup(fs.id, fs.name)                      AS form_schema,
       jsonb_build_object('enabled', q.processing, 'form_schema', call_center.cc_get_lookup(fs.id, fs.name), 'sec',
                          q.processing_sec, 'renewal_sec', q.processing_renewal_sec) AS task_processing,
	   call_center.cc_get_lookup(au.id, au.name)                                     AS grantee
from q
         left join flow.calendar c on q.calendar_id = c.id
		 left join directory.wbt_auth au on au.id = q.grantee_id
         left join directory.wbt_user uc on uc.id = q.created_by
         left join directory.wbt_user u on u.id = q.updated_by
         left join call_center.cc_list cl on q.dnc_list_id = cl.id
         left join flow.acr_routing_scheme s on q.schema_id = s.id
         LEFT JOIN flow.acr_routing_scheme ds ON q.do_schema_id = ds.id
         LEFT JOIN flow.acr_routing_scheme afs ON q.after_schema_id = afs.id
		 LEFT JOIN flow.acr_routing_scheme fs ON q.form_schema_id = fs.id
         left join call_center.cc_team ct on q.team_id = ct.id
         left join storage.media_files mf on mf.id = q.ringtone_id`,
		map[string]interface{}{
			"Strategy":             queue.Strategy,
			"Enabled":              queue.Enabled,
			"Payload":              queue.Payload.ToSafeBytes(),
			"CalendarId":           queue.Calendar.GetSafeId(),
			"Priority":             queue.Priority,
			"UpdatedAt":            queue.UpdatedAt,
			"Name":                 queue.Name,
			"Variables":            queue.Variables.ToJson(),
			"DomainId":             queue.DomainId,
			"DncListId":            queue.DncListId(),
			"Type":                 queue.Type,
			"TeamId":               queue.TeamId(),
			"CreatedAt":            queue.CreatedAt,
			"CreatedBy":            queue.CreatedBy.GetSafeId(),
			"UpdatedBy":            queue.UpdatedBy.GetSafeId(),
			"Description":          queue.Description,
			"SchemaId":             queue.SchemaId(),
			"DoSchemaId":           queue.DoSchemaId(),
			"AfterSchemaId":        queue.AfterSchemaId(),
			"RingtoneId":           queue.RingtoneId(),
			"StickyAgent":          queue.StickyAgent,
			"Processing":           queue.Processing,
			"ProcessingSec":        queue.ProcessingSec,
			"ProcessingRenewalSec": queue.ProcessingRenewalSec,
			"FormSchemaId":         queue.FormSchema.GetSafeId(),
			"GranteeId":            queue.Grantee.GetSafeId(),
		}); nil != err {
		return nil, model.NewCustomCodeError("store.sql_queue.save.app_error", fmt.Sprintf("name=%v, %v", queue.Name, err.Error()), extractCodeFromErr(err))
	} else {
		return out, nil
	}
}

func (s SqlQueueStore) GetAllPage(ctx context.Context, domainId int64, search *model.SearchQueue) ([]*model.Queue, model.AppError) {
	var queues []*model.Queue

	f := map[string]interface{}{
		"DomainId": domainId,
		"Ids":      pq.Array(search.Ids),
		"Q":        search.GetQ(),
		"Types":    pq.Array(search.Types),
	}

	err := s.ListQueryMaster(ctx, &queues, search.ListRequest,
		`domain_id = :DomainId 
			and ( (:Ids::int[] isnull or id = any(:Ids) )  
			and ( (:Types::int[] isnull or "type" = any(:Types) ) ) 
			and (:Q::varchar isnull or (name ilike :Q::varchar or description ilike :Q::varchar ) ))`,
		model.Queue{}, f)
	if err != nil {
		return nil, model.NewInternalError("store.sql_queue.get_all.app_error", err.Error())
	}

	return queues, nil
}

func (s SqlQueueStore) GetAllPageByGroups(ctx context.Context, domainId int64, groups []int, search *model.SearchQueue) ([]*model.Queue, model.AppError) {
	var queues []*model.Queue

	f := map[string]interface{}{
		"Groups":   pq.Array(groups),
		"Access":   auth_manager.PERMISSION_ACCESS_READ.Value(),
		"DomainId": domainId,
		"Ids":      pq.Array(search.Ids),
		"Q":        search.GetQ(),
	}

	err := s.ListQueryMaster(ctx, &queues, search.ListRequest,
		`domain_id = :DomainId and  (
					exists(select 1
					  from call_center.cc_queue_acl acl
					  where acl.dc = t.domain_id and acl.object = t.id and acl.subject = any(:Groups::int[]) and acl.access&:Access = :Access)
		  	) and ( (:Ids::int[] isnull or id = any(:Ids) )  and (:Q::varchar isnull or (name ilike :Q::varchar or description ilike :Q::varchar ) ))`,
		model.Queue{}, f)
	if err != nil {
		return nil, model.NewInternalError("store.sql_queue.get_all.app_error", err.Error())
	}

	return queues, nil
}

func (s SqlQueueStore) Get(ctx context.Context, domainId int64, id int64) (*model.Queue, model.AppError) {
	var queue *model.Queue
	if err := s.GetReplica().WithContext(ctx).SelectOne(&queue, `
select q.id,
       q.strategy,
       q.enabled,
       q.payload,
       q.priority,
       q.updated_at,
       q.name,
       q.variables,
       q.domain_id,
       q.type,
       q.created_at,
       call_center.cc_get_lookup(uc.id, uc.name)         as created_by,
       call_center.cc_get_lookup(u.id, u.name)           as updated_by,
       call_center.cc_get_lookup(c.id, c.name)           as calendar,
       call_center.cc_get_lookup(cl.id, cl.name)         as dnc_list,
       call_center.cc_get_lookup(ct.id, ct.name)         as team,
       q.description,
       call_center.cc_get_lookup(s.id, s.name)           as schema,
       call_center.cc_get_lookup(ds.id, ds.name)                      AS do_schema,
       call_center.cc_get_lookup(afs.id, afs.name)                      AS after_schema,
       call_center.cc_get_lookup(q.ringtone_id, mf.name) as ringtone,
	   q.sticky_agent,
	   q.processing,
	   q.processing_sec,
	   q.processing_renewal_sec,
	   call_center.cc_get_lookup(fs.id, fs.name)                      AS form_schema,
       jsonb_build_object('enabled', q.processing, 'form_schema', call_center.cc_get_lookup(fs.id, fs.name), 'sec',
                          q.processing_sec, 'renewal_sec', q.processing_renewal_sec) AS task_processing,
	   call_center.cc_get_lookup(au.id, au.name)                                     AS grantee
from call_center.cc_queue q
         left join flow.calendar c on q.calendar_id = c.id
	     left join directory.wbt_auth au on au.id = q.grantee_id
         left join directory.wbt_user uc on uc.id = q.created_by
         left join directory.wbt_user u on u.id = q.updated_by
         left join call_center.cc_list cl on q.dnc_list_id = cl.id
         left join flow.acr_routing_scheme s on q.schema_id = s.id
         LEFT JOIN flow.acr_routing_scheme ds ON q.do_schema_id = ds.id
         LEFT JOIN flow.acr_routing_scheme afs ON q.after_schema_id = afs.id
		 LEFT JOIN flow.acr_routing_scheme fs ON q.form_schema_id = fs.id
         left join call_center.cc_team ct on q.team_id = ct.id
         left join storage.media_files mf on mf.id = q.ringtone_id 	
where q.domain_id = :DomainId and q.id = :Id
		`, map[string]interface{}{"Id": id, "DomainId": domainId}); err != nil {
		return nil, model.NewCustomCodeError("store.sql_queue.get.app_error", fmt.Sprintf("Id=%v, %s", id, err.Error()), extractCodeFromErr(err))
	} else {
		return queue, nil
	}
}

func (s SqlQueueStore) Update(ctx context.Context, queue *model.Queue) (*model.Queue, model.AppError) {
	err := s.GetMaster().WithContext(ctx).SelectOne(&queue, `with q as (
    update call_center.cc_queue q
set updated_at = :UpdatedAt,
    updated_by = :UpdatedBy,
    strategy = :Strategy,
    enabled = :Enabled,
    payload = :Payload,
    calendar_id = :CalendarId,
    priority = :Priority,
    name = :Name,
    variables = :Variables,
    dnc_list_id = :DncListId,
    type = :Type,
    team_id = :TeamId,
	description = :Description,
	schema_id = :SchemaId,
	ringtone_id = :RingtoneId,
	do_schema_id = :DoSchemaId,
	after_schema_id = :AfterSchemaId,
	sticky_agent = :StickyAgent,
	processing = :Processing,
	processing_sec = :ProcessingSec,
    processing_renewal_sec = :ProcessingRenewalSec,
	form_schema_id = :FormSchemaId,
	grantee_id = :GranteeId
where q.id = :Id and q.domain_id = :DomainId
    returning *
)
select q.id,
       q.strategy,
       q.enabled,
       q.payload,
       q.priority,
       q.updated_at,
       q.name,
       q.variables,
       q.domain_id,
       q.type,
       q.created_at,
       call_center.cc_get_lookup(uc.id, uc.name)         as created_by,
       call_center.cc_get_lookup(u.id, u.name)           as updated_by,
       call_center.cc_get_lookup(c.id, c.name)           as calendar,
       call_center.cc_get_lookup(cl.id, cl.name)         as dnc_list,
       call_center.cc_get_lookup(ct.id, ct.name)         as team,
       q.description,
       call_center.cc_get_lookup(s.id, s.name)           as schema,
       call_center.cc_get_lookup(ds.id, ds.name)                      AS do_schema,
       call_center.cc_get_lookup(afs.id, afs.name)                      AS after_schema,
       call_center.cc_get_lookup(q.ringtone_id, mf.name) as ringtone,
	   q.sticky_agent,
	   q.processing,
	   q.processing_sec,
	   q.processing_renewal_sec,
	   call_center.cc_get_lookup(fs.id, fs.name)                      AS form_schema,
       jsonb_build_object('enabled', q.processing, 'form_schema', call_center.cc_get_lookup(fs.id, fs.name), 'sec',
                          q.processing_sec, 'renewal_sec', q.processing_renewal_sec) AS task_processing,
	   call_center.cc_get_lookup(au.id, au.name)                                     AS grantee
from  q
         left join flow.calendar c on q.calendar_id = c.id
		 left join directory.wbt_auth au on au.id = q.grantee_id
         left join directory.wbt_user uc on uc.id = q.created_by
         left join directory.wbt_user u on u.id = q.updated_by
         left join call_center.cc_list cl on q.dnc_list_id = cl.id
         left join flow.acr_routing_scheme s on q.schema_id = s.id
         LEFT JOIN flow.acr_routing_scheme ds ON q.do_schema_id = ds.id
         LEFT JOIN flow.acr_routing_scheme afs ON q.after_schema_id = afs.id
		 LEFT JOIN flow.acr_routing_scheme fs ON q.form_schema_id = fs.id
         left join call_center.cc_team ct on q.team_id = ct.id
         left join storage.media_files mf on mf.id = q.ringtone_id`, map[string]interface{}{
		"UpdatedAt":            queue.UpdatedAt,
		"UpdatedBy":            queue.UpdatedBy.GetSafeId(),
		"Strategy":             queue.Strategy,
		"Enabled":              queue.Enabled,
		"Payload":              queue.Payload.ToSafeBytes(),
		"CalendarId":           queue.Calendar.GetSafeId(),
		"Priority":             queue.Priority,
		"Name":                 queue.Name,
		"Variables":            queue.Variables.ToJson(),
		"DncListId":            queue.DncListId(),
		"Type":                 queue.Type,
		"TeamId":               queue.TeamId(),
		"SchemaId":             queue.SchemaId(),
		"Id":                   queue.Id,
		"DomainId":             queue.DomainId,
		"Description":          queue.Description,
		"RingtoneId":           queue.RingtoneId(),
		"DoSchemaId":           queue.DoSchemaId(),
		"AfterSchemaId":        queue.AfterSchemaId(),
		"StickyAgent":          queue.StickyAgent,
		"Processing":           queue.Processing,
		"ProcessingSec":        queue.ProcessingSec,
		"ProcessingRenewalSec": queue.ProcessingRenewalSec,
		"FormSchemaId":         queue.FormSchema.GetSafeId(),
		"GranteeId":            queue.Grantee.GetSafeId(),
	})
	if err != nil {
		return nil, model.NewCustomCodeError("store.sql_queue.update.app_error", fmt.Sprintf("Id=%v, %s", queue.Id, err.Error()), extractCodeFromErr(err))
	}
	return queue, nil
}

func (s SqlQueueStore) Delete(ctx context.Context, domainId, id int64) model.AppError {
	if _, err := s.GetMaster().WithContext(ctx).Exec(`delete from call_center.cc_queue c where c.id=:Id and c.domain_id = :DomainId`,
		map[string]interface{}{"Id": id, "DomainId": domainId}); err != nil {
		return model.NewInternalError("store.sql_queue.delete.app_error", fmt.Sprintf("Id=%v, %s", id, err.Error()))
	}
	return nil
}

// QueueReportGeneral TODO call_center.cc_agent_channel not unique join
func (s SqlQueueStore) QueueReportGeneral(ctx context.Context, domainId int64, supervisorId int64, groups []int, access auth_manager.PermissionAccess, search *model.SearchQueueReportGeneral) (*model.QueueReportGeneralAgg, model.AppError) {
	var report *model.QueueReportGeneralAgg
	err := s.GetReplica().WithContext(ctx).SelectOne(&report, `
with queues  as  (
    select *
    from call_center.cc_queue q
    where  q.id in (
        with x as (
            select a.user_id, a.id agent_id, a.supervisor, a.domain_id
            from directory.wbt_user u
                     inner join call_center.cc_agent a on a.user_id = u.id and a.domain_id = u.dc
            where u.id = :UserSupervisorId
              and u.dc = :DomainId
        )
        select distinct qs.queue_id
        from x
                 left join lateral (
            select a.id, a.auditor_ids && array [x.user_id] aud
            from call_center.cc_agent a
            where (a.user_id = x.user_id or (a.supervisor_ids && array [x.agent_id]))
            union
            distinct
            select a.id, a.auditor_ids && array [x.user_id] aud
            from call_center.cc_team t
                     inner join call_center.cc_agent a on a.team_id = t.id
            where t.admin_ids && array[x.agent_id]
            ) a on true
                 inner join call_center.cc_skill_in_agent sa on sa.agent_id = a.id
                 inner join call_center.cc_queue_skill qs
                            on qs.skill_id = sa.skill_id and sa.capacity between qs.min_capacity and qs.max_capacity
			 where sa.enabled and qs.enabled
        union
        select q.id
        from call_center.cc_queue q
        where q.domain_id = :DomainId
          and q.grantee_id = any (:Groups) and q.enabled
    ) and q.enabled
 ),
     queue_ag as (
        select distinct
               q.id queue_id,
               array_agg(distinct a.id) filter ( where status = 'online' ) agent_on_ids,
               array_agg(distinct a.id) filter ( where status = 'offline' ) agent_off_ids,
               array_agg(distinct a.id) filter ( where status in ('pause', 'break_out') ) agent_p_ids,
               array_agg(distinct a.id) filter ( where status = 'online' and ac.channel isnull and ac.state = 'waiting' ) free,
               array_agg(distinct a.id) total
        from queues q
            inner join call_center.cc_agent a on a.domain_id = q.domain_id
            inner join call_center.cc_agent_channel ac on ac.agent_id = a.id
            inner join call_center.cc_queue_skill qs on qs.queue_id = q.id and qs.enabled
            inner join call_center.cc_skill_in_agent sia on sia.agent_id = a.id and sia.enabled
        where (q.team_id isnull or a.team_id = q.team_id) and qs.skill_id = sia.skill_id and sia.capacity between qs.min_capacity and qs.max_capacity
        group by rollup (q.id)
     ),
items as materialized (
    select call_center.cc_get_lookup(q.id, q.name) queue,
           call_center.cc_get_lookup(ct.id, ct.name) team,
           jsonb_build_object('online', coalesce(array_length(queue_ag.agent_on_ids, 1), 0),
                                  'pause', coalesce(array_length(queue_ag.agent_p_ids, 1), 0),
                                  'offline', coalesce(array_length(queue_ag.agent_off_ids, 1), 0),
                                  'free', coalesce(array_length(queue_ag.free, 1), 0),
                                  'total', coalesce(array_length(queue_ag.total, 1), 0)
               ) agent_status,

           coalesce(ag.abandoned::int, 0) missed,
           (select count(*) from call_center.cc_member_attempt a where a.queue_id = q.id and a.bridged_at notnull) processed,
           coalesce(case when q.type = 1 then (select count(*) from call_center.cc_member_attempt a1 where a1.queue_id = q.id and a1.bridged_at isnull)
               else (select sum(s.member_waiting) from call_center.cc_queue_statistics s where s.queue_id = q.id) end, 0) waiting,
           coalesce(ag.count, 0) count,
           coalesce(ag.transferred, 0) transferred,
		   0 attempts,
           coalesce(ag.bridged, 0) bridged,
           coalesce(ag.abandoned::int, 0) abandoned,

           coalesce(ag.sum_bill_sec, 0) sum_bill_sec,
           coalesce(ag.sl20, 0) sl20,
           coalesce(ag.sl30, 0) sl30,
           coalesce(ag.avg_wrap_sec, 0) avg_wrap_sec,
           coalesce(ag.avg_awt_sec, 0) avg_awt_sec,
           coalesce(ag.max_awt_sec, 0) max_awt_sec,
           coalesce(ag.avg_asa_sec, 0) avg_asa_sec,
           coalesce(ag.avg_aht_sec, 0) avg_aht_sec
    from queues q
        left join queue_ag on queue_ag.queue_id = q.id
        left join call_center.cc_team ct on q.team_id = ct.id
        left join lateral (
            select
                   t.queue_id,
                   count(*) as count,
                   count(*) filter ( where t.bridged_at notnull ) * 100.0 / count(*) as bridged,
                   count(*) filter ( where t.bridged_at isnull  ) * 100.0 / count(*) as abandoned,
                   count(*) filter ( where t.result = 'transfer' )  as transferred,
				   case when count(*)::decimal > 0 then
			       	(((count(*) filter ( where t.bridged_at - t.joined_at < interval '20 sec')::decimal) / count(*)::decimal) * 100)
				   else 0::decimal end as sl20,
				   case when count(*)::decimal > 0 then
			       	(((count(*) filter ( where t.bridged_at - t.joined_at < interval '30 sec')::decimal) / count(*)::decimal) * 100)
				   else 0::decimal end as sl30,
                   extract(EPOCH from sum(t.leaving_at - t.bridged_at) filter ( where t.bridged_at notnull )) sum_bill_sec,
                   extract(EPOCH from avg(t.reporting_at - t.leaving_at) filter ( where t.reporting_at notnull )) avg_wrap_sec,
                   extract(EPOCH from avg(t.bridged_at - t.offering_at) filter ( where t.bridged_at notnull )) avg_awt_sec, --TODO!!! FIXME
                   extract(epoch from max(t.bridged_at - t.offering_at) filter ( where t.bridged_at notnull )) max_awt_sec,
                   extract(epoch from avg(t.bridged_at - t.joined_at) filter ( where t.bridged_at notnull )) avg_asa_sec,
                   extract(epoch from avg( GREATEST(t.leaving_at, t.reporting_at) - t.bridged_at ) filter ( where t.bridged_at notnull )) avg_aht_sec
            from call_center.cc_member_attempt_history t
            where t.domain_id = :DomainId and t.joined_at between :From::timestamptz and :To::timestamptz
                and t.queue_id = q.id
            group by 1
    ) ag on true
    where ( :QueueIds::int[] isnull or q.id = any(:QueueIds) )
        and ( :Types::int[] isnull or q.type = any(:Types) )
        and ( :TeamIds::int[] isnull or q.team_id = any(:TeamIds) )
        and (:Q::varchar isnull or (q.name ilike :Q::varchar ) )
    order by q.priority desc
    limit :Limit
    offset :Offset
)
select
    (select jsonb_agg(items) from items) as items,
    (select jsonb_build_object(
            'online', coalesce(array_length(agent_on_ids, 1), 0),
            'pause', coalesce(array_length(agent_p_ids, 1), 0),
            'offline', coalesce(array_length(agent_off_ids, 1), 0),
            'free', coalesce(array_length(free, 1), 0),
            'total', coalesce(array_length(total, 1), 0)
                            ) from queue_ag where queue_ag.queue_id isnull ) aggs
`, map[string]interface{}{
		"DomainId":         domainId,
		"UserSupervisorId": supervisorId,
		"Groups":           pq.Array(groups),
		//"Access":   access.Value(),
		"From":     model.GetBetweenFromTime(&search.JoinedAt),
		"To":       model.GetBetweenToTime(&search.JoinedAt),
		"Q":        search.GetQ(),
		"QueueIds": pq.Array(search.QueueIds),
		"TeamIds":  pq.Array(search.TeamIds),
		"Types":    pq.Array(search.Types),
		"Limit":    search.GetLimit(),
		"Offset":   search.GetOffset(),
	})

	if err != nil {
		return nil, model.NewCustomCodeError("store.sql_queue.report_general.app_error", err.Error(), extractCodeFromErr(err))
	}

	return report, nil
}
