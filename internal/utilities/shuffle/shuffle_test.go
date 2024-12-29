package shuffle

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"reflect"
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	hashUtil "github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
)

// TestSerialize tests the Serialize function across various data types and values.
// TestSerializeFixedLength verifies that SerializeFixedLength correctly encodes integers to fixed-length octets.
func TestSerializeFixedLength(t *testing.T) {
	tests := []struct {
		x       types.U64
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
		input  types.ByteSequence
		output types.U64
	}{
		{types.ByteSequence{0x01}, 1},
		{types.ByteSequence{0x80}, 128},
		{types.ByteSequence{0x8A, 0x33, 0x0B, 0xDF}, 3742053258},
		{types.ByteSequence{0x88, 0x77, 0x66, 0x55, 0x44, 0x33, 0x22, 0x11}, 1234605616436508552},
	}

	for _, tc := range testCases {
		if DeserializeFixedLength(tc.input) != tc.output {
			t.Errorf("The output of DeserializeFixedLength is not correct")
		}
	}
}

func TestBlake2bHash(t *testing.T) {
	input := []byte{0x00, 0x01, 0x02, 0x03}

	if len(hashUtil.Blake2bHash(input)) != 32 {
		t.Errorf("The length of the hash is not correct")
	}
}

// TestNumericSequenceFromHash tests the numericSequenceFromHash function.
func TestNumericSequenceFromHash(t *testing.T) {
	// Create a hash as randomness
	hash := hashUtil.Blake2bHash([]byte{0x00, 0x01, 0x02, 0x03})

	// The output length of the numeric sequence
	length := types.U32(10)

	numericSequence := numericSequenceFromHash(hash, length)

	if types.U32(len(numericSequence)) != length {
		t.Errorf("The length of the numeric sequence is not correct")
	}

	for i := types.U32(0); i < length; i++ {
		if numericSequence[i] == 0 {
			t.Errorf("The numeric sequence is not generated correctly")
		}
	}
}

func TestShuffle(t *testing.T) {
	// Create a numeric sequence
	s := []types.U32{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	// Create a hash as randomness
	hash := hashUtil.Blake2bHash([]byte{0x00, 0x01, 0x02, 0x03})
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
func GenerateRandomHash() types.OpaqueHash {
	randomBytes := make([]byte, 32)
	_, err := rand.Read(randomBytes)
	if err != nil {
		panic(err)
	}
	return hashUtil.Blake2bHash(randomBytes)
}

func factorial(n int) int {
	if n == 0 {
		return 1
	}
	return n * factorial(n-1)
}

func printStatisicTable(original []types.U32, counts map[string]int, iterations int, tolerancePercentage int) {
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
	const debugMode = false

	counts := make(map[string]int)
	original := []types.U32{1, 2, 3}

	for i := 0; i < iterations; i++ {
		hash := GenerateRandomHash()
		shuffled := Shuffle(original, hash)
		key := fmt.Sprint(shuffled)
		counts[key]++
	}

	expectedCount := iterations / len(counts)
	tolerance := expectedCount * tolerancePercentage / 100

	if debugMode {
		printStatisicTable(original, counts, iterations, tolerancePercentage)
	}

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

// createNumericSlice creates a numeric slice from 0 to n-1
func createNumericSlice(n types.U32) []types.U32 {
	slice := make([]types.U32, n)
	for i := types.U32(0); i < n; i++ {
		slice[i] = i
	}
	return slice
}

// TestShuffleWithJamTestVectors tests the Shuffle function with the test
// vectors from jamtestvectors repository.
// However, it's still in pull request and not merged yet.
// https://github.com/w3f/jamtestvectors/blob/bd1247eae47bab10de1fa4be14a644b85e923024/shuffle/shuffle_tests.json
func TestShuffleWithJamTestVectors(t *testing.T) {
	type TestCase struct {
		Input   types.U32   `json:"input"`
		Entropy string      `json:"entropy"`
		Output  []types.U32 `json:"output"`
	}

	// Open the JSON file
	file, err := os.Open("shuffle_tests.json")
	if err != nil {
		t.Errorf("Error opening file: %v", err)
		return
	}
	defer file.Close()

	// Read the file content
	byteValue, err := io.ReadAll(file)
	if err != nil {
		t.Errorf("Error reading file: %v", err)
		return
	}

	// Unmarshal the JSON data
	var testCases []TestCase
	err = json.Unmarshal(byteValue, &testCases)
	if err != nil {
		t.Errorf("Error unmarshalling JSON: %v", err)
		return
	}

	for _, testCase := range testCases {
		inputSequence := createNumericSlice(testCase.Input)
		entropy, _ := hex.DecodeString(testCase.Entropy)
		shuffled := Shuffle(inputSequence, types.OpaqueHash(entropy))

		if !reflect.DeepEqual(shuffled, testCase.Output) {
			t.Errorf("The output of Shuffle is not correct")
		}
	}
}
