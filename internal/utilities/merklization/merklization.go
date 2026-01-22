package merklization

import (
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
)

const NODE_SIZE = 512

// BranchEncoding encodes a branch node.
func BranchEncoding(left, right types.OpaqueHash) types.BitSequence {
	branchPrefixBit := types.BitSequence{false}
	leftBits := utilities.BytesToBits(left[:])[1:]
	rightBits := utilities.BytesToBits(right[:])

	encoding := types.BitSequence{}
	encoding = append(encoding, branchPrefixBit...) // 1 bit
	encoding = append(encoding, leftBits...)        // 255 bits
	encoding = append(encoding, rightBits...)       // 256 bits

	return encoding // 512 bits
}

func embeddedValueLeaf(key types.StateKey, value types.ByteSequence) types.BitSequence {
	leftPrefixBit := types.BitSequence{true}                       // 1 bit
	embeddedValueLeafPrefixBit := types.BitSequence{false}         // 1 bit
	prefix := append(leftPrefixBit, embeddedValueLeafPrefixBit...) // 2 bits

	valueSize := types.U32(len(value))
	serializedValueSize := utilities.SerializeFixedLength(valueSize, 1)
	valueSizeBits := utilities.BytesToBits(serializedValueSize)

	encoding := types.BitSequence{}
	encoding = append(encoding, prefix...)
	encoding = append(encoding, valueSizeBits[2:]...)
	encoding = append(encoding, utilities.BytesToBits(key[:])...)
	encoding = append(encoding, utilities.BytesToBits(value)...)

	// Calculate the size, if it has space left, fill it with 0
	if len(encoding) < NODE_SIZE {
		encoding = append(encoding, make(types.BitSequence, NODE_SIZE-len(encoding))...)
	}

	return encoding
}

func regularLeaf(key types.StateKey, value types.ByteSequence) types.BitSequence {
	leftPrefixBit := types.BitSequence{true}                                    // 1 bit
	regularLeafPrefixBit := types.BitSequence{true}                             // 1 bit
	fillZeroBits := types.BitSequence{false, false, false, false, false, false} // 6 bits

	encoding := types.BitSequence{}
	encoding = append(encoding, leftPrefixBit...)
	encoding = append(encoding, regularLeafPrefixBit...)
	encoding = append(encoding, fillZeroBits...)
	encoding = append(encoding, utilities.BytesToBits(key[:])...)
	valueHash := hash.Blake2bHash(value)
	encoding = append(encoding, utilities.BytesToBits(valueHash[:])...)

	return encoding
}

func LeafEncoding(key types.StateKey, value types.ByteSequence) types.BitSequence {
	if len(value) <= 32 {
		return embeddedValueLeaf(key, value)
	} else {
		return regularLeaf(key, value)
	}
}

// Convert a types.BitSequence to a string
func bitSequenceToString(bitSequence types.BitSequence) string {
	str := ""
	for _, bit := range bitSequence {
		if bit {
			str += "1"
		} else {
			str += "0"
		}
	}
	return str
}

func BitSequenceToString(bitSequence types.BitSequence) string {
	return bitSequenceToString(bitSequence)
}

type MerklizationInput map[[31]byte]types.StateKeyVal

// Shift the key left by 1 bit
func shiftKeyLeft(key [31]byte) [31]byte {
	var result [31]byte
	for i := 0; i < 30; i++ {
		result[i] = (key[i] << 1) | (key[i+1] >> 7)
	}
	result[30] = key[30] << 1
	return result
}

func Merklization(d MerklizationInput) types.OpaqueHash {
	if len(d) == 0 {
		// zero hash
		return types.OpaqueHash{}
	}

	// FIXME: 為什麼 graypaper 要寫 {(k, v)}, 而不是判斷長度？
	if len(d) == 1 {
		for _, stateKeyVal := range d {
			leftEncoding := LeafEncoding(stateKeyVal.Key, stateKeyVal.Value)
			bytes, _ := utilities.BitsToBytes(leftEncoding)
			return hash.Blake2bHash(bytes)
		}
	}

	l := make(MerklizationInput)
	r := make(MerklizationInput)
	for key, value := range d {
		// check the first bit: 0 -> left, 1 -> right
		firstBit := (key[0] & 0x80) == 0

		shiftedKey := shiftKeyLeft(key)

		if firstBit {
			l[shiftedKey] = value
		} else {
			r[shiftedKey] = value
		}
	}

	branchEncoding := BranchEncoding(Merklization(l), Merklization(r))
	bytes, _ := utilities.BitsToBytes(branchEncoding)
	return hash.Blake2bHash(bytes)
}

// basic Merklization function
// $M_{\sigma}(\sigma)$
// Input: $T(\sigma)$ a dictionary, serialized states
func MerklizationSerializedState(serializedState types.StateKeyVals) types.StateRoot {
	merklizationInput := make(MerklizationInput)

	// Convert the StateKeyVals to merklization input
	for _, stateKeyVal := range serializedState {
		merklizationInput[stateKeyVal.Key] = stateKeyVal
	}

	return types.StateRoot(Merklization(merklizationInput))
}

// MerklizationState is a function that takes a state and returns the
// Merklization of the state.
// (D.5)
func MerklizationState(state types.State) types.StateRoot {
	// serializedState, err := StateSerialize(state)
	serializedState, _ := StateEncoder(state)

	return MerklizationSerializedState(serializedState)
}
