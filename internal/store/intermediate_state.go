package store

import (
	"sync"

	types "github.com/New-JAMneration/JAM-Protocol/internal/types"
)

type IntermediateStates struct {
	mu    sync.RWMutex
	state *IntermediateState
}

type IntermediateState struct {
	BetaDagger               types.BlocksHistory
	RhoDagger                types.AvailabilityAssignments
	RhoDoubleDagger          types.AvailabilityAssignments
	DeltaDagger              types.ServiceAccountState
	DeltaDoubleDagger        types.ServiceAccountState
	BeefyCommitmentOutputs   BeefyCommitmentOutputs
	AvailableWorkReports     AvailableWorkReports
	PresentWorkReports       PresentWorkReports
	AccumulatedWorkReports   AccumulatedWorkReports
	QueuedWorkReports        QueuedWorkReports
	AccumulatableWorkReports AccumulatableWorkReports
}

func NewIntermediateStates() *IntermediateStates {
	return &IntermediateStates{
		state: &IntermediateState{
			BetaDagger:               types.BlocksHistory{},
			RhoDagger:                types.AvailabilityAssignments{},
			RhoDoubleDagger:          types.AvailabilityAssignments{},
			DeltaDagger:              types.ServiceAccountState{},
			DeltaDoubleDagger:        types.ServiceAccountState{},
			BeefyCommitmentOutputs:   *NewBeefyCommitmentOutputs(),
			AvailableWorkReports:     *NewAvailableWorkReports(),
			PresentWorkReports:       *NewPresentWorkReports(),
			AccumulatedWorkReports:   *NewAccumulatedWorkReports(),
			QueuedWorkReports:        *NewQueuedWorkReports(),
			AccumulatableWorkReports: *NewAccumulatableWorkReports(),
		},
	}
}

// BetaDagger
func (s *IntermediateStates) GetBetaDagger() types.BlocksHistory {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state.BetaDagger
}

func (s *IntermediateStates) SetBetaDagger(betaDagger types.BlocksHistory) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.BetaDagger = betaDagger
}

// RhoDagger
func (s *IntermediateStates) GetRhoDagger() types.AvailabilityAssignments {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state.RhoDagger
}

func (s *IntermediateStates) SetRhoDagger(rhoDagger types.AvailabilityAssignments) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.RhoDagger = rhoDagger
}

func (s *IntermediateStates) SetRhoDoubleDagger(RhoDoubleDagger types.AvailabilityAssignments) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.RhoDoubleDagger = RhoDoubleDagger
}

func (s *IntermediateStates) GetRhoDoubleDagger() types.AvailabilityAssignments {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state.RhoDoubleDagger
}

func (s *IntermediateStates) SetDeltaDagger(deltaDagger types.ServiceAccountState) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.DeltaDagger = deltaDagger
}

func (s *IntermediateStates) GetDeltaDagger() types.ServiceAccountState {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state.DeltaDagger
}

func (s *IntermediateStates) SetDeltaDoubleDagger(deltaDoubleDagger types.ServiceAccountState) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.DeltaDoubleDagger = deltaDoubleDagger
}

func (s *IntermediateStates) GetDeltaDoubleDagger() types.ServiceAccountState {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state.DeltaDoubleDagger
}

func (s *IntermediateStates) SetBeefyCommitmentOutput(c types.AccumulatedServiceOutput) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.BeefyCommitmentOutputs.SetBeefyCommitmentOutput(c)
}

func (s *IntermediateStates) GetBeefyCommitmentOutput() types.AccumulatedServiceOutput {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state.BeefyCommitmentOutputs.GetBeefyCommitmentOutput()
}

func (s *IntermediateStates) SetAvailableWorkReports(w []types.WorkReport) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.AvailableWorkReports.SetAvailableWorkReports(w)
}

func (s *IntermediateStates) GetAvailableWorkReports() []types.WorkReport {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state.AvailableWorkReports.GetAvailableWorkReports()
}

func (s *IntermediateStates) SetPresentWorkReports(w []types.WorkReport) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.PresentWorkReports.SetPresentWorkReports(w)
}

func (s *IntermediateStates) GetPresentWorkReports() []types.WorkReport {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state.PresentWorkReports.GetPresentWorkReports()
}

func (s *IntermediateStates) SetAccumulatedWorkReports(w []types.WorkReport) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.AccumulatedWorkReports.SetAccumulatedWorkReports(w)
}

func (s *IntermediateStates) GetAccumulatedWorkReports() []types.WorkReport {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state.AccumulatedWorkReports.GetAccumulatedWorkReports()
}

func (s *IntermediateStates) SetQueuedWorkReports(w types.ReadyQueueItem) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.QueuedWorkReports.SetQueuedWorkReports(w)
}

func (s *IntermediateStates) GetQueuedWorkReports() types.ReadyQueueItem {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state.QueuedWorkReports.GetQueuedWorkReports()
}

func (s *IntermediateStates) SetAccumulatableWorkReports(w []types.WorkReport) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.AccumulatableWorkReports.SetAccumulatableWorkReports(w)
}

func (s *IntermediateStates) GetAccumulatableWorkReports() []types.WorkReport {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state.AccumulatableWorkReports.GetAccumulatableWorkReports()
}
