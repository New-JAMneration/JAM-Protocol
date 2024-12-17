package hash

import (
	"bytes"
	"encoding/hex"
	"golang.org/x/crypto/sha3"
	"testing"
)

func TestBlake2bHash(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
		want  int // hash length should be 32
	}{
		{
			name:  "empty input",
			input: []byte{},
			want:  32,
		},
		{
			name:  "normal string",
			input: []byte("hello world"),
			want:  32,
		},
		{
			name:  "binary data",
			input: []byte{0x00, 0xFF, 0xAA, 0x55},
			want:  32,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Blake2bHash(tt.input)
			if len(got) != tt.want {
				t.Errorf("Blake2bHash() length = %v, want %v", len(got), tt.want)
			}
			// Test consistency
			got2 := Blake2bHash(tt.input)
			gotBytes := make([]byte, len(got))
			got2Bytes := make([]byte, len(got2))
			copy(gotBytes, got[:])
			copy(got2Bytes, got2[:])
			t.Logf("Input: %s", tt.input)
			t.Logf("Hash result (length=%d): %x", len(gotBytes), gotBytes)
			t.Logf("Hash result 2 (length=%d): %x", len(got2Bytes), got2Bytes)
			if !bytes.Equal(gotBytes, got2Bytes) {
				t.Errorf("Blake2bHash() not consistent for same input\ngot: %x\ngot2: %x", gotBytes, got2Bytes)
			}
		})
	}
}

func TestBlake2bHashPartial(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
		x     int
		want  int
	}{
		{
			name:  "empty input, 16 bytes",
			input: []byte{},
			x:     16,
			want:  16,
		},
		{
			name:  "normal string, 24 bytes",
			input: []byte("hello world"),
			x:     24,
			want:  24,
		},
		{
			name:  "binary data, 8 bytes",
			input: []byte{0x00, 0xFF, 0xAA, 0x55},
			x:     8,
			want:  8,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Blake2bHashPartial(tt.input, tt.x)
			if len(got) != tt.want {
				t.Errorf("Blake2bHashPartial() length = %v, want %v", len(got), tt.want)
			}
			// Test consistency
			got2 := Blake2bHashPartial(tt.input, tt.x)
			gotBytes := make([]byte, len(got))
			got2Bytes := make([]byte, len(got2))
			copy(gotBytes, got[:])
			copy(got2Bytes, got2[:])
			t.Logf("Input: %s", tt.input)
			t.Logf("Hash result (length=%d): %x", len(gotBytes), gotBytes)
			t.Logf("Hash result 2 (length=%d): %x", len(got2Bytes), got2Bytes)
			if !bytes.Equal(gotBytes, got2Bytes) {
				t.Errorf("Blake2bHashPartial() not consistent for same input\ngot: %x\ngot2: %x", gotBytes, got2Bytes)
			}
		})
	}
}

func TestKeccakHash(t *testing.T) {
	hash := sha3.NewLegacyKeccak256()

	var buf []byte
	//hash.Write([]byte{0xcc})
	hash.Write(decodeHex("cc"))
	buf = hash.Sum(nil)

	t.Logf("ouput: %v", hex.EncodeToString(buf))
	t.Logf("expected: EEAD6DBFC7340A56CAEDC044696A168870549A6A7F6F56961E84A54BD9970B8A")
}

func decodeHex(s string) []byte {
	b, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}

	return b
}
