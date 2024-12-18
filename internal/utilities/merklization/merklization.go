package merklization

import (
	"golang.org/x/crypto/blake2b"

	jamTypes "github.com/New-JAMneration/JAM-Protocol/internal/jam_types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities"
)

const NODE_SIZE = 512

// bits converts a byte sequence to a bit sequence.
// The function defined in graypaper 3.7.3. Boolean Values
func bits(s jamTypes.ByteSequence) jamTypes.BitSequence {
	// Initialize the bit sequence.
	b := jamTypes.BitSequence{}

	// Iterate over the byte sequence.
	for _, v := range s {
		// Iterate over the bits in the byte.
		for i := 0; i < 8; i++ {
			// Append the bit to the bit sequence.
			boolValue := ((v >> uint(7-i)) & 1) == 1
			b = append(b, boolValue)
		}
	}

	// Return the bit sequence.
	return b
}

// BranchEncoding encodes a branch node.
func BranchEncoding(left, right jamTypes.OpaqueHash) jamTypes.BitSequence {
	var leftByteSequence jamTypes.ByteSequence = left[:]
	var rightByteSequence jamTypes.ByteSequence = right[:]

	branchPrefixBit := jamTypes.BitSequence{false} // prefix with 0, 1 bit
	leftBits := bits(leftByteSequence)             // left, 255 bits
	rightBits := bits(rightByteSequence)           // right, 256 bits

	encoding := append(branchPrefixBit, append(leftBits[1:], rightBits...)...)

	return encoding
}

func embeddedValueLeaf(key jamTypes.OpaqueHash, value jamTypes.ByteSequence) jamTypes.BitSequence {
	var keyByteSequence jamTypes.ByteSequence = key[:]
	var valueByteSequence jamTypes.ByteSequence = value[:]

	leftPrefixBit := jamTypes.BitSequence{true}                    // 1 bit
	embeddedValueLeafPrefixBit := jamTypes.BitSequence{false}      // 1 bit
	prefix := append(leftPrefixBit, embeddedValueLeafPrefixBit...) // 2 bits

	valueSize := jamTypes.U32(len(value))
	serializedValueSize := utilities.SerializeFixedLength(valueSize, 1)
	valueSizeBits := bits(serializedValueSize)

	encoding := jamTypes.BitSequence{}
	encoding = append(encoding, prefix...)
	encoding = append(encoding, valueSizeBits[:6]...)
	encoding = append(encoding, bits(keyByteSequence)[:248]...)
	encoding = append(encoding, bits(valueByteSequence)...)

	// Calculate the size, if it has space left, fill it with 0
	if len(encoding) < NODE_SIZE {
		encoding = append(encoding, make(jamTypes.BitSequence, NODE_SIZE-len(encoding))...)
	}

	return encoding
}

func regularLeaf(key jamTypes.OpaqueHash, value jamTypes.ByteSequence) jamTypes.BitSequence {
	var keyByteSequence jamTypes.ByteSequence = key[:]

	leftPrefixBit := jamTypes.BitSequence{true}                                    // 1 bit
	regularLeafPrefixBit := jamTypes.BitSequence{true}                             // 1 bit
	fillZeroBits := jamTypes.BitSequence{false, false, false, false, false, false} // 6 bits

	encoding := jamTypes.BitSequence{}
	encoding = append(encoding, leftPrefixBit...)
	encoding = append(encoding, regularLeafPrefixBit...)
	encoding = append(encoding, fillZeroBits...)
	encoding = append(encoding, bits(keyByteSequence)[:248]...)
	valueHash := blake2b.Sum256(value)
	encoding = append(encoding, bits(valueHash[:])...)

	return encoding
}

func LeafEncoding(key jamTypes.OpaqueHash, value jamTypes.ByteSequence) jamTypes.BitSequence {
	if len(value) <= 32 {
		return embeddedValueLeaf(key, value)
	} else {
		return regularLeaf(key, value)
	}
}
