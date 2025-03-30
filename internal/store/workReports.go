package store

import (
	"sync"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

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
