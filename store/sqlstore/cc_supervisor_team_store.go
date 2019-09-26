package sqlstore

import (
	"fmt"
	"github.com/webitel/engine/model"
	"github.com/webitel/engine/store"
	"net/http"
)

type SqlSupervisorTeamStore struct {
	SqlStore
}

func NewSqlSupervisorTeamStore(sqlStore SqlStore) store.SupervisorTeamStore {
	us := &SqlSupervisorTeamStore{sqlStore}
	return us
}

func (s SqlSupervisorTeamStore) Create(supervisor *model.SupervisorInTeam) (*model.SupervisorInTeam, *model.AppError) {
	var out *model.SupervisorInTeam
	if err := s.GetMaster().SelectOne(&out, `with i as (
    insert into cc_supervisor_in_team (agent_id, team_id)
    values (:AgentId, :TeamId)
    returning *
)
select i.id, i.team_id, cc_get_lookup(ca.id, u.name) as agent
from i
    inner join cc_agent ca on i.agent_id = ca.id
    inner join directory.wbt_user u on u.id = ca.user_id`,
		map[string]interface{}{
			"AgentId": supervisor.Agent.Id,
			"TeamId":  supervisor.TeamId,
		}); nil != err {
		return nil, model.NewAppError("SqlSupervisorTeamStore.Save", "store.sql_supervisor_team.save.app_error", nil,
			fmt.Sprintf("name=%v teamId=%v, %v", supervisor.Agent.Id, supervisor.TeamId, err.Error()), extractCodeFromErr(err))
	} else {
		return out, nil
	}
}

func (s SqlSupervisorTeamStore) GetAllPage(domainId, teamId int64, offset, limit int) ([]*model.SupervisorInTeam, *model.AppError) {
	var supervisors []*model.SupervisorInTeam

	if _, err := s.GetReplica().Select(&supervisors,
		`select i.id, i.team_id, cc_get_lookup(ca.id, u.name) as agent
				from cc_supervisor_in_team i
					inner join cc_team te on te.id = i.team_id
					inner join cc_agent ca on i.agent_id = ca.id
					inner join directory.wbt_user u on u.id = ca.user_id
				where i.team_id = :TeamId and te.domain_id = :DomainId
				order by i.id
				limit :Limit
				offset :Offset`, map[string]interface{}{
			"DomainId": domainId,
			"TeamId":   teamId,
			"Limit":    limit,
			"Offset":   offset,
		}); err != nil {
		return nil, model.NewAppError("SqlSupervisorTeamStore.GetAllPage", "store.sql_supervisor_team.get_all.app_error",
			nil, err.Error(), extractCodeFromErr(err))
	} else {
		return supervisors, nil
	}
}

func (s SqlSupervisorTeamStore) Get(domainId, teamId, id int64) (*model.SupervisorInTeam, *model.AppError) {
	var supervisor *model.SupervisorInTeam

	if err := s.GetReplica().SelectOne(&supervisor,
		`select i.id, i.team_id, cc_get_lookup(ca.id, u.name) as agent
				from cc_supervisor_in_team i
					inner join cc_team te on te.id = i.team_id
					inner join cc_agent ca on i.agent_id = ca.id
					inner join directory.wbt_user u on u.id = ca.user_id
				where i.id = :Id and i.team_id = :TeamId and te.domain_id = :DomainId
				`, map[string]interface{}{
			"Id":       id,
			"DomainId": domainId,
			"TeamId":   teamId,
		}); err != nil {
		return nil, model.NewAppError("SqlSupervisorTeamStore.Get", "store.sql_supervisor_team.get.app_error",
			nil, err.Error(), extractCodeFromErr(err))
	} else {
		return supervisor, nil
	}
}

func (s SqlSupervisorTeamStore) Update(supervisor *model.SupervisorInTeam) (*model.SupervisorInTeam, *model.AppError) {
	err := s.GetMaster().SelectOne(&supervisor, `with i as (
    update cc_supervisor_in_team
    set agent_id = :AgentId 
    where id = :Id and team_id = :TeamId
    returning *
)
select i.id, i.team_id, cc_get_lookup(ca.id, u.name) as agent
from i
    inner join cc_team te on te.id = i.team_id
    inner join cc_agent ca on i.agent_id = ca.id
    inner join directory.wbt_user u on u.id = ca.user_id`, map[string]interface{}{
		"AgentId": supervisor.Agent.Id,
		"Id":      supervisor.Id,
		"TeamId":  supervisor.TeamId,
	})
	if err != nil {
		return nil, model.NewAppError("SqlResourceTeamStore.Update", "store.sql_resource_team.update.app_error", nil,
			fmt.Sprintf("Id=%v, %s", supervisor.Id, err.Error()), extractCodeFromErr(err))
	}
	return supervisor, nil
}

func (s SqlSupervisorTeamStore) Delete(teamId, id int64) *model.AppError {
	if _, err := s.GetMaster().Exec(`delete from cc_supervisor_in_team c where c.id=:Id and c.team_id = :TeamId`,
		map[string]interface{}{"Id": id, "TeamId": teamId}); err != nil {
		return model.NewAppError("SqlSupervisorTeamStore.Delete", "store.sql_resource_team.delete.app_error", nil,
			fmt.Sprintf("Id=%v, %s", id, err.Error()), http.StatusInternalServerError)
	}
	return nil
}
