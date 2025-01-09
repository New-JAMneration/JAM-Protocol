package mmr

import (
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities"
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
	Peaks  []types.MmrPeak
	hashFn HashFunction
}

// NewMMR creates a new empty Merkle Mountain Range
func NewMMR(hashFn HashFunction) *MMR {
	if hashFn == nil {
		return nil
	}
	return &MMR{
		Peaks:  make([]types.MmrPeak, 0),
		hashFn: hashFn,
	}
}

// NewMMR creates a new empty Merkle Mountain Range
func NewMMRFromPeaks(peaks []types.MmrPeak, hashFn HashFunction) *MMR {
	if hashFn == nil {
		return nil
	}
	return &MMR{
		Peaks:  peaks,
		hashFn: hashFn,
	}
}

// concatenateAndHash combines two byte slices and hashes the result
func (m *MMR) concatenateAndHash(left, right types.MmrPeak) types.MmrPeak {
	leftBytes := [32]byte(*left)
	rightBytes := [32]byte(*right)
	val := m.hashFn(append(leftBytes[:], rightBytes[:]...))
	return &val
}

// R (Replace) function from E.8 in the Gray Paper
// "R: ([T], N, T) → [T]
//
//	(s, i, v) ↦ s' where s' = s except s'i = v"
//
// This function replaces the value at index i with value v in sequence s
func (m *MMR) Replace(sequence []types.MmrPeak, index int, value types.MmrPeak) []types.MmrPeak {
	// Create a new sequence copying the original
	result := make([]types.MmrPeak, len(sequence))
	copy(result, sequence)

	// Replace the value at the specified index
	if index < len(sequence) {
		result[index] = value
	}

	return result
}

// P
func (m *MMR) P(peaks []types.MmrPeak, l types.MmrPeak, n int) []types.MmrPeak {
	// if n >= l
	if n >= len(m.Peaks) {
		return append(m.Peaks, l)
	}

	// 2. if peaks[n] is empty
	if m.Peaks[n] == nil {
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

func (m *MMR) AppendOne(data types.MmrPeak) []types.MmrPeak {
	if data == nil || len(*data) == 0 {
		return m.Peaks
	}
	newPeaks := m.P(m.Peaks, data, 0)
	m.Peaks = newPeaks
	return newPeaks
}

// MmrWrapper converts an external types.Mmr into your internal MMR structure.
func MmrWrapper(ext *types.Mmr, hashFn HashFunction) *MMR {
	if ext == nil || hashFn == nil {
		return nil
	}

	// Build internal MMR
	return NewMMRFromPeaks(ext.Peaks, hashFn)
}

func (m *MMR) Serialize() types.ByteSequence {
	serialItems := []utilities.Serializable{}

	for _, peak := range m.Peaks {
		// empty
		if peak == nil || len(*peak) == 0 {
			serialItems = append(serialItems, utilities.U64Wrapper{})
		} else {
			serialItems = append(serialItems, utilities.SerializableSequence{utilities.U64Wrapper{Value: 1}, utilities.ByteArray32Wrapper{Value: types.ByteArray32(*peak)}})
		}
	}

	disc := utilities.Discriminator{
		Value: serialItems,
	}

	return disc.Serialize()
}
