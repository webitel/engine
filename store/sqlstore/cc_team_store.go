package sqlstore

import (
	"context"
	"fmt"

	"github.com/lib/pq"

	"github.com/webitel/engine/model"
	"github.com/webitel/engine/pkg/wbt/auth_manager"
	"github.com/webitel/engine/store"
)

type SqlAgentTeamStore struct {
	SqlStore
}

func NewSqlAgentTeamStore(sqlStore SqlStore) store.AgentTeamStore {
	us := &SqlAgentTeamStore{sqlStore}
	return us
}

func (s SqlAgentTeamStore) Create(ctx context.Context, team *model.AgentTeam) (*model.AgentTeam, model.AppError) {
	var out *model.AgentTeam
	if err := s.GetMaster().WithContext(ctx).SelectOne(&out, `with t as (
    insert into call_center.cc_team (domain_id, name, description, strategy, max_no_answer, wrap_up_time,
                     no_answer_delay_time, call_timeout, updated_at, created_at, created_by, updated_by,
                     admin_ids, invite_chat_timeout, task_accept_timeout, forecast_calculation_id)
    values (:DomainId, :Name, :Description, :Strategy, :MaxNoAnswer, :WrapUpTime,
                    :NoAnswerDelayTime, :CallTimeout, :UpdatedAt, :CreatedAt, :CreatedBy,  :UpdatedBy, :AdminIds, :InviteChatTimeout, :TaskAcceptTimeout, :ForecastCalculationId)
    returning *
)
select t.id,
       t.name,
       t.description,
       t.strategy,
       t.max_no_answer,
       t.wrap_up_time,
       t.no_answer_delay_time,
       t.call_timeout,
	   t.invite_chat_timeout,
	   t.task_accept_timeout,
       t.updated_at,
       (SELECT jsonb_agg(adm."user") AS jsonb_agg
        FROM call_center.cc_agent_with_user adm
		WHERE adm.id = any(t.admin_ids)) as admin,
       t.domain_id,
       call_center.cc_get_lookup(fc.id, fc.name) AS forecast_calculation
from t
	left join wfm.forecast_calculation fc on fc.id = t.forecast_calculation_id`,
		map[string]interface{}{
			"DomainId":              team.DomainId,
			"Name":                  team.Name,
			"Description":           team.Description,
			"Strategy":              team.Strategy,
			"MaxNoAnswer":           team.MaxNoAnswer,
			"WrapUpTime":            team.WrapUpTime,
			"NoAnswerDelayTime":     team.NoAnswerDelayTime,
			"CallTimeout":           team.CallTimeout,
			"InviteChatTimeout":     team.InviteChatTimeout,
			"TaskAcceptTimeout":     team.TaskAcceptTimeout,
			"CreatedAt":             team.CreatedAt,
			"CreatedBy":             team.CreatedBy.GetSafeId(),
			"UpdatedAt":             team.UpdatedAt,
			"UpdatedBy":             team.UpdatedBy.GetSafeId(),
			"AdminIds":              pq.Array(model.LookupIds(team.Admin)),
			"ForecastCalculationId": team.ForecastCalculation.GetSafeId(),
		}); nil != err {
		return nil, model.NewCustomCodeError("store.sql_agent_team.save.app_error", fmt.Sprintf("name=%v, %v", team.Name, err.Error()), extractCodeFromErr(err))
	} else {
		return out, nil
	}
}

func (s SqlAgentTeamStore) CheckAccess(ctx context.Context, domainId, id int64, groups []int, access auth_manager.PermissionAccess) (bool, model.AppError) {
	res, err := s.GetReplica().WithContext(ctx).SelectNullInt(`select 1
		where exists(
          select 1
          from call_center.cc_team_acl a
          where a.dc = :DomainId
            and a.object = :Id
            and a.subject = any (:Groups::int[])
            and a.access & :Access = :Access
        )`, map[string]interface{}{"DomainId": domainId, "Id": id, "Groups": pq.Array(groups), "Access": access.Value()})

	if err != nil {
		return false, model.NewInternalError("store.sql_agent_team.access.app_error", fmt.Sprintf("id=%v, domain_id=%v %v", id, domainId, err.Error()))
	}

	return (res.Valid && res.Int64 == 1), nil
}

func (s SqlAgentTeamStore) GetAllPage(ctx context.Context, domainId int64, search *model.SearchAgentTeam) ([]*model.AgentTeam, model.AppError) {

	var teams []*model.AgentTeam

	f := map[string]interface{}{
		"DomainId": domainId,
		"Ids":      pq.Array(search.Ids),
		"Q":        search.GetQ(),
		"AdminIds": pq.Array(search.AdminIds),
		"Strategy": pq.Array(search.Strategy),
	}

	err := s.ListQuery(ctx, &teams, search.ListRequest,
		`domain_id = :DomainId and ( (:Ids::int[] isnull or id = any(:Ids) ) 
			and (:AdminIds::int[] isnull or admin_ids && :AdminIds )
			and (:Strategy::varchar[] isnull or strategy = any(:Strategy) )
			and (:Q::varchar isnull or (t.name ilike :Q::varchar or t.description ilike :Q::varchar or t.strategy ilike :Q::varchar ) ) )`,
		model.AgentTeam{}, f)
	if err != nil {
		return nil, model.NewInternalError("store.sql_agent_team.get_all.app_error", err.Error())
	}

	return teams, nil
}

