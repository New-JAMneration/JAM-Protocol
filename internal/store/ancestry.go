package store

import (
	"sync"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

// AncestryStore represents a thread-safe ancestry storage
// The maximum length is MaxLookupAge
type AncestryStore struct {
	mu       sync.RWMutex
	ancestry types.Ancestry
}

// NewAncestryStore creates a new AncestryStore
func NewAncestryStore() *AncestryStore {
	return &AncestryStore{
		ancestry: make(types.Ancestry, 0, types.MaxLookupAge),
	}
}

// AppendAncestry appends ancestry items to the store.
// It maintains a maximum length of MaxLookupAge.
func (a *AncestryStore) AppendAncestry(newAncestry types.Ancestry) {
	if len(newAncestry) == 0 {
		return
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	// Append new ancestry items
	a.ancestry = append(a.ancestry, newAncestry...)

	// Trim to MaxLookupAge if exceeded
	if len(a.ancestry) > types.MaxLookupAge {
		// Keep only the most recent MaxLookupAge items
		startIdx := len(a.ancestry) - types.MaxLookupAge
		a.ancestry = a.ancestry[startIdx:]
	}
}

// GetAncestry returns the current ancestry.
// Returns a copy to prevent external modification.
func (a *AncestryStore) GetAncestry() types.Ancestry {
	a.mu.RLock()
	defer a.mu.RUnlock()

	// Return a copy to prevent external modification
	result := make(types.Ancestry, len(a.ancestry))
	copy(result, a.ancestry)
	return result
}

// Clear clears all ancestry
func (a *AncestryStore) Clear() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.ancestry = make(types.Ancestry, 0, types.MaxLookupAge)
}
