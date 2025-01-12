package store

import (
	"sync"

	types "github.com/New-JAMneration/JAM-Protocol/internal/types"
)

type IntermediateStates struct {
	mu    sync.RWMutex
	state *IntermediateState
}

type IntermediateState struct {
	Delta      types.ServiceAccountState
	BetaDagger types.BlocksHistory
	RhoDagger  types.AvailabilityAssignments
}

func NewIntermediateStates() *IntermediateStates {
	return &IntermediateStates{
		state: &IntermediateState{
			Delta:      types.ServiceAccountState{},
			BetaDagger: types.BlocksHistory{},
			RhoDagger:  types.AvailabilityAssignments{},
		},
	}
}

// Delta
func (s *IntermediateStates) GetDelta() types.ServiceAccountState {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state.Delta
}

func (s *IntermediateStates) SetDelta(delta types.ServiceAccountState) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Delta = delta
}

// BetaDagger
func (s *IntermediateStates) GetBetaDagger() types.BlocksHistory {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state.BetaDagger
}

func (s *IntermediateStates) SetBetaDagger(betaDagger types.BlocksHistory) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.BetaDagger = betaDagger
}

// RhoDagger
func (s *IntermediateStates) GetRhoDagger() types.AvailabilityAssignments {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state.RhoDagger
}

func (s *IntermediateStates) SetRhoDagger(rhoDagger types.AvailabilityAssignments) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.RhoDagger = rhoDagger
}
