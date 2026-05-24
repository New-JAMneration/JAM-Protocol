package store

import (
	"fmt"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/merklization"
)

// NodeReader abstracts reading trie nodes (for testing with in-memory maps).
type NodeReader interface {
	GetNode(hash types.OpaqueHash) (merklization.TrieNode, error)
	GetNodeValue(node merklization.TrieNode) ([]byte, error)
}

// IncrementalMerklize computes a new trie root by traversing the prior trie along
// dirty paths and reusing unchanged subtrees.
//
// Algorithm: at each branch node, partition dirty entries by bit[depth] into left/right.
// Subtrees with no dirty entries reuse the old child hash. Subtrees with dirty entries
// recurse. At leaf level, re-encode with new value or handle insert/delete transitions.
func IncrementalMerklize(
	priorRoot types.OpaqueHash,
	dirtyEntries []DirtyEntry,
	reader NodeReader,
	storeNode merklization.StoreNodeFunc,
	storeValue merklization.StoreValueFunc,
) (types.OpaqueHash, error) {
	if len(dirtyEntries) == 0 {
		return priorRoot, nil
	}
	return incrementalMerklizeNode(priorRoot, dirtyEntries, 0, reader, storeNode, storeValue)
}

func incrementalMerklizeNode(
	nodeHash types.OpaqueHash,
	dirtyEntries []DirtyEntry,
	depth int,
	reader NodeReader,
	storeNode merklization.StoreNodeFunc,
	storeValue merklization.StoreValueFunc,
) (types.OpaqueHash, error) {
	// All dirty entries are deletes and we hit an empty subtree
	if nodeHash == (types.OpaqueHash{}) {
		// Check if there are inserts — if so, build a new subtree
		var inserts []DirtyEntry
		for _, e := range dirtyEntries {
			if !e.IsDelete {
				inserts = append(inserts, e)
			}
		}
		if len(inserts) == 0 {
			return types.OpaqueHash{}, nil
		}
		return buildSubtree(inserts, depth, storeNode, storeValue)
	}

	node, err := reader.GetNode(nodeHash)
	if err != nil {
		return types.OpaqueHash{}, fmt.Errorf("incremental: get node %x: %w", nodeHash[:8], err)
	}

	if node.IsLeaf() {
		return handleLeaf(node, nodeHash, dirtyEntries, depth, reader, storeNode, storeValue)
	}

	return handleBranch(node, dirtyEntries, depth, reader, storeNode, storeValue)
}

func handleLeaf(
	node merklization.TrieNode,
	nodeHash types.OpaqueHash,
	dirtyEntries []DirtyEntry,
	depth int,
	reader NodeReader,
	storeNode merklization.StoreNodeFunc,
	storeValue merklization.StoreValueFunc,
) (types.OpaqueHash, error) {
	existingKey := node.GetLeafKey()

	// Find if the existing leaf's key is in the dirty set
	var existingDirty *DirtyEntry
	var otherDirty []DirtyEntry
	for i := range dirtyEntries {
		if dirtyEntries[i].Key == existingKey {
			existingDirty = &dirtyEntries[i]
		} else {
			otherDirty = append(otherDirty, dirtyEntries[i])
		}
	}

	// Case 1: existing leaf is deleted
	if existingDirty != nil && existingDirty.IsDelete {
		if len(otherDirty) == 0 {
			return types.OpaqueHash{}, nil
		}
		// Build subtree from remaining inserts
		return buildSubtree(otherDirty, depth, storeNode, storeValue)
	}

	// Determine the value for the existing key
	var existingValue []byte
	if existingDirty != nil {
		// Modified
		existingValue = existingDirty.NewValue
	} else {
		// Unchanged — get original value
		var err error
		existingValue, err = reader.GetNodeValue(node)
		if err != nil {
			return types.OpaqueHash{}, fmt.Errorf("incremental: get leaf value: %w", err)
		}
	}

	// Case 2: no other dirty entries — just re-encode the (possibly modified) leaf
	if len(otherDirty) == 0 {
		if existingDirty == nil {
			// Leaf unchanged, reuse hash
			return nodeHash, nil
		}
		return encodeAndStoreLeaf(existingKey, existingValue, storeNode, storeValue)
	}

	// Case 3: leaf insert → branch split
	// Build a subtree from existing leaf + new inserts
	entries := []DirtyEntry{{Key: existingKey, NewValue: existingValue}}
	for _, e := range otherDirty {
		if !e.IsDelete {
			entries = append(entries, e)
		}
	}
	return buildSubtree(entries, depth, storeNode, storeValue)
}

