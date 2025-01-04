package safrole

import (
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

func TestGetEpochIndex(t *testing.T) {
	// Mock types.EpochLength
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
}

func TestGetSlotIndex(t *testing.T) {
	// Mock types.EpochLength
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

// TODO: Add tests for GetNewSafroleState, UpdateBandersnatchKeyRoot and
// GetBandersnatchRingRootCommmitment
