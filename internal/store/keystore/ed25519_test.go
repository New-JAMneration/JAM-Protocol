package keystore_test

import (
	"encoding/hex"
	"path/filepath"
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/store/keystore"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities"
	"github.com/hdevalence/ed25519consensus"
)

// TestVector represents a single test vector from vectors.json
type TestVector struct {
	Number      int    `json:"number"`
	Desc        string `json:"desc"`
	PK          string `json:"pk"`           // Public key (hex)
	R           string `json:"r"`            // R point (hex)
	S           string `json:"s"`            // s scalar (hex)
	Msg         string `json:"msg"`          // Message (hex)
	PKCanonical bool   `json:"pk_canonical"` // Whether PK is canonical
	RCanonical  bool   `json:"r_canonical"`  // Whether R is canonical
}

// loadTestVectors loads test vectors from vectors.json
func loadTestVectors(t *testing.T) []TestVector {
	// Get the path to vectors.json relative to the test file
	// Adjust this path based on your project structure
	vectorsPath := filepath.Join("..", "..", "..", "pkg", "test_data", "jam-conformance", "crypto", "ed25519", "vectors.json")

	// Alternative: use absolute path from workspace root
	// vectorsPath := "pkg/test_data/jam-conformance/crypto/ed25519/vectors.json"

	vectors, err := utilities.GetTestFromJson[[]TestVector](vectorsPath)
	if err != nil {
		t.Fatalf("Failed to load test vectors: %v", err)
	}

	return vectors
}

// hexDecode decodes a hex string to bytes
func hexDecode(s string) []byte {
	data, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return data
}

// TestEd25519ConsensusVectors tests all 196 test vectors against ed25519consensus
// According to ZIP 215, all vectors should pass verification
func TestEd25519ConsensusVectors(t *testing.T) {
	vectors := loadTestVectors(t)

	passed := 0
	failed := 0

	for _, vec := range vectors {
		t.Run(vec.Desc, func(t *testing.T) {
			// Decode hex strings to bytes
			publicKey := hexDecode(vec.PK)
			rPoint := hexDecode(vec.R)
			sScalar := hexDecode(vec.S)
			message := hexDecode(vec.Msg)

			// Construct signature: sig = R_bytes || s_bytes (64 bytes total)
			signature := make([]byte, 64)
			copy(signature[:32], rPoint)
			copy(signature[32:], sScalar)

			// Verify using ed25519consensus (ZIP 215 compliant)
			valid := ed25519consensus.Verify(publicKey, message, signature)
			// valid := ed25519.Verify(publicKey, message, signature)

			if !valid {
				failed++
				t.Errorf("Vector #%d (%s) failed verification\n"+
					"  PK: %s\n"+
					"  R:  %s\n"+
					"  s:  %s\n"+
					"  msg: %s\n"+
					"  PK canonical: %v, R canonical: %v",
					vec.Number, vec.Desc,
					vec.PK, vec.R, vec.S, vec.Msg,
					vec.PKCanonical, vec.RCanonical)
			} else {
				passed++
			}
		})
	}

	t.Logf("Test results: %d passed, %d failed out of %d total vectors",
		passed, failed, len(vectors))

	// According to ZIP 215, all vectors should pass
	if failed > 0 {
		t.Errorf("Expected all vectors to pass (ZIP 215 compliance), but %d failed", failed)
	}
}

// TestEd25519KeyPairVerify tests using the Ed25519KeyPair.Verify method
func TestEd25519KeyPairVerify(t *testing.T) {
	vectors := loadTestVectors(t)

	// Create a key pair (we'll use the first vector's public key)
	if len(vectors) == 0 {
		t.Fatal("No test vectors available")
	}
	passed := 0
	failed := 0

	for _, vec := range vectors {
		t.Run(vec.Desc, func(t *testing.T) {
			publicKey := hexDecode(vec.PK)

			// Create Ed25519KeyPair with the public key
			keyPair := &keystore.Ed25519KeyPair{
				Public: publicKey,
			}

			// Test verification using the key pair's Verify method
			rPoint := hexDecode(vec.R)
			sScalar := hexDecode(vec.S)
			message := hexDecode(vec.Msg)

			signature := make([]byte, 64)
			copy(signature[:32], rPoint)
			copy(signature[32:], sScalar)

			valid := keyPair.Verify(message, signature)

			if !valid {
				failed++
				t.Errorf("KeyPair.Verify failed for vector #%d", vec.Number)
			} else {
				passed++
				t.Logf("KeyPair.Verify passed for vector #%d", vec.Number)
			}
		})
	}
	t.Logf("Test results: %d passed, %d failed out of %d total vectors",
		passed, failed, len(vectors))

	// According to ZIP 215, all vectors should pass
	if failed > 0 {
		t.Errorf("Expected all vectors to pass (ZIP 215 compliance), but %d failed", failed)
	}
}
