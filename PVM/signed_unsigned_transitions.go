package PVM

import (
	"fmt"
)

// A.12
// UnsignedToBits converts an unsigned integer to a slice of bits.
func UnsignedToBits(x uint64, n uint) ([]bool, error) {
	// n should be in the range of 1 ~ 8
	if n < 1 || n > 8 {
		return nil, fmt.Errorf("n should be in the range of 1 ~ 8: got %d", n)
	}

	bitSize := int(8 * n)

	// input should be in the range of [0, 2^(8n)-1]
	maxValue := uint64(1)<<uint(bitSize) - 1
	if x > maxValue {
		return nil, fmt.Errorf("UnsignedToBits: x >= 2^(8n): got %d", x)
	}

	y := make([]bool, bitSize) // make a slice of bool with length bitSize

	for i := 0; i < bitSize; i++ {
		afterShift := x >> uint64(bitSize-1-i)
		y[i] = (afterShift & 1) == 1
	}

	return y, nil
}

// A.13
// BitsToUnsigned converts a slice of bits to an unsigned integer.
func BitsToUnsigned(x []bool, n uint) (uint64, error) {
	// n should be in the range of 1 ~ 8
	if n < 1 || n > 8 {
		return 0, fmt.Errorf("n should be in the range of 1 ~ 8: got %d", n)
	}

	var y uint64

	bitsLength := len(x)

	if bitsLength != int(8*n) {
		return 0, fmt.Errorf("BitsToUnsigned: len(x) != 8n: got %d", bitsLength)
	}

	for i, bit := range x {
		if bit {
			binaryValue := uint64(1) << uint64(bitsLength-1-i) // binary value of the bit
			y |= binaryValue                                   // add the value to y
		}
	}

	return y, nil
}

// A.14
// ReverseUnsignedToBits converts an unsigned integer to a slice of bits in
// reverse order.
func ReverseUnsignedToBits(x uint64, n uint) ([]bool, error) {
	// n should be in the range of 1 ~ 8
	if n < 1 || n > 8 {
		return nil, fmt.Errorf("n should be in the range of 1 ~ 8: got %d", n)
	}

	bitSize := int(8 * n)

	// input should be in the range of [0, 2^(8n)-1]
	maxValue := uint64(1) << uint(bitSize)
	if x >= maxValue {
		return nil, fmt.Errorf("ReverseUnsignedToBits: x >= 2^(8n): got %d", x)
	}

	y := make([]bool, bitSize) // make a slice of bool with length bitSize

	for i := 0; i < bitSize; i++ {
		y[i] = (x >> i & 1) == 1
	}

	return y, nil
}

// A.15
// ReverseBitsToUnsigned converts a slice of bits in reverse order to an
// unsigned integer.
func ReverseBitsToUnsigned(x []bool, n uint) (uint64, error) {
	// n should be in the range of 1 ~ 8
	if n < 1 || n > 8 {
		return 0, fmt.Errorf("n should be in the range of 1 ~ 8: got %d", n)
	}

	var y uint64

	if len(x) != int(8*n) {
		return 0, fmt.Errorf("ReverseBitsToUnsigned: len(x) != 8n: got %d", len(x))
	}

	for i, bit := range x {
		if bit {
			reversedBinaryValue := uint64(1) << uint64(i) // reversed binary value of the bit
			y |= reversedBinaryValue                      // add the value to y
		}
	}

	return y, nil
}

// A.16 SignExtend
func SignExtend(n uint8, x uint64) (uint64, error) {
	switch n {
	case 0:
		return x, nil
	case 1:
		return uint64(int64(int8(x))), nil
	case 2:
		return uint64(int64(int16(x))), nil
	case 3:
		// shift := uint(64 - n*8)
		x &= 0xFFFFFF // filter high bits
		return uint64(int64(x<<40) >> 40), nil
	case 4:
		return uint64(int64(int32(x))), nil
	case 8:
		return x, nil
	default:
		return 0, fmt.Errorf("invalid byte count: got %d", n)
	}
}
