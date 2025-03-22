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
	BetaDagger        types.BlocksHistory
	RhoDagger         types.AvailabilityAssignments
	RhoDoubleDagger   types.AvailabilityAssignments
	DeltaDagger       types.ServiceAccountState
	DeltaDoubleDagger types.ServiceAccountState
}

func NewIntermediateStates() *IntermediateStates {
	return &IntermediateStates{
		state: &IntermediateState{
			BetaDagger:      types.BlocksHistory{},
			RhoDagger:       types.AvailabilityAssignments{},
			RhoDoubleDagger: types.AvailabilityAssignments{},
		},
	}
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

func (s *IntermediateStates) SetRhoDoubleDagger(RhoDoubleDagger types.AvailabilityAssignments) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.RhoDoubleDagger = RhoDoubleDagger
}

func (s *IntermediateStates) GetRhoDoubleDagger() types.AvailabilityAssignments {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state.RhoDoubleDagger
}

func (s *IntermediateStates) SetDeltaDagger(deltaDagger types.ServiceAccountState) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.DeltaDagger = deltaDagger
}

func (s *IntermediateStates) GetDeltaDagger() types.ServiceAccountState {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state.DeltaDagger
}

func (s *IntermediateStates) SetDeltaDoubleDagger(deltaDoubleDagger types.ServiceAccountState) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.DeltaDoubleDagger = deltaDoubleDagger
}

func (s *IntermediateStates) GetDeltaDoubleDagger() types.ServiceAccountState {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state.DeltaDoubleDagger
}
