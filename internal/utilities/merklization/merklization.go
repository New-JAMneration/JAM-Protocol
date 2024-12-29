package merklization

import (
	"errors"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
)

const NODE_SIZE = 512

// bytesToBits converts a byte sequence to a bit sequence.
// The function defined in graypaper 3.7.3. Boolean Values
func bytesToBits(s types.ByteSequence) types.BitSequence {
	// Initialize the bit sequence.
	bitSequence := types.BitSequence{}

	// Iterate over the byte sequence.
	for _, byte := range s {
		// Iterate over the bits in the byte.
		for i := 0; i < 8; i++ {
			// Calculate the bit.
			bit := ((byte >> uint(7-i)) & 1)

			// Append the bit (bool) to the bit sequence.
			bitSequence = append(bitSequence, bit == 1)
		}
	}

	// Return the bit sequence.
	return bitSequence
}

// bitsToBytes converts a BitSequence to a ByteSequence
func bitsToBytes(bits types.BitSequence) (types.ByteSequence, error) {
	if len(bits)%8 != 0 {
		return nil, errors.New("bit sequence length must be a multiple of 8")
	}

	byteLength := len(bits) / 8
	bytes := make(types.ByteSequence, byteLength)

	for i, bit := range bits {
		if bit {
			bytes[i/8] |= 1 << uint(7-i%8)
		}
	}

	return bytes, nil
}

// BranchEncoding encodes a branch node.
func BranchEncoding(left, right types.OpaqueHash) types.BitSequence {
	branchPrefixBit := types.BitSequence{false}
	leftBits := bytesToBits(left[:])[1:]
	rightBits := bytesToBits(right[:])

	encoding := types.BitSequence{}
	encoding = append(encoding, branchPrefixBit...) // 1 bit
	encoding = append(encoding, leftBits...)        // 255 bits
	encoding = append(encoding, rightBits...)       // 256 bits

	return encoding // 512 bits
}

func embeddedValueLeaf(key types.OpaqueHash, value types.ByteSequence) types.BitSequence {
	leftPrefixBit := types.BitSequence{true}                       // 1 bit
	embeddedValueLeafPrefixBit := types.BitSequence{false}         // 1 bit
	prefix := append(leftPrefixBit, embeddedValueLeafPrefixBit...) // 2 bits

	valueSize := types.U32(len(value))
	serializedValueSize := utilities.SerializeFixedLength(valueSize, 1)
	valueSizeBits := bytesToBits(serializedValueSize)

	encoding := types.BitSequence{}
	encoding = append(encoding, prefix...)
	encoding = append(encoding, valueSizeBits[2:]...)
	encoding = append(encoding, bytesToBits(key[:])[:248]...)
	encoding = append(encoding, bytesToBits(value)...)

	// Calculate the size, if it has space left, fill it with 0
	if len(encoding) < NODE_SIZE {
		encoding = append(encoding, make(types.BitSequence, NODE_SIZE-len(encoding))...)
	}

	return encoding
}

func regularLeaf(key types.OpaqueHash, value types.ByteSequence) types.BitSequence {
	leftPrefixBit := types.BitSequence{true}                                    // 1 bit
	regularLeafPrefixBit := types.BitSequence{true}                             // 1 bit
	fillZeroBits := types.BitSequence{false, false, false, false, false, false} // 6 bits

	encoding := types.BitSequence{}
	encoding = append(encoding, leftPrefixBit...)
	encoding = append(encoding, regularLeafPrefixBit...)
	encoding = append(encoding, fillZeroBits...)
	encoding = append(encoding, bytesToBits(key[:])[:248]...)
	valueHash := hash.Blake2bHash(value)
	encoding = append(encoding, bytesToBits(valueHash[:])...)

	return encoding
}

func LeafEncoding(key types.OpaqueHash, value types.ByteSequence) types.BitSequence {
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

// (D.2)
// $T(\sigma)$
// Serialized States
type SerializedState map[types.OpaqueHash]types.ByteSequence

type SerializedStateKeyValue struct {
	key   types.OpaqueHash
	value types.ByteSequence
}

// INFO: Convert the BitSequence to a bitstrings, because we cannot use []bool as a
// key in a map
type MerklizationInput map[string]SerializedStateKeyValue

func Merklization(d MerklizationInput) types.OpaqueHash {
	if len(d) == 0 {
		// zero hash
		return types.OpaqueHash{}
	}

	// FIXME: 為什麼 graypaper 要寫 {(k, v)}, 而不是判斷長度？
	if len(d) == 1 {
		for _, value := range d {
			leftEncoding := LeafEncoding(value.key, value.value)
			bytes, _ := bitsToBytes(leftEncoding)
			return hash.Blake2bHash(bytes)
		}
	}

	l := make(MerklizationInput)
	r := make(MerklizationInput)
	for key, value := range d {
		isLeft := key[0] == '0'
		if isLeft {
			l[key] = value
		}

		isRight := key[0] == '1'
		if isRight {
			r[key] = value
		}
	}

	branchEncoding := BranchEncoding(Merklization(l), Merklization(r))
	bytes, _ := bitsToBytes(branchEncoding)
	return hash.Blake2bHash(bytes)
}

// basic Merklization function
// $M_{\sigma}(\sigma)$
// Input: $T(\sigma)$ a dictionary, serialized states
func MerklizationState(serializedState SerializedState) {
	merklizationInput := make(MerklizationInput)

	for stateKey, stateValue := range serializedState {
		key := bitSequenceToString(bytesToBits(stateKey[:]))
		value := SerializedStateKeyValue{stateKey, stateValue}

		merklizationInput[key] = value
	}

	Merklization(merklizationInput)
}
