package store

import (
	"sync"

	jamTypes "github.com/New-JAMneration/JAM-Protocol/internal/types"
)

type PriorStates struct {
	mu    sync.RWMutex
	state *jamTypes.State
}

func NewPriorStates() *PriorStates {
	return &PriorStates{
		state: &jamTypes.State{},
	}
}

func (s *PriorStates) GetState() jamTypes.State {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return *s.state
}

func (s *PriorStates) GenerateGenesisState(state jamTypes.State) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state = &state
}

func (s *PriorStates) SetKappa(kappa jamTypes.ValidatorsData) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Kappa = kappa

}

func (s *PriorStates) SetLambda(lambda jamTypes.ValidatorsData) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Lambda = lambda
}

func (s *PriorStates) AddPsiOffenders(offender jamTypes.Ed25519Public) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Psi.Offenders = append(s.state.Psi.Offenders, offender)
}
