package merkle_tree

import (
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

type HashOrByteSequence struct {
	Hash         types.OpaqueHash
	ByteSequence types.ByteSequence
}

// N: Calculates the Merkle root from integers.
func N(v []types.ByteSequence, hashFunc func(types.ByteSequence) types.OpaqueHash) (output HashOrByteSequence) {
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
		merge := types.ByteSequence("node")
		if len(a.ByteSequence) > 0 {
			merge = append(merge, a.ByteSequence...)
		} else {
			merge = append(merge, types.ByteSequence(a.Hash[:])...)
		}
		if len(b.ByteSequence) > 0 {
			merge = append(merge, b.ByteSequence...)
		} else {
			merge = append(merge, types.ByteSequence(b.Hash[:])...)
		}
		output.Hash = hashFunc(merge) // Combine hashes of left and right subtrees
		return output
	}
}

// Mb: Well-balanced binary Merkle function
func Mb(v []types.ByteSequence, hashFunc func(types.ByteSequence) types.OpaqueHash) (output types.OpaqueHash) {
	if len(v) == 1 {
		output = hashFunc(v[0][:])
		return output
	} else {
		return Mb(v, hashFunc) // Use N for multiple elements
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
func T(v []types.ByteSequence, i types.U32, hashFunc func(types.ByteSequence) types.OpaqueHash) (output []types.OpaqueHash) {
	if len(v) > 1 {
		suffix := T(Ps(v, i), i-PI(v, i), hashFunc) // Recursive call for suffix
		first := N(Ps(v, i), hashFunc)              // Calculate hash of prefix
		output = append([]types.OpaqueHash{first.Hash}, suffix...)
		return output
	} else {
		return output
	}
}

// Lx: Function provides a single page of hashed leaves
func Lx(x types.U8, v []types.ByteSequence, i types.U32, hashFunc func(types.ByteSequence) types.OpaqueHash) []types.OpaqueHash {
	ret := make([]types.OpaqueHash, 0)
	for idx := i * (1 << x); idx < min(i*(1<<x)+(1<<x), types.U32(len(v))); idx++ {
		merge := types.ByteSequence("leaf")
		merge = append(merge, v[idx][:]...)
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
	for i := 0; i < sz; i++ {
		if i < len(v) {
			merge := types.ByteSequence("leaf")
			merge = append(merge, v[i][:]...)
			ret[i] = hashFunc(merge)
		} else {
			ret[i] = types.OpaqueHash{} // Zero hash for padding
		}
	}
	return ret
}

// Jx: Function provides the Merkle path to a single page

func Jx(x types.U8, v []types.ByteSequence, i types.U32, hashFunc func(types.ByteSequence) types.OpaqueHash) []types.OpaqueHash {
	C_res := C(v, hashFunc)
	var seq []types.ByteSequence
	for _, hash := range C_res {
		seq = append(seq, types.ByteSequence(hash[:]))
	}
	return T(seq, i*(1<<x), hashFunc)
}

// M: Constant-depth binary Merkle function
func M(v []types.ByteSequence, hashFunc func(types.ByteSequence) types.OpaqueHash) types.OpaqueHash {
	C_res := C(v, hashFunc)
	var seq []types.ByteSequence
	for _, hash := range C_res {
		seq = append(seq, types.ByteSequence(hash[:]))
	}
	return N(seq, hashFunc).Hash
}
