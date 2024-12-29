package store

import (
	"sync"
	"time"

	jamTypes "github.com/New-JAMneration/JAM-Protocol/internal/types"
)

// AncestorHeaders represents 24 hours of ancestor headers
// We only require implmentations to store headers of ancestors which were
// authored in the previous L = 24 hours of any block B they wish to validate.
// graypaper (5.3)
type AncestorHeaders struct {
	mu              sync.RWMutex
	ancestorHeaders []jamTypes.Header
}

// NewAncestorHeaders creates a new AncestorHeaders
func NewAncestorHeaders() *AncestorHeaders {
	return &AncestorHeaders{
		ancestorHeaders: make([]jamTypes.Header, 0),
	}
}

// AddHeader adds a header to the ancestorHeaders
func (a *AncestorHeaders) AddHeader(header jamTypes.Header) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.ancestorHeaders = append(a.ancestorHeaders, header)
	a.updateAncestorHeaders()
}

// GetHeaders returns the ancestorHeaders
func (a *AncestorHeaders) GetHeaders() []jamTypes.Header {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.ancestorHeaders
}

// We only require implmentations to store headers of ancestors which were
// authored in the previous L = 24 hours of any block B they wish to validate.
// graypaper (5.3)
// The ancestorHeaders is ordered by the slot, we can remove the
// headers older than 24 hours by checking the slot of the header from the
// oldest one.
func (a *AncestorHeaders) updateAncestorHeaders() {
	const slotPeriod = 6

	// Get the current time
	currentTime := time.Now()

	// Check the slot of the header from the oldest one
	for i, header := range a.ancestorHeaders {
		// Get the time of the header
		headerTime := time.Unix(int64(header.Slot*slotPeriod), 0)

		// If the header is older than 24 hours, remove the header
		if currentTime.Sub(headerTime) > 24*time.Hour {
			a.ancestorHeaders = a.ancestorHeaders[i+1:]
		}
	}
}
