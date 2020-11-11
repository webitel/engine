package sqlstore

import (
	"fmt"
	"github.com/lib/pq"
	"github.com/webitel/engine/model"
	"github.com/webitel/engine/store"
	"net/http"
	"strings"
)

type SqlCallStore struct {
	SqlStore
}

func NewSqlCallStore(sqlStore SqlStore) store.CallStore {
	us := &SqlCallStore{sqlStore}
	return us
}

func (s SqlCallStore) GetActive(domainId int64, search *model.SearchCall) ([]*model.Call, *model.AppError) {
	var out []*model.Call

	f := map[string]interface{}{
		"Domain":       domainId,
		"Limit":        search.GetLimit(),
		"Offset":       search.GetOffset(),
		"From":         model.GetBetweenFromTime(search.CreatedAt),
		"To":           model.GetBetweenToTime(search.CreatedAt),
		"Q":            search.GetQ(),
		"UserIds":      pq.Array(search.UserIds),
		"QueueIds":     pq.Array(search.QueueIds),
		"TeamIds":      pq.Array(search.TeamIds),
		"AgentIds":     pq.Array(search.AgentIds),
		"MemberIds":    pq.Array(search.MemberIds),
		"GatewayIds":   pq.Array(search.GatewayIds),
		"SkipParent":   search.SkipParent,
		"ParentId":     search.ParentId,
		"Number":       search.Number,
		"Direction":    pq.Array(search.Direction),
		"Missed":       search.Missed,
		"AnsweredFrom": model.GetBetweenFromTime(search.AnsweredAt),
		"AnsweredTo":   model.GetBetweenToTime(search.AnsweredAt),
		"DurationFrom": model.GetBetweenFrom(search.Duration),
		"DurationTo":   model.GetBetweenTo(search.Duration),
	}

	err := s.ListQuery(&out, search.ListRequest,
		`domain_id = :Domain and direction notnull
	and (:Q::text isnull or destination ~ :Q  or  from_number ~ :Q or  to_number ~ :Q)
	and ( (:From::timestamptz isnull or :To::timestamptz isnull) or created_at between :From and :To )
	and (:UserIds::int8[] isnull or user_id = any(:UserIds))
	and (:QueueIds::int[] isnull or queue_id = any(:QueueIds) )
	and (:TeamIds::int[] isnull or team_id = any(:TeamIds) )  
	and (:AgentIds::int[] isnull or agent_id = any(:AgentIds) )
	and (:MemberIds::int8[] isnull or member_id = any(:MemberIds) )
	and (:GatewayIds::int8[] isnull or gateway_id = any(:GatewayIds) )
	and (:Number::varchar isnull or from_number ilike :Number::varchar or to_number ilike :Number::varchar or destination ilike :Number::varchar)
	and ( (:SkipParent::bool isnull or not :SkipParent::bool is true ) or parent_id isnull)
	and (:ParentId::varchar isnull or parent_id = :ParentId )
	and ( (:AnsweredFrom::timestamptz isnull or :AnsweredTo::timestamptz isnull) or answered_at between :AnsweredFrom and :AnsweredTo )
	and ( (:DurationFrom::int8 isnull or :DurationTo::int8 isnull) or duration between :DurationFrom and :DurationTo )
	and (:Direction::varchar[] isnull or direction = any(:Direction) )
	and (:Missed::bool isnull or (:Missed and answered_at isnull))`,
		model.Call{}, f)
	if err != nil {
		return nil, model.NewAppError("SqlCallStore.GetActive", "store.sql_call.get_active.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	return out, nil
}

// fixme
func (s SqlCallStore) GetUserActiveCall(domainId, userId int64) ([]*model.Call, *model.AppError) {
	var res []*model.Call
	_, err := s.GetMaster().Select(&res, `select
       row_to_json(at) task,
       "id", "app_id", "state", "timestamp", "parent_id", "user", "extension", "gateway", "direction", "destination", "from", "to", "variables",
		"created_at", "answered_at", "bridged_at", "hangup_at", "duration", "hold_sec", "wait_sec", "bill_sec",
		"queue", "member", "team", "agent", "joined_at", "leaving_at", "reporting_at", "queue_bridged_at",
		"queue_wait_sec", "queue_duration_sec", "reporting_sec", "display"
from cc_call_active_list c
    left join lateral (
    select a.id as attempt_id, a.channel, a.queue_id, a.member_id, a.member_call_id as member_channel_id,
           a.agent_call_id as agent_channel_id, a.destination as communication,
           t.post_processing as reporting
    from cc_member_attempt a
        inner join cc_queue q on q.id = a.queue_id
        inner join cc_team t on t.id = q.team_id
    where a.id = c.attempt_id and a.agent_call_id = c.id
) at on true
where user_id = :UserId and domain_id = :DomainId`, map[string]interface{}{
		"UserId":   userId,
		"DomainId": domainId,
	})

	if err != nil {
		return nil, model.NewAppError("SqlCallStore.GetUserActiveCall", "store.sql_call.get_user_active.app_error", nil, err.Error(), extractCodeFromErr(err))
	}

	return res, nil
}

func (s SqlCallStore) Get(domainId int64, id string) (*model.Call, *model.AppError) {
	var out *model.Call

	err := s.GetMaster().SelectOne(&out, `
select c.id, c.app_id, c.state, c."timestamp", c.direction, c.destination, c.parent_id, c.created_at,
   json_build_object('type', coalesce(c.from_type, ''), 'number', coalesce(c.from_number, ''), 'id', coalesce(c.from_id, ''), 'name', coalesce(c.from_name, '')) "from",
   json_build_object('type', coalesce(c.to_type, ''), 'number', coalesce(c.to_number, ''), 'id', coalesce(c.to_id, ''), 'name', coalesce(c.to_name, '')) "to",
   (extract(epoch from now() -  c.created_at))::int8 duration	
from cc_calls c
where c.domain_id = :Domain and c.id = :Id`, map[string]interface{}{
		"Domain": domainId,
		"Id":     id,
	})

	if err != nil {
		return nil, model.NewAppError("SqlCallStore.Get", "store.sql_call.get.app_error", nil, err.Error(), extractCodeFromErr(err))
	}

	return out, nil
}

func (s SqlCallStore) GetInstance(domainId int64, id string) (*model.CallInstance, *model.AppError) {
	var inst *model.CallInstance
	err := s.GetMaster().SelectOne(&inst, `select c.id, c.app_id, c.state
from cc_calls c
where c.id = :Id and c.domain_id = :Domain`, map[string]interface{}{
		"Id":     id,
		"Domain": domainId,
	})
	if err != nil {
		return nil, model.NewAppError("SqlCallStore.GetInstance", "store.sql_call.get_instance.app_error", nil, err.Error(), extractCodeFromErr(err))
	}

	return inst, nil
}

func (s SqlCallStore) GetHistory(domainId int64, search *model.SearchHistoryCall) ([]*model.HistoryCall, *model.AppError) {
	var out []*model.HistoryCall

	f := map[string]interface{}{
		"Domain":          domainId,
		"Limit":           search.GetLimit(),
		"Offset":          search.GetOffset(),
		"From":            model.GetBetweenFromTime(search.CreatedAt),
		"To":              model.GetBetweenToTime(search.CreatedAt),
		"Q":               search.GetQ(),
		"UserIds":         pq.Array(search.UserIds),
		"QueueIds":        pq.Array(search.QueueIds),
		"TeamIds":         pq.Array(search.TeamIds),
		"AgentIds":        pq.Array(search.AgentIds),
		"MemberIds":       pq.Array(search.MemberIds),
		"GatewayIds":      pq.Array(search.GatewayIds),
		"SkipParent":      search.SkipParent,
		"ParentId":        search.ParentId,
		"Number":          search.Number,
		"Cause":           search.Cause,
		"HasFile":         search.HasFile,
		"Direction":       search.Direction,
		"Missed":          search.Missed,
		"AnsweredFrom":    model.GetBetweenFromTime(search.AnsweredAt),
		"AnsweredTo":      model.GetBetweenToTime(search.AnsweredAt),
		"DurationFrom":    model.GetBetweenFrom(search.Duration),
		"DurationTo":      model.GetBetweenTo(search.Duration),
		"StoredAtFrom":    model.GetBetweenFromTime(search.StoredAt),
		"StoredAtTo":      model.GetBetweenToTime(search.StoredAt),
		"Ids":             pq.Array(search.Ids),
		"TransferFromIds": pq.Array(search.TransferFromIds),
		"TransferToIds":   pq.Array(search.TransferToIds),
		"DependencyIds":   pq.Array(search.DependencyIds),
		"Tags":            pq.Array(search.Tags),
	}

	err := s.ListQuery(&out, search.ListRequest,
		`domain_id = :Domain 
	and (:Q::text isnull or destination ~ :Q  or  from_number ~ :Q or  to_number ~ :Q or id = :Q)
	and ( (:From::timestamptz isnull or :To::timestamptz isnull) or created_at between :From and :To )
	and ( (:StoredAtFrom::timestamptz isnull or :StoredAtTo::timestamptz isnull) or stored_at between :StoredAtFrom and :StoredAtTo )
	and (:UserIds::int8[] isnull or user_id = any(:UserIds))
	and (:Ids::varchar[] isnull or id = any(:Ids))
	and (:TransferFromIds::varchar[] isnull or transfer_from = any(:TransferFromIds))
	and (:TransferToIds::varchar[] isnull or transfer_to = any(:TransferToIds))
	and (:QueueIds::int[] isnull or queue_id = any(:QueueIds) )
	and (:TeamIds::int[] isnull or team_id = any(:TeamIds) )  
	and (:AgentIds::int[] isnull or agent_id = any(:AgentIds) )
	and (:MemberIds::int8[] isnull or member_id = any(:MemberIds) )
	and (:GatewayIds::int8[] isnull or gateway_id = any(:GatewayIds) )
	and (:Number::varchar isnull or from_number ilike :Number::varchar or to_number ilike :Number::varchar or destination ilike :Number::varchar)
	and ( (:SkipParent::bool isnull or not :SkipParent::bool is true ) or parent_id isnull)
	and (:ParentId::varchar isnull or parent_id = :ParentId )
	and (:HasFile::bool is not true or files notnull )
	and (:Cause::varchar isnull or cause = :Cause )
	and ( (:AnsweredFrom::timestamptz isnull or :AnsweredTo::timestamptz isnull) or answered_at between :AnsweredFrom and :AnsweredTo )
	and ( (:DurationFrom::int8 isnull or :DurationTo::int8 isnull) or duration between :DurationFrom and :DurationTo )
	and (:Direction::varchar isnull or direction = :Direction )
	and (:Missed::bool isnull or (:Missed and answered_at isnull))
	and (:Tags::varchar[] isnull or (tags && :Tags))
	and (:DependencyIds::varchar[] isnull or id in (
		with recursive a as (
			select t.id
			from cc_calls_history t
			where id = any(:DependencyIds)
			union all
			select t.id
			from cc_calls_history t, a
			where t.parent_id = a.id or t.transfer_from = a.id
		)
		select id
		from a
		where not a.id = any(:DependencyIds)
	))
`,
		model.HistoryCall{}, f)
	if err != nil {
		return nil, model.NewAppError("SqlCallStore.GetHistory", "store.sql_call.get_history.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	return out, nil
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

//TODO
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

func (s SqlCallStore) Aggregate(domainId int64, aggs *model.CallAggregate) ([]*model.AggregateResult, *model.AppError) {

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
	from cc_calls_history h
		left join cc_agent ca on h.agent_id = ca.id
		left join directory.wbt_user ua on ua.id = ca.user_id
		left join directory.wbt_user u on u.id = h.user_id
		left join directory.sip_gateway g on g.id = h.gateway_id
		left join cc_queue q on q.id = h.queue_id
		left join cc_team t on t.id = h.team_id
	where h.domain_id = :Domain 
		and (:Q::text isnull or h.destination ~ :Q  or  h.from_number ~ :Q or  h.to_number ~ :Q or h.id = :Q)
		and ( (:From::timestamptz isnull or :To::timestamptz isnull) or h.created_at between :From and :To )
		and ( (:StoredAtFrom::timestamptz isnull or :StoredAtTo::timestamptz isnull) or h.stored_at between :StoredAtFrom and :StoredAtTo )
		and (:UserIds::int8[] isnull or h.user_id = any(:UserIds))
		and (:Ids::varchar[] isnull or h.id = any(:Ids))
		and (:TransferFromIds::varchar[] isnull or h.transfer_from = any(:TransferFromIds))
		and (:TransferToIds::varchar[] isnull or h.transfer_to = any(:TransferToIds))
		and (:QueueIds::int[] isnull or h.queue_id = any(:QueueIds) )
		and (:TeamIds::int[] isnull or h.team_id = any(:TeamIds) )  
		and (:AgentIds::int[] isnull or h.agent_id = any(:AgentIds) )
		and (:MemberIds::int8[] isnull or h.member_id = any(:MemberIds) )
		and (:GatewayIds::int8[] isnull or h.gateway_id = any(:GatewayIds) )
		and (:Number::varchar isnull or h.from_number ilike :Number::varchar or h.to_number ilike :Number::varchar or h.destination ilike :Number::varchar)
		and ( (:SkipParent::bool isnull or not :SkipParent::bool is true ) or h.parent_id isnull)
		and (:ParentId::varchar isnull or h.parent_id = :ParentId )
		and (:CauseArr::varchar[] isnull or h.cause = any(:CauseArr) )
		and ( (:AnsweredFrom::timestamptz isnull or :AnsweredTo::timestamptz isnull) or h.answered_at between :AnsweredFrom and :AnsweredTo )
		and (:Directions::varchar[] isnull or h.direction = any(:Directions) )
		and (:Missed::bool isnull or (:Missed and h.answered_at isnull))
		and (:Tags::varchar[] isnull or (h.tags && :Tags))
		and (:DependencyIds::varchar[] isnull or h.id in (
			with recursive a as (
				select t.id
				from cc_calls_history t
				where t.id = any(:DependencyIds)
				union all
				select t.id
				from cc_calls_history t, a
				where t.parent_id = a.id or t.transfer_from = a.id
			)
			select a.id
			from a
			where not a.id = any(:DependencyIds)
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
		"Number":          aggs.Number,
		"CauseArr":        pq.Array(aggs.CauseArr),
		"Directions":      pq.Array(aggs.Directions),
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
	}

	for i, v := range aggs.Aggs {
		if i > 0 {
			sql += "union all "
		}
		sql += "select " + QuoteLiteral(v.Name) + " as name, (select data from " + QuoteIdentifier(v.Name) + ") as data "
	}

	var res []*model.AggregateResult

	_, err := s.GetReplica().Select(&res, sql, f)
	if err != nil {
		return nil,
			model.NewAppError("SqlCallStore.Aggregate", "store.sql_call.aggregate.app_error", nil, err.Error(), extractCodeFromErr(err))
	}

	return res, nil
}

func (s SqlCallStore) BridgeInfo(domainId int64, fromId, toId string) (*model.BridgeCall, *model.AppError) {
	var res *model.BridgeCall
	err := s.GetMaster().SelectOne(&res, `select coalesce(c.bridged_id, c.id) from_id, coalesce(c2.bridged_id, c2.id) to_id, c.app_id
from cc_calls c,
     cc_calls c2
where c.id = :FromId and c2.id = :ToId and c.domain_id = :DomainId and c2.domain_id = :DomainId`, map[string]interface{}{
		"DomainId": domainId,
		"FromId":   fromId,
		"ToId":     toId,
	})
	if err != nil {
		return nil, model.NewAppError("SqlCallStore.GetBridgeInfo", "store.sql_call.get_bridge_info.app_error", nil, err.Error(), extractCodeFromErr(err))
	} else {
		return res, nil
	}
}

func (s SqlCallStore) LastFile(domainId int64, id string) (int64, *model.AppError) {
	fileId, err := s.GetReplica().SelectInt(`select f.id
from storage.files f
where f.domain_id = :DomainId and f.uuid = (
    select coalesce(c.parent_id, c.id)
    from cc_calls_history c
    where c.id = :Id and c.domain_id = :DomainId
    limit 1
)`, map[string]interface{}{
		"DomainId": domainId,
		"Id":       id,
	})

	if err != nil {
		return 0, model.NewAppError("SqlCallStore.LastFile", "store.sql_call.get_last_file.app_error", nil, err.Error(), extractCodeFromErr(err))
	}

	return fileId, nil
}

func (s SqlCallStore) BridgedId(id string) (string, *model.AppError) {
	res, err := s.GetReplica().SelectStr(`select coalesce(c.bridged_id, c.parent_id, c.id)
from call_center.cc_calls c
where id = :Id`, map[string]string{
		"Id": id,
	})

	if err != nil {
		return "", model.NewAppError("SqlCallStore.BridgedId", "store.sql_call.get_bridge_id.app_error", nil, err.Error(), extractCodeFromErr(err))
	} else {
		return res, nil
	}
}
