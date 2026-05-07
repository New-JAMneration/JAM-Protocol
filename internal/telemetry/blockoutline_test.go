package telemetry

import (
	"bytes"
	"encoding/hex"
	"testing"
)

// makeSampleBlockOutline returns a BlockOutline with deterministic
// content, used by the layout, roundtrip, and golden vector tests.
func makeSampleBlockOutline() BlockOutline {
	var hash [32]byte
	for i := range hash {
		hash[i] = 0x10 + byte(i)
	}
	return BlockOutline{
		Size:            0x12345678,
		HeaderHash:      hash,
		Tickets:         3,
		Preimages:       7,
		PreimagesBytes:  1024,
		Guarantees:      5,
		Assurances:      4,
		DisputeVerdicts: 0,
	}
}

// Encode produces exactly blockOutlineEncodedSize bytes. Catches
// accidental field reordering or size drift.
func TestBlockOutline_EncodedSize(t *testing.T) {
	enc := makeSampleBlockOutline().Encode()
	if len(enc) != blockOutlineEncodedSize {
		t.Fatalf("encoded size = %d, want %d", len(enc), blockOutlineEncodedSize)
	}
}

// Encode → ReadBlockOutline must reproduce the original. Standard
// roundtrip property.
func TestBlockOutline_Roundtrip(t *testing.T) {
	want := makeSampleBlockOutline()
	enc := want.Encode()
	got, err := NewDecoder(enc).ReadBlockOutline()
	if err != nil {
		t.Fatalf("ReadBlockOutline: %v", err)
	}
	if got != want {
		t.Errorf("roundtrip mismatch:\n got %+v\nwant %+v", got, want)
	}
}

// Trailing bytes left in the buffer after a successful read indicate a
// field-size mismatch or a missing field.
func TestBlockOutline_DecoderConsumesExactlyEncodedSize(t *testing.T) {
	enc := makeSampleBlockOutline().Encode()
	d := NewDecoder(enc)
	if _, err := d.ReadBlockOutline(); err != nil {
		t.Fatalf("ReadBlockOutline: %v", err)
	}
	if !d.Done() {
		t.Errorf("decoder has %d bytes left after read", d.Remaining())
	}
}

// Truncated input must error per field rather than silently zeroing.
func TestBlockOutline_TruncatedInputErrors(t *testing.T) {
	enc := makeSampleBlockOutline().Encode()
	for cut := 0; cut < len(enc); cut++ {
		d := NewDecoder(enc[:cut])
		if _, err := d.ReadBlockOutline(); err == nil {
			t.Errorf("cut at %d: expected error, got nil", cut)
		}
	}
}

// Golden vector: byte-for-byte for a fixed input. Catches accidental
// endianness / field order regressions.
//
// TODO(jip3-golden-source): replace with externally-verified vector
// once GP / JIP-3 / JamTART reference fixtures are available.
func TestBlockOutline_GoldenVector(t *testing.T) {
	enc := makeSampleBlockOutline().Encode()

	// Layout (offsets in bytes):
	//   [0:4]   Size            = 0x12345678 LE = 78 56 34 12
	//   [4:36]  HeaderHash      = 0x10..0x2F
	//   [36:40] Tickets         = 3 LE
	//   [40:44] Preimages       = 7 LE
	//   [44:48] PreimagesBytes  = 1024 LE = 00 04 00 00
	//   [48:52] Guarantees      = 5 LE
	//   [52:56] Assurances      = 4 LE
	//   [56:60] DisputeVerdicts = 0 LE
	expectedHex := "" +
		"78563412" +
		"101112131415161718191a1b1c1d1e1f202122232425262728292a2b2c2d2e2f" +
		"03000000" +
		"07000000" +
		"00040000" +
		"05000000" +
		"04000000" +
		"00000000"

	expected, err := hex.DecodeString(expectedHex)
	if err != nil {
		t.Fatalf("decode expected: %v", err)
	}
	if !bytes.Equal(enc, expected) {
		t.Fatalf("golden vector mismatch:\n  got:  %x\n  want: %x", enc, expected)
	}
}

// Zero-value BlockOutline encodes to all-zero bytes.
func TestBlockOutline_ZeroEncodesToZeroes(t *testing.T) {
	enc := BlockOutline{}.Encode()
	if len(enc) != blockOutlineEncodedSize {
		t.Fatalf("encoded size = %d, want %d", len(enc), blockOutlineEncodedSize)
	}
	for i, b := range enc {
		if b != 0 {
			t.Errorf("byte %d = 0x%02x, want 0x00", i, b)
		}
	}
}
