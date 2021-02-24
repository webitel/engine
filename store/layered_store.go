package store

import (
	"context"
)

type LayeredStoreDatabaseLayer interface {
	LayeredStoreSupplier
	Store
}

type LayeredStore struct {
	TmpContext     context.Context
	DatabaseLayer  LayeredStoreDatabaseLayer
	LayerChainHead LayeredStoreSupplier
}

func NewLayeredStore(db LayeredStoreDatabaseLayer) Store {
	store := &LayeredStore{
		TmpContext:    context.TODO(),
		DatabaseLayer: db,
	}

	return store
}

type QueryFunction func(LayeredStoreSupplier) *LayeredStoreSupplierResult

func (s *LayeredStore) RunQuery(queryFunction QueryFunction) StoreChannel {
	storeChannel := make(StoreChannel)

	go func() {
		result := queryFunction(s.LayerChainHead)
		storeChannel <- result.StoreResult
	}()

	return storeChannel
}

func (s *LayeredStore) User() UserStore {
	return s.DatabaseLayer.User()
}

func (s *LayeredStore) Calendar() CalendarStore {
	return s.DatabaseLayer.Calendar()
}

func (s *LayeredStore) Skill() SkillStore {
	return s.DatabaseLayer.Skill()
}

func (s *LayeredStore) AgentTeam() AgentTeamStore {
	return s.DatabaseLayer.AgentTeam()
}

func (s *LayeredStore) Agent() AgentStore {
	return s.DatabaseLayer.Agent()
}

func (s *LayeredStore) AgentSkill() AgentSkillStore {
	return s.DatabaseLayer.AgentSkill()
}

func (s *LayeredStore) OutboundResource() OutboundResourceStore {
	return s.DatabaseLayer.OutboundResource()
}

func (s *LayeredStore) OutboundResourceGroup() OutboundResourceGroupStore {
	return s.DatabaseLayer.OutboundResourceGroup()
}

func (s *LayeredStore) OutboundResourceInGroup() OutboundResourceInGroupStore {
	return s.DatabaseLayer.OutboundResourceInGroup()
}

func (s *LayeredStore) RoutingSchema() RoutingSchemaStore {
	return s.DatabaseLayer.RoutingSchema()
}

func (s *LayeredStore) RoutingOutboundCall() RoutingOutboundCallStore {
	return s.DatabaseLayer.RoutingOutboundCall()
}

func (s *LayeredStore) RoutingVariable() RoutingVariableStore {
	return s.DatabaseLayer.RoutingVariable()
}

func (s *LayeredStore) Queue() QueueStore {
	return s.DatabaseLayer.Queue()
}

func (s *LayeredStore) QueueResource() QueueResourceStore {
	return s.DatabaseLayer.QueueResource()
}

func (s *LayeredStore) QueueSkill() QueueSkillStore {
	return s.DatabaseLayer.QueueSkill()
}

func (s *LayeredStore) Bucket() BucketSore {
	return s.DatabaseLayer.Bucket()
}

func (s *LayeredStore) BucketInQueue() BucketInQueueStore {
	return s.DatabaseLayer.BucketInQueue()
}

func (s *LayeredStore) CommunicationType() CommunicationTypeStore {
	return s.DatabaseLayer.CommunicationType()
}

func (s *LayeredStore) Member() MemberStore {
	return s.DatabaseLayer.Member()
}

func (s *LayeredStore) List() ListStore {
	return s.DatabaseLayer.List()
}

func (s *LayeredStore) Call() CallStore {
	return s.DatabaseLayer.Call()
}

func (s *LayeredStore) EmailProfile() EmailProfileStore {
	return s.DatabaseLayer.EmailProfile()
}

func (s *LayeredStore) Chat() ChatStore {
	return s.DatabaseLayer.Chat()
}

func (s *LayeredStore) Region() RegionStore {
	return s.DatabaseLayer.Region()
}

func (s *LayeredStore) PauseCause() PauseCauseStore {
	return s.DatabaseLayer.PauseCause()
}
