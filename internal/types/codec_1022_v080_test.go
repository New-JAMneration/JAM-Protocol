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

// TestSetErasureParameters_SyncOnValidatorCountChange is the regression for
// the #1035 review finding: every path that changes ValidatorsCount must
// re-derive the erasure parameters (GP v0.8.0 eq:ecoriginalshards), or a node
// keeps a stale coding rate and produces divergent erasure roots.
func TestSetErasureParameters_SyncOnValidatorCountChange(t *testing.T) {
	t.Cleanup(SetTinyMode)

	assertParams := func(t *testing.T, wantData, wantTotal, wantWE, wantWP int) {
		t.Helper()
		if DataShards != wantData || TotalShards != wantTotal {
			t.Errorf("shards = %d:%d, want %d:%d", DataShards, TotalShards, wantData, wantTotal)
		}
		if ECBasicSize != wantWE || ECPiecesPerSegment != wantWP {
			t.Errorf("W_E/W_P = %d/%d, want %d/%d", ECBasicSize, ECPiecesPerSegment, wantWE, wantWP)
		}
	}

	// Mode switches keep everything in sync.
	SetFullMode()
	assertParams(t, 342, 1023, 684, 6)
	SetTinyMode()
	assertParams(t, 3, 6, 6, 684)

	// The custom-config path funnels through SetErasureParameters directly.
	SetErasureParameters(1023)
	assertParams(t, 342, 1023, 684, 6)
	SetErasureParameters(6)
	assertParams(t, 3, 6, 6, 684)
}

// TestApplyProtocolParameters_ErasureSync covers the chainspec path: applying
// protocol parameters with a new validator count re-derives the erasure
// parameters, and a chainspec whose W_E / W_P disagree with the values
// derived from V is rejected.
func TestApplyProtocolParameters_ErasureSync(t *testing.T) {
	t.Cleanup(SetTinyMode)

	fullPP := func() ProtocolParameters {
		return ProtocolParameters{
			BI: 10, BL: 1, BS: 100,
			C: 341, D: LookupAnchorMaxAge + 4800, E: 600,
			GA: 10_000_000, GI: 50_000_000, GR: 5_000_000_000, GT: 3_500_000_000,
			H: 8, I: 16, J: 8, K: 16, L: 14400, N: 2, O: 8, P: 6, Q: 80, R: 10,
			T: 128, U: 5, V: 1023,
			WA: 64_000, WB: 13_791_360, WC: 4_000_000,
			WE: 684, WM: 3072, WP: 6, WR: 48 * 1024, WT: 128, WX: 3072,
			Y: 500,
		}
	}

	// Valid full chainspec: erasure parameters follow V.
	SetTinyMode() // start from the stale tiny values the bug left behind
	if err := applyProtocolParametersImpl(fullPP()); err != nil {
		t.Fatalf("apply full chainspec: %v", err)
	}
	if DataShards != 342 || TotalShards != 1023 || ECBasicSize != 684 || ECPiecesPerSegment != 6 {
		t.Errorf("after full chainspec: shards %d:%d W_E %d W_P %d, want 342:1023 684 6",
			DataShards, TotalShards, ECBasicSize, ECPiecesPerSegment)
	}

	// A chainspec whose W_E disagrees with the V-derived value is rejected.
	bad := fullPP()
	bad.WE = 4
	if err := applyProtocolParametersImpl(bad); err == nil {
		t.Errorf("chainspec with mismatched W_E must be rejected")
	}
}
