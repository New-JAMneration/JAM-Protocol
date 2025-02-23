package types

import (
	"encoding/json"
	"io"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// Constants
const (
	MODE                 = "full" // tiny or full
	JSON_EXTENTION       = ".json"
	BIN_EXTENTION        = ".bin"
	JAM_TEST_VECTORS_DIR = "../../pkg/test_data/jam-test-vectors/"
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

func GetJsonFilename(filename string) string {
	return filename + JSON_EXTENTION
}

func GetBinFilename(filename string) string {
	return filename + BIN_EXTENTION
}

func CompareBinaryData(data1 []byte, data2 []byte) bool {
	if len(data1) != len(data2) {
		return false
	}

	for i := 0; i < len(data1); i++ {
		if data1[i] != data2[i] {
			return false
		}
	}

	return true
}

// Codec
func TestEncodeJamTestVectorsCodec(t *testing.T) {
	// The Codec test cases only support tiny mode
	BACKUP_TEST_MODE := TEST_MODE
	if TEST_MODE != "tiny" {
		SetTinyMode()
		log.Println("⚠️  Codec test cases only support tiny mode")
	}

	testCases := map[reflect.Type][]string{
		reflect.TypeOf(AssurancesExtrinsic{}): {
			"assurances_extrinsic",
		},
		reflect.TypeOf(Block{}): {
			"block",
		},
		reflect.TypeOf(DisputesExtrinsic{}): {
			"disputes_extrinsic",
		},
		reflect.TypeOf(Extrinsic{}): {
			"extrinsic",
		},
		reflect.TypeOf(GuaranteesExtrinsic{}): {
			"guarantees_extrinsic",
		},
		reflect.TypeOf(Header{}): {
			"header_0",
			"header_1",
		},
		reflect.TypeOf(PreimagesExtrinsic{}): {
			"preimages_extrinsic",
		},
		reflect.TypeOf(RefineContext{}): {
			"refine_context",
		},
		reflect.TypeOf(TicketsExtrinsic{}): {
			"tickets_extrinsic",
		},
		reflect.TypeOf(WorkItem{}): {
			"work_item",
		},
		reflect.TypeOf(WorkPackage{}): {
			"work_package",
		},
		reflect.TypeOf(WorkReport{}): {
			"work_report",
		},
		reflect.TypeOf(WorkResult{}): {
			"work_result_0",
			"work_result_1",
		},
	}

	dir := filepath.Join(JAM_TEST_VECTORS_DIR, "codec", "data")

	for structType, fileNames := range testCases {
		for _, filename := range fileNames {
			// Read json file
			jsonFilename := GetJsonFilename(filepath.Join(dir, filename))
			data, err := LoadJAMTestJsonCase(jsonFilename, structType)
			if err != nil {
				t.Fatalf("Failed to read JSON file: %v", err)
			}

			structValue := reflect.New(structType).Elem()
			structValue.Set(reflect.ValueOf(data))

			// Encode the JSON data
			encoder := NewEncoder()
			encoded, err := encoder.Encode(structValue.Addr().Interface())
			if err != nil {
				t.Fatalf("Failed to encode JSON data: %v", err)
			}

			// Read binary file
			binFilename := GetBinFilename(filepath.Join(dir, filename))
			binData, err := LoadJAMTestBinaryCase(binFilename)
			if err != nil {
				t.Fatalf("Failed to read binary file: %v", err)
			}

			// Compare the binary data
			if !CompareBinaryData(encoded, binData) {
				log.Printf("❌ [%s] %s", TEST_MODE, filename)
				t.Fatalf("Binary data is not equal to the expected data")
			} else {
				log.Printf("✅ [%s] %s", TEST_MODE, filename)
			}
		}
	}

	// Reset the test mode
	if BACKUP_TEST_MODE == "tiny" {
		SetTinyMode()
	} else {
		SetFullMode()
	}
}
