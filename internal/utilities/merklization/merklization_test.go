package merklization

import (
	"errors"
	"reflect"
	"testing"

	jamTypes "github.com/New-JAMneration/JAM-Protocol/internal/jam_types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
)

func TestBytesToBits(t *testing.T) {
	testCases := []struct {
		input    jamTypes.ByteSequence
		expected jamTypes.BitSequence
	}{
		{jamTypes.ByteSequence{}, jamTypes.BitSequence{}},
		{jamTypes.ByteSequence{0}, jamTypes.BitSequence{false, false, false, false, false, false, false, false}},
		{jamTypes.ByteSequence{1}, jamTypes.BitSequence{false, false, false, false, false, false, false, true}},
		{jamTypes.ByteSequence{10}, jamTypes.BitSequence{false, false, false, false, true, false, true, false}},
		{jamTypes.ByteSequence{100}, jamTypes.BitSequence{false, true, true, false, false, true, false, false}},
		{jamTypes.ByteSequence{128}, jamTypes.BitSequence{true, false, false, false, false, false, false, false}},
		{jamTypes.ByteSequence{255}, jamTypes.BitSequence{true, true, true, true, true, true, true, true}},
		{jamTypes.ByteSequence{0, 0}, jamTypes.BitSequence{false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false}},
		{jamTypes.ByteSequence{160, 0}, jamTypes.BitSequence{true, false, true, false, false, false, false, false, false, false, false, false, false, false, false, false}},
	}

	for _, tc := range testCases {
		actual := bytesToBits(tc.input)

		if len(actual) != len(tc.expected) {
			t.Errorf("Expected %v, got %v", tc.expected, actual)
		}

		if !reflect.DeepEqual(actual, tc.expected) {
			t.Errorf("Expected %v, got %v", tc.expected, actual)
		}
	}
}

func TestBitsToBytes(t *testing.T) {
	testCases := []struct {
		input         jamTypes.BitSequence
		expected      jamTypes.ByteSequence
		expectedError error
	}{
		{jamTypes.BitSequence{}, jamTypes.ByteSequence{}, nil},
		{jamTypes.BitSequence{false, false, false, false, false, false, false, false}, jamTypes.ByteSequence{0}, nil},
		{jamTypes.BitSequence{false, false, false, false, false, false, false, true}, jamTypes.ByteSequence{1}, nil},
		{jamTypes.BitSequence{false, false, false, false, true, false, true, false}, jamTypes.ByteSequence{10}, nil},
		{jamTypes.BitSequence{false, true, true, false, false, true, false, false}, jamTypes.ByteSequence{100}, nil},
		{jamTypes.BitSequence{true, false, false, false, false, false, false, false}, jamTypes.ByteSequence{128}, nil},
		{jamTypes.BitSequence{true, true, true, true, true, true, true, true}, jamTypes.ByteSequence{255}, nil},
		{jamTypes.BitSequence{false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false}, jamTypes.ByteSequence{0, 0}, nil},
		{jamTypes.BitSequence{true, false, true, false, false, false, false, false, false, false, false, false, false, false, false, false}, jamTypes.ByteSequence{160, 0}, nil},
		{jamTypes.BitSequence{false}, nil, errors.New("bit sequence length must be a multiple of 8")},
		{jamTypes.BitSequence{true, false, false, false, false, false, false, false, false}, nil, errors.New("bit sequence length must be a multiple of 8")},
	}

	for _, tc := range testCases {
		actual, error := bitsToBytes(tc.input)

		// Check error message
		if error != nil {
			if error.Error() != tc.expectedError.Error() {
				t.Errorf("Expected %v, got %v", tc.expectedError, error)
			}
		}

		if len(actual) != len(tc.expected) {
			t.Errorf("Expected %v, got %v", tc.expected, actual)
		}

		if !reflect.DeepEqual(actual, tc.expected) {
			t.Errorf("Expected %v, got %v", tc.expected, actual)
		}
	}
}

func TestBranchEncoding(t *testing.T) {
	testCases := []struct {
		left  jamTypes.OpaqueHash
		right jamTypes.OpaqueHash
	}{
		{jamTypes.OpaqueHash{}, jamTypes.OpaqueHash{}},
		{hash.Blake2bHash([]byte("left")), hash.Blake2bHash([]byte("right"))},
	}

	for _, tc := range testCases {
		actual := BranchEncoding(tc.left, tc.right)

		if len(actual) != NODE_SIZE {
			t.Errorf("Expected %v, got %v", NODE_SIZE, actual)
		}

		// Branch encoding first bit should be 0 (The node is branch)
		if actual[0] != false {
			t.Errorf("Expected %v, got %v", false, actual[0])
		}

		leftBits := bytesToBits(tc.left[:])
		rightBits := bytesToBits(tc.right[:])

		// Left bits should be 255 bits
		// Branch encoding [1:256] should be equal to left bits
		for i := 1; i < 256; i++ {
			if actual[i] != leftBits[i] {
				t.Errorf("Expected %v, got %v", leftBits[i], actual[i])
			}
		}

		// Right bits should be 256 bits
		// Branch encoding [256:512] should be equal to right bits
		for i := 256; i < 512; i++ {
			if actual[i] != rightBits[i-256] {
				t.Errorf("Expected %v, got %v", rightBits[i-256], actual[i])
			}
		}
	}
}

