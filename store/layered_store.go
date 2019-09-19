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

func (s *LayeredStore) RoutingScheme() RoutingSchemeStore {
	return s.DatabaseLayer.RoutingScheme()
}

func (s *LayeredStore) RoutingInboundCall() RoutingInboundCallStore {
	return s.DatabaseLayer.RoutingInboundCall()
}

func (s *LayeredStore) RoutingOutboundCall() RoutingOutboundCallStore {
	return s.DatabaseLayer.RoutingOutboundCall()
}
