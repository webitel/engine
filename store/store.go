package store

import (
	"context"
	"github.com/Masterminds/squirrel"
	"golang.org/x/oauth2"

	"github.com/webitel/engine/auth_manager"
	"github.com/webitel/engine/model"
)

type StoreResult struct {
	Data interface{}
	Err  model.AppError
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
	TeamHook() TeamHookStore
	TeamTrigger() TeamTriggerStore
	Agent() AgentStore
	AgentSkill() AgentSkillStore
	Queue() QueueStore
	QueueResource() QueueResourceStore
	QueueSkill() QueueSkillStore
	QueueHook() QueueHookStore
	Bucket() BucketStore
	BucketInQueue() BucketInQueueStore
	OutboundResource() OutboundResourceStore
	OutboundResourceGroup() OutboundResourceGroupStore
	OutboundResourceInGroup() OutboundResourceInGroupStore
	CommunicationType() CommunicationTypeStore
	List() ListStore

	Member() MemberStore

	RoutingSchema() RoutingSchemaStore
	RoutingOutboundCall() RoutingOutboundCallStore
	RoutingVariable() RoutingVariableStore

	Call() CallStore

	EmailProfile() EmailProfileStore
	Chat() ChatStore

	Region() RegionStore

	PauseCause() PauseCauseStore
	Notification() NotificationStore

	ChatPlan() ChatPlanStore
	Trigger() TriggerStore

	AuditForm() AuditFormStore
	AuditRate() AuditRateStore
	PresetQuery() PresetQueryStore
	SystemSettings() SystemSettingsStore
	WebHook() WebHookStore
	SchemeVersion() SchemeVersionsStore
	SchemeVariable() SchemeVariablesStore
}

// todo deprecated
type ChatStore interface {
	OpenedConversations(ctx context.Context, domainId, userId int64) ([]*model.Conversation, model.AppError)
	ValidDomain(ctx context.Context, domainId int64, profileId int64) model.AppError
}

type UserStore interface {
	CheckAccess(ctx context.Context, domainId, id int64, groups []int, access auth_manager.PermissionAccess) (bool, model.AppError)
	GetCallInfo(ctx context.Context, userId, domainId int64) (*model.UserCallInfo, model.AppError)
	GetCallInfoEndpoint(ctx context.Context, domainId int64, e *model.EndpointRequest, isOnline bool) (*model.UserCallInfo, model.AppError)
	DefaultWebRTCDeviceConfig(ctx context.Context, userId, domainId int64) (*model.UserDeviceConfig, model.AppError)
	DefaultSipDeviceConfig(ctx context.Context, userId, domainId int64) (*model.UserSipDeviceConfig, model.AppError)
}

type CalendarStore interface {
	Create(ctx context.Context, calendar *model.Calendar) (*model.Calendar, model.AppError)
	GetAllPage(ctx context.Context, domainId int64, search *model.SearchCalendar) ([]*model.Calendar, model.AppError)
	GetAllPageByGroups(ctx context.Context, domainId int64, groups []int, search *model.SearchCalendar) ([]*model.Calendar, model.AppError)
	Get(ctx context.Context, domainId int64, id int64) (*model.Calendar, model.AppError)
	Update(ctx context.Context, calendar *model.Calendar) (*model.Calendar, model.AppError)
	Delete(ctx context.Context, domainId, id int64) model.AppError

	GetTimezoneAllPage(ctx context.Context, search *model.SearchTimezone) ([]*model.Timezone, model.AppError)

	CheckAccess(ctx context.Context, domainId, id int64, groups []int, access auth_manager.PermissionAccess) (bool, model.AppError)
}

type SkillStore interface {
	Create(ctx context.Context, skill *model.Skill) (*model.Skill, model.AppError)
	Get(ctx context.Context, domainId int64, id int64) (*model.Skill, model.AppError)
	GetAllPage(ctx context.Context, domainId int64, search *model.SearchSkill) ([]*model.Skill, model.AppError)
	Delete(ctx context.Context, domainId, id int64) model.AppError
	Update(ctx context.Context, skill *model.Skill) (*model.Skill, model.AppError)
}

type AgentTeamStore interface {
	CheckAccess(ctx context.Context, domainId, id int64, groups []int, access auth_manager.PermissionAccess) (bool, model.AppError)

	Create(ctx context.Context, team *model.AgentTeam) (*model.AgentTeam, model.AppError)
	GetAllPage(ctx context.Context, domainId int64, search *model.SearchAgentTeam) ([]*model.AgentTeam, model.AppError)
	GetAllPageByGroups(ctx context.Context, domainId int64, groups []int, search *model.SearchAgentTeam) ([]*model.AgentTeam, model.AppError)
	Get(ctx context.Context, domainId int64, id int64) (*model.AgentTeam, model.AppError)
	Update(ctx context.Context, domainId int64, team *model.AgentTeam) (*model.AgentTeam, model.AppError)
	Delete(ctx context.Context, domainId, id int64) model.AppError
}

