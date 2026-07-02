package utilities

import (
	"bytes"
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
)

// TestExtrinsicHashPreimagesComponent_V080 covers the GP v0.8.0 header.tex
// change to the extrinsic-hash sequence: the preimages component is
// p = E(var[(E4(s), blake(d))]) — each blob committed by its Blake2b hash —
// instead of the full C.15 preimage encoding.
func TestExtrinsicHashPreimagesComponent_V080(t *testing.T) {
	blob := types.ByteSequence{0xDE, 0xAD, 0xBE, 0xEF}
	preimages := types.PreimagesExtrinsic{
		{Requester: 0x0A0B0C0D, Blob: blob},
	}

	encoded, err := p(preimages)
	if err != nil {
		t.Fatalf("p: %v", err)
	}

	// Layout: count prefix(1, value 1) ++ requester E4(4) ++ blake(blob)(32).
	const want = 1 + 4 + 32
	if len(encoded) != want {
		t.Fatalf("encoded length = %d, want %d", len(encoded), want)
	}
	if encoded[0] != 1 {
		t.Errorf("count prefix = %d, want 1", encoded[0])
	}
	// requester (0x0A0B0C0D little-endian)
	if encoded[1] != 0x0D || encoded[2] != 0x0C || encoded[3] != 0x0B || encoded[4] != 0x0A {
		t.Errorf("requester bytes = % x, want 0d 0c 0b 0a", encoded[1:5])
	}
	// blob committed by hash, not inlined
	blobHash := hash.Blake2bHash(blob)
	if !bytes.Equal(encoded[5:], blobHash[:]) {
		t.Errorf("blob commitment = %x, want blake2b(blob) = %x", encoded[5:], blobHash)
	}

	// Empty extrinsic encodes as a bare zero count.
	empty, err := p(types.PreimagesExtrinsic{})
	if err != nil {
		t.Fatalf("p(empty): %v", err)
	}
	if len(empty) != 1 || empty[0] != 0 {
		t.Errorf("empty preimages encoding = % x, want a single 0 byte", empty)
	}
}
