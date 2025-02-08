package utilities

import (
	"encoding/json"
	"io"
	"os"
	"reflect"
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

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

func TestEncoder(t *testing.T) {
	testCases := map[reflect.Type][]string{
		reflect.TypeOf(types.AssurancesExtrinsic{}): {
			"assurances_extrinsic",
		},
		reflect.TypeOf(types.Block{}): {
			"block",
		},
		reflect.TypeOf(types.DisputesExtrinsic{}): {
			"disputes_extrinsic",
		},
		reflect.TypeOf(types.Extrinsic{}): {
			"extrinsic",
		},
		reflect.TypeOf(types.GuaranteesExtrinsic{}): {
			"guarantees_extrinsic",
		},
		reflect.TypeOf(types.Header{}): {
			"header_0",
			"header_1",
		},
		reflect.TypeOf(types.PreimagesExtrinsic{}): {
			"preimages_extrinsic",
		},
		reflect.TypeOf(types.RefineContext{}): {
			"refine_context",
		},
		reflect.TypeOf(types.TicketsExtrinsic{}): {
			"tickets_extrinsic",
		},
		reflect.TypeOf(types.WorkItem{}): {
			"work_item",
		},
		reflect.TypeOf(types.WorkPackage{}): {
			"work_package",
		},
		reflect.TypeOf(types.WorkReport{}): {
			"work_report",
		},
		reflect.TypeOf(types.WorkResult{}): {
			"work_result_0",
			"work_result_1",
		},
	}

	dir := "../../pkg/test_data/jam-test-vectors/codec/data/"
	jsonExtention := ".json"
	binExtention := ".bin"

	for structType, fileNames := range testCases {
		for _, filename := range fileNames {
			filePath := dir + filename + jsonExtention
			data, err := LoadJAMTestJsonCase(filePath, structType)
			if err != nil {
				t.Errorf("Failed to load test case from %s: %v", filename, err)
			}

			encoder := NewEncoder()
			encodeResult, err := encoder.Encode(data)
			if err != nil {
				t.Errorf("Failed to encode test case from %s: %v", filename, err)
			}

			binFilePath := dir + filename + binExtention
			binData, err := LoadJAMTestBinaryCase(binFilePath)

			// compare the encoded data with the binary data
			if string(encodeResult) != string(binData) {
				t.Errorf("encoded data does not match the binary data")
			}
		}
	}
}
