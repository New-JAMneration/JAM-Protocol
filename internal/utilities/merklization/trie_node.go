package merklization

import "github.com/New-JAMneration/JAM-Protocol/internal/types"

// TrieNode is the 64-byte on-wire representation of a trie node.
// Branch: {left[0]&0x7F, left[1:32], right[0:32]}
// Embedded leaf (value <= 32B): {0x80|len, key[:31], value, padding}
// Regular leaf (value > 32B): {0xC0, key[:31], blake2b(value)}
type TrieNode [64]byte

// IsBranch returns true if this is a branch node (MSB of byte 0 is 0).
func (n TrieNode) IsBranch() bool { return n[0]&0x80 == 0 }

// IsLeaf returns true if this is a leaf node (MSB of byte 0 is 1).
func (n TrieNode) IsLeaf() bool { return n[0]&0x80 != 0 }

// IsEmbeddedLeaf returns true if this is an embedded leaf (value <= 32B).
func (n TrieNode) IsEmbeddedLeaf() bool { return n.IsLeaf() && n[0]&0x40 == 0 }

// GetBranchHashes returns left and right child hashes from a branch node.
// Note: left[0] has MSB cleared (& 0x7F from encoding). This does NOT affect
// DB lookup (key uses hash[1:32]), but left[0] != original child hash byte 0.
func (n TrieNode) GetBranchHashes() (left, right types.OpaqueHash) {
	copy(left[:], n[:32])
	copy(right[:], n[32:])
	return
}

// GetLeafKey returns the 31-byte StateKey stored in a leaf node.
func (n TrieNode) GetLeafKey() types.StateKey {
	var key types.StateKey
	copy(key[:], n[1:32])
	return key
}

// GetLeafValue returns the embedded value from an embedded leaf node.
// For regular leaves, use GetLeafValueHash instead.
func (n TrieNode) GetLeafValue() []byte {
	if !n.IsEmbeddedLeaf() {
		return nil
	}
	length := int(n[0] & 0x3F)
	if length > 32 {
		return nil
	}
	val := make([]byte, length)
	copy(val, n[32:32+length])
	return val
}

// GetLeafValueHash returns the blake2b hash of the value from a regular leaf.
func (n TrieNode) GetLeafValueHash() types.OpaqueHash {
	var h types.OpaqueHash
	copy(h[:], n[32:])
	return h
}

// StoreNodeFunc is called for each trie node (leaf and branch) during merklize.
// hash is the Blake2b hash of the 64-byte node.
type StoreNodeFunc func(hash types.OpaqueHash, node TrieNode) error

// StoreValueFunc is called for regular leaf values (> 32 bytes) during merklize.
type StoreValueFunc func(value []byte) error
