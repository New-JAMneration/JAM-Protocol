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

func newTestTrie() *store.Trie {
	return store.NewTrie(memory.NewDatabase())
}

func makePairs(n int) types.StateKeyVals {
	pairs := make(types.StateKeyVals, n)
	for i := range pairs {
		var key types.StateKey
		key[0] = byte(i >> 8)
		key[1] = byte(i)
		pairs[i] = types.StateKeyVal{
			Key:   key,
			Value: []byte{byte(i), byte(i + 1), byte(i + 2), byte(i + 3)},
		}
	}
	return pairs
}

func makeLargeValuePairs(n int) types.StateKeyVals {
	pairs := make(types.StateKeyVals, n)
	for i := range pairs {
		var key types.StateKey
		key[0] = byte(i >> 8)
		key[1] = byte(i)
		value := make([]byte, 64)
		for j := range value {
			value[j] = byte(i + j)
		}
		pairs[i] = types.StateKeyVal{Key: key, Value: value}
	}
	return pairs
}

func TestMerklizeAndCommit_SingleEntry(t *testing.T) {
	tr := newTestTrie()
	pairs := makePairs(1)

	root, err := tr.MerklizeAndCommit(pairs)
	require.NoError(t, err)
	assert.NotEqual(t, types.StateRoot{}, root)

	exists, err := tr.TrieExists(types.OpaqueHash(root))
	require.NoError(t, err)
	assert.True(t, exists)
}

func TestMerklizeAndCommit_MultipleEntries(t *testing.T) {
	tr := newTestTrie()
	pairs := makePairs(10)

	root, err := tr.MerklizeAndCommit(pairs)
	require.NoError(t, err)
	assert.NotEqual(t, types.StateRoot{}, root)

	exists, err := tr.TrieExists(types.OpaqueHash(root))
	require.NoError(t, err)
	assert.True(t, exists)
}

func TestMerklizeAndCommit_EmptyInput(t *testing.T) {
	tr := newTestTrie()

	root, err := tr.MerklizeAndCommit(types.StateKeyVals{})
	require.NoError(t, err)
	assert.Equal(t, types.StateRoot{}, root)
}

func TestMerklizeOnly_MatchesMerklizeAndCommit(t *testing.T) {
	tr := newTestTrie()
	pairs := makePairs(10)

	rootCommit, err := tr.MerklizeAndCommit(pairs)
	require.NoError(t, err)

	rootOnly, err := tr.MerklizeOnly(pairs)
	require.NoError(t, err)

	assert.Equal(t, rootCommit, rootOnly)
}

func TestGetNode_Found(t *testing.T) {
	tr := newTestTrie()
	pairs := makePairs(1)

	root, err := tr.MerklizeAndCommit(pairs)
	require.NoError(t, err)

	node, err := tr.GetNode(types.OpaqueHash(root))
	require.NoError(t, err)
	assert.True(t, node.IsLeaf())
}

func TestGetNode_NotFound(t *testing.T) {
	tr := newTestTrie()
	fakeHash := types.OpaqueHash{0xFF, 0xAA, 0xBB}

	_, err := tr.GetNode(fakeHash)
	assert.ErrorIs(t, err, store.ErrNotFound)
}

func TestGetNodeValue_EmbeddedLeaf(t *testing.T) {
	tr := newTestTrie()
	pairs := makePairs(1)

	root, err := tr.MerklizeAndCommit(pairs)
	require.NoError(t, err)

	node, err := tr.GetNode(types.OpaqueHash(root))
	require.NoError(t, err)
	require.True(t, node.IsLeaf())
	require.True(t, node.IsEmbeddedLeaf())

	value, err := tr.GetNodeValue(node)
	require.NoError(t, err)
	assert.Equal(t, []byte(pairs[0].Value), value)
}

func TestGetNodeValue_RegularLeaf(t *testing.T) {
	tr := newTestTrie()
	pairs := makeLargeValuePairs(1)

	root, err := tr.MerklizeAndCommit(pairs)
	require.NoError(t, err)

	node, err := tr.GetNode(types.OpaqueHash(root))
	require.NoError(t, err)
	require.True(t, node.IsLeaf())
	require.False(t, node.IsEmbeddedLeaf())

	value, err := tr.GetNodeValue(node)
	require.NoError(t, err)
	assert.Equal(t, []byte(pairs[0].Value), value)
}

