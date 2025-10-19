package merkle_tree

import (
	"fmt"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

type HashOrByteSequence struct {
	Hash         types.OpaqueHash
	ByteSequence types.ByteSequence
}

func GetDataFromHashOrByteSequence(input HashOrByteSequence) types.ByteSequence {
	if len(input.ByteSequence) > 0 {
		return input.ByteSequence
	}
	return input.Hash[:]
}

// N: Calculates the Merkle root from integers.
func N(v []types.ByteSequence, hashFunc func(types.ByteSequence) types.OpaqueHash) (output HashOrByteSequence) {
	if len(v) == 0 {
		// H0
		output.Hash = types.OpaqueHash{} // zero hash
		return output
	} else if len(v) == 1 {
		output.ByteSequence = v[0]
		return output
	} else {
		mid := len(v) / 2
		left := v[:mid]
		right := v[mid:]
		// $node + N(left) + N(right)
		a := N(left, hashFunc)
		b := N(right, hashFunc)
		merge := types.ByteSequence("node")
		merge = append(merge, GetDataFromHashOrByteSequence(a)...)
		merge = append(merge, GetDataFromHashOrByteSequence(b)...)
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
		return N(v, hashFunc).Hash // Use N for multiple elements
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
func T(v []types.ByteSequence, i types.U32, hashFunc func(types.ByteSequence) types.OpaqueHash) (output []HashOrByteSequence) {
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

	// Use full HashOrByteSequence from N, not just .Hash
	sibling := N(siblingHalf, hashFunc)

	// Recursive trace on the half we're traversing
	suffix := T(traverseHalf, newIndex, hashFunc)

	// Prepend sibling to suffix
	return append([]HashOrByteSequence{sibling}, suffix...)
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
	// ... max(0,⌈log2(max(1,|v|))−x⌉)
	num := max(1, len(v))
	log := 0
	for (1 << log) < num {
		log++
	}
	sz := max(0, log-int(x))
	res := T(seq, i*(1<<x), hashFunc)
	ret := []types.OpaqueHash{}
	for i := 0; i < sz; i++ {
		ret = append(ret, types.OpaqueHash(GetDataFromHashOrByteSequence(res[i])))
	}
	return ret
}

// M: Constant-depth binary Merkle function
func M(v []types.ByteSequence, hashFunc func(types.ByteSequence) types.OpaqueHash) types.OpaqueHash {
	C_res := C(v, hashFunc)
	var seq []types.ByteSequence
	for _, hash := range C_res {
		seq = append(seq, types.ByteSequence(hash[:]))
	}
	return types.OpaqueHash(GetDataFromHashOrByteSequence(N(seq, hashFunc)))
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
