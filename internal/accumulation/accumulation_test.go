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
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
	m "github.com/New-JAMneration/JAM-Protocol/internal/utilities/merklization"
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
			inputDelta[delta.ID] = serviceAccount
		}

		// Get store instance and required states
		blockchain.ResetInstance()
		cs := blockchain.GetInstance()

		block := types.Block{
			Extrinsic: types.Extrinsic{
				Preimages: preimages.Input.Preimages,
			},
		}
		cs.AddBlock(block)
		cs.GetPriorStates().SetDelta(inputDelta)
		cs.GetIntermediateStates().SetDeltaDoubleDagger(inputDelta)
		cs.GetPosteriorStates().SetTau(preimages.Input.Slot)

		/*
			STF
		*/
		// Preimage
		_ = cs.GetPriorStateUnmatchedKeyVals() // legacy fallback pool no longer consulted; Step 7.5 will remove the accessor
		accumulateErr := ValidatePreimageExtrinsics(preimages.Input.Preimages, inputDelta)
		if accumulateErr == nil {
			accumulateErr = ProcessPreimageExtrinsics()
		}
		// Get output state
		outputDelta := cs.GetPosteriorStates().GetDelta()

		statistics.UpdateServiceActivityStatistics(cs.GetLatestBlock().Extrinsic)

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
			} else if !reflect.DeepEqual(cs.GetPosteriorStates().GetPi().Services, preimages.PostState.Statistics) {
				t.Logf("❌ [%s] %s", types.TEST_MODE, binFile)
				t.Fatalf("Service statistics do not match expected: %v, but got %v", preimages.PostState.Statistics, cs.GetPosteriorStates().GetPi().Services)
			} else {
				t.Logf("🟢 [%s] %s", types.TEST_MODE, binFile)
			}
		}
	}
}

// report4-like service IDs from fuzz state_diff on π (Statistics).
const (
	testServiceA    types.ServiceID      = 993378590  // 0x3b35c11e
	testServiceB    types.ServiceID      = 1097212405 // 0x416621f5
	testAuthorIndex types.ValidatorIndex = 5
)

func newEmptyServiceAccount() types.ServiceAccount {
	return types.ServiceAccount{
		PreimageLookup: make(types.PreimagesMapEntry),
		LookupDict:     make(types.LookupMetaMapEntry),
		StorageDict:    make(types.Storage),
	}
}

func makeTestPreimageBlob(size int, fill byte) types.ByteSequence {
	b := make(types.ByteSequence, size)
	for i := range b {
		b[i] = fill
	}
	return b
}

// setupPartialPreimageFilterChain configures report4-like E_P:
//   - two preimages (service A 40B, service B 25B), sorted by requester;
//   - A is already in PreimageLookup → filter rejects integration;
//   - B is eligible via unmatched δ₄ key-val.
func setupPartialPreimageFilterChain(t *testing.T) (
	cs *blockchain.ChainState,
	blobA, blobB types.ByteSequence,
) {
	t.Helper()

	blobA = makeTestPreimageBlob(40, 'a')
	blobB = makeTestPreimageBlob(25, 'b')
	hashA := hash.Blake2bHash(blobA)
	hashB := hash.Blake2bHash(blobB)

	accountA := newEmptyServiceAccount()
	accountA.PreimageLookup[hashA] = blobA

	accountB := newEmptyServiceAccount()

	delta := types.ServiceAccountState{
		testServiceA: accountA,
		testServiceB: accountB,
	}

	lookupKeyB := types.LookupMetaMapkey{Hash: hashB, Length: types.U32(len(blobB))}
	keyVals := types.StateKeyVals{
		m.EncodeDelta4KeyVal(testServiceB, lookupKeyB, types.TimeSlotSet{}),
	}

	blockchain.ResetInstance()
	cs = blockchain.GetInstance()
	cs.AddBlock(types.Block{
		Header: types.Header{
			AuthorIndex: testAuthorIndex,
		},
		Extrinsic: types.Extrinsic{
			Preimages: types.PreimagesExtrinsic{
				{Requester: testServiceA, Blob: blobA},
				{Requester: testServiceB, Blob: blobB},
			},
		},
	})
	cs.GetIntermediateStates().SetDeltaDoubleDagger(delta)
	cs.GetPosteriorStates().SetTau(100)
	cs.SetPostStateUnmatchedKeyVals(keyVals)

	pi := cs.GetPosteriorStates().GetPi()
	if len(pi.ValsCurr) <= int(testAuthorIndex) {
		pi.ValsCurr = make(types.ValidatorsStatistics, int(testAuthorIndex)+1)
		cs.GetPosteriorStates().SetPi(pi)
	}

	return cs, blobA, blobB
}