func (s SqlAgentTeamStore) GetAllPageByGroups(ctx context.Context, domainId int64, groups []int, search *model.SearchAgentTeam) ([]*model.AgentTeam, model.AppError) {
	var teams []*model.AgentTeam

	f := map[string]interface{}{
		"Groups":   pq.Array(groups),
		"Access":   auth_manager.PERMISSION_ACCESS_READ.Value(),
		"DomainId": domainId,
		"Ids":      pq.Array(search.Ids),
		"Q":        search.GetQ(),
		"AdminIds": pq.Array(search.AdminIds),
		"Strategy": pq.Array(search.Strategy),
	}

	err := s.ListQuery(ctx, &teams, search.ListRequest,
		`domain_id = :DomainId and (
				exists(select 1
				  from call_center.cc_team_acl a
				  where a.dc = t.domain_id and a.object = t.id and a.subject = any(:Groups::int[]) and a.access&:Access = :Access)
			  ) and ( (:Ids::int[] isnull or id = any(:Ids) ) 
			and (:AdminIds::int[] isnull or admin_ids && :AdminIds )
			and (:Strategy::varchar[] isnull or strategy = any(:Strategy) )
			and (:Q::varchar isnull or (t.name ilike :Q::varchar or t.description ilike :Q::varchar or t.strategy ilike :Q::varchar ) ) )`,
		model.AgentTeam{}, f)
	if err != nil {
		return nil, model.NewInternalError("store.sql_agent_team.get_all.app_error", err.Error())
	}

	return teams, nil
}

func (s SqlAgentTeamStore) Get(ctx context.Context, domainId int64, id int64) (*model.AgentTeam, model.AppError) {
	var team *model.AgentTeam
	if err := s.GetReplica().WithContext(ctx).SelectOne(&team, `select t.id,
       t.name,
       t.description,
       t.strategy,
       t.max_no_answer,
       t.wrap_up_time,
       t.no_answer_delay_time,
       t.call_timeout,
       t.invite_chat_timeout,
       t.task_accept_timeout,
       t.updated_at,
       (SELECT jsonb_agg(adm."user") AS jsonb_agg
        FROM call_center.cc_agent_with_user adm
		WHERE adm.id = any(t.admin_ids)) as admin,
	    call_center.cc_get_lookup(fc.id, fc.name) AS forecast_calculation
from call_center.cc_team t
	left join wfm.forecast_calculation fc on fc.id = t.forecast_calculation_id
where t.domain_id = :DomainId and t.id = :Id`, map[string]interface{}{
		"Id":       id,
		"DomainId": domainId,
	}); err != nil {
		return nil, model.NewCustomCodeError("store.sql_agent_team.get.app_error", fmt.Sprintf("Id=%v, %s", id, err.Error()), extractCodeFromErr(err))
	} else {
		return team, nil
	}
}

func (s SqlAgentTeamStore) Update(ctx context.Context, domainId int64, team *model.AgentTeam) (*model.AgentTeam, model.AppError) {
	err := s.GetMaster().WithContext(ctx).SelectOne(&team, `with t as (
    update call_center.cc_team
    set name = :Name,
        description = :Description,
        strategy = :Strategy,
        max_no_answer = :MaxNoAnswer,
        wrap_up_time = :WrapUpTime,
        no_answer_delay_time = :NoAnswerDelayTime,
        call_timeout = :CallTimeout,
        invite_chat_timeout = :InviteChatTimeout,
        task_accept_timeout = :TaskAcceptTimeout,
        updated_at = :UpdatedAt,
        updated_by = :UpdatedBy,
        admin_ids = :AdminIds,
		forecast_calculation_id = :ForecastCalculationId
    where id = :Id and domain_id = :DomainId
    returning *
)
select t.id,
       t.name,
       t.description,
       t.strategy,
       t.max_no_answer,
       t.wrap_up_time,
       t.no_answer_delay_time,
       t.call_timeout,
       t.invite_chat_timeout,
       t.task_accept_timeout,
       t.updated_at,
       (SELECT jsonb_agg(adm."user") AS jsonb_agg
        FROM call_center.cc_agent_with_user adm
		WHERE adm.id = any(t.admin_ids)) as admin,
       t.domain_id,
		call_center.cc_get_lookup(fc.id, fc.name) AS forecast_calculation
from t
	left join wfm.forecast_calculation fc on fc.id = t.forecast_calculation_id`, map[string]interface{}{
		"Id":                    team.Id,
		"DomainId":              domainId,
		"Name":                  team.Name,
		"Description":           team.Description,
		"Strategy":              team.Strategy,
		"MaxNoAnswer":           team.MaxNoAnswer,
		"WrapUpTime":            team.WrapUpTime,
		"NoAnswerDelayTime":     team.NoAnswerDelayTime,
		"CallTimeout":           team.CallTimeout,
		"InviteChatTimeout":     team.InviteChatTimeout,
		"TaskAcceptTimeout":     team.TaskAcceptTimeout,
		"UpdatedAt":             team.UpdatedAt,
		"UpdatedBy":             team.UpdatedBy.GetSafeId(),
		"AdminIds":              pq.Array(model.LookupIds(team.Admin)),
		"ForecastCalculationId": team.ForecastCalculation.GetSafeId(),
	})
	if err != nil {
		return nil, model.NewCustomCodeError("store.sql_agent_team.update.app_error", fmt.Sprintf("Id=%v, %s", team.Id, err.Error()), extractCodeFromErr(err))
	}
	return team, nil
}

func (s SqlAgentTeamStore) Delete(ctx context.Context, domainId, id int64) model.AppError {
	if _, err := s.GetMaster().WithContext(ctx).Exec(`delete from call_center.cc_team c where c.id=:Id and c.domain_id = :DomainId`,
		map[string]interface{}{"Id": id, "DomainId": domainId}); err != nil {
		return model.NewInternalError("store.sql_agent_team.delete.app_error", fmt.Sprintf("Id=%v, %s", id, err.Error()))
	}
	return nil
}
