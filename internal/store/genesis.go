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
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return data, nil
}

func DecodeBlockFromBin(data []byte) (*types.Block, error) {
	block := &types.Block{}
	decoder := types.NewDecoder()
	err := decoder.Decode(data, block)
	if err != nil {
		return nil, fmt.Errorf("failed to decode block: %w", err)
	}

	return block, nil
}

func DecodeStateFromBin(data []byte) (*types.State, error) {
	state := &types.State{}
	decoder := types.NewDecoder()
	err := decoder.Decode(data, state)
	if err != nil {
		return nil, fmt.Errorf("failed to decode state: %w", err)
	}

	return state, nil
}

// This function will read the types.GenesisBlockPath file and return the
// genesis block
func GetGenesisBlock() (*types.Block, error) {
	filename := types.GenesisBlockPath
	data, err := GetBytesFromFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to get bytes from file: %w", err)
	}

	return DecodeBlockFromBin(data)
}

// This function will read the types.GenesisStatePath file and return the
// genesis state
func GetGenesisState() (*types.State, error) {
	filename := types.GenesisStatePath
	data, err := GetBytesFromFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to get bytes from file: %w", err)
	}

	return DecodeStateFromBin(data)
}

// This function will read the genesis block from the given binary file and
// return it
func GetGenesisBlockFromBinFile(filename string) (*types.Block, error) {
	data, err := GetBytesFromFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to get bytes from file: %w", err)
	}

	return DecodeBlockFromBin(data)
}

// This function will read the genesis block from the given JSON file and return
// it
func GetGenesisBlockFromJson(filename string) (*types.Block, error) {
	data, err := GetBytesFromFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to get bytes from file: %w", err)
	}

	// Unmarshal the JSON data
	var block types.Block
	err = json.Unmarshal(data, &block)
	if err != nil {
		return nil, err
	}

	return &block, nil
}

// This function will read the genesis state from the given binary file and
// return it
func GetGenesisStateFromBinFile(filename string) (*types.State, error) {
	data, err := GetBytesFromFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to get bytes from file: %w", err)
	}

	return DecodeStateFromBin(data)
}

// This function will read the genesis state from the given JSON file and return
// it
func GetGenesisStateFromJson(filename string) (*types.State, error) {
	data, err := GetBytesFromFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to get bytes from file: %w", err)
	}

	// Unmarshal the JSON data
	var state types.State
	err = json.Unmarshal(data, &state)
	if err != nil {
		return nil, err
	}

	return &state, nil
}
