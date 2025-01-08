package store

import (
	"sync"

	jamTypes "github.com/New-JAMneration/JAM-Protocol/internal/types"
)

type IntermediateStates struct {
	mu    sync.RWMutex
	state *IntermediateState
}

type IntermediateState struct {
	RhoDagger jamTypes.AvailabilityAssignments
}

func NewIntermediateStates() *IntermediateStates {
	return &IntermediateStates{
		state: &IntermediateState{
			RhoDagger: jamTypes.AvailabilityAssignments{},
		},
	}
}

func (s *IntermediateStates) GetRhoDagger() jamTypes.AvailabilityAssignments {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state.RhoDagger
}

func (s *IntermediateStates) SetRhoDagger(rhoDagger jamTypes.AvailabilityAssignments) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.RhoDagger = rhoDagger
}
