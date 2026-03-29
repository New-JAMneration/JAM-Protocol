package types

import (
	"testing"
)

// TestGetPutEncoder verifies that pooled encoders are reusable and produce
// correct results across get/put cycles.
func TestGetPutEncoder(t *testing.T) {
	// First cycle: get, encode, put.
	enc1 := GetEncoder()
	ts := TimeSlot(42)
	data1, err := enc1.Encode(&ts)
	if err != nil {
		t.Fatalf("first encode failed: %v", err)
	}
	PutEncoder(enc1)

	// Second cycle: the pool may return the same encoder.
	enc2 := GetEncoder()
	data2, err := enc2.Encode(&ts)
	if err != nil {
		t.Fatalf("second encode failed: %v", err)
	}
	PutEncoder(enc2)

	if len(data1) == 0 || len(data2) == 0 {
		t.Fatal("encoded data should not be empty")
	}

	// Both encodes of the same value must produce identical bytes.
	if string(data1) != string(data2) {
		t.Errorf("pooled encoder produced different results: %x vs %x", data1, data2)
	}
}

// TestPutEncoderClearsHashSegmentMap ensures that PutEncoder resets the
// HashSegmentMap field so stale state is not leaked to the next caller.
func TestPutEncoderClearsHashSegmentMap(t *testing.T) {
	enc := GetEncoder()
	enc.HashSegmentMap = HashSegmentMap{
		OpaqueHash{1}: OpaqueHash{2},
	}
	PutEncoder(enc)

	enc2 := GetEncoder()
	if enc2.HashSegmentMap != nil {
		t.Error("HashSegmentMap should be nil after PutEncoder")
	}
	PutEncoder(enc2)
}
