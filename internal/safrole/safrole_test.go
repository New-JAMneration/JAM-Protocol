package safrole

import (
	"reflect"
	"testing"

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

	vrfHandler, _ := CreateVRFHandler(bandersnatchKeys)
	commitment, _ := vrfHandler.GetCommitment()

	if types.BandersnatchRingCommitment(commitment) != expectedCommitment {
		t.Errorf("Expected commitment %v, got %v", expectedCommitment, commitment)
	}
}

func TestKeyRotate(t *testing.T) {
	timeslot := types.TimeSlot(1)
	timeslotPrime := types.TimeSlot(12)
	s := store.GetInstance()
	state := s.GetPriorState()
	offendersMark := types.OffendersMark{}

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

	state.Kappa = priorKappa
	state.Lambda = priorLambda
	state.Iota = priorIota
	state.Gamma.GammaK = priorGammaK
	state.Gamma.GammaZ = priorGammaZ

	s.GenerateGenesisState(state)

	KeyRotate(timeslot, timeslotPrime, state, offendersMark)

	// Get posterior state
	posteriorState := s.GetPosteriorState()
	if !reflect.DeepEqual(posteriorState.Gamma.GammaK, priorIota) {
		t.Errorf("Expected GammaK to be %v, got %v", priorIota, posteriorState.Gamma.GammaK)
	}
	if !reflect.DeepEqual(posteriorState.Kappa, priorGammaK) {
		t.Errorf("Expected Kappa to be %v, got %v", priorGammaK, posteriorState.Kappa)
	}
	if !reflect.DeepEqual(posteriorState.Lambda, priorKappa) {
		t.Errorf("Expected Lambda to be %v, got %v", priorKappa, posteriorState.Lambda)
	}
	if posteriorState.Gamma.GammaZ != priorGammaZ {
		t.Errorf("Expected GammaZ to be %v, got %v", priorGammaZ, posteriorState.Gamma.GammaZ)
	}
}

// TODO: Add tests for GetNewSafroleState, UpdateBandersnatchKeyRoot