type AgentStore interface {
	AgentCC(ctx context.Context, domainId int64, userId int64) (*model.AgentCC, model.AppError)
	CheckAccess(ctx context.Context, domainId, id int64, groups []int, access auth_manager.PermissionAccess) (bool, model.AppError)
	AccessAgents(ctx context.Context, domainId int64, agentIds []int64, groups []int, access auth_manager.PermissionAccess) ([]int64, model.AppError)

	Create(ctx context.Context, agent *model.Agent) (*model.Agent, model.AppError)
	GetAllPage(ctx context.Context, domainId int64, search *model.SearchAgent) ([]*model.Agent, model.AppError)
	GetAllPageByGroups(ctx context.Context, domainId int64, groups []int, search *model.SearchAgent) ([]*model.Agent, model.AppError)
	GetActiveTask(ctx context.Context, domainId, id int64) ([]*model.CCTask, model.AppError)
	Get(ctx context.Context, domainId int64, id int64) (*model.Agent, model.AppError)
	Update(ctx context.Context, agent *model.Agent) (*model.Agent, model.AppError)
	Delete(ctx context.Context, domainId, id int64) model.AppError
	SetStatus(ctx context.Context, domainId, agentId int64, status string, payload interface{}) (bool, model.AppError)

	GetSession(ctx context.Context, domainId, userId int64) (*model.AgentSession, model.AppError)

	PauseCause(ctx context.Context, domainId int64, fromUserId, toAgentId int64, allowChange bool) ([]*model.AgentPauseCause, model.AppError)

	/* stats */
	CallStatistics(ctx context.Context, domainId int64, search *model.SearchAgentCallStatistics) ([]*model.AgentCallStatistics, model.AppError)
	TodayStatistics(ctx context.Context, domainId int64, agentId *int64, userId *int64) (*model.AgentStatistics, model.AppError)

	/* view */
	InQueue(ctx context.Context, domainId, id int64, search *model.SearchAgentInQueue) ([]*model.AgentInQueue, model.AppError)
	QueueStatistic(ctx context.Context, domainId, agentId int64) ([]*model.AgentInQueueStatistic, model.AppError)
	HistoryState(ctx context.Context, domainId int64, search *model.SearchAgentState) ([]*model.AgentState, model.AppError)

	/*Lookups*/
	LookupNotExistsUsers(ctx context.Context, domainId int64, search *model.SearchAgentUser) ([]*model.AgentUser, model.AppError)
	LookupNotExistsUsersByGroups(ctx context.Context, domainId int64, groups []int, search *model.SearchAgentUser) ([]*model.AgentUser, model.AppError)

	StatusStatistic(ctx context.Context, domainId int64, supervisorUserId int64, groups []int, access auth_manager.PermissionAccess, search *model.SearchAgentStatusStatistic) ([]*model.AgentStatusStatistics, model.AppError)
	SupervisorAgentItem(ctx context.Context, domainId int64, agentId int64, t *model.FilterBetween) (*model.SupervisorAgentItem, model.AppError)
	DistributeInfoByUserId(ctx context.Context, domainId, userId int64, channel string) (*model.DistributeAgentInfo, model.AppError)

	UsersStatus(ctx context.Context, domainId int64, search *model.SearchUserStatus) ([]*model.UserStatus, model.AppError)
	UsersStatusByGroup(ctx context.Context, domainId int64, groups []int, search *model.SearchUserStatus) ([]*model.UserStatus, model.AppError)
}

type TeamHookStore interface {
	Create(ctx context.Context, domainId int64, teamId int64, in *model.TeamHook) (*model.TeamHook, model.AppError)
	Get(ctx context.Context, domainId int64, teamId int64, id uint32) (*model.TeamHook, model.AppError)
	GetAllPage(ctx context.Context, domainId int64, teamId int64, search *model.SearchTeamHook) ([]*model.TeamHook, model.AppError)
	Update(ctx context.Context, domainId int64, teamId int64, qh *model.TeamHook) (*model.TeamHook, model.AppError)
	Delete(ctx context.Context, domainId int64, teamId int64, id uint32) model.AppError
}

type TeamTriggerStore interface {
	Create(ctx context.Context, domainId int64, teamId int64, in *model.TeamTrigger) (*model.TeamTrigger, model.AppError)
	Get(ctx context.Context, domainId int64, teamId int64, id uint32) (*model.TeamTrigger, model.AppError)
	GetAllPage(ctx context.Context, domainId int64, teamId int64, search *model.SearchTeamTrigger) ([]*model.TeamTrigger, model.AppError)
	Update(ctx context.Context, domainId int64, teamId int64, qt *model.TeamTrigger) (*model.TeamTrigger, model.AppError)
	Delete(ctx context.Context, domainId int64, teamId int64, id uint32) model.AppError
}