func TestEmbeddedValueLeaf(t *testing.T) {
	testCases := []struct {
		key   jamTypes.OpaqueHash
		value jamTypes.ByteSequence
	}{
		{jamTypes.OpaqueHash{}, jamTypes.ByteSequence{}},
		{hash.Blake2bHash([]byte("embedded_value_leaf_test_1")), jamTypes.ByteSequence{1, 2, 3}},
		{hash.Blake2bHash([]byte("embedded_value_leaf_test_2")), jamTypes.ByteSequence{1, 11, 111, 0, 3, 0}},
	}

	for _, tc := range testCases {
		actual := embeddedValueLeaf(tc.key, tc.value)

		if len(actual) != NODE_SIZE {
			t.Errorf("Expected %v, got %v", NODE_SIZE, len(actual))
		}

		// Check the first bit is 1 (The node is a leaf)
		if actual[0] != true {
			t.Errorf("Expected %v, got %v", true, actual[0])
		}

		// Check the second bit is 0 (The leaf is embedded value leaf)
		if actual[1] != false {
			t.Errorf("Expected %v, got %v", false, actual[1])
		}

		// Check the serialized output of the value size
		valueSize := jamTypes.U32(len(tc.value))
		serializedValueSize := utilities.SerializeFixedLength(valueSize, 1)
		valueSizeBits := bytesToBits(serializedValueSize)
		if !reflect.DeepEqual(actual[2:8], valueSizeBits[:6]) {
			t.Errorf("Expected %v, got %v", valueSizeBits[:6], actual[2:8])
		}

		// Check the key bits
		keyBits := bytesToBits(tc.key[:])
		if !reflect.DeepEqual(actual[8:256], keyBits[:248]) {
			t.Errorf("Expected %v, got %v", keyBits[:248], actual[8:256])
		}

		// Check the value bits
		valueBits := bytesToBits(tc.value)
		valueSizeInBits := len(valueBits)
		if !reflect.DeepEqual(actual[256:256+valueSizeInBits], valueBits) {
			t.Errorf("Expected %v, got %v", valueBits, actual[256:256+valueSizeInBits])
		}

		// Check the remaining bits are 0
		for i := 256 + valueSizeInBits; i < NODE_SIZE; i++ {
			if actual[i] != false {
				t.Errorf("Expected %v, got %v", false, actual[i])
			}
		}
	}
}

func TestRegularLeaf(t *testing.T) {
	testCases := []struct {
		key   jamTypes.OpaqueHash
		value jamTypes.ByteSequence
	}{
		{jamTypes.OpaqueHash{}, jamTypes.ByteSequence{}},
		{hash.Blake2bHash([]byte("regular_leaf_test_1")), jamTypes.ByteSequence{
			1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16,
			17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32,
		}},
		{hash.Blake2bHash([]byte("regular_leaf_test_2")), jamTypes.ByteSequence{
			0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
			0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		}},
	}

	for _, tc := range testCases {
		actual := regularLeaf(tc.key, tc.value)

		if len(actual) != NODE_SIZE {
			t.Errorf("Expected %v, got %v", NODE_SIZE, len(actual))
		}

		// Check the first bit is 1 (The node is a leaf)
		if actual[0] != true {
			t.Errorf("Expected %v, got %v", true, actual[0])
		}

		// Check the second bit is 1 (The leaf is regular leaf)
		if actual[1] != true {
			t.Errorf("Expected %v, got %v", true, actual[1])
		}

		// Check the remaining bits are 0 in first byte
		for i := 2; i < 8; i++ {
			if actual[i] != false {
				t.Errorf("Expected %v, got %v", false, actual[i])
			}
		}

		// Check the key bits
		keyBits := bytesToBits(tc.key[:])
		if !reflect.DeepEqual(actual[8:256], keyBits[:248]) {
			t.Errorf("Expected %v, got %v", keyBits[:248], actual[8:256])
		}

		// Check the hash value bits
		valueHash := hash.Blake2bHash(tc.value)
		valueHashBits := bytesToBits(valueHash[:])
		if !reflect.DeepEqual(actual[256:512], valueHashBits) {
			t.Errorf("Expected %v, got %v", valueHashBits, actual[256:512])
		}
	}
}

func TestLeafEncoding(t *testing.T) {
	testCases := []struct {
		key   jamTypes.OpaqueHash
		value jamTypes.ByteSequence
	}{
		{jamTypes.OpaqueHash{}, jamTypes.ByteSequence{}},
		{hash.Blake2bHash([]byte("embedded_value_leaf_test_1")), jamTypes.ByteSequence{1, 2, 3}},
		{hash.Blake2bHash([]byte("embedded_value_leaf_test_2")), jamTypes.ByteSequence{1, 11, 111, 0, 3, 0}},
		{hash.Blake2bHash([]byte("regular_leaf_test_1")), jamTypes.ByteSequence{
			1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16,
			17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32,
		}},
		{hash.Blake2bHash([]byte("regular_leaf_test_2")), jamTypes.ByteSequence{
			0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
			0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		}},
	}

	for _, tc := range testCases {
		actual := LeafEncoding(tc.key, tc.value)

		if len(actual) != NODE_SIZE {
			t.Errorf("Expected %v, got %v", NODE_SIZE, len(actual))
		}

		// Check the first bit is 1 (The node is a leaf)
		if actual[0] != true {
			t.Errorf("Expected %v, got %v", true, actual[0])
		}

		// Calculate the size of value
		valueSize := len(tc.value)
		if valueSize <= 32 {
			// This is an embedded value leaf, its second bit should be 0
			if actual[1] != false {
				t.Errorf("Expected %v, got %v", false, actual[1])
			}
		} else {
			// This is a regular leaf, its second bit should be 1
			if actual[1] != true {
				t.Errorf("Expected %v, got %v", true, actual[1])
			}
		}
	}
}
