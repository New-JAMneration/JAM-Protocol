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

// INFO: Convert the BitSequence to a bitstrings, because we cannot use []bool as a
// key in a map
type MerklizationInput map[string]types.StateKeyVal

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
		isLeft := key[0] == '0'
		if isLeft {
			l[key[1:]] = value
		}

		isRight := key[0] == '1'
		if isRight {
			r[key[1:]] = value
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
		key := bitSequenceToString(utilities.BytesToBits(stateKeyVal.Key[:]))
		value := types.StateKeyVal{
			Key:   stateKeyVal.Key,
			Value: stateKeyVal.Value,
		}

		merklizationInput[key] = value
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