type AgentSkillStore interface {
	Create(ctx context.Context, agent *model.AgentSkill) (*model.AgentSkill, model.AppError)
	BulkCreate(ctx context.Context, domainId, agentId int64, skills []*model.AgentSkill) ([]int64, model.AppError)
	GetById(ctx context.Context, domainId, agentId, id int64) (*model.AgentSkill, model.AppError)
	Update(ctx context.Context, agentSkill *model.AgentSkill) (*model.AgentSkill, model.AppError)
	UpdateMany(ctx context.Context, domainId int64, search model.SearchAgentSkill, path model.AgentSkillPatch) ([]*model.AgentSkill, model.AppError)
	GetAllPage(ctx context.Context, domainId int64, search *model.SearchAgentSkillList) ([]*model.AgentSkill, model.AppError)
	DeleteById(ctx context.Context, agentId, id int64) model.AppError
	Delete(ctx context.Context, domainId int64, search model.SearchAgentSkill) ([]*model.AgentSkill, model.AppError)

	LookupNotExistsAgent(ctx context.Context, domainId, agentId int64, search *model.SearchAgentSkillList) ([]*model.Skill, model.AppError)

	CreateMany(ctx context.Context, domainId int64, in *model.AgentsSkills) ([]*model.AgentSkill, model.AppError)
	HasDisabledSkill(ctx context.Context, domainId int64, skillId int64) (bool, model.AppError)
}

type OutboundResourceStore interface {
	CheckAccess(ctx context.Context, domainId, id int64, groups []int, access auth_manager.PermissionAccess) (bool, model.AppError)
	Create(ctx context.Context, resource *model.OutboundCallResource) (*model.OutboundCallResource, model.AppError)
	GetAllPage(ctx context.Context, domainId int64, search *model.SearchOutboundCallResource) ([]*model.OutboundCallResource, model.AppError)
	GetAllPageByGroups(ctx context.Context, domainId int64, groups []int, search *model.SearchOutboundCallResource) ([]*model.OutboundCallResource, model.AppError)
	Get(ctx context.Context, domainId int64, id int64) (*model.OutboundCallResource, model.AppError)
	Update(ctx context.Context, resource *model.OutboundCallResource) (*model.OutboundCallResource, model.AppError)
	Delete(ctx context.Context, domainId, id int64) model.AppError

	SaveDisplay(ctx context.Context, d *model.ResourceDisplay) (*model.ResourceDisplay, model.AppError)
	SaveDisplays(ctx context.Context, resourceId int64, d []*model.ResourceDisplay) ([]*model.ResourceDisplay, model.AppError)
	GetDisplayAllPage(ctx context.Context, domainId, resourceId int64, search *model.SearchResourceDisplay) ([]*model.ResourceDisplay, model.AppError)
	GetDisplay(ctx context.Context, domainId, resourceId, id int64) (*model.ResourceDisplay, model.AppError)
	UpdateDisplay(ctx context.Context, domainId int64, display *model.ResourceDisplay) (*model.ResourceDisplay, model.AppError)
	DeleteDisplay(ctx context.Context, domainId, resourceId, id int64) model.AppError
	DeleteDisplays(ctx context.Context, resourceId int64, ids []int64) model.AppError
}

type OutboundResourceGroupStore interface {
	CheckAccess(ctx context.Context, domainId, id int64, groups []int, access auth_manager.PermissionAccess) (bool, model.AppError)
	Create(ctx context.Context, group *model.OutboundResourceGroup) (*model.OutboundResourceGroup, model.AppError)
	GetAllPage(ctx context.Context, domainId int64, search *model.SearchOutboundResourceGroup) ([]*model.OutboundResourceGroup, model.AppError)
	GetAllPageByGroups(ctx context.Context, domainId int64, groups []int, search *model.SearchOutboundResourceGroup) ([]*model.OutboundResourceGroup, model.AppError)
	Get(ctx context.Context, domainId int64, id int64) (*model.OutboundResourceGroup, model.AppError)
	Update(ctx context.Context, group *model.OutboundResourceGroup) (*model.OutboundResourceGroup, model.AppError)
	Delete(ctx context.Context, domainId, id int64) model.AppError
}

type OutboundResourceInGroupStore interface {
	Create(ctx context.Context, domainId int64, res *model.OutboundResourceInGroup) (*model.OutboundResourceInGroup, model.AppError)
	GetAllPage(ctx context.Context, domainId, groupId int64, search *model.SearchOutboundResourceInGroup) ([]*model.OutboundResourceInGroup, model.AppError)
	Get(ctx context.Context, domainId, groupId, id int64) (*model.OutboundResourceInGroup, model.AppError)
	Update(ctx context.Context, domainId int64, res *model.OutboundResourceInGroup) (*model.OutboundResourceInGroup, model.AppError)
	Delete(ctx context.Context, domainId, groupId, id int64) model.AppError
}

