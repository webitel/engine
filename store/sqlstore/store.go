package sqlstore

import (
	"github.com/go-gorp/gorp"
	_ "github.com/lib/pq"
	"github.com/webitel/engine/store"
)

type SqlStore interface {
	GetMaster() *gorp.DbMap
	GetReplica() *gorp.DbMap
	GetAllConns() []*gorp.DbMap

	User() store.UserStore
	Calendar() store.CalendarStore
	Skill() store.SkillStore
	AgentTeam() store.AgentTeamStore
	Agent() store.AgentStore
	AgentSkill() store.AgentSkillStore
	ResourceTeam() store.ResourceTeamStore
	OutboundResource() store.OutboundResourceStore
	OutboundResourceGroup() store.OutboundResourceGroupStore
	OutboundResourceInGroup() store.OutboundResourceInGroupStore
	Queue() store.QueueStore
	QueueResource() store.QueueResourceStore
	Bucket() store.BucketSore
	BucketInQueue() store.BucketInQueueStore
	QueueRouting() store.QueueRoutingStore
	SupervisorTeam() store.SupervisorTeamStore
	CommunicationType() store.CommunicationTypeStore
	List() store.ListStore

	Member() store.MemberStore

	RoutingSchema() store.RoutingSchemaStore
	RoutingInboundCall() store.RoutingInboundCallStore
	RoutingOutboundCall() store.RoutingOutboundCallStore
	RoutingVariable() store.RoutingVariableStore
}
