package sqlstore

import (
	"context"
	"fmt"

	"github.com/lib/pq"
	"github.com/webitel/engine/model"
	"github.com/webitel/engine/pkg/wbt/auth_manager"
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

func (s SqlQueueStore) RbacUniqueQueues(ctx context.Context, domainId int64, queueIds []int64, groups []int) ([]int32, model.AppError) {
	var res []int32
	_, err := s.GetReplica().WithContext(ctx).
		Select(&res, `select distinct object
from call_center.cc_queue_acl a
where a.dc = :DomainId
    and (:QueueIds::int[] isnull or object = any (:QueueIds::int[]))
    and a.subject = any (:Groups::int[])
    and a.access & :Access = :Access;`, map[string]any{
			"DomainId": domainId,
			"QueueIds": pq.Array(queueIds),
			"Groups":   pq.Array(groups),
			"Access":   auth_manager.PERMISSION_ACCESS_UPDATE.Value(),
		})

	if err != nil {
		return nil, model.NewCustomCodeError("store.sql_queue.rbac_queues.app_error", err.Error(), extractCodeFromErr(err))
	}

	return res, nil
}

func (s SqlQueueStore) Create(ctx context.Context, queue *model.Queue) (*model.Queue, model.AppError) {
	query := `
		with q as (
			insert into call_center.cc_queue (
				strategy, enabled, payload, calendar_id,
				priority, updated_at, name, variables,
				domain_id, dnc_list_id, type, team_id,
				created_at, created_by, updated_by, description,
				ringtone_id, schema_id, do_schema_id, after_schema_id,
				sticky_agent, processing, processing_sec, processing_renewal_sec,
				form_schema_id, grantee_id, tags, prolongation_enabled,
				prolongation_repeats_number, prolongation_time_sec, 
				prolongation_is_timeout_retry
			)
			values (
				:Strategy, :Enabled, :Payload, :CalendarId,
				:Priority, :UpdatedAt, :Name, :Variables,
				:DomainId, :DncListId, :Type, :TeamId,
				:CreatedAt, :CreatedBy, :UpdatedBy, :Description,
				:RingtoneId, :SchemaId, :DoSchemaId, :AfterSchemaId,
				:StickyAgent, :Processing, :ProcessingSec, :ProcessingRenewalSec,
				:FormSchemaId, :GranteeId, :Tags, :ProlongationEnabled,
				:ProlongationRepeatsNumber, :ProlongationTimeSec, :ProlongationIsTimeoutRetry
			)
			returning *
		)
		select
			q.id,
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
			q.description,
			q.sticky_agent,
			q.processing,
			q.processing_sec,
			q.processing_renewal_sec,
			q.prolongation_enabled,
			q.prolongation_repeats_number,
			q.prolongation_time_sec,
			q.prolongation_is_timeout_retry,
			q.tags,
			call_center.cc_get_lookup(uc.id, uc.name) 			as created_by,
			call_center.cc_get_lookup(u.id, u.name) 			as updated_by,
			call_center.cc_get_lookup(c.id, c.name) 			as calendar,
			call_center.cc_get_lookup(cl.id, cl.name) 			as dnc_list,
			call_center.cc_get_lookup(ct.id, ct.name) 			as team,
			call_center.cc_get_lookup(s.id, s.name) 			as schema,
			call_center.cc_get_lookup(ds.id, ds.name) 			as do_schema,
			call_center.cc_get_lookup(afs.id, afs.name) 		as after_schema,
			call_center.cc_get_lookup(q.ringtone_id, mf.name) 	as ringtone,
			call_center.cc_get_lookup(fs.id, fs.name) 			as form_schema,
			call_center.cc_get_lookup(au.id, au.name) 			as grantee,
			jsonb_build_object (
				'enabled', q.processing,
				'form_schema', call_center.cc_get_lookup(fs.id, fs.name),
				'sec', q.processing_sec,
				'renewal_sec', q.processing_renewal_sec,
				'prolongation_options',
					case
						when q.prolongation_enabled then jsonb_build_object (
							'prolongation_enabled', q.prolongation_enabled,
							'prolongation_repeats_number', q.prolongation_repeats_number,
							'prolongation_time_sec', q.prolongation_time_sec,
							'prolongation_is_timeout_retry', q.prolongation_is_timeout_retry
						)
						else null
					end
			) as task_processing
		from
			q
		left join
			flow.calendar c on q.calendar_id = c.id
		left join
			directory.wbt_auth au on au.id = q.grantee_id
		left join
			directory.wbt_user uc on uc.id = q.created_by
		left join
			directory.wbt_user u on u.id = q.updated_by
		left join
			call_center.cc_list cl on q.dnc_list_id = cl.id
		left join
			flow.acr_routing_scheme s on q.schema_id = s.id
		left join
			flow.acr_routing_scheme ds on q.do_schema_id = ds.id
		left join
			flow.acr_routing_scheme afs on q.after_schema_id = afs.id
		left join
			flow.acr_routing_scheme fs on q.form_schema_id = fs.id
		left join
			call_center.cc_team ct on q.team_id = ct.id
		left join
			storage.media_files mf on mf.id = q.ringtone_id
	`

	args := map[string]any{
		"Strategy":                   queue.Strategy,
		"Enabled":                    queue.Enabled,
		"Payload":                    queue.Payload.ToSafeBytes(),
		"CalendarId":                 queue.Calendar.GetSafeId(),
		"Priority":                   queue.Priority,
		"UpdatedAt":                  queue.UpdatedAt,
		"Name":                       queue.Name,
		"Variables":                  queue.Variables.ToJson(),
		"DomainId":                   queue.DomainId,
		"DncListId":                  queue.DncListId(),
		"Type":                       queue.Type,
		"TeamId":                     queue.TeamId(),
		"CreatedAt":                  queue.CreatedAt,
		"CreatedBy":                  queue.CreatedBy.GetSafeId(),
		"UpdatedBy":                  queue.UpdatedBy.GetSafeId(),
		"Description":                queue.Description,
		"SchemaId":                   queue.SchemaId(),
		"DoSchemaId":                 queue.DoSchemaId(),
		"AfterSchemaId":              queue.AfterSchemaId(),
		"RingtoneId":                 queue.RingtoneId(),
		"StickyAgent":                queue.StickyAgent,
		"Processing":                 queue.Processing,
		"ProcessingSec":              queue.ProcessingSec,
		"ProcessingRenewalSec":       queue.ProcessingRenewalSec,
		"FormSchemaId":               queue.FormSchema.GetSafeId(),
		"GranteeId":                  queue.Grantee.GetSafeId(),
		"Tags":                       pq.Array(queue.Tags),
		"ProlongationEnabled":        queue.TaskProcessing.ProlongationOptions.ProlongationEnabled,
		"ProlongationRepeatsNumber":  queue.TaskProcessing.ProlongationOptions.RepeatsNumber,
		"ProlongationTimeSec":        queue.TaskProcessing.ProlongationOptions.ProlongationTimeSec,
		"ProlongationIsTimeoutRetry": queue.TaskProcessing.ProlongationOptions.IsTimeoutRetry,
	}

	var out *model.Queue
	if err := s.GetMaster().WithContext(ctx).SelectOne(&out, query, args); err != nil {
		return nil, model.NewCustomCodeError("store.sql_queue.save.app_error", fmt.Sprintf("name=%v, %v", queue.Name, err.Error()), extractCodeFromErr(err))
	}

	return out, nil
}

func (s SqlQueueStore) GetAllPage(ctx context.Context, domainId int64, search *model.SearchQueue) ([]*model.Queue, model.AppError) {
	var queues []*model.Queue

	f := map[string]interface{}{
		"DomainId": domainId,
		"Ids":      pq.Array(search.Ids),
		"Q":        search.GetQ(),
		"Types":    pq.Array(search.Types),
		"TeamIds":  pq.Array(search.TeamIds),
		"Tags":     pq.Array(search.Tags),
		"Enabled":  search.Enabled,
	}

	err := s.ListQueryMaster(ctx, &queues, search.ListRequest,
		`domain_id = :DomainId 
			and ( (:Ids::int[] isnull or id = any(:Ids) )  
			and ( (:Types::int[] isnull or "type" = any(:Types) ) ) 
			and ( :TeamIds::int[] isnull or "team_id" = any(:TeamIds) ) 
			and (:Tags::varchar[] isnull or tags && :Tags::varchar[])
			and (:Enabled::bool isnull or enabled)
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
		"Types":    pq.Array(search.Types),
		"TeamIds":  pq.Array(search.TeamIds),
		"Tags":     pq.Array(search.Tags),
		"Enabled":  search.Enabled,
	}

	err := s.ListQueryMaster(ctx, &queues, search.ListRequest,
		`domain_id = :DomainId and  (
					exists(select 1
					  from call_center.cc_queue_acl acl
					  where acl.dc = t.domain_id and acl.object = t.id and acl.subject = any(:Groups::int[]) and acl.access&:Access = :Access)
		  	) 
			and ( (:Ids::int[] isnull or id = any(:Ids) )  
			and ( (:Types::int[] isnull or "type" = any(:Types) ) ) 
			and ( :TeamIds::int[] isnull or "team_id" = any(:TeamIds) ) 
			and (:Tags::varchar[] isnull or tags && :Tags::varchar[])
			and (:Enabled::bool isnull or enabled)
			and (:Q::varchar isnull or (name ilike :Q::varchar or description ilike :Q::varchar ) ))`,
		model.Queue{}, f)
	if err != nil {
		return nil, model.NewInternalError("store.sql_queue.get_all.app_error", err.Error())
	}

	return queues, nil
}

func (s SqlQueueStore) Get(ctx context.Context, domainId int64, id int64) (*model.Queue, model.AppError) {
	query := `
		select
			q.id,
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
			q.description,
			q.sticky_agent,
			q.processing,
			q.processing_sec,
			q.processing_renewal_sec,
			q.prolongation_enabled,
			q.prolongation_repeats_number,
			q.prolongation_time_sec,
			q.prolongation_is_timeout_retry,
			q.tags,
			call_center.cc_get_lookup(uc.id, uc.name) 			as created_by,
			call_center.cc_get_lookup(u.id, u.name) 			as updated_by,
			call_center.cc_get_lookup(c.id, c.name) 			as calendar,
			call_center.cc_get_lookup(cl.id, cl.name) 			as dnc_list,
			call_center.cc_get_lookup(ct.id, ct.name) 			as team,
			call_center.cc_get_lookup(s.id, s.name) 			as schema,
			call_center.cc_get_lookup(ds.id, ds.name) 			as do_schema,
			call_center.cc_get_lookup(afs.id, afs.name) 		as after_schema,
			call_center.cc_get_lookup(q.ringtone_id, mf.name) 	as ringtone,
			call_center.cc_get_lookup(fs.id, fs.name) 			as form_schema,
			call_center.cc_get_lookup(au.id, au.name) 			as grantee,
			jsonb_build_object (
				'enabled', q.processing,
				'form_schema', call_center.cc_get_lookup(fs.id, fs.name),
				'sec', q.processing_sec,
				'renewal_sec', q.processing_renewal_sec,
				'prolongation_options',
					case
						when q.prolongation_enabled then jsonb_build_object (
							'prolongation_enabled', q.prolongation_enabled,
							'prolongation_repeats_number', q.prolongation_repeats_number,
							'prolongation_time_sec', q.prolongation_time_sec,
							'prolongation_is_timeout_retry', q.prolongation_is_timeout_retry
						)
						else null
					end
			) as task_processing
			from
				call_center.cc_queue q
			left join
				flow.calendar c on q.calendar_id = c.id
			left join
				directory.wbt_auth au on au.id = q.grantee_id
			left join
				directory.wbt_user uc on uc.id = q.created_by
			left join
				directory.wbt_user u on u.id = q.updated_by
			left join
				call_center.cc_list cl on q.dnc_list_id = cl.id
			left join
				flow.acr_routing_scheme s on q.schema_id = s.id
			left join
				flow.acr_routing_scheme ds on q.do_schema_id = ds.id
			left join
				flow.acr_routing_scheme afs on q.after_schema_id = afs.id
			left join
				flow.acr_routing_scheme fs on q.form_schema_id = fs.id
			left join
				call_center.cc_team ct on q.team_id = ct.id
			left join
				storage.media_files mf on mf.id = q.ringtone_id
			where
				q.domain_id = :DomainId 
				and q.id = :Id
	`
	args := map[string]any{
		"Id":       id,
		"DomainId": domainId,
	}

	var queue *model.Queue
	if err := s.GetReplica().WithContext(ctx).SelectOne(&queue, query, args); err != nil {
		return nil, model.NewCustomCodeError("store.sql_queue.get.app_error", fmt.Sprintf("Id=%v, %s", id, err.Error()), extractCodeFromErr(err))
	}

	return queue, nil
}

func (s SqlQueueStore) Update(ctx context.Context, queue *model.Queue) (*model.Queue, model.AppError) {
	query := `
		with q as (
			update 
				call_center.cc_queue q
			set
				updated_at 	= :UpdatedAt,
				updated_by 	= :UpdatedBy,
				strategy 	= :Strategy,
				enabled 	= :Enabled,
				payload 	= :Payload,
				calendar_id = :CalendarId,
				priority 	= :Priority,
				name 		= :Name,
				variables 	= :Variables,
				dnc_list_id = :DncListId,
    			type 		= :Type,
    			team_id 	= :TeamId,
				description = :Description,
				schema_id 	= :SchemaId,
				ringtone_id = :RingtoneId,
				do_schema_id = :DoSchemaId,
				after_schema_id = :AfterSchemaId,
				sticky_agent = :StickyAgent,
				processing 	= :Processing,
				processing_sec = :ProcessingSec,
    			processing_renewal_sec = :ProcessingRenewalSec,
				form_schema_id = :FormSchemaId,
				grantee_id = :GranteeId,
    			tags 	= :Tags,
				prolongation_enabled 			= :ProlongationEnabled,
				prolongation_repeats_number 	= :ProlongationRepeatsNumber,
				prolongation_time_sec 			= :ProlongationTimeSec,
				prolongation_is_timeout_retry 	= :ProlongationIsTimeoutRetry
			where
				q.id = :Id
				and q.domain_id = :DomainId
			returning *
		)
		select
			q.id,
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
			q.description,
			q.sticky_agent,
			q.processing,
			q.processing_sec,
			q.processing_renewal_sec,
			q.prolongation_enabled,
			q.prolongation_repeats_number,
			q.prolongation_time_sec,
			q.prolongation_is_timeout_retry,
			q.tags,
			call_center.cc_get_lookup(uc.id, uc.name) 			as created_by,
			call_center.cc_get_lookup(u.id, u.name) 			as updated_by,
			call_center.cc_get_lookup(c.id, c.name) 			as calendar,
			call_center.cc_get_lookup(cl.id, cl.name) 			as dnc_list,
			call_center.cc_get_lookup(ct.id, ct.name) 			as team,
			call_center.cc_get_lookup(s.id, s.name) 			as schema,
			call_center.cc_get_lookup(ds.id, ds.name) 			as do_schema,
			call_center.cc_get_lookup(afs.id, afs.name) 		as after_schema,
			call_center.cc_get_lookup(q.ringtone_id, mf.name) 	as ringtone,
			call_center.cc_get_lookup(fs.id, fs.name) 			as form_schema,
			call_center.cc_get_lookup(au.id, au.name) 			as grantee,
			jsonb_build_object (
				'enabled', q.processing,
				'form_schema', call_center.cc_get_lookup(fs.id, fs.name),
				'sec', q.processing_sec,
				'renewal_sec', q.processing_renewal_sec,
				'prolongation_options',
					case
						when q.prolongation_enabled then jsonb_build_object (
							'prolongation_enabled', q.prolongation_enabled,
							'prolongation_repeats_number', q.prolongation_repeats_number,
							'prolongation_time_sec', q.prolongation_time_sec,
							'prolongation_is_timeout_retry', q.prolongation_is_timeout_retry
						)
						else null
					end
			) as task_processing
		from
			q
		left join
			flow.calendar c on q.calendar_id = c.id
		left join
			directory.wbt_auth au on au.id = q.grantee_id
		left join
			directory.wbt_user uc on uc.id = q.created_by
		left join
			directory.wbt_user u on u.id = q.updated_by
		left join
			call_center.cc_list cl on q.dnc_list_id = cl.id
		left join
			flow.acr_routing_scheme s on q.schema_id = s.id
		left join
			flow.acr_routing_scheme ds on q.do_schema_id = ds.id
		left join
			flow.acr_routing_scheme afs on q.after_schema_id = afs.id
		left join
			flow.acr_routing_scheme fs on q.form_schema_id = fs.id
		left join
			call_center.cc_team ct on q.team_id = ct.id
		left join
			storage.media_files mf on mf.id = q.ringtone_id
	`
	args := map[string]any{
		"UpdatedAt":                  queue.UpdatedAt,
		"UpdatedBy":                  queue.UpdatedBy.GetSafeId(),
		"Strategy":                   queue.Strategy,
		"Enabled":                    queue.Enabled,
		"Payload":                    queue.Payload.ToSafeBytes(),
		"CalendarId":                 queue.Calendar.GetSafeId(),
		"Priority":                   queue.Priority,
		"Name":                       queue.Name,
		"Variables":                  queue.Variables.ToJson(),
		"DncListId":                  queue.DncListId(),
		"Type":                       queue.Type,
		"TeamId":                     queue.TeamId(),
		"SchemaId":                   queue.SchemaId(),
		"Id":                         queue.Id,
		"DomainId":                   queue.DomainId,
		"Description":                queue.Description,
		"RingtoneId":                 queue.RingtoneId(),
		"DoSchemaId":                 queue.DoSchemaId(),
		"AfterSchemaId":              queue.AfterSchemaId(),
		"StickyAgent":                queue.StickyAgent,
		"Processing":                 queue.Processing,
		"ProcessingSec":              queue.ProcessingSec,
		"ProcessingRenewalSec":       queue.ProcessingRenewalSec,
		"FormSchemaId":               queue.FormSchema.GetSafeId(),
		"GranteeId":                  queue.Grantee.GetSafeId(),
		"Tags":                       pq.Array(queue.Tags),
		"ProlongationEnabled":        queue.TaskProcessing.ProlongationOptions.ProlongationEnabled,
		"ProlongationRepeatsNumber":  queue.TaskProcessing.ProlongationOptions.RepeatsNumber,
		"ProlongationTimeSec":        queue.TaskProcessing.ProlongationOptions.ProlongationTimeSec,
		"ProlongationIsTimeoutRetry": queue.TaskProcessing.ProlongationOptions.IsTimeoutRetry,
	}

	if err := s.GetMaster().WithContext(ctx).SelectOne(&queue, query, args); err != nil {
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
// FIXME hot fix WTEL-4333
func sortQueueReportGeneral(val string) string {
	sort, field := orderBy(val)

	switch field {
	case "queue":
		field = "q.name"
	case "team":
		field = "ct.name"
	case "agent_status":
		field = "coalesce(array_length(queue_ag.total, 1), 0)"
	case "missed":
		field = "4"
	case "processed":
		field = "5"
	case "waiting":
		field = "6"
	case "count":
		field = "7"
	case "transferred":
		field = "8"
	case "bridged":
		field = "9"
	case "abandoned":
		field = "10"
	case "sum_bill_sec":
		field = "11"
	case "sl20":
		field = "12"
	case "sl30":
		field = "13"
	case "avg_wrap_sec":
		field = "14"
	case "avg_awt_sec":
		field = "15"
	case "max_awt_sec":
		field = "16"
	case "avg_asa_sec":
		field = "17"
	case "avg_aht_sec":
		field = "18"
	default:
		sort = "desc"
		field = "q.priority"
	}

	return fmt.Sprintf("%s %s", field, sort)
}
func (s SqlQueueStore) QueueReportGeneral(ctx context.Context, domainId int64, supervisorId int64, groups []int, access auth_manager.PermissionAccess, search *model.SearchQueueReportGeneral) (*model.QueueReportGeneralAgg, model.AppError) {
	var report *model.QueueReportGeneralAgg

	err := s.GetMaster().WithContext(ctx).SelectOne(&report, `
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
               array_agg(distinct a.id) filter ( where status = 'online' and ac.state = 'waiting' ) free,
               array_agg(distinct a.id) total
        from queues q
            inner join call_center.cc_agent a on a.domain_id = q.domain_id
            inner join call_center.cc_agent_channel ac on ac.agent_id = a.id and ac.channel = case when q.type in (0,1,2,3,4,5) then  'call'
                                                                                when q.type in (6) then 'chat' else 'task' end
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
           coalesce(case when q.type in (1, 6) then (select count(*) from call_center.cc_member_attempt a1 where a1.queue_id = q.id and a1.bridged_at isnull)
               else (select sum(s.member_waiting) from call_center.cc_queue_statistics s where s.queue_id = q.id) end, 0) waiting,
           coalesce(ag.count, 0) count,
           coalesce(ag.transferred, 0) transferred,
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
    order by `+sortQueueReportGeneral(search.Sort)+`
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

// todo
func (s SqlQueueStore) ListTags(ctx context.Context, domainId int64, search *model.ListRequest) ([]*model.Tag, model.AppError) {
	var res []*model.Tag
	if search.Sort == "" {
		search.Sort = "name"
	}
	st, f := orderBy(search.Sort)
	sort := fmt.Sprintf("order by %s %s", QuoteIdentifier(f), st)

	q := `with tags as (
    select distinct tag as name
    from call_center.cc_queue s,
         unnest(s.tags) tag
    where s.domain_id = :DomainId
        and (:Q::varchar isnull or tag ilike :Q::varchar)
)
select *
from tags
%s
limit :Limit
offset :Offset`

	_, err := s.GetReplica().WithContext(ctx).Select(&res, fmt.Sprintf(q, sort), map[string]interface{}{
		"DomainId": domainId,
		"Q":        search.GetQ(),
		"Limit":    search.GetLimit(),
		"Offset":   search.GetOffset(),
	})

	if err != nil {
		return nil, model.NewCustomCodeError("store.sql_queue.tags.app_error", err.Error(), extractCodeFromErr(err))
	}

	return res, nil
}

func (s SqlQueueStore) GetGlobalState(ctx context.Context, domainId int64) (bool, model.AppError) {
	query := `
		select not exists (
			select 
				1
			from 
				call_center.cc_queue q
			where
				q.domain_id = :DomainId
				and q.enabled <> true
		) as is_all_enabled
	`
	params := map[string]any{
		"DomainId": domainId,
	}

	var isAllEnabled bool
	if err := s.GetReplica().WithContext(ctx).SelectOne(&isAllEnabled, query, params); err != nil {
		return false, model.NewCustomCodeError("sqlstore.sql_queue.get_global_state.app_error", err.Error(), extractCodeFromErr(err))
	}

	return isAllEnabled, nil
}

func (s SqlQueueStore) SetGlobalState(ctx context.Context, domainId int64, newState bool, updatedBy *model.Lookup) (int32, model.AppError) {
	query := `
		update
			call_center.cc_queue q
		set
			updated_by = :UpdatedBy,
			updated_at = :UpdatedAt,
			enabled = :Enabled
		where
			q.domain_id = :DomainId
			and q.enabled <> :Enabled
	`
	params := map[string]any{
		"UpdatedBy": updatedBy.GetSafeId(),
		"UpdatedAt": model.GetMillis(),
		"Enabled":   newState,
		"DomainId":  domainId,
	}

	res, err := s.GetMaster().WithContext(ctx).Exec(query, params)
	if err != nil {
		return -1, model.NewCustomCodeError("sqlstore.sql_queue.set_global_state.app_error", err.Error(), extractCodeFromErr(err))
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return -1, model.NewCustomCodeError("sqlstore.sql_queue.set_global_state.app_error", err.Error(), extractCodeFromErr(err))
	}

	return int32(rowsAffected), nil
}