func handleBranch(
	node merklization.TrieNode,
	dirtyEntries []DirtyEntry,
	depth int,
	reader NodeReader,
	storeNode merklization.StoreNodeFunc,
	storeValue merklization.StoreValueFunc,
) (types.OpaqueHash, error) {
	leftHash, rightHash := node.GetBranchHashes()

	// Partition dirty entries by bit at depth
	var leftDirty, rightDirty []DirtyEntry
	for _, e := range dirtyEntries {
		if bitAt(e.Key[:], depth) {
			rightDirty = append(rightDirty, e)
		} else {
			leftDirty = append(leftDirty, e)
		}
	}

	// Recurse left
	newLeftHash := leftHash
	if len(leftDirty) > 0 {
		var err error
		newLeftHash, err = incrementalMerklizeNode(leftHash, leftDirty, depth+1, reader, storeNode, storeValue)
		if err != nil {
			return types.OpaqueHash{}, err
		}
	}

	// Recurse right
	newRightHash := rightHash
	if len(rightDirty) > 0 {
		var err error
		newRightHash, err = incrementalMerklizeNode(rightHash, rightDirty, depth+1, reader, storeNode, storeValue)
		if err != nil {
			return types.OpaqueHash{}, err
		}
	}

	// Handle collapse: if both children are empty, subtree is gone
	if newLeftHash == (types.OpaqueHash{}) && newRightHash == (types.OpaqueHash{}) {
		return types.OpaqueHash{}, nil
	}

	// If one child is empty and the other is a leaf, collapse (promote the leaf).
	// If the other child is a branch, keep the degenerate branch (matches full merklize).
	if newLeftHash == (types.OpaqueHash{}) {
		shouldCollapse, collapseHash, err := tryCollapse(newRightHash, reader, storeNode)
		if err != nil {
			return types.OpaqueHash{}, err
		}
		if shouldCollapse {
			return collapseHash, nil
		}
	}
	if newRightHash == (types.OpaqueHash{}) {
		shouldCollapse, collapseHash, err := tryCollapse(newLeftHash, reader, storeNode)
		if err != nil {
			return types.OpaqueHash{}, err
		}
		if shouldCollapse {
			return collapseHash, nil
		}
	}

	// Encode new branch (including degenerate cases with one zero child)
	return encodeAndStoreBranch(newLeftHash, newRightHash, storeNode)
}

// tryCollapse checks if the remaining child is a leaf; if so, collapses by
// promoting the leaf. Returns (shouldCollapse, hash, error).
// Re-hashes the node because childHash from GetBranchHashes() may have byte[0]
// MSB masked (branch encoding uses left[0] & 0x7F).
func tryCollapse(childHash types.OpaqueHash, reader NodeReader, storeNode merklization.StoreNodeFunc) (bool, types.OpaqueHash, error) {
	childNode, err := reader.GetNode(childHash)
	if err != nil {
		// If we can't read the child (it was just created by buildSubtree and not
		// persisted to this reader), re-hash from the hash we have
		return false, types.OpaqueHash{}, nil
	}
	if childNode.IsLeaf() {
		// Promote the leaf — re-hash to recover the true hash
		trueHash := hash.Blake2bHash(childNode[:])
		return true, trueHash, nil
	}
	// Child is a branch — cannot collapse, will create degenerate branch
	return false, types.OpaqueHash{}, nil
}

// buildSubtree builds a fresh subtree from dirty entries using the same algorithm
// as full merklize (partition-by-bit, always creating branch nodes for len > 1).
func buildSubtree(
	entries []DirtyEntry,
	depth int,
	storeNode merklization.StoreNodeFunc,
	storeValue merklization.StoreValueFunc,
) (types.OpaqueHash, error) {
	if len(entries) == 0 {
		return types.OpaqueHash{}, nil
	}
	if len(entries) == 1 {
		return encodeAndStoreLeaf(entries[0].Key, entries[0].NewValue, storeNode, storeValue)
	}

	// Partition by bit at depth (same semantics as full merklize's partitionByBit)
	var left, right []DirtyEntry
	for _, e := range entries {
		if bitAt(e.Key[:], depth) {
			right = append(right, e)
		} else {
			left = append(left, e)
		}
	}

	leftHash, err := buildSubtree(left, depth+1, storeNode, storeValue)
	if err != nil {
		return types.OpaqueHash{}, err
	}
	rightHash, err := buildSubtree(right, depth+1, storeNode, storeValue)
	if err != nil {
		return types.OpaqueHash{}, err
	}

	// Always create a branch when len(entries) > 1, even if one side is empty.
	// This matches full merklize behavior which encodes branch(leftHash, zeroHash)
	// for degenerate cases where all entries share the same bit at this depth.
	return encodeAndStoreBranch(leftHash, rightHash, storeNode)
}

func encodeAndStoreLeaf(
	key types.StateKey,
	value []byte,
	storeNode merklization.StoreNodeFunc,
	storeValue merklization.StoreValueFunc,
) (types.OpaqueHash, error) {
	if len(value) > 32 && storeValue != nil {
		if err := storeValue(value); err != nil {
			return types.OpaqueHash{}, err
		}
	}

	nodeBytes := encodeLeafNode(key, value)
	nodeHash := hash.Blake2bHash(nodeBytes[:])

	if storeNode != nil {
		if err := storeNode(nodeHash, merklization.TrieNode(nodeBytes)); err != nil {
			return types.OpaqueHash{}, err
		}
	}
	return nodeHash, nil
}

func encodeAndStoreBranch(
	leftHash, rightHash types.OpaqueHash,
	storeNode merklization.StoreNodeFunc,
) (types.OpaqueHash, error) {
	nodeBytes := encodeBranchNode(leftHash, rightHash)
	nodeHash := hash.Blake2bHash(nodeBytes[:])

	if storeNode != nil {
		if err := storeNode(nodeHash, merklization.TrieNode(nodeBytes)); err != nil {
			return types.OpaqueHash{}, err
		}
	}
	return nodeHash, nil
}

// encodeLeafNode encodes a leaf node as [64]byte.
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

// encodeBranchNode encodes a branch node as [64]byte.
func encodeBranchNode(left, right types.OpaqueHash) [64]byte {
	var node [64]byte
	node[0] = left[0] & 0x7F
	copy(node[1:32], left[1:])
	copy(node[32:], right[:])
	return node
}

func bitAt(key []byte, depth int) bool {
	byteIdx := depth / 8
	bitMask := byte(1 << (7 - depth%8))
	return key[byteIdx]&bitMask != 0
}