type RoutingSchemaStore interface {
	Create(ctx context.Context, scheme *model.RoutingSchema) (*model.RoutingSchema, model.AppError)
	GetAllPage(ctx context.Context, domainId int64, search *model.SearchRoutingSchema) ([]*model.RoutingSchema, model.AppError)
	Get(ctx context.Context, domainId int64, id int64) (*model.RoutingSchema, model.AppError)
	Update(ctx context.Context, scheme *model.RoutingSchema) (*model.RoutingSchema, model.AppError)
	Delete(ctx context.Context, domainId, id int64) model.AppError

	ListTags(ctx context.Context, domainId int64, search *model.SearchRoutingSchemaTag) ([]*model.RoutingSchemaTag, model.AppError)
}

type RoutingOutboundCallStore interface {
	Create(ctx context.Context, routing *model.RoutingOutboundCall) (*model.RoutingOutboundCall, model.AppError)
	GetAllPage(ctx context.Context, domainId int64, search *model.SearchRoutingOutboundCall) ([]*model.RoutingOutboundCall, model.AppError)
	Get(ctx context.Context, domainId, id int64) (*model.RoutingOutboundCall, model.AppError)
	Update(ctx context.Context, routing *model.RoutingOutboundCall) (*model.RoutingOutboundCall, model.AppError)
	Delete(ctx context.Context, domainId, id int64) model.AppError

	ChangePosition(ctx context.Context, domainId, fromId, toId int64) model.AppError
}

type RoutingVariableStore interface {
	Create(ctx context.Context, variable *model.RoutingVariable) (*model.RoutingVariable, model.AppError)
	GetAllPage(ctx context.Context, domainId int64, offset, limit int) ([]*model.RoutingVariable, model.AppError) //FIXME
	Get(ctx context.Context, domainId int64, id int64) (*model.RoutingVariable, model.AppError)
	Update(ctx context.Context, variable *model.RoutingVariable) (*model.RoutingVariable, model.AppError)
	Delete(ctx context.Context, domainId, id int64) model.AppError
}

type QueueStore interface {
	CheckAccess(ctx context.Context, domainId, id int64, groups []int, access auth_manager.PermissionAccess) (bool, model.AppError)
	Create(ctx context.Context, queue *model.Queue) (*model.Queue, model.AppError)
	GetAllPage(ctx context.Context, domainId int64, search *model.SearchQueue) ([]*model.Queue, model.AppError)
	GetAllPageByGroups(ctx context.Context, domainId int64, groups []int, search *model.SearchQueue) ([]*model.Queue, model.AppError)
	Get(ctx context.Context, domainId int64, id int64) (*model.Queue, model.AppError)
	Update(ctx context.Context, queue *model.Queue) (*model.Queue, model.AppError)
	Delete(ctx context.Context, domainId, id int64) model.AppError

	QueueReportGeneral(ctx context.Context, domainId int64, supervisorId int64, groups []int, access auth_manager.PermissionAccess, search *model.SearchQueueReportGeneral) (*model.QueueReportGeneralAgg, model.AppError)
	ListTags(ctx context.Context, domainId int64, search *model.ListRequest) ([]*model.Tag, model.AppError)
}

type QueueResourceStore interface {
	Create(ctx context.Context, queueResource *model.QueueResourceGroup) (*model.QueueResourceGroup, model.AppError)
	Get(ctx context.Context, domainId, queueId, id int64) (*model.QueueResourceGroup, model.AppError)
	GetAllPage(ctx context.Context, domainId, queueId int64, search *model.SearchQueueResourceGroup) ([]*model.QueueResourceGroup, model.AppError)
	Update(ctx context.Context, domainId int64, queueResourceGroup *model.QueueResourceGroup) (*model.QueueResourceGroup, model.AppError)
	Delete(ctx context.Context, queueId, id int64) model.AppError
}

type QueueSkillStore interface {
	Create(ctx context.Context, domainId int64, in *model.QueueSkill) (*model.QueueSkill, model.AppError)
	Get(ctx context.Context, domainId int64, queueId, id uint32) (*model.QueueSkill, model.AppError)
	GetAllPage(ctx context.Context, domainId int64, search *model.SearchQueueSkill) ([]*model.QueueSkill, model.AppError)
	Update(ctx context.Context, domainId int64, skill *model.QueueSkill) (*model.QueueSkill, model.AppError)
	Delete(ctx context.Context, domainId int64, queueId, id uint32) model.AppError
}

type QueueHookStore interface {
	Create(ctx context.Context, domainId int64, queueId uint32, in *model.QueueHook) (*model.QueueHook, model.AppError)
	Get(ctx context.Context, domainId int64, queueId, id uint32) (*model.QueueHook, model.AppError)
	GetAllPage(ctx context.Context, domainId int64, queueId uint32, search *model.SearchQueueHook) ([]*model.QueueHook, model.AppError)
	Update(ctx context.Context, domainId int64, queueId uint32, qh *model.QueueHook) (*model.QueueHook, model.AppError)
	Delete(ctx context.Context, domainId int64, queueId, id uint32) model.AppError
}

