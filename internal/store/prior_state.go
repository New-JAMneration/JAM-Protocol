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

func (s *PriorStates) SetDelta(delta types.ServiceAccountState) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Delta = delta
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

func (s *PriorStates) SetEta(eta types.EntropyBuffer) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Eta = eta
}

func (s *PriorStates) SetTau(tau types.TimeSlot) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Tau = tau
}

func (s *PriorStates) SetGammaA(gammaA []types.TicketBody) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Gamma.GammaA = gammaA
}

func (s *PriorStates) SetPsiG(psiG []types.WorkReportHash) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Psi.Good = psiG
}

func (s *PriorStates) SetPsiB(psiB []types.WorkReportHash) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Psi.Bad = psiB
}

func (s *PriorStates) SetPsiW(psiW []types.WorkReportHash) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Psi.Wonky = psiW
}

func (s *PriorStates) SetPsiO(psiO []types.Ed25519Public) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Psi.Offenders = psiO
}

func (s *PriorStates) SetRho(rho types.AvailabilityAssignments) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Rho = rho
}
