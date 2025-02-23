package types_test

import (
	"log"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	jamtests_accmuluate "github.com/New-JAMneration/JAM-Protocol/jamtests/accumulate"
	jamtests_assurances "github.com/New-JAMneration/JAM-Protocol/jamtests/assurances"
	jamtests_authorizations "github.com/New-JAMneration/JAM-Protocol/jamtests/authorizations"
	jamtests_disputes "github.com/New-JAMneration/JAM-Protocol/jamtests/disputes"
	jamtests_reports "github.com/New-JAMneration/JAM-Protocol/jamtests/reports"
	jamtests_safrole "github.com/New-JAMneration/JAM-Protocol/jamtests/safrole"
	jamtests_statistics "github.com/New-JAMneration/JAM-Protocol/jamtests/statistics"
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
		structType := reflect.TypeOf(jamtests_accmuluate.AccumulateTestCase{})
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
