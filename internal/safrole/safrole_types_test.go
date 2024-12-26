package safrole

import (
	"fmt"
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

func TestEpochKeysValidate(t *testing.T) {
	// Mock jam_types.EpochLength
	jam_types.EpochLength = 5

	// Test case 1: correct number of items
	keys := EpochKeys(make([]BandersnatchKey, jam_types.EpochLength))
	err := keys.Validate()
	if err != nil {
		t.Errorf("EpochKeys.Validate returned an unexpected error for valid input: %v", err)
	}

	// Test case 2: incorrect number of items
	keys = EpochKeys(make([]BandersnatchKey, jam_types.EpochLength-1))
	err = keys.Validate()
	if err == nil {
		t.Error("EpochKeys.Validate did not return an error for invalid input")
	} else if err.Error() != fmt.Sprintf("EpochKeys must have exactly %d entries, got %d", jam_types.EpochLength, len(keys)) {
		t.Errorf("Unexpected error message: got '%v'", err)
	}
}

func TestTicketsBodiesValidate(t *testing.T) {
	// Mock jam_types.EpochLength
	jam_types.EpochLength = 5

	// Test case 1: correct number of items
	tickets := TicketsBodies(make([]TicketBody, jam_types.EpochLength))
	err := tickets.Validate()
	if err != nil {
		t.Errorf("TicketsBodies.Validate returned an unexpected error for valid input: %v", err)
	}

	// Test case 2: incorrect number of items
	tickets = TicketsBodies(make([]TicketBody, jam_types.EpochLength-1))
	err = tickets.Validate()
	if err == nil {
		t.Error("TicketsBodies.Validate did not return an error for invalid input")
	} else if err.Error() != fmt.Sprintf("TicketsBodies must have exactly %d entries, got %d", jam_types.EpochLength, len(tickets)) {
		t.Errorf("Unexpected error message: got '%v'", err)
	}
}

func TestEpochMarkValidate(t *testing.T) {
	// Mock jam_types.ValidatorsCount
	jam_types.ValidatorsCount = 3

	// Test case 1: 正確的數量
	mark := EpochMark{
		Validators: make([]BandersnatchKey, jam_types.ValidatorsCount),
	}
	err := mark.Validate()
	if err != nil {
		t.Errorf("EpochMark.Validate returned an unexpected error for valid input: %v", err)
	}

	// Test case 2: 錯誤的數量
	mark.Validators = make([]BandersnatchKey, jam_types.ValidatorsCount-1)
	err = mark.Validate()
	if err == nil {
		t.Error("EpochMark.Validate did not return an error for invalid input")
	} else if err.Error() != fmt.Sprintf("EpochMark must have exactly %d validators, got %d", jam_types.ValidatorsCount, len(mark.Validators)) {
		t.Errorf("Unexpected error message: got '%v'", err)
	}
}

func TestStateIntegrity(t *testing.T) {
	// Mock jam_types.ValidatorsCount
	jam_types.ValidatorsCount = 3
	jam_types.EpochLength = 2

	// Create a valid state
	state := State{
		Tau: 10,
		Eta: [4]OpaqueHash{},
		Lambda: ValidatorsData{
			{Bandersnatch: BandersnatchKey{}, Ed25519: Ed25519Key{}, Bls: BlsKey{}, Metadata: [128]U8{}},
			{Bandersnatch: BandersnatchKey{}, Ed25519: Ed25519Key{}, Bls: BlsKey{}, Metadata: [128]U8{}},
			{Bandersnatch: BandersnatchKey{}, Ed25519: Ed25519Key{}, Bls: BlsKey{}, Metadata: [128]U8{}},
		},
	}

	// Test Lambda validation
	err := state.Lambda.Validate()
	if err != nil {
		t.Errorf("ValidatorsData.Validate returned an unexpected error for valid input: %v", err)
	}

	// Modify Lambda to make it invalid
	state.Lambda = ValidatorsData{
		{Bandersnatch: BandersnatchKey{}, Ed25519: Ed25519Key{}, Bls: BlsKey{}, Metadata: [128]U8{}},
	}
	err = state.Lambda.Validate()
	if err == nil {
		t.Error("ValidatorsData.Validate did not return an error for invalid input")
	} else if err.Error() != fmt.Sprintf("ValidatorsData must have exactly %d entries, got %d", jam_types.ValidatorsCount, len(state.Lambda)) {
		t.Errorf("Unexpected error message: got '%v'", err)
	}
}
