package merkle_tree

import (
	"github.com/New-JAMneration/JAM-Protocol/internal/jam_types"
	jamTypes "github.com/New-JAMneration/JAM-Protocol/internal/jam_types"
	hashUtil "github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
)

// H: Calculates the hash of a given string.
func H(s string) int {
	// ... implementation of hash function ...
	return 0 // Replace with actual hash calculation
}

// N: Calculates the Merkle root from integers.
func N(v []jamTypes.OpaqueHash) (output jamTypes.OpaqueHash) {
	if len(v) == 0 {
		return output
	} else if len(v) == 1 {
		return v[0] // Base case: single element
	} else {
		mid := len(v) / 2
		left := v[:mid]
		right := v[mid:]
		// TODO() add $node:
		// $node + N(left) + N(right)
		a := N(left)
		b := N(right)
		combined := append(a[:], b[:]...)
		output = hashUtil.Blake2bHash(jam_types.ByteSequence(combined)) // Combine hashes of left and right subtrees
		return output
	}
}

// Mb: Well-balanced binary Merkle function
func Mb(v []jamTypes.OpaqueHash) (output jamTypes.OpaqueHash) {
	if len(v) == 1 {
		output = hashUtil.Blake2bHash(jam_types.ByteSequence(v[0][:]))
		return output
	} else {
		return N(v) // Use N for multiple elements
	}
}

func Mb_wHashFunction(v []jamTypes.OpaqueHash, hashFunc func(jamTypes.ByteSequence) jamTypes.OpaqueHash) (output jamTypes.OpaqueHash) {
	if len(v) == 1 {
		output = hashFunc(jam_types.ByteSequence(v[0][:]))
		return output
	} else {
		return N(v) // Use N for multiple elements
	}
}

// Ps: Find the half based on the given index.
// TODO(): check U32
func Ps(v []jamTypes.OpaqueHash, i jamTypes.U32) []jamTypes.OpaqueHash {
	mid := jamTypes.U32(len(v) / 2)
	if i < mid {
		return v[:mid] // Left half
	} else {
		return v[mid:] // Right half
	}
}

// PI: Determines the index of the parent node in a complete binary tree.
func PI(v []jamTypes.OpaqueHash, i jamTypes.U32) jamTypes.U32 {
	half := jamTypes.U32((len(v) + 1) / 2)
	if i < half {
		return 0 // Left subtree
	} else {
		return half // Right subtree
	}
}

// T: Traces the path from the root to a leaf node, returning opposite nodes at each level to justify data inclusion.
// TODO add type of merkle proof
func T(v []jamTypes.OpaqueHash, i jamTypes.U32) (output []jamTypes.OpaqueHash) {
	if len(v) > 1 {
		suffix := T(Ps(v, i), i-PI(v, i))         // Recursive call for suffix
		first := N(Ps(v, jamTypes.U32(len(v))-i)) // Calculate hash of prefix
		// Convert first (ByteArray32) to ByteSequence
		firstSlice := jamTypes.OpaqueHash(first[:])

		// Concatenate firstSlice and suffix
		output = append([]jamTypes.OpaqueHash{firstSlice}, suffix...)
		return output
	} else {
		return output
	}
}

// Lx: Function provides a single page of hashed leaves
func Lx(v []jamTypes.OpaqueHash, i jamTypes.U32) []jamTypes.OpaqueHash {
	pow2 := jamTypes.U32(1)
	for i*pow2*2 < jamTypes.U32(len(v)) {
		pow2 *= 2
	}
	i *= jamTypes.U32(pow2)

	ret := make([]jamTypes.OpaqueHash, 0)
	for idx := i; idx < min(i+pow2, jamTypes.U32(len(v))); idx++ {
		ret = append(ret, hashUtil.Blake2bHash(jam_types.ByteSequence(v[idx][:])))
	}
	return ret
}

// C: Pads a slice with zero hashes to the nearest power of 2.
func C(v []jamTypes.OpaqueHash) []jamTypes.OpaqueHash {
	sz := 1
	for sz < len(v) {
		sz *= 2
	}
	ret := make([]jamTypes.OpaqueHash, sz)
	for i := 0; i < sz; i++ {
		if i < len(v) {
			ret[i] = hashUtil.Blake2bHash(jam_types.ByteSequence(v[i][:])) // TODO $leaf
		} else {
			ret[i] = jamTypes.OpaqueHash{} // Zero hash for padding
		}
	}
	return ret
}

// Jx: Function provides the Merkle path to a single page
func Jx(v []jamTypes.OpaqueHash, i jamTypes.U32) []jamTypes.OpaqueHash {
	for i*2 < jamTypes.U32(len(v)) {
		i *= 2
	}
	return T(C(v), i)
}

// M: Constant-depth binary Merkle function
func M(v []jamTypes.OpaqueHash) jamTypes.OpaqueHash {
	return N(C(v))
}