type CommunicationTypeStore interface {
	Create(ctx context.Context, domainId int64, comm *model.CommunicationType) (*model.CommunicationType, model.AppError)
	GetAllPage(ctx context.Context, domainId int64, search *model.SearchCommunicationType) ([]*model.CommunicationType, model.AppError)
	Get(ctx context.Context, domainId int64, id int64) (*model.CommunicationType, model.AppError)
	Update(ctx context.Context, domainId int64, cType *model.CommunicationType) (*model.CommunicationType, model.AppError)
	Delete(ctx context.Context, domainId int64, id int64) model.AppError
}

type MemberStore interface {
	Create(ctx context.Context, domainId int64, member *model.Member) (*model.Member, model.AppError)
	BulkCreate(ctx context.Context, domainId, queueId int64, fileName string, members []*model.Member) ([]int64, model.AppError)
	SearchMembers(ctx context.Context, domainId int64, search *model.SearchMemberRequest) ([]*model.Member, model.AppError)
	Get(ctx context.Context, domainId, queueId, id int64) (*model.Member, model.AppError)
	Update(ctx context.Context, domainId int64, member *model.Member) (*model.Member, model.AppError)
	Delete(ctx context.Context, queueId, id int64) model.AppError
	MultiDelete(ctx context.Context, domainId int64, del *model.MultiDeleteMembers, withoutMembers bool) ([]*model.Member, model.AppError)
	ResetMembers(ctx context.Context, domainId int64, req *model.ResetMembers) (int64, model.AppError)

	// Move to new store
	AttemptsList(ctx context.Context, memberId int64) ([]*model.MemberAttempt, model.AppError) //FIXME
	SearchAttempts(ctx context.Context, domainId int64, search *model.SearchAttempts) ([]*model.Attempt, model.AppError)
	SearchAttemptsHistory(ctx context.Context, domainId int64, search *model.SearchAttempts) ([]*model.AttemptHistory, model.AppError)
	ListOfflineQueueForAgent(ctx context.Context, domainId int64, search *model.SearchOfflineQueueMembers) ([]*model.OfflineMember, model.AppError)

	// Appointments
	GetAppointmentWidget(ctx context.Context, uri string) (*model.AppointmentWidget, model.AppError)
	GetAppointment(ctx context.Context, memberId int64) (*model.Appointment, model.AppError)
	CreateAppointment(ctx context.Context, profile *model.AppointmentProfile, app *model.Appointment) (*model.Appointment, model.AppError)
	CancelAppointment(ctx context.Context, memberId int64, reason string) model.AppError
}

type BucketStore interface {
	Create(ctx context.Context, bucket *model.Bucket) (*model.Bucket, model.AppError)
	CheckAccess(ctx context.Context, domainId, id int64, groups []int, access auth_manager.PermissionAccess) (bool, model.AppError)
	GetAllPage(ctx context.Context, domainId int64, search *model.SearchBucket) ([]*model.Bucket, model.AppError)
	Get(ctx context.Context, domainId int64, id int64) (*model.Bucket, model.AppError)
	Update(ctx context.Context, bucket *model.Bucket) (*model.Bucket, model.AppError)
	Delete(ctx context.Context, domainId int64, id int64) model.AppError
}

type BucketInQueueStore interface {
	Create(ctx context.Context, queueBucket *model.QueueBucket) (*model.QueueBucket, model.AppError)
	Get(ctx context.Context, domainId, queueId, id int64) (*model.QueueBucket, model.AppError)
	GetAllPage(ctx context.Context, domainId, queueId int64, search *model.SearchQueueBucket) ([]*model.QueueBucket, model.AppError)
	Update(ctx context.Context, domainId int64, queueBucket *model.QueueBucket) (*model.QueueBucket, model.AppError)
	Delete(ctx context.Context, queueId, id int64) model.AppError
}

type ListStore interface {
	Create(ctx context.Context, list *model.List) (*model.List, model.AppError)
	CheckAccess(ctx context.Context, domainId, id int64, groups []int, access auth_manager.PermissionAccess) (bool, model.AppError)
	GetAllPage(ctx context.Context, domainId int64, search *model.SearchList) ([]*model.List, model.AppError)
	GetAllPageByGroups(ctx context.Context, domainId int64, groups []int, search *model.SearchList) ([]*model.List, model.AppError)
	Get(ctx context.Context, domainId int64, id int64) (*model.List, model.AppError)
	Update(ctx context.Context, list *model.List) (*model.List, model.AppError)
	Delete(ctx context.Context, domainId, id int64) model.AppError

	//Communications
	CreateCommunication(ctx context.Context, comm *model.ListCommunication) (*model.ListCommunication, model.AppError)
	GetAllPageCommunication(ctx context.Context, domainId, listId int64, search *model.SearchListCommunication) ([]*model.ListCommunication, model.AppError)
	GetCommunication(ctx context.Context, domainId, listId int64, id int64) (*model.ListCommunication, model.AppError)
	UpdateCommunication(ctx context.Context, domainId int64, communication *model.ListCommunication) (*model.ListCommunication, model.AppError)
	DeleteCommunication(ctx context.Context, domainId, listId, id int64) model.AppError
}

