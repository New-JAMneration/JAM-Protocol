package mmr

import (
	"bytes"
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
)

func HelperFromM(mmrPeak types.MmrPeak) types.OpaqueHash {
	if mmrPeak == nil {
		return types.OpaqueHash{}
	}
	return types.OpaqueHash(*mmrPeak)
}
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

	original := []types.MmrPeak{&hashA, nil, nil, &hashB}
	// Replace the second index (which is empty) with hashB
	newSeq := m.Replace(original, 1, &hashB)

	if len(newSeq) != len(original) {
		t.Errorf("expected length %d, got %d", len(original), len(newSeq))
	}

	// Check if index 1 got replaced
	if *newSeq[1] != hashB {
		t.Errorf("expected newSeq[1] to be hashB, got %x", newSeq[1])
	}

	// Ensure other indices remained unchanged
	if *newSeq[0] != hashA {
		t.Errorf("expected newSeq[0] to remain hashA, got %x", newSeq[0])
	}
	if *newSeq[3] != hashB {
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
	m.AppendOne(&hashA)
	if len(m.Peaks) != 1 {
		t.Errorf("expected 1 peak after appending 1 item, got %d", len(m.Peaks))
	}
	if *m.Peaks[0] != hashA {
		t.Errorf("expected peaks[0] to be hashA, got %v", m.Peaks)
	}

	// Append hashB (may merge or expand depending on the logic in P)
	m.AppendOne(&hashB)
	if len(m.Peaks) < 1 {
		t.Fatalf("expected at least 1 peak, got %d", len(m.Peaks))
	}

	// We at least know the MMR has updated peaks
	foundAB := false
	for _, peak := range m.Peaks {
		if *peak != hash.Blake2bHash(types.ByteSequence(append(hashA[:], hashB[:]...))) {
			foundAB = true
			break
		}
	}
	if !foundAB {
		t.Errorf("did not find hashB in peaks after AppendOne(hashB)")
	}
}

// TestImportMmr checks that ImportMmr correctly maps the external Mmr into our internal MMR.
func TestImportMmr(t *testing.T) {
	// 1. Define some test peaks:
	peak1 := hash.Blake2bHash([]byte("Alice"))
	peak2 := hash.Blake2bHash([]byte("Bob"))

	// 2. Create an external Mmr with three peaks (with one nil to test behavior).
	extMmr := &types.Mmr{
		Peaks: []types.MmrPeak{
			&peak1,
			nil,
			&peak2,
		},
	}

	// 4. Convert the external Mmr to your internal MMR.
	myMmr := MmrWrapper(extMmr, hash.Blake2bHash)
	if myMmr == nil {
		t.Fatalf("Expected non-nil MMR, got nil")
	}

	// 5. Check that the new MMR has the correct number of peaks.
	if len(myMmr.Peaks) != 3 {
		t.Fatalf("Expected 3 peaks, got %d", len(myMmr.Peaks))
	}

	// 6. Compare each peak:
	if *myMmr.Peaks[0] != peak1 {
		t.Errorf("Expected myMmr.Peaks[0] = %v, got %v", peak1, myMmr.Peaks[0])
	}
	if myMmr.Peaks[1] != nil {
		t.Errorf("Expected myMmr.Peaks[1] = nil, got %v", myMmr.Peaks[1])
	}
	if *myMmr.Peaks[2] != peak2 {
		t.Errorf("Expected myMmr.Peaks[2] = %v, got %v", peak2, myMmr.Peaks[2])
	}
}

