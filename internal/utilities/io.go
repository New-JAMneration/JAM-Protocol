package utilities

import (
	"fmt"
	"io"
	"os"

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
