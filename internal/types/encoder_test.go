package types_test

import (
	"log"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities"
	jamtests_accumulate "github.com/New-JAMneration/JAM-Protocol/jamtests/accumulate"
	jamtests_assurances "github.com/New-JAMneration/JAM-Protocol/jamtests/assurances"
	jamtests_authorizations "github.com/New-JAMneration/JAM-Protocol/jamtests/authorizations"
	jamtests_disputes "github.com/New-JAMneration/JAM-Protocol/jamtests/disputes"
	jamtests_history "github.com/New-JAMneration/JAM-Protocol/jamtests/history"
	jamtests_preimages "github.com/New-JAMneration/JAM-Protocol/jamtests/preimages"
	jamtests_reports "github.com/New-JAMneration/JAM-Protocol/jamtests/reports"
	jamtests_safrole "github.com/New-JAMneration/JAM-Protocol/jamtests/safrole"
	jamtests_statistics "github.com/New-JAMneration/JAM-Protocol/jamtests/statistics"
	jamtests_trace "github.com/New-JAMneration/JAM-Protocol/jamtests/trace"
)

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
	BACKUP_TEST_MODE := types.TEST_MODE
	if types.TEST_MODE != "tiny" {
		types.SetTinyMode()
		log.Println("⚠️  Codec test cases only support tiny mode")
	}

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
			// Read json file
			jsonFilename := GetJsonFilename(filepath.Join(dir, filename))
			data, err := LoadJAMTestJsonCase(jsonFilename, structType)
			if err != nil {
				t.Fatalf("Failed to read JSON file: %v", err)
			}

			structValue := reflect.New(structType).Elem()
			structValue.Set(reflect.ValueOf(data))

			// Encode the JSON data
			encoder := types.NewEncoder()
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
				log.Printf("❌ [%s] %s", types.TEST_MODE, filename)
				t.Fatalf("Binary data is not equal to the expected data")
			} else {
				log.Printf("✅ [%s] %s", types.TEST_MODE, filename)
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

// Statistics
func TestEncodeJamTestVectorsStatistics(t *testing.T) {
	dir := filepath.Join(JAM_TEST_VECTORS_DIR, "statistics", types.TEST_MODE)

	// Read json files
	jsonFiles, err := GetTargetExtensionFiles(dir, JSON_EXTENTION)
	if err != nil {
		t.Errorf("Failed to get JSON files: %v", err)
	}

	for _, jsonFile := range jsonFiles {
		// read json file
		jsonFilePath := filepath.Join(dir, jsonFile)
		structType := reflect.TypeOf(jamtests_statistics.StatisticsTestCase{})
		data, err := LoadJAMTestJsonCase(jsonFilePath, structType)
		if err != nil {
			t.Fatalf("Failed to read JSON file: %v", err)
		}

		structValue := reflect.New(structType).Elem()
		structValue.Set(reflect.ValueOf(data))

		// Encode the JSON data
		encoder := types.NewEncoder()
		encoded, err := encoder.Encode(structValue.Addr().Interface())
		if err != nil {
			t.Fatalf("Failed to encode JSON data: %v", err)
		}

		// Read binary file
		filename := jsonFile[:len(jsonFile)-len(JSON_EXTENTION)]
		binFileName := GetBinFilename(filename)
		binFilePath := filepath.Join(dir, binFileName)
		binData, err := LoadJAMTestBinaryCase(binFilePath)
		if err != nil {
			t.Fatalf("Failed to read binary file: %v", err)
		}

		// Compare the binary data
		if !CompareBinaryData(encoded, binData) {
			log.Printf("❌ [%s] %s", types.TEST_MODE, filename)
			t.Fatalf("Binary data is not equal to the expected data")
		} else {
			log.Printf("✅ [%s] %s", types.TEST_MODE, filename)
		}
	}
}

// Safrole
func TestEncodeJamTestVectorsSafrole(t *testing.T) {
	dir := filepath.Join(JAM_TEST_VECTORS_DIR, "safrole", types.TEST_MODE)

	// Read json files
	jsonFiles, err := GetTargetExtensionFiles(dir, JSON_EXTENTION)
	if err != nil {
		t.Errorf("Failed to get JSON files: %v", err)
	}

	for _, jsonFile := range jsonFiles {
		// read json file
		jsonFilePath := filepath.Join(dir, jsonFile)
		structType := reflect.TypeOf(jamtests_safrole.SafroleTestCase{})
		data, err := LoadJAMTestJsonCase(jsonFilePath, structType)
		if err != nil {
			t.Fatalf("Failed to read JSON file: %v", err)
		}

		structValue := reflect.New(structType).Elem()
		structValue.Set(reflect.ValueOf(data))

		// Encode the JSON data
		encoder := types.NewEncoder()
		encoded, err := encoder.Encode(structValue.Addr().Interface())
		if err != nil {
			t.Fatalf("Failed to encode JSON data: %v", err)
		}

		// Read binary file
		filename := jsonFile[:len(jsonFile)-len(JSON_EXTENTION)]
		binFileName := GetBinFilename(filename)
		binFilePath := filepath.Join(dir, binFileName)
		binData, err := LoadJAMTestBinaryCase(binFilePath)
		if err != nil {
			t.Fatalf("Failed to read binary file: %v", err)
		}

		// Compare the binary data
		if !CompareBinaryData(encoded, binData) {
			log.Printf("❌ [%s] %s", types.TEST_MODE, filename)
			t.Fatalf("Binary data is not equal to the expected data")
		} else {
			log.Printf("✅ [%s] %s", types.TEST_MODE, filename)
		}
	}
}

// Reports
func TestEncodeJamTestVectorsReports(t *testing.T) {
	dir := filepath.Join(JAM_TEST_VECTORS_DIR, "reports", types.TEST_MODE)

	// Read json files
	jsonFiles, err := GetTargetExtensionFiles(dir, JSON_EXTENTION)
	if err != nil {
		t.Errorf("Failed to get JSON files: %v", err)
	}

	for _, jsonFile := range jsonFiles {
		// read json file
		jsonFilePath := filepath.Join(dir, jsonFile)
		structType := reflect.TypeOf(jamtests_reports.ReportsTestCase{})
		data, err := LoadJAMTestJsonCase(jsonFilePath, structType)
		if err != nil {
			t.Fatalf("Failed to read JSON file: %v", err)
		}

		structValue := reflect.New(structType).Elem()
		structValue.Set(reflect.ValueOf(data))

		// Encode the JSON data
		encoder := types.NewEncoder()
		encoded, err := encoder.Encode(structValue.Addr().Interface())
		if err != nil {
			t.Fatalf("Failed to encode JSON data: %v", err)
		}

		// Read binary file
		filename := jsonFile[:len(jsonFile)-len(JSON_EXTENTION)]
		binFileName := GetBinFilename(filename)
		binFilePath := filepath.Join(dir, binFileName)
		binData, err := LoadJAMTestBinaryCase(binFilePath)
		if err != nil {
			t.Fatalf("Failed to read binary file: %v", err)
		}

		// Compare the binary data
		if !CompareBinaryData(encoded, binData) {
			log.Printf("❌ [%s] %s", types.TEST_MODE, filename)
			t.Fatalf("Binary data is not equal to the expected data")
		} else {
			log.Printf("✅ [%s] %s", types.TEST_MODE, filename)
		}
	}
}

// Disputes
func TestEncodeJamTestVectorsDisputes(t *testing.T) {
	dir := filepath.Join(JAM_TEST_VECTORS_DIR, "disputes", types.TEST_MODE)

	// Read json files
	jsonFiles, err := GetTargetExtensionFiles(dir, JSON_EXTENTION)
	if err != nil {
		t.Errorf("Failed to get JSON files: %v", err)
	}

	for _, jsonFile := range jsonFiles {
		// read json file
		jsonFilePath := filepath.Join(dir, jsonFile)
		structType := reflect.TypeOf(jamtests_disputes.DisputeTestCase{})
		data, err := LoadJAMTestJsonCase(jsonFilePath, structType)
		if err != nil {
			t.Fatalf("Failed to read JSON file: %v", err)
		}

		structValue := reflect.New(structType).Elem()
		structValue.Set(reflect.ValueOf(data))

		// Encode the JSON data
		encoder := types.NewEncoder()
		encoded, err := encoder.Encode(structValue.Addr().Interface())
		if err != nil {
			t.Fatalf("Failed to encode JSON data: %v", err)
		}

		// Read binary file
		filename := jsonFile[:len(jsonFile)-len(JSON_EXTENTION)]
		binFileName := GetBinFilename(filename)
		binFilePath := filepath.Join(dir, binFileName)
		binData, err := LoadJAMTestBinaryCase(binFilePath)
		if err != nil {
			t.Fatalf("Failed to read binary file: %v", err)
		}

		// Compare the binary data
		if !CompareBinaryData(encoded, binData) {
			log.Printf("❌ [%s] %s", types.TEST_MODE, filename)
			t.Fatalf("Binary data is not equal to the expected data")
		} else {
			log.Printf("✅ [%s] %s", types.TEST_MODE, filename)
		}
	}
}

// Assurances
func TestEncodeJamTestVectorsAssurances(t *testing.T) {
	dir := filepath.Join(JAM_TEST_VECTORS_DIR, "assurances", types.TEST_MODE)

	// Read json files
	jsonFiles, err := GetTargetExtensionFiles(dir, JSON_EXTENTION)
	if err != nil {
		t.Errorf("Failed to get JSON files: %v", err)
	}

	for _, jsonFile := range jsonFiles {
		// read json file
		jsonFilePath := filepath.Join(dir, jsonFile)
		structType := reflect.TypeOf(jamtests_assurances.AssuranceTestCase{})

		data, err := LoadJAMTestJsonCase(jsonFilePath, structType)
		if err != nil {
			t.Fatalf("Failed to read JSON file: %v", err)
		}

		structValue := reflect.New(structType).Elem()
		structValue.Set(reflect.ValueOf(data))

		// Encode the JSON data
		encoder := types.NewEncoder()
		encoded, err := encoder.Encode(structValue.Addr().Interface())
		if err != nil {
			t.Fatalf("Failed to encode JSON data: %v", err)
		}

		// Read binary file
		filename := jsonFile[:len(jsonFile)-len(JSON_EXTENTION)]
		binFileName := GetBinFilename(filename)
		binFilePath := filepath.Join(dir, binFileName)
		binData, err := LoadJAMTestBinaryCase(binFilePath)
		if err != nil {
			t.Fatalf("Failed to read binary file: %v", err)
		}

		// Compare the binary data
		if !CompareBinaryData(encoded, binData) {
			log.Printf("❌ [%s] %s", types.TEST_MODE, filename)
			t.Fatalf("Binary data is not equal to the expected data")
		} else {
			log.Printf("✅ [%s] %s", types.TEST_MODE, filename)
		}
	}
}

// Authorizations
func TestEncodeJamTestVectorsAuthorizations(t *testing.T) {
	dir := filepath.Join(JAM_TEST_VECTORS_DIR, "authorizations", types.TEST_MODE)

	// Read json files
	jsonFiles, err := GetTargetExtensionFiles(dir, JSON_EXTENTION)
	if err != nil {
		t.Errorf("Failed to get JSON files: %v", err)
	}

	for _, jsonFile := range jsonFiles {
		// read json file
		jsonFilePath := filepath.Join(dir, jsonFile)
		structType := reflect.TypeOf(jamtests_authorizations.AuthorizationTestCase{})
		data, err := LoadJAMTestJsonCase(jsonFilePath, structType)
		if err != nil {
			t.Fatalf("Failed to read JSON file: %v", err)
		}

		structValue := reflect.New(structType).Elem()
		structValue.Set(reflect.ValueOf(data))

		// Encode the JSON data
		encoder := types.NewEncoder()
		encoded, err := encoder.Encode(structValue.Addr().Interface())
		if err != nil {
			t.Fatalf("Failed to encode JSON data: %v", err)
		}

		// Read binary file
		filename := jsonFile[:len(jsonFile)-len(JSON_EXTENTION)]
		binFileName := GetBinFilename(filename)
		binFilePath := filepath.Join(dir, binFileName)
		binData, err := LoadJAMTestBinaryCase(binFilePath)
		if err != nil {
			t.Fatalf("Failed to read binary file: %v", err)
		}

		// Compare the binary data
		if !CompareBinaryData(encoded, binData) {
			log.Printf("❌ [%s] %s", types.TEST_MODE, filename)
			t.Fatalf("Binary data is not equal to the expected data")
		} else {
			log.Printf("✅ [%s] %s", types.TEST_MODE, filename)
		}
	}
}

// Accumulate
func TestEncodeJamTestVectorsAccumulate(t *testing.T) {
	dir := filepath.Join(JAM_TEST_VECTORS_DIR, "accumulate", types.TEST_MODE)

	// Read json files
	jsonFiles, err := GetTargetExtensionFiles(dir, JSON_EXTENTION)
	if err != nil {
		t.Errorf("Failed to get JSON files: %v", err)
	}

	for _, jsonFile := range jsonFiles {
		// read json file
		jsonFilePath := filepath.Join(dir, jsonFile)
		structType := reflect.TypeOf(jamtests_accumulate.AccumulateTestCase{})
		data, err := LoadJAMTestJsonCase(jsonFilePath, structType)
		if err != nil {
			t.Fatalf("Failed to read JSON file: %v", err)
		}

		structValue := reflect.New(structType).Elem()
		structValue.Set(reflect.ValueOf(data))

		// Encode the JSON data
		encoder := types.NewEncoder()
		encoded, err := encoder.Encode(structValue.Addr().Interface())
		if err != nil {
			t.Fatalf("Failed to encode JSON data: %v", err)
		}

		// Read binary file
		filename := jsonFile[:len(jsonFile)-len(JSON_EXTENTION)]
		binFileName := GetBinFilename(filename)
		binFilePath := filepath.Join(dir, binFileName)
		binData, err := LoadJAMTestBinaryCase(binFilePath)
		if err != nil {
			t.Fatalf("Failed to read binary file: %v", err)
		}

		// Compare the binary data
		if !CompareBinaryData(encoded, binData) {
			log.Printf("❌ [%s] %s", types.TEST_MODE, filename)
			t.Fatalf("Binary data is not equal to the expected data")
		} else {
			log.Printf("✅ [%s] %s", types.TEST_MODE, filename)
		}
	}
}

// Preimages
func TestEncodeJamTestVectorsPreimages(t *testing.T) {
	BACKUP_TEST_MODE := types.TEST_MODE
	if types.TEST_MODE != "tiny" {
		types.SetTinyMode()
		log.Println("⚠️  Preimages test cases only support tiny mode")
	}

	dir := filepath.Join(JAM_TEST_VECTORS_DIR, "preimages", "data")

	// Read json files
	jsonFiles, err := GetTargetExtensionFiles(dir, JSON_EXTENTION)
	if err != nil {
		t.Errorf("Failed to get JSON files: %v", err)
	}

	for _, jsonFile := range jsonFiles {
		// read json file
		jsonFilePath := filepath.Join(dir, jsonFile)
		structType := reflect.TypeOf(jamtests_preimages.PreimageTestCase{})
		data, err := LoadJAMTestJsonCase(jsonFilePath, structType)
		if err != nil {
			t.Fatalf("Failed to read JSON file: %v", err)
		}

		structValue := reflect.New(structType).Elem()
		structValue.Set(reflect.ValueOf(data))

		// Encode the JSON data
		encoder := types.NewEncoder()
		encoded, err := encoder.Encode(structValue.Addr().Interface())
		if err != nil {
			t.Fatalf("Failed to encode JSON data: %v", err)
		}

		// Read binary file
		filename := jsonFile[:len(jsonFile)-len(JSON_EXTENTION)]
		binFileName := GetBinFilename(filename)
		binFilePath := filepath.Join(dir, binFileName)
		binData, err := LoadJAMTestBinaryCase(binFilePath)
		if err != nil {
			t.Fatalf("Failed to read binary file: %v", err)
		}

		// Compare the binary data
		if !CompareBinaryData(encoded, binData) {
			log.Printf("❌ [%s] %s", types.TEST_MODE, filename)
			t.Fatalf("Binary data is not equal to the expected data")
		} else {
			log.Printf("✅ [%s] %s", types.TEST_MODE, filename)
		}
	}

	// Reset the test mode
	if BACKUP_TEST_MODE == "tiny" {
		types.SetTinyMode()
	} else {
		types.SetFullMode()
	}
}

// History
func TestEncodeJamTestVectorsHistory(t *testing.T) {
	BACKUP_TEST_MODE := types.TEST_MODE
	if types.TEST_MODE != "tiny" {
		types.SetTinyMode()
		log.Println("⚠️  History test cases only support tiny mode")
	}

	dir := filepath.Join(JAM_TEST_VECTORS_DIR, "history", "data")

	// Read json files
	jsonFiles, err := GetTargetExtensionFiles(dir, JSON_EXTENTION)
	if err != nil {
		t.Errorf("Failed to get JSON files: %v", err)
	}

	for _, jsonFile := range jsonFiles {
		// read json file
		jsonFilePath := filepath.Join(dir, jsonFile)
		structType := reflect.TypeOf(jamtests_history.HistoryTestCase{})
		data, err := LoadJAMTestJsonCase(jsonFilePath, structType)
		if err != nil {
			t.Fatalf("Failed to read JSON file: %v", err)
		}

		structValue := reflect.New(structType).Elem()
		structValue.Set(reflect.ValueOf(data))

		// Encode the JSON data
		encoder := types.NewEncoder()
		encoded, err := encoder.Encode(structValue.Addr().Interface())
		if err != nil {
			t.Fatalf("Failed to encode JSON data: %v", err)
		}

		// Read binary file
		filename := jsonFile[:len(jsonFile)-len(JSON_EXTENTION)]
		binFileName := GetBinFilename(filename)
		binFilePath := filepath.Join(dir, binFileName)
		binData, err := LoadJAMTestBinaryCase(binFilePath)
		if err != nil {
			t.Fatalf("Failed to read binary file: %v", err)
		}

		// Compare the binary data
		if !CompareBinaryData(encoded, binData) {
			log.Printf("❌ [%s] %s", types.TEST_MODE, filename)
			t.Fatalf("Binary data is not equal to the expected data")
		} else {
			log.Printf("✅ [%s] %s", types.TEST_MODE, filename)
		}
	}

	// Reset the test mode
	if BACKUP_TEST_MODE == "tiny" {
		types.SetTinyMode()
	} else {
		types.SetFullMode()
	}
}

func TestEncodeJamTestNetGenesisBlock(t *testing.T) {
	BACKUP_TEST_MODE := types.TEST_MODE
	if types.TEST_MODE != "tiny" {
		types.SetTinyMode()
		log.Println("⚠️  genesis block only support tiny mode")
	}

	filename := "../../pkg/test_data/jamtestnet/chainspecs/blocks/genesis-tiny.json"

	// Read json file
	structType := reflect.TypeOf(types.Block{})
	data, err := LoadJAMTestJsonCase(filename, structType)
	if err != nil {
		t.Fatalf("Failed to read JSON file: %v", err)
	}

	structValue := reflect.New(structType).Elem()
	structValue.Set(reflect.ValueOf(data))

	// Encode the JSON data
	encoder := types.NewEncoder()
	encoded, err := encoder.Encode(structValue.Addr().Interface())
	if err != nil {
		t.Fatalf("Failed to encode JSON data: %v", err)
	}

	// Read binary file
	binFilePath := "../../pkg/test_data/jamtestnet/chainspecs/blocks/genesis-tiny.bin"

	binData, err := LoadJAMTestBinaryCase(binFilePath)
	if err != nil {
		t.Fatalf("Failed to read binary file: %v", err)
	}

	// Compare the binary data
	if !CompareBinaryData(encoded, binData) {
		log.Printf("❌ [%s] %s", types.TEST_MODE, "genesis")
		t.Fatalf("Binary data is not equal to the expected data")
	} else {
		log.Printf("✅ [%s] %s", types.TEST_MODE, "genesis")
	}

	// Reset the test mode
	if BACKUP_TEST_MODE == "tiny" {
		types.SetTinyMode()
	} else {
		types.SetFullMode()
	}
}

func TestEncodeJamTestNetGenesisState(t *testing.T) {
	BACKUP_TEST_MODE := types.TEST_MODE
	if types.TEST_MODE != "tiny" {
		types.SetTinyMode()
		log.Println("⚠️  genesis state only support tiny mode")
	}

	filename := "../../pkg/test_data/jamtestnet/chainspecs/state_snapshots/genesis-tiny.json"

	// Read json file
	structType := reflect.TypeOf(types.State{})
	data, err := LoadJAMTestJsonCase(filename, structType)
	if err != nil {
		t.Fatalf("Failed to read JSON file: %v", err)
	}

	structValue := reflect.New(structType).Elem()
	structValue.Set(reflect.ValueOf(data))

	// Encode the JSON data
	encoder := types.NewEncoder()
	encoded, err := encoder.Encode(structValue.Addr().Interface())
	if err != nil {
		t.Fatalf("Failed to encode JSON data: %v", err)
	}

	// Read binary file
	binFilePath := "../../pkg/test_data/jamtestnet/chainspecs/state_snapshots/genesis-tiny.bin"

	binData, err := LoadJAMTestBinaryCase(binFilePath)
	if err != nil {
		t.Fatalf("Failed to read binary file: %v", err)
	}

	// Compare the binary data
	if !CompareBinaryData(encoded, binData) {
		log.Printf("❌ [%s] %s", types.TEST_MODE, "genesis-tiny")
		t.Fatalf("Binary data is not equal to the expected data")
	} else {
		log.Printf("✅ [%s] %s", types.TEST_MODE, "genesis-tiny")
	}

	// Reset the test mode
	if BACKUP_TEST_MODE == "tiny" {
		types.SetTinyMode()
	} else {
		types.SetFullMode()
	}
}

func TestEncodeJamTestNetBlock(t *testing.T) {
	BACKUP_TEST_MODE := types.TEST_MODE
	if types.TEST_MODE != "tiny" {
		types.SetTinyMode()
		log.Println("⚠️  jamtestnet block test cases only support tiny mode")
	}

	dirNames := []string{
		"assurances",
		"fallback",
		"orderedaccumulation",
		"safrole",
	}

	for _, dirName := range dirNames {
		dir := filepath.Join(JAM_TEST_NET_DIR, "data", dirName, "blocks")

		files, err := GetTargetExtensionFiles(dir, JSON_EXTENTION)
		if err != nil {
			t.Errorf("Error: %v", err)
		}

		for _, file := range files {
			jsonPath := filepath.Join(dir, file)
			structType := reflect.TypeOf(types.Block{})
			data, err := LoadJAMTestJsonCase(jsonPath, structType)
			if err != nil {
				t.Fatalf("Failed to read JSON file: %v", err)
			}

			structValue := reflect.New(structType).Elem()
			structValue.Set(reflect.ValueOf(data))

			// Encode the JSON data
			encoder := types.NewEncoder()
			encoded, err := encoder.Encode(structValue.Addr().Interface())
			if err != nil {
				t.Fatalf("Failed to encode JSON data: %v", err)
			}

			// Read binary file
			filename := file[:len(file)-len(JSON_EXTENTION)]
			binFileName := GetBinFilename(filename)
			binFilePath := filepath.Join(dir, binFileName)
			binData, err := LoadJAMTestBinaryCase(binFilePath)
			if err != nil {
				t.Fatalf("Failed to read binary file: %v", err)
			}

			// Compare the binary data
			if !CompareBinaryData(encoded, binData) {
				log.Printf("❌ [%s] [%s] %s", types.TEST_MODE, dirName, file)
				t.Fatalf("Binary data is not equal to the expected data")
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

// INFO: We cannot pass this test because they didn't implement the sort for the
// map
// FIXME: Waiting for the vectors to be updated to pass this test
// func TestEncodeJamTestNetState(t *testing.T) {
// 	dirNames := []string{
// 		"assurances",
// 		"fallback",
// 		"orderedaccumulation",
// 		"safrole",
// 	}

// 	for _, dirName := range dirNames {
// 		dir := filepath.Join(JAM_TEST_NET_DIR, "data", dirName, "state_snapshots")

// 		files, err := GetTargetExtensionFiles(dir, JSON_EXTENTION)
// 		if err != nil {
// 			t.Errorf("Error: %v", err)
// 		}

// 		for _, file := range files {
// 			jsonPath := filepath.Join(dir, file)
// 			structType := reflect.TypeOf(types.State{})
// 			data, err := LoadJAMTestJsonCase(jsonPath, structType)
// 			if err != nil {
// 				t.Fatalf("Failed to read JSON file: %v", err)
// 			}

// 			structValue := reflect.New(structType).Elem()
// 			structValue.Set(reflect.ValueOf(data))

// 			// Encode the JSON data
// 			encoder := types.NewEncoder()
// 			encoded, err := encoder.Encode(structValue.Addr().Interface())
// 			if err != nil {
// 				t.Fatalf("Failed to encode JSON data: %v", err)
// 			}

// 			// Read binary file
// 			filename := file[:len(file)-len(JSON_EXTENTION)]
// 			binFileName := GetBinFilename(filename)
// 			binFilePath := filepath.Join(dir, binFileName)
// 			binData, err := LoadJAMTestBinaryCase(binFilePath)
// 			if err != nil {
// 				t.Fatalf("Failed to read binary file: %v", err)
// 			}

// 			// Compare the binary data
// 			if !CompareBinaryData(encoded, binData) {
// 				log.Printf("❌ [%s] [%s] %s", types.TEST_MODE, dirName, file)
// 				t.Errorf("Error: %v", err)
// 			} else {
// 				log.Printf("✅ [%s] [%s] %s", types.TEST_MODE, dirName, file)
// 			}
// 		}
// 	}
// }

// Encode json and decode the json, we have to get the same data
func TestEncodeDecodeJamTestNetState(t *testing.T) {
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

		files, err := GetTargetExtensionFiles(dir, JSON_EXTENTION)
		if err != nil {
			t.Errorf("Error: %v", err)
		}

		for _, file := range files {
			jsonPath := filepath.Join(dir, file)
			structType := reflect.TypeOf(types.State{})
			data, err := LoadJAMTestJsonCase(jsonPath, structType)
			if err != nil {
				t.Fatalf("Failed to read JSON file: %v", err)
			}

			structValue := reflect.New(structType).Elem()
			structValue.Set(reflect.ValueOf(data))

			// Encode the JSON data
			encoder := types.NewEncoder()
			encoded, err := encoder.Encode(structValue.Addr().Interface())
			if err != nil {
				t.Fatalf("Failed to encode JSON data: %v", err)
			}

			// Decode the encoded data
			decoder := types.NewDecoder()
			decoded := types.State{}
			err = decoder.Decode(encoded, &decoded)
			if err != nil {
				t.Fatalf("Failed to decode encoded data: %v", err)
			}

			// Compare two state struct
			if !reflect.DeepEqual(data, decoded) {
				log.Printf("❌ [%s] [%s] %s", types.TEST_MODE, dirName, file)
				t.Errorf("Decoded data is not equal to the expected data")
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

func TestEncodeJamTestNetTransitions(t *testing.T) {
	dirnames := []string{
		"assurances",
		"generic",
		"orderedaccumulation",
	}

	for _, dirname := range dirnames {
		dir := filepath.Join(utilities.JAM_TEST_NET_DIR, "data", dirname, "state_transitions")
		jsonTestFiles, err := GetTargetExtensionFiles(dir, JSON_EXTENTION)
		if err != nil {
			t.Fatalf("Failed to get JSON files: %v", err)
		}
		binTestFiles, err := GetTargetExtensionFiles(dir, BIN_EXTENTION)
		if err != nil {
			t.Fatalf("Failed to get BIN files: %v", err)
		}

		for i := 0; i < len(jsonTestFiles); i++ {
			jsonTestFile := filepath.Join(dir, jsonTestFiles[i])
			binTestFile := filepath.Join(dir, binTestFiles[i])

			// Decode the JSON data
			jsonData, err := utilities.GetTestFromJson[jamtests_trace.TraceTestCase](jsonTestFile)
			if err != nil {
				t.Fatalf("Failed to decode JSON data: %v", err)
			}

			// Encode the JSON data
			encoder := types.NewEncoder()
			encoded, err := encoder.Encode(&jsonData)
			if err != nil {
				t.Fatalf("Failed to encode JSON data: %v", err)
			}

			// Read the binary file
			binData, err := utilities.GetBytesFromFile(binTestFile)
			if err != nil {
				t.Fatalf("Failed to read binary file: %v", err)
			}

			// Compare the binary data
			if !CompareBinaryData(encoded, binData) {
				log.Printf("❌ [%s] %s", dirname, jsonTestFiles[i])
				t.Fatalf("Binary data is not equal to the expected data")
			} else {
				log.Printf("✅ [%s] %s", dirname, jsonTestFiles[i])
			}
		}
	}
}
