package main

import (
	"encoding/hex"
	"fmt"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

func parseStateRoot(stateRootStr string) (types.StateRoot, error) {
	// Remove 0x prefix if present
	if len(stateRootStr) > 2 && stateRootStr[:2] == "0x" {
		stateRootStr = stateRootStr[2:]
	}

	var stateRoot types.StateRoot
	hashBytes, err := hex.DecodeString(stateRootStr)
	if err != nil {
		return types.StateRoot{}, err
	}

	if len(hashBytes) != 32 {
		return types.StateRoot{}, fmt.Errorf("state root must be 32 bytes, got %d bytes", len(hashBytes))
	}

	copy(stateRoot[:], hashBytes)
	return stateRoot, nil
}
