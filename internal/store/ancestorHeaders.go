package store

import (
	"sync"

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
}

// GetHeaders returns the ancestorHeaders
func (a *AncestorHeaders) GetHeaders() []jamTypes.Header {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.ancestorHeaders
}
