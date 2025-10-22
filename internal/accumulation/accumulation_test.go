package accumulation

import (
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/statistics"
	"github.com/New-JAMneration/JAM-Protocol/internal/store"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	utils "github.com/New-JAMneration/JAM-Protocol/internal/utilities"
	jamtests_accumulate "github.com/New-JAMneration/JAM-Protocol/jamtests/accumulate"
	jamtests_preimages "github.com/New-JAMneration/JAM-Protocol/jamtests/preimages"
	"github.com/google/go-cmp/cmp"
)

func TestMain(m *testing.M) {
	// Set the test mode
	types.SetTestMode()

	// Run the tests
	os.Exit(m.Run())
}

func TestPreimageTestVectors(t *testing.T) {
	dir := filepath.Join(utils.JAM_TEST_VECTORS_DIR, "stf", "preimages", types.TEST_MODE)

	// Read binary files
	binFiles, err := utils.GetTargetExtensionFiles(dir, utils.BIN_EXTENTION)
	if err != nil {
		t.Errorf("Error: %v", err)
	}

	for _, binFile := range binFiles {
		// Read the binary file
		binPath := filepath.Join(dir, binFile)

		// Load preimages test case
		preimages := &jamtests_preimages.PreimageTestCase{}

		err := utils.GetTestFromBin(binPath, preimages)
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

		// Get store instance and required states
		store.ResetInstance()
		s := store.GetInstance()

		block := types.Block{
			Header: types.Header{
				Slot: preimages.Input.Slot,
			},
			Extrinsic: types.Extrinsic{
				Preimages: preimages.Input.Preimages,
			},
		}
		s.AddBlock(block)

		s.GetIntermediateStates().SetDeltaDoubleDagger(inputDelta)
		s.GetPosteriorStates().SetTau(preimages.Input.Slot)

		/*
			STF
		*/
		accumulateErr := ProcessPreimageExtrinsics()

		// Get output state
		outputDelta := s.GetPosteriorStates().GetDelta()

		statistics.UpdateServiceActivityStatistics(s.GetLatestBlock().Extrinsic)

		// Validate output state
		if preimages.Output.Err != nil {
			if accumulateErr == nil || accumulateErr.Error() != preimages.Output.Err.Error() {
				t.Logf("❌ [%s] %s", types.TEST_MODE, binFile)
				t.Fatalf("Should raise Error %v but got %v", preimages.Output.Err, accumulateErr)
			} else {
				t.Logf("ErrorCode matched: expected %v, got %v", preimages.Output.Err, accumulateErr)
				t.Logf("🔴 [%s] %s", types.TEST_MODE, binFile)
			}
		} else {
			if accumulateErr != nil {
				t.Logf("❌ [%s] %s", types.TEST_MODE, binFile)
				t.Fatalf("No Error expected but got %v", accumulateErr)
			} else if !reflect.DeepEqual(outputDelta, inputDelta) {
				t.Logf("❌ [%s] %s", types.TEST_MODE, binFile)
				diff := cmp.Diff(inputDelta, outputDelta)
				t.Fatalf("Result States are not equal: %v", diff)
			} else if !reflect.DeepEqual(s.GetPosteriorStates().GetPi().Services, preimages.PostState.Statistics) {
				t.Logf("❌ [%s] %s", types.TEST_MODE, binFile)
				t.Fatalf("Service statistics do not match expected: %v, but got %v", preimages.PostState.Statistics, s.GetPosteriorStates().GetPi().Services)
			} else {
				t.Logf("🟢 [%s] %s", types.TEST_MODE, binFile)
			}
		}
	}
}

func TestAccumulateTestVectors(t *testing.T) {
	dir := filepath.Join(utils.JAM_TEST_VECTORS_DIR, "stf", "accumulate", types.TEST_MODE)

	// Read binary files
	binFiles, err := utils.GetTargetExtensionFiles(dir, utils.BIN_EXTENTION)
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
	// Load accumulate test case
	testCase := &jamtests_accumulate.AccumulateTestCase{}
	err := utils.GetTestFromBin(binPath, testCase)
	if err != nil {
		t.Errorf("Error: %v", err)
	}

	// Setup test state
	setupTestState(testCase.PreState, testCase.Input)

	// Execute accumulation
	// 12.1, 12.2
	err = ProcessAccumulation()
	if err != nil {
		t.Errorf("ProcessAccumulation raised error: %v", err)
	}

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
	s.GetProcessingBlockPointer().SetSlot(input.Slot)
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
	s.GetIntermediateStates().SetAvailableWorkReports(input.Reports)
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
	ourTheta := s.GetPosteriorStates().GetTheta()
	if !reflect.DeepEqual(ourTheta, expectedState.ReadyQueue) {
		// log.Printf("len of queue reports expected: %d, got: %d", len(expectedState.ReadyQueue), len(s.GetPosteriorStates().GetTheta()))
		for i := range ourTheta {
			if expectedState.ReadyQueue[i] == nil {
				expectedState.ReadyQueue[i] = []types.ReadyRecord{}
			}
			if ourTheta[i] == nil {
				ourTheta[i] = []types.ReadyRecord{}
			}
			diff := cmp.Diff(ourTheta[i], expectedState.ReadyQueue[i])
			if len(diff) != 0 {
				t.Errorf("Theta[%d] Diff:\n%v", i, diff)
			}
		}

		// t.Errorf("Ready queue does not match expected:\n%v,\nbut got \n%v\nDiff:\n%v", expectedState.ReadyQueue, ourTheta, diff)
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
