package types_test

import (
	"encoding/json"
	"io"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities"
)

// Constants
const (
	MODE                 = "tiny" // tiny or full
	JSON_EXTENTION       = ".json"
	BIN_EXTENTION        = ".bin"
	JAM_TEST_VECTORS_DIR = "../../pkg/test_data/jam-test-vectors/"
)

func TestDecoderBlock(t *testing.T) {
	blockBinPath := filepath.Join(JAM_TEST_VECTORS_DIR, "codec", "data", "block.bin")
	blockBinData, err := LoadJAMTestBinaryCase(blockBinPath)
	if err != nil {
		t.Errorf("Error: %v", err)
	}

	block := types.Block{}
	decoder := types.NewDecoder()
	err = decoder.Decode(blockBinData, &block)
	if err != nil {
		log.Fatal(err)
	}

	// Load json file
	blockJsonPath := filepath.Join(JAM_TEST_VECTORS_DIR, "codec", "data", "block.json")
	blockJsonData, err := LoadJAMTestJsonCase(blockJsonPath, reflect.TypeOf(types.Block{}))
	if err != nil {
		t.Errorf("Error: %v", err)
	}

	// Convert blockJsonData to Block struct
	blockJson := blockJsonData.(types.Block)

	// convert the block to string and compare
	blockString, _ := json.Marshal(block)
	blockJsonString, _ := json.Marshal(blockJson)

	if string(blockString) != string(blockJsonString) {
		t.Errorf("Error: %v", err)
	}

	encoder := utilities.NewEncoder()
	encodedBlock, err := encoder.Encode(block)
	if err != nil {
		t.Errorf("Error: %v", err)
	}

	// Compare the two blocks
	if !reflect.DeepEqual(blockBinData, encodedBlock) {
		t.Errorf("Error: %v", err)
	}
}

func LoadJAMTestJsonCase(filename string, structType reflect.Type) (interface{}, error) {
	// Create a new instance of the struct
	structValue := reflect.New(structType).Elem()

	// Open the file
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
	err = json.Unmarshal(byteValue, structValue.Addr().Interface())
	if err != nil {
		return nil, err
	}

	// Return the struct
	return structValue.Interface(), nil
}

func LoadJAMTestBinaryCase(filename string) ([]byte, error) {
	// read binary file and return byte array
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

	return byteValue, nil
}