type CallStore interface {
	GetHistory(ctx context.Context, domainId int64, search *model.SearchHistoryCall) ([]*model.HistoryCall, model.AppError)
	GetHistoryByGroups(ctx context.Context, domainId int64, userSupervisorId int64, groups []int, search *model.SearchHistoryCall) ([]*model.HistoryCall, model.AppError)
	Aggregate(ctx context.Context, domainId int64, aggs *model.CallAggregate) ([]*model.AggregateResult, model.AppError)
	GetActive(ctx context.Context, domainId int64, search *model.SearchCall) ([]*model.Call, model.AppError)
	GetActiveByGroups(ctx context.Context, domainId int64, userSupervisorId int64, groups []int, search *model.SearchCall) ([]*model.Call, model.AppError)
	Get(ctx context.Context, domainId int64, id string) (*model.Call, model.AppError)
	GetInstance(ctx context.Context, domainId int64, id string) (*model.CallInstance, model.AppError)
	BridgeInfo(ctx context.Context, domainId int64, fromId, toId string) (*model.BridgeCall, model.AppError)
	BridgedId(ctx context.Context, id string) (string, model.AppError)
	LastFile(ctx context.Context, domainId int64, id string) (int64, model.AppError)
	GetUserActiveCall(ctx context.Context, domainId, userId int64) ([]*model.Call, model.AppError)
	SetEmptySeverCall(ctx context.Context, domainId int64, id string) (*model.CallServiceHangup, model.AppError)
	SetVariables(ctx context.Context, domainId int64, id string, vars model.StringMap) (*model.CallDomain, model.AppError)
	GetSipId(ctx context.Context, domainId int64, userId int64, id string) (string, model.AppError)

	CreateAnnotation(ctx context.Context, annotation *model.CallAnnotation) (*model.CallAnnotation, model.AppError)
	GetAnnotation(ctx context.Context, id int64) (*model.CallAnnotation, model.AppError)
	UpdateAnnotation(ctx context.Context, domainId int64, annotation *model.CallAnnotation) (*model.CallAnnotation, model.AppError)
	DeleteAnnotation(ctx context.Context, id int64) model.AppError
	GetEavesdropInfo(ctx context.Context, domainId int64, id string) (*model.EavesdropInfo, model.AppError)

	GetOwnerUserCall(ctx context.Context, id string) (*int64, model.AppError)
	UpdateHistoryCall(ctx context.Context, domainId int64, id string, upd *model.HistoryCallPatch) model.AppError
	SetContactId(ctx context.Context, domainId int64, id string, contactId int64) model.AppError

	FromNumber(ctx context.Context, domainId int64, userId int64, id string) (string, model.AppError)
	SetHideMissedLeg(ctx context.Context, domainId int64, userId int64, id string) model.AppError
	SetHideMissedAllParent(ctx context.Context, domainId int64, userId int64, id string) model.AppError
	FromNumberWithUserIds(ctx context.Context, domainId int64, userId int64, id string) (model.RedialFrom, model.AppError)
}

type EmailProfileStore interface {
	Create(ctx context.Context, domainId int64, p *model.EmailProfile) (*model.EmailProfile, model.AppError)
	GetAllPage(ctx context.Context, domainId int64, search *model.SearchEmailProfile) ([]*model.EmailProfile, model.AppError)
	Get(ctx context.Context, domainId int64, id int) (*model.EmailProfile, model.AppError)
	Update(ctx context.Context, domainId int64, p *model.EmailProfile) (*model.EmailProfile, model.AppError)
	Delete(ctx context.Context, domainId int64, id int) model.AppError

	SetupOAuth2(ctx context.Context, id int, token *oauth2.Token) model.AppError
	CountEnabledByDomain(ctx context.Context, domainId int64) (int, model.AppError)
}

type RegionStore interface {
	Create(ctx context.Context, domainId int64, region *model.Region) (*model.Region, model.AppError)
	GetAllPage(ctx context.Context, domainId int64, search *model.SearchRegion) ([]*model.Region, model.AppError)
	Get(ctx context.Context, domainId int64, id int64) (*model.Region, model.AppError)
	Update(ctx context.Context, domainId int64, region *model.Region) (*model.Region, model.AppError)
	Delete(ctx context.Context, domainId int64, id int64) model.AppError
}

type PauseCauseStore interface {
	Create(ctx context.Context, domainId int64, cause *model.PauseCause) (*model.PauseCause, model.AppError)
	GetAllPage(ctx context.Context, domainId int64, search *model.SearchPauseCause) ([]*model.PauseCause, model.AppError)
	Get(ctx context.Context, domainId int64, id uint32) (*model.PauseCause, model.AppError)
	Update(ctx context.Context, domainId int64, region *model.PauseCause) (*model.PauseCause, model.AppError)
	Delete(ctx context.Context, domainId int64, id uint32) model.AppError
}

