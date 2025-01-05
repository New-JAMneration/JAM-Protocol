package store

import (
	"sync"

	types "github.com/New-JAMneration/JAM-Protocol/internal/types"
)

type States struct {
	mu    sync.RWMutex
	state *types.State
}

func NewStates() *States {
	return &States{
		state: &types.State{},
	}
}

func (s *States) GetState() types.State {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return *s.state
}

func (s *States) GenerateGenesisState(state types.State) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state = &state
}

func (s *States) SetBeta(beta types.BlocksHistory) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Beta = beta
}

func (s *States) SetKappa(kappa types.ValidatorsData) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Kappa = kappa

}

func (s *States) SetLambda(lambda types.ValidatorsData) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Lambda = lambda
}

func (s *States) AddPsiOffenders(offender types.Ed25519Public) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Psi.Offenders = append(s.state.Psi.Offenders, offender)
}
