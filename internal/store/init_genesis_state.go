package store

import (
	"encoding/json"
	"io"
	"os"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

func GetInitGenesisState(filename string) (types.StateTest, error) {
	file, err := os.Open(filename)
	if err != nil {
		return types.StateTest{}, err
	}
	defer file.Close()

	// Read the file content
	byteValue, err := io.ReadAll(file)
	if err != nil {
		return types.StateTest{}, err
	}

	// Unmarshal the JSON data
	var state types.StateTest
	err = json.Unmarshal(byteValue, &state)
	if err != nil {
		return types.StateTest{}, err
	}

	return state, nil
}
