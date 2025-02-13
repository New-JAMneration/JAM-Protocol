package PolkaVM

import (
	"fmt"
)

// graypaper 0.6.1

// A.10
// UnsignedToSigned converts an unsigned integer to a signed integer.
func UnsignedToSigned(a uint64, n uint) (int64, error) {
	// n should be in the range of 1 ~ 8
	if n < 1 || n > 8 {
		return 0, fmt.Errorf("n should be in the range of 1 ~ 8")
	}

	bitSize := 8 * n                      // 8n
	signBit := uint64(1) << (bitSize - 1) // 2^(8n-1)
	maxValue := uint64(1) << bitSize      // 2^(8n)

	// input should be in the range of [0, 2^(8n)-1]
	if a >= maxValue {
		return 0, fmt.Errorf("UnsignedToSigned: a >= 2^(8n)")
	}

	if a < signBit {
		return int64(a), nil // a < 2^(8n-1)
	}
	return int64(a - maxValue), nil // a >= 2^(8n-1)
}

// A.11
// SignedToUnsigned converts a signed integer to an unsigned integer.
func SignedToUnsigned(a int64, n uint) (uint64, error) {
	// n should be in the range of 1 ~ 8
	if n < 1 || n > 8 {
		return 0, fmt.Errorf("n should be in the range of 1 ~ 8")
	}

	bitSize := 8 * n
	maxValue := uint64(1) << bitSize // 2^(8n)

	// input should be in the range of [-(2^(8n-1)), 2^(8n-1)-1]
	lowerBound := -(int64(maxValue) >> 1)
	upperBound := int64(maxValue) >> 1

	if (a < lowerBound) || (a >= upperBound) {
		return 0, fmt.Errorf("SignedToUnsigned: input: %d out of range (%d ~ %d)", a, lowerBound, upperBound)
	}

	return (maxValue + uint64(a)) % maxValue, nil
}

// A.12
// UnsignedToBits converts an unsigned integer to a slice of bits.
func UnsignedToBits(x uint64, n uint) ([]bool, error) {
	// n should be in the range of 1 ~ 8
	if n < 1 || n > 8 {
		return nil, fmt.Errorf("n should be in the range of 1 ~ 8")
	}

	bitSize := int(8 * n)

	// input should be in the range of [0, 2^(8n)-1]
	maxValue := uint64(1) << uint(bitSize)
	if x >= maxValue {
		return nil, fmt.Errorf("UnsignedToBits: x >= 2^(8n)")
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
		return 0, fmt.Errorf("n should be in the range of 1 ~ 8")
	}

	var y uint64

	bitsLength := len(x)

	if bitsLength != int(8*n) {
		return 0, fmt.Errorf("BitsToUnsigned: len(x) != 8n")
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
		return nil, fmt.Errorf("n should be in the range of 1 ~ 8")
	}

	bitSize := int(8 * n)

	// input should be in the range of [0, 2^(8n)-1]
	maxValue := uint64(1) << uint(bitSize)
	if x >= maxValue {
		return nil, fmt.Errorf("ReverseUnsignedToBits: x >= 2^(8n)")
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
		return 0, fmt.Errorf("n should be in the range of 1 ~ 8")
	}

	var y uint64

	if len(x) != int(8*n) {
		return 0, fmt.Errorf("ReverseBitsToUnsigned: len(x) != 8n")
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
func SignExtend(n int, x uint64) (uint64, error) {
	if n < 0 || n > 8 || n == 5 || n == 6 || n == 7 {
		return 0, fmt.Errorf("invalid byte count")
	}
	if n == 8 || n == 0 {
		return x, nil
	}
	if x >= (1 << (8 * n)) {
		return 0, fmt.Errorf("x (%d) exceeds the maximum value for %d bytes", x, 8*n)
	}
	var mul, add uint64
	add = x >> (8*n - 1)
	mul = 0
	for i := 8 * n; i < 64; i++ {
		mul |= (1 << i)
	}
	add *= mul
	x += add
	return x, nil
}
