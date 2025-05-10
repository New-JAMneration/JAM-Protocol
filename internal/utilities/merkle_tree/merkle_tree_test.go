package merkle_tree

import (
	"fmt"
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
	"github.com/test-go/testify/require"
)

func TestJ0_VerifyMerkleProof_AllLeaves(t *testing.T) {
	var segment []types.ByteSequence
	for i := 0; i < 8; i++ {
		segment = append(segment, types.ByteSequence{byte(i)})
	}

	root := M(segment, hash.Blake2bHash)

	for n := 0; n < len(segment); n++ {
		proof := Jx(0, segment, types.U32(n), hash.Blake2bHash)
		ok := VerifyMerkleProof(segment[n], proof, n, hash.Blake2bHash, root)
		require.True(t, ok, "failed on leaf index %d", n)
	}
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

func TestN(t *testing.T) {
	hash := hash.Blake2bHash
	h1 := hash(append(types.ByteSequence("leaf"), 1))
	h2 := hash(append(types.ByteSequence("leaf"), 2))
	h3 := hash(append(types.ByteSequence("leaf"), 3))
	h4 := hash(append(types.ByteSequence("leaf"), 4))
	t.Run("empty slice", func(t *testing.T) {
		var empty []types.ByteSequence
		result := N(empty, hash)
		require.Equal(t, types.OpaqueHash{}, result)
	})
	t.Run("single element", func(t *testing.T) {
		data := []types.ByteSequence{{42}}
		data[0] = h1[:]
		result := N(data, hash)
		require.Equal(t, h1, result)
	})
	t.Run("two elements", func(t *testing.T) {
		data := []types.ByteSequence{
			types.ByteSequence(h1[:]),
			types.ByteSequence(h2[:]),
		}

		merge := types.ByteSequence("node")
		merge = append(merge, data[0]...)
		merge = append(merge, data[1]...)
		expected := hash(merge)

		result := N(data, hash)
		require.Equal(t, expected, result)
	})
	t.Run("three elements", func(t *testing.T) {
		data := []types.ByteSequence{
			h1[:],
			h2[:],
			h3[:],
		}

		// Merkle tree:
		// left : data[0]
		// right: hash(data[1], data[2])
		// root : hash(node, left, right)
		left := data[0]

		right := types.ByteSequence("node")
		right = append(right, data[1]...)
		right = append(right, data[2]...)
		right_hash := hash(right)

		merge := types.ByteSequence("node")
		merge = append(merge, left...)
		merge = append(merge, right_hash[:]...)
		expected := hash(merge)

		result := N(data, hash)
		require.Equal(t, expected, result)
	})
	t.Run("four elements", func(t *testing.T) {
		data := []types.ByteSequence{
			h1[:],
			h2[:],
			h3[:],
			h4[:],
		}

		// data[0-1]
		left := types.ByteSequence("node")
		left = append(left, data[0]...)
		left = append(left, data[1]...)
		leftHash := hash(left)

		// data[2-3]
		right := types.ByteSequence("node")
		right = append(right, data[2]...)
		right = append(right, data[3]...)
		rightHash := hash(right)

		// data[0-3]
		merge := types.ByteSequence("node")
		merge = append(merge, leftHash[:]...)
		merge = append(merge, rightHash[:]...)
		expected := hash(merge)

		result := N(data, hash)
		require.Equal(t, expected, result)
	})

}

func TestM(t *testing.T) {
	hash := hash.Blake2bHash

	t.Run("one leaf only", func(t *testing.T) {
		data := []types.ByteSequence{{99}}
		expected := hash(append(types.ByteSequence("leaf"), 99))
		result := M(data, hash)
		require.Equal(t, expected, result)
	})

	t.Run("non-power-of-2 leaf padding", func(t *testing.T) {
		data := []types.ByteSequence{{1}, {2}, {3}} // Should pad to 4
		result := M(data, hash)
		require.NotEqual(t, types.OpaqueHash{}, result)
		require.Len(t, C(data, hash), 4)
	})
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
	require.Len(t, path, 2)
}

func TestJ0_Equals_T(t *testing.T) {
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

func TestN_PanicsOnInvalidSingleElement(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for single element not 32 bytes, but none occurred")
		} else {
			fmt.Println("✅ Panic caught as expected:", r)
		}
	}()

	hash := hash.Blake2bHash
	invalid := []types.ByteSequence{{0x01, 0x02, 0x03}} // Not 32 bytes
	N(invalid, hash)                                    // should panic
}
