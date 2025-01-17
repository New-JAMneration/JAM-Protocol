package safrole

import (
	"reflect"
	"testing"
	"time"

	"github.com/New-JAMneration/JAM-Protocol/internal/store"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
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
		result := GetEpochIndex(test.input)
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
		result := GetSlotIndex(test.input)
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
		if actual := ValidatorIsOffender(testCase.validator, testCase.offenders); actual != testCase.isOffender {
			t.Errorf("ValidatorIsOffender(%v, %v) = %t, expected %t", testCase.validator, testCase.offenders, actual, testCase.isOffender)
		}
	}
}

func TestGetBandersnatchRingRootCommmitment(t *testing.T) {
	expectedCommitmentStr := "0xa949a60ad754d683d398a0fb674a9bbe525ca26b0b0b9c8d79f210291b40d286d9886a9747a4587d497f2700baee229ca72c54ad652e03e74f35f075d0189a40d41e5ee65703beb5d7ae8394da07aecf9056b98c61156714fd1d9982367bee2992e630ae2b14e758ab0960e372172203f4c9a41777dadd529971d7ab9d23ab29fe0e9c85ec450505dde7f5ac038274cf"
	expectedCommitment := types.BandersnatchRingCommitment(hex2Bytes(expectedCommitmentStr))

	bandersnatchKeys := []types.BandersnatchPublic{
		types.BandersnatchPublic(hex2Bytes("0x5e465beb01dbafe160ce8216047f2155dd0569f058afd52dcea601025a8d161d")),
		types.BandersnatchPublic(hex2Bytes("0x3d5e5a51aab2b048f8686ecd79712a80e3265a114cc73f14bdb2a59233fb66d0")),
		types.BandersnatchPublic(hex2Bytes("0xaa2b95f7572875b0d0f186552ae745ba8222fc0b5bd456554bfe51c68938f8bc")),
		types.BandersnatchPublic(hex2Bytes("0x7f6190116d118d643a98878e294ccf62b509e214299931aad8ff9764181a4e33")),
		types.BandersnatchPublic(hex2Bytes("0x48e5fcdce10e0b64ec4eebd0d9211c7bac2f27ce54bca6f7776ff6fee86ab3e3")),
		types.BandersnatchPublic(hex2Bytes("0xf16e5352840afb47e206b5c89f560f2611835855cf2e6ebad1acc9520a72591d")),
	}

	var proverIdx uint = 0
	vrfHandler, _ := CreateRingVRFHandler(bandersnatchKeys, proverIdx)
	defer vrfHandler.Free()
	commitment, _ := vrfHandler.GetCommitment()

	if types.BandersnatchRingCommitment(commitment) != expectedCommitment {
		t.Errorf("Expected commitment %v, got %v", expectedCommitment, commitment)
	}
}

func TestKeyRotate(t *testing.T) {
	s := store.GetInstance()
	priorState := s.GetPriorStates()
	posteriorState := s.GetPosteriorStates()

	now := time.Now().UTC()
	timeInSecond := uint64(now.Sub(types.JamCommonEra).Seconds())
	tauPrime := types.TimeSlot(timeInSecond / uint64(types.SlotPeriod))

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

	fakeValidators := LoadFakeValidators()

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
	priorGammaZ := types.BandersnatchRingCommitment(hex2Bytes(gammaZ))

	priorState.SetKappa(priorKappa)
	priorState.SetLambda(priorLambda)
	priorState.SetIota(priorIota)
	priorState.SetGammaK(priorGammaK)
	priorState.SetGammaZ(priorGammaZ)

	s.GenerateGenesisState(priorState.GetState())

	KeyRotate()

	// Get posterior state
	posteriorState = s.GetPosteriorStates()
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
	fakeValidators := LoadFakeValidators()

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
	s := store.GetInstance()
	s.GetPosteriorStates().SetPsiO(types.OffendersMark{})

	newValidators := ReplaceOffenderKeys(validatorsData)

	// Check if the new validators data has the same length as the original
	// validators data
	if len(newValidators) != len(validatorsData) {
		t.Errorf("Expected newValidators to have %d elements, got %d", len(validatorsData), len(newValidators))
	}
}

func TestReplaceOffenderKeys(t *testing.T) {
	// Load fake validators
	fakeValidators := LoadFakeValidators()

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
	s := store.GetInstance()
	s.GetPosteriorStates().SetPsiO(types.OffendersMark{fakeValidators[0].Ed25519})

	newValidators := ReplaceOffenderKeys(validatorsData)

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
