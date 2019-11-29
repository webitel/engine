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
	User() UserStore
	Calendar() CalendarStore
	Skill() SkillStore
	AgentTeam() AgentTeamStore
	Agent() AgentStore
	AgentSkill() AgentSkillStore
	ResourceTeam() ResourceTeamStore
	Queue() QueueStore
	Bucket() BucketSore
	BucketInQueue() BucketInQueueStore
	QueueRouting() QueueRoutingStore
	SupervisorTeam() SupervisorTeamStore
	OutboundResource() OutboundResourceStore
	OutboundResourceGroup() OutboundResourceGroupStore
	OutboundResourceInGroup() OutboundResourceInGroupStore
	CommunicationType() CommunicationTypeStore
	List() ListStore

	Member() MemberStore

	RoutingScheme() RoutingSchemeStore
	RoutingInboundCall() RoutingInboundCallStore
	RoutingOutboundCall() RoutingOutboundCallStore
	RoutingVariable() RoutingVariableStore
}

type UserStore interface {
	GetCallInfo(userId, domainId int64) (*model.UserCallInfo, *model.AppError)
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

	CreateAcceptOfDay(domainId, calendarId int64, timeRange *model.CalendarAcceptOfDay) (*model.CalendarAcceptOfDay, *model.AppError)
	GetAcceptOfDayById(domainId, calendarId, id int64) (*model.CalendarAcceptOfDay, *model.AppError)
	GetAcceptOfDayAllPage(calendarId int64) ([]*model.CalendarAcceptOfDay, *model.AppError)
	UpdateAcceptOfDay(calendarId int64, rangeTime *model.CalendarAcceptOfDay) (*model.CalendarAcceptOfDay, *model.AppError)
	DeleteAcceptOfDay(domainId, calendarId, id int64) *model.AppError

	CreateExcept(domainId, calendarId int64, except *model.CalendarExceptDate) (*model.CalendarExceptDate, *model.AppError)
	GetExceptById(domainId, calendarId, id int64) (*model.CalendarExceptDate, *model.AppError)
	GetExceptAllPage(calendarId int64) ([]*model.CalendarExceptDate, *model.AppError)
	UpdateExceptDate(calendarId int64, except *model.CalendarExceptDate) (*model.CalendarExceptDate, *model.AppError)
	DeleteExceptDate(domainId, calendarId, id int64) *model.AppError
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

	SaveDisplay(d *model.ResourceDisplay) (*model.ResourceDisplay, *model.AppError)
	GetDisplayAllPage(domainId, resourceId int64, offset, limit int) ([]*model.ResourceDisplay, *model.AppError)
	GetDisplay(domainId, resourceId, id int64) (*model.ResourceDisplay, *model.AppError)
	UpdateDisplay(domainId int64, display *model.ResourceDisplay) (*model.ResourceDisplay, *model.AppError)
	DeleteDisplay(domainId, resourceId, id int64) *model.AppError
}

type OutboundResourceGroupStore interface {
	CheckAccess(domainId, id int64, groups []int, access model.PermissionAccess) (bool, *model.AppError)
	Create(group *model.OutboundResourceGroup) (*model.OutboundResourceGroup, *model.AppError)
	GetAllPage(domainId int64, offset, limit int) ([]*model.OutboundResourceGroup, *model.AppError)
	GetAllPageByGroups(domainId int64, groups []int, offset, limit int) ([]*model.OutboundResourceGroup, *model.AppError)
	Get(domainId int64, id int64) (*model.OutboundResourceGroup, *model.AppError)
	Update(group *model.OutboundResourceGroup) (*model.OutboundResourceGroup, *model.AppError)
	Delete(domainId, id int64) *model.AppError
}

