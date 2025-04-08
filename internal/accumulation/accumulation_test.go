package accumulation

import (
	"encoding/json"
	"io"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/store"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	jamtests_accumulate "github.com/New-JAMneration/JAM-Protocol/jamtests/accumulate"
	jamtests_preimages "github.com/New-JAMneration/JAM-Protocol/jamtests/preimages"
	"github.com/google/go-cmp/cmp"
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
			// Create or get ServiceAccount, ensure its internal maps are initialized
			serviceAccount := types.ServiceAccount{
				ServiceInfo:    types.ServiceInfo{},
				PreimageLookup: make(types.PreimagesMapEntry),
				LookupDict:     make(types.LookupMetaMapEntry),
				StorageDict:    make(types.Storage),
			}

			// Fill PreimageLookup
			for _, preimage := range delta.Data.Preimages {
				serviceAccount.PreimageLookup[preimage.Hash] = preimage.Blob
			}

			// Fill LookupDict
			for _, lookup := range delta.Data.LookupMeta {
				serviceAccount.LookupDict[types.LookupMetaMapkey{
					Hash:   lookup.Key.Hash,
					Length: lookup.Key.Length,
				}] = lookup.Val
			}

			// Store ServiceAccount into inputDelta
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
				t.Logf("❌ [%s] %s", types.TEST_MODE, binFile)
				t.Fatalf("Should raise Error %v but got %v", preimages.Output.Err, accumulateErr)
			} else {
				t.Logf("Error: %v", accumulateErr)
				t.Logf("🔴 [%s] %s", types.TEST_MODE, binFile)
			}
		} else {
			if !reflect.DeepEqual(outputDelta, inputDelta) {
				t.Logf("❌ [%s] %s", types.TEST_MODE, binFile)
				t.Fatalf("Result States are not equal: %v", accumulateErr)
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
	dir := filepath.Join(JAM_TEST_VECTORS_DIR, "accumulate", types.TEST_MODE)

	// Read binary files
	binFiles, err := GetTargetExtensionFiles(dir, BIN_EXTENTION)
	if err != nil {
		t.Errorf("Error: %v", err)
	}

	for _, binFile := range binFiles {
		// if binFile != "accumulate_ready_queued_reports-1.bin" {
		// 	continue
		// }
		// Read the binary file
		binPath := filepath.Join(dir, binFile)
		t.Logf("▶ Processing [%s] %s", types.TEST_MODE, binFile)

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
	setupTestState(testCase.PreState, testCase.Input)

	// Execute accumulation

	// 12.1~2
	GetAccumulatedHashes()
	// Those two functions should be modified to get W from store
	UpdateImmediatelyAccumulateWorkReports(testCase.Input.Reports)
	UpdateQueuedWorkReports(testCase.Input.Reports)
	UpdateAccumulatableWorkReports()

	// 12.3
	err = DeferredTransfers()
	if err != nil {
		t.Errorf("DeferredTransfers raised error: %v", err)
	}

	// Validate final state
	validateFinalState(t, testCase.PostState)
	// The account does not change in testvector
	if !reflect.DeepEqual(testCase.PreState.Accounts, testCase.PostState.Accounts) {
		t.Errorf("Accounts do not match expected:\n%v,\nbut got \n%v", testCase.PreState.Accounts, testCase.PostState.Accounts)
	}
}

// Setup test state
func setupTestState(preState jamtests_accumulate.AccumulateState, input jamtests_accumulate.AccumulateInput) {
	s := store.GetInstance()

	// Set time slot
	s.GetPriorStates().SetTau(preState.Slot)
	s.GetIntermediateHeaderPointer().SetSlot(input.Slot)
	s.GetPosteriorStates().SetTau(input.Slot)

	// Set entropy
	s.GetPriorStates().SetEta(types.EntropyBuffer{preState.Entropy})
	s.GetPosteriorStates().SetEta0(preState.Entropy)

	// Set ready queue
	s.GetPriorStates().SetTheta(preState.ReadyQueue)

	// Set accumulated reports
	s.GetPriorStates().SetXi(preState.Accumulated)

	// Set privileges
	s.GetPriorStates().SetChi(preState.Privileges)

	// Set accounts
	inputDelta := make(types.ServiceAccountState)
	for serviceId, delta := range preState.Accounts {
		// Create or get ServiceAccount, ensure its internal maps are initialized
		serviceAccount := types.ServiceAccount{
			ServiceInfo:    delta.Data.Service,
			PreimageLookup: make(types.PreimagesMapEntry),
			LookupDict:     make(types.LookupMetaMapEntry),
			StorageDict:    make(types.Storage),
		}

		// Fill PreimageLookup
		for _, preimage := range delta.Data.Preimages {
			serviceAccount.PreimageLookup[preimage.Hash] = preimage.Blob
		}

		// Store ServiceAccount into inputDelta
		inputDelta[types.ServiceId(serviceId)] = serviceAccount
	}
	s.GetPriorStates().SetDelta(inputDelta)

	sort.Slice(input.Reports, func(i, j int) bool {
		return input.Reports[i].CoreIndex < input.Reports[j].CoreIndex
	})
	s.GetAccumulatableWorkReportsPointer().SetAccumulatableWorkReports(input.Reports)
}

// Validate final state
func validateFinalState(t *testing.T, expectedState jamtests_accumulate.AccumulateState) {
	s := store.GetInstance()

	// Validate time slot (passed)
	if s.GetPosteriorStates().GetTau() != expectedState.Slot {
		t.Errorf("Time slot does not match expected: %v, but got %v", expectedState.Slot, s.GetPosteriorStates().GetTau())
	}

	// Validate entropy (passed)
	if !reflect.DeepEqual(s.GetPosteriorStates().GetEta(), types.EntropyBuffer{expectedState.Entropy}) {
		t.Errorf("Entropy does not match expected: %v, but got %v", expectedState.Entropy, s.GetPosteriorStates().GetEta())
	}

	// Validate ready queue reports (passed expect nil and [])
	if !reflect.DeepEqual(s.GetPosteriorStates().GetTheta(), expectedState.ReadyQueue) {
		log.Printf("len of expected: %d, got: %d", len(expectedState.ReadyQueue), len(s.GetPosteriorStates().GetTheta()))
		diff := cmp.Diff(s.GetPosteriorStates().GetTheta(), expectedState.ReadyQueue)
		t.Errorf("Ready queue does not match expected:\n%v,\nbut got \n%v\nDiff:\n%v", expectedState.ReadyQueue, s.GetPosteriorStates().GetTheta(), diff)
	}

	// Validate accumulated reports (passed by implementing sort)
	if !reflect.DeepEqual(s.GetPosteriorStates().GetXi(), expectedState.Accumulated) {
		diff := cmp.Diff(s.GetPosteriorStates().GetXi(), expectedState.Accumulated)
		t.Errorf("Accumulated reports do not match expected:\n%v,but got \n%v\nDiff:\n%v", expectedState.Accumulated, s.GetPosteriorStates().GetXi(), diff)
	}

	// Validate privileges (passed)
	if !reflect.DeepEqual(s.GetPosteriorStates().GetChi(), expectedState.Privileges) {
		t.Errorf("Privileges do not match expected:\n%v,\nbut got %v", expectedState.Privileges, s.GetPosteriorStates().GetChi())
	}
}
