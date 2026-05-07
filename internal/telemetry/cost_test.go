package telemetry

import (
	"bytes"
	"encoding/hex"
	"testing"
)

// makeSampleExecCost returns an ExecCost with deterministic content.
func makeSampleExecCost(seed uint64) ExecCost {
	return ExecCost{
		GasUsed:      seed,
		ElapsedNanos: seed*2 + 1,
	}
}

func TestExecCost_EncodedSize(t *testing.T) {
	if got := len(makeSampleExecCost(1).Encode()); got != execCostEncodedSize {
		t.Errorf("ExecCost size = %d, want %d", got, execCostEncodedSize)
	}
}

func TestExecCost_Roundtrip(t *testing.T) {
	want := makeSampleExecCost(0xDEADBEEF)
	got, err := NewDecoder(want.Encode()).ReadExecCost()
	if err != nil {
		t.Fatalf("ReadExecCost: %v", err)
	}
	if got != want {
		t.Errorf("roundtrip mismatch:\n got %+v\nwant %+v", got, want)
	}
}

// Zero-fill: AccumulateCost / IsAuthorizedCost / RefineCost are emitted
// zero-filled per #775 Q2 until PVM cost instrumentation lands. Encode
// must produce all-zero bytes for a zero-value struct so callers can
// rely on it.
func TestExecCost_ZeroEncodesToZeroes(t *testing.T) {
	enc := ExecCost{}.Encode()
	if len(enc) != execCostEncodedSize {
		t.Fatalf("encoded size = %d, want %d", len(enc), execCostEncodedSize)
	}
	for i, b := range enc {
		if b != 0 {
			t.Errorf("byte %d = 0x%02x, want 0x00", i, b)
		}
	}
}

// IsAuthorizedCost roundtrip + size.
func TestIsAuthorizedCost_RoundtripAndSize(t *testing.T) {
	want := IsAuthorizedCost{
		Total:        makeSampleExecCost(10),
		CompileNanos: 0x1122334455667788,
		HostCalls:    makeSampleExecCost(20),
	}
	enc := want.Encode()
	if len(enc) != isAuthorizedCostEncodedSize {
		t.Fatalf("size = %d, want %d", len(enc), isAuthorizedCostEncodedSize)
	}
	got, err := NewDecoder(enc).ReadIsAuthorizedCost()
	if err != nil {
		t.Fatalf("ReadIsAuthorizedCost: %v", err)
	}
	if got != want {
		t.Errorf("roundtrip mismatch:\n got %+v\nwant %+v", got, want)
	}
}

func TestIsAuthorizedCost_ZeroEncodesToZeroes(t *testing.T) {
	enc := IsAuthorizedCost{}.Encode()
	if len(enc) != isAuthorizedCostEncodedSize {
		t.Fatalf("encoded size = %d, want %d", len(enc), isAuthorizedCostEncodedSize)
	}
	for i, b := range enc {
		if b != 0 {
			t.Errorf("byte %d = 0x%02x, want 0x00", i, b)
		}
	}
}

// RefineCost roundtrip + size + zero-fill.
func TestRefineCost_RoundtripAndSize(t *testing.T) {
	want := RefineCost{
		Total:            makeSampleExecCost(100),
		CompileNanos:     999,
		HistoricalLookup: makeSampleExecCost(101),
		MachineExpunge:   makeSampleExecCost(102),
		PeekPokePages:    makeSampleExecCost(103),
		Invoke:           makeSampleExecCost(104),
		Other:            makeSampleExecCost(105),
	}
	enc := want.Encode()
	if len(enc) != refineCostEncodedSize {
		t.Fatalf("size = %d, want %d", len(enc), refineCostEncodedSize)
	}
	got, err := NewDecoder(enc).ReadRefineCost()
	if err != nil {
		t.Fatalf("ReadRefineCost: %v", err)
	}
	if got != want {
		t.Errorf("roundtrip mismatch:\n got %+v\nwant %+v", got, want)
	}
}

func TestRefineCost_ZeroEncodesToZeroes(t *testing.T) {
	enc := RefineCost{}.Encode()
	if len(enc) != refineCostEncodedSize {
		t.Fatalf("encoded size = %d, want %d", len(enc), refineCostEncodedSize)
	}
	for i, b := range enc {
		if b != 0 {
			t.Errorf("byte %d = 0x%02x, want 0x00", i, b)
		}
	}
}

