package utilities

import (
	"bytes"
	"testing"

	jamtypes "github.com/New-JAMneration/JAM-Protocol/internal/jam_types"
)

// TestSerialize tests the Serialize function across various data types and values.
// TestSerializeFixedLength verifies that SerializeFixedLength correctly encodes integers to fixed-length octets.
func TestSerializeFixedLength(t *testing.T) {
	tests := []struct {
		x       jamtypes.U64
		l       jamtypes.U64
		wantHex []byte
	}{
		{0, 0, []byte{}},
		{1, 1, []byte{0x01}},
		{128, 1, []byte{0x80}},
		{1, 8, []byte{0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
		{256, 2, []byte{0x00, 0x01}},
		{10000, 2, []byte{0x10, 0x27}},
		{65535, 2, []byte{0xFF, 0xFF}},
		{65535, 3, []byte{0xFF, 0xFF, 0x00}},
		{0x1122334455667788, 8, []byte{0x88, 0x77, 0x66, 0x55, 0x44, 0x33, 0x22, 0x11}},
	}

	for _, tt := range tests {
		got := SerializeFixedLength(tt.x, tt.l)
		if !bytes.Equal(got, tt.wantHex) {
			t.Errorf("SerializeFixedLength(%d, %d) = %X, want %X", tt.x, tt.l, got, tt.wantHex)
		}
	}
}

func TestSerializeFixedLengthU32(t *testing.T) {
	tests := []struct {
		x       jamtypes.U32
		l       jamtypes.U32
		wantHex []byte
	}{
		{0, 0, []byte{}},
		{1, 1, []byte{0x01}},
		{128, 1, []byte{0x80}},
		{1, 8, []byte{0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
		{256, 2, []byte{0x00, 0x01}},
		{10000, 2, []byte{0x10, 0x27}},
		{65535, 2, []byte{0xFF, 0xFF}},
		{65535, 3, []byte{0xFF, 0xFF, 0x00}},
	}

	for _, tt := range tests {
		got := SerializeFixedLength(tt.x, tt.l)
		if !bytes.Equal(got, tt.wantHex) {
			t.Errorf("SerializeFixedLength(%d, %d) = %X, want %X", tt.x, tt.l, got, tt.wantHex)
		}
	}
}

// TestSerializeGeneral checks that SerializeGeneral meets the defined conditions.
func TestSerializeU64(t *testing.T) {
	tests := []struct {
		x       jamtypes.U64
		wantHex []byte
	}{
		// According to the specification:
		// x=0 => E(x)=[0]
		{0, []byte{0x00}},

		// For small values, often the form is a single byte if it fits in the correct range
		{1, []byte{0xFF, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
		{127, []byte{0xFF, 0x7F, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}}, // Should still fit in one byte if it satisfies condition 2^(7*1)=128

		{128, []byte{0x80, 0x80}},
		{255, []byte{0x80, 0xFF}},

		// Larger values near 2^(7*l)
		// l = 1
		// prefix = 128 + 63 = 191 = 0xBF
		// reminder = 255 mode 256 = 0xFF
		{16383, []byte{0xBF, 0xFF}}, // Example: Just before requiring a longer prefix
	}

	for _, tt := range tests {
		got := SerializeU64(tt.x)
		if !bytes.Equal(got, tt.wantHex) {
			t.Errorf("Serialize(%d) = %X, want %X", tt.x, got, tt.wantHex)
		}
	}
}

func TestBitSequenceWrapper(t *testing.T) {
	bits := jamtypes.BitSequence{true, false, true, true, false, false, false, true} // 8 bits
	w := BitSequenceWrapper{Bits: bits, IsVariableLength: false}
	got := w.Serialize()

	// Bit pattern: LSB first, bits: 1,0,1,1,0,0,0,1 = 0x9B (binary 110110011)
	// Let's compute manually:
	// bit0 = true => 1
	// bit1 = false => 0
	// bit2 = true => 1<<2=4
	// bit3 = true => 1<<3=8
	// bit4 = false =>0
	// bit5 = false =>0
	// bit6 = false =>0
	// bit7 = true =>1<<7=128
	// sum = 1+0+4+8+0+0+0+128 = 141 decimal = 0x8D in hex, check carefully:
	// Actually, let's write the bits in order: bit0=1,bit1=0,bit2=1,bit3=1,bit4=0,bit5=0,bit6=0,bit7=1
	// binary: 10001101 in binary = 0x8D indeed.

	want := []byte{0x8D}
	if !bytes.Equal(got, want) {
		t.Errorf("BitSequenceWrapper.Serialize() = %X, want %X", got, want)
	}
}

func TestMapWarpper_Empty(t *testing.T) {
	m := MapWarpper{Value: make(map[Comparable]Serializable)}
	got := m.Serialize()
	if len(got) != 0 {
		t.Errorf("MapWarpper(empty).Serialize() = %X, want empty", got)
	}
}

func TestMapWarpper_Simple(t *testing.T) {
	// Create a dictionary:
	// Key: StringOctets, Value: U64Wrapper
	// Keys: "apple", "banana"
	m := MapWarpper{
		Value: map[Comparable]Serializable{
			StringOctets("banana"): U64Wrapper{Value: 200},
			StringOctets("apple"):  U64Wrapper{Value: 100},
		},
	}

	got := m.Serialize()
	if len(got) == 0 {
		t.Fatal("MapWarpper.Serialize() returned empty, expected some output")
	}

	if !bytes.Contains(got, U64Wrapper{2}.Serialize()) {
		t.Errorf("Serialized output does not contain length")
	}
	// We know keys will be sorted as "apple" < "banana".
	// Discriminator: length of seq = 2 pairs
	// Each pair = { E(k), E(v) } = { "apple", SerializeU64(100) } followed by { "banana", SerializeU64(200) }.
	// Let's just check that "apple" appears before "banana" in the serialized output and that 100 appears before 200.
	if !bytes.Contains(got, []byte("apple")) {
		t.Errorf("Serialized output does not contain 'apple'")
	}
	if !bytes.Contains(got, []byte("banana")) {
		t.Errorf("Serialized output does not contain 'banana'")
	}

	// Check order: 'apple' should come before 'banana'
	idxApple := bytes.Index(got, []byte("apple"))
	idxBanana := bytes.Index(got, []byte("banana"))
	if idxApple == -1 || idxBanana == -1 || idxApple > idxBanana {
		t.Errorf("'apple' should appear before 'banana' in the output")
	}
}

func TestSetWarpper_Empty(t *testing.T) {
	s := SetWarpper{Value: []Comparable{}}
	got := s.Serialize()
	if len(got) != 0 {
		t.Errorf("SetWarpper(empty).Serialize() = %X, want empty", got)
	}
}

func TestSetWarpper_Simple(t *testing.T) {
	// They should be sorted alphabetically: "alice", "bob", "charlie"
	s := SetWarpper{
		Value: []Comparable{
			StringOctets("charlie"),
			StringOctets("alice"),
			StringOctets("bob"),
		},
	}

	got := s.Serialize()
	if len(got) == 0 {
		t.Fatal("SetWarpper.Serialize() returned empty, expected some output")
	}

	// Check that all elements appear
	if !bytes.Contains(got, []byte("alice")) ||
		!bytes.Contains(got, []byte("bob")) ||
		!bytes.Contains(got, []byte("charlie")) {
		t.Errorf("SetWarpper.Serialize() output missing some elements")
	}

	// Check order: 'alice' < 'bob' < 'charlie'
	idxAlice := bytes.Index(got, []byte("alice"))
	idxBob := bytes.Index(got, []byte("bob"))
	idxCharlie := bytes.Index(got, []byte("charlie"))

	if idxAlice == -1 || idxBob == -1 || idxCharlie == -1 {
		t.Errorf("Some elements not found in serialized set")
	}
	if !(idxAlice < idxBob && idxBob < idxCharlie) {
		t.Errorf("Elements are not in alphabetical order: got=%s", got)
	}
}

func TestDiscriminatorSerialization(t *testing.T) {
	// Define test cases
	tests := []struct {
		name     string
		input    Discriminator
		expected jamtypes.ByteSequence
	}{
		{
			name:     "Empty Discriminator",
			input:    Discriminator{Value: []Serializable{}},
			expected: append(WrapU64(0).Serialize(), []byte{}...), // Length 0, no values
		},
		{
			name: "Non-Empty Discriminator",
			input: Discriminator{
				Value: []Serializable{
					U8Wrapper{Value: 1},
					U16Wrapper{Value: 300},
					U32Wrapper{Value: 100000},
				},
			},
			expected: func() jamtypes.ByteSequence {
				// Manually construct the expected output
				lengthPrefix := WrapU64(3).Serialize() // Length prefix: 3 elements
				seq := append(U8Wrapper{Value: 1}.Serialize(),
					U16Wrapper{Value: 300}.Serialize()...)
				seq = append(seq, U32Wrapper{Value: 100000}.Serialize()...)

				return append(lengthPrefix, seq...)
			}(),
		},
	}

	// Run each test case
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := tt.input.Serialize()

			if !bytes.Equal(output, tt.expected) {
				t.Errorf("Test %s failed. Expected: %v, Got: %v", tt.name, tt.expected, output)
			}
		})
	}
}