type OutboundResourceInGroupStore interface {
	Create(domainId, resourceId, groupId int64) (*model.OutboundResourceInGroup, *model.AppError)
	GetAllPage(domainId, groupId int64, offset, limit int) ([]*model.OutboundResourceInGroup, *model.AppError)
	Get(domainId, groupId, id int64) (*model.OutboundResourceInGroup, *model.AppError)
	Update(domainId int64, res *model.OutboundResourceInGroup) (*model.OutboundResourceInGroup, *model.AppError)
	Delete(domainId, groupId, id int64) *model.AppError
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

type QueueRoutingStore interface {
	Create(routing *model.QueueRouting) (*model.QueueRouting, *model.AppError)
	GetAllPage(domainId, queueId int64, offset, limit int) ([]*model.QueueRouting, *model.AppError)
	Get(domainId, queueId int64, id int64) (*model.QueueRouting, *model.AppError)
	Update(qr *model.QueueRouting) (*model.QueueRouting, *model.AppError)
	Delete(queueId, id int64) *model.AppError
}

type SupervisorTeamStore interface {
	Create(supervisor *model.SupervisorInTeam) (*model.SupervisorInTeam, *model.AppError)
	GetAllPage(domainId, teamId int64, offset, limit int) ([]*model.SupervisorInTeam, *model.AppError)
	Get(domainId, teamId, id int64) (*model.SupervisorInTeam, *model.AppError)
	Update(supervisor *model.SupervisorInTeam) (*model.SupervisorInTeam, *model.AppError)
	Delete(teamId, id int64) *model.AppError
}

type CommunicationTypeStore interface {
	Create(comm *model.CommunicationType) (*model.CommunicationType, *model.AppError)
	GetAllPage(domainId int64, offset, limit int) ([]*model.CommunicationType, *model.AppError)
	Get(domainId int64, id int64) (*model.CommunicationType, *model.AppError)
	Update(cType *model.CommunicationType) (*model.CommunicationType, *model.AppError)
	Delete(domainId int64, id int64) *model.AppError
}

type MemberStore interface {
	Create(member *model.Member) (*model.Member, *model.AppError)
	BulkCreate(queueId int64, members []*model.Member) ([]int64, *model.AppError)
	GetAllPage(domainId, queueId int64, offset, limit int) ([]*model.Member, *model.AppError)
	Get(domainId, queueId, id int64) (*model.Member, *model.AppError)
	Update(domainId int64, member *model.Member) (*model.Member, *model.AppError)
	Delete(queueId, id int64) *model.AppError
}

type BucketSore interface {
	Create(bucket *model.Bucket) (*model.Bucket, *model.AppError)
	CheckAccess(domainId, id int64, groups []int, access model.PermissionAccess) (bool, *model.AppError)
	GetAllPage(domainId int64, offset, limit int) ([]*model.Bucket, *model.AppError)
	GetAllPageByGroups(domainId int64, groups []int, offset, limit int) ([]*model.Bucket, *model.AppError)
	Get(domainId int64, id int64) (*model.Bucket, *model.AppError)
	Update(bucket *model.Bucket) (*model.Bucket, *model.AppError)
	Delete(domainId int64, id int64) *model.AppError
}

type BucketInQueueStore interface {
	Create(queueBucket *model.QueueBucket) (*model.QueueBucket, *model.AppError)
	Get(domainId, queueId, id int64) (*model.QueueBucket, *model.AppError)
	GetAllPage(domainId, queueId int64, offset, limit int) ([]*model.QueueBucket, *model.AppError)
	Update(domainId int64, queueBucket *model.QueueBucket) (*model.QueueBucket, *model.AppError)
	Delete(queueId, id int64) *model.AppError
}

type ListStore interface {
	Create(list *model.List) (*model.List, *model.AppError)
	CheckAccess(domainId, id int64, groups []int, access model.PermissionAccess) (bool, *model.AppError)
	GetAllPage(domainId int64, offset, limit int) ([]*model.List, *model.AppError)
	GetAllPageByGroups(domainId int64, groups []int, offset, limit int) ([]*model.List, *model.AppError)
	Get(domainId int64, id int64) (*model.List, *model.AppError)
	Update(list *model.List) (*model.List, *model.AppError)
	Delete(domainId, id int64) *model.AppError

	//Communications
	CreateCommunication(comm *model.ListCommunication) (*model.ListCommunication, *model.AppError)
	GetAllPageCommunication(domainId, listId int64, offset, limit int) ([]*model.ListCommunication, *model.AppError)
	GetCommunication(domainId, listId int64, id int64) (*model.ListCommunication, *model.AppError)
	UpdateCommunication(domainId int64, communication *model.ListCommunication) (*model.ListCommunication, *model.AppError)
	DeleteCommunication(domainId, listId, id int64) *model.AppError
}
