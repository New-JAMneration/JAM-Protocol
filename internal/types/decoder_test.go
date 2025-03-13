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
	MODE                 = "full" // tiny or full
	JSON_EXTENTION       = ".json"
	BIN_EXTENTION        = ".bin"
	JAM_TEST_VECTORS_DIR = "../../pkg/test_data/jam-test-vectors/"
	JAM_TEST_NET_DIR     = "../../pkg/test_data/jamtestnet/"
)

func TestMain(m *testing.M) {
	// Set the test mode
	types.SetTestMode()

	// Run the tests
	os.Exit(m.Run())
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

// Codec
func TestDecodeJamTestVectorsCodec(t *testing.T) {
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
			// Read the binary file
			binPath := filepath.Join(dir, filename+BIN_EXTENTION)
			binData, err := LoadJAMTestBinaryCase(binPath)
			if err != nil {
				t.Errorf("Error: %v", err)
			}

			// Decode the binary data
			decoder := types.NewDecoder()
			structValue := reflect.New(structType).Elem()
			err = decoder.Decode(binData, structValue.Addr().Interface())
			if err != nil {
				t.Errorf("Error: %v", err)
			}

			// Read the json file
			jsonPath := filepath.Join(dir, filename+JSON_EXTENTION)
			jsonData, err := LoadJAMTestJsonCase(jsonPath, structType)
			if err != nil {
				t.Errorf("Error: %v", err)
			}

			// Compare the two structs
			if !reflect.DeepEqual(structValue.Interface(), jsonData) {
				log.Printf("❌ [%s] %s", types.TEST_MODE, filename)
				t.Errorf("Error: %v", err)
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
func TestDecodeJamTestVectorsStatistics(t *testing.T) {
	dir := filepath.Join(JAM_TEST_VECTORS_DIR, "statistics", types.TEST_MODE)

	// Read binary files
	binFiles, err := GetTargetExtensionFiles(dir, BIN_EXTENTION)
	if err != nil {
		t.Errorf("Error: %v", err)
	}

	for _, binFile := range binFiles {
		// Read the binary file
		binPath := filepath.Join(dir, binFile)
		binData, err := LoadJAMTestBinaryCase(binPath)
		if err != nil {
			t.Errorf("Error: %v", err)
		}

		// Decode the binary data
		decoder := types.NewDecoder()
		statistics := &jamtests_statistics.StatisticsTestCase{}
		err = decoder.Decode(binData, statistics)
		if err != nil {
			t.Errorf("Error: %v", err)
		}

		// Read the json file
		filename := binFile[:len(binFile)-len(BIN_EXTENTION)]
		jsonFileName := GetJsonFilename(filename)
		jsonFilePath := filepath.Join(dir, jsonFileName)
		jsonData, err := LoadJAMTestJsonCase(jsonFilePath, reflect.TypeOf(&jamtests_statistics.StatisticsTestCase{}))
		if err != nil {
			t.Errorf("Error: %v", err)
		}

		// Compare the two structs
		if !reflect.DeepEqual(statistics, jsonData) {
			log.Printf("❌ [%s] %s", types.TEST_MODE, binFile)
			t.Errorf("Error: %v", err)
		} else {
			log.Printf("✅ [%s] %s", types.TEST_MODE, binFile)
		}
	}
}

// Safrole
func TestDecodeJamTestVectorsSafrole(t *testing.T) {
	dir := filepath.Join(JAM_TEST_VECTORS_DIR, "safrole", types.TEST_MODE)

	// Read binary files
	binFiles, err := GetTargetExtensionFiles(dir, BIN_EXTENTION)
	if err != nil {
		t.Errorf("Error: %v", err)
	}

	for _, binFile := range binFiles {
		// Read the binary file
		binPath := filepath.Join(dir, binFile)
		binData, err := LoadJAMTestBinaryCase(binPath)
		if err != nil {
			t.Errorf("Error: %v", err)
		}

		// Decode the binary data
		decoder := types.NewDecoder()
		safrole := &jamtests_safrole.SafroleTestCase{}
		err = decoder.Decode(binData, safrole)
		if err != nil {
			t.Errorf("Error: %v", err)
		}

		// Read the json file
		filename := binFile[:len(binFile)-len(BIN_EXTENTION)]
		jsonFileName := GetJsonFilename(filename)
		jsonFilePath := filepath.Join(dir, jsonFileName)
		jsonData, err := LoadJAMTestJsonCase(jsonFilePath, reflect.TypeOf(&jamtests_safrole.SafroleTestCase{}))
		if err != nil {
			t.Errorf("Error: %v", err)
		}

		// Compare the two structs
		if !reflect.DeepEqual(safrole, jsonData) {
			log.Printf("❌ [%s] %s", types.TEST_MODE, binFile)
			t.Errorf("Error: %v", err)
		} else {
			log.Printf("✅ [%s] %s", types.TEST_MODE, binFile)
		}
	}
}

// Reports
func TestDecodeJamTestVectorsReports(t *testing.T) {
	dir := filepath.Join(JAM_TEST_VECTORS_DIR, "reports", types.TEST_MODE)

	// Read binary files
	binFiles, err := GetTargetExtensionFiles(dir, BIN_EXTENTION)
	if err != nil {
		t.Errorf("Error: %v", err)
	}

	for _, binFile := range binFiles {
		// Read the binary file
		binPath := filepath.Join(dir, binFile)
		binData, err := LoadJAMTestBinaryCase(binPath)
		if err != nil {
			t.Errorf("Error: %v", err)
		}

		// Decode the binary data
		decoder := types.NewDecoder()
		reports := &jamtests_reports.ReportsTestCase{}
		err = decoder.Decode(binData, reports)
		if err != nil {
			t.Errorf("Error: %v", err)
		}

		// Read the json file
		filename := binFile[:len(binFile)-len(BIN_EXTENTION)]
		jsonFileName := GetJsonFilename(filename)
		jsonFilePath := filepath.Join(dir, jsonFileName)
		jsonData, err := LoadJAMTestJsonCase(jsonFilePath, reflect.TypeOf(&jamtests_reports.ReportsTestCase{}))
		if err != nil {
			t.Errorf("Error: %v", err)
		}

		// Compare the two structs
		if !reflect.DeepEqual(reports, jsonData) {
			log.Printf("❌ [%s] %s", types.TEST_MODE, binFile)
			t.Errorf("Error: %v", err)
		} else {
			log.Printf("✅ [%s] %s", types.TEST_MODE, binFile)
		}
	}
}

// Disputes
func TestDecodeJamTestVectorsDisputes(t *testing.T) {
	dir := filepath.Join(JAM_TEST_VECTORS_DIR, "disputes", types.TEST_MODE)

	// Read binary files
	binFiles, err := GetTargetExtensionFiles(dir, BIN_EXTENTION)
	if err != nil {
		t.Errorf("Error: %v", err)
	}

	for _, binFile := range binFiles {
		// Read the binary file
		binPath := filepath.Join(dir, binFile)
		binData, err := LoadJAMTestBinaryCase(binPath)
		if err != nil {
			t.Errorf("Error: %v", err)
		}

		// Decode the binary data
		decoder := types.NewDecoder()
		disputes := &jamtests_disputes.DisputeTestCase{}
		err = decoder.Decode(binData, disputes)
		if err != nil {
			t.Errorf("Error: %v", err)
		}

		// Read the json file
		filename := binFile[:len(binFile)-len(BIN_EXTENTION)]
		jsonFileName := GetJsonFilename(filename)
		jsonFilePath := filepath.Join(dir, jsonFileName)
		jsonData, err := LoadJAMTestJsonCase(jsonFilePath, reflect.TypeOf(&jamtests_disputes.DisputeTestCase{}))
		if err != nil {
			t.Errorf("Error: %v", err)
		}

		// Compare the two structs
		if !reflect.DeepEqual(disputes, jsonData) {
			log.Printf("❌ [%s] %s", types.TEST_MODE, binFile)
			t.Errorf("Error: %v", err)
		} else {
			log.Printf("✅ [%s] %s", types.TEST_MODE, binFile)
		}
	}
}

// Assurances
func TestDecodeJamTestVectorsAssurances(t *testing.T) {
	dir := filepath.Join(JAM_TEST_VECTORS_DIR, "assurances", types.TEST_MODE)

	// Read binary files
	binFiles, err := GetTargetExtensionFiles(dir, BIN_EXTENTION)
	if err != nil {
		t.Errorf("Error: %v", err)
	}

	for _, binFile := range binFiles {
		// Read the binary file
		binPath := filepath.Join(dir, binFile)
		binData, err := LoadJAMTestBinaryCase(binPath)
		if err != nil {
			t.Errorf("Error: %v", err)
		}

		// Decode the binary data
		decoder := types.NewDecoder()
		assurances := &jamtests_assurances.AssuranceTestCase{}
		err = decoder.Decode(binData, assurances)
		if err != nil {
			t.Errorf("Error: %v", err)
		}

		// Read the json file
		filename := binFile[:len(binFile)-len(BIN_EXTENTION)]
		jsonFileName := GetJsonFilename(filename)
		jsonFilePath := filepath.Join(dir, jsonFileName)
		jsonData, err := LoadJAMTestJsonCase(jsonFilePath, reflect.TypeOf(&jamtests_assurances.AssuranceTestCase{}))
		if err != nil {
			t.Errorf("Error: %v", err)
		}

		// Compare the two structs
		if !reflect.DeepEqual(assurances, jsonData) {
			log.Printf("❌ [%s] %s", types.TEST_MODE, binFile)
			t.Errorf("Error: %v", err)
		} else {
			log.Printf("✅ [%s] %s", types.TEST_MODE, binFile)
		}
	}
}

// authorizations
func TestDecodeJamTestVectorsAuthorizations(t *testing.T) {
	dir := filepath.Join(JAM_TEST_VECTORS_DIR, "authorizations", types.TEST_MODE)

	// Read binary files
	binFiles, err := GetTargetExtensionFiles(dir, BIN_EXTENTION)
	if err != nil {
		t.Errorf("Error: %v", err)
	}

	for _, binFile := range binFiles {
		// Read the binary file
		binPath := filepath.Join(dir, binFile)
		binData, err := LoadJAMTestBinaryCase(binPath)
		if err != nil {
			t.Errorf("Error: %v", err)
		}

		// Decode the binary data
		decoder := types.NewDecoder()
		authorizations := &jamtests_authorizations.AuthorizationTestCase{}
		err = decoder.Decode(binData, authorizations)
		if err != nil {
			t.Errorf("Error: %v", err)
		}

		// Read the json file
		filename := binFile[:len(binFile)-len(BIN_EXTENTION)]
		jsonFileName := GetJsonFilename(filename)
		jsonFilePath := filepath.Join(dir, jsonFileName)
		jsonData, err := LoadJAMTestJsonCase(jsonFilePath, reflect.TypeOf(&jamtests_authorizations.AuthorizationTestCase{}))
		if err != nil {
			t.Errorf("Error: %v", err)
		}

		// Compare the two structs
		if !reflect.DeepEqual(authorizations, jsonData) {
			log.Printf("❌ [%s] %s", types.TEST_MODE, binFile)
			t.Errorf("Error: %v", err)
		} else {
			log.Printf("✅ [%s] %s", types.TEST_MODE, binFile)
		}
	}
}

// accumulate
func TestDecodeJamTestVectorsAccumulate(t *testing.T) {
	dir := filepath.Join(JAM_TEST_VECTORS_DIR, "accumulate", types.TEST_MODE)

	// Read binary files
	binFiles, err := GetTargetExtensionFiles(dir, BIN_EXTENTION)
	if err != nil {
		t.Errorf("Error: %v", err)
	}

	for _, binFile := range binFiles {
		// Read the binary file
		binPath := filepath.Join(dir, binFile)
		binData, err := LoadJAMTestBinaryCase(binPath)
		if err != nil {
			t.Errorf("Error: %v", err)
		}

		// Decode the binary data
		decoder := types.NewDecoder()
		accumulate := &jamtests_accmuluate.AccumulateTestCase{}
		err = decoder.Decode(binData, accumulate)
		if err != nil {
			t.Errorf("Error: %v", err)
		}

		// Read the json file
		filename := binFile[:len(binFile)-len(BIN_EXTENTION)]
		jsonFileName := GetJsonFilename(filename)
		jsonFilePath := filepath.Join(dir, jsonFileName)
		jsonData, err := LoadJAMTestJsonCase(jsonFilePath, reflect.TypeOf(&jamtests_accmuluate.AccumulateTestCase{}))
		if err != nil {
			t.Errorf("Error: %v", err)
		}

		// Compare the two structs
		if !reflect.DeepEqual(accumulate, jsonData) {
			log.Printf("❌ [%s] %s", types.TEST_MODE, binFile)
			t.Errorf("Error: %v", err)
		} else {
			log.Printf("✅ [%s] %s", types.TEST_MODE, binFile)
		}
	}
}

// preimages
func TestDecodeJamTestVectorsPreimages(t *testing.T) {
	BACKUP_TEST_MODE := types.TEST_MODE
	if types.TEST_MODE != "tiny" {
		types.SetTinyMode()
		log.Println("⚠️  Preimages test cases only support tiny mode")
	}

	dir := filepath.Join(JAM_TEST_VECTORS_DIR, "preimages", "data")

	// Read binary files
	binFiles, err := GetTargetExtensionFiles(dir, BIN_EXTENTION)
	if err != nil {
		t.Errorf("Error: %v", err)
	}

	for _, binFile := range binFiles {
		// Read the binary file
		binPath := filepath.Join(dir, binFile)
		binData, err := LoadJAMTestBinaryCase(binPath)
		if err != nil {
			t.Errorf("Error: %v", err)
		}

		// Decode the binary data
		decoder := types.NewDecoder()
		preimages := &jamtests_preimages.PreimageTestCase{}
		err = decoder.Decode(binData, preimages)
		if err != nil {
			t.Errorf("Error: %v", err)
		}

		// Read the json file
		filename := binFile[:len(binFile)-len(BIN_EXTENTION)]
		jsonFileName := GetJsonFilename(filename)
		jsonFilePath := filepath.Join(dir, jsonFileName)
		jsonData, err := LoadJAMTestJsonCase(jsonFilePath, reflect.TypeOf(&jamtests_preimages.PreimageTestCase{}))
		if err != nil {
			t.Errorf("Error: %v", err)
		}

		// Compare the two structs
		if !reflect.DeepEqual(preimages, jsonData) {
			log.Printf("❌ [%s] %s", types.TEST_MODE, binFile)
			t.Errorf("Error: %v", err)
		} else {
			log.Printf("✅ [%s] %s", types.TEST_MODE, binFile)
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
func TestDecodeJamTestVectorsHistory(t *testing.T) {
	BACKUP_TEST_MODE := types.TEST_MODE
	if types.TEST_MODE != "tiny" {
		types.SetTinyMode()
		log.Println("⚠️  History test cases only support tiny mode")
	}

	dir := filepath.Join(JAM_TEST_VECTORS_DIR, "history", "data")

	// Read binary files
	binFiles, err := GetTargetExtensionFiles(dir, BIN_EXTENTION)
	if err != nil {
		t.Errorf("Error: %v", err)
	}

	for _, binFile := range binFiles {
		// Read the binary file
		binPath := filepath.Join(dir, binFile)
		binData, err := LoadJAMTestBinaryCase(binPath)
		if err != nil {
			t.Errorf("Error: %v", err)
		}

		// Decode the binary data
		decoder := types.NewDecoder()
		history := &jamtests_history.HistoryTestCase{}
		err = decoder.Decode(binData, history)
		if err != nil {
			t.Errorf("Error: %v", err)
		}

		// Read the json file
		filename := binFile[:len(binFile)-len(BIN_EXTENTION)]
		jsonFileName := GetJsonFilename(filename)
		jsonFilePath := filepath.Join(dir, jsonFileName)
		jsonData, err := LoadJAMTestJsonCase(jsonFilePath, reflect.TypeOf(&jamtests_history.HistoryTestCase{}))
		if err != nil {
			t.Errorf("Error: %v", err)
		}

		// Compare the two structs
		if !reflect.DeepEqual(history, jsonData) {
			log.Printf("❌ [%s] %s", types.TEST_MODE, binFile)
			t.Errorf("Error: %v", err)
		} else {
			log.Printf("✅ [%s] %s", types.TEST_MODE, binFile)
		}
	}

	// Reset the test mode
	if BACKUP_TEST_MODE == "tiny" {
		types.SetTinyMode()
	} else {
		types.SetFullMode()
	}
}

func ReadBinaryFile(filename string) ([]byte, error) {
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

func TestDecodeJamTestNetGenesisBlock(t *testing.T) {
	BACKUP_TEST_MODE := types.TEST_MODE
	if types.TEST_MODE != "tiny" {
		types.SetTinyMode()
		log.Println("⚠️  Genesis block only support tiny mode")
	}

	filename := "../../pkg/test_data/jamtestnet/chainspecs/blocks/genesis-tiny.bin"

	// Read the binary file
	binData, err := ReadBinaryFile(filename)
	if err != nil {
		t.Errorf("Error: %v", err)
	}

	// Decode the binary data
	decoder := types.NewDecoder()
	block := &types.Block{}
	err = decoder.Decode(binData, block)
	if err != nil {
		t.Errorf("Error: %v", err)
	}

	// Read the json file
	jsonFilePath := "../../pkg/test_data/jamtestnet/chainspecs/blocks/genesis-tiny.json"
	jsonData, err := LoadJAMTestJsonCase(jsonFilePath, reflect.TypeOf(&types.Block{}))
	if err != nil {
		t.Errorf("Error: %v", err)
	}

	// Compare the two structs
	if !reflect.DeepEqual(block, jsonData) {
		log.Printf("❌ [%s] %s", types.TEST_MODE, filename)
		t.Errorf("Error: %v", err)
	} else {
		log.Printf("✅ [%s] %s", types.TEST_MODE, filename)
	}

	// Reset the test mode
	if BACKUP_TEST_MODE == "tiny" {
		types.SetTinyMode()
	} else {
		types.SetFullMode()
	}
}

func TestDecodeJamTestNetBlock(t *testing.T) {
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

		files, err := GetTargetExtensionFiles(dir, BIN_EXTENTION)
		if err != nil {
			t.Errorf("Error: %v", err)
		}

		for _, file := range files {
			// Read the binary file
			binPath := filepath.Join(dir, file)
			binData, err := ReadBinaryFile(binPath)
			if err != nil {
				t.Errorf("Error: %v", err)
			}

			// Decode the binary data
			decoder := types.NewDecoder()
			block := &types.Block{}
			err = decoder.Decode(binData, block)
			if err != nil {
				t.Errorf("Error: %v", err)
			}

			// Read the json file
			filename := file[:len(file)-len(BIN_EXTENTION)]
			jsonFileName := GetJsonFilename(filename)
			jsonFilePath := filepath.Join(dir, jsonFileName)
			jsonData, err := LoadJAMTestJsonCase(jsonFilePath, reflect.TypeOf(&types.Block{}))
			if err != nil {
				t.Errorf("Error: %v", err)
			}

			// Compare the two structs
			if !reflect.DeepEqual(block, jsonData) {
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
			binData, err := ReadBinaryFile(binPath)
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
			jsonFileName := GetJsonFilename(filename)
			jsonFilePath := filepath.Join(dir, jsonFileName)
			jsonData, err := LoadJAMTestJsonCase(jsonFilePath, reflect.TypeOf(&types.State{}))
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

func TestDecodeJamTestNetGenesisState(t *testing.T) {
	BACKUP_TEST_MODE := types.TEST_MODE
	if types.TEST_MODE != "tiny" {
		types.SetTinyMode()
		log.Println("⚠️  genesis state only support tiny mode")
	}

	filename := "../../pkg/test_data/jamtestnet/chainspecs/state_snapshots/genesis-tiny.bin"

	// Read the binary file
	binData, err := ReadBinaryFile(filename)
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
	jsonFilePath := "../../pkg/test_data/jamtestnet/chainspecs/state_snapshots/genesis-tiny.json"
	jsonData, err := LoadJAMTestJsonCase(jsonFilePath, reflect.TypeOf(&types.State{}))
	if err != nil {
		t.Errorf("Error: %v", err)
	}

	// Compare the two structs
	if !reflect.DeepEqual(state, jsonData) {
		log.Printf("❌ [%s] %s", types.TEST_MODE, "genesis-tiny")
		t.Errorf("Error: %v", err)
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
