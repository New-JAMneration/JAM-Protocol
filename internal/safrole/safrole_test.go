package safrole

import (
	"testing"

	jamTypes "github.com/New-JAMneration/JAM-Protocol/internal/jam_types"
)

func TestGetEpochIndexEpochLength600(t *testing.T) {
	jamTypes.EpochLength = 600

	testCases := []struct {
		t        jamTypes.TimeSlot
		expected jamTypes.TimeSlot
	}{
		{0, 0},
		{1, 0},
		{2, 0},
		{599, 0},
		{600, 1},
		{601, 1},
		{1199, 1},
		{1200, 2},
	}

	for _, testCase := range testCases {
		if actual := GetEpochIndex(testCase.t); actual != testCase.expected {
			t.Errorf("GetEpochIndex(%d) = %d, expected %d", testCase.t, actual, testCase.expected)
		}
	}
}

func TestGetEpochIndexEpochLength12(t *testing.T) {
	jamTypes.EpochLength = 12

	testCases := []struct {
		t        jamTypes.TimeSlot
		expected jamTypes.TimeSlot
	}{
		{0, 0},
		{1, 0},
		{2, 0},
		{11, 0},
		{12, 1},
		{13, 1},
		{23, 1},
		{24, 2},
	}

	for _, testCase := range testCases {
		if actual := GetEpochIndex(testCase.t); actual != testCase.expected {
			t.Errorf("GetEpochIndex(%d) = %d, expected %d", testCase.t, actual, testCase.expected)
		}
	}
}

func TestGetSlotIndexEpochLength600(t *testing.T) {
	jamTypes.EpochLength = 600

	testCases := []struct {
		t        jamTypes.TimeSlot
		expected jamTypes.TimeSlot
	}{
		{0, 0},
		{1, 1},
		{2, 2},
		{599, 599},
		{600, 0},
		{601, 1},
		{1199, 599},
		{1200, 0},
	}

	for _, testCase := range testCases {
		if actual := GetSlotIndex(testCase.t); actual != testCase.expected {
			t.Errorf("GetSlotIndex(%d) = %d, expected %d", testCase.t, actual, testCase.expected)
		}
	}
}

func TestGetSlotIndexEpochLength12(t *testing.T) {
	jamTypes.EpochLength = 12

	testCases := []struct {
		t        jamTypes.TimeSlot
		expected jamTypes.TimeSlot
	}{
		{0, 0},
		{1, 1},
		{2, 2},
		{11, 11},
		{12, 0},
		{13, 1},
		{23, 11},
		{24, 0},
	}

	for _, testCase := range testCases {
		if actual := GetSlotIndex(testCase.t); actual != testCase.expected {
			t.Errorf("GetSlotIndex(%d) = %d, expected %d", testCase.t, actual, testCase.expected)
		}
	}
}

func TestValidatorIsOffender(t *testing.T) {
	offendersMark := jamTypes.OffendersMark{}
	offenderValidator := jamTypes.Validator{
		Bandersnatch: jamTypes.BandersnatchPublic{},
		Ed25519:      jamTypes.Ed25519Public{1, 2, 3},
		Bls:          jamTypes.BlsPublic{},
		Metadata:     jamTypes.ValidatorMetadata{},
	}
	offendersMark = append(offendersMark, offenderValidator.Ed25519)

	testCases := []struct {
		validator  ValidatorData
		offenders  jamTypes.OffendersMark
		isOffender bool
	}{
		{
			ValidatorData{
				Bandersnatch: BandersnatchKey{},
				Ed25519:      Ed25519Key{1, 2, 3},
				Bls:          BlsKey{},
				Metadata:     [128]U8{},
			},
			offendersMark,
			true,
		},
		{
			ValidatorData{
				Bandersnatch: BandersnatchKey{},
				Ed25519:      Ed25519Key{1, 2, 2},
				Bls:          BlsKey{},
				Metadata:     [128]U8{},
			},
			offendersMark,
			false,
		},
		{
			ValidatorData{
				Bandersnatch: BandersnatchKey{},
				Ed25519:      Ed25519Key{2, 2, 2},
				Bls:          BlsKey{},
				Metadata:     [128]U8{},
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

// TODO: Add tests for GetNewSafroleState, UpdateBandersnatchKeyRoot and
// GetBandersnatchRingRootCommmitment
