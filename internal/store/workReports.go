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
