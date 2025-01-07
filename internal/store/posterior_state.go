package store

import (
	"sync"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
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

func (s *PosteriorStates) SetKappa(kappa types.ValidatorsData) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Kappa = kappa
}

func (s *PosteriorStates) SetLambda(lambda types.ValidatorsData) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Lambda = lambda
}

func (s *PosteriorStates) SetGammaK(gammaK types.ValidatorsData) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Gamma.GammaK = gammaK
}

func (s *PosteriorStates) SetGammaZ(gammaZ types.BandersnatchRingCommitment) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Gamma.GammaZ = gammaZ
}
