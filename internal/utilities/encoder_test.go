package utilities

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
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

func TestEncodeCodec(t *testing.T) {
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
				fmt.Println("❌", "[Codec]", filename)
				t.Errorf("encoded data does not match the binary data")
			} else {
				fmt.Println("✅", "[Codec]", filename)
			}
		}
	}
}

func getTargetExtensionFiles(dir string, extension string) ([]string, error) {
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

func TestEncodeJAMTestStatistics(t *testing.T) {
	mode := "tiny" // tiny or full
	dir := "../../pkg/test_data/jam-test-vectors/statistics/" + mode + "/"
	jsonExtention := ".json"
	binExtention := ".bin"

	// Get json files
	jsonFiles, err := getTargetExtensionFiles(dir, jsonExtention)
	if err != nil {
		t.Errorf("Failed to get json files: %v", err)
	}

	// Read the json files
	for _, jsonFile := range jsonFiles {
		// Load the json file
		filePath := dir + jsonFile
		data, err := LoadJAMTestJsonCase(filePath, reflect.TypeOf(jamtests_statistics.StatisticsTestCase{}))
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
		filename := jsonFile[:len(jsonFile)-len(jsonExtention)]
		binFile := filename + binExtention
		binData, err := LoadJAMTestBinaryCase(dir + binFile)

		// compare the encoded data with the binary data
		if string(encodeResult) != string(binData) {
			fmt.Println("❌", "[", mode, "]", filename)
			t.Errorf("encoded data does not match the binary data")
		} else {
			fmt.Println("✅", "[", mode, "]", filename)
		}
	}
}

func TestEncodeJAMTestSafrole(t *testing.T) {
	mode := "tiny" // tiny or full
	dir := "../../pkg/test_data/jam-test-vectors/safrole/" + mode + "/"
	jsonExtention := ".json"
	binExtention := ".bin"

	// Get json files
	jsonFiles, err := getTargetExtensionFiles(dir, jsonExtention)
	if err != nil {
		t.Errorf("Failed to get json files: %v", err)
	}

	// Read the json files
	for _, jsonFile := range jsonFiles {
		// Load the json file
		filePath := dir + jsonFile
		data, err := LoadJAMTestJsonCase(filePath, reflect.TypeOf(jamtests_safrole.SafroleTestCase{}))
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
		filename := jsonFile[:len(jsonFile)-len(jsonExtention)]
		binFile := filename + binExtention
		binData, err := LoadJAMTestBinaryCase(dir + binFile)

		// compare the encoded data with the binary data
		if string(encodeResult) != string(binData) {
			fmt.Println("❌", "[", mode, "]", filename)
			t.Errorf("encoded data does not match the binary data")
		} else {
			fmt.Println("✅", "[", mode, "]", filename)
		}
	}
}

func TestEncodeJAMTestReport(t *testing.T) {
	mode := "tiny" // tiny or full
	dir := "../../pkg/test_data/jam-test-vectors/reports/" + mode + "/"
	jsonExtention := ".json"
	binExtention := ".bin"

	// Get json files
	jsonFiles, err := getTargetExtensionFiles(dir, jsonExtention)
	if err != nil {
		t.Errorf("Failed to get json files: %v", err)
	}

	// Read the json files
	for _, jsonFile := range jsonFiles {
		// Load the json file
		filePath := dir + jsonFile
		data, err := LoadJAMTestJsonCase(filePath, reflect.TypeOf(jamtests_reports.ReportsTestCase{}))
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
		filename := jsonFile[:len(jsonFile)-len(jsonExtention)]
		binFile := filename + binExtention
		binData, err := LoadJAMTestBinaryCase(dir + binFile)

		// compare the encoded data with the binary data
		if string(encodeResult) != string(binData) {
			fmt.Println("❌", "[", mode, "]", filename)
			t.Errorf("encoded data does not match the binary data")
		} else {
			fmt.Println("✅", "[", mode, "]", filename)
		}
	}
}

func TestEncodeJAMTestAuthorizations(t *testing.T) {
	mode := "tiny" // tiny or full
	dir := "../../pkg/test_data/jam-test-vectors/authorizations/" + mode + "/"
	jsonExtention := ".json"
	binExtention := ".bin"

	// Get json files
	jsonFiles, err := getTargetExtensionFiles(dir, jsonExtention)
	if err != nil {
		t.Errorf("Failed to get json files: %v", err)
	}

	// Read the json files
	for _, jsonFile := range jsonFiles {
		// Load the json file
		filePath := dir + jsonFile
		data, err := LoadJAMTestJsonCase(filePath, reflect.TypeOf(jamtests_authorizations.AuthorizationTestCase{}))
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
		filename := jsonFile[:len(jsonFile)-len(jsonExtention)]
		binFile := filename + binExtention
		binData, err := LoadJAMTestBinaryCase(dir + binFile)

		// compare the encoded data with the binary data
		if string(encodeResult) != string(binData) {
			fmt.Println("❌", "[", mode, "]", filename)
			t.Errorf("encoded data does not match the binary data")
		} else {
			fmt.Println("✅", "[", mode, "]", filename)
		}
	}
}

func TestEncodeJAMTestAssurances(t *testing.T) {
	mode := "tiny" // tiny or full
	dir := "../../pkg/test_data/jam-test-vectors/assurances/" + mode + "/"
	jsonExtention := ".json"
	binExtention := ".bin"

	// Get json files
	jsonFiles, err := getTargetExtensionFiles(dir, jsonExtention)
	if err != nil {
		t.Errorf("Failed to get json files: %v", err)
	}

	// Read the json files
	for _, jsonFile := range jsonFiles {
		// Load the json file
		filePath := dir + jsonFile
		data, err := LoadJAMTestJsonCase(filePath, reflect.TypeOf(jamtests_assurances.AssuranceTestCase{}))
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
		filename := jsonFile[:len(jsonFile)-len(jsonExtention)]
		binFile := filename + binExtention
		binData, err := LoadJAMTestBinaryCase(dir + binFile)

		// compare the encoded data with the binary data
		if string(encodeResult) != string(binData) {
			fmt.Println("❌", "[", mode, "]", filename)
			t.Errorf("encoded data does not match the binary data")
		} else {
			fmt.Println("✅", "[", mode, "]", filename)
		}
	}
}

func TestEncodeJAMTestDisputes(t *testing.T) {
	mode := "tiny" // tiny or full
	dir := "../../pkg/test_data/jam-test-vectors/disputes/" + mode + "/"
	jsonExtention := ".json"
	binExtention := ".bin"

	// Get json files
	jsonFiles, err := getTargetExtensionFiles(dir, jsonExtention)
	if err != nil {
		t.Errorf("Failed to get json files: %v", err)
	}

	// Read the json files
	for _, jsonFile := range jsonFiles {
		// Load the json file
		filePath := dir + jsonFile
		data, err := LoadJAMTestJsonCase(filePath, reflect.TypeOf(jamtests_disputes.DisputeTestCase{}))
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
		filename := jsonFile[:len(jsonFile)-len(jsonExtention)]
		binFile := filename + binExtention
		binData, err := LoadJAMTestBinaryCase(dir + binFile)

		// compare the encoded data with the binary data
		if string(encodeResult) != string(binData) {
			fmt.Println("❌", "[", mode, "]", filename)
			t.Errorf("encoded data does not match the binary data")
		} else {
			fmt.Println("✅", "[", mode, "]", filename)
		}
	}
}

func TestEncodeJAMTestAccumulate(t *testing.T) {
	mode := "tiny" // tiny or full
	dir := "../../pkg/test_data/jam-test-vectors/accumulate/" + mode + "/"
	jsonExtention := ".json"
	binExtention := ".bin"

	// Get json files
	jsonFiles, err := getTargetExtensionFiles(dir, jsonExtention)
	if err != nil {
		t.Errorf("Failed to get json files: %v", err)
	}

	// Read the json files
	for _, jsonFile := range jsonFiles {
		// Load the json file
		filePath := dir + jsonFile
		data, err := LoadJAMTestJsonCase(filePath, reflect.TypeOf(jamtests_accmuluate.AccumulateTestCase{}))
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
		filename := jsonFile[:len(jsonFile)-len(jsonExtention)]
		binFile := filename + binExtention
		binData, err := LoadJAMTestBinaryCase(dir + binFile)

		// compare the encoded data with the binary data
		if string(encodeResult) != string(binData) {
			fmt.Println("❌", "[", mode, "]", filename)
			t.Errorf("encoded data does not match the binary data")
		} else {
			fmt.Println("✅", "[", mode, "]", filename)
		}
	}
}

func TestEncodeJAMTestPreimages(t *testing.T) {
	dir := "../../pkg/test_data/jam-test-vectors/preimages/data/"
	jsonExtention := ".json"
	binExtention := ".bin"

	// Get json files
	jsonFiles, err := getTargetExtensionFiles(dir, jsonExtention)
	if err != nil {
		t.Errorf("Failed to get json files: %v", err)
	}

	// Read the json files
	for _, jsonFile := range jsonFiles {
		// Load the json file
		filePath := dir + jsonFile
		data, err := LoadJAMTestJsonCase(filePath, reflect.TypeOf(jamtests_preimages.PreimageTestCase{}))
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
		filename := jsonFile[:len(jsonFile)-len(jsonExtention)]
		binFile := filename + binExtention
		binData, err := LoadJAMTestBinaryCase(dir + binFile)

		// compare the encoded data with the binary data
		if string(encodeResult) != string(binData) {
			fmt.Println("❌", filename)
			t.Errorf("encoded data does not match the binary data")
		} else {
			fmt.Println("✅", filename)
		}
	}
}

func TestEncodeJAMTestHistory(t *testing.T) {
	dir := "../../pkg/test_data/jam-test-vectors/history/data/"
	jsonExtention := ".json"
	binExtention := ".bin"

	// Get json files
	jsonFiles, err := getTargetExtensionFiles(dir, jsonExtention)
	if err != nil {
		t.Errorf("Failed to get json files: %v", err)
	}

	// Read the json files
	for _, jsonFile := range jsonFiles {
		// Load the json file
		filePath := dir + jsonFile
		data, err := LoadJAMTestJsonCase(filePath, reflect.TypeOf(jamtests_history.HistoryTestCase{}))
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
		filename := jsonFile[:len(jsonFile)-len(jsonExtention)]
		binFile := filename + binExtention
		binData, err := LoadJAMTestBinaryCase(dir + binFile)

		// compare the encoded data with the binary data
		if string(encodeResult) != string(binData) {
			fmt.Println("❌", filename)
			t.Errorf("encoded data does not match the binary data")
		} else {
			fmt.Println("✅", filename)
		}
	}
}
