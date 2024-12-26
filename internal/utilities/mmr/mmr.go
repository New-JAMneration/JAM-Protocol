package mmr

import (
	"log"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

// Node represents a node in the Merkle Mountain Range
type Node struct {
	Hash   types.OpaqueHash
	Height uint64
}

// HashFunction defines the type for hash functions that take a single input
// Input: byte slice to be hashed
// Output: resulting hash as byte slice
type HashFunction func(input types.ByteSequence) types.OpaqueHash

// MMR represents the Merkle Mountain Range structure
// From E.2 of the Polkadot Gray Paper:
// "The Merkle mountain range (MMR) is an append-only cryptographic data structure
// which yields a commitment to a sequence of values"
type MMR struct {
	Peaks  []*types.OpaqueHash
	hashFn HashFunction
}

// NewMMR creates a new empty Merkle Mountain Range
func NewMMR(hashFn HashFunction) *MMR {
	if hashFn == nil {
		return nil
	}
	return &MMR{
		Peaks:  make([]*types.OpaqueHash, 0),
		hashFn: hashFn,
	}
}

// NewMMR creates a new empty Merkle Mountain Range
func NewMMRFromPeaks(peaks []*types.OpaqueHash, hashFn HashFunction) *MMR {
	if hashFn == nil {
		return nil
	}
	return &MMR{
		Peaks:  peaks,
		hashFn: hashFn,
	}
}

// concatenateAndHash combines two byte slices and hashes the result
func (m *MMR) concatenateAndHash(left, right *types.OpaqueHash) *types.OpaqueHash {
	leftSlice := left[:]
	rightSlice := right[:]
	val := m.hashFn(append(leftSlice, rightSlice...))
	return &val
}

// R (Replace) function from E.8 in the Gray Paper
// "R: ([T], N, T) → [T]
//
//	(s, i, v) ↦ s' where s' = s except s'i = v"
//
// This function replaces the value at index i with value v in sequence s
func (m *MMR) Replace(sequence []*types.OpaqueHash, index int, value *types.OpaqueHash) []*types.OpaqueHash {
	// Create a new sequence copying the original
	result := make([]*types.OpaqueHash, len(sequence))
	copy(result, sequence)

	// Replace the value at the specified index
	if index < len(sequence) {
		result[index] = value
	}

	return result
}

// P
func (m *MMR) P(peaks []*types.OpaqueHash, l *types.OpaqueHash, n int) []*types.OpaqueHash {
	// if n >= l
	if n >= len(m.Peaks) {
		return append(m.Peaks, l)
	}

	// 2. if peaks[n] is empty
	if len(m.Peaks[n]) == 0 {
		return m.Replace(m.Peaks, n, l)
	}

	// 3.
	current := m.Peaks[n]
	// 3.1 clean the position n
	peaks = m.Replace(m.Peaks, n, nil)
	// 3.2 new hash
	newHash := m.concatenateAndHash(current, l)
	// 3.3 next n+1
	return m.P(peaks, newHash, n+1)
}

func (m *MMR) AppendOne(data *types.OpaqueHash) []*types.OpaqueHash {
	if data == nil {
		return m.Peaks
	}
	newPeaks := m.P(m.Peaks, data, 0)
	m.Peaks = newPeaks
	return newPeaks
}

// fromMmrPeaks converts external MmrPeak slices into your jamtypes.OpaqueHash slices.
func fromMmrPeaks(peaks []types.MmrPeak) []*types.OpaqueHash {
	result := make([]*types.OpaqueHash, len(peaks))
	for i, peak := range peaks {
		log.Printf("peak: %v", peak)
		if peak != nil {
			// peak is a pointer to OpaqueHash
			// OpaqueHash itself could be []byte or [32]byte depending on how jamtypes defines it
			result[i] = peak
		} else {
			// Decide how you want to handle nil peaks:
			result[i] = nil // or jamtypes.OpaqueHash{} if you prefer an empty hash
		}
	}
	return result
}

// MmrWrapper converts an external types.Mmr into your internal MMR structure.
func MmrWrapper(ext *types.Mmr, hashFn HashFunction) *MMR {
	if ext == nil || hashFn == nil {
		return nil
	}

	// Build internal MMR
	return NewMMRFromPeaks(fromMmrPeaks(ext.Peaks), hashFn)
}
