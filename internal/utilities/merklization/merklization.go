package merklization

import (
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
)

// encodeBranchNode encodes a branch node as [64]byte with zero heap allocation.
// Layout: {left[0] & 0x7F, left[1:32], right[0:32]}
func encodeBranchNode(left, right types.OpaqueHash) [64]byte {
	var node [64]byte
	node[0] = left[0] & 0x7F
	copy(node[1:32], left[1:])
	copy(node[32:], right[:])
	return node
}

// encodeLeafNode encodes a leaf node as [64]byte with zero heap allocation.
// Embedded leaf (value <= 32 bytes): {0x80 | len(value), key[:31], value, zero-padding}
// Regular leaf (value > 32 bytes):   {0xC0, key[:31], blake2b(value)}
func encodeLeafNode(key types.StateKey, value []byte) [64]byte {
	var node [64]byte
	if len(value) <= 32 {
		node[0] = 0x80 | byte(len(value))
		copy(node[1:32], key[:])
		copy(node[32:], value)
	} else {
		node[0] = 0xC0
		copy(node[1:32], key[:])
		h := hash.Blake2bHash(value)
		copy(node[32:], h[:])
	}
	return node
}

// EncodeLeafNodeHash computes the leaf hash for (key, value) using [64]byte encoding.
func EncodeLeafNodeHash(key types.StateKey, value []byte) types.OpaqueHash {
	node := encodeLeafNode(key, value)
	return hash.Blake2bHash(node[:])
}

// partitionByBit partitions entries in-place based on the bit at position depth.
// Returns pivot index: entries[:pivot] have bit=0 (left), entries[pivot:] have bit=1 (right).
func partitionByBit(entries []types.StateKeyVal, depth int) int {
	byteIdx := depth / 8
	bitMask := byte(1 << (7 - depth%8))
	left := 0
	for right := range entries {
		if entries[right].Key[byteIdx]&bitMask == 0 {
			entries[left], entries[right] = entries[right], entries[left]
			left++
		}
	}
	return left
}

// merklize computes the Merkle root hash using in-place partition and [64]byte encoding.
func merklize(entries []types.StateKeyVal, depth int) types.OpaqueHash {
	if len(entries) == 0 {
		return types.OpaqueHash{}
	}
	if len(entries) == 1 {
		node := encodeLeafNode(entries[0].Key, entries[0].Value)
		return hash.Blake2bHash(node[:])
	}

	pivot := partitionByBit(entries, depth)
	leftHash := merklize(entries[:pivot], depth+1)
	rightHash := merklize(entries[pivot:], depth+1)

	node := encodeBranchNode(leftHash, rightHash)
	return hash.Blake2bHash(node[:])
}

// merklizeWithCache computes the Merkle root hash with leaf-level caching.
func merklizeWithCache(entries []types.StateKeyVal, depth int, cache LeafHashCache) types.OpaqueHash {
	if len(entries) == 0 {
		return types.OpaqueHash{}
	}
	if len(entries) == 1 {
		if cache != nil {
			return cache(entries[0].Key, entries[0].Value)
		}
		node := encodeLeafNode(entries[0].Key, entries[0].Value)
		return hash.Blake2bHash(node[:])
	}

	pivot := partitionByBit(entries, depth)
	leftHash := merklizeWithCache(entries[:pivot], depth+1, cache)
	rightHash := merklizeWithCache(entries[pivot:], depth+1, cache)

	node := encodeBranchNode(leftHash, rightHash)
	return hash.Blake2bHash(node[:])
}

// MerklizationSerializedState computes the Merkle root from serialized state key-vals.
func MerklizationSerializedState(serializedState types.StateKeyVals) types.StateRoot {
	entries := make([]types.StateKeyVal, len(serializedState))
	copy(entries, serializedState)
	return types.StateRoot(merklize(entries, 0))
}

// MerklizationSerializedStateWithCache computes the Merkle root with key-level caching.
func MerklizationSerializedStateWithCache(
	serializedState types.StateKeyVals,
	cache LeafHashCache,
) types.StateRoot {
	entries := make([]types.StateKeyVal, len(serializedState))
	copy(entries, serializedState)

	if cache != nil {
		return types.StateRoot(merklizeWithCache(entries, 0, cache))
	}
	return types.StateRoot(merklize(entries, 0))
}

// MerklizationState computes the Merkle root from a State.
func MerklizationState(state types.State) types.StateRoot {
	serializedState, _ := StateEncoder(state)
	return MerklizationSerializedState(serializedState)
}
