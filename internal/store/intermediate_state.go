package store

import (
	"sync"

	jamTypes "github.com/New-JAMneration/JAM-Protocol/internal/types"
)

type IntermediateStates struct {
	mu    sync.RWMutex
	state *jamTypes.State
}

func NewIntermediateStates() *IntermediateStates {
	return &IntermediateStates{
		state: &jamTypes.State{},
	}
}

func (s *IntermediateStates) GetState() jamTypes.State {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return *s.state
}

func (s *IntermediateStates) SetRho(rho jamTypes.AvailabilityAssignments) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Rho = rho
}
