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

	// PrintMerkleTree(segment, hash.Blake2bHash)
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
func TestN_EmptyInput(t *testing.T) {
	hash := hash.Blake2bHash
	var empty []types.ByteSequence
	require.Equal(t, types.OpaqueHash{}, N(empty, hash))
}

func TestM_SingleElement(t *testing.T) {
	hash := hash.Blake2bHash
	input := []types.ByteSequence{{42}}
	expected := hash(append(types.ByteSequence("leaf"), 42))
	require.Equal(t, expected, M(input, hash))
}

func TestC_Padding(t *testing.T) {
	hash := hash.Blake2bHash
	input := []types.ByteSequence{{1}, {2}, {3}}
	out := C(input, hash)
	require.Len(t, out, 4)
	require.NotEqual(t, types.OpaqueHash{}, out[0])
	require.Equal(t, types.OpaqueHash{}, out[3])
}

func TestT_PathLength(t *testing.T) {
	hash := hash.Blake2bHash
	input := []types.ByteSequence{{0}, {1}, {2}, {3}}
	c := C(input, hash)
	var C []types.ByteSequence
	for _, val := range c {
		C = append(C, types.ByteSequence(val[:]))
	}

	path := T(C, 2, hash)
	require.Len(t, path, 2) // depth-2 tree，包含兩個 sibling
}

func TestJx_Equals_T(t *testing.T) {
	hash := hash.Blake2bHash
	input := []types.ByteSequence{{0}, {1}, {2}, {3}}
	output := C(input, hash)
	C := make([]types.ByteSequence, len(output))
	for i, val := range output {
		C[i] = types.ByteSequence(val[:])
	}
	tPath := T(C, 2, hash)
	jxPath := Jx(0, input, 2, hash)
	require.Equal(t, tPath, jxPath)
}
