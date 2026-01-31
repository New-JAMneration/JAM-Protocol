package blockchain

import (
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
)

// leafCacheEntry stores the latest (valueHash, leafHash) for a stateKey.
// Only one entry per stateKey; updated when value changes.
type leafCacheEntry struct {
	valueHash types.OpaqueHash // Blake2bHash(value)
	leafHash  types.OpaqueHash // Merkle leaf hash for (key, value)
}

// KeyLevelCache keeps at most one entry per stateKey.
// For a given stateKey, we only recompute the leaf hash when its value actually changes.
type KeyLevelCache struct {
	entries map[types.StateKey]leafCacheEntry
}

// NewKeyLevelCache creates a new key-level cache.
func NewKeyLevelCache() *KeyLevelCache {
	return &KeyLevelCache{
		entries: make(map[types.StateKey]leafCacheEntry),
	}
}

// GetLeafHash returns the cached leaf hash if the key's value matches the cached valueHash.
// Also returns the computed valueHash so the caller can reuse it (e.g. when storing on miss).
// Returns (leafHash, valueHash, true) on hit, (zero, valueHash, false) on miss.
func (c *KeyLevelCache) GetLeafHash(key types.StateKey, value []byte) (leafHash types.OpaqueHash, valueHash types.OpaqueHash, ok bool) {
	valueHash = hash.Blake2bHash(value)
	entry, found := c.entries[key]
	if !found || entry.valueHash != valueHash {
		return types.OpaqueHash{}, valueHash, false
	}
	return entry.leafHash, valueHash, true
}

// Len returns the number of entries in the cache.
func (c *KeyLevelCache) Len() int {
	return len(c.entries)
}

// PutLeafHash stores the leaf hash for (key, value). Caller must pass the same valueHash
// returned from GetLeafHash (or computed once) to avoid hashing value twice.
func (c *KeyLevelCache) PutLeafHash(key types.StateKey, valueHash types.OpaqueHash, leafHash types.OpaqueHash) {
	c.entries[key] = leafCacheEntry{valueHash: valueHash, leafHash: leafHash}
}

// GetOrComputeLeafHash gets a cached leaf hash or computes it.
// ValueHash is computed once and reused for PutLeafHash on miss.
func (c *KeyLevelCache) GetOrComputeLeafHash(
	key types.StateKey,
	value []byte,
	computeFn func(types.StateKey, []byte) types.OpaqueHash,
) types.OpaqueHash {
	leafHash, valueHash, ok := c.GetLeafHash(key, value)
	if ok {
		return leafHash
	}
	leafHash = computeFn(key, value)
	c.PutLeafHash(key, valueHash, leafHash)
	return leafHash
}

// Clear removes all entries from the cache.
func (c *KeyLevelCache) Clear() {
	c.entries = make(map[types.StateKey]leafCacheEntry)
}
