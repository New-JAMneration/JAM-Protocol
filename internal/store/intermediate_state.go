package store

import (
	"sync"

	types "github.com/New-JAMneration/JAM-Protocol/internal/types"
)

type IntermediateStates struct {
	mu    sync.RWMutex
	state *types.State
}

func NewIntermediateStates() *IntermediateStates {
	return &IntermediateStates{
		state: &types.State{},
	}
}

func (s *IntermediateStates) GetState() types.State {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return *s.state
}

func (s *IntermediateStates) GenerateGenesisState(state types.State) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state = &state
}

func (s *IntermediateStates) SetBeta(beta types.BlocksHistory) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Beta = beta
}
