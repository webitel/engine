package sqlstore

import (
	"fmt"
	"github.com/lib/pq"
	"github.com/webitel/engine/auth_manager"
	"github.com/webitel/engine/model"
	"github.com/webitel/engine/store"
	"net/http"
)

type SqlAgentTeamStore struct {
	SqlStore
}

func NewSqlAgentTeamStore(sqlStore SqlStore) store.AgentTeamStore {
	us := &SqlAgentTeamStore{sqlStore}
	return us
}

func (s SqlAgentTeamStore) Create(team *model.AgentTeam) (*model.AgentTeam, *model.AppError) {
	var out *model.AgentTeam
	if err := s.GetMaster().SelectOne(&out, `with t as (
    insert into cc_team (domain_id, name, description, strategy, max_no_answer, wrap_up_time,
                     no_answer_delay_time, call_timeout, updated_at, created_at, created_by, updated_by,
                     admin_id)
    values (:DomainId, :Name, :Description, :Strategy, :MaxNoAnswer, :WrapUpTime,
                    :NoAnswerDelayTime, :CallTimeout, :UpdatedAt, :CreatedAt, :CreatedBy,  :UpdatedBy, :AdminId)
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
       t.updated_at,
       adm.user as admin,
       t.domain_id
from t
    left join cc_agent_with_user adm on adm.id = t.admin_id`,
		map[string]interface{}{
			"DomainId":          team.DomainId,
			"Name":              team.Name,
			"Description":       team.Description,
			"Strategy":          team.Strategy,
			"MaxNoAnswer":       team.MaxNoAnswer,
			"WrapUpTime":        team.WrapUpTime,
			"NoAnswerDelayTime": team.NoAnswerDelayTime,
			"CallTimeout":       team.CallTimeout,
			"CreatedAt":         team.CreatedAt,
			"CreatedBy":         team.CreatedBy.Id,
			"UpdatedAt":         team.UpdatedAt,
			"UpdatedBy":         team.UpdatedBy.Id,
			"AdminId":           team.Admin.GetSafeId(),
		}); nil != err {
		return nil, model.NewAppError("SqlAgentTeamStore.Save", "store.sql_agent_team.save.app_error", nil,
			fmt.Sprintf("name=%v, %v", team.Name, err.Error()), extractCodeFromErr(err))
	} else {
		return out, nil
	}
}

func (s SqlAgentTeamStore) CheckAccess(domainId, id int64, groups []int, access auth_manager.PermissionAccess) (bool, *model.AppError) {
	res, err := s.GetReplica().SelectNullInt(`select 1
		where exists(
          select 1
          from cc_team_acl a
          where a.dc = :DomainId
            and a.object = :Id
            and a.subject = any (:Groups::int[])
            and a.access & :Access = :Access
        )`, map[string]interface{}{"DomainId": domainId, "Id": id, "Groups": pq.Array(groups), "Access": access.Value()})

	if err != nil {
		return false, model.NewAppError("SqlAgentTeamStore.CheckAccess", "store.sql_agent_team.access.app_error", nil,
			fmt.Sprintf("id=%v, domain_id=%v %v", id, domainId, err.Error()), http.StatusInternalServerError)
	}

	return (res.Valid && res.Int64 == 1), nil
}

