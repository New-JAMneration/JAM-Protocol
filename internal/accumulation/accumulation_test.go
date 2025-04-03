package accumulation

import (
	"encoding/json"
	"io"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/store"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	jamtests_accumulate "github.com/New-JAMneration/JAM-Protocol/jamtests/accumulate"
	jamtests_preimages "github.com/New-JAMneration/JAM-Protocol/jamtests/preimages"
)

func TestDecodeJamTestVectorsAccumulateFile(t *testing.T) {
	file := "../../pkg/test_data/jam-test-vectors/accumulate/tiny/accumulate_ready_queued_reports-1.bin"

	data, err := store.GetBytesFromFile(file)
	if err != nil {
		t.Errorf("Error getting bytes from file: %v", err)
	}

	decoder := types.NewDecoder()
	accumulateTestCase := &jamtests_accumulate.AccumulateTestCase{}
	err = decoder.Decode(data, accumulateTestCase)
	if err != nil {
		t.Errorf("Error decoding AccumulateTestCase: %v", err)
	}

	// You can access the fields of the AccumulateTestCase struct
	// e.g. accumulateTestCase.Input.Slot

	// // Chapter 12.3
	// err = DeferredTransfers()
	// if err != nil {
	// 	t.Errorf("Error in DeferredTransfers: %v", err)
	// }
}

// ---

