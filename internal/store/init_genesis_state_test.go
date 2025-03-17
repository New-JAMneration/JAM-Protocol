package store

import (
	"encoding/json"
	"io"
	"os"
	"reflect"
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

func ReadFile(filename string) ([]byte, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Read the file content
	bytes, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	return bytes, nil
}

func TestLoadGenesisState(t *testing.T) {
	jsonFile := "../../pkg/test_data/jamtestnet/chainspecs/state_snapshots/genesis-tiny.json"
	bytes, err := ReadFile(jsonFile)
	if err != nil {
		t.Fatalf("Error reading file: %v", err)
	}

	// Parse it with types.State
	var state types.State
	err = json.Unmarshal(bytes, &state)
	if err != nil {
		t.Fatalf("Error unmarshalling JSON: %v", err)
	}

	// Compare the result with the expected value from binary file
	binFile := "../../pkg/test_data/jamtestnet/chainspecs/state_snapshots/genesis-tiny.bin"
	bytes, err = ReadFile(binFile)
	if err != nil {
		t.Fatalf("Error reading file: %v", err)
	}

	// Decode the binary file
	var expected types.State
	decoder := types.NewDecoder()
	err = decoder.Decode(bytes, &expected)
	if err != nil {
		t.Fatalf("Error decoding binary: %v", err)
	}

	// Compare two states with reflect.DeepEqual
	if !reflect.DeepEqual(state, expected) {
		t.Errorf("State does not match expected value: got %v, expected %v", state, expected)
	}
}
