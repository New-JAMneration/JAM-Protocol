package types

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"reflect"
	"testing"
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

func GetBlockFromJson(filename string) (*Block, error) {
	data, err := GetBytesFromFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to get bytes from file: %v", err)
	}

	// Unmarshal the JSON data
	var block Block
	err = json.Unmarshal(data, &block)
	if err != nil {
		return nil, err
	}

	return &block, nil
}

func TestDecodeBlock(t *testing.T) {
	blockBinFile := "../../pkg/test_data/jamtestnet/chainspecs/blocks/genesis-tiny.bin"

	// Get the bytes from the file
	data, err := GetBytesFromFile(blockBinFile)
	if err != nil {
		t.Errorf("Error getting bytes from file: %v", err)
	}

	// Create a new block (pointer)
	block := &Block{}

	// Decode the block from the binary file
	decoder := NewDecoder()
	err = decoder.Decode(data, block)
	if err != nil {
		t.Errorf("Error decoding block: %v", err)
	}

	// Check the decode block is same as expected block
	// Read json file and unmarshal it to get the expected block
	blockJsonFile := "../../pkg/test_data/jamtestnet/chainspecs/blocks/genesis-tiny.json"
	expectedBlock, err := GetBlockFromJson(blockJsonFile)
	if err != nil {
		t.Errorf("Error getting block from json file: %v", err)
	}

	// Compare the two blocks
	if !reflect.DeepEqual(block, expectedBlock) {
		t.Errorf("Blocks do not match")
	}
}

func TestDecodeU64(t *testing.T) {
	data := []byte{0, 228, 11, 84, 2, 0, 0, 0}

	decoder := NewDecoder()
	u64 := U64(0)
	err := decoder.Decode(data, &u64)
	if err != nil {
		t.Errorf("Error decoding U64: %v", err)
	}

	expected := U64(10000000000)

	if u64 != expected {
		t.Errorf("Decoded U64 does not match expected")
	}
}

func TestDecodeServiceId(t *testing.T) {
	data := []byte{100, 0, 0, 0, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32}

	decoder := NewDecoder()
	serviceId := ServiceId(0)
	err := decoder.Decode(data, &serviceId)
	if err != nil {
		t.Errorf("Error decoding ServiceId: %v", err)
	}

	expected := ServiceId(100)

	if serviceId != expected {
		t.Errorf("Decoded ServiceId does not match expected")
	}
}
