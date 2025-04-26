package utilities

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

/*
	See internal/accumulation/accumulation_test.go TestPreimageTestVectors to know how to import test vector
*/

// Constants
const (
	JSON_EXTENTION       = ".json"
	BIN_EXTENTION        = ".bin"
	JAM_TEST_VECTORS_DIR = "../../pkg/test_data/jam-test-vectors/"
	JAM_TEST_NET_DIR     = "../../pkg/test_data/jamtestnet/"
)

func GetTargetExtensionFiles(dir string, extension string) ([]string, error) {
	// Get all files in the directory
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	// Get all files with the target extension
	var targetFiles []string
	for _, file := range files {
		fileName := file.Name()
		if fileName[len(fileName)-len(extension):] == extension {
			targetFiles = append(targetFiles, fileName)
		}
	}

	return targetFiles, nil
}

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

func GetTestFromBin[T any](filename string, jamtests_case *T) error {
	data, err := GetBytesFromFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read file: %v", err)
	}

	// Decode the binary data
	decoder := types.NewDecoder()
	err = decoder.Decode(data, jamtests_case)
	if err != nil {
		return fmt.Errorf("failed to decode binary data: %v", err)
	}

	// Return the decoded data
	return nil
}

func GetTestFromJson[T any](filename string) (T, error) {
	var jamtests_case T
	data, err := GetBytesFromFile(filename)
	if err != nil {
		return jamtests_case, fmt.Errorf("failed to read file: %v", err)
	}

	// Decode the json data
	err = json.Unmarshal(data, &jamtests_case)
	if err != nil {
		return jamtests_case, fmt.Errorf("failed to decode json data: %v", err)
	}

	// Return the decoded data
	return jamtests_case, nil
}

func TestDecodeJamTestNetState(t *testing.T) {
	BACKUP_TEST_MODE := types.TEST_MODE
	if types.TEST_MODE != "tiny" {
		types.SetTinyMode()
		log.Println("⚠️  jamtestnet state test cases only support tiny mode")
	}

	dirNames := []string{
		"assurances",
		"fallback",
		"orderedaccumulation",
		"safrole",
	}

	for _, dirName := range dirNames {
		dir := filepath.Join(JAM_TEST_NET_DIR, "data", dirName, "state_snapshots")

		files, err := GetTargetExtensionFiles(dir, BIN_EXTENTION)
		if err != nil {
			t.Errorf("Error: %v", err)
		}

		for _, file := range files {
			// Read the binary file
			binPath := filepath.Join(dir, file)
			binData, err := GetBytesFromFile(binPath)
			if err != nil {
				t.Errorf("Error: %v", err)
			}

			// Decode the binary data
			decoder := types.NewDecoder()
			state := &types.State{}
			err = decoder.Decode(binData, state)
			if err != nil {
				t.Errorf("Error: %v", err)
			}

			// Read the json file
			filename := file[:len(file)-len(BIN_EXTENTION)]
			jsonFilePath := filepath.Join(dir, filename+JSON_EXTENTION)
			jsonData, err := GetTestFromJson[types.State](jsonFilePath)
			if err != nil {
				t.Errorf("Error: %v", err)
			}

			// Compare the two structs
			if !reflect.DeepEqual(state, jsonData) {
				log.Printf("❌ [%s] [%s] %s", types.TEST_MODE, dirName, file)
				t.Errorf("Error: %v", err)
			} else {
				log.Printf("✅ [%s] [%s] %s", types.TEST_MODE, dirName, file)
			}
		}
	}

	// Reset the test mode
	if BACKUP_TEST_MODE == "tiny" {
		types.SetTinyMode()
	} else {
		types.SetFullMode()
	}
}
