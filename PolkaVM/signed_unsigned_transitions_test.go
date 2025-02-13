package PolkaVM

import (
	"fmt"
	"testing"
)

func TestUnsignedToSigned(t *testing.T) {
	testCases := []struct {
		a      uint64
		n      uint
		result int64
		err    error
	}{
		{0, 0, 0, fmt.Errorf("n should be in the range of 1 ~ 8")},
		{1, 0, 0, fmt.Errorf("n should be in the range of 1 ~ 8")},
		{1, 10, 0, fmt.Errorf("n should be in the range of 1 ~ 8")},
		{0, 1, 0, nil},
		{1, 1, 1, nil},
		{32, 1, 32, nil},
		{100, 1, 100, nil},
		{127, 1, 127, nil},
		{128, 1, -128, nil},
		{200, 1, -56, nil},
		{255, 1, -1, nil},
		{256, 1, 0, fmt.Errorf("UnsignedToSigned: a >= 2^(8n)")},
		{32767, 2, 32767, nil},
		{32768, 2, -32768, nil},
		{65535, 2, -1, nil},
		{65536, 2, 0, fmt.Errorf("UnsignedToSigned: a >= 2^(8n)")},
	}

	for _, tc := range testCases {
		result, err := UnsignedToSigned(tc.a, tc.n)
		if result != tc.result {
			t.Errorf("Expected %d, got %d", tc.result, result)
		}

		if err != nil && tc.err == nil {
			t.Errorf("Expected nil error, got %v", err)
		}
	}
}

func TestSignedToUnsigned(t *testing.T) {
	testCases := []struct {
		a      int64
		n      uint
		result uint64
		err    error
	}{
		{0, 0, 0, fmt.Errorf("n should be in the range of 1 ~ 8")},
		{0, 1, 0, nil},
		{1, 1, 1, nil},
		{1, 10, 0, fmt.Errorf("n should be in the range of 1 ~ 8")},
		{32, 1, 32, nil},
		{100, 1, 100, nil},
		{127, 1, 127, nil},
		{128, 1, 0, fmt.Errorf("SignedToUnsigned: input: 128 out of range (-128 ~ 127)")},
		{255, 2, 255, nil},
		{-128, 1, 128, nil},
		{-56, 1, 200, nil},
		{-1, 1, 255, nil},
		{32767, 2, 32767, nil},
		{-32768, 2, 32768, nil},
		{50000, 2, 0, fmt.Errorf("SignedToUnsigned: input: 50000 out of range (-32768 ~ 32767)")},
	}

	for _, tc := range testCases {
		result, err := SignedToUnsigned(tc.a, tc.n)
		if result != tc.result {
			t.Errorf("Expected %d, got %d", tc.result, result)
		}

		if err != nil && tc.err == nil {
			t.Errorf("Expected nil error, got %v", err)
		}
	}
}

func TestUnsignedToBits(t *testing.T) {
	testCases := []struct {
		x      uint64
		n      uint
		result []bool
		err    error
	}{
		{0, 0, nil, fmt.Errorf("n should be in the range of 1 ~ 8")},
		{1, 10, nil, fmt.Errorf("n should be in the range of 1 ~ 8")},
		{0, 1, []bool{false, false, false, false, false, false, false, false}, nil},
		{1, 1, []bool{false, false, false, false, false, false, false, true}, nil},
		{32, 1, []bool{false, false, true, false, false, false, false, false}, nil},
		{255, 1, []bool{true, true, true, true, true, true, true, true}, nil},
		{256, 1, nil, fmt.Errorf("UnsignedToBits: x >= 2^(8n)")},
		{255, 2, []bool{false, false, false, false, false, false, false, false, true, true, true, true, true, true, true, true}, nil},
	}

	for _, tc := range testCases {
		result, err := UnsignedToBits(tc.x, tc.n)

		if err != nil && tc.err == nil {
			t.Errorf("Expected nil error, got %v", err)
		}

		if err == nil {
			for i := range tc.result {
				if result[i] != tc.result[i] {
					t.Errorf("Expected %v, got %v", tc.result, result)
				}
			}
		}
	}
}

func TestBitsToUnsigned(t *testing.T) {
	testCases := []struct {
		x      []bool
		n      uint
		result uint64
		err    error
	}{
		{[]bool{false, false, false, false, false, false, false, false}, 0, 0, fmt.Errorf("n should be in the range of 1 ~ 8")},
		{[]bool{false, false, false, false, false, false, false, false}, 10, 0, fmt.Errorf("n should be in the range of 1 ~ 8")},
		{[]bool{false, false, false, false, false, false, false, false}, 1, 0, nil},
		{[]bool{false, false, false, false, false, false, false, true}, 1, 1, nil},
		{[]bool{false, false, true, false, false, false, false, false}, 1, 32, nil},
		{[]bool{true, true, true, true, true, true, true, true}, 1, 255, nil},
		{[]bool{false, false, false, false, false, false, false, false, true, true, true, true, true, true, true, true}, 2, 255, nil},
		{[]bool{false, false, false, false, false, false, false, false, true, true, true, true, true, true, true, true}, 1, 0, fmt.Errorf("BitsToUnsigned: len(x) != 8n")},
	}

	for _, tc := range testCases {
		result, err := BitsToUnsigned(tc.x, tc.n)

		if err != nil && tc.err == nil {
			t.Errorf("Expected nil error, got %v", err)
		}

		if err == nil && result != tc.result {
			t.Errorf("Expected %d, got %d", tc.result, result)
		}
	}
}

