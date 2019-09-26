package store

import (
	"time"

	"github.com/webitel/engine/model"
)

type StoreResult struct {
	Data interface{}
	Err  *model.AppError
}

type StoreChannel chan StoreResult

func Do(f func(result *StoreResult)) StoreChannel {
	storeChannel := make(StoreChannel, 1)
	go func() {
		result := StoreResult{}
		f(&result)
		storeChannel <- result
		close(storeChannel)
	}()
	return storeChannel
}

func Must(sc StoreChannel) interface{} {
	r := <-sc
	if r.Err != nil {

		time.Sleep(time.Second)
		panic(r.Err)
	}

	return r.Data
}

type Store interface {
	Calendar() CalendarStore
	Skill() SkillStore
	AgentTeam() AgentTeamStore
	Agent() AgentStore
	AgentSkill() AgentSkillStore
	ResourceTeam() ResourceTeamStore
	Queue() QueueStore
	SupervisorTeam() SupervisorTeamStore
	OutboundResource() OutboundResourceStore
	RoutingScheme() RoutingSchemeStore
	RoutingInboundCall() RoutingInboundCallStore
	RoutingOutboundCall() RoutingOutboundCallStore
	RoutingVariable() RoutingVariableStore
}

type CalendarStore interface {
	Create(calendar *model.Calendar) (*model.Calendar, *model.AppError)
	GetAllPage(domainId int64, offset, limit int) ([]*model.Calendar, *model.AppError)
	GetAllPageByGroups(domainId int64, groups []int, offset, limit int) ([]*model.Calendar, *model.AppError)
	Get(domainId int64, id int64) (*model.Calendar, *model.AppError)
	Update(calendar *model.Calendar) (*model.Calendar, *model.AppError)
	Delete(domainId, id int64) *model.AppError

	GetTimezoneAllPage(offset, limit int) ([]*model.Timezone, *model.AppError)

	CheckAccess(domainId, id int64, groups []int, access model.PermissionAccess) (bool, *model.AppError)

	GetAcceptOfDay(calendarId int64) ([]*model.CalendarAcceptOfDay, *model.AppError)
}

type SkillStore interface {
	Create(skill *model.Skill) (*model.Skill, *model.AppError)
	Get(domainId int64, id int64) (*model.Skill, *model.AppError)
	GetAllPage(domainId int64, offset, limit int) ([]*model.Skill, *model.AppError)
	Delete(domainId, id int64) *model.AppError
	Update(skill *model.Skill) (*model.Skill, *model.AppError)
}

type AgentTeamStore interface {
	CheckAccess(domainId, id int64, groups []int, access model.PermissionAccess) (bool, *model.AppError)

	Create(team *model.AgentTeam) (*model.AgentTeam, *model.AppError)
	GetAllPage(domainId int64, offset, limit int) ([]*model.AgentTeam, *model.AppError)
	GetAllPageByGroups(domainId int64, groups []int, offset, limit int) ([]*model.AgentTeam, *model.AppError)
	Get(domainId int64, id int64) (*model.AgentTeam, *model.AppError)
	Update(team *model.AgentTeam) (*model.AgentTeam, *model.AppError)
	Delete(domainId, id int64) *model.AppError
}

type AgentStore interface {
	CheckAccess(domainId, id int64, groups []int, access model.PermissionAccess) (bool, *model.AppError)

	Create(agent *model.Agent) (*model.Agent, *model.AppError)
	GetAllPage(domainId int64, offset, limit int) ([]*model.Agent, *model.AppError)
	GetAllPageByGroups(domainId int64, groups []int, offset, limit int) ([]*model.Agent, *model.AppError)
	Get(domainId int64, id int64) (*model.Agent, *model.AppError)
	Update(agent *model.Agent) (*model.Agent, *model.AppError)
	Delete(domainId, id int64) *model.AppError
	SetStatus(domainId, agentId int64, status string, payload interface{}) (bool, *model.AppError)
}

type AgentSkillStore interface {
	Create(agent *model.AgentSkill) (*model.AgentSkill, *model.AppError)
	GetById(domainId, agentId, id int64) (*model.AgentSkill, *model.AppError)
	Update(agentSkill *model.AgentSkill) (*model.AgentSkill, *model.AppError)
	GetAllPage(domainId, agentId int64, offset, limit int) ([]*model.AgentSkill, *model.AppError)
	Delete(agentId, id int64) *model.AppError
}