type NotificationStore interface {
	Create(ctx context.Context, notification *model.Notification) (*model.Notification, model.AppError)
	Close(ctx context.Context, id, userId int64) (*model.Notification, model.AppError)
	Accept(ctx context.Context, id, userId int64) (*model.Notification, model.AppError)
}

type ChatPlanStore interface {
	Create(ctx context.Context, domainId int64, plan *model.ChatPlan) (*model.ChatPlan, model.AppError)
	GetAllPage(ctx context.Context, domainId int64, search *model.SearchChatPlan) ([]*model.ChatPlan, model.AppError)
	Get(ctx context.Context, domainId int64, id int32) (*model.ChatPlan, model.AppError)
	Update(ctx context.Context, domainId int64, plan *model.ChatPlan) (*model.ChatPlan, model.AppError)
	Delete(ctx context.Context, domainId int64, id int32) model.AppError
	GetSchemaId(ctx context.Context, domainId int64, id int32) (model.Lookup, model.AppError)
}

type TriggerStore interface {
	CheckAccess(ctx context.Context, domainId int64, id int32, groups []int, access auth_manager.PermissionAccess) (bool, model.AppError)
	Create(ctx context.Context, domainId int64, trigger *model.Trigger) (*model.Trigger, model.AppError)
	GetAllPage(ctx context.Context, domainId int64, search *model.SearchTrigger) ([]*model.Trigger, model.AppError)
	GetAllPageByGroup(ctx context.Context, domainId int64, groups []int, search *model.SearchTrigger) ([]*model.Trigger, model.AppError)
	Get(ctx context.Context, domainId int64, id int32) (*model.Trigger, model.AppError)
	Update(ctx context.Context, domainId int64, trigger *model.Trigger) (*model.Trigger, model.AppError)
	Delete(ctx context.Context, domainId int64, id int32) model.AppError

	CreateJob(ctx context.Context, domainId int64, triggerId int32, vars map[string]string) (*model.TriggerJob, model.AppError)
	GetAllJobs(ctx context.Context, triggerId int32, search *model.SearchTriggerJob) ([]*model.TriggerJob, model.AppError)
}

type AuditFormStore interface {
	CheckAccess(ctx context.Context, domainId int64, id int32, groups []int, access auth_manager.PermissionAccess) (bool, model.AppError)
	Create(ctx context.Context, domainId int64, form *model.AuditForm) (*model.AuditForm, model.AppError)
	GetAllPage(ctx context.Context, domainId int64, search *model.SearchAuditForm) ([]*model.AuditForm, model.AppError)
	GetAllPageByGroup(ctx context.Context, domainId int64, groups []int, search *model.SearchAuditForm) ([]*model.AuditForm, model.AppError)
	Get(ctx context.Context, domainId int64, id int32) (*model.AuditForm, model.AppError)
	Update(ctx context.Context, domainId int64, form *model.AuditForm) (*model.AuditForm, model.AppError)
	Delete(ctx context.Context, domainId int64, id int32) model.AppError
	SetEditable(ctx context.Context, id int32, editable bool) model.AppError
}

type AuditRateStore interface {
	Create(ctx context.Context, domainId int64, rate *model.AuditRate) (*model.AuditRate, model.AppError)
	GetAllPage(ctx context.Context, domainId int64, search *model.SearchAuditRate) ([]*model.AuditRate, model.AppError)
	Get(ctx context.Context, domainId int64, id int64) (*model.AuditRate, model.AppError)
	FormId(ctx context.Context, domainId, id int64) (int32, model.AppError)
}

type PresetQueryStore interface {
	Create(ctx context.Context, domainId, userId int64, preset *model.PresetQuery) (*model.PresetQuery, model.AppError)
	GetAllPage(ctx context.Context, domainId, userId int64, search *model.SearchPresetQuery) ([]*model.PresetQuery, model.AppError)
	Get(ctx context.Context, domainId, userId int64, id int32) (*model.PresetQuery, model.AppError)
	Update(ctx context.Context, domainId, userId int64, preset *model.PresetQuery) (*model.PresetQuery, model.AppError)
	Delete(ctx context.Context, domainId, userId int64, id int32) model.AppError
}

type SystemSettingsStore interface {
	Create(ctx context.Context, domainId int64, setting *model.SystemSetting) (*model.SystemSetting, model.AppError)
	GetAllPage(ctx context.Context, domainId int64, search *model.SearchSystemSetting) ([]*model.SystemSetting, model.AppError)
	Get(ctx context.Context, domainId int64, id int32) (*model.SystemSetting, model.AppError)
	Update(ctx context.Context, domainId int64, setting *model.SystemSetting) (*model.SystemSetting, model.AppError)
	Delete(ctx context.Context, domainId int64, id int32) model.AppError
	ValueByName(ctx context.Context, domainId int64, name string) (model.SysValue, model.AppError)
	Available(ctx context.Context, domainId int64, search *model.ListRequest) ([]string, model.AppError)
}

