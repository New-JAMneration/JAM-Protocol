package merklization

import (
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
)

// LeafHashCache is a callback function type for caching leaf hashes
// Returns the cached hash and true if found, or zero hash and false if not cached
type LeafHashCache func(key types.StateKey, value []byte) (types.OpaqueHash, bool)

// MerklizationSerializedStateWithCache computes the Merkle root with key-level caching.
// The cache callback is used to retrieve and store leaf hashes, avoiding recomputation
// for unchanged keys.
func MerklizationSerializedStateWithCache(
	serializedState types.StateKeyVals,
	cache LeafHashCache,
) types.StateRoot {
	merklizationInput := make(MerklizationInput)

	// Convert StateKeyVals to merklization input
	for _, stateKeyVal := range serializedState {
		// StateKey is already [31]byte, use it directly as map key
		merklizationInput[stateKeyVal.Key] = stateKeyVal
	}

	// Use the cached version of Merklization if we have cache
	if cache != nil {
		return types.StateRoot(MerklizationWithLeafCache(merklizationInput, cache))
	}

	return types.StateRoot(Merklization(merklizationInput))
}

// MerklizationWithLeafCache is like Merklization but uses cached leaf hashes when available.
// This avoids recomputing leaf hashes for unchanged keys.
func MerklizationWithLeafCache(d MerklizationInput, cache LeafHashCache) types.OpaqueHash {
	if len(d) == 0 {
		return types.OpaqueHash{}
	}

	if len(d) == 1 {
		for _, stateKeyVal := range d {
			// Try cache first
			if cache != nil {
				if cached, ok := cache(stateKeyVal.Key, stateKeyVal.Value); ok {
					return cached
				}
			}

			// Compute if not cached
			leftEncoding := LeafEncoding(stateKeyVal.Key, stateKeyVal.Value)
			bytes, _ := utilities.BitsToBytes(leftEncoding)
			return hash.Blake2bHash(bytes)
		}
	}

	// Split into left and right subtrees
	l := make(MerklizationInput)
	r := make(MerklizationInput)
	for key, value := range d {
		// check the first bit: 0 -> left, 1 -> right
		firstBit := (key[0] & 0x80) == 0

		shiftedKey := shiftKeyLeft(key)

		if firstBit {
			l[shiftedKey] = value
		} else {
			r[shiftedKey] = value
		}
	}

	// Recursively compute left and right subtree hashes
	leftHash := MerklizationWithLeafCache(l, cache)
	rightHash := MerklizationWithLeafCache(r, cache)

	branchEncoding := BranchEncoding(leftHash, rightHash)
	bytes, _ := utilities.BitsToBytes(branchEncoding)
	return hash.Blake2bHash(bytes)
}
