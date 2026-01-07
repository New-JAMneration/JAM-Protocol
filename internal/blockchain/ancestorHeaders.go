package blockchain

import (
	"sync"
	"time"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

// AncestorHeaders represents 24 hours of ancestor headers
// We only require implmentations to cs headers of ancestors which were
// authored in the previous L = 24 hours of any block B they wish to validate.
// graypaper (5.3)
type AncestorHeaders struct {
	mu              sync.RWMutex
	ancestorHeaders []types.Header
}

// NewAncestorHeaders creates a new AncestorHeaders
func NewAncestorHeaders() *AncestorHeaders {
	return &AncestorHeaders{
		ancestorHeaders: make([]types.Header, 0),
	}
}

// AddHeader adds a header to the ancestorHeaders
func (a *AncestorHeaders) AddHeader(header types.Header) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.ancestorHeaders = append(a.ancestorHeaders, header)
	a.updateAncestorHeaders()
}

// GetHeaders returns the ancestorHeaders
func (a *AncestorHeaders) GetHeaders() []types.Header {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.ancestorHeaders
}

// We only require implmentations to cs headers of ancestors which were
// authored in the previous L = 24 hours of any block B they wish to validate.
// graypaper (5.3)
// The ancestorHeaders is ordered by the slot, we can remove the
// headers older than 24 hours by checking the slot of the header from the
// oldest one.
func (a *AncestorHeaders) updateAncestorHeaders() {
	// Get the current time
	currentTime := time.Now().UTC()

	// Check the slot of the header from the oldest one
	for i, header := range a.ancestorHeaders {
		// Get the time of the header
		headerTime := types.JamCommonEra.Add(time.Duration(header.Slot*types.TimeSlot(types.SlotPeriod)) * time.Second)

		// If the header is older than 24 hours, remove the header
		if currentTime.Sub(headerTime) > 24*time.Hour {
			a.ancestorHeaders = a.ancestorHeaders[i+1:]
		}
	}
}