// Constants
const (
	MODE                 = "tiny" // tiny or full
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

		// Set up test input state
		inputDelta := make(types.ServiceAccountState)
		for _, delta := range preimages.PreState.Accounts {
			// 先創建或獲取 ServiceAccount，確保其內部 maps 已初始化
			serviceAccount := types.ServiceAccount{
				ServiceInfo:    types.ServiceInfo{},
				PreimageLookup: make(types.PreimagesMapEntry),
				LookupDict:     make(types.LookupMetaMapEntry),
				StorageDict:    make(types.Storage),
			}

			// 填充 PreimageLookup
			for _, preimage := range delta.Data.Preimages {
				serviceAccount.PreimageLookup[preimage.Hash] = preimage.Blob
			}

			// 填充 LookupDict
			for _, lookup := range delta.Data.LookupMeta {
				serviceAccount.LookupDict[types.LookupMetaMapkey{
					Hash:   lookup.Key.Hash,
					Length: lookup.Key.Length,
				}] = lookup.Val
			}

			// 將 ServiceAccount 存入 inputDelta
			inputDelta[delta.Id] = serviceAccount
		}
		inputEp := preimages.Input.Preimages
		inputSlot := preimages.Input.Slot
		// Get store instance and required states
		s := store.GetInstance()
		s.GetProcessingBlockPointer().SetPreimagesExtrinsic(inputEp)
		s.GetIntermediateStates().SetDeltaDoubleDagger(inputDelta)
		s.GetPosteriorStates().SetTau(inputSlot)
		accumulateErr := ProcessPreimageExtrinsics()
		// Get output state
		outputDelta := s.GetPosteriorStates().GetDelta()
		// Validate output state
		if preimages.Output.Err != nil {
			if accumulateErr == nil {
				t.Logf("🔴 [%s] %s", types.TEST_MODE, binFile)
				t.Fatalf("Should raise Error %v but got %v", preimages.Output.Err, accumulateErr)
			} else {
				t.Logf("Error: %v", accumulateErr)
				t.Logf("🔴 [%s] %s", types.TEST_MODE, binFile)
			}
		} else {
			if !reflect.DeepEqual(outputDelta, inputDelta) {
				t.Logf("❌ [%s] %s", types.TEST_MODE, binFile)
				t.Fatalf("Error: %v", accumulateErr)
			} else {
				t.Logf("🟢 [%s] %s", types.TEST_MODE, binFile)
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

func TestDecodeJamTestVectorsAccumulate(t *testing.T) {
	dir := filepath.Join(JAM_TEST_VECTORS_DIR, "accumulate", MODE)

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
		accumulate := &jamtests_accumulate.AccumulateTestCase{}
		err = decoder.Decode(binData, accumulate)
		if err != nil {
			t.Errorf("Error: %v", err)
		}

		// Read the json file
		filename := strings.TrimSuffix(binFile, BIN_EXTENTION)
		jsonFileName := GetJsonFilename(filename)
		jsonFilePath := filepath.Join(dir, jsonFileName)
		jsonData, err := LoadJAMTestJsonCase(jsonFilePath, reflect.TypeOf(&jamtests_accumulate.AccumulateTestCase{}))
		if err != nil {
			t.Errorf("Error: %v", err)
		}

		// Compare the two structs
		if !reflect.DeepEqual(accumulate, jsonData) {
			t.Logf("❌ [%s] %s", MODE, filename)
			t.Errorf("Error: %v", err)
		} else {
			t.Logf("✅ [%s] %s", MODE, filename)
		}

		// Test accumulate
		testAccumulateFile(t, binPath)

	}
}

func testAccumulateFile(t *testing.T, binPath string) {
	// Read the binary file
	binData, err := LoadJAMTestBinaryCase(binPath)
	if err != nil {
		t.Errorf("Error: %v", err)
	}

	// Decode the binary data
	decoder := types.NewDecoder()
	testCase := &jamtests_accumulate.AccumulateTestCase{}
	err = decoder.Decode(binData, testCase)
	if err != nil {
		t.Errorf("Error: %v", err)
		return
	}

	// Setup test state
	setupTestState(testCase.PreState)

	// Execute accumulation
	result, err := Process(t, testCase.Input)
	if err != nil {
		// Check if expected error
		if testCase.Output.Err == nil {
			t.Errorf("執行時出現非預期錯誤: %v", err)
		}
		return
	}

	// Validate result
	if !reflect.DeepEqual(result, testCase.Output.Ok) {
		t.Errorf("結果不符合預期:\n實際: %+v\n預期: %+v", result, testCase.Output.Ok)
	}

	// Validate final state
	validateFinalState(t, testCase.PostState)
}

// Setup test state
func setupTestState(preState jamtests_accumulate.AccumulateState) {
	s := store.GetInstance()

	// Set time slot
	s.GetPriorStates().SetTau(preState.Slot)

	// Set entropy
	s.GetPriorStates().SetEta(types.EntropyBuffer{preState.Entropy})

	// Set accounts
	delta := types.ServiceAccountState{}
	for serviceID, account := range preState.Accounts {
		preimagesMap := types.PreimagesMapEntry{}
		for _, preimage := range account.Data.Preimages {
			preimagesMap[preimage.Hash] = types.ByteSequence(preimage.Blob)
		}
		delta[types.ServiceId(serviceID)] = types.ServiceAccount{
			ServiceInfo:    account.Data.Service,
			PreimageLookup: preimagesMap,
			LookupDict:     types.LookupMetaMapEntry{},
			StorageDict:    types.Storage{},
		}
	}
	s.GetPriorStates().SetDelta(delta)

	// Set ready queue
	// (Here needs to implement specific setting method)

	// 設置已累積報告和其他狀態
	// (這裡需要實現具體設置方法)
}

// 驗證系統的最終狀態
func validateFinalState(t *testing.T, expectedState jamtests_accumulate.AccumulateState) {
	s := store.GetInstance()

	// 驗證時間槽
	if s.GetPriorStates().GetTau() != expectedState.Slot {
		t.Errorf("時間槽不符合預期: 實際 %v, 預期 %v", s.GetPriorStates().GetTau(), expectedState.Slot)
	}

	// Validate entropy
	if !reflect.DeepEqual(s.GetPriorStates().GetEta(), types.EntropyBuffer{expectedState.Entropy}) {
		t.Errorf("熵值不符合預期")
	}

	// Validate ready queue and accumulated reports
	// (Here needs to implement specific validation method)
}

// Process function executes accumulation
func Process(t *testing.T, input jamtests_accumulate.AccumulateInput) (*types.AccumulateRoot, error) {
	// 這裡實現實際的累積邏輯
	// 處理 input.Reports 並返回結果

	// 示例:
	return &types.AccumulateRoot{
		// 填充相應字段
	}, nil
}
