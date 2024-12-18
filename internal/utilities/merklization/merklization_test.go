package merklization

import (
	"testing"

	jamTypes "github.com/New-JAMneration/JAM-Protocol/internal/jam_types"
)

func TestBits(t *testing.T) {
	testCases := []struct {
		input    jamTypes.ByteSequence
		expected jamTypes.BitSequence
	}{
		{jamTypes.ByteSequence{}, jamTypes.BitSequence{}},
		{jamTypes.ByteSequence{0}, jamTypes.BitSequence{false, false, false, false, false, false, false, false}},
		{jamTypes.ByteSequence{1}, jamTypes.BitSequence{false, false, false, false, false, false, false, true}},
		{jamTypes.ByteSequence{128}, jamTypes.BitSequence{true, false, false, false, false, false, false, false}},
		{jamTypes.ByteSequence{255}, jamTypes.BitSequence{true, true, true, true, true, true, true, true}},
		{jamTypes.ByteSequence{0, 0}, jamTypes.BitSequence{false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false}},
		{jamTypes.ByteSequence{160, 0}, jamTypes.BitSequence{true, false, true, false, false, false, false, false, false, false, false, false, false, false, false, false}},
	}

	for _, tc := range testCases {
		actual := bits(tc.input)

		if len(actual) != len(tc.expected) {
			t.Errorf("Expected %v, got %v", tc.expected, actual)
		}

		// Compare the bit sequences.
		for i := 0; i < len(actual); i++ {
			if actual[i] != tc.expected[i] {
				t.Errorf("Expected %v, got %v", tc.expected, actual)
			}
		}
	}
}

func TestBranchEncoding(t *testing.T) {
	testCases := []struct {
		left     jamTypes.OpaqueHash
		right    jamTypes.OpaqueHash
		expected jamTypes.BitSequence
	}{
		{jamTypes.OpaqueHash{}, jamTypes.OpaqueHash{}, jamTypes.BitSequence{}},
	}

	for _, tc := range testCases {
		actual := BranchEncoding(tc.left, tc.right)

		if len(actual) != NODE_SIZE {
			t.Errorf("Expected %v, got %v", tc.expected, actual)
		}
	}
}

func TestEmbeddedValueLeaf(t *testing.T) {
	testCases := []struct {
		key      jamTypes.OpaqueHash
		value    jamTypes.ByteSequence
		expected jamTypes.BitSequence
	}{
		{jamTypes.OpaqueHash{}, jamTypes.ByteSequence{}, jamTypes.BitSequence{}},
	}

	for _, tc := range testCases {
		actual := embeddedValueLeaf(tc.key, tc.value)

		if len(actual) != NODE_SIZE {
			t.Errorf("Expected %v, got %v", NODE_SIZE, len(actual))
		}
	}
}

func TestRegularLeaf(t *testing.T) {
	testCases := []struct {
		key      jamTypes.OpaqueHash
		value    jamTypes.ByteSequence
		expected jamTypes.BitSequence
	}{
		{jamTypes.OpaqueHash{}, jamTypes.ByteSequence{}, jamTypes.BitSequence{}},
	}

	for _, tc := range testCases {
		actual := regularLeaf(tc.key, tc.value)

		if len(actual) != NODE_SIZE {
			t.Errorf("Expected %v, got %v", NODE_SIZE, len(actual))
		}
	}
}

func TestLeafEncoding(t *testing.T) {
	testCases := []struct {
		key      jamTypes.OpaqueHash
		value    jamTypes.ByteSequence
		expected jamTypes.BitSequence
	}{
		{jamTypes.OpaqueHash{}, jamTypes.ByteSequence{}, jamTypes.BitSequence{}},
	}

	for _, tc := range testCases {
		actual := LeafEncoding(tc.key, tc.value)

		if len(actual) != NODE_SIZE {
			t.Errorf("Expected %v, got %v", NODE_SIZE, len(actual))
		}
	}
}