func TestMMR_Serialize(t *testing.T) {
	tests := []struct {
		name     string
		peaks    []types.MmrPeak
		expected types.ByteSequence
	}{
		{
			name:     "Empty Peaks",
			peaks:    []types.MmrPeak{},
			expected: utilities.Discriminator{Value: []utilities.Serializable{}}.Serialize(),
		},
		{
			name:  "Peaks with Nil pointer Arrays",
			peaks: []types.MmrPeak{nil, nil},
			expected: utilities.Discriminator{
				Value: []utilities.Serializable{
					utilities.U64Wrapper{},
					utilities.U64Wrapper{},
				},
			}.Serialize(),
		},
		{
			name: "Peaks with Non-Empty Arrays",
			peaks: []types.MmrPeak{
				{0x1, 0x2},
				{0x3, 0x4, 0x5},
			},
			expected: utilities.Discriminator{
				Value: []utilities.Serializable{
					utilities.SerializableSequence{
						utilities.U64Wrapper{Value: 1},
						utilities.ByteArray32Wrapper{Value: types.ByteArray32{0x1, 0x2}},
					},
					utilities.SerializableSequence{
						utilities.U64Wrapper{Value: 1},
						utilities.ByteArray32Wrapper{Value: types.ByteArray32{0x3, 0x4, 0x5}},
					},
				},
			}.Serialize(),
		},
		{
			name: "Peaks with Mixed Arrays",
			peaks: []types.MmrPeak{
				{0x1, 0x2},
				nil,
				{0x3, 0x4, 0x5},
			},
			expected: utilities.Discriminator{
				Value: []utilities.Serializable{
					utilities.SerializableSequence{
						utilities.U64Wrapper{Value: 1},
						utilities.ByteArray32Wrapper{Value: types.ByteArray32{0x1, 0x2}},
					},
					utilities.U64Wrapper{},
					utilities.SerializableSequence{
						utilities.U64Wrapper{Value: 1},
						utilities.ByteArray32Wrapper{Value: types.ByteArray32{0x3, 0x4, 0x5}},
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
				t.Errorf("Serialize output mismatch, expected %v, got %v", tt.expected, actual)
			}
		})
	}
}

// TestSuperPeak exercises the MMR.SuperPeak function.
func TestSuperPeak(t *testing.T) {
	// Create a simple MMR instance that uses Keccak as its hashing function.
	newTestMMR := func() *MMR {
		return &MMR{
			Peaks: []types.MmrPeak{},
			// If your MMR normally uses m.hashFn, just wrap hash.KeccakHash here:
			// i.e., hashFn: func(input types.ByteSequence) types.OpaqueHash { ... }
			hashFn: func(input types.ByteSequence) types.OpaqueHash {
				h := hash.Blake2bHash(input)
				return h
			},
		}
	}

	t.Run("no peaks", func(t *testing.T) {
		m := newTestMMR()
		got := m.SuperPeak([]types.MmrPeak{})
		expected := types.OpaqueHash{}
		if !bytes.Equal(got[:], expected[:]) {
			t.Errorf("SuperPeak([]) = %v; want [32]byte{}", *got)
		}
	})

	t.Run("single peak", func(t *testing.T) {
		m := newTestMMR()

		// Make a “peak” with some recognizable bytes:
		var p1 types.OpaqueHash
		copy(p1[:], []byte("peak-1"))

		peaks := []types.MmrPeak{&p1}
		got := m.SuperPeak(peaks)

		if got == nil {
			t.Fatalf("SuperPeak([p1]) = nil; want the peak itself")
		}
		if !bytes.Equal(p1[:], (*got)[:]) {
			t.Errorf("SuperPeak([p1]) = %x; want %x", *got, p1)
		}
	})

	t.Run("two peaks", func(t *testing.T) {
		m := newTestMMR()

		var p1, p2 types.OpaqueHash
		copy(p1[:], []byte("peak-1"))
		copy(p2[:], []byte("peak-2"))

		peaks := []types.MmrPeak{&p1, &p2}
		got := m.SuperPeak(peaks)

		// Manually compute the expected:
		//   seq = "peak" + p1 + p2
		seq := []byte("peak")
		seq = append(seq, p1[:]...)
		seq = append(seq, p2[:]...)
		want := hash.KeccakHash(seq)

		if got == nil {
			t.Fatalf("SuperPeak([p1,p2]) = nil; want a 32-byte hash")
		}
		if !bytes.Equal(want[:], (*got)[:]) {
			t.Errorf("SuperPeak([p1,p2]) = %x; want %x", *got, want)
		}
	})

	t.Run("three peaks", func(t *testing.T) {
		m := newTestMMR()

		// p1, p2, p3
		var p1, p2, p3 types.OpaqueHash
		copy(p1[:], []byte("peak-1"))
		copy(p2[:], []byte("peak-2"))
		copy(p3[:], []byte("peak-3"))

		peaks := []types.MmrPeak{&p1, &p2, &p3}
		got := m.SuperPeak(peaks)

		// Let’s do a manual fold:
		// partial = hash("peak"+p1+p2)
		seqPartial := []byte("peak")
		seqPartial = append(seqPartial, p1[:]...)
		seqPartial = append(seqPartial, p2[:]...)
		partial := hash.KeccakHash(seqPartial)

		// final = hash("peak"+partial+p3)
		seqFinal := []byte("peak")
		seqFinal = append(seqFinal, partial[:]...)
		seqFinal = append(seqFinal, p3[:]...)
		want := hash.KeccakHash(seqFinal)

		if got == nil {
			t.Fatalf("SuperPeak([p1,p2,p3]) = nil; want 32-byte hash")
		}
		if !bytes.Equal(want[:], (*got)[:]) {
			t.Errorf("SuperPeak([p1,p2,p3]) = %x; want %x", *got, want)
		}
	})
}
