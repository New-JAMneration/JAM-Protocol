package shuffle

import (
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	hashUtil "github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
)

// SerializeFixedLength corresponds to E_l in the given specification (C.5).
// It serializes a non-negative integer x into exactly l octets in little-endian order.
// If l=0, returns an empty slice.
// FIXME: Update this function after serialization.go is merged into main.
func SerializeFixedLength(x types.U64, l int) types.ByteSequence {
	if l == 0 {
		return []byte{}
	}
	out := make([]byte, l)
	for i := 0; i < l; i++ {
		out[i] = byte(x & 0xFF)
		x >>= 8
	}
	return out
}

// DeserializeFixedLength deserializes a ByteSequence into a jamtypes.U64.
// FIXME: Update this function after serialization.go is merged into main.
func DeserializeFixedLength(data types.ByteSequence) types.U64 {
	var x types.U64
	for i := len(data) - 1; i >= 0; i-- {
		x <<= 8
		x |= types.U64(data[i])
	}
	return x
}

// numericSequenceFromHash generates a numeric sequence from a hash.
// The function defined in graypaper F.2 $\mathcal{Q}_l$
func numericSequenceFromHash(hash types.OpaqueHash, length types.U32) []types.U32 {
	const serializeLength = 4

	numericSequence := make([]types.U32, length)

	for i := types.U32(0); i < length; i++ {
		floor := i / 8

		// Serialize the floor value
		serializeOutput := SerializeFixedLength(types.U64(floor), serializeLength)

		// Concatenate the hash with the serialized output
		hashOutput := hashUtil.Blake2bHash(types.ByteSequence(append(hash[:], serializeOutput...)))

		// Select a slice of 4 bytes from the hashOutput
		selectRange := types.U32(4)
		startIndex := types.U32((4 * i) % 32)
		hashOutputSlice := hashOutput[startIndex : startIndex+selectRange]

		// Deserialize the hashOutputSlice
		numericValue := types.U32(DeserializeFixedLength(types.ByteSequence(hashOutputSlice)))
		numericSequence[i] = numericValue
	}

	return numericSequence
}

// FisherYatesShuffle is a recursive implementation of the Fisher-Yates shuffle
// algorithm.
// The function requires a sequence of numbers and a sequence of random numbers.
// It returns a shuffled sequence of numbers.
// It's defined in graypaper F.1 $\mathcal{F}$
func FisherYatesShuffle(s []types.U32, r []types.U32) []types.U32 {
	l := len(s)

	// If the sequence is empty, return an empty slice
	if l == 0 {
		return make([]types.U32, 0)
	}

	// Calculate the index
	index := r[0] % types.U32(l)

	// The selected element
	selected := s[index]

	// Swap elements
	s[index], s[l-1] = s[l-1], s[index]

	// Recursively shuffle the remaining elements
	shuffledRest := FisherYatesShuffle(s[:l-1], r[1:])

	// Return the shuffled sequence
	return append([]types.U32{selected}, shuffledRest...)
}

// Shuffle is the main function that shuffles a sequence of numbers.
// The function requires a sequence of numbers and a hash.
// The hash as a seed to generate a numeric sequence.
// It returns a shuffled sequence of numbers.
// It's defined in graypaper F.3 $\mathcal{F}$
func Shuffle(s []types.U32, hash types.OpaqueHash) []types.U32 {
	length := types.U32(len(s))
	numericSequence := numericSequenceFromHash(hash, length)

	return FisherYatesShuffle(s, numericSequence)
}
