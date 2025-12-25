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

// KeepAncestryUpTo keeps only ancestry items up to and including the specified headerHash.
// If the headerHash is not found, it clears all ancestry.
func (a *AncestryStore) KeepAncestryUpTo(headerHash types.HeaderHash) {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Find the index of the headerHash in ancestry
	foundIdx := -1
	for i := len(a.ancestry) - 1; i >= 0; i-- {
		if a.ancestry[i].HeaderHash == headerHash {
			foundIdx = i
			break
		}
	}

	if foundIdx == -1 {
		// HeaderHash not found, clear all ancestry
		a.ancestry = make(types.Ancestry, 0, types.MaxLookupAge)
		return
	}

	// Keep only items up to and including the found index
	a.ancestry = a.ancestry[:foundIdx+1]
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
