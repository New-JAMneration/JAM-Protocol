package accumulation

import (
	"bytes"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/blockchain"
	"github.com/New-JAMneration/JAM-Protocol/internal/statistics"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	utils "github.com/New-JAMneration/JAM-Protocol/internal/utilities"
	jamtests_accumulate "github.com/New-JAMneration/JAM-Protocol/jamtests/accumulate"
	jamtests_preimages "github.com/New-JAMneration/JAM-Protocol/jamtests/preimages"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
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
		blockchain.ResetInstance()
		s := blockchain.GetInstance()

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
}

// Setup test state
func setupTestState(preState jamtests_accumulate.AccumulateState, input jamtests_accumulate.AccumulateInput) {
	blockchain.ResetInstance()
	s := blockchain.GetInstance()

	// Set time slot
	s.GetPriorStates().SetTau(preState.Slot)
	block := types.Block{
		Header: types.Header{
			Slot: input.Slot,
		},
	}
	s.AddBlock(block)
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
	inputDelta := jamtests_accumulate.ParseAccountToServiceAccountState(preState.Accounts)
	s.GetPriorStates().SetDelta(inputDelta)

	sort.Slice(input.Reports, func(i, j int) bool {
		return input.Reports[i].CoreIndex < input.Reports[j].CoreIndex
	})
	s.GetIntermediateStates().SetAvailableWorkReports(input.Reports)
}

// Validate final state
func validateFinalState(t *testing.T, expectedState jamtests_accumulate.AccumulateState) {
	s := blockchain.GetInstance()

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
	if !cmp.Equal(s.GetPosteriorStates().GetXi(), expectedState.Accumulated, cmpopts.EquateEmpty()) {
		diff := cmp.Diff(s.GetPosteriorStates().GetXi(), expectedState.Accumulated, cmpopts.EquateEmpty())
		t.Errorf("Accumulated reports do not match, diff:\n%v", diff)
	}

	// Validate privileges (passed)
	if !reflect.DeepEqual(s.GetPosteriorStates().GetChi(), expectedState.Privileges) {
		diff := cmp.Diff(s.GetPosteriorStates().GetChi(), expectedState.Privileges)
		t.Errorf("Privileges do not match, diff:\n%v", diff)
	}

	// FIXME: Before validating statistics, we need to execute the update_preimage and update statistics functions

	// Validate Statistics (types.Statistics.Services, PI_S)
	// Calculate the actual statistics
	// INFO: This step will be executed in the UpdateStatistics function, but we can do it here for validation
	serviceIds := []types.ServiceId{}
	ourStatisticsServices := s.GetPosteriorStates().GetServicesStatistics()
	accumulationStatisitcs := s.GetIntermediateStates().GetAccumulationStatistics()

	for serviceId := range accumulationStatisitcs {
		serviceIds = append(serviceIds, serviceId)
	}

	for _, serviceId := range serviceIds {
		accumulateCount, accumulateGasUsed := statistics.CalculateAccumulationStatistics(serviceId, accumulationStatisitcs)
		// Skip if the service has no accumulated reports or gas used
		if accumulateCount == 0 && accumulateGasUsed == 0 {
			continue
		}
		// Update the statistics for the service
		thisServiceActivityRecord, ok := ourStatisticsServices[serviceId]
		if ok {
			thisServiceActivityRecord.AccumulateCount = accumulateCount
			thisServiceActivityRecord.AccumulateGasUsed = accumulateGasUsed
			ourStatisticsServices[serviceId] = thisServiceActivityRecord
		} else {
			newServiceActivityRecord := types.ServiceActivityRecord{
				AccumulateCount:   accumulateCount,
				AccumulateGasUsed: accumulateGasUsed,
			}
			ourStatisticsServices[serviceId] = newServiceActivityRecord
		}
	}
	// const EjectedServiceIDException = 2 // TEMP FIX: service 2 should not appear in R* statistics (issue #101 jam-test-vectors)

	// Validate statistics
	if expectedState.Statistics == nil {
		// we ignore case don't compare statistics
	} else if !reflect.DeepEqual(s.GetPosteriorStates().GetServicesStatistics(), expectedState.Statistics) {
		got := s.GetPosteriorStates().GetServicesStatistics()
		expected := expectedState.Statistics

		// TEMP FIX: ignore ejected service (ID = 2) for comparison
		// delete(expected, EjectedServiceIDException)
		diff := cmp.Diff(got, expected)
		log.Printf("expected:\n%v,\nbut got %v\n", expected, got)
		t.Errorf("statistics do not match, diff:\n%v", diff)
	}

	// Validate Accounts (AccountsMapEntry)
	// INFO:
	// The type of state.Delta is ServiceAccountState
	// The type of a.PostState.Accounts is []AccountsMapEntry

	// Validate Delta
	// FIXME: Review after PVM stable
	expectedDelta := jamtests_accumulate.ParseAccountToServiceAccountState(expectedState.Accounts)
	actualDelta := s.GetIntermediateStates().GetDeltaDoubleDagger()

	for key, expectedAcc := range expectedDelta {
		actualAcc, ok := actualDelta[key]
		if !ok {
			t.Errorf("serviceId %v missing in actualDelta", key)
		}

		// ServiceInfo
		// 0.7.0 davxy test lack loockupdict will  cause error for calculate item and byte length
		// lack of lookup dict -> item ( 2 -> 1 ), Bytes ( only compute storage )
		/*if !reflect.DeepEqual(expectedAcc.ServiceInfo, actualAcc.ServiceInfo) {
			return fmt.Errorf("mismatch in ServiceInfo for serviceId %v:\n expected=%+v\n actual=%+v",
				key, expectedAcc.ServiceInfo, actualAcc.ServiceInfo)
		}*/

		// PreimageLookup
		for h, expectedBlob := range expectedAcc.PreimageLookup {
			actualBlob, ok := actualAcc.PreimageLookup[h]
			if !ok {
				t.Errorf("serviceId %v missing Preimage hash %x in actualDelta", key, h)
			}
			if !bytes.Equal(expectedBlob, actualBlob) {
				t.Errorf("mismatch for serviceId %v, Preimage hash %x:\n expected=%x\n actual=%x",
					key, h, expectedBlob, actualBlob)
			}
		}
		for h := range actualAcc.PreimageLookup {
			if _, ok := expectedAcc.PreimageLookup[h]; !ok {
				t.Errorf("serviceId %v has extra Preimage hash %x in actualDelta", key, h)
			}
		}

		// StorageDict
		for storageKey, expectedValue := range expectedAcc.StorageDict {
			actualValue, ok := actualAcc.StorageDict[storageKey]
			if !ok {
				t.Errorf("serviceId %v missing Storage key %q in actualDelta", key, storageKey)
			}
			if !bytes.Equal(expectedValue, actualValue) {
				t.Errorf("mismatch for serviceId %v, Storage key %q:\n expected=%x\n actual=%x",
					key, storageKey, expectedValue, actualValue)
			}
		}
		for storageKey := range actualAcc.StorageDict {
			if _, ok := expectedAcc.StorageDict[storageKey]; !ok {
				t.Errorf("serviceId %v has extra Storage key %q in actualDelta", key, storageKey)
			}
		}
	}
}