// AccumulateCost is the largest of the four; verify all 12 fields
// roundtrip correctly.
func TestAccumulateCost_RoundtripAndSize(t *testing.T) {
	want := AccumulateCost{
		AccumulateCalls:           1,
		TransfersProcessed:        2,
		ItemsAccumulated:          3,
		Total:                     makeSampleExecCost(10),
		CompileNanos:              0xAABBCCDDEEFF1122,
		ReadWrite:                 makeSampleExecCost(20),
		Lookup:                    makeSampleExecCost(30),
		QuerySolicitForgetProvide: makeSampleExecCost(40),
		InfoNewUpgradeEject:       makeSampleExecCost(50),
		Transfer:                  makeSampleExecCost(60),
		TotalTransferGas:          0x1234567890ABCDEF,
		Other:                     makeSampleExecCost(70),
	}
	enc := want.Encode()
	if len(enc) != accumulateCostEncodedSize {
		t.Fatalf("size = %d, want %d", len(enc), accumulateCostEncodedSize)
	}
	got, err := NewDecoder(enc).ReadAccumulateCost()
	if err != nil {
		t.Fatalf("ReadAccumulateCost: %v", err)
	}
	if got != want {
		t.Errorf("roundtrip mismatch:\n got %+v\nwant %+v", got, want)
	}
}

func TestAccumulateCost_ZeroEncodesToZeroes(t *testing.T) {
	enc := AccumulateCost{}.Encode()
	if len(enc) != accumulateCostEncodedSize {
		t.Fatalf("encoded size = %d, want %d", len(enc), accumulateCostEncodedSize)
	}
	for i, b := range enc {
		if b != 0 {
			t.Errorf("byte %d = 0x%02x, want 0x00", i, b)
		}
	}
}

