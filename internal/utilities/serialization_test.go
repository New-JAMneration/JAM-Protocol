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
		l       int
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

func TestTrivialEncodings(t *testing.T) {
	tests := []struct {
		input any
		want  []byte
	}{
		// E(∅) = []
		{nil, []byte{}},

		// E(x∈Y) = x if x is a byte-sequence
		{[]byte{0xDE, 0xAD}, []byte{0xDE, 0xAD}},
		{"Hello", []byte("Hello")},

		// Tuples/Sequences: E({a,b,...}) = E(a)||E(b)||...
		{
			[]any{[]byte("A"), []byte("B")},
			[]byte("AB"),
		},
		{
			[]any{"Hi", []byte{0x01, 0x02}, "World"},
			append(append([]byte("Hi"), 0x01, 0x02), []byte("World")...),
		},
		// Sequence example (C.1.3):
		{
			[]any{"SeqItem1", []byte{0xAB}, "SeqItem2"},
			append(append([]byte("SeqItem1"), 0xAB), []byte("SeqItem2")...),
		},
	}

	for _, tt := range tests {
		got := Serialize(tt.input)
		if !bytes.Equal(got, tt.want) {
			t.Errorf("Serialize(%v) = %X, want %X", tt.input, got, tt.want)
		}
	}
}
