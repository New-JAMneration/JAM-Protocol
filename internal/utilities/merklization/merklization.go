package merklization

import (
	"errors"

	"golang.org/x/crypto/blake2b"

	jamTypes "github.com/New-JAMneration/JAM-Protocol/internal/jam_types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities"
)

const NODE_SIZE = 512

// bytesToBits converts a byte sequence to a bit sequence.
// The function defined in graypaper 3.7.3. Boolean Values
func bytesToBits(s jamTypes.ByteSequence) jamTypes.BitSequence {
	// Initialize the bit sequence.
	bitSequence := jamTypes.BitSequence{}

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
func bitsToBytes(bits jamTypes.BitSequence) (jamTypes.ByteSequence, error) {
	if len(bits)%8 != 0 {
		return nil, errors.New("bit sequence length must be a multiple of 8")
	}

	byteLength := len(bits) / 8
	bytes := make(jamTypes.ByteSequence, byteLength)

	for i, bit := range bits {
		if bit {
			bytes[i/8] |= 1 << uint(7-i%8)
		}
	}

	return bytes, nil
}

// BranchEncoding encodes a branch node.
func BranchEncoding(left, right jamTypes.OpaqueHash) jamTypes.BitSequence {
	branchPrefixBit := jamTypes.BitSequence{false}
	leftBits := bytesToBits(left[:])[1:]
	rightBits := bytesToBits(right[:])

	encoding := jamTypes.BitSequence{}
	encoding = append(encoding, branchPrefixBit...) // 1 bit
	encoding = append(encoding, leftBits...)        // 255 bits
	encoding = append(encoding, rightBits...)       // 256 bits

	return encoding // 512 bits
}

func embeddedValueLeaf(key jamTypes.OpaqueHash, value jamTypes.ByteSequence) jamTypes.BitSequence {
	leftPrefixBit := jamTypes.BitSequence{true}                    // 1 bit
	embeddedValueLeafPrefixBit := jamTypes.BitSequence{false}      // 1 bit
	prefix := append(leftPrefixBit, embeddedValueLeafPrefixBit...) // 2 bits

	valueSize := jamTypes.U32(len(value))
	serializedValueSize := utilities.SerializeFixedLength(valueSize, 1)
	valueSizeBits := bytesToBits(serializedValueSize)

	encoding := jamTypes.BitSequence{}
	encoding = append(encoding, prefix...)
	encoding = append(encoding, valueSizeBits[:6]...)
	encoding = append(encoding, bytesToBits(key[:])[:248]...)
	encoding = append(encoding, bytesToBits(value)...)

	// Calculate the size, if it has space left, fill it with 0
	if len(encoding) < NODE_SIZE {
		encoding = append(encoding, make(jamTypes.BitSequence, NODE_SIZE-len(encoding))...)
	}

	return encoding
}

func regularLeaf(key jamTypes.OpaqueHash, value jamTypes.ByteSequence) jamTypes.BitSequence {
	leftPrefixBit := jamTypes.BitSequence{true}                                    // 1 bit
	regularLeafPrefixBit := jamTypes.BitSequence{true}                             // 1 bit
	fillZeroBits := jamTypes.BitSequence{false, false, false, false, false, false} // 6 bits

	encoding := jamTypes.BitSequence{}
	encoding = append(encoding, leftPrefixBit...)
	encoding = append(encoding, regularLeafPrefixBit...)
	encoding = append(encoding, fillZeroBits...)
	encoding = append(encoding, bytesToBits(key[:])[:248]...)
	valueHash := blake2b.Sum256(value)
	encoding = append(encoding, bytesToBits(valueHash[:])...)

	return encoding
}

func LeafEncoding(key jamTypes.OpaqueHash, value jamTypes.ByteSequence) jamTypes.BitSequence {
	if len(value) <= 32 {
		return embeddedValueLeaf(key, value)
	} else {
		return regularLeaf(key, value)
	}
}
