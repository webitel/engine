package store

import (
	"github.com/webitel/engine/auth_manager"
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

type Store interface {
	User() UserStore
	Calendar() CalendarStore
	Skill() SkillStore
	AgentTeam() AgentTeamStore
	Agent() AgentStore
	AgentSkill() AgentSkillStore
	ResourceTeam() ResourceTeamStore
	Queue() QueueStore
	QueueResource() QueueResourceStore
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

	RoutingSchema() RoutingSchemaStore
	RoutingInboundCall() RoutingInboundCallStore
	RoutingOutboundCall() RoutingOutboundCallStore
	RoutingVariable() RoutingVariableStore

	Call() CallStore

	EmailProfile() EmailProfileStore
}

type UserStore interface {
	CheckAccess(domainId, id int64, groups []int, access auth_manager.PermissionAccess) (bool, *model.AppError)
	GetCallInfo(userId, domainId int64) (*model.UserCallInfo, *model.AppError)
	DefaultDeviceConfig(userId, domainId int64) (*model.UserDeviceConfig, *model.AppError)
}

type CalendarStore interface {
	Create(calendar *model.Calendar) (*model.Calendar, *model.AppError)
	GetAllPage(domainId int64, search *model.SearchCalendar) ([]*model.Calendar, *model.AppError)
	GetAllPageByGroups(domainId int64, groups []int, search *model.SearchCalendar) ([]*model.Calendar, *model.AppError)
	Get(domainId int64, id int64) (*model.Calendar, *model.AppError)
	Update(calendar *model.Calendar) (*model.Calendar, *model.AppError)
	Delete(domainId, id int64) *model.AppError

	GetTimezoneAllPage(search *model.SearchTimezone) ([]*model.Timezone, *model.AppError)

	CheckAccess(domainId, id int64, groups []int, access auth_manager.PermissionAccess) (bool, *model.AppError)
}

type SkillStore interface {
	Create(skill *model.Skill) (*model.Skill, *model.AppError)
	Get(domainId int64, id int64) (*model.Skill, *model.AppError)
	GetAllPage(domainId int64, search *model.SearchSkill) ([]*model.Skill, *model.AppError)
	Delete(domainId, id int64) *model.AppError
	Update(skill *model.Skill) (*model.Skill, *model.AppError)
}

type AgentTeamStore interface {
	CheckAccess(domainId, id int64, groups []int, access auth_manager.PermissionAccess) (bool, *model.AppError)

	Create(team *model.AgentTeam) (*model.AgentTeam, *model.AppError)
	GetAllPage(domainId int64, search *model.SearchAgentTeam) ([]*model.AgentTeam, *model.AppError)
	GetAllPageByGroups(domainId int64, groups []int, search *model.SearchAgentTeam) ([]*model.AgentTeam, *model.AppError)
	Get(domainId int64, id int64) (*model.AgentTeam, *model.AppError)
	Update(team *model.AgentTeam) (*model.AgentTeam, *model.AppError)
	Delete(domainId, id int64) *model.AppError
}

type AgentStore interface {
	CheckAccess(domainId, id int64, groups []int, access auth_manager.PermissionAccess) (bool, *model.AppError)

	Create(agent *model.Agent) (*model.Agent, *model.AppError)
	GetAllPage(domainId int64, search *model.SearchAgent) ([]*model.Agent, *model.AppError)
	GetAllPageByGroups(domainId int64, groups []int, search *model.SearchAgent) ([]*model.Agent, *model.AppError)
	Get(domainId int64, id int64) (*model.Agent, *model.AppError)
	Update(agent *model.Agent) (*model.Agent, *model.AppError)
	Delete(domainId, id int64) *model.AppError
	SetStatus(domainId, agentId int64, status string, payload interface{}) (bool, *model.AppError)

	GetSession(domainId, userId int64) (*model.AgentSession, *model.AppError)

	/* stats */
	CallStatistics(domainId int64, search *model.SearchAgentCallStatistics) ([]*model.AgentCallStatistics, *model.AppError)

	/* view */
	InTeam(domainId, id int64, search *model.SearchAgentInTeam) ([]*model.AgentInTeam, *model.AppError)
	InQueue(domainId, id int64, search *model.SearchAgentInQueue) ([]*model.AgentInQueue, *model.AppError)
	QueueStatistic(domainId, agentId int64) ([]*model.AgentInQueueStatistic, *model.AppError)
	HistoryState(domainId int64, search *model.SearchAgentState) ([]*model.AgentState, *model.AppError)

	/*Lookups*/
	LookupNotExistsUsers(domainId int64, search *model.SearchAgentUser) ([]*model.AgentUser, *model.AppError)
	LookupNotExistsUsersByGroups(domainId int64, groups []int, search *model.SearchAgentUser) ([]*model.AgentUser, *model.AppError)

	StatusStatistic(domainId int64, search *model.SearchAgentStatusStatistic) ([]*model.AgentStatusStatistics, *model.AppError)
}

type AgentSkillStore interface {
	Create(agent *model.AgentSkill) (*model.AgentSkill, *model.AppError)
	GetById(domainId, agentId, id int64) (*model.AgentSkill, *model.AppError)
	Update(agentSkill *model.AgentSkill) (*model.AgentSkill, *model.AppError)
	GetAllPage(domainId, agentId int64, search *model.SearchAgentSkill) ([]*model.AgentSkill, *model.AppError)
	Delete(agentId, id int64) *model.AppError

	LookupNotExistsAgent(domainId, agentId int64, search *model.SearchAgentSkill) ([]*model.Skill, *model.AppError)
}

type ResourceTeamStore interface {
	Create(in *model.ResourceInTeam) (*model.ResourceInTeam, *model.AppError)
	Get(domainId, teamId int64, id int64) (*model.ResourceInTeam, *model.AppError)
	GetAllPage(domainId, teamId int64, search *model.SearchResourceInTeam) ([]*model.ResourceInTeam, *model.AppError)
	Update(resource *model.ResourceInTeam) (*model.ResourceInTeam, *model.AppError)
	Delete(domainId, teamId int64, id int64) *model.AppError
}

type OutboundResourceStore interface {
	CheckAccess(domainId, id int64, groups []int, access auth_manager.PermissionAccess) (bool, *model.AppError)
	Create(resource *model.OutboundCallResource) (*model.OutboundCallResource, *model.AppError)
	GetAllPage(domainId int64, search *model.SearchOutboundCallResource) ([]*model.OutboundCallResource, *model.AppError)
	GetAllPageByGroups(domainId int64, groups []int, search *model.SearchOutboundCallResource) ([]*model.OutboundCallResource, *model.AppError)
	Get(domainId int64, id int64) (*model.OutboundCallResource, *model.AppError)
	Update(resource *model.OutboundCallResource) (*model.OutboundCallResource, *model.AppError)
	Delete(domainId, id int64) *model.AppError

	SaveDisplay(d *model.ResourceDisplay) (*model.ResourceDisplay, *model.AppError)
	GetDisplayAllPage(domainId, resourceId int64, search *model.SearchResourceDisplay) ([]*model.ResourceDisplay, *model.AppError)
	GetDisplay(domainId, resourceId, id int64) (*model.ResourceDisplay, *model.AppError)
	UpdateDisplay(domainId int64, display *model.ResourceDisplay) (*model.ResourceDisplay, *model.AppError)
	DeleteDisplay(domainId, resourceId, id int64) *model.AppError
}

type OutboundResourceGroupStore interface {
	CheckAccess(domainId, id int64, groups []int, access auth_manager.PermissionAccess) (bool, *model.AppError)
	Create(group *model.OutboundResourceGroup) (*model.OutboundResourceGroup, *model.AppError)
	GetAllPage(domainId int64, search *model.SearchOutboundResourceGroup) ([]*model.OutboundResourceGroup, *model.AppError)
	GetAllPageByGroups(domainId int64, groups []int, search *model.SearchOutboundResourceGroup) ([]*model.OutboundResourceGroup, *model.AppError)
	Get(domainId int64, id int64) (*model.OutboundResourceGroup, *model.AppError)
	Update(group *model.OutboundResourceGroup) (*model.OutboundResourceGroup, *model.AppError)
	Delete(domainId, id int64) *model.AppError
}

type OutboundResourceInGroupStore interface {
	Create(domainId, resourceId, groupId int64) (*model.OutboundResourceInGroup, *model.AppError)
	GetAllPage(domainId, groupId int64, search *model.SearchOutboundResourceInGroup) ([]*model.OutboundResourceInGroup, *model.AppError)
	Get(domainId, groupId, id int64) (*model.OutboundResourceInGroup, *model.AppError)
	Update(domainId int64, res *model.OutboundResourceInGroup) (*model.OutboundResourceInGroup, *model.AppError)
	Delete(domainId, groupId, id int64) *model.AppError
}

type RoutingSchemaStore interface {
	Create(scheme *model.RoutingSchema) (*model.RoutingSchema, *model.AppError)
	GetAllPage(domainId int64, search *model.SearchRoutingSchema) ([]*model.RoutingSchema, *model.AppError)
	Get(domainId int64, id int64) (*model.RoutingSchema, *model.AppError)
	Update(scheme *model.RoutingSchema) (*model.RoutingSchema, *model.AppError)
	Delete(domainId, id int64) *model.AppError
}

//FIXME
type RoutingInboundCallStore interface {
	Create(routing *model.RoutingInboundCall) (*model.RoutingInboundCall, *model.AppError)
	GetAllPage(domainId int64, offset, limit int) ([]*model.RoutingInboundCall, *model.AppError)
	Get(domainId, id int64) (*model.RoutingInboundCall, *model.AppError)
	Update(routing *model.RoutingInboundCall) (*model.RoutingInboundCall, *model.AppError)
	Delete(domainId, id int64) *model.AppError
}

type RoutingOutboundCallStore interface {
	Create(routing *model.RoutingOutboundCall) (*model.RoutingOutboundCall, *model.AppError)
	GetAllPage(domainId int64, search *model.SearchRoutingOutboundCall) ([]*model.RoutingOutboundCall, *model.AppError)
	Get(domainId, id int64) (*model.RoutingOutboundCall, *model.AppError)
	Update(routing *model.RoutingOutboundCall) (*model.RoutingOutboundCall, *model.AppError)
	Delete(domainId, id int64) *model.AppError

	ChangePosition(domainId, fromId, toId int64) *model.AppError
}

type RoutingVariableStore interface {
	Create(variable *model.RoutingVariable) (*model.RoutingVariable, *model.AppError)
	GetAllPage(domainId int64, offset, limit int) ([]*model.RoutingVariable, *model.AppError) //FIXME
	Get(domainId int64, id int64) (*model.RoutingVariable, *model.AppError)
	Update(variable *model.RoutingVariable) (*model.RoutingVariable, *model.AppError)
	Delete(domainId, id int64) *model.AppError
}

type QueueStore interface {
	CheckAccess(domainId, id int64, groups []int, access auth_manager.PermissionAccess) (bool, *model.AppError)
	Create(queue *model.Queue) (*model.Queue, *model.AppError)
	GetAllPage(domainId int64, search *model.SearchQueue) ([]*model.Queue, *model.AppError)
	GetAllPageByGroups(domainId int64, groups []int, search *model.SearchQueue) ([]*model.Queue, *model.AppError)
	Get(domainId int64, id int64) (*model.Queue, *model.AppError)
	Update(queue *model.Queue) (*model.Queue, *model.AppError)
	Delete(domainId, id int64) *model.AppError

	QueueReportGeneral(domainId int64, search *model.SearchQueueReportGeneral) ([]*model.QueueReportGeneral, *model.AppError)
}

type QueueResourceStore interface {
	Create(queueResource *model.QueueResourceGroup) (*model.QueueResourceGroup, *model.AppError)
	Get(domainId, queueId, id int64) (*model.QueueResourceGroup, *model.AppError)
	GetAllPage(domainId, queueId int64, search *model.SearchQueueResourceGroup) ([]*model.QueueResourceGroup, *model.AppError)
	Update(domainId int64, queueResourceGroup *model.QueueResourceGroup) (*model.QueueResourceGroup, *model.AppError)
	Delete(queueId, id int64) *model.AppError
}

//FIXME delete
type QueueRoutingStore interface {
	Create(routing *model.QueueRouting) (*model.QueueRouting, *model.AppError)
	GetAllPage(domainId, queueId int64, offset, limit int) ([]*model.QueueRouting, *model.AppError)
	Get(domainId, queueId int64, id int64) (*model.QueueRouting, *model.AppError)
	Update(qr *model.QueueRouting) (*model.QueueRouting, *model.AppError)
	Delete(queueId, id int64) *model.AppError
}

type SupervisorTeamStore interface {
	Create(supervisor *model.SupervisorInTeam) (*model.SupervisorInTeam, *model.AppError)
	GetAllPage(domainId, teamId int64, search *model.SearchSupervisorInTeam) ([]*model.SupervisorInTeam, *model.AppError)
	Get(domainId, teamId, id int64) (*model.SupervisorInTeam, *model.AppError)
	Update(supervisor *model.SupervisorInTeam) (*model.SupervisorInTeam, *model.AppError)
	Delete(teamId, id int64) *model.AppError
}

type CommunicationTypeStore interface {
	Create(comm *model.CommunicationType) (*model.CommunicationType, *model.AppError)
	GetAllPage(domainId int64, search *model.SearchCommunicationType) ([]*model.CommunicationType, *model.AppError)
	Get(domainId int64, id int64) (*model.CommunicationType, *model.AppError)
	Update(cType *model.CommunicationType) (*model.CommunicationType, *model.AppError)
	Delete(domainId int64, id int64) *model.AppError
}

type MemberStore interface {
	Create(domainId int64, member *model.Member) (*model.Member, *model.AppError)
	BulkCreate(domainId, queueId int64, members []*model.Member) ([]int64, *model.AppError)
	GetAllPage(domainId, queueId int64, search *model.SearchMemberRequest) ([]*model.Member, *model.AppError)
	SearchMembers(domainId int64, search *model.SearchMemberRequest) ([]*model.Member, *model.AppError)
	Get(domainId, queueId, id int64) (*model.Member, *model.AppError)
	Update(domainId int64, member *model.Member) (*model.Member, *model.AppError)
	Delete(queueId, id int64) *model.AppError
	MultiDelete(queueId int64, ids []int64) ([]*model.Member, *model.AppError)

	// Move to new store
	AttemptsList(memberId int64) ([]*model.MemberAttempt, *model.AppError) //FIXME
	SearchAttempts(domainId int64, search *model.SearchAttempts) ([]*model.Attempt, *model.AppError)
	SearchAttemptsHistory(domainId int64, search *model.SearchAttempts) ([]*model.AttemptHistory, *model.AppError)
	ListOfflineQueueForAgent(domainId int64, search *model.SearchOfflineQueueMembers) ([]*model.OfflineMember, *model.AppError)
}

type BucketSore interface {
	Create(bucket *model.Bucket) (*model.Bucket, *model.AppError)
	CheckAccess(domainId, id int64, groups []int, access auth_manager.PermissionAccess) (bool, *model.AppError)
	GetAllPage(domainId int64, search *model.SearchBucket) ([]*model.Bucket, *model.AppError)
	GetAllPageByGroups(domainId int64, groups []int, search *model.SearchBucket) ([]*model.Bucket, *model.AppError)
	Get(domainId int64, id int64) (*model.Bucket, *model.AppError)
	Update(bucket *model.Bucket) (*model.Bucket, *model.AppError)
	Delete(domainId int64, id int64) *model.AppError
}

type BucketInQueueStore interface {
	Create(queueBucket *model.QueueBucket) (*model.QueueBucket, *model.AppError)
	Get(domainId, queueId, id int64) (*model.QueueBucket, *model.AppError)
	GetAllPage(domainId, queueId int64, search *model.SearchQueueBucket) ([]*model.QueueBucket, *model.AppError)
	Update(domainId int64, queueBucket *model.QueueBucket) (*model.QueueBucket, *model.AppError)
	Delete(queueId, id int64) *model.AppError
}

type ListStore interface {
	Create(list *model.List) (*model.List, *model.AppError)
	CheckAccess(domainId, id int64, groups []int, access auth_manager.PermissionAccess) (bool, *model.AppError)
	GetAllPage(domainId int64, search *model.SearchList) ([]*model.List, *model.AppError)
	GetAllPageByGroups(domainId int64, groups []int, search *model.SearchList) ([]*model.List, *model.AppError)
	Get(domainId int64, id int64) (*model.List, *model.AppError)
	Update(list *model.List) (*model.List, *model.AppError)
	Delete(domainId, id int64) *model.AppError

	//Communications
	CreateCommunication(comm *model.ListCommunication) (*model.ListCommunication, *model.AppError)
	GetAllPageCommunication(domainId, listId int64, search *model.SearchListCommunication) ([]*model.ListCommunication, *model.AppError)
	GetCommunication(domainId, listId int64, id int64) (*model.ListCommunication, *model.AppError)
	UpdateCommunication(domainId int64, communication *model.ListCommunication) (*model.ListCommunication, *model.AppError)
	DeleteCommunication(domainId, listId, id int64) *model.AppError
}

type CallStore interface {
	GetHistory(domainId int64, search *model.SearchHistoryCall) ([]*model.HistoryCall, *model.AppError)
	GetActive(domainId int64, search *model.SearchCall) ([]*model.Call, *model.AppError)
	Get(domainId int64, id string) (*model.Call, *model.AppError)
	GetInstance(domainId int64, id string) (*model.CallInstance, *model.AppError)
	BridgeInfo(domainId int64, fromId, toId string) (*model.BridgeCall, *model.AppError)
	BridgedId(id string) (string, *model.AppError)
	LastFile(domainId int64, id string) (int64, *model.AppError)
}

type EmailProfileStore interface {
	Create(p *model.EmailProfile) (*model.EmailProfile, *model.AppError)
	GetAllPage(domainId int64, search *model.SearchEmailProfile) ([]*model.EmailProfile, *model.AppError)
	Get(domainId int64, id int) (*model.EmailProfile, *model.AppError)
	Update(p *model.EmailProfile) (*model.EmailProfile, *model.AppError)
	Delete(domainId int64, id int) *model.AppError
}

type NotificationStore interface {
}
