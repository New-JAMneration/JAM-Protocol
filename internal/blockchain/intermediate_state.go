package blockchain

import (
	"sync"

	types "github.com/New-JAMneration/JAM-Protocol/internal/types"
)

type IntermediateStates struct {
	mu    sync.RWMutex
	state *IntermediateState
}

type IntermediateState struct {
	BetaHDagger       types.BlocksHistory
	RhoDagger         types.AvailabilityAssignments
	RhoDoubleDagger   types.AvailabilityAssignments
	DeltaDagger       types.ServiceAccountState
	DeltaDoubleDagger types.ServiceAccountState
	// (11.16) \mathbf{W} GP 0.6.4
	AvailableWorkReports []types.WorkReport
	PresentWorkReports   []types.WorkReport
	// (12.4) \mathbf{W}^! GP 0.6.4
	AccumulatedWorkReports []types.WorkReport
	// (12.5) \mathbf{W}^Q GP 0.6.4
	QueuedWorkReports types.ReadyQueueItem
	// (12.11) \mathbf{W}^* GP 0.6.4
	AccumulatableWorkReports []types.WorkReport
	AccumulationStatistics   types.AccumulationStatistics
	// (7.7) b: MR(β′_B) GP 0.6.7 Only used for test-vector
	MmrCommitment types.OpaqueHash
}

func NewIntermediateStates() *IntermediateStates {
	return &IntermediateStates{
		state: &IntermediateState{
			BetaHDagger:              types.BlocksHistory{},
			RhoDagger:                make(types.AvailabilityAssignments, types.CoresCount),
			RhoDoubleDagger:          make(types.AvailabilityAssignments, types.CoresCount),
			DeltaDagger:              types.ServiceAccountState{},
			DeltaDoubleDagger:        types.ServiceAccountState{},
			AvailableWorkReports:     []types.WorkReport{},
			PresentWorkReports:       []types.WorkReport{},
			AccumulatedWorkReports:   []types.WorkReport{},
			QueuedWorkReports:        types.ReadyQueueItem{},
			AccumulatableWorkReports: []types.WorkReport{},
			AccumulationStatistics:   types.AccumulationStatistics{},
			MmrCommitment:            types.OpaqueHash{},
		},
	}
}

// BetaHDagger
func (s *IntermediateStates) GetBetaHDagger() types.BlocksHistory {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state.BetaHDagger
}

func (s *IntermediateStates) SetBetaHDagger(betaHDagger types.BlocksHistory) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.BetaHDagger = betaHDagger
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

func (s *IntermediateStates) SetAvailableWorkReports(w []types.WorkReport) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.AvailableWorkReports = w
}

func (s *IntermediateStates) GetAvailableWorkReports() []types.WorkReport {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state.AvailableWorkReports
}

func (s *IntermediateStates) SetPresentWorkReports(w []types.WorkReport) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.PresentWorkReports = w
}

func (s *IntermediateStates) GetPresentWorkReports() []types.WorkReport {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state.PresentWorkReports
}

func (s *IntermediateStates) SetAccumulatedWorkReports(w []types.WorkReport) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.AccumulatedWorkReports = w
}

func (s *IntermediateStates) GetAccumulatedWorkReports() []types.WorkReport {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state.AccumulatedWorkReports
}

func (s *IntermediateStates) SetQueuedWorkReports(w types.ReadyQueueItem) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.QueuedWorkReports = w
}

func (s *IntermediateStates) GetQueuedWorkReports() types.ReadyQueueItem {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state.QueuedWorkReports
}

func (s *IntermediateStates) SetAccumulatableWorkReports(w []types.WorkReport) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.AccumulatableWorkReports = w
}

func (s *IntermediateStates) GetAccumulatableWorkReports() []types.WorkReport {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state.AccumulatableWorkReports
}

func (s *IntermediateStates) SetAccumulationStatistics(w types.AccumulationStatistics) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.AccumulationStatistics = w
}

func (s *IntermediateStates) GetAccumulationStatistics() types.AccumulationStatistics {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state.AccumulationStatistics
}

func (s *IntermediateStates) SetMmrCommitment(mmrCommitment types.OpaqueHash) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.MmrCommitment = mmrCommitment
}

func (s *IntermediateStates) GetMmrCommitment() types.OpaqueHash {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state.MmrCommitment
}
