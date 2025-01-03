package store

import (
	"sync"

	jamTypes "github.com/New-JAMneration/JAM-Protocol/internal/types"
)

type States struct {
	mu    sync.RWMutex
	state *jamTypes.State
}

func NewStates() *States {
	return &States{
		state: &jamTypes.State{},
	}
}

func (s *States) GetState() jamTypes.State {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return *s.state
}

func (s *States) GenerateGenesisState(state jamTypes.State) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state = &state
}

func (s *States) SetKappa(kappa jamTypes.ValidatorsData) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Kappa = kappa

}

func (s *States) SetLambda(lambda jamTypes.ValidatorsData) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Lambda = lambda
}

func (s *States) AddPsiOffenders(offender jamTypes.Ed25519Public) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Psi.Offenders = append(s.state.Psi.Offenders, offender)
}
