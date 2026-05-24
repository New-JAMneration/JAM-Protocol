package store

import (
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/New-JAMneration/JAM-Protocol/internal/database"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/merklization"
)

const (
	prefixTrieNode         byte = 0x03
	prefixTrieNodeValue    byte = 0x04
	prefixTrieNodeRefCount byte = 0x05
)

var (
	ErrNotFound     = errors.New("trie: not found")
	ErrNotLeafNode  = errors.New("trie: not a leaf node")
	ErrRefCountZero = errors.New("trie: ref count already zero")
)

// Trie manages trie node persistence and reference counting in a key-value database.
type Trie struct {
	db database.Database
}

// NewTrie creates a new Trie backed by the given database.
func NewTrie(db database.Database) *Trie {
	return &Trie{db: db}
}

func makeTrieKey(prefix byte, suffix []byte) []byte {
	key := make([]byte, 1+len(suffix))
	key[0] = prefix
	copy(key[1:], suffix)
	return key
}

// MerklizeAndCommit computes the state root and persists all trie nodes + leaf values
// into the database, then increments refcount for each new node.
func (t *Trie) MerklizeAndCommit(pairs types.StateKeyVals) (types.StateRoot, error) {
	batch := t.db.NewBatch()

	var newNodes []types.OpaqueHash

	storeNode := func(nodeHash types.OpaqueHash, node merklization.TrieNode) error {
		newNodes = append(newNodes, nodeHash)
		return batch.Put(makeTrieKey(prefixTrieNode, nodeHash[1:]), node[:])
	}

	storeValue := func(value []byte) error {
		valueHash := hash.Blake2bHash(value)
		return batch.Put(makeTrieKey(prefixTrieNodeValue, valueHash[:]), value)
	}

	root, err := merklization.MerklizationSerializedStateWithCache(pairs, nil, storeNode, storeValue)
	if err != nil {
		batch.Close()
		return types.StateRoot{}, fmt.Errorf("trie: merklize error: %w", err)
	}

	if err := batch.Commit(); err != nil {
		batch.Close()
		return types.StateRoot{}, fmt.Errorf("trie: batch commit error: %w", err)
	}

	if err := batch.Close(); err != nil {
		return types.StateRoot{}, fmt.Errorf("trie: batch close error: %w", err)
	}

	for _, nodeHash := range newNodes {
		if err := t.IncreaseNodeRefCount(nodeHash); err != nil {
			return types.StateRoot{}, fmt.Errorf("trie: increase ref count for %x: %w", nodeHash, err)
		}
	}

	return root, nil
}

// MerklizeOnly computes the state root without persisting anything (dry-run).
func (t *Trie) MerklizeOnly(pairs types.StateKeyVals) (types.StateRoot, error) {
	return merklization.MerklizationSerializedState(pairs)
}

// GetNode retrieves a trie node by its hash.
func (t *Trie) GetNode(nodeHash types.OpaqueHash) (merklization.TrieNode, error) {
	data, found, err := t.db.Get(makeTrieKey(prefixTrieNode, nodeHash[1:]))
	if err != nil {
		return merklization.TrieNode{}, fmt.Errorf("trie: get node %x: %w", nodeHash, err)
	}
	if !found {
		return merklization.TrieNode{}, ErrNotFound
	}
	var node merklization.TrieNode
	copy(node[:], data)
	return node, nil
}

// GetNodeValue retrieves the value from a leaf node.
// For embedded leaves the value is extracted directly; for regular leaves
// the value is fetched from the database using the value hash.
func (t *Trie) GetNodeValue(node merklization.TrieNode) ([]byte, error) {
	if !node.IsLeaf() {
		return nil, ErrNotLeafNode
	}

	if node.IsEmbeddedLeaf() {
		return node.GetLeafValue(), nil
	}

	valueHash := node.GetLeafValueHash()
	data, found, err := t.db.Get(makeTrieKey(prefixTrieNodeValue, valueHash[:]))
	if err != nil {
		return nil, fmt.Errorf("trie: get value %x: %w", valueHash, err)
	}
	if !found {
		return nil, ErrNotFound
	}
	return data, nil
}