// Trailing bytes after a Cost decode indicate field-size drift.
func TestCosts_DecoderConsumesExactlyEncodedSize(t *testing.T) {
	cases := []struct {
		name string
		fn   func() (int, error)
	}{
		{"ExecCost", func() (int, error) {
			d := NewDecoder(ExecCost{}.Encode())
			if _, err := d.ReadExecCost(); err != nil {
				return 0, err
			}
			return d.Remaining(), nil
		}},
		{"IsAuthorizedCost", func() (int, error) {
			d := NewDecoder(IsAuthorizedCost{}.Encode())
			if _, err := d.ReadIsAuthorizedCost(); err != nil {
				return 0, err
			}
			return d.Remaining(), nil
		}},
		{"RefineCost", func() (int, error) {
			d := NewDecoder(RefineCost{}.Encode())
			if _, err := d.ReadRefineCost(); err != nil {
				return 0, err
			}
			return d.Remaining(), nil
		}},
		{"AccumulateCost", func() (int, error) {
			d := NewDecoder(AccumulateCost{}.Encode())
			if _, err := d.ReadAccumulateCost(); err != nil {
				return 0, err
			}
			return d.Remaining(), nil
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rem, err := tc.fn()
			if err != nil {
				t.Fatalf("decode: %v", err)
			}
			if rem != 0 {
				t.Errorf("decoder has %d trailing bytes", rem)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Golden vectors: byte-for-byte for fixed inputs. Catch endianness or
// field-order regressions that roundtrip + size tests miss (e.g. swapping
// two same-typed adjacent fields, or putting AccumulateCost.TotalTransferGas
// in the wrong position relative to Transfer / Other).
//
// Self-consistent (generated from the encoder we commit). Replace with
// externally-verified vectors once GP / JIP-3 / JamTART reference data
// is available — see TODO in blockoutline_test.go.
// ---------------------------------------------------------------------------

func TestExecCost_GoldenVector(t *testing.T) {
	c := ExecCost{
		GasUsed:      0x0807060504030201,
		ElapsedNanos: 0x1817161514131211,
	}
	// u64 LE GasUsed ++ u64 LE ElapsedNanos
	want := mustHex(t, "0102030405060708"+"1112131415161718")
	if got := c.Encode(); !bytes.Equal(got, want) {
		t.Fatalf("ExecCost golden mismatch:\n  got:  %x\n  want: %x", got, want)
	}
}

func TestIsAuthorizedCost_GoldenVector(t *testing.T) {
	c := IsAuthorizedCost{
		Total:        ExecCost{GasUsed: 1, ElapsedNanos: 2},
		CompileNanos: 3,
		HostCalls:    ExecCost{GasUsed: 4, ElapsedNanos: 5},
	}
	want := mustHex(t,
		"0100000000000000"+"0200000000000000"+ // Total
			"0300000000000000"+ // CompileNanos
			"0400000000000000"+"0500000000000000") // HostCalls
	if got := c.Encode(); !bytes.Equal(got, want) {
		t.Fatalf("IsAuthorizedCost golden mismatch:\n  got:  %x\n  want: %x", got, want)
	}
}

func TestRefineCost_GoldenVector(t *testing.T) {
	c := RefineCost{
		Total:            ExecCost{GasUsed: 1, ElapsedNanos: 2},
		CompileNanos:     3,
		HistoricalLookup: ExecCost{GasUsed: 4, ElapsedNanos: 5},
		MachineExpunge:   ExecCost{GasUsed: 6, ElapsedNanos: 7},
		PeekPokePages:    ExecCost{GasUsed: 8, ElapsedNanos: 9},
		Invoke:           ExecCost{GasUsed: 10, ElapsedNanos: 11},
		Other:            ExecCost{GasUsed: 12, ElapsedNanos: 13},
	}
	want := mustHex(t,
		"0100000000000000"+"0200000000000000"+ // Total
			"0300000000000000"+ // CompileNanos
			"0400000000000000"+"0500000000000000"+ // HistoricalLookup
			"0600000000000000"+"0700000000000000"+ // MachineExpunge
			"0800000000000000"+"0900000000000000"+ // PeekPokePages
			"0a00000000000000"+"0b00000000000000"+ // Invoke
			"0c00000000000000"+"0d00000000000000") // Other
	if got := c.Encode(); !bytes.Equal(got, want) {
		t.Fatalf("RefineCost golden mismatch:\n  got:  %x\n  want: %x", got, want)
	}
}

// AccumulateCost is the trickiest: TotalTransferGas sits between the
// Transfer ExecCost and the Other ExecCost. Easy to swap by accident in
// a future refactor. The golden vector pins the order.
func TestAccumulateCost_GoldenVector(t *testing.T) {
	c := AccumulateCost{
		AccumulateCalls:           0x11,
		TransfersProcessed:        0x22,
		ItemsAccumulated:          0x33,
		Total:                     ExecCost{GasUsed: 1, ElapsedNanos: 2},
		CompileNanos:              3,
		ReadWrite:                 ExecCost{GasUsed: 4, ElapsedNanos: 5},
		Lookup:                    ExecCost{GasUsed: 6, ElapsedNanos: 7},
		QuerySolicitForgetProvide: ExecCost{GasUsed: 8, ElapsedNanos: 9},
		InfoNewUpgradeEject:       ExecCost{GasUsed: 10, ElapsedNanos: 11},
		Transfer:                  ExecCost{GasUsed: 12, ElapsedNanos: 13},
		TotalTransferGas:          0x44,
		Other:                     ExecCost{GasUsed: 14, ElapsedNanos: 15},
	}
	want := mustHex(t,
		"11000000"+"22000000"+"33000000"+ // u32 counts
			"0100000000000000"+"0200000000000000"+ // Total
			"0300000000000000"+ // CompileNanos
			"0400000000000000"+"0500000000000000"+ // ReadWrite
			"0600000000000000"+"0700000000000000"+ // Lookup
			"0800000000000000"+"0900000000000000"+ // QuerySolicitForgetProvide
			"0a00000000000000"+"0b00000000000000"+ // InfoNewUpgradeEject
			"0c00000000000000"+"0d00000000000000"+ // Transfer
			"4400000000000000"+ // TotalTransferGas — pinned BEFORE Other
			"0e00000000000000"+"0f00000000000000") // Other
	if got := c.Encode(); !bytes.Equal(got, want) {
		t.Fatalf("AccumulateCost golden mismatch:\n  got:  %x\n  want: %x", got, want)
	}
}

func mustHex(t *testing.T, s string) []byte {
	t.Helper()
	b, err := hex.DecodeString(s)
	if err != nil {
		t.Fatalf("invalid hex in test: %v", err)
	}
	return b
}

// Truncated input must error rather than silently zeroing fields.
func TestCosts_TruncatedInputErrors(t *testing.T) {
	cases := []struct {
		name string
		enc  []byte
		read func(*Decoder) error
	}{
		{"ExecCost", makeSampleExecCost(1).Encode(), func(d *Decoder) error {
			_, err := d.ReadExecCost()
			return err
		}},
		{"IsAuthorizedCost", IsAuthorizedCost{Total: makeSampleExecCost(1)}.Encode(), func(d *Decoder) error {
			_, err := d.ReadIsAuthorizedCost()
			return err
		}},
		{"RefineCost", RefineCost{Total: makeSampleExecCost(1)}.Encode(), func(d *Decoder) error {
			_, err := d.ReadRefineCost()
			return err
		}},
		{"AccumulateCost", AccumulateCost{AccumulateCalls: 1}.Encode(), func(d *Decoder) error {
			_, err := d.ReadAccumulateCost()
			return err
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			for cut := 0; cut < len(tc.enc); cut++ {
				if err := tc.read(NewDecoder(tc.enc[:cut])); err == nil {
					t.Errorf("cut at %d: expected error, got nil", cut)
				}
			}
		})
	}
}
