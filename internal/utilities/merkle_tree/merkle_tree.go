package merkle_tree

import (
	jamTypes "github.com/New-JAMneration/JAM-Protocol/internal/types"
)

type HashOrByteSequence struct {
	Hash         jamTypes.OpaqueHash
	ByteSequence jamTypes.ByteSequence
}

// N: Calculates the Merkle root from integers.
func N(v []jamTypes.ByteSequence, hashFunc func(jamTypes.ByteSequence) jamTypes.OpaqueHash) (output HashOrByteSequence) {
	if len(v) == 0 {
		return output
	} else if len(v) == 1 {
		output.ByteSequence = v[0] // Base case: single element
		return output
	} else {
		mid := len(v) / 2
		left := v[:mid]
		right := v[mid:]
		// $node + N(left) + N(right)
		a := N(left, hashFunc)
		b := N(right, hashFunc)
		merge := jamTypes.ByteSequence("node")
		if len(a.ByteSequence) > 0 {
			merge = append(merge, a.ByteSequence...)
		} else {
			merge = append(merge, jamTypes.ByteSequence(a.Hash[:])...)
		}
		if len(b.ByteSequence) > 0 {
			merge = append(merge, b.ByteSequence...)
		} else {
			merge = append(merge, jamTypes.ByteSequence(b.Hash[:])...)
		}
		output.Hash = hashFunc(merge) // Combine hashes of left and right subtrees
		return output
	}
}

// Mb: Well-balanced binary Merkle function
func Mb(v []jamTypes.ByteSequence, hashFunc func(jamTypes.ByteSequence) jamTypes.OpaqueHash) (output jamTypes.OpaqueHash) {
	if len(v) == 1 {
		output = hashFunc(v[0][:])
		return output
	} else {
		return Mb(v, hashFunc) // Use N for multiple elements
	}
}

// Ps: Find the half based on the given index.
func Ps(v []jamTypes.ByteSequence, i jamTypes.U32) []jamTypes.ByteSequence {
	mid := jamTypes.U32(len(v) / 2)
	if i < mid {
		return v[:mid] // Left half
	} else {
		return v[mid:] // Right half
	}
}

// PI: Determines the index of the parent node in a complete binary tree.
func PI(v []jamTypes.ByteSequence, i jamTypes.U32) jamTypes.U32 {
	half := jamTypes.U32((len(v) + 1) / 2)
	if i < half {
		return 0 // Left subtree
	} else {
		return half // Right subtree
	}
}

// T: Traces the path from the root to a leaf node, returning opposite nodes at each level to justify data inclusion.
func T(v []jamTypes.ByteSequence, i jamTypes.U32, hashFunc func(jamTypes.ByteSequence) jamTypes.OpaqueHash) (output []jamTypes.OpaqueHash) {
	if len(v) > 1 {
		suffix := T(Ps(v, i), i-PI(v, i), hashFunc) // Recursive call for suffix
		first := N(Ps(v, i), hashFunc)              // Calculate hash of prefix
		output = append([]jamTypes.OpaqueHash{first.Hash}, suffix...)
		return output
	} else {
		return output
	}
}

// Lx: Function provides a single page of hashed leaves
func Lx(v []jamTypes.ByteSequence, i jamTypes.U32, hashFunc func(jamTypes.ByteSequence) jamTypes.OpaqueHash) []jamTypes.OpaqueHash {
	pow2 := jamTypes.U32(1)
	for i*pow2*2 < jamTypes.U32(len(v)) {
		pow2 *= 2
	}
	i *= jamTypes.U32(pow2)

	ret := make([]jamTypes.OpaqueHash, 0)
	for idx := i; idx < min(i+pow2, jamTypes.U32(len(v))); idx++ {
		merge := jamTypes.ByteSequence("leaf")
		merge = append(merge, v[idx][:]...)
		ret = append(ret, hashFunc(merge))
	}
	return ret
}

// C: Pads a slice with zero hashes to the nearest power of 2.
func C(v []jamTypes.ByteSequence, hashFunc func(jamTypes.ByteSequence) jamTypes.OpaqueHash) []jamTypes.OpaqueHash {
	sz := 1
	for sz < len(v) {
		sz *= 2
	}
	ret := make([]jamTypes.OpaqueHash, sz)
	for i := 0; i < sz; i++ {
		if i < len(v) {
			merge := jamTypes.ByteSequence("leaf")
			merge = append(merge, v[i][:]...)
			ret[i] = hashFunc(merge)
		} else {
			ret[i] = jamTypes.OpaqueHash{} // Zero hash for padding
		}
	}
	return ret
}

// Jx: Function provides the Merkle path to a single page

func Jx(v []jamTypes.ByteSequence, i jamTypes.U32, hashFunc func(jamTypes.ByteSequence) jamTypes.OpaqueHash) []jamTypes.OpaqueHash {
	for i*2 < jamTypes.U32(len(v)) {
		i *= 2
	}
	C_res := C(v, hashFunc)
	var seq []jamTypes.ByteSequence
	for _, hash := range C_res {
		seq = append(seq, jamTypes.ByteSequence(hash[:]))
	}
	return T(seq, i, hashFunc)
}

// M: Constant-depth binary Merkle function
func M(v []jamTypes.ByteSequence, hashFunc func(jamTypes.ByteSequence) jamTypes.OpaqueHash) jamTypes.OpaqueHash {
	C_res := C(v, hashFunc)
	var seq []jamTypes.ByteSequence
	for _, hash := range C_res {
		seq = append(seq, jamTypes.ByteSequence(hash[:]))
	}
	return N(seq, hashFunc).Hash
}