// TrieExists checks whether a trie with the given root hash exists in the database.
func (t *Trie) TrieExists(rootHash types.OpaqueHash) (bool, error) {
	_, err := t.GetNode(rootHash)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// IncreaseNodeRefCount increments the reference count for a node.
// If the node has no existing refcount entry, it is initialized to 1.
func (t *Trie) IncreaseNodeRefCount(nodeHash types.OpaqueHash) error {
	key := makeTrieKey(prefixTrieNodeRefCount, nodeHash[1:])
	data, found, err := t.db.Get(key)
	if err != nil {
		return fmt.Errorf("trie: get ref count %x: %w", nodeHash, err)
	}

	var newCount uint64
	if found {
		newCount = binary.LittleEndian.Uint64(data) + 1
	} else {
		newCount = 1
	}

	countBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(countBytes, newCount)
	return t.db.Put(key, countBytes)
}

// DecreaseNodeRefCount decrements the reference count for a node.
// Returns the new count after decrement.
func (t *Trie) DecreaseNodeRefCount(nodeHash types.OpaqueHash) (uint64, error) {
	key := makeTrieKey(prefixTrieNodeRefCount, nodeHash[1:])
	data, found, err := t.db.Get(key)
	if err != nil {
		return 0, fmt.Errorf("trie: get ref count %x: %w", nodeHash, err)
	}
	if !found {
		return 0, fmt.Errorf("trie: ref count not found for %x", nodeHash)
	}

	current := binary.LittleEndian.Uint64(data)
	if current == 0 {
		return 0, fmt.Errorf("trie: %w for %x", ErrRefCountZero, nodeHash)
	}

	newCount := current - 1
	countBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(countBytes, newCount)

	if err := t.db.Put(key, countBytes); err != nil {
		return 0, fmt.Errorf("trie: put ref count %x: %w", nodeHash, err)
	}
	return newCount, nil
}

// GetNodeRefCount returns the current reference count for a node.
func (t *Trie) GetNodeRefCount(nodeHash types.OpaqueHash) (uint64, error) {
	key := makeTrieKey(prefixTrieNodeRefCount, nodeHash[1:])
	data, found, err := t.db.Get(key)
	if err != nil {
		return 0, fmt.Errorf("trie: get ref count %x: %w", nodeHash, err)
	}
	if !found {
		return 0, ErrNotFound
	}
	return binary.LittleEndian.Uint64(data), nil
}

// DeleteTrie recursively deletes a trie starting from the root hash.
// The root is force-deleted regardless of refcount; children are only deleted
// when their refcount reaches zero.
func (t *Trie) DeleteTrie(rootHash types.OpaqueHash) error {
	return t.deleteNode(rootHash, true)
}

func (t *Trie) deleteNode(nodeHash types.OpaqueHash, forceDelete bool) error {
	node, err := t.GetNode(nodeHash)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil
		}
		return fmt.Errorf("trie: delete get node %x: %w", nodeHash, err)
	}

	newCount, err := t.DecreaseNodeRefCount(nodeHash)
	if err != nil {
		return fmt.Errorf("trie: delete decrease ref %x: %w", nodeHash, err)
	}

	if newCount > 0 && !forceDelete {
		return nil
	}

	if node.IsBranch() {
		leftHash, rightHash := node.GetBranchHashes()

		if leftHash != (types.OpaqueHash{}) {
			if err := t.deleteNode(leftHash, false); err != nil {
				return fmt.Errorf("trie: delete left child: %w", err)
			}
		}
		if rightHash != (types.OpaqueHash{}) {
			if err := t.deleteNode(rightHash, false); err != nil {
				return fmt.Errorf("trie: delete right child: %w", err)
			}
		}
	} else if node.IsLeaf() && !node.IsEmbeddedLeaf() {
		valueHash := node.GetLeafValueHash()
		if err := t.db.Delete(makeTrieKey(prefixTrieNodeValue, valueHash[:])); err != nil {
			return fmt.Errorf("trie: delete value %x: %w", valueHash, err)
		}
	}

	if err := t.db.Delete(makeTrieKey(prefixTrieNode, nodeHash[1:])); err != nil {
		return fmt.Errorf("trie: delete node %x: %w", nodeHash, err)
	}
	if err := t.db.Delete(makeTrieKey(prefixTrieNodeRefCount, nodeHash[1:])); err != nil {
		return fmt.Errorf("trie: delete ref count %x: %w", nodeHash, err)
	}

	return nil
}

// DeleteAll removes all trie data (nodes, values, refcounts) from the database.
// Uses Iterator+Batch+Delete since Database interface lacks DeleteRange.
func (t *Trie) DeleteAll() error {
	prefixes := []byte{prefixTrieNode, prefixTrieNodeValue, prefixTrieNodeRefCount}
	for _, prefix := range prefixes {
		if err := t.deleteByPrefix([]byte{prefix}); err != nil {
			return fmt.Errorf("trie: delete all prefix 0x%02x: %w", prefix, err)
		}
	}
	return nil
}

func (t *Trie) deleteByPrefix(prefix []byte) error {
	iter, err := t.db.NewIterator(prefix, nil)
	if err != nil {
		return err
	}
	defer iter.Close()

	batch := t.db.NewBatch()
	count := 0
	for iter.Next() {
		if err := batch.Delete(iter.Key()); err != nil {
			batch.Close()
			return err
		}
		count++
		if count%1000 == 0 {
			if err := batch.Commit(); err != nil {
				batch.Close()
				return err
			}
			batch.Close()
			batch = t.db.NewBatch()
		}
	}
	if err := iter.Error(); err != nil {
		batch.Close()
		return err
	}
	if count%1000 != 0 {
		if err := batch.Commit(); err != nil {
			batch.Close()
			return err
		}
	}
	batch.Close()
	return nil
}

// TrieNodePrefix returns the trie node prefix byte (for Iterator/DeleteRange usage).
func TrieNodePrefix() byte { return prefixTrieNode }

// TrieNodeValuePrefix returns the trie node value prefix byte.
func TrieNodeValuePrefix() byte { return prefixTrieNodeValue }

// TrieNodeRefCountPrefix returns the trie node refcount prefix byte.
func TrieNodeRefCountPrefix() byte { return prefixTrieNodeRefCount }
