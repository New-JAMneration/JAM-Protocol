package merkle_tree

import (
	"fmt"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

// Package-level constants
var (
	// nodePrefix is the prefix used for internal node hashing in N function
	nodePrefix = []byte("node")
	// leafPrefix is the prefix used for leaf node hashing in Lx and C functions
	leafPrefix = []byte("leaf")
	// zeroHash is the zero OpaqueHash used for empty Merkle trees
	zeroHash = types.OpaqueHash{}
)

// N: Calculates the Merkle root from integers.
func N(v []types.ByteSequence, hashFunc func(types.ByteSequence) types.OpaqueHash) types.ByteSequence {
	// [[]] should result zero hash
	if len(v) == 0 || v[0] == nil {
		// H0 - return zero hash as bytes
		return types.ByteSequence(zeroHash[:])
	} else if len(v) == 1 {
		// Single element: return raw data
		return v[0]
	} else {
		mid := (len(v) + 1) / 2
		left := v[:mid]
		right := v[mid:]
		// $node + N(left) + N(right)
		a := N(left, hashFunc)
		b := N(right, hashFunc)

		// Pre-allocate buffer with exact capacity to avoid multiple reallocations
		merge := make([]byte, 0, len(nodePrefix)+len(a)+len(b))
		merge = append(merge, nodePrefix...)
		merge = append(merge, a...)
		merge = append(merge, b...)

		// Return hash as ByteSequence
		hash := hashFunc(merge)
		return types.ByteSequence(hash[:])
	}
}

// Mb: Well-balanced binary Merkle function
func Mb(v []types.ByteSequence, hashFunc func(types.ByteSequence) types.OpaqueHash) types.OpaqueHash {
	// [[]] should go to N
	if len(v) == 1 && v[0] != nil {
		return hashFunc(v[0])
	} else {
		// N returns ByteSequence, convert to OpaqueHash
		return types.OpaqueHash(N(v, hashFunc))
	}
}

// Ps: Find the half based on the given index.
func Ps(v []types.ByteSequence, i types.U32) []types.ByteSequence {
	mid := types.U32(len(v) / 2)
	if i < mid {
		return v[:mid] // Left half
	} else {
		return v[mid:] // Right half
	}
}

// PI: Determines the index of the parent node in a complete binary tree.
func PI(v []types.ByteSequence, i types.U32) types.U32 {
	half := types.U32((len(v) + 1) / 2)
	if i < half {
		return 0 // Left subtree
	} else {
		return half // Right subtree
	}
}

// T: Traces the path from the root to a leaf node, returning opposite nodes at each level to justify data inclusion.
func T(v []types.ByteSequence, i types.U32, hashFunc func(types.ByteSequence) types.OpaqueHash) (output []types.ByteSequence) {
	if len(v) <= 1 {
		return output
	}
	mid := types.U32(len(v) / 2)
	var siblingHalf []types.ByteSequence
	var traverseHalf []types.ByteSequence
	var newIndex types.U32

	if i < mid {
		siblingHalf = v[mid:]  // right is sibling
		traverseHalf = v[:mid] // go left
		newIndex = i
	} else {
		siblingHalf = v[:mid]  // left is sibling
		traverseHalf = v[mid:] // go right
		newIndex = i - mid
	}

	// N returns ByteSequence directly
	sibling := N(siblingHalf, hashFunc)

	// Recursive trace on the half we're traversing
	suffix := T(traverseHalf, newIndex, hashFunc)

	// Pre-allocate capacity to avoid reallocation
	// sibling is a single ByteSequence, so capacity should be 1 + len(suffix)
	result := make([]types.ByteSequence, 0, 1+len(suffix))
	result = append(result, sibling)
	result = append(result, suffix...)
	return result
}

// Lx: Function provides a single page of hashed leaves
func Lx(x types.U8, v []types.ByteSequence, i types.U32, hashFunc func(types.ByteSequence) types.OpaqueHash) []types.OpaqueHash {
	pageSize := 1 << x
	start := i * types.U32(pageSize)
	end := min(start+types.U32(pageSize), types.U32(len(v)))

	// Pre-allocate capacity to avoid dynamic resizing
	ret := make([]types.OpaqueHash, 0, end-start)

	var merge types.ByteSequence

	for idx := start; idx < end; idx++ {
		// Reset merge buffer and append "leaf" prefix
		merge = merge[:0]
		merge = append(merge, leafPrefix...)
		merge = append(merge, v[idx]...)
		ret = append(ret, hashFunc(merge))
	}
	return ret
}

// C: Pads a slice with zero hashes to the nearest power of 2.
func C(v []types.ByteSequence, hashFunc func(types.ByteSequence) types.OpaqueHash) []types.OpaqueHash {
	sz := 1
	for sz < len(v) {
		sz *= 2
	}
	ret := make([]types.OpaqueHash, sz)
	var merge types.ByteSequence
	for i := 0; i < sz; i++ {
		if i < len(v) {
			// Reset merge buffer and append "leaf" prefix + value
			merge = merge[:0]
			merge = append(merge, leafPrefix...)
			merge = append(merge, v[i]...)
			ret[i] = hashFunc(merge)
		} else {
			// constant for zero hash
			ret[i] = zeroHash
		}
	}
	return ret
}

// Jx: Function provides the Merkle path to a single page
func Jx(x types.U8, v []types.ByteSequence, i types.U32, hashFunc func(types.ByteSequence) types.OpaqueHash) []types.OpaqueHash {
	C_res := C(v, hashFunc)

	// Pre-allocate with exact size to avoid reallocation
	seq := make([]types.ByteSequence, len(C_res))
	for i, hash := range C_res {
		// Direct slice conversion
		seq[i] = types.ByteSequence(hash[:])
	}
	// ... max(0,⌈log2(max(1,|v|))−x⌉)
	num := max(1, len(v))
	log := 0
	for (1 << log) < num {
		log++
	}
	sz := max(0, log-int(x))
	res := T(seq, i*(1<<x), hashFunc)

	// Pre-allocate with exact size to avoid reallocation
	// res is now []types.ByteSequence, convert to []types.OpaqueHash
	ret := make([]types.OpaqueHash, sz)
	for i := 0; i < sz && i < len(res); i++ {
		// Direct copy with bounds check
		copy(ret[i][:], res[i])
	}
	return ret
}

// M: Constant-depth binary Merkle function
func M(v []types.ByteSequence, hashFunc func(types.ByteSequence) types.OpaqueHash) types.OpaqueHash {
	C_res := C(v, hashFunc)

	// Pre-allocate with exact size to avoid reallocation
	seq := make([]types.ByteSequence, len(C_res))
	for i, hash := range C_res {
		// Direct slice conversion
		seq[i] = types.ByteSequence(hash[:])
	}

	// N returns ByteSequence, convert to OpaqueHash
	nResult := N(seq, hashFunc)
	var hash types.OpaqueHash
	copy(hash[:], nResult)
	return hash
}

func VerifyMerkleProof(leaf []byte, proof []types.OpaqueHash, index int, hashFunc func(types.ByteSequence) types.OpaqueHash, root types.OpaqueHash) bool {
	h := append(types.ByteSequence("leaf"), leaf...)
	current := hashFunc(h)
	fmt.Printf("Leaf hash: %x\n", current)

	for level := len(proof) - 1; level >= 0; level-- {
		sibling := proof[level]
		var node []byte
		if index%2 == 0 {
			node = append(types.ByteSequence("node"), current[:]...)
			node = append(node, sibling[:]...)
			// fmt.Printf("Level %d: Left %x, Right %x\n", level, current, sibling)
		} else {
			node = append(types.ByteSequence("node"), sibling[:]...)
			node = append(node, current[:]...)
			// fmt.Printf("Level %d: Left %x, Right %x\n", level, sibling, current)
		}
		current = hashFunc(node)
		// fmt.Printf("Level %d result: %x\n", level, current)
		index /= 2
	}

	// fmt.Printf("Computed root: %x\n", current)
	// fmt.Printf("Expected root: %x\n", root)
	return current == root
}