func TestReverseUnsignedToBits(t *testing.T) {
	testCases := []struct {
		x      uint64
		n      uint
		result []bool
		err    error
	}{
		{0, 0, nil, fmt.Errorf("n should be in the range of 1 ~ 8")},
		{1, 10, nil, fmt.Errorf("n should be in the range of 1 ~ 8")},
		{0, 1, []bool{false, false, false, false, false, false, false, false}, nil},
		{1, 1, []bool{true, false, false, false, false, false, false, false}, nil},
		{32, 1, []bool{false, false, false, false, false, true, false, false}, nil},
		{127, 1, []bool{true, true, true, true, true, true, true, false}, nil},
		{128, 1, []bool{false, false, false, false, false, false, false, true}, nil},
		{255, 1, []bool{true, true, true, true, true, true, true, true}, nil},
		{256, 1, nil, fmt.Errorf("UnsignedToBits: x >= 2^(8n)")},
		{255, 2, []bool{true, true, true, true, true, true, true, true, false, false, false, false, false, false, false, false}, nil},
		{1, 2, []bool{true, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false}, nil},
	}

	for _, tc := range testCases {
		result, err := ReverseUnsignedToBits(tc.x, tc.n)

		if err != nil && tc.err == nil {
			t.Errorf("Expected nil error, got %v", err)
		}

		if err == nil {
			for i := range tc.result {
				if result[i] != tc.result[i] {
					t.Errorf("Expected %v, got %v", tc.result, result)
				}
			}
		}
	}
}

func TestReverseBitsToUnsigned(t *testing.T) {
	testCases := []struct {
		x      []bool
		n      uint
		result uint64
		err    error
	}{
		{[]bool{false, false, false, false, false, false, false, false}, 0, 0, fmt.Errorf("n should be in the range of 1 ~ 8")},
		{[]bool{false, false, false, false, false, false, false, false}, 10, 0, fmt.Errorf("n should be in the range of 1 ~ 8")},
		{[]bool{false, false, false, false, false, false, false, false}, 1, 0, nil},
		{[]bool{true, false, false, false, false, false, false, false}, 1, 1, nil},
		{[]bool{false, false, false, false, false, true, false, false}, 1, 32, nil},
		{[]bool{true, true, true, true, true, true, true, false}, 1, 127, nil},
		{[]bool{false, false, false, false, false, false, false, true}, 1, 128, nil},
		{[]bool{true, true, true, true, true, true, true, true}, 1, 255, nil},
		{[]bool{true, true, true, true, true, true, true, true, false, false, false, false, false, false, false, false}, 2, 255, nil},
		{[]bool{true, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false}, 2, 1, nil},
		{[]bool{true, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false}, 1, 0, fmt.Errorf("BitsToUnsigned: len(x) != 8n")},
	}

	for _, tc := range testCases {
		result, err := ReverseBitsToUnsigned(tc.x, tc.n)

		if err != nil && tc.err == nil {
			t.Errorf("Expected nil error, got %v", err)
		}

		if err == nil && result != tc.result {
			t.Errorf("Expected %d, got %d", tc.result, result)
		}
	}
}

func TestSignExtend(t *testing.T) {
	testCases := []struct {
		name           string
		n              int
		x              uint64
		expectedOutput uint64
		expectedError  error
	}{
		{"ValidInput1", 1, 151, 0xffffffffffffff97, nil},
		{"ValidInput2", 2, 11121, 0x2b71, nil},
		{"ValidInput3", 8, 123456, 123456, nil},
		{"ValidInput4", 3, 123456, 0x1e240, nil},
		{"ValidInput5", 4, 123456789, 0x75bcd15, nil},
		{"InvalidInput1", 9, 0, 0, fmt.Errorf("invalid byte count")},
		{"InvalidInput2", -1, 0, 0, fmt.Errorf("invalid byte count")},
		{"InvalidInput3", 1, 1000, 0, fmt.Errorf("x (1000) exceeds the maximum value for 8 bytes")},
		{"InvalidInput4", 7, 11, 0, fmt.Errorf("invalid byte count")},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			output, err := SignExtend(tc.n, tc.x)
			if output != tc.expectedOutput {
				t.Errorf("Expected output %x, but got %x", tc.expectedOutput, output)
			}
			if (err != nil && tc.expectedError == nil) || (err == nil && tc.expectedError != nil) || (err != nil && tc.expectedError != nil && err.Error() != tc.expectedError.Error()) {
				t.Errorf("Expected error %v, but got %v", tc.expectedError, err)
			}
		})
	}
}
