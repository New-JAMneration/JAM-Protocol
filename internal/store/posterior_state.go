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
