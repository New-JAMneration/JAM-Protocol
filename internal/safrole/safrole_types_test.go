package safrole

import (
	"fmt"
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/jam_types"
	"github.com/stretchr/testify/assert"
)

func TestEpochKeysValidate(t *testing.T) {
	// Mock jam_types.EpochLength
	jam_types.EpochLength = 5

	// Test case 1: correct number of items
	keys := EpochKeys(make([]BandersnatchKey, jam_types.EpochLength))
	err := keys.Validate()
	assert.NoError(t, err, "EpochKeys.Validate should not return an error for valid input")

	// Test case 2: incorrect number of items
	keys = EpochKeys(make([]BandersnatchKey, jam_types.EpochLength-1))
	err = keys.Validate()
	assert.Error(t, err, "EpochKeys.Validate should return an error for invalid input")
	assert.EqualError(t, err, fmt.Sprintf("EpochKeys must have exactly %d entries, got %d", jam_types.EpochLength, len(keys)))
}

func TestTicketsBodiesValidate(t *testing.T) {
	// Mock jam_types.EpochLength
	jam_types.EpochLength = 5

	// Test case 1: correct number of items
	tickets := TicketsBodies(make([]TicketBody, jam_types.EpochLength))
	err := tickets.Validate()
	assert.NoError(t, err, "TicketsBodies.Validate should not return an error for valid input")

	// Test case 2: incorrect number of items
	tickets = TicketsBodies(make([]TicketBody, jam_types.EpochLength-1))
	err = tickets.Validate()
	assert.Error(t, err, "TicketsBodies.Validate should return an error for invalid input")
	assert.EqualError(t, err, fmt.Sprintf("TicketsBodies must have exactly %d entries, got %d", jam_types.EpochLength, len(tickets)))
}

func TestEpochMarkValidate(t *testing.T) {
	// Mock jam_types.ValidatorsCount
	jam_types.ValidatorsCount = 3

	// Test case 1: 正確的數量
	mark := EpochMark{
		Validators: make([]BandersnatchKey, jam_types.ValidatorsCount),
	}
	err := mark.Validate()
	assert.NoError(t, err, "EpochMark.Validate should not return an error for valid input")

	// Test case 2: 錯誤的數量
	mark.Validators = make([]BandersnatchKey, jam_types.ValidatorsCount-1)
	err = mark.Validate()
	assert.Error(t, err, "EpochMark.Validate should return an error for invalid input")
	assert.EqualError(t, err, fmt.Sprintf("EpochMark must have exactly %d validators, got %d", jam_types.ValidatorsCount, len(mark.Validators)))
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
	assert.NoError(t, err, "ValidatorsData.Validate should not return an error for valid input")

	// Modify Lambda to make it invalid
	state.Lambda = ValidatorsData{
		{Bandersnatch: BandersnatchKey{}, Ed25519: Ed25519Key{}, Bls: BlsKey{}, Metadata: [128]U8{}},
	}
	err = state.Lambda.Validate()
	assert.Error(t, err, "ValidatorsData.Validate should return an error for invalid input")
	assert.EqualError(t, err, fmt.Sprintf("ValidatorsData must have exactly %d entries, got %d", jam_types.ValidatorsCount, len(state.Lambda)))
}
