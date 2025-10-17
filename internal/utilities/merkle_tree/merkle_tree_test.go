package merkle_tree

import (
	"fmt"
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
	"github.com/test-go/testify/require"
)

func TestJ0_VerifyMerkleProof_AllLeavesWithPadding(t *testing.T) {
	var segment []types.ByteSequence
	for i := 0; i < 7; i++ {
		segment = append(segment, types.ByteSequence{byte(i)})
	}

	root := M(segment, hash.Blake2bHash)

	for n := 0; n < len(segment); n++ {
		proof := Jx(0, segment, types.U32(n), hash.Blake2bHash)
		ok := VerifyMerkleProof(segment[n], proof, n, hash.Blake2bHash, root)
		require.True(t, ok, "failed on leaf index %d", n)
	}
}

func TestN(t *testing.T) {
	hash := hash.Blake2bHash
	// leafs
	h1 := hash(append(types.ByteSequence("leaf"), 1))
	h2 := hash(append(types.ByteSequence("leaf"), 2))
	h3 := hash(append(types.ByteSequence("leaf"), 3))
	h4 := hash(append(types.ByteSequence("leaf"), 4))
	t.Run("empty slice", func(t *testing.T) {
		var empty []types.ByteSequence
		result := N(empty, hash).Hash
		require.Equal(t, types.OpaqueHash{}, result)
	})
	t.Run("single element", func(t *testing.T) {
		data := []types.ByteSequence{{42}}
		data[0] = h1[:]
		result := types.OpaqueHash(GetDataFromHashOrByteSequence(N(data, hash)))
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

		result := N(data, hash).Hash
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

		result := N(data, hash).Hash
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

		result := N(data, hash).Hash
		require.Equal(t, expected, result)
	})

}

func TestM(t *testing.T) {
	hash := hash.Blake2bHash
	t.Run("empty", func(t *testing.T) {
		data := []types.ByteSequence{}
		expected := types.OpaqueHash{}
		result := M(data, hash)
		require.Equal(t, expected, result)
	})
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

func TestT_OutputLength(t *testing.T) {
	hash := hash.Blake2bHash

	// test all n values from 2 to 64 (powers of 2)
	// n = 2, 4, 8, 16, 32, 64
	// log2(n) = 1, 2, 3, 4, 5, 6
	for n := 2; n <= 64; n *= 2 {
		t.Run(fmt.Sprintf("len=%d", n), func(t *testing.T) {
			var input []types.ByteSequence
			for i := 0; i < n; i++ {
				input = append(input, types.ByteSequence{byte(i)})
			}

			C_res := C(input, hash)
			var hashedLeaves []types.ByteSequence
			for _, h := range C_res {
				hashedLeaves = append(hashedLeaves, types.ByteSequence(h[:]))
			}

			leafIndex := types.U32(n / 2)
			path := T(hashedLeaves, leafIndex, hash)

			expectedLen := 0
			for (1 << expectedLen) < len(hashedLeaves) {
				expectedLen++
			}
			// output length should be log2(n)
			require.Len(t, path, expectedLen, "T() output length mismatch for len=%d", n)
		})
	}
}

func TestJx_OutputLength(t *testing.T) {
	hash := hash.Blake2bHash

	// test all n values from 2 to 64 (powers of 2)
	// n = 2, 4, 8, 16, 32, 64
	// log2(n) = 1, 2, 3, 4, 5, 6
	for n := 2; n <= 64; n *= 2 {
		maxLog := 0
		for (1 << maxLog) < n {
			maxLog++
		}

		// test all x values from 0 to maxLog
		for x := 0; x <= maxLog; x++ {
			t.Run(fmt.Sprintf("len=%d_x=%d", n, x), func(t *testing.T) {
				var input []types.ByteSequence
				for i := 0; i < n; i++ {
					input = append(input, types.ByteSequence{byte(i)})
				}

				pageIndex := types.U32(0)
				path := Jx(types.U8(x), input, pageIndex, hash)

				expectedLen := max(0, maxLog-x)
				require.Len(t, path, expectedLen,
					"Jx output length mismatch: expected %d, got %d (len=%d, x=%d)",
					expectedLen, len(path), n, x)
			})
		}
	}
}

func TestM_WithPadding(t *testing.T) {
	hash := hash.Blake2bHash

	// input with 3 elements, should be padded to 4
	// and then hashed
	input := []types.ByteSequence{
		[]byte("a"),
		[]byte("b"),
		[]byte("c"),
	}

	// leafs
	l0 := hash(append([]byte("leaf"), input[0]...))
	l1 := hash(append([]byte("leaf"), input[1]...))
	l2 := hash(append([]byte("leaf"), input[2]...))
	l3 := types.OpaqueHash{} // zero padding

	n0 := hash(append(append([]byte("node"), l0[:]...), l1[:]...)) // left
	n1 := hash(append(append([]byte("node"), l2[:]...), l3[:]...)) // right

	expected := hash(append(append([]byte("node"), n0[:]...), n1[:]...))

	root := M(input, hash)
	require.Equal(t, expected, root)
}
