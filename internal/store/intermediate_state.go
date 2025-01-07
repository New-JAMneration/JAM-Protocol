package store

import (
	"sync"

	jamTypes "github.com/New-JAMneration/JAM-Protocol/internal/types"
)

type IntermediateStates struct {
	mu        sync.RWMutex
	RhoDagger *jamTypes.AvailabilityAssignments
}

func NewIntermediateStates() *IntermediateStates {
	return &IntermediateStates{
		RhoDagger: &jamTypes.AvailabilityAssignments{},
	}
}

func (s *IntermediateStates) GetRhoDagger() jamTypes.AvailabilityAssignments {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return *s.RhoDagger
}

func (s *IntermediateStates) SetRhoDagger(rhoDagger jamTypes.AvailabilityAssignments) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.RhoDagger = &rhoDagger
}
