package store

import (
	"sync"

	jamTypes "github.com/New-JAMneration/JAM-Protocol/internal/types"
)

type PosteriorStates struct {
	mu    sync.RWMutex
	state *jamTypes.State
}

func NewPosteriorStates() *PosteriorStates {
	return &PosteriorStates{
		state: &jamTypes.State{},
	}
}

func (s *PosteriorStates) GetState() jamTypes.State {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return *s.state
}

func (s *PosteriorStates) SetRho(rho jamTypes.AvailabilityAssignments) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Rho = rho
}

func (s *PosteriorStates) SetPsiG(psiG []jamTypes.WorkReportHash) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Psi.Good = psiG
}

func (s *PosteriorStates) SetPsiB(psiB []jamTypes.WorkReportHash) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Psi.Bad = psiB
}

func (s *PosteriorStates) SetPsiW(psiW []jamTypes.WorkReportHash) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Psi.Wonky = psiW
}

func (s *PosteriorStates) SetPsiO(psiO []jamTypes.Ed25519Public) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Psi.Offenders = psiO
}
