package store

import (
	"sync"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

// AccumulatedServiceOutput (12.15) B used in Chapter 7 (7.3)
type BeefyCommitmentOutputs struct {
	mu                    sync.RWMutex
	serviceCommitmentPair types.AccumulatedServiceOutput
}

func NewBeefyCommitmentOutputs() *BeefyCommitmentOutputs {
	return &BeefyCommitmentOutputs{
		serviceCommitmentPair: make(types.AccumulatedServiceOutput),
	}
}

func (b *BeefyCommitmentOutputs) GetBeefyCommitmentOutput() types.AccumulatedServiceOutput {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.serviceCommitmentPair
}

func (b *BeefyCommitmentOutputs) SetBeefyCommitmentOutput(c types.AccumulatedServiceOutput) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.serviceCommitmentPair = c
}

// accumulatedWorkReports
// W^! (accumulated immediately)
type AccumulatedWorkReports struct {
	mu          sync.RWMutex
	workReports []types.WorkReport
}

func NewAccumulatedWorkReports() *AccumulatedWorkReports {
	return &AccumulatedWorkReports{
		workReports: []types.WorkReport{},
	}
}

func (a *AccumulatedWorkReports) GetAccumulatedWorkReports() []types.WorkReport {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.workReports
}

func (a *AccumulatedWorkReports) SetAccumulatedWorkReports(w []types.WorkReport) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.workReports = w
}

// queuedWorkReports
// W^Q (queued execution)
type QueuedWorkReports struct {
	mu          sync.RWMutex
	workReports types.ReadyQueueItem
}

func NewQueuedWorkReports() *QueuedWorkReports {
	return &QueuedWorkReports{
		workReports: types.ReadyQueueItem{},
	}
}

func (q *QueuedWorkReports) GetQueuedWorkReports() types.ReadyQueueItem {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return q.workReports
}

func (q *QueuedWorkReports) SetQueuedWorkReports(w types.ReadyQueueItem) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.workReports = w
}

// accumulatableWorkReports
// W^* (accumulatable work-reports in this block)
type AccumulatableWorkReports struct {
	mu          sync.RWMutex
	workReports []types.WorkReport
}

func NewAccumulatableWorkReports() *AccumulatableWorkReports {
	return &AccumulatableWorkReports{
		workReports: []types.WorkReport{},
	}
}

func (a *AccumulatableWorkReports) GetAccumulatableWorkReports() []types.WorkReport {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.workReports
}

func (a *AccumulatableWorkReports) SetAccumulatableWorkReports(w []types.WorkReport) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.workReports = w
}

// AvailableWorkReports
// W (available work-reports) (11.16)
type AvailableWorkReports struct {
	mu          sync.RWMutex
	workReports []types.WorkReport
}

func NewAvailableWorkReports() *AvailableWorkReports {
	return &AvailableWorkReports{
		workReports: []types.WorkReport{},
	}
}

func (a *AvailableWorkReports) GetAvailableWorkReports() []types.WorkReport {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.workReports
}

func (a *AvailableWorkReports) SetAvailableWorkReports(w []types.WorkReport) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.workReports = w
}

// PresentWorkReports
// w (the set of work-reports in the present extrinsic) (11.28)
type PresentWorkReports struct {
	mu          sync.RWMutex
	workReports []types.WorkReport
}

func NewPresentWorkReports() *PresentWorkReports {
	return &PresentWorkReports{
		workReports: []types.WorkReport{},
	}
}

func (p *PresentWorkReports) GetPresentWorkReports() []types.WorkReport {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.workReports
}

func (p *PresentWorkReports) SetPresentWorkReports(w []types.WorkReport) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.workReports = w
}