func assertPreimageEqual(t *testing.T, got types.Preimage, wantRequester types.ServiceID, wantBlob types.ByteSequence) {
	t.Helper()
	if got.Requester != wantRequester || !preimageBlobEqual(got.Blob, wantBlob) {
		t.Fatalf("preimage: got requester=%d len=%d, want requester=%d len=%d",
			got.Requester, len(got.Blob), wantRequester, len(wantBlob))
	}
}

func preimageBlobEqual(a, b types.ByteSequence) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestFilterPreimageExtrinsics_PartialIntegrate_PreservesBlockEP(t *testing.T) {
	cs, blobA, blobB := setupPartialPreimageFilterChain(t)

	eps := cs.GetLatestBlock().Extrinsic.Preimages
	delta := cs.GetIntermediateStates().GetDeltaDoubleDagger()

	filtered, _ := filterPreimageExtrinsics(eps, delta)
	blockEps := cs.GetLatestBlock().Extrinsic.Preimages

	if len(filtered) != 1 {
		t.Fatalf("filtered len: got %d want 1", len(filtered))
	}
	assertPreimageEqual(t, filtered[0], testServiceB, blobB)

	if len(blockEps) != 2 {
		t.Fatalf("block E_P len: got %d want 2", len(blockEps))
	}
	assertPreimageEqual(t, blockEps[0], testServiceA, blobA)
	assertPreimageEqual(t, blockEps[1], testServiceB, blobB)
}

func TestStatistics_AfterPartialFilter_UsesFullEP(t *testing.T) {
	cs, blobA, blobB := setupPartialPreimageFilterChain(t)

	eps := cs.GetLatestBlock().Extrinsic.Preimages
	delta := cs.GetIntermediateStates().GetDeltaDoubleDagger()
	_, _ = filterPreimageExtrinsics(eps, delta)

	extrinsic := cs.GetLatestBlock().Extrinsic
	statistics.UpdateCurrentStatistics(extrinsic)
	statistics.UpdateServiceActivityStatistics(extrinsic)

	pi := cs.GetPosteriorStates().GetPi()
	rec := pi.ValsCurr[testAuthorIndex]
	svcA := pi.Services[testServiceA]
	svcB := pi.Services[testServiceB]

	if rec.PreImages != 2 {
		t.Fatalf("π_V p: got %d want 2", rec.PreImages)
	}
	if rec.PreImagesSize != types.U32(len(blobA)+len(blobB)) {
		t.Fatalf("π_V d: got %d want %d", rec.PreImagesSize, len(blobA)+len(blobB))
	}
	if svcA.ProvidedCount != 1 || svcA.ProvidedSize != types.U32(len(blobA)) {
		t.Fatalf("π_S service A: got p=%d d=%d want 1/%d", svcA.ProvidedCount, svcA.ProvidedSize, len(blobA))
	}
	if svcB.ProvidedCount != 1 || svcB.ProvidedSize != types.U32(len(blobB)) {
		t.Fatalf("π_S service B: got p=%d d=%d want 1/%d", svcB.ProvidedCount, svcB.ProvidedSize, len(blobB))
	}
}

