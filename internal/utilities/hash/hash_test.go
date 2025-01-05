package hash

import (
	"bytes"
	"encoding/hex"
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

// TestBlake2bHash tests the Blake2bHash function.
// We use the test examples from polkadot-js/common to ensure the correctness of
// our algorithms
func TestBlake2bHash(t *testing.T) {
	hash := Blake2bHash(types.ByteSequence("abc"))

	// Expected hash in opaque hash format
	excpected := types.OpaqueHash{189, 221, 129, 60, 99, 66, 57, 114, 49, 113, 239, 63, 238, 152, 87, 155, 148, 150, 78, 59, 177, 203, 62, 66, 114, 98, 200, 192, 104, 213, 35, 25}

	if !bytes.Equal(hash[:], excpected[:]) {
		t.Errorf("Expected: %v, got: %v", excpected, hash)
	}

	// Expected hash in hex format
	expectedHex := "bddd813c634239723171ef3fee98579b94964e3bb1cb3e427262c8c068d52319"

	if hex.EncodeToString(hash[:]) != expectedHex {
		t.Errorf("Expected: %v, got: %v", expectedHex, hex.EncodeToString(hash[:]))
	}
}

// TestBlake2bHashPartial tests the Blake2bHashPartial function.
// We use the test examples from polkadot-js/common to ensure the correctness of
// our algorithms
func TestBlake2bHashPartial(t *testing.T) {
	input := types.ByteSequence("abc")
	partialHash := Blake2bHashPartial(input, 4)

	// Expected hash in opaque hash format
	excpected := types.ByteSequence{189, 221, 129, 60}

	if !bytes.Equal(partialHash[:], excpected[:]) {
		t.Errorf("Expected: %v, got: %v", excpected, partialHash)
	}

	// Expected hash in hex format
	expectedHex := "bddd813c"

	if hex.EncodeToString(partialHash[:]) != expectedHex {
		t.Errorf("Expected: %v, got: %v", expectedHex, hex.EncodeToString(partialHash[:]))
	}
}

// TestKeccakHash tests the KeccakHash function.
// We use the test examples from polkadot-js/common to ensure the correctness of
// our algorithms
func TestKeccakHash(t *testing.T) {
	testCases := []struct {
		input       types.ByteSequence
		expected    types.OpaqueHash
		expectedHex string
	}{
		{
			types.ByteSequence("test"), types.OpaqueHash{
				156, 34, 255, 95, 33, 240, 184, 27, 17, 62, 99, 247, 219, 109, 169, 79,
				237, 239, 17, 178, 17, 155, 64, 136, 184, 150, 100, 251, 154, 60, 182, 88,
			}, "9c22ff5f21f0b81b113e63f7db6da94fedef11b2119b4088b89664fb9a3cb658",
		},
		{
			types.ByteSequence("test value"), types.OpaqueHash{
				45, 7, 54, 75, 92, 35, 28, 86, 206, 99, 212, 148, 48, 224, 133, 234,
				48, 51, 199, 80, 104, 139, 165, 50, 178, 64, 41, 18, 76, 38, 202, 94,
			}, "2d07364b5c231c56ce63d49430e085ea3033c750688ba532b24029124c26ca5e",
		},
	}

	for _, tc := range testCases {
		hash := KeccakHash(tc.input)

		if !bytes.Equal(hash[:], tc.expected[:]) {
			t.Errorf("Expected: %v, got: %v", tc.expected, hash)
		}

		if hex.EncodeToString(hash[:]) != tc.expectedHex {
			t.Errorf("Expected: %v, got: %v", tc.expectedHex, hex.EncodeToString(hash[:]))
		}
	}
}