type WebHookStore interface {
	Create(ctx context.Context, domainId int64, hook *model.WebHook) (*model.WebHook, model.AppError)
	GetAllPage(ctx context.Context, domainId int64, search *model.SearchWebHook) ([]*model.WebHook, model.AppError)
	Get(ctx context.Context, domainId int64, id int32) (*model.WebHook, model.AppError)
	Update(ctx context.Context, domainId int64, hook *model.WebHook) (*model.WebHook, model.AppError)
	Delete(ctx context.Context, domainId int64, id int32) model.AppError
}

type SchemeVersionsStore interface {
	Search(ctx context.Context, searchOpts *model.ListRequest, filters any) ([]*model.SchemeVersion, model.AppError)
}

type SchemeVariablesStore interface {
	Create(ctx context.Context, domainId int64, variable *model.SchemeVariable) (*model.SchemeVariable, model.AppError)
	Search(ctx context.Context, domainId int64, searchOpts *model.ListRequest, filters any) ([]*model.SchemeVariable, model.AppError)
	Get(ctx context.Context, domainId int64, id int32) (*model.SchemeVariable, model.AppError)
	Update(ctx context.Context, domainId int64, variable *model.SchemeVariable) (*model.SchemeVariable, model.AppError)
	Delete(ctx context.Context, domainId int64, id int32) model.AppError
}

// ApplyFiltersToBuilder determines type of {filters} parameter and applies {filters} to the {base} according to the determined type.
// columnAlias is additional parameter applied to every model.Filter existing in {filters} and checks if {model.Filter.Column} has alias in the {columnAlias}
func ApplyFiltersToBuilderBulk(base any, columnAlias map[string]string, filters any) (any, model.AppError) {
	if filters == nil {
		return base, nil
	}
	switch data := filters.(type) {
	case *model.FilterNode:
		switch data.Connection {
		case model.AND:
			result := squirrel.And{}
			for _, bunch := range data.Nodes {
				switch bunchType := bunch.(type) {
				case model.FilterNode:
					lowerResult, err := ApplyFiltersToBuilderBulk(result, columnAlias, bunchType)
					if err != nil {
						return nil, err
					}
					switch newData := lowerResult.(type) {
					case squirrel.And:
						result = append(result, newData)
					}
				case model.Filter:
					result = append(result, applyFilter(&bunchType, columnAlias))
				}
			}

			switch baseType := base.(type) {
			case squirrel.And:
				base = append(baseType, result)
			case squirrel.Or:
				base = append(baseType, result)
			case squirrel.SelectBuilder:
				base = baseType.Where(result)
			}
			return base, nil
		case model.OR:
			result := squirrel.Or{}
			for _, bunch := range data.Nodes {
				switch v := bunch.(type) {
				case model.FilterNode:
					lowerResult, err := ApplyFiltersToBuilderBulk(result, columnAlias, v)
					if err != nil {
						return nil, err
					}
					switch newData := lowerResult.(type) {
					case squirrel.And:
						result = append(result, newData)
					}
				case model.Filter:
					result = append(result, applyFilter(&v, columnAlias))
				}
			}
			switch baseType := base.(type) {
			case squirrel.And:
				base = append(baseType, result)
			case squirrel.Or:
				base = append(baseType, result)
			case squirrel.SelectBuilder:
				base = baseType.Where(result)
			}
			return base, nil
		}
	case *model.Filter:
		switch baseType := base.(type) {
		case squirrel.And:
			base = append(baseType, applyFilter(data, columnAlias))
		case squirrel.Or:
			base = append(baseType, applyFilter(data, columnAlias))
		case squirrel.SelectBuilder:
			base = baseType.Where(applyFilter(data, columnAlias))
		}
	}

	return base, nil
}

// Apply filter performs convertation between model.Filter and squirrel.Sqlizer.
// columnAlias is additional parameter to determine if model.Filter in the Column property has alias of the column and NOT the real DB column name.
func applyFilter(filter *model.Filter, columnsAlias map[string]string) squirrel.Sqlizer {
	columnName := filter.Column
	if columnsAlias != nil {
		if alias, ok := columnsAlias[columnName]; ok {
			columnName = alias
		}
	}
	var result squirrel.Sqlizer
	switch filter.ComparisonType {
	case model.GreaterThan:
		result = squirrel.Gt{columnName: filter.Value}
	case model.GreaterThanOrEqual:
		result = squirrel.GtOrEq{columnName: filter.Value}
	case model.LessThan:
		result = squirrel.Lt{columnName: filter.Value}
	case model.LessThanOrEqual:
		result = squirrel.LtOrEq{columnName: filter.Value}
	case model.NotEqual:
		result = squirrel.NotEq{columnName: filter.Value}
	case model.Like:
		result = squirrel.Like{columnName: filter.Value}
	case model.ILike:
		result = squirrel.ILike{columnName: filter.Value}
	default:
		result = squirrel.Eq{columnName: filter.Value}
	}
	return result
}