type ResourceTeamStore interface {
	Create(in *model.ResourceInTeam) (*model.ResourceInTeam, *model.AppError)
	Get(domainId, teamId int64, id int64) (*model.ResourceInTeam, *model.AppError)
	GetAllPage(domainId, teamId int64, offset, limit int) ([]*model.ResourceInTeam, *model.AppError)
	Update(resource *model.ResourceInTeam) (*model.ResourceInTeam, *model.AppError)
	Delete(domainId, teamId int64, id int64) *model.AppError
}

type OutboundResourceStore interface {
	CheckAccess(domainId, id int64, groups []int, access model.PermissionAccess) (bool, *model.AppError)
	Create(resource *model.OutboundCallResource) (*model.OutboundCallResource, *model.AppError)
	GetAllPage(domainId int64, offset, limit int) ([]*model.OutboundCallResource, *model.AppError)
	GetAllPageByGroups(domainId int64, groups []int, offset, limit int) ([]*model.OutboundCallResource, *model.AppError)
	Get(domainId int64, id int64) (*model.OutboundCallResource, *model.AppError)
	Update(resource *model.OutboundCallResource) (*model.OutboundCallResource, *model.AppError)
	Delete(domainId, id int64) *model.AppError
}

type RoutingSchemeStore interface {
	Create(scheme *model.RoutingScheme) (*model.RoutingScheme, *model.AppError)
	GetAllPage(domainId int64, offset, limit int) ([]*model.RoutingScheme, *model.AppError)
	Get(domainId int64, id int64) (*model.RoutingScheme, *model.AppError)
	Update(scheme *model.RoutingScheme) (*model.RoutingScheme, *model.AppError)
	Delete(domainId, id int64) *model.AppError
}

type RoutingInboundCallStore interface {
	Create(routing *model.RoutingInboundCall) (*model.RoutingInboundCall, *model.AppError)
	GetAllPage(domainId int64, offset, limit int) ([]*model.RoutingInboundCall, *model.AppError)
	Get(domainId, id int64) (*model.RoutingInboundCall, *model.AppError)
	Update(routing *model.RoutingInboundCall) (*model.RoutingInboundCall, *model.AppError)
	Delete(domainId, id int64) *model.AppError
}

type RoutingOutboundCallStore interface {
	Create(routing *model.RoutingOutboundCall) (*model.RoutingOutboundCall, *model.AppError)
	GetAllPage(domainId int64, offset, limit int) ([]*model.RoutingOutboundCall, *model.AppError)
	Get(domainId, id int64) (*model.RoutingOutboundCall, *model.AppError)
	Update(routing *model.RoutingOutboundCall) (*model.RoutingOutboundCall, *model.AppError)
	Delete(domainId, id int64) *model.AppError
}

type RoutingVariableStore interface {
	Create(variable *model.RoutingVariable) (*model.RoutingVariable, *model.AppError)
	GetAllPage(domainId int64, offset, limit int) ([]*model.RoutingVariable, *model.AppError)
	Get(domainId int64, id int64) (*model.RoutingVariable, *model.AppError)
	Update(variable *model.RoutingVariable) (*model.RoutingVariable, *model.AppError)
	Delete(domainId, id int64) *model.AppError
}

type QueueStore interface {
	CheckAccess(domainId, id int64, groups []int, access model.PermissionAccess) (bool, *model.AppError)
	Create(queue *model.Queue) (*model.Queue, *model.AppError)
	GetAllPage(domainId int64, offset, limit int) ([]*model.Queue, *model.AppError)
	GetAllPageByGroups(domainId int64, groups []int, offset, limit int) ([]*model.Queue, *model.AppError)
	Get(domainId int64, id int64) (*model.Queue, *model.AppError)
	Update(queue *model.Queue) (*model.Queue, *model.AppError)
	Delete(domainId, id int64) *model.AppError
}

type SupervisorTeamStore interface {
	Create(supervisor *model.SupervisorInTeam) (*model.SupervisorInTeam, *model.AppError)
	GetAllPage(domainId, teamId int64, offset, limit int) ([]*model.SupervisorInTeam, *model.AppError)
	Get(domainId, teamId, id int64) (*model.SupervisorInTeam, *model.AppError)
	Update(supervisor *model.SupervisorInTeam) (*model.SupervisorInTeam, *model.AppError)
	Delete(teamId, id int64) *model.AppError
}
