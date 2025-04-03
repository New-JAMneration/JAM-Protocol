package store

import (
	"sync"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

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
