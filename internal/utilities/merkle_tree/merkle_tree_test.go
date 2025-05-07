package merkle_tree

import (
	"fmt"
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
	"github.com/test-go/testify/require"
)

func TestJ0_VerifyMerkleProof(t *testing.T) {
	var segment []types.ByteSequence
	for i := 0; i < 8; i++ {
		segment = append(segment, types.ByteSequence{byte(i)})
	}
	fmt.Println("segment:", segment)
	root := M(segment, hash.Blake2bHash)

	n := 2
	proof := Jx(0, segment, types.U32(n), hash.Blake2bHash)
	for i, p := range proof {
		fmt.Printf("proof[%d]: %x\n", i, p)
	}
	ok := VerifyMerkleProof(segment[n], proof, n, hash.Blake2bHash, root)
	require.True(t, ok)
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
			fmt.Printf("Level %d: Left %x, Right %x\n", level, current, sibling)
		} else {
			node = append(types.ByteSequence("node"), sibling[:]...)
			node = append(node, current[:]...)
			fmt.Printf("Level %d: Left %x, Right %x\n", level, sibling, current)
		}
		current = hashFunc(node)
		fmt.Printf("Level %d result: %x\n", level, current)
		index /= 2
	}

	fmt.Printf("Computed root: %x\n", current)
	fmt.Printf("Expected root: %x\n", root)
	return current == root
}
