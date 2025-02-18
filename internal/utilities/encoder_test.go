package utilities

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"

	jamtests_accmuluate "github.com/New-JAMneration/JAM-Protocol/jamtests/accumulate"
	jamtests_assurances "github.com/New-JAMneration/JAM-Protocol/jamtests/assurances"
	jamtests_authorizations "github.com/New-JAMneration/JAM-Protocol/jamtests/authorizations"
	jamtests_disputes "github.com/New-JAMneration/JAM-Protocol/jamtests/disputes"
	jamtests_history "github.com/New-JAMneration/JAM-Protocol/jamtests/history"
	jamtests_preimages "github.com/New-JAMneration/JAM-Protocol/jamtests/preimages"
	jamtests_reports "github.com/New-JAMneration/JAM-Protocol/jamtests/reports"
	jamtests_safrole "github.com/New-JAMneration/JAM-Protocol/jamtests/safrole"
	jamtests_statistics "github.com/New-JAMneration/JAM-Protocol/jamtests/statistics"
)

// Constants
const (
	MODE                 = "tiny" // tiny or full
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

func GetJsonFilename(filename string) string {
	return filename + JSON_EXTENTION
}

func GetBinFilename(filename string) string {
	return filename + BIN_EXTENTION
}

func TestEncodeJamTestVectorsCodec(t *testing.T) {
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

	dir := filepath.Join(JAM_TEST_VECTORS_DIR, "codec", "data")

	for structType, fileNames := range testCases {
		for _, filename := range fileNames {
			// Read the json file
			jsonFileName := GetJsonFilename(filename)
			jsonFilePath := filepath.Join(dir, jsonFileName)
			data, err := LoadJAMTestJsonCase(jsonFilePath, structType)
			if err != nil {
				t.Errorf("Failed to load test case from %s: %v", filename, err)
			}

			// Encode the data
			encoder := NewEncoder()
			encodeResult, err := encoder.Encode(data)
			if err != nil {
				t.Errorf("Failed to encode test case from %s: %v", filename, err)
			}

			// Read the binary file
			binFilename := GetBinFilename(filename)
			binFilePath := filepath.Join(dir, binFilename)
			binData, err := LoadJAMTestBinaryCase(binFilePath)

			// Compare the encoded data with the binary data
			if !CompareBinaryData(encodeResult, binData) {
				fmt.Println("❌", "[ ---- ]", filename)
				t.Errorf("encoded data does not match the binary data")
			} else {
				fmt.Println("✅", "[ ---- ]", filename)
			}
		}
	}
}

func TestEncodeJamTestVectorsStatistics(t *testing.T) {
	dir := filepath.Join(JAM_TEST_VECTORS_DIR, "statistics", MODE)

	// Get json files
	jsonFiles, err := GetTargetExtensionFiles(dir, JSON_EXTENTION)
	if err != nil {
		t.Errorf("Failed to get json files: %v", err)
	}

	// Read the json files
	for _, jsonFile := range jsonFiles {
		// Load the json file
		jsonFilePath := filepath.Join(dir, jsonFile)
		data, err := LoadJAMTestJsonCase(jsonFilePath, reflect.TypeOf(jamtests_statistics.StatisticsTestCase{}))
		if err != nil {
			t.Errorf("Failed to load test case from %s: %v", jsonFile, err)
		}

		// Encode the data
		encoder := NewEncoder()
		encodeResult, err := encoder.Encode(data)
		if err != nil {
			t.Errorf("Failed to encode test case from %s: %v", jsonFile, err)
		}

		// Load the bin file
		filename := jsonFile[:len(jsonFile)-len(JSON_EXTENTION)]
		binFileName := GetBinFilename(filename)
		binFilePath := filepath.Join(dir, binFileName)
		binData, err := LoadJAMTestBinaryCase(binFilePath)

		// compare the encoded data with the binary data
		if !CompareBinaryData(encodeResult, binData) {
			fmt.Println("❌", "[", MODE, "]", filename)
			t.Errorf("encoded data does not match the binary data")
		} else {
			fmt.Println("✅", "[", MODE, "]", filename)
		}
	}
}

func TestEncodeJamTestVectorsSafrole(t *testing.T) {
	dir := filepath.Join(JAM_TEST_VECTORS_DIR, "safrole", MODE)

	// Get json files
	jsonFiles, err := GetTargetExtensionFiles(dir, JSON_EXTENTION)
	if err != nil {
		t.Errorf("Failed to get json files: %v", err)
	}

	// Read the json files
	for _, jsonFile := range jsonFiles {
		// Load the json file
		jsonFilePath := filepath.Join(dir, jsonFile)
		data, err := LoadJAMTestJsonCase(jsonFilePath, reflect.TypeOf(jamtests_safrole.SafroleTestCase{}))
		if err != nil {
			t.Errorf("Failed to load test case from %s: %v", jsonFile, err)
		}

		// Encode the data
		encoder := NewEncoder()
		encodeResult, err := encoder.Encode(data)
		if err != nil {
			t.Errorf("Failed to encode test case from %s: %v", jsonFile, err)
		}

		// Load the bin file
		filename := jsonFile[:len(jsonFile)-len(JSON_EXTENTION)]
		binFileName := GetBinFilename(filename)
		binFilePath := filepath.Join(dir, binFileName)
		binData, err := LoadJAMTestBinaryCase(binFilePath)

		// compare the encoded data with the binary data
		if !CompareBinaryData(encodeResult, binData) {
			fmt.Println("❌", "[", MODE, "]", filename)
			t.Errorf("encoded data does not match the binary data")
		} else {
			fmt.Println("✅", "[", MODE, "]", filename)
		}
	}
}

func TestEncodeJamTestVectorsReport(t *testing.T) {
	dir := filepath.Join(JAM_TEST_VECTORS_DIR, "reports", MODE)

	// Get json files
	jsonFiles, err := GetTargetExtensionFiles(dir, JSON_EXTENTION)
	if err != nil {
		t.Errorf("Failed to get json files: %v", err)
	}

	// Read the json files
	for _, jsonFile := range jsonFiles {
		// Load the json file
		jsonFilePath := filepath.Join(dir, jsonFile)
		data, err := LoadJAMTestJsonCase(jsonFilePath, reflect.TypeOf(jamtests_reports.ReportsTestCase{}))
		if err != nil {
			t.Errorf("Failed to load test case from %s: %v", jsonFile, err)
		}

		// Encode the data
		encoder := NewEncoder()
		encodeResult, err := encoder.Encode(data)
		if err != nil {
			t.Errorf("Failed to encode test case from %s: %v", jsonFile, err)
		}

		// Load the bin file
		filename := jsonFile[:len(jsonFile)-len(JSON_EXTENTION)]
		binFileName := GetBinFilename(filename)
		binFilePath := filepath.Join(dir, binFileName)
		binData, err := LoadJAMTestBinaryCase(binFilePath)

		// compare the encoded data with the binary data
		if !CompareBinaryData(encodeResult, binData) {
			fmt.Println("❌", "[", MODE, "]", filename)
			t.Errorf("encoded data does not match the binary data")
		} else {
			fmt.Println("✅", "[", MODE, "]", filename)
		}
	}
}

func TestEncodeJamTestVectorsAuthorizations(t *testing.T) {
	dir := filepath.Join(JAM_TEST_VECTORS_DIR, "authorizations", MODE)

	// Get json files
	jsonFiles, err := GetTargetExtensionFiles(dir, JSON_EXTENTION)
	if err != nil {
		t.Errorf("Failed to get json files: %v", err)
	}

	// Read the json files
	for _, jsonFile := range jsonFiles {
		// Load the json file
		jsonFilePath := filepath.Join(dir, jsonFile)
		data, err := LoadJAMTestJsonCase(jsonFilePath, reflect.TypeOf(jamtests_authorizations.AuthorizationTestCase{}))
		if err != nil {
			t.Errorf("Failed to load test case from %s: %v", jsonFile, err)
		}

		// Encode the data
		encoder := NewEncoder()
		encodeResult, err := encoder.Encode(data)
		if err != nil {
			t.Errorf("Failed to encode test case from %s: %v", jsonFile, err)
		}

		// Load the bin file
		filename := jsonFile[:len(jsonFile)-len(JSON_EXTENTION)]
		binFileName := GetBinFilename(filename)
		binFilePath := filepath.Join(dir, binFileName)
		binData, err := LoadJAMTestBinaryCase(binFilePath)

		// compare the encoded data with the binary data
		if !CompareBinaryData(encodeResult, binData) {
			fmt.Println("❌", "[", MODE, "]", filename)
			t.Errorf("encoded data does not match the binary data")
		} else {
			fmt.Println("✅", "[", MODE, "]", filename)
		}
	}
}

func TestEncodeJamTestVectorsAssurances(t *testing.T) {
	dir := filepath.Join(JAM_TEST_VECTORS_DIR, "assurances", MODE)

	// Get json files
	jsonFiles, err := GetTargetExtensionFiles(dir, JSON_EXTENTION)
	if err != nil {
		t.Errorf("Failed to get json files: %v", err)
	}

	// Read the json files
	for _, jsonFile := range jsonFiles {
		// Load the json file
		jsonFilePath := filepath.Join(dir, jsonFile)
		data, err := LoadJAMTestJsonCase(jsonFilePath, reflect.TypeOf(jamtests_assurances.AssurancesTestCase{}))
		if err != nil {
			t.Errorf("Failed to load test case from %s: %v", jsonFile, err)
		}

		// Encode the data
		encoder := NewEncoder()
		encodeResult, err := encoder.Encode(data)
		if err != nil {
			t.Errorf("Failed to encode test case from %s: %v", jsonFile, err)
		}

		// Load the bin file
		filename := jsonFile[:len(jsonFile)-len(JSON_EXTENTION)]
		binFileName := GetBinFilename(filename)
		binFilePath := filepath.Join(dir, binFileName)
		binData, err := LoadJAMTestBinaryCase(binFilePath)

		// compare the encoded data with the binary data
		if !CompareBinaryData(encodeResult, binData) {
			fmt.Println("❌", "[", MODE, "]", filename)
			t.Errorf("encoded data does not match the binary data")
		} else {
			fmt.Println("✅", "[", MODE, "]", filename)
		}
	}
}

func TestEncodeJamTestVectorsDisputes(t *testing.T) {
	dir := filepath.Join(JAM_TEST_VECTORS_DIR, "disputes", MODE)

	// Get json files
	jsonFiles, err := GetTargetExtensionFiles(dir, JSON_EXTENTION)
	if err != nil {
		t.Errorf("Failed to get json files: %v", err)
	}

	// Read the json files
	for _, jsonFile := range jsonFiles {
		// Load the json file
		jsonFilePath := filepath.Join(dir, jsonFile)
		data, err := LoadJAMTestJsonCase(jsonFilePath, reflect.TypeOf(jamtests_disputes.DisputesTestCase{}))
		if err != nil {
			t.Errorf("Failed to load test case from %s: %v", jsonFile, err)
		}

		// Encode the data
		encoder := NewEncoder()
		encodeResult, err := encoder.Encode(data)
		if err != nil {
			t.Errorf("Failed to encode test case from %s: %v", jsonFile, err)
		}

		// Load the bin file
		filename := jsonFile[:len(jsonFile)-len(JSON_EXTENTION)]
		binFileName := GetBinFilename(filename)
		binFilePath := filepath.Join(dir, binFileName)
		binData, err := LoadJAMTestBinaryCase(binFilePath)

		// compare the encoded data with the binary data
		if !CompareBinaryData(encodeResult, binData) {
			fmt.Println("❌", "[", MODE, "]", filename)
			t.Errorf("encoded data does not match the binary data")
		} else {
			fmt.Println("✅", "[", MODE, "]", filename)
		}
	}
}

func TestEncodeJamTestVectorsAccumulate(t *testing.T) {
	dir := filepath.Join(JAM_TEST_VECTORS_DIR, "accumulate", MODE)

	// Get json files
	jsonFiles, err := GetTargetExtensionFiles(dir, JSON_EXTENTION)
	if err != nil {
		t.Errorf("Failed to get json files: %v", err)
	}

	// Read the json files
	for _, jsonFile := range jsonFiles {
		// Load the json file
		jsonFilePath := filepath.Join(dir, jsonFile)
		data, err := LoadJAMTestJsonCase(jsonFilePath, reflect.TypeOf(jamtests_accmuluate.AccumulateTestCase{}))
		if err != nil {
			t.Errorf("Failed to load test case from %s: %v", jsonFile, err)
		}

		// Encode the data
		encoder := NewEncoder()
		encodeResult, err := encoder.Encode(data)
		if err != nil {
			t.Errorf("Failed to encode test case from %s: %v", jsonFile, err)
		}

		// Load the bin file
		filename := jsonFile[:len(jsonFile)-len(JSON_EXTENTION)]
		binFileName := GetBinFilename(filename)
		binFilePath := filepath.Join(dir, binFileName)
		binData, err := LoadJAMTestBinaryCase(binFilePath)

		// compare the encoded data with the binary data
		if !CompareBinaryData(encodeResult, binData) {
			fmt.Println("❌", "[", MODE, "]", filename)
			t.Errorf("encoded data does not match the binary data")
		} else {
			fmt.Println("✅", "[", MODE, "]", filename)
		}
	}
}

func TestEncodeJamTestVectorsPreimages(t *testing.T) {
	dir := filepath.Join(JAM_TEST_VECTORS_DIR, "preimages", "data")

	// Get json files
	jsonFiles, err := GetTargetExtensionFiles(dir, JSON_EXTENTION)
	if err != nil {
		t.Errorf("Failed to get json files: %v", err)
	}

	// Read the json files
	for _, jsonFile := range jsonFiles {
		// Load the json file
		jsonFilePath := filepath.Join(dir, jsonFile)
		data, err := LoadJAMTestJsonCase(jsonFilePath, reflect.TypeOf(jamtests_preimages.PreimageTestCase{}))
		if err != nil {
			t.Errorf("Failed to load test case from %s: %v", jsonFile, err)
		}

		// Encode the data
		encoder := NewEncoder()
		encodeResult, err := encoder.Encode(data)
		if err != nil {
			t.Errorf("Failed to encode test case from %s: %v", jsonFile, err)
		}

		// Load the bin file
		filename := jsonFile[:len(jsonFile)-len(JSON_EXTENTION)]
		binFileName := GetBinFilename(filename)
		binFilePath := filepath.Join(dir, binFileName)
		binData, err := LoadJAMTestBinaryCase(binFilePath)

		// compare the encoded data with the binary data
		if !CompareBinaryData(encodeResult, binData) {
			fmt.Println("❌", "[ ---- ]", filename)
			t.Errorf("encoded data does not match the binary data")
		} else {
			fmt.Println("✅", "[ ---- ]", filename)
		}
	}
}

func TestEncodeJamTestVectorsHistory(t *testing.T) {
	dir := filepath.Join(JAM_TEST_VECTORS_DIR, "history", "data")

	// Get json files
	jsonFiles, err := GetTargetExtensionFiles(dir, JSON_EXTENTION)
	if err != nil {
		t.Errorf("Failed to get json files: %v", err)
	}

	// Read the json files
	for _, jsonFile := range jsonFiles {
		// Load the json file
		jsonFilePath := filepath.Join(dir, jsonFile)
		data, err := LoadJAMTestJsonCase(jsonFilePath, reflect.TypeOf(jamtests_history.HistoryTestCase{}))
		if err != nil {
			t.Errorf("Failed to load test case from %s: %v", jsonFile, err)
		}

		// Encode the data
		encoder := NewEncoder()
		encodeResult, err := encoder.Encode(data)
		if err != nil {
			t.Errorf("Failed to encode test case from %s: %v", jsonFile, err)
		}

		// Load the bin file
		filename := jsonFile[:len(jsonFile)-len(JSON_EXTENTION)]
		binFileName := GetBinFilename(filename)
		binFilePath := filepath.Join(dir, binFileName)
		binData, err := LoadJAMTestBinaryCase(binFilePath)

		// compare the encoded data with the binary data
		if !CompareBinaryData(encodeResult, binData) {
			fmt.Println("❌", "[ ---- ]", filename)
			t.Errorf("encoded data does not match the binary data")
		} else {
			fmt.Println("✅", "[ ---- ]", filename)
		}
	}
}
