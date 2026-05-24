package store_test

import (
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/database/provider/memory"
	"github.com/New-JAMneration/JAM-Protocol/internal/store"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/merklization"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIncrementalMerklize_NoDirtyEntries verifies that no changes = same root.
func TestIncrementalMerklize_NoDirtyEntries(t *testing.T) {
	tr := store.NewTrie(memory.NewDatabase())
	pairs := types.StateKeyVals{
		{Key: types.StateKey{0x00}, Value: []byte{1, 2, 3}},
		{Key: types.StateKey{0x80}, Value: []byte{4, 5, 6}},
	}
	root, err := tr.MerklizeAndCommit(pairs)
	require.NoError(t, err)

	newRoot, err := store.IncrementalMerklize(
		types.OpaqueHash(root), nil, tr, nil, nil,
	)
	require.NoError(t, err)
	assert.Equal(t, types.OpaqueHash(root), newRoot)
}

// TestIncrementalMerklize_ModifyLeaf modifies a single leaf value.
func TestIncrementalMerklize_ModifyLeaf(t *testing.T) {
	tr := store.NewTrie(memory.NewDatabase())
	pairs := types.StateKeyVals{
		{Key: types.StateKey{0x00}, Value: []byte{1, 2, 3}},
		{Key: types.StateKey{0x80}, Value: []byte{4, 5, 6}},
	}
	root, err := tr.MerklizeAndCommit(pairs)
	require.NoError(t, err)

	// Modify second entry
	dirty := []store.DirtyEntry{
		{Key: types.StateKey{0x80}, NewValue: []byte{7, 8, 9}},
	}

	var nodes []struct {
		Hash types.OpaqueHash
		Node merklization.TrieNode
	}
	storeNode := func(h types.OpaqueHash, n merklization.TrieNode) error {
		nodes = append(nodes, struct {
			Hash types.OpaqueHash
			Node merklization.TrieNode
		}{h, n})
		return nil
	}

	newRoot, err := store.IncrementalMerklize(
		types.OpaqueHash(root), dirty, tr, storeNode, nil,
	)
	require.NoError(t, err)
	assert.NotEqual(t, types.OpaqueHash(root), newRoot)

	// Verify against full merklize with new value
	expectedPairs := types.StateKeyVals{
		{Key: types.StateKey{0x00}, Value: []byte{1, 2, 3}},
		{Key: types.StateKey{0x80}, Value: []byte{7, 8, 9}},
	}
	expectedRoot, err := tr.MerklizeOnly(expectedPairs)
	require.NoError(t, err)
	assert.Equal(t, types.OpaqueHash(expectedRoot), newRoot)
}

// TestIncrementalMerklize_InsertLeaf inserts a new leaf (branch split).
func TestIncrementalMerklize_InsertLeaf(t *testing.T) {
	tr := store.NewTrie(memory.NewDatabase())
	pairs := types.StateKeyVals{
		{Key: types.StateKey{0x00}, Value: []byte{1, 2, 3}},
	}
	root, err := tr.MerklizeAndCommit(pairs)
	require.NoError(t, err)

	// Insert a new key
	dirty := []store.DirtyEntry{
		{Key: types.StateKey{0x80}, NewValue: []byte{4, 5, 6}},
	}

	newRoot, err := store.IncrementalMerklize(
		types.OpaqueHash(root), dirty, tr, nil, nil,
	)
	require.NoError(t, err)

	expectedPairs := types.StateKeyVals{
		{Key: types.StateKey{0x00}, Value: []byte{1, 2, 3}},
		{Key: types.StateKey{0x80}, Value: []byte{4, 5, 6}},
	}
	expectedRoot, err := tr.MerklizeOnly(expectedPairs)
	require.NoError(t, err)
	assert.Equal(t, types.OpaqueHash(expectedRoot), newRoot)
}

// TestIncrementalMerklize_DeleteLeaf_BranchCollapse deletes a leaf causing branch collapse.
func TestIncrementalMerklize_DeleteLeaf_BranchCollapse(t *testing.T) {
	tr := store.NewTrie(memory.NewDatabase())
	pairs := types.StateKeyVals{
		{Key: types.StateKey{0x00}, Value: []byte{1, 2, 3}},
		{Key: types.StateKey{0x80}, Value: []byte{4, 5, 6}},
	}
	root, err := tr.MerklizeAndCommit(pairs)
	require.NoError(t, err)

	// Delete the second entry → branch collapses to single leaf
	dirty := []store.DirtyEntry{
		{Key: types.StateKey{0x80}, IsDelete: true},
	}

	newRoot, err := store.IncrementalMerklize(
		types.OpaqueHash(root), dirty, tr, nil, nil,
	)
	require.NoError(t, err)

	expectedPairs := types.StateKeyVals{
		{Key: types.StateKey{0x00}, Value: []byte{1, 2, 3}},
	}
	expectedRoot, err := tr.MerklizeOnly(expectedPairs)
	require.NoError(t, err)
	assert.Equal(t, types.OpaqueHash(expectedRoot), newRoot)
}

// TestIncrementalMerklize_DeleteAllLeaves deletes all entries → empty trie.
func TestIncrementalMerklize_DeleteAllLeaves(t *testing.T) {
	tr := store.NewTrie(memory.NewDatabase())
	pairs := types.StateKeyVals{
		{Key: types.StateKey{0x00}, Value: []byte{1, 2, 3}},
		{Key: types.StateKey{0x80}, Value: []byte{4, 5, 6}},
	}
	root, err := tr.MerklizeAndCommit(pairs)
	require.NoError(t, err)

	dirty := []store.DirtyEntry{
		{Key: types.StateKey{0x00}, IsDelete: true},
		{Key: types.StateKey{0x80}, IsDelete: true},
	}

	newRoot, err := store.IncrementalMerklize(
		types.OpaqueHash(root), dirty, tr, nil, nil,
	)
	require.NoError(t, err)
	assert.Equal(t, types.OpaqueHash{}, newRoot)
}

// TestIncrementalMerklize_MultipleChanges tests mixed insert/delete/modify.
func TestIncrementalMerklize_MultipleChanges(t *testing.T) {
	tr := store.NewTrie(memory.NewDatabase())
	pairs := types.StateKeyVals{
		{Key: types.StateKey{0x10}, Value: []byte{0xAA}},
		{Key: types.StateKey{0x20}, Value: []byte{0xBB}},
		{Key: types.StateKey{0x80}, Value: []byte{0xCC}},
		{Key: types.StateKey{0xC0}, Value: []byte{0xDD}},
	}
	root, err := tr.MerklizeAndCommit(pairs)
	require.NoError(t, err)

	dirty := []store.DirtyEntry{
		{Key: types.StateKey{0x20}, IsDelete: true},         // delete
		{Key: types.StateKey{0x40}, NewValue: []byte{0xEE}}, // insert
		{Key: types.StateKey{0xC0}, NewValue: []byte{0xFF}}, // modify
	}

	newRoot, err := store.IncrementalMerklize(
		types.OpaqueHash(root), dirty, tr, nil, nil,
	)
	require.NoError(t, err)

	expectedPairs := types.StateKeyVals{
		{Key: types.StateKey{0x10}, Value: []byte{0xAA}},
		{Key: types.StateKey{0x40}, Value: []byte{0xEE}},
		{Key: types.StateKey{0x80}, Value: []byte{0xCC}},
		{Key: types.StateKey{0xC0}, Value: []byte{0xFF}},
	}
	expectedRoot, err := tr.MerklizeOnly(expectedPairs)
	require.NoError(t, err)
	assert.Equal(t, types.OpaqueHash(expectedRoot), newRoot)
}

// TestIncrementalMerklize_LargeValue tests with values > 32 bytes (regular leaf).
func TestIncrementalMerklize_LargeValue(t *testing.T) {
	tr := store.NewTrie(memory.NewDatabase())
	largeVal := make([]byte, 64)
	for i := range largeVal {
		largeVal[i] = byte(i)
	}
	pairs := types.StateKeyVals{
		{Key: types.StateKey{0x00}, Value: largeVal},
	}
	root, err := tr.MerklizeAndCommit(pairs)
	require.NoError(t, err)

	newLargeVal := make([]byte, 64)
	for i := range newLargeVal {
		newLargeVal[i] = byte(i + 1)
	}
	dirty := []store.DirtyEntry{
		{Key: types.StateKey{0x00}, NewValue: newLargeVal},
	}

	var storedValues [][]byte
	storeValue := func(v []byte) error {
		cp := make([]byte, len(v))
		copy(cp, v)
		storedValues = append(storedValues, cp)
		return nil
	}

	newRoot, err := store.IncrementalMerklize(
		types.OpaqueHash(root), dirty, tr, nil, storeValue,
	)
	require.NoError(t, err)

	expectedPairs := types.StateKeyVals{
		{Key: types.StateKey{0x00}, Value: newLargeVal},
	}
	expectedRoot, err := tr.MerklizeOnly(expectedPairs)
	require.NoError(t, err)
	assert.Equal(t, types.OpaqueHash(expectedRoot), newRoot)
	assert.Len(t, storedValues, 1)
	assert.Equal(t, newLargeVal, storedValues[0])
}
