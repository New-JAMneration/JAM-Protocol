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
// storeNode / storeValue are optional callbacks; pass nil for dry-run (original behaviour).
func merklize(entries []types.StateKeyVal, depth int, storeNode StoreNodeFunc, storeValue StoreValueFunc) (types.OpaqueHash, error) {
	if len(entries) == 0 {
		return types.OpaqueHash{}, nil
	}
	if len(entries) == 1 {
		value := entries[0].Value
		if len(value) > 32 && storeValue != nil {
			if err := storeValue(value); err != nil {
				return types.OpaqueHash{}, err
			}
		}
		node := encodeLeafNode(entries[0].Key, value)
		nodeHash := hash.Blake2bHash(node[:])
		if storeNode != nil {
			if err := storeNode(nodeHash, TrieNode(node)); err != nil {
				return types.OpaqueHash{}, err
			}
		}
		return nodeHash, nil
	}

	pivot := partitionByBit(entries, depth)
	leftHash, err := merklize(entries[:pivot], depth+1, storeNode, storeValue)
	if err != nil {
		return types.OpaqueHash{}, err
	}
	rightHash, err := merklize(entries[pivot:], depth+1, storeNode, storeValue)
	if err != nil {
		return types.OpaqueHash{}, err
	}

	node := encodeBranchNode(leftHash, rightHash)
	nodeHash := hash.Blake2bHash(node[:])
	if storeNode != nil {
		if err := storeNode(nodeHash, TrieNode(node)); err != nil {
			return types.OpaqueHash{}, err
		}
	}
	return nodeHash, nil
}

// merklizeWithCache computes the Merkle root hash with leaf-level caching.
// storeNode / storeValue are optional callbacks; pass nil for dry-run.
func merklizeWithCache(entries []types.StateKeyVal, depth int, cache LeafHashCache, storeNode StoreNodeFunc, storeValue StoreValueFunc) (types.OpaqueHash, error) {
	if len(entries) == 0 {
		return types.OpaqueHash{}, nil
	}
	if len(entries) == 1 {
		value := entries[0].Value
		if len(value) > 32 && storeValue != nil {
			if err := storeValue(value); err != nil {
				return types.OpaqueHash{}, err
			}
		}
		var leafHash types.OpaqueHash
		if cache != nil {
			leafHash = cache(entries[0].Key, value)
		} else {
			node := encodeLeafNode(entries[0].Key, value)
			leafHash = hash.Blake2bHash(node[:])
		}
		if storeNode != nil {
			node := encodeLeafNode(entries[0].Key, value)
			if err := storeNode(leafHash, TrieNode(node)); err != nil {
				return types.OpaqueHash{}, err
			}
		}
		return leafHash, nil
	}

	pivot := partitionByBit(entries, depth)
	leftHash, err := merklizeWithCache(entries[:pivot], depth+1, cache, storeNode, storeValue)
	if err != nil {
		return types.OpaqueHash{}, err
	}
	rightHash, err := merklizeWithCache(entries[pivot:], depth+1, cache, storeNode, storeValue)
	if err != nil {
		return types.OpaqueHash{}, err
	}

	node := encodeBranchNode(leftHash, rightHash)
	nodeHash := hash.Blake2bHash(node[:])
	if storeNode != nil {
		if err := storeNode(nodeHash, TrieNode(node)); err != nil {
			return types.OpaqueHash{}, err
		}
	}
	return nodeHash, nil
}

// MerklizationSerializedState computes the Merkle root from serialized state key-vals.
func MerklizationSerializedState(serializedState types.StateKeyVals) (types.StateRoot, error) {
	entries := make([]types.StateKeyVal, len(serializedState))
	copy(entries, serializedState)
	h, err := merklize(entries, 0, nil, nil)
	return types.StateRoot(h), err
}

// MerklizationSerializedStateWithCache computes the Merkle root with key-level caching.
// storeNode / storeValue are optional callbacks for trie node persistence.
func MerklizationSerializedStateWithCache(
	serializedState types.StateKeyVals,
	cache LeafHashCache,
	storeNode StoreNodeFunc,
	storeValue StoreValueFunc,
) (types.StateRoot, error) {
	entries := make([]types.StateKeyVal, len(serializedState))
	copy(entries, serializedState)

	if cache != nil {
		h, err := merklizeWithCache(entries, 0, cache, storeNode, storeValue)
		return types.StateRoot(h), err
	}
	h, err := merklize(entries, 0, storeNode, storeValue)
	return types.StateRoot(h), err
}

// MerklizationState computes the Merkle root from a State.
func MerklizationState(state types.State) (types.StateRoot, error) {
	serializedState, err := StateEncoder(state)
	if err != nil {
		return types.StateRoot{}, err
	}
	return MerklizationSerializedState(serializedState)
}
