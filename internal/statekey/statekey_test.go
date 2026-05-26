package statekey

import (
	"testing"

	"golang.org/x/crypto/blake2b"
)

func TestInterleave_Layout(t *testing.T) {
	serviceID := uint32(700)
	preimage := []byte{
		0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef,
		0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef,
		0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef,
		0x01, 0x23, 0x45,
	}

	out := Interleave(serviceID, preimage)

	n := [4]byte{
		byte(serviceID),
		byte(serviceID >> 8),
		byte(serviceID >> 16),
		byte(serviceID >> 24),
	}
	digest := blake2b.Sum256(preimage)

	// Verify interleave layout: [n0, h0, n1, h1, n2, h2, n3, h3, h4, ..., h26]
	for i := 0; i < 4; i++ {
		if out[2*i] != n[i] {
			t.Errorf("position %d: expected n[%d]=0x%02x, got 0x%02x", 2*i, i, n[i], out[2*i])
		}
		if out[2*i+1] != digest[i] {
			t.Errorf("position %d: expected h[%d]=0x%02x, got 0x%02x", 2*i+1, i, digest[i], out[2*i+1])
		}
	}
	for i := 4; i < 27; i++ {
		if out[i+4] != digest[i] {
			t.Errorf("position %d: expected h[%d]=0x%02x, got 0x%02x", i+4, i, digest[i], out[i+4])
		}
	}
}

func TestStorage_Prefix(t *testing.T) {
	rawKey := []byte{0xAA, 0xBB}
	out := Storage(42, rawKey)

	expectedPreimage := make([]byte, 4+len(rawKey))
	expectedPreimage[0] = 0xFF
	expectedPreimage[1] = 0xFF
	expectedPreimage[2] = 0xFF
	expectedPreimage[3] = 0xFF
	copy(expectedPreimage[4:], rawKey)

	expected := Interleave(42, expectedPreimage)
	if out != expected {
		t.Errorf("Storage: expected %x, got %x", expected, out)
	}
}

func TestPreimageMeta_Prefix(t *testing.T) {
	var hash [32]byte
	for i := range hash {
		hash[i] = byte(i)
	}
	length := uint32(100)
	out := PreimageMeta(42, hash, length)

	expectedPreimage := make([]byte, 4+32)
	expectedPreimage[0] = byte(length)
	expectedPreimage[1] = byte(length >> 8)
	expectedPreimage[2] = byte(length >> 16)
	expectedPreimage[3] = byte(length >> 24)
	copy(expectedPreimage[4:], hash[:])

	expected := Interleave(42, expectedPreimage)
	if out != expected {
		t.Errorf("PreimageMeta: expected %x, got %x", expected, out)
	}
}

func TestPreimageLookup_Prefix(t *testing.T) {
	var hash [32]byte
	for i := range hash {
		hash[i] = byte(i + 10)
	}
	out := PreimageLookup(42, hash)

	expectedPreimage := make([]byte, 4+32)
	expectedPreimage[0] = 0xFE
	expectedPreimage[1] = 0xFF
	expectedPreimage[2] = 0xFF
	expectedPreimage[3] = 0xFF
	copy(expectedPreimage[4:], hash[:])

	expected := Interleave(42, expectedPreimage)
	if out != expected {
		t.Errorf("PreimageLookup: expected %x, got %x", expected, out)
	}
}

func TestDeterministic(t *testing.T) {
	key1 := Storage(100, []byte{1, 2, 3})
	key2 := Storage(100, []byte{1, 2, 3})
	if key1 != key2 {
		t.Error("same inputs must produce same output")
	}

	key3 := Storage(100, []byte{1, 2, 4})
	if key1 == key3 {
		t.Error("different inputs should produce different output")
	}
}
