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
	BetaDagger      types.BlocksHistory
	RhoDagger       types.AvailabilityAssignments
	RhoDoubleDagger types.AvailabilityAssignments
	TauInput        types.TimeSlot
	EntropyInput    types.Entropy
}

func NewIntermediateStates() *IntermediateStates {
	return &IntermediateStates{
		state: &IntermediateState{
			BetaDagger:      types.BlocksHistory{},
			RhoDagger:       types.AvailabilityAssignments{},
			RhoDoubleDagger: types.AvailabilityAssignments{},
			TauInput:        0,
			EntropyInput:    types.Entropy{},
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

// TauInput
func (s *IntermediateStates) GetTauInput() types.TimeSlot {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state.TauInput
}

func (s *IntermediateStates) SetTauInput(TauInput types.TimeSlot) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.TauInput = TauInput
}

// EntropyInput Y(H_v)
func (s *IntermediateStates) GetEntropyInput() types.Entropy {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state.EntropyInput
}

func (s *IntermediateStates) SetEntropyInput(EntropyInput types.Entropy) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.EntropyInput = EntropyInput
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
