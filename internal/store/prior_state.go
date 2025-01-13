package store

import (
	"sync"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

type PriorStates struct {
	mu    sync.RWMutex
	state *types.State
}

func NewPriorStates() *PriorStates {
	return &PriorStates{
		state: &types.State{},
	}
}

func (s *PriorStates) GetState() types.State {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return *s.state
}

func (s *PriorStates) GenerateGenesisState(state types.State) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state = &state
}

func (s *PriorStates) SetBeta(beta types.BlocksHistory) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Beta = beta
}

func (s *PriorStates) SetKappa(kappa types.ValidatorsData) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Kappa = kappa
}

func (s *PriorStates) SetLambda(lambda types.ValidatorsData) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Lambda = lambda
}

func (s *PriorStates) AddPsiOffenders(offender types.Ed25519Public) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Psi.Offenders = append(s.state.Psi.Offenders, offender)
}

func (s *PriorStates) SetTau(tau types.TimeSlot) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Tau = tau
}

func (s *PriorStates) GetEta() types.EntropyBuffer {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state.Eta
}

func (s *PriorStates) SetEta(eta types.EntropyBuffer) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Eta = eta
}

// gamma_a
func (s *PriorStates) SetGammaA(gammaA types.TicketsAccumulator) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Gamma.GammaA = gammaA
}