func (s SqlAgentTeamStore) GetAllPage(domainId int64, search *model.SearchAgentTeam) ([]*model.AgentTeam, *model.AppError) {

	var teams []*model.AgentTeam

	f := map[string]interface{}{
		"DomainId": domainId,
		"Ids":      pq.Array(search.Ids),
		"Q":        search.GetQ(),
		"AdminIds": pq.Array(search.AdminIds),
		"Strategy": pq.Array(search.Strategy),
	}

	err := s.ListQuery(&teams, search.ListRequest,
		`domain_id = :DomainId and ( (:Ids::int[] isnull or id = any(:Ids) ) 
			and (:AdminIds::int[] isnull or admin_id = any(:AdminIds) )
			and (:Strategy::varchar[] isnull or strategy = any(:Strategy) )
			and (:Q::varchar isnull or (t.name ilike :Q::varchar or t.description ilike :Q::varchar or t.strategy ilike :Q::varchar ) ) )`,
		model.AgentTeam{}, f)
	if err != nil {
		return nil, model.NewAppError("SqlAgentTeamStore.GetAllPage", "store.sql_agent_team.get_all.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	return teams, nil
}

func (s SqlAgentTeamStore) GetAllPageByGroups(domainId int64, groups []int, search *model.SearchAgentTeam) ([]*model.AgentTeam, *model.AppError) {
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

	err := s.ListQuery(&teams, search.ListRequest,
		`domain_id = :DomainId and (
				exists(select 1
				  from cc_team_acl a
				  where a.dc = t.domain_id and a.object = t.id and a.subject = any(:Groups::int[]) and a.access&:Access = :Access)
			  ) and ( (:Ids::int[] isnull or id = any(:Ids) ) 
			and (:AdministratorIds::int[] isnull or admin_id = any(:AdministratorIds) )
			and (:Strategy::varchar[] isnull or strategy = any(:Strategy) )
			and (:Q::varchar isnull or (t.name ilike :Q::varchar or t.description ilike :Q::varchar or t.strategy ilike :Q::varchar ) ) )`,
		model.AgentTeam{}, f)
	if err != nil {
		return nil, model.NewAppError("SqlAgentTeamStore.GetAllPageByGroups", "store.sql_agent_team.get_all.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	return teams, nil
}

func (s SqlAgentTeamStore) Get(domainId int64, id int64) (*model.AgentTeam, *model.AppError) {
	var team *model.AgentTeam
	if err := s.GetReplica().SelectOne(&team, `select t.id,
       t.name,
       t.description,
       t.strategy,
       t.max_no_answer,
       t.wrap_up_time,
       t.no_answer_delay_time,
       t.call_timeout,
       t.updated_at,
       adm.user as admin
from cc_team t
    left join cc_agent_with_user adm on adm.id = t.admin_id
where t.domain_id = :DomainId and t.id = :Id`, map[string]interface{}{
		"Id":       id,
		"DomainId": domainId,
	}); err != nil {
		return nil, model.NewAppError("SqlAgentTeamStore.Get", "store.sql_agent_team.get.app_error", nil,
			fmt.Sprintf("Id=%v, %s", id, err.Error()), extractCodeFromErr(err))
	} else {
		return team, nil
	}
}

func (s SqlAgentTeamStore) Update(domainId int64, team *model.AgentTeam) (*model.AgentTeam, *model.AppError) {
	err := s.GetMaster().SelectOne(&team, `with t as (
    update cc_team
    set name = :Name,
        description = :Description,
        strategy = :Strategy,
        max_no_answer = :MaxNoAnswer,
        wrap_up_time = :WrapUpTime,
        no_answer_delay_time = :NoAnswerDelayTime,
        call_timeout = :CallTimeout,
        updated_at = :UpdatedAt,
        updated_by = :UpdatedBy,
        admin_id = :AdminId
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
       t.updated_at,
       adm.user as admin,
       t.domain_id
from t
    left join cc_agent_with_user adm on adm.id = t.admin_id`, map[string]interface{}{
		"Id":                team.Id,
		"DomainId":          domainId,
		"Name":              team.Name,
		"Description":       team.Description,
		"Strategy":          team.Strategy,
		"MaxNoAnswer":       team.MaxNoAnswer,
		"WrapUpTime":        team.WrapUpTime,
		"NoAnswerDelayTime": team.NoAnswerDelayTime,
		"CallTimeout":       team.CallTimeout,
		"UpdatedAt":         team.UpdatedAt,
		"UpdatedBy":         team.UpdatedBy.Id,
		"AdminId":           team.Admin.GetSafeId(),
	})
	if err != nil {
		return nil, model.NewAppError("SqlAgentTeamStore.Update", "store.sql_agent_team.update.app_error", nil,
			fmt.Sprintf("Id=%v, %s", team.Id, err.Error()), extractCodeFromErr(err))
	}
	return team, nil
}

func (s SqlAgentTeamStore) Delete(domainId, id int64) *model.AppError {
	if _, err := s.GetMaster().Exec(`delete from cc_team c where c.id=:Id and c.domain_id = :DomainId`,
		map[string]interface{}{"Id": id, "DomainId": domainId}); err != nil {
		return model.NewAppError("SqlAgentTeamStore.Delete", "store.sql_agent_team.delete.app_error", nil,
			fmt.Sprintf("Id=%v, %s", id, err.Error()), http.StatusInternalServerError)
	}
	return nil
}
