package store

import (
	"reflect"
	"testing"
)

func TestGetGenesisBlock(t *testing.T) {
	block, err := GetGenesisBlock()
	if err != nil {
		t.Fatalf("Error getting genesis block: %v", err)
	}

	blockJsonFile := "../../pkg/test_data/jamtestnet/chainspecs/blocks/genesis-tiny.json"
	expectedBlock, err := GetGenesisBlockFromJson(blockJsonFile)
	if err != nil {
		t.Fatalf("Error getting expected genesis block: %v", err)
	}

	// Compare the two blocks
	if !reflect.DeepEqual(block, expectedBlock) {
		t.Fatalf("Genesis block does not match expected block")
	}
}

func TestGetGenesisState(t *testing.T) {
	state, err := GetGenesisState()
	if err != nil {
		t.Fatalf("Error getting genesis state: %v", err)
	}

	stateJsonFile := "../../pkg/test_data/jamtestnet/chainspecs/state_snapshots/genesis-tiny.json"
	expectedState, err := GetGenesisStateFromJson(stateJsonFile)
	if err != nil {
		t.Fatalf("Error getting expected genesis state: %v", err)
	}

	// Compare the two states
	if !reflect.DeepEqual(state, expectedState) {
		t.Fatalf("Genesis state does not match expected state")
	}
}
