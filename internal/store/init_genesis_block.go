package store

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

func GetBytesFromFile(filename string) ([]byte, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %v", err)
	}

	return data, nil
}

func GetGenesisBlockFromBin() (*types.Block, error) {
	filename := "../../pkg/test_data/jamtestnet/chainspecs/blocks/genesis-tiny.bin"
	data, err := GetBytesFromFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to get bytes from file: %v", err)
	}

	block := &types.Block{}
	decoder := types.NewDecoder()
	err = decoder.Decode(data, block)
	if err != nil {
		return nil, fmt.Errorf("failed to decode block: %v", err)
	}

	return block, nil
}

func GetGenesisBlockFromJson(filename string) (*types.Block, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Read the file content
	byteValue, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	// Unmarshal the JSON data
	var block types.Block
	err = json.Unmarshal(byteValue, &block)
	if err != nil {
		return nil, err
	}

	return &block, nil
}