func TestProcessPreimageExtrinsics_PartialIntegrate(t *testing.T) {
	cs, blobA, blobB := setupPartialPreimageFilterChain(t)
	hashB := hash.Blake2bHash(blobB)

	if err := ProcessPreimageExtrinsics(); err != nil {
		t.Fatalf("ProcessPreimageExtrinsics: %v", err)
	}

	blockEps := cs.GetLatestBlock().Extrinsic.Preimages
	if len(blockEps) != 2 {
		t.Fatalf("block E_P len after process: got %d want 2", len(blockEps))
	}
	assertPreimageEqual(t, blockEps[0], testServiceA, blobA)
	assertPreimageEqual(t, blockEps[1], testServiceB, blobB)

	delta := cs.GetPosteriorStates().GetDelta()
	accB := delta[testServiceB]
	if _, ok := accB.PreimageLookup[hashB]; !ok {
		t.Fatal("δ should contain integrated preimage B")
	}
	accA := delta[testServiceA]
	if len(accA.LookupDict) != 0 {
		t.Fatalf("service A should have no new lookup from rejected E_P, got %d entries", len(accA.LookupDict))
	}

	extrinsic := cs.GetLatestBlock().Extrinsic
	statistics.UpdateCurrentStatistics(extrinsic)
	statistics.UpdateServiceActivityStatistics(extrinsic)

	pi := cs.GetPosteriorStates().GetPi()
	if pi.ValsCurr[testAuthorIndex].PreImages != 2 || pi.ValsCurr[testAuthorIndex].PreImagesSize != 65 {
		t.Fatalf("π_V: got p=%d d=%d want 2/65",
			pi.ValsCurr[testAuthorIndex].PreImages, pi.ValsCurr[testAuthorIndex].PreImagesSize)
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
	cs := blockchain.GetInstance()

	// Set time slot
	cs.GetPriorStates().SetTau(preState.Slot)
	block := types.Block{
		Header: types.Header{
			Slot: input.Slot,
		},
	}
	cs.AddBlock(block)
	cs.GetPosteriorStates().SetTau(input.Slot)

	// Set entropy
	cs.GetPriorStates().SetEta(types.EntropyBuffer{preState.Entropy})
	cs.GetPosteriorStates().SetEta0(preState.Entropy)

	// Set ready queue
	cs.GetPriorStates().SetVartheta(preState.ReadyQueue)

	// Set accumulated reports
	cs.GetPriorStates().SetXi(preState.Accumulated)

	// Set privileges
	cs.GetPriorStates().SetChi(preState.Privileges)

	// Set accounts
	inputDelta := jamtests_accumulate.ParseAccountToServiceAccountState(preState.Accounts)
	cs.GetPriorStates().SetDelta(inputDelta)

	sort.Slice(input.Reports, func(i, j int) bool {
		return input.Reports[i].CoreIndex < input.Reports[j].CoreIndex
	})
	cs.GetIntermediateStates().SetAvailableWorkReports(input.Reports)
}

// Validate final state
func validateFinalState(t *testing.T, expectedState jamtests_accumulate.AccumulateState) {
	cs := blockchain.GetInstance()

	// Validate time slot (passed)
	if cs.GetPosteriorStates().GetTau() != expectedState.Slot {
		t.Errorf("Time slot does not match expected: %v, but got %v", expectedState.Slot, cs.GetPosteriorStates().GetTau())
	}

	// Validate entropy (passed)
	if !reflect.DeepEqual(cs.GetPosteriorStates().GetEta(), types.EntropyBuffer{expectedState.Entropy}) {
		t.Errorf("Entropy does not match expected: %v, but got %v", expectedState.Entropy, cs.GetPosteriorStates().GetEta())
	}

	// Validate ready queue reports (passed expect nil and [])
	ourTheta := cs.GetPosteriorStates().GetVartheta()
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
	if !cmp.Equal(cs.GetPosteriorStates().GetXi(), expectedState.Accumulated, cmpopts.EquateEmpty()) {
		diff := cmp.Diff(cs.GetPosteriorStates().GetXi(), expectedState.Accumulated, cmpopts.EquateEmpty())
		t.Errorf("Accumulated reports do not match, diff:\n%v", diff)
	}

	// Validate privileges (passed)
	if !reflect.DeepEqual(cs.GetPosteriorStates().GetChi(), expectedState.Privileges) {
		diff := cmp.Diff(cs.GetPosteriorStates().GetChi(), expectedState.Privileges)
		t.Errorf("Privileges do not match, diff:\n%v", diff)
	}

	// FIXME: Before validating statistics, we need to execute the update_preimage and update statistics functions

	// Validate Statistics (types.Statistics.Services, PI_S)
	// Calculate the actual statistics
	// INFO: This step will be executed in the UpdateStatistics function, but we can do it here for validation
	serviceIDs := []types.ServiceID{}
	ourStatisticsServices := cs.GetPosteriorStates().GetServicesStatistics()
	accumulationStatisitcs := cs.GetIntermediateStates().GetAccumulationStatistics()

	for serviceID := range accumulationStatisitcs {
		serviceIDs = append(serviceIDs, serviceID)
	}

	for _, serviceID := range serviceIDs {
		accumulateCount, accumulateGasUsed := statistics.CalculateAccumulationStatistics(serviceID, accumulationStatisitcs)
		// Skip if the service has no accumulated reports or gas used
		if accumulateCount == 0 && accumulateGasUsed == 0 {
			continue
		}
		// Update the statistics for the service
		thisServiceActivityRecord, ok := ourStatisticsServices[serviceID]
		if ok {
			thisServiceActivityRecord.AccumulateCount = accumulateCount
			thisServiceActivityRecord.AccumulateGasUsed = accumulateGasUsed
			ourStatisticsServices[serviceID] = thisServiceActivityRecord
		} else {
			newServiceActivityRecord := types.ServiceActivityRecord{
				AccumulateCount:   accumulateCount,
				AccumulateGasUsed: accumulateGasUsed,
			}
			ourStatisticsServices[serviceID] = newServiceActivityRecord
		}
	}
	// const EjectedServiceIDException = 2 // TEMP FIX: service 2 should not appear in R* statistics (issue #101 jam-test-vectors)

	// Validate statistics
	if expectedState.Statistics == nil {
		// we ignore case don't compare statistics
	} else if !reflect.DeepEqual(cs.GetPosteriorStates().GetServicesStatistics(), expectedState.Statistics) {
		got := cs.GetPosteriorStates().GetServicesStatistics()
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
	actualDelta := cs.GetIntermediateStates().GetDeltaDoubleDagger()

	for key, expectedAcc := range expectedDelta {
		actualAcc, ok := actualDelta[key]
		if !ok {
			t.Errorf("serviceID %v missing in actualDelta", key)
		}

		// ServiceInfo
		// 0.7.0 davxy test lack loockupdict will  cause error for calculate item and byte length
		// lack of lookup dict -> item ( 2 -> 1 ), Bytes ( only compute storage )
		/*if !reflect.DeepEqual(expectedAcc.ServiceInfo, actualAcc.ServiceInfo) {
			return fmt.Errorf("mismatch in ServiceInfo for serviceID %v:\n expected=%+v\n actual=%+v",
				key, expectedAcc.ServiceInfo, actualAcc.ServiceInfo)
		}*/

		// PreimageLookup
		for h, expectedBlob := range expectedAcc.PreimageLookup {
			actualBlob, ok := actualAcc.PreimageLookup[h]
			if !ok {
				t.Errorf("serviceID %v missing Preimage hash %x in actualDelta", key, h)
			}
			if !bytes.Equal(expectedBlob, actualBlob) {
				t.Errorf("mismatch for serviceID %v, Preimage hash %x:\n expected=%x\n actual=%x",
					key, h, expectedBlob, actualBlob)
			}
		}
		for h := range actualAcc.PreimageLookup {
			if _, ok := expectedAcc.PreimageLookup[h]; !ok {
				t.Errorf("serviceID %v has extra Preimage hash %x in actualDelta", key, h)
			}
		}

		// StorageDict
		for storageKey, expectedValue := range expectedAcc.StorageDict {
			actualValue, ok := actualAcc.StorageDict[storageKey]
			if !ok {
				t.Errorf("serviceID %v missing Storage key %q in actualDelta", key, storageKey)
			}
			if !bytes.Equal(expectedValue, actualValue) {
				t.Errorf("mismatch for serviceID %v, Storage key %q:\n expected=%x\n actual=%x",
					key, storageKey, expectedValue, actualValue)
			}
		}
		for storageKey := range actualAcc.StorageDict {
			if _, ok := expectedAcc.StorageDict[storageKey]; !ok {
				t.Errorf("serviceID %v has extra Storage key %q in actualDelta", key, storageKey)
			}
		}
	}
}
