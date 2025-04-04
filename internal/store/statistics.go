package store

import (
	"sync"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

// AccumulationStatistics
type AccumulationStatistics struct {
	mu                     sync.RWMutex
	accumulationStatistics types.AccumulationStatistics
}

func NewAccumulationStatistics() *AccumulationStatistics {
	return &AccumulationStatistics{
		accumulationStatistics: types.AccumulationStatistics{},
	}
}

func (a *AccumulationStatistics) GetAccumulationStatistics() types.AccumulationStatistics {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.accumulationStatistics
}

func (a *AccumulationStatistics) SetAccumulationStatistics(w types.AccumulationStatistics) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.accumulationStatistics = w
}

// DeferredTransfersStatistics
type DeferredTransfersStatistics struct {
	mu                          sync.RWMutex
	deferredTransfersStatistics types.DeferredTransfersStatistics
}

func NewDeferredTransfersStatistics() *DeferredTransfersStatistics {
	return &DeferredTransfersStatistics{
		deferredTransfersStatistics: types.DeferredTransfersStatistics{},
	}
}

func (d *DeferredTransfersStatistics) GetDeferredTransfersStatistics() types.DeferredTransfersStatistics {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.deferredTransfersStatistics
}

func (d *DeferredTransfersStatistics) SetDeferredTransfersStatistics(w types.DeferredTransfersStatistics) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.deferredTransfersStatistics = w
}
