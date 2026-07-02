package types

import (
	"reflect"
	"testing"
)

// TestEncodeAvailabilityAssignment_V080Guarantee covers the GP v0.8.0
// eq:reportingstate change: each rho entry stores the FULL guarantee
// (report ++ E4(slot) ++ var(credential)) followed by the E4 assignment
// timestamp; v0.7.x stored only the bare work report.
func TestEncodeAvailabilityAssignment_V080Guarantee(t *testing.T) {
	assignment := AvailabilityAssignment{
		Guarantee: ReportGuarantee{
			Report: WorkReport{
				PackageSpec: WorkPackageSpec{Hash: WorkPackageHash{0x11}},
				CoreIndex:   1,
			},
			Slot: 0x01020304,
			Signatures: []ValidatorSignature{
				{ValidatorIndex: 2, Signature: Ed25519Signature{0x22}},
			},
		},
		AssignedSlot: 0x0A0B0C0D,
	}

	encoder := NewEncoder()
	encoded, err := encoder.Encode(&assignment)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}

	// The assignment timestamp (0x0A0B0C0D little-endian) is the trailing E4.
	n := len(encoded)
	if n < 4 || encoded[n-4] != 0x0D || encoded[n-3] != 0x0C || encoded[n-2] != 0x0B || encoded[n-1] != 0x0A {
		t.Errorf("trailing assigned-slot bytes = % x, want 0d 0c 0b 0a", encoded[n-4:])
	}
	// The credential precedes it: signature(64) preceded by E2(index=2),
	// preceded by the var length prefix (1). The v0.7.x layout had none of
	// this between the report and the timestamp.
	sigStart := n - 4 - 64
	if encoded[sigStart] != 0x22 {
		t.Errorf("credential signature[0] = %x, want 22", encoded[sigStart])
	}
	if idx := sigStart - 2; encoded[idx] != 2 || encoded[idx+1] != 0 {
		t.Errorf("credential validator index bytes = % x, want 02 00", encoded[idx:idx+2])
	}
	if prefix := sigStart - 3; encoded[prefix] != 1 {
		t.Errorf("credential length prefix = %d, want 1", encoded[prefix])
	}
	// The guarantee slot (0x01020304 little-endian) precedes the credential.
	if off := sigStart - 3 - 4; encoded[off] != 0x04 || encoded[off+1] != 0x03 ||
		encoded[off+2] != 0x02 || encoded[off+3] != 0x01 {
		t.Errorf("guarantee slot bytes = % x, want 04 03 02 01", encoded[off:off+4])
	}

	decoder := NewDecoder()
	var got AvailabilityAssignment
	if err := decoder.Decode(encoded, &got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.Guarantee.Report.PackageSpec.Hash != assignment.Guarantee.Report.PackageSpec.Hash {
		t.Errorf("report hash mismatch after round-trip")
	}
	if got.Guarantee.Slot != assignment.Guarantee.Slot {
		t.Errorf("guarantee slot = %d, want %d", got.Guarantee.Slot, assignment.Guarantee.Slot)
	}
	if !reflect.DeepEqual(got.Guarantee.Signatures, assignment.Guarantee.Signatures) {
		t.Errorf("credential mismatch after round-trip:\n got %+v\nwant %+v", got.Guarantee.Signatures, assignment.Guarantee.Signatures)
	}
	if got.AssignedSlot != assignment.AssignedSlot {
		t.Errorf("assigned slot = %d, want %d", got.AssignedSlot, assignment.AssignedSlot)
	}
}
