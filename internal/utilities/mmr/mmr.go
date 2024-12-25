package mmr

import (
	jamtypes "github.com/New-JAMneration/JAM-Protocol/internal/jam_types"
)

// Node represents a node in the Merkle Mountain Range
type Node struct {
	Hash   jamtypes.OpaqueHash
	Height uint64
}

// HashFunction defines the type for hash functions that take a single input
// Input: byte slice to be hashed
// Output: resulting hash as byte slice
type HashFunction func(input jamtypes.ByteSequence) jamtypes.OpaqueHash

// MMR represents the Merkle Mountain Range structure
// From E.2 of the Polkadot Gray Paper:
// "The Merkle mountain range (MMR) is an append-only cryptographic data structure
// which yields a commitment to a sequence of values"
type MMR struct {
	Peaks  []jamtypes.OpaqueHash
	hashFn HashFunction
}

// NewMMR creates a new empty Merkle Mountain Range
func NewMMR(hashFn HashFunction) *MMR {
	if hashFn == nil {
		return nil
	}
	return &MMR{
		Peaks:  make([]jamtypes.OpaqueHash, 0),
		hashFn: hashFn,
	}
}

// concatenateAndHash combines two byte slices and hashes the result
func (m *MMR) concatenateAndHash(left, right jamtypes.OpaqueHash) jamtypes.OpaqueHash {
	leftSlice := left[:]
	rightSlice := right[:]
	return m.hashFn(append(leftSlice, rightSlice...))
}

// R (Replace) function from E.8 in the Gray Paper
// "R: ([T], N, T) → [T]
//
//	(s, i, v) ↦ s' where s' = s except s'i = v"
//
// This function replaces the value at index i with value v in sequence s
func (m *MMR) Replace(sequence []jamtypes.OpaqueHash, index int, value jamtypes.OpaqueHash) []jamtypes.OpaqueHash {
	// Create a new sequence copying the original
	result := make([]jamtypes.OpaqueHash, len(sequence))
	copy(result, sequence)

	// Replace the value at the specified index
	if index < len(sequence) {
		result[index] = value
	}

	return result
}

// P
func (m *MMR) P(peaks []jamtypes.OpaqueHash, l jamtypes.OpaqueHash, n int) []jamtypes.OpaqueHash {
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
	peaks = m.Replace(m.Peaks, n, jamtypes.OpaqueHash{})
	// 3.2 new hash
	newHash := m.concatenateAndHash(current, l)
	// 3.3 next n+1
	return m.P(peaks, newHash, n+1)
}

func (m *MMR) AppendOne(data jamtypes.OpaqueHash) []jamtypes.OpaqueHash {
	newPeaks := m.P(m.Peaks, data, 0)
	m.Peaks = newPeaks
	return newPeaks
}
