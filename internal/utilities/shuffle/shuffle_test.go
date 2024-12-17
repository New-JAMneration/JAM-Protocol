package shuffle

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/jam_types"
	"golang.org/x/crypto/blake2b"
)

// TestSerialize tests the Serialize function across various data types and values.
// TestSerializeFixedLength verifies that SerializeFixedLength correctly encodes integers to fixed-length octets.
func TestSerializeFixedLength(t *testing.T) {
	tests := []struct {
		x       jam_types.U64
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

func TestDeserializeFixedLength(t *testing.T) {
	// multi test cases
	testCases := []struct {
		input  jam_types.ByteSequence
		output jam_types.U64
	}{
		{jam_types.ByteSequence{0x01}, 1},
		{jam_types.ByteSequence{0x80}, 128},
		{jam_types.ByteSequence{0x8A, 0x33, 0x0B, 0xDF}, 3742053258},
		{jam_types.ByteSequence{0x88, 0x77, 0x66, 0x55, 0x44, 0x33, 0x22, 0x11}, 1234605616436508552},
	}

	for _, tc := range testCases {
		if DeserializeFixedLength(tc.input) != tc.output {
			t.Errorf("The output of DeserializeFixedLength is not correct")
		}
	}
}

func TestBlake2bHash(t *testing.T) {
	input := []byte{0x00, 0x01, 0x02, 0x03}
	hash := blake2bHash(input)

	if len(hash) != 32 {
		t.Errorf("The length of the hash is not correct")
	}
}

// TestNumericSequenceFromHash tests the numericSequenceFromHash function.
func TestNumericSequenceFromHash(t *testing.T) {
	// Create a hash as randomness
	hash := blake2bHash([]byte{0x00, 0x01, 0x02, 0x03})

	// The output length of the numeric sequence
	length := jam_types.U32(10)

	numericSequence := numericSequenceFromHash(hash, length)

	if jam_types.U32(len(numericSequence)) != length {
		t.Errorf("The length of the numeric sequence is not correct")
	}

	for i := jam_types.U32(0); i < length; i++ {
		if numericSequence[i] == 0 {
			t.Errorf("The numeric sequence is not generated correctly")
		}
	}
}

func TestShuffle(t *testing.T) {
	// Create a numeric sequence
	s := []jam_types.U32{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	// Create a hash as randomness
	hash := blake2bHash([]byte{0x00, 0x01, 0x02, 0x03})
	shuffled := Shuffle(s, hash)

	if len(shuffled) != len(s) {
		t.Errorf("The length of the shuffled sequence is not correct")
	}

	for i := 0; i < len(s); i++ {
		if shuffled[i] == 0 {
			t.Errorf("The shuffled sequence is not generated correctly")
		}
	}
}

// GenerateRandomHash generates a random 32-byte hash
func GenerateRandomHash() jam_types.OpaqueHash {
	randomBytes := make([]byte, 32)
	_, err := rand.Read(randomBytes)
	if err != nil {
		panic(err)
	}
	return blake2b.Sum256(randomBytes)
}

func factorial(n int) int {
	if n == 0 {
		return 1
	}
	return n * factorial(n-1)
}

func printStatisicTable(original []jam_types.U32, counts map[string]int, iterations int, tolerancePercentage int) {
	averagePercentage := float32(100 / float32(len(counts)))
	expectedCount := iterations / len(counts)

	fmt.Println("Original sequence:", original)
	fmt.Println("Iterations:", iterations)
	fmt.Printf("tolerance percentage: %d%%\n", tolerancePercentage)
	fmt.Printf("Average percentage: %.2f%%\n\n", averagePercentage)

	fmt.Printf("+----------------+--------------+-----------------+\n")
	fmt.Printf("| Permutation    | Count        | Percent         |\n")
	fmt.Printf("+----------------+--------------+-----------------+\n")
	for key, count := range counts {
		percentage := float64(count) / float64(iterations) * 100
		deviationCount := count - expectedCount
		deviationPercentage := percentage - float64(averagePercentage)

		fmt.Printf("| %-14s | %-5d (%+4d) | %5.2f%% (%+5.2f%%) |\n", key, count, deviationCount, percentage, deviationPercentage)
	}

	fmt.Printf("+----------------+--------------+-----------------+\n")
}

func TestShuffleRandomness(t *testing.T) {
	const iterations = 10000
	const tolerancePercentage = 5

	counts := make(map[string]int)
	original := []jam_types.U32{1, 2, 3}

	for i := 0; i < iterations; i++ {
		hash := GenerateRandomHash()
		shuffled := Shuffle(original, hash)
		key := fmt.Sprint(shuffled)
		counts[key]++
	}

	expectedCount := iterations / len(counts)
	tolerance := expectedCount * tolerancePercentage / 100

	// printStatisicTable(original, counts, iterations, tolerancePercentage)

	// Check the number of unique permutations
	if len(counts) != factorial(len(original)) {
		t.Errorf("The number of unique permutations is not correct")
	}

	// Check the percentage of each permutation
	for _, count := range counts {
		if count < expectedCount-tolerance || count > expectedCount+tolerance {
			t.Errorf("The percentage of each permutation is not in tolerance")
		}
	}
}
