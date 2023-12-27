package sqlstore

import (
	"context"
	"github.com/go-gorp/gorp"
	_ "github.com/lib/pq"
	"github.com/webitel/engine/model"
	"github.com/webitel/engine/store"
)

type SqlStore interface {
	GetMaster() *gorp.DbMap
	GetReplica() *gorp.DbMap
	GetAllConns() []*gorp.DbMap

	ListQuery(ctx context.Context, out interface{}, req model.ListRequest, where string, e Entity, params map[string]interface{}) error
	ListQueryMaster(ctx context.Context, out interface{}, req model.ListRequest, where string, e Entity, params map[string]interface{}) error
	One(ctx context.Context, out interface{}, where string, e Entity, params map[string]interface{}) error

	ListQueryTimeout(ctx context.Context, out interface{}, req model.ListRequest, where string, e Entity, params map[string]interface{}) error
	// todo
	ListQueryFromSchema(ctx context.Context, out interface{}, schema string, req model.ListRequest, where string, e Entity, params map[string]interface{}) error

	User() store.UserStore
	Calendar() store.CalendarStore
	Skill() store.SkillStore
	AgentTeam() store.AgentTeamStore
	Agent() store.AgentStore
	AgentSkill() store.AgentSkillStore
	OutboundResource() store.OutboundResourceStore
	OutboundResourceGroup() store.OutboundResourceGroupStore
	OutboundResourceInGroup() store.OutboundResourceInGroupStore
	Queue() store.QueueStore
	QueueResource() store.QueueResourceStore
	QueueSkill() store.QueueSkillStore
	QueueHook() store.QueueHookStore
	Bucket() store.BucketStore
	BucketInQueue() store.BucketInQueueStore
	CommunicationType() store.CommunicationTypeStore
	List() store.ListStore

	Member() store.MemberStore

	RoutingSchema() store.RoutingSchemaStore
	RoutingOutboundCall() store.RoutingOutboundCallStore
	RoutingVariable() store.RoutingVariableStore

	Call() store.CallStore
	EmailProfile() store.EmailProfileStore
	Chat() store.ChatStore
	ChatPlan() store.ChatPlanStore

	Region() store.RegionStore
	PauseCause() store.PauseCauseStore
	Notification() store.NotificationStore
	Trigger() store.TriggerStore
	AuditForm() store.AuditFormStore
	AuditRate() store.AuditRateStore
	PresetQuery() store.PresetQueryStore
	SystemSettings() store.SystemSettingsStore
	//SchemeVersion() store.SystemSettingsStore
}