func TestGetNodeValue_BranchNodeReturnsError(t *testing.T) {
	tr := newTestTrie()
	pairs := makePairs(4)

	root, err := tr.MerklizeAndCommit(pairs)
	require.NoError(t, err)

	node, err := tr.GetNode(types.OpaqueHash(root))
	require.NoError(t, err)
	require.True(t, node.IsBranch())

	_, err = tr.GetNodeValue(node)
	assert.ErrorIs(t, err, store.ErrNotLeafNode)
}

func TestTrieExists_NotFound(t *testing.T) {
	tr := newTestTrie()

	exists, err := tr.TrieExists(types.OpaqueHash{1, 2, 3})
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestRefCount_Basic(t *testing.T) {
	tr := newTestTrie()
	pairs := makePairs(1)

	root, err := tr.MerklizeAndCommit(pairs)
	require.NoError(t, err)

	count, err := tr.GetNodeRefCount(types.OpaqueHash(root))
	require.NoError(t, err)
	assert.Equal(t, uint64(1), count)
}

func TestRefCount_IncreaseTwice(t *testing.T) {
	tr := newTestTrie()
	pairs := makePairs(1)

	root, err := tr.MerklizeAndCommit(pairs)
	require.NoError(t, err)

	err = tr.IncreaseNodeRefCount(types.OpaqueHash(root))
	require.NoError(t, err)

	count, err := tr.GetNodeRefCount(types.OpaqueHash(root))
	require.NoError(t, err)
	assert.Equal(t, uint64(2), count)
}

func TestRefCount_Decrease(t *testing.T) {
	tr := newTestTrie()
	pairs := makePairs(1)

	root, err := tr.MerklizeAndCommit(pairs)
	require.NoError(t, err)

	err = tr.IncreaseNodeRefCount(types.OpaqueHash(root))
	require.NoError(t, err)

	newCount, err := tr.DecreaseNodeRefCount(types.OpaqueHash(root))
	require.NoError(t, err)
	assert.Equal(t, uint64(1), newCount)
}

func TestRefCount_DecreaseToZero(t *testing.T) {
	tr := newTestTrie()
	pairs := makePairs(1)

	root, err := tr.MerklizeAndCommit(pairs)
	require.NoError(t, err)

	newCount, err := tr.DecreaseNodeRefCount(types.OpaqueHash(root))
	require.NoError(t, err)
	assert.Equal(t, uint64(0), newCount)
}

func TestRefCount_DecreaseNotFound(t *testing.T) {
	tr := newTestTrie()
	_, err := tr.DecreaseNodeRefCount(types.OpaqueHash{0xAA})
	assert.Error(t, err)
}

func TestRefCount_GetNotFound(t *testing.T) {
	tr := newTestTrie()
	_, err := tr.GetNodeRefCount(types.OpaqueHash{0xAA})
	assert.ErrorIs(t, err, store.ErrNotFound)
}

func TestDeleteTrie_SingleLeaf(t *testing.T) {
	tr := newTestTrie()
	pairs := makePairs(1)

	root, err := tr.MerklizeAndCommit(pairs)
	require.NoError(t, err)

	err = tr.DeleteTrie(types.OpaqueHash(root))
	require.NoError(t, err)

	exists, err := tr.TrieExists(types.OpaqueHash(root))
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestDeleteTrie_BranchTree(t *testing.T) {
	tr := newTestTrie()
	pairs := makePairs(8)

	root, err := tr.MerklizeAndCommit(pairs)
	require.NoError(t, err)

	err = tr.DeleteTrie(types.OpaqueHash(root))
	require.NoError(t, err)

	exists, err := tr.TrieExists(types.OpaqueHash(root))
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestDeleteTrie_NonExistentRoot(t *testing.T) {
	tr := newTestTrie()
	err := tr.DeleteTrie(types.OpaqueHash{0xDE, 0xAD})
	require.NoError(t, err)
}

func TestDeleteTrie_RegularLeafValueDeleted(t *testing.T) {
	tr := newTestTrie()
	pairs := makeLargeValuePairs(1)

	root, err := tr.MerklizeAndCommit(pairs)
	require.NoError(t, err)

	node, err := tr.GetNode(types.OpaqueHash(root))
	require.NoError(t, err)

	err = tr.DeleteTrie(types.OpaqueHash(root))
	require.NoError(t, err)

	_, err = tr.GetNodeValue(node)
	assert.ErrorIs(t, err, store.ErrNotFound)
}

func TestSharedNodes_RefCountCorrect(t *testing.T) {
	tr := newTestTrie()

	pairs1 := types.StateKeyVals{
		{Key: types.StateKey{0x00}, Value: []byte{1, 2, 3}},
		{Key: types.StateKey{0x80}, Value: []byte{4, 5, 6}},
	}
	pairs2 := types.StateKeyVals{
		{Key: types.StateKey{0x00}, Value: []byte{1, 2, 3}},
		{Key: types.StateKey{0x80}, Value: []byte{7, 8, 9}},
	}

	root1, err := tr.MerklizeAndCommit(pairs1)
	require.NoError(t, err)

	root2, err := tr.MerklizeAndCommit(pairs2)
	require.NoError(t, err)

	assert.NotEqual(t, root1, root2, "different values should produce different roots")

	// The left child (key=0x00, value=[1,2,3]) is shared between both tries.
	// Its refcount should be 2 after two MerklizeAndCommit calls.
	node1, err := tr.GetNode(types.OpaqueHash(root1))
	require.NoError(t, err)
	require.True(t, node1.IsBranch())

	leftHash, _ := node1.GetBranchHashes()
	count, err := tr.GetNodeRefCount(leftHash)
	require.NoError(t, err)
	assert.Equal(t, uint64(2), count, "shared leaf should have refcount=2")
}

func TestSharedRootGuard(t *testing.T) {
	tr := newTestTrie()
	pairs := makePairs(4)

	root1, err := tr.MerklizeAndCommit(pairs)
	require.NoError(t, err)

	// Commit the same pairs again — produces the same root hash.
	root2, err := tr.MerklizeAndCommit(pairs)
	require.NoError(t, err)
	require.Equal(t, root1, root2)

	rootHash := types.OpaqueHash(root1)
	count, err := tr.GetNodeRefCount(rootHash)
	require.NoError(t, err)
	assert.Equal(t, uint64(2), count, "same root committed twice should have refcount=2")

	// Delete one trie — root is force-deleted, but refcount was 2.
	err = tr.DeleteTrie(rootHash)
	require.NoError(t, err)

	// After DeleteTrie with forceDelete=true, the root node itself is removed.
	// This is the known behaviour documented in Todo.md (shared root guard issue).
	// The second trie's root becomes a dangling reference.
	exists, err := tr.TrieExists(rootHash)
	require.NoError(t, err)
	assert.False(t, exists, "forceDelete=true removes root even if refcount > 0")
}

func TestMerklizeAndCommit_Deterministic(t *testing.T) {
	pairs := makePairs(16)

	tr1 := newTestTrie()
	root1, err := tr1.MerklizeAndCommit(pairs)
	require.NoError(t, err)

	tr2 := newTestTrie()
	root2, err := tr2.MerklizeAndCommit(pairs)
	require.NoError(t, err)

	assert.Equal(t, root1, root2, "same input must produce same root")
}

func TestMerklizeOnly_DoesNotPersist(t *testing.T) {
	tr := newTestTrie()
	pairs := makePairs(4)

	root, err := tr.MerklizeOnly(pairs)
	require.NoError(t, err)
	assert.NotEqual(t, types.StateRoot{}, root)

	exists, err := tr.TrieExists(types.OpaqueHash(root))
	require.NoError(t, err)
	assert.False(t, exists, "MerklizeOnly should not persist nodes")
}

// TestCallbackConsistency verifies that merklize and merklizeWithCache produce
// the same state root and the same set of (hash, node) pairs via callbacks.
func TestCallbackConsistency(t *testing.T) {
	pairs := makePairs(16)

	type nodeEntry struct {
		Hash types.OpaqueHash
		Node merklization.TrieNode
	}

	var nodesPlain []nodeEntry
	var valuesPlain [][]byte

	storeNodePlain := func(h types.OpaqueHash, n merklization.TrieNode) error {
		nodesPlain = append(nodesPlain, nodeEntry{h, n})
		return nil
	}
	storeValuePlain := func(v []byte) error {
		cp := make([]byte, len(v))
		copy(cp, v)
		valuesPlain = append(valuesPlain, cp)
		return nil
	}

	entriesPlain := pairs.DeepCopy()
	rootPlain, err := merklization.MerklizationSerializedStateWithCache(
		entriesPlain, nil, storeNodePlain, storeValuePlain,
	)
	require.NoError(t, err)

	var nodesCached []nodeEntry
	var valuesCached [][]byte

	storeNodeCached := func(h types.OpaqueHash, n merklization.TrieNode) error {
		nodesCached = append(nodesCached, nodeEntry{h, n})
		return nil
	}
	storeValueCached := func(v []byte) error {
		cp := make([]byte, len(v))
		copy(cp, v)
		valuesCached = append(valuesCached, cp)
		return nil
	}

	// Use a trivial cache that always computes (no actual caching)
	trivialCache := func(key types.StateKey, value []byte) types.OpaqueHash {
		return merklization.EncodeLeafNodeHash(key, value)
	}

	entriesCached := pairs.DeepCopy()
	rootCached, err := merklization.MerklizationSerializedStateWithCache(
		entriesCached, trivialCache, storeNodeCached, storeValueCached,
	)
	require.NoError(t, err)

	// (a) Same state root
	assert.Equal(t, rootPlain, rootCached, "plain and cached must produce same root")

	// (b) Same set of (hash, node) pairs (order may differ due to partition, but with
	//     same input both should traverse identically)
	require.Equal(t, len(nodesPlain), len(nodesCached), "same number of nodes")
	plainSet := make(map[types.OpaqueHash]merklization.TrieNode)
	for _, e := range nodesPlain {
		plainSet[e.Hash] = e.Node
	}
	for _, e := range nodesCached {
		node, exists := plainSet[e.Hash]
		assert.True(t, exists, "cached node %x not in plain set", e.Hash)
		if exists {
			assert.Equal(t, node, e.Node)
		}
	}

	// (c) Same set of stored values
	assert.Equal(t, len(valuesPlain), len(valuesCached))
}

// TestCallbackConsistency_LargeValues tests with values > 32 bytes that trigger storeValue.
func TestCallbackConsistency_LargeValues(t *testing.T) {
	pairs := makeLargeValuePairs(8)

	type nodeEntry struct {
		Hash types.OpaqueHash
		Node merklization.TrieNode
	}

	var nodesPlain []nodeEntry
	var valuesPlain [][]byte

	storeNodePlain := func(h types.OpaqueHash, n merklization.TrieNode) error {
		nodesPlain = append(nodesPlain, nodeEntry{h, n})
		return nil
	}
	storeValuePlain := func(v []byte) error {
		cp := make([]byte, len(v))
		copy(cp, v)
		valuesPlain = append(valuesPlain, cp)
		return nil
	}

	entriesPlain := pairs.DeepCopy()
	rootPlain, err := merklization.MerklizationSerializedStateWithCache(
		entriesPlain, nil, storeNodePlain, storeValuePlain,
	)
	require.NoError(t, err)

	var nodesCached []nodeEntry
	var valuesCached [][]byte

	storeNodeCached := func(h types.OpaqueHash, n merklization.TrieNode) error {
		nodesCached = append(nodesCached, nodeEntry{h, n})
		return nil
	}
	storeValueCached := func(v []byte) error {
		cp := make([]byte, len(v))
		copy(cp, v)
		valuesCached = append(valuesCached, cp)
		return nil
	}

	trivialCache := func(key types.StateKey, value []byte) types.OpaqueHash {
		return merklization.EncodeLeafNodeHash(key, value)
	}

	entriesCached := pairs.DeepCopy()
	rootCached, err := merklization.MerklizationSerializedStateWithCache(
		entriesCached, trivialCache, storeNodeCached, storeValueCached,
	)
	require.NoError(t, err)

	assert.Equal(t, rootPlain, rootCached)
	require.Equal(t, len(nodesPlain), len(nodesCached))

	plainSet := make(map[types.OpaqueHash]merklization.TrieNode)
	for _, e := range nodesPlain {
		plainSet[e.Hash] = e.Node
	}
	for _, e := range nodesCached {
		node, exists := plainSet[e.Hash]
		assert.True(t, exists, "cached node %x not in plain set", e.Hash)
		if exists {
			assert.Equal(t, node, e.Node)
		}
	}

	require.Equal(t, len(valuesPlain), len(valuesCached))
	for i := range valuesPlain {
		assert.Equal(t, valuesPlain[i], valuesCached[i])
	}
}
