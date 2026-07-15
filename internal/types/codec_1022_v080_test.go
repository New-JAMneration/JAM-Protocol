package types

import (
	"reflect"
	"testing"
)

// TestOriginalShards_V080 covers GP v0.8.0 eq:ecoriginalshards and the
// per-mode erasure parameters derived from it.
func TestOriginalShards_V080(t *testing.T) {
	if got := OriginalShards(6); got != 3 {
		t.Errorf("OriginalShards(6) = %d, want 3", got)
	}
	if got := OriginalShards(1023); got != 342 {
		t.Errorf("OriginalShards(1023) = %d, want 342", got)
	}

	// The test binary runs in tiny mode: 3:6, with segment parameters derived
	// from original_shards.
	if DataShards != 3 || TotalShards != 6 {
		t.Errorf("tiny erasure config = %d:%d, want 3:6", DataShards, TotalShards)
	}
	if ECBasicSize != 2*DataShards {
		t.Errorf("ECBasicSize = %d, want %d (2 * original_shards)", ECBasicSize, 2*DataShards)
	}
	if ECBasicSize*ECPiecesPerSegment != SegmentSize {
		t.Errorf("ECBasicSize * ECPiecesPerSegment = %d, want SegmentSize %d",
			ECBasicSize*ECPiecesPerSegment, SegmentSize)
	}
}

// TestEncodeTicketAttempt_V080FixedByte covers the GP v0.8.0 serialization
// change: the ticket entry-index is a fixed single byte (encode[1]); v0.7.x
// used the compact natural encoding, which takes 2 bytes for values >= 128.
func TestEncodeTicketAttempt_V080FixedByte(t *testing.T) {
	// 199 is the maximum entry-index under the dynamic cap with the smallest
	// validator set (n = ceil(2*600/6) = 200); compact encoding would take 2
	// bytes here.
	attempt := TicketAttempt(199)

	encoder := NewEncoder()
	encoded, err := encoder.Encode(&attempt)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	if len(encoded) != 1 || encoded[0] != 199 {
		t.Fatalf("encoded = % x, want the single byte c7", encoded)
	}

	decoder := NewDecoder()
	var got TicketAttempt
	if err := decoder.Decode(encoded, &got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got != attempt {
		t.Errorf("round-trip = %d, want %d", got, attempt)
	}

	// Out-of-range values must error rather than truncate.
	tooBig := TicketAttempt(256)
	if _, err := encoder.Encode(&tooBig); err == nil {
		t.Errorf("encoding attempt 256 must fail, got nil error")
	}
}

// TestEncodeValidatorsData_V080LengthPrefix covers the GP v0.8.0 merklization
// C(4)/C(7)-C(9) change: validator-set sequences are length-prefixed (var);
// v0.7.x emitted them fixed-length.
func TestEncodeValidatorsData_V080LengthPrefix(t *testing.T) {
	data := make(ValidatorsData, ValidatorsCount)
	for i := range data {
		data[i].Ed25519 = Ed25519Public{byte(i + 1)}
	}

	encoder := NewEncoder()
	encoded, err := encoder.Encode(&data)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}

	// The length prefix (value ValidatorsCount) must lead the sequence,
	// followed by ValidatorsCount fixed-size validator records (32+32+144+128).
	want := 1 + ValidatorsCount*(32+32+144+128)
	if len(encoded) != want {
		t.Fatalf("encoded length = %d, want %d", len(encoded), want)
	}
	if encoded[0] != byte(ValidatorsCount) {
		t.Errorf("length prefix = %d, want %d", encoded[0], ValidatorsCount)
	}

	decoder := NewDecoder()
	var got ValidatorsData
	if err := decoder.Decode(encoded, &got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !reflect.DeepEqual(data, got) {
		t.Errorf("round-trip mismatch")
	}
}
