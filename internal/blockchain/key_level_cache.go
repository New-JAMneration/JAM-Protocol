package blockchain

import (
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
)

// KeyLevelCache caches leaf hashes for individual state keys.
type KeyLevelCache struct {
	// Cache: key + valueHash -> leaf hash
	cache map[cacheKey]types.OpaqueHash
}

type cacheKey struct {
	stateKey  types.StateKey
	valueHash types.OpaqueHash
}

// NewKeyLevelCache creates a new key-level cache
func NewKeyLevelCache() *KeyLevelCache {
	return &KeyLevelCache{
		cache: make(map[cacheKey]types.OpaqueHash),
	}
}

// GetLeafHash attempts to get a cached leaf hash for a key-value pair.
// Returns the cached hash and true if found, or zero hash and false if not cached.
func (c *KeyLevelCache) GetLeafHashCache(key types.StateKey, value []byte) (types.OpaqueHash, bool) {
	valueHash := hash.Blake2bHash(value)
	ck := cacheKey{stateKey: key, valueHash: valueHash}

	if leafHash, ok := c.cache[ck]; ok {
		return leafHash, true
	}

	return types.OpaqueHash{}, false
}

// PutLeafHash stores a leaf hash in the cache.
func (c *KeyLevelCache) AddLeafHashCache(key types.StateKey, value []byte, leafHash types.OpaqueHash) {
	valueHash := hash.Blake2bHash(value)
	ck := cacheKey{stateKey: key, valueHash: valueHash}
	c.cache[ck] = leafHash
}

// GetOrComputeLeafHash gets a cached leaf hash or computes it.
func (c *KeyLevelCache) GetOrComputeLeafHash(
	key types.StateKey,
	value []byte,
	computeFn func(types.StateKey, []byte) types.OpaqueHash,
) types.OpaqueHash {
	if leafHash, ok := c.GetLeafHashCache(key, value); ok {
		return leafHash
	}

	leafHash := computeFn(key, value)
	c.AddLeafHashCache(key, value, leafHash)
	return leafHash
}

// Clear removes all entries from the cache.
func (c *KeyLevelCache) Clear() {
	c.cache = make(map[cacheKey]types.OpaqueHash)
}
