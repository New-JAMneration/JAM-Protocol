package store

import (
	"encoding/json"
	"io"
	"os"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

func GetInitGenesisBlock(filename string) (types.Block, error) {
	file, err := os.Open(filename)
	if err != nil {
		return types.Block{}, err
	}
	defer file.Close()

	// Read the file content
	byteValue, err := io.ReadAll(file)
	if err != nil {
		return types.Block{}, err
	}

	// Unmarshal the JSON data
	var block types.Block
	err = json.Unmarshal(byteValue, &block)
	if err != nil {
		return types.Block{}, err
	}

	return block, nil
}
