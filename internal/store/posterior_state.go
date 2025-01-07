package store

import (
	"sync"

	types "github.com/New-JAMneration/JAM-Protocol/internal/types"
)

type PosteriorStates struct {
	mu    sync.RWMutex
	state *types.State
}

func NewPosteriorStates() *PosteriorStates {
	return &PosteriorStates{
		state: &types.State{},
	}
}

func (s *PosteriorStates) GetState() types.State {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return *s.state
}

func (s *PosteriorStates) GenerateGenesisState(state types.State) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state = &state
}

func (s *PosteriorStates) SetBeta(beta types.BlocksHistory) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Beta = beta
}

func (s *PosteriorStates) SetPsiG(psiG []types.WorkReportHash) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Psi.Good = psiG
}

func (s *PosteriorStates) SetPsiB(psiB []types.WorkReportHash) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Psi.Bad = psiB
}

func (s *PosteriorStates) SetPsiW(psiW []types.WorkReportHash) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Psi.Wonky = psiW
}

func (s *PosteriorStates) SetPsiO(psiO []types.Ed25519Public) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Psi.Offenders = psiO
}
