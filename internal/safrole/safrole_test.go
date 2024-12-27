package safrole_test

import (
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/safrole"
	jamTypes "github.com/New-JAMneration/JAM-Protocol/internal/types"
)

func TestGetEpochIndex(t *testing.T) {
	// Mock jamTypes.EpochLength
	jamTypes.EpochLength = 10

	// Test various time slot inputs
	tests := []struct {
		input    jamTypes.TimeSlot
		expected jamTypes.TimeSlot
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
}

func TestGetSlotIndex(t *testing.T) {
	// Mock jamTypes.EpochLength
	jamTypes.EpochLength = 10

	// Test various time slot inputs
	tests := []struct {
		input    jamTypes.TimeSlot
		expected jamTypes.TimeSlot
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
}
