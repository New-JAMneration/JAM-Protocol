package mmr

import (
	"bytes"
	"testing"

	jamtypes "github.com/New-JAMneration/JAM-Protocol/internal/jam_types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
)

func TestNewMMR(t *testing.T) {
	t.Run("nil hash function returns nil MMR", func(t *testing.T) {
		m := NewMMR(nil)
		if m != nil {
			t.Errorf("expected nil MMR when hashFn is nil, got %v", m)
		}
	})

	t.Run("valid hash function returns non-nil MMR", func(t *testing.T) {
		m := NewMMR(hash.Blake2bHash)
		if m == nil {
			t.Errorf("expected non-nil MMR, got nil")
		}
	})
}

func TestMMR_Replace(t *testing.T) {
	hashA := hash.Blake2bHash([]byte("A"))
	hashB := hash.Blake2bHash([]byte("B"))

	m := NewMMR(hash.Blake2bHash)
	if m == nil {
		t.Fatal("failed to create MMR")
	}

	original := []jamtypes.OpaqueHash{hashA, {}, {}, hashB}
	// Replace the second index (which is empty) with hashB
	newSeq := m.Replace(original, 1, hashB)

	if len(newSeq) != len(original) {
		t.Errorf("expected length %d, got %d", len(original), len(newSeq))
	}

	// Check if index 1 got replaced
	if !bytes.Equal(newSeq[1], hashB) {
		t.Errorf("expected newSeq[1] to be hashB, got %x", newSeq[1])
	}

	// Ensure other indices remained unchanged
	if !bytes.Equal(newSeq[0], hashA) {
		t.Errorf("expected newSeq[0] to remain hashA, got %x", newSeq[0])
	}
	if !bytes.Equal(newSeq[3], hashA) {
		t.Errorf("expected newSeq[3] to remain hashB, got %x", newSeq[3])
	}
}

func TestMMR_AppendOne(t *testing.T) {
	m := NewMMR(hash.Blake2bHash)
	if m == nil {
		t.Fatal("failed to create MMR")
	}

	hashA := hash.Blake2bHash([]byte("Alice"))
	hashB := hash.Blake2bHash([]byte("Bob"))

	// Append hashA
	m.AppendOne(hashA)
	if len(m.Peaks) != 1 {
		t.Errorf("expected 1 peak after appending 1 item, got %d", len(m.Peaks))
	}
	if !bytes.Equal(m.Peaks[0], hashA) {
		t.Errorf("expected peaks[0] to be hashA, got %x", m.Peaks)
	}

	// Append hashB (may merge or expand depending on the logic in P)
	m.AppendOne(hashB)
	if len(m.Peaks) < 1 {
		t.Fatalf("expected at least 1 peak, got %d", len(m.Peaks))
	}

	// We at least know the MMR has updated peaks
	foundAB := false
	// Preallocate slice with the required size to avoid multiple allocations
	target := make([]byte, len(hashA)+len(hashB))
	copy(target, hashA)
	copy(target[len(hashA):], hashB)

	targetHash := hash.Blake2bHash(target)
	for _, peak := range m.Peaks {
		if bytes.Equal(peak, targetHash) {
			foundAB = true
			break
		}
	}
	if !foundAB {
		t.Errorf("did not find hashB in peaks after AppendOne(hashB)")
	}
}

func TestMMR_Serialize(t *testing.T) {
	tests := []struct {
		name     string
		peaks    []jamtypes.OpaqueHash
		expected jamtypes.ByteSequence
	}{
		{
			name:     "Empty Peaks",
			peaks:    []jamtypes.OpaqueHash{},
			expected: utilities.Discriminator{Value: []utilities.Serializable{}}.Serialize(),
		},
		{
			name:  "Peaks with Empty Arrays",
			peaks: []jamtypes.OpaqueHash{{}, {}},
			expected: utilities.Discriminator{
				Value: []utilities.Serializable{
					utilities.U64Wrapper{},
					utilities.U64Wrapper{},
				},
			}.Serialize(),
		},
		{
			name: "Peaks with Non-Empty Arrays",
			peaks: []jamtypes.OpaqueHash{
				{0x1, 0x2},
				{0x3, 0x4, 0x5},
			},
			expected: utilities.Discriminator{
				Value: []utilities.Serializable{
					utilities.SerializableSequence{
						utilities.U64Wrapper{Value: 1},
						utilities.ByteArray32Wrapper{Value: jamtypes.ByteArray32{0x1, 0x2}},
					},
					utilities.SerializableSequence{
						utilities.U64Wrapper{Value: 1},
						utilities.ByteArray32Wrapper{Value: jamtypes.ByteArray32{0x3, 0x4, 0x5}},
					},
				},
			}.Serialize(),
		},
		{
			name: "Peaks with Mixed Arrays",
			peaks: []jamtypes.OpaqueHash{
				{0x1, 0x2},
				{},
				{0x3, 0x4, 0x5},
			},
			expected: utilities.Discriminator{
				Value: []utilities.Serializable{
					utilities.SerializableSequence{
						utilities.U64Wrapper{Value: 1},
						utilities.ByteArray32Wrapper{Value: jamtypes.ByteArray32{0x1, 0x2}},
					},
					utilities.U64Wrapper{},
					utilities.SerializableSequence{
						utilities.U64Wrapper{Value: 1},
						utilities.ByteArray32Wrapper{Value: jamtypes.ByteArray32{0x3, 0x4, 0x5}},
					},
				},
			}.Serialize(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MMR{Peaks: tt.peaks}

			// Call Serialize
			actual := m.Serialize()

			// Assert equality of the actual and expected results
			if !bytes.Equal(tt.expected, actual) {
				t.Errorf("Serialize output mismatch")
			}
		})
	}
}
