package safrole_test

import (
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/New-JAMneration/JAM-Protocol/internal/blockchain"
	"github.com/New-JAMneration/JAM-Protocol/internal/safrole"
	"github.com/New-JAMneration/JAM-Protocol/internal/stf"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	utils "github.com/New-JAMneration/JAM-Protocol/internal/utilities"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/merklization"
	jamtests_safrole "github.com/New-JAMneration/JAM-Protocol/jamtests/safrole"
	jamtests_trace "github.com/New-JAMneration/JAM-Protocol/jamtests/trace"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestGetEpochIndex(t *testing.T) {
	// Mock types.EpochLength
	backupEpochLength := types.EpochLength
	types.EpochLength = 10

	// Test various time slot inputs
	tests := []struct {
		input    types.TimeSlot
		expected types.TimeSlot
	}{
		{0, 0},  // Time slot 0, epoch 0
		{9, 0},  // Time slot 9, epoch 0
		{10, 1}, // Time slot 10, epoch 1
		{20, 2}, // Time slot 20, epoch 2
		{25, 2}, // Time slot 25, epoch 2
	}

	for _, test := range tests {
		result := safrole.GetEpochIndex(test.input)
		if result != test.expected {
			t.Errorf("For input %v, expected epoch %v but got %v", test.input, test.expected, result)
		}
	}

	types.EpochLength = backupEpochLength
}

func TestGetSlotIndex(t *testing.T) {
	// Mock types.EpochLength
	backupEpochLength := types.EpochLength
	types.EpochLength = 10

	// Test various time slot inputs
	tests := []struct {
		input    types.TimeSlot
		expected types.TimeSlot
	}{
		{0, 0},  // time slot 0, slot index 0
		{9, 9},  // time slot 9, slot index 9
		{10, 0}, // time slot 10, slot index 0
		{20, 0}, // time slot 20, slot index 0
		{25, 5}, // time slot 25, slot index 5
	}

	for _, test := range tests {
		result := safrole.GetSlotIndex(test.input)
		if result != test.expected {
			t.Errorf("For input %v, expected slotIndex %v but got %v", test.input, test.expected, result)
		}
	}

	types.EpochLength = backupEpochLength
}

func TestValidatorIsOffender(t *testing.T) {
	offendersMark := types.OffendersMark{}
	offenderValidator := types.Validator{
		Bandersnatch: types.BandersnatchPublic{},
		Ed25519:      types.Ed25519Public{1, 2, 3},
		Bls:          types.BlsPublic{},
		Metadata:     types.ValidatorMetadata{},
	}
	offendersMark = append(offendersMark, offenderValidator.Ed25519)

	testCases := []struct {
		validator  types.Validator
		offenders  types.OffendersMark
		isOffender bool
	}{
		{
			types.Validator{
				Bandersnatch: types.BandersnatchPublic{},
				Ed25519:      types.Ed25519Public{1, 2, 3},
				Bls:          types.BlsPublic{},
				Metadata:     types.ValidatorMetadata{},
			},
			offendersMark,
			true,
		},
		{
			types.Validator{
				Bandersnatch: types.BandersnatchPublic{},
				Ed25519:      types.Ed25519Public{1, 2, 2},
				Bls:          types.BlsPublic{},
				Metadata:     types.ValidatorMetadata{},
			},
			offendersMark,
			false,
		},
		{
			types.Validator{
				Bandersnatch: types.BandersnatchPublic{},
				Ed25519:      types.Ed25519Public{2, 2, 2},
				Bls:          types.BlsPublic{},
				Metadata:     types.ValidatorMetadata{},
			},
			offendersMark,
			false,
		},
	}

	for _, testCase := range testCases {
		if actual := safrole.ValidatorIsOffender(testCase.validator, testCase.offenders); actual != testCase.isOffender {
			t.Errorf("ValidatorIsOffender(%v, %v) = %t, expected %t", testCase.validator, testCase.offenders, actual, testCase.isOffender)
		}
	}
}

func TestKeyRotate(t *testing.T) {
	s := blockchain.GetInstance()
	priorState := s.GetPriorStates()

	now := time.Now().UTC()
	timeInSecond := uint64(now.Sub(types.JamCommonEra).Seconds())
	tauPrime := types.TimeSlot(timeInSecond / uint64(types.SlotPeriod))
	e := safrole.GetEpochIndex(tauPrime)
	ePrime := safrole.GetEpochIndex(tauPrime)

	// Add a block to the store
	s.AddBlock(types.Block{
		Header: types.Header{
			Slot: tauPrime - types.TimeSlot(types.EpochLength),
		},
	})

	// Set offendersMark
	s.GetPosteriorStates().SetPsiO(types.OffendersMark{})

	// Simulate previous time slot to trigger key rotation
	priorState.SetTau(tauPrime - types.TimeSlot(types.EpochLength))

	fakeValidators := safrole.LoadFakeValidators()

	priorKappa := types.ValidatorsData{}
	for _, fakeValidator := range fakeValidators {
		priorKappa = append(priorKappa, types.Validator{
			Bandersnatch: fakeValidator.Bandersnatch,
			Ed25519:      fakeValidator.Ed25519,
			Bls:          fakeValidator.BLS,
			Metadata:     types.ValidatorMetadata{},
		})
	}

	priorGammaK := types.ValidatorsData{}
	for _, fakeValidator := range fakeValidators {
		priorGammaK = append(priorGammaK, types.Validator{
			Bandersnatch: fakeValidator.Bandersnatch,
			Ed25519:      fakeValidator.Ed25519,
			Bls:          fakeValidator.BLS,
			Metadata:     types.ValidatorMetadata{},
		})
	}

	priorIota := types.ValidatorsData{}
	for _, fakeValidator := range fakeValidators {
		priorIota = append(priorIota, types.Validator{
			Bandersnatch: fakeValidator.Bandersnatch,
			Ed25519:      fakeValidator.Ed25519,
			Bls:          fakeValidator.BLS,
			Metadata:     types.ValidatorMetadata{},
		})
	}

	priorLambda := types.ValidatorsData{}
	for _, fakeValidator := range fakeValidators {
		priorLambda = append(priorLambda, types.Validator{
			Bandersnatch: fakeValidator.Bandersnatch,
			Ed25519:      fakeValidator.Ed25519,
			Bls:          fakeValidator.BLS,
			Metadata:     types.ValidatorMetadata{},
		})
	}

	gammaZ := "0xa949a60ad754d683d398a0fb674a9bbe525ca26b0b0b9c8d79f210291b40d286d9886a9747a4587d497f2700baee229ca72c54ad652e03e74f35f075d0189a40d41e5ee65703beb5d7ae8394da07aecf9056b98c61156714fd1d9982367bee2992e630ae2b14e758ab0960e372172203f4c9a41777dadd529971d7ab9d23ab29fe0e9c85ec450505dde7f5ac038274cf"
	priorGammaZ := types.BandersnatchRingCommitment(safrole.Hex2Bytes(gammaZ))

	priorState.SetKappa(priorKappa)
	priorState.SetLambda(priorLambda)
	priorState.SetIota(priorIota)
	priorState.SetGammaK(priorGammaK)
	priorState.SetGammaZ(priorGammaZ)

	s.GenerateGenesisState(priorState.GetState())

	safrole.KeyRotate(e, ePrime)

	// Get posterior state
	posteriorState := s.GetPosteriorStates()
	if !reflect.DeepEqual(posteriorState.GetGammaK(), priorIota) {
		t.Errorf("Expected GammaK to be %v, got %v", priorIota, posteriorState.GetGammaK())
	}
	if !reflect.DeepEqual(posteriorState.GetKappa(), priorGammaK) {
		t.Errorf("Expected Kappa to be %v, got %v", priorGammaK, posteriorState.GetKappa())
	}
	if !reflect.DeepEqual(posteriorState.GetLambda(), priorKappa) {
		t.Errorf("Expected Lambda to be %v, got %v", priorKappa, posteriorState.GetLambda())
	}
	if posteriorState.GetGammaZ() != priorGammaZ {
		t.Errorf("Expected GammaZ to be %v, got %v", priorGammaZ, posteriorState.GetGammaZ())
	}
}

func TestReplaceOffenderKeysEmptyOffenders(t *testing.T) {
	// Load fake validators
	fakeValidators := safrole.LoadFakeValidators()

	// Create validators data
	validatorsData := types.ValidatorsData{}
	for _, fakeValidator := range fakeValidators {
		validatorsData = append(validatorsData, types.Validator{
			Bandersnatch: fakeValidator.Bandersnatch,
			Ed25519:      fakeValidator.Ed25519,
			Bls:          fakeValidator.BLS,
			Metadata:     types.ValidatorMetadata{},
		})
	}

	// Set posterior state offenders to empty
	s := blockchain.GetInstance()
	s.GetPosteriorStates().SetPsiO(types.OffendersMark{})

	newValidators := safrole.ReplaceOffenderKeys(validatorsData)

	// Check if the new validators data has the same length as the original
	// validators data
	if len(newValidators) != len(validatorsData) {
		t.Errorf("Expected newValidators to have %d elements, got %d", len(validatorsData), len(newValidators))
	}
}

func TestReplaceOffenderKeys(t *testing.T) {
	// Load fake validators
	fakeValidators := safrole.LoadFakeValidators()

	// Create validators data
	validatorsData := types.ValidatorsData{}
	for _, fakeValidator := range fakeValidators {
		validatorsData = append(validatorsData, types.Validator{
			Bandersnatch: fakeValidator.Bandersnatch,
			Ed25519:      fakeValidator.Ed25519,
			Bls:          fakeValidator.BLS,
			Metadata:     types.ValidatorMetadata{},
		})
	}

	// Set posterior state offenders to the first validator
	s := blockchain.GetInstance()
	s.GetPosteriorStates().SetPsiO(types.OffendersMark{fakeValidators[0].Ed25519})

	newValidators := safrole.ReplaceOffenderKeys(validatorsData)

	// Check if the new validators data has the same length as the original
	// validators data
	if len(newValidators) != len(validatorsData) {
		t.Errorf("Expected newValidators to have %d elements, got %d", len(validatorsData), len(newValidators))
	}

	// Check if the new validators data has the same elements as the original
	// validators data, except for the offender
	for i, newValidator := range newValidators {
		if newValidator.Ed25519 == fakeValidators[0].Ed25519 {
			t.Errorf("Expected newValidators[%d] to be different from the offender, got %v", i, newValidator)
		}
	}

	if newValidators[0].Ed25519 != (types.Ed25519Public{}) {
		t.Errorf("Expected newValidators[0] to be zeroed out, got %v", newValidators[0])
	}
}

// TODO: Add tests for GetNewSafroleState, UpdateBandersnatchKeyRoot

func TestMain(m *testing.M) {
	// Set the test mode
	types.SetTestMode()

	// Run the tests
	os.Exit(m.Run())
}

func TestSafroleTestVectors(t *testing.T) {
	dir := filepath.Join(utils.JAM_TEST_VECTORS_DIR, "stf", "safrole", types.TEST_MODE)

	// Read binary files
	binFiles, err := utils.GetTargetExtensionFiles(dir, utils.BIN_EXTENTION)
	if err != nil {
		t.Errorf("Error: %v", err)
	}

	for _, binFile := range binFiles {
		// if binFile != "enact-epoch-change-with-no-tickets-2.bin" {
		// 	continue
		// }
		// Read the binary file
		binPath := filepath.Join(dir, binFile)
		t.Logf("▶ Processing [%s] %s", types.TEST_MODE, binFile)

		// Test accumulate
		testSafroleFile(t, binPath, binFile)
	}
}

func testSafroleFile(t *testing.T, binPath string, binFile string) {
	// Load accumulate test case
	testCase := &jamtests_safrole.SafroleTestCase{}
	err := utils.GetTestFromBin(binPath, testCase)
	if err != nil {
		t.Errorf("Error: %v", err)
	}

	setupTestState(testCase.PreState, testCase.Input)

	// Measure time for OuterUsedSafrole
	start := time.Now()
	errCode := safrole.OuterUsedSafrole()
	duration := time.Since(start)
	t.Logf("OuterUsedSafrole took: %v", duration)

	validateFinalState(t, binFile, testCase, errCode)
}

// Setup test state
func setupTestState(preState jamtests_safrole.SafroleState, input jamtests_safrole.SafroleInput) {
	blockchain.ResetInstance()
	storeInstance := blockchain.GetInstance()

	storeInstance.GetPriorStates().SetTau(preState.Tau)
	storeInstance.GetProcessingBlockPointer().SetSlot(input.Slot)
	storeInstance.GetPosteriorStates().SetTau(input.Slot)

	storeInstance.GetPriorStates().SetEta(preState.Eta)
	// Set eta^prime_0 here
	hash_input := append(preState.Eta[0][:], input.Entropy[:]...)
	storeInstance.GetPosteriorStates().SetEta0(types.Entropy(hash.Blake2bHash(hash_input)))

	storeInstance.GetPriorStates().SetLambda(preState.Lambda)
	storeInstance.GetPriorStates().SetKappa(preState.Kappa)
	storeInstance.GetPriorStates().SetGammaK(preState.GammaK)
	storeInstance.GetPriorStates().SetIota(preState.Iota)
	storeInstance.GetPriorStates().SetGammaA(preState.GammaA)
	storeInstance.GetPriorStates().SetGammaS(preState.GammaS)
	storeInstance.GetPriorStates().SetGammaZ(preState.GammaZ)

	storeInstance.GetPosteriorStates().SetPsiO(preState.PostOffenders)

	// Add block with TicketsExtrinsic
	block := types.Block{
		Extrinsic: types.Extrinsic{
			Tickets: input.Extrinsic,
		},
	}
	storeInstance.AddBlock(block)
}

// Validate final state
func validateFinalState(t *testing.T, binFile string,
	testCase *jamtests_safrole.SafroleTestCase,
	errCode *types.ErrorCode,
) {
	s := blockchain.GetInstance()

	ourEpochMarker := s.GetProcessingBlockPointer().GetEpochMark()
	ourTicketsMark := s.GetProcessingBlockPointer().GetTicketsMark()
	expectedErr := testCase.Output.Err
	expectedOk := testCase.Output.Ok

	// Handle safrole output error
	safroleOutputErrCode := -1 // -1 indicates invalid error code

	if expectedErr == nil && expectedOk != nil { // OK case
		log.Printf("input safroleOkInfo: %v", *expectedOk)

		if !reflect.DeepEqual(expectedOk.EpochMark, ourEpochMarker) {
			if expectedOk.EpochMark != nil && ourEpochMarker != nil {
				diff := cmp.Diff(*expectedOk.EpochMark, *ourEpochMarker)
				t.Logf("❌ [%s] %s", types.TEST_MODE, binFile)
				t.Fatalf("epoch marker mismatch: %v", diff)
			} else {
				t.Logf("❌ [%s] %s", types.TEST_MODE, binFile)
				t.Fatalf("epoch marker mismatch: expected %v\n, got %v\n", expectedOk.EpochMark, ourEpochMarker)
			}
		}

		if !reflect.DeepEqual(expectedOk.TicketsMark, ourTicketsMark) {
			if expectedOk.TicketsMark != nil && ourTicketsMark != nil {
				diff := cmp.Diff(*expectedOk.TicketsMark, *ourTicketsMark)
				t.Logf("❌ [%s] %s", types.TEST_MODE, binFile)
				t.Fatalf("tickets mark mismatch: %v", diff)
			} else {
				t.Logf("❌ [%s] %s", types.TEST_MODE, binFile)
				t.Fatalf("tickets mark mismatch: expected %v\n, got %v\n", expectedOk.TicketsMark, ourTicketsMark)
			}
		}

		/*
			Validate iota doesn't change
		*/
		// if !reflect.DeepEqual(testCase.PostState.Iota, testCase.PreState.Iota) {
		// 	diff := cmp.Diff(testCase.PreState.Iota, testCase.PostState.Iota)
		// 	t.Logf("[%s] %s: iota should change: %v", types.TEST_MODE, binFile, diff)
		// } else {
		// 	t.Logf("[%s] %s: iota should not change", types.TEST_MODE, binFile)
		// }
		validateState(t, testCase.PostState)

		t.Logf("🟢 [%s] %s", types.TEST_MODE, binFile)
	} else { // Error case
		safroleOutputErrCode = int(*expectedErr)
		log.Printf("input ErrCode: %v", safroleOutputErrCode)

		if errCode != nil && safroleOutputErrCode == int(*errCode) {
			if !reflect.DeepEqual(testCase.PreState, testCase.PostState) {
				diff := cmp.Diff(testCase.PreState, testCase.PostState)
				t.Errorf("error case state mismatch: %v", diff)
			}
			t.Logf("🔴 [%s] %s", types.TEST_MODE, binFile)
		} else {
			t.Errorf("error code mismatch: expected %v, got %v", safroleOutputErrCode, *errCode)
			t.Logf("❌ [%s] %s", types.TEST_MODE, binFile)
		}
	}
}

func validateState(t *testing.T, expectedState jamtests_safrole.SafroleState) {
	storeInstance := blockchain.GetInstance()
	posteriorState := storeInstance.GetPosteriorStates()

	checks := []struct {
		name        string
		expected    interface{}
		actual      interface{}
		compareOpts []cmp.Option
	}{
		{"tau", expectedState.Tau, posteriorState.GetTau(), nil},
		{"eta", expectedState.Eta, posteriorState.GetEta(), nil},
		{"lambda", expectedState.Lambda, posteriorState.GetLambda(), nil},
		{"kappa", expectedState.Kappa, posteriorState.GetKappa(), nil},
		{"gamma_k", expectedState.GammaK, posteriorState.GetGammaK(), nil},
		{"gamma_a", expectedState.GammaA, posteriorState.GetGammaA(), []cmp.Option{cmpopts.EquateEmpty()}},
		{"gamma_s", expectedState.GammaS, posteriorState.GetGammaS(), nil},
		{"gamma_z", expectedState.GammaZ, posteriorState.GetGammaZ(), nil},
		{"post_offenders", expectedState.PostOffenders, posteriorState.GetPsiO(), nil},
	}

	for _, check := range checks {
		var equal bool
		if check.compareOpts != nil {
			equal = cmp.Equal(check.expected, check.actual, check.compareOpts...)
		} else {
			equal = reflect.DeepEqual(check.expected, check.actual)
		}

		if !equal {
			var diff string
			if check.compareOpts != nil {
				diff = cmp.Diff(check.expected, check.actual, check.compareOpts...)
			} else {
				diff = cmp.Diff(check.expected, check.actual)
			}
			t.Errorf("%s mismatch:\n%v", check.name, diff)
			return
		}
	}
}

func TestJamtestvectorsTraces(t *testing.T) {
	dirsPath := filepath.Join(utils.JAM_TEST_VECTORS_DIR, "traces")

	// Set test mode
	types.SetTestMode()

	// Get all dirs in dirsPath
	dirs, err := os.ReadDir(dirsPath)
	if err != nil {
		t.Errorf("Error reading dirs in %s: %v", dirsPath, err)
		return
	}

	totalPassed := 0
	totalFailed := 0

	for _, dirEntry := range dirs {
		log.Printf("dirEntry: %s", dirEntry.Name())
		if !dirEntry.IsDir() || dirEntry.Name() != "safrole" { // Modify to run other traces
			continue
		}

		dirPath := filepath.Join(dirsPath, dirEntry.Name())
		t.Logf("Processing directory: %s", dirEntry.Name())

		// Get all bin files in directory
		fileNames, err := utils.GetTargetExtensionFiles(dirPath, utils.BIN_EXTENTION)
		if err != nil {
			t.Errorf("Error getting files from directory %s: %v", dirPath, err)
			continue
		}

		// Find and setup genesis
		genesisFileFound := false
		genesisFilePath := ""

		for _, fileName := range fileNames {
			if strings.Contains(fileName, "genesis") {
				genesisFilePath = filepath.Join(dirPath, fileName)
				genesisFileFound = true
				break
			}
		}

		if !genesisFileFound {
			t.Logf("Warning: genesis not found in %s, skipping directory", dirPath)
			continue
		}

		// Setup genesis state
		blockchain.ResetInstance()
		instance := blockchain.GetInstance()

		var state types.State
		var block types.Block

		genesisTestCase := &jamtests_trace.Genesis{}
		err = utils.GetTestFromBin(genesisFilePath, genesisTestCase)
		if err != nil {
			t.Errorf("Failed to read genesis: %v", err)
			continue
		}

		state, _, err = merklization.StateKeyValsToState(genesisTestCase.State.KeyVals)
		if err != nil {
			t.Errorf("Failed to parse state key-vals to state: %v", err)
			continue
		}

		block.Header = genesisTestCase.Header
		instance.GenerateGenesisBlock(block)
		instance.GenerateGenesisState(state)

		// Run trace tests
		dirPassed := 0
		dirFailed := 0

		for idx, fileName := range fileNames {
			// Skip genesis file
			if strings.Contains(fileName, "genesis") {
				continue
			}

			filePath := filepath.Join(dirPath, fileName)
			testStart := time.Now()
			t.Logf("------------------{%v, %s}--------------------", idx, fileName)

			// State commit before each test (post-state update to pre-state, tau_prime+1)
			blockchain.GetInstance().StateCommit()

			// Read trace test case
			traceTestCase := &jamtests_trace.TraceTestCase{}
			err := utils.GetTestFromBin(filePath, traceTestCase)
			if err != nil {
				t.Logf("got error during parsing: %v", err)
				dirFailed++
				continue
			}

			// Setup block from trace test case
			instance.AddBlock(traceTestCase.Block)

			// Run STF with time measurement
			stfStart := time.Now()
			_, outputErr := stf.RunSTF()
			stfDuration := time.Since(stfStart)

			testDuration := time.Since(testStart)

			if outputErr != nil {
				t.Logf("❌ STF output error: %v (STF: %v, Total: %v)", outputErr, stfDuration, testDuration)
				dirFailed++
			} else {
				// Validate state root
				posteriorState := instance.GetPosteriorStates()
				postKeyVals, err := merklization.StateEncoder(posteriorState.GetState())
				if err != nil {
					t.Errorf("Failed to encode state: %v", err)
					continue
				}
				actualStateRoot := merklization.MerklizationSerializedState(postKeyVals)
				expectedStateRoot := traceTestCase.PostState.StateRoot

				if actualStateRoot != expectedStateRoot {
					t.Logf("❌ State root mismatch: expected 0x%x, got 0x%x (STF: %v, Total: %v)",
						expectedStateRoot, actualStateRoot, stfDuration, testDuration)
					dirFailed++
				} else {
					dirPassed++
					t.Logf("✅ passed (STF: %v, Total: %v)", stfDuration, testDuration)
				}
			}
		}

		totalPassed += dirPassed
		totalFailed += dirFailed
		t.Logf("Directory %s: Passed: %d, Failed: %d", dirEntry.Name(), dirPassed, dirFailed)
	}

	t.Logf("========================================")
	t.Logf("Total: Passed: %d, Failed: %d", totalPassed, totalFailed)
}
