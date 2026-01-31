package jamtests

import (
	"fmt"
	"os"

	"github.com/New-JAMneration/JAM-Protocol/internal/blockchain"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	m "github.com/New-JAMneration/JAM-Protocol/internal/utilities/merklization"
)

var disableStateDiff = os.Getenv("DEBUG_STATE_DIFF") == "1"

func (s *TraceTestCase) Dump() error {
	// Add block, state
	blockchain.GetInstance().AddBlock(s.Block)

	return nil
}

func (s *TraceTestCase) GetPostState() interface{} {
	return s.PostState
}

func (s *TraceTestCase) GetOutput() interface{} {
	return nil
}

func (s *TraceTestCase) ExpectError() error {
	return nil
}

func (s *TraceTestCase) Validate() error {
	stateRoot := m.MerklizationState(blockchain.GetInstance().GetPosteriorStates().GetState())

	if stateRoot != s.PostState.StateRoot {
		if disableStateDiff {
			return fmt.Errorf("state root mismatch: got %x, want %x", stateRoot, s.PostState.StateRoot)
		}
		if err := s.CmpKeyVal(stateRoot); err != nil {
			return fmt.Errorf("compare key-val error: %w", err)
		}
	}

	return nil
}

func (s *TraceTestCase) CmpKeyVal(stateRoot types.StateRoot) error {
	keyVals, err := m.StateEncoder(blockchain.GetInstance().GetPosteriorStates().GetState())
	if err != nil {
		return fmt.Errorf("state encode keyVals failed: %w", err)
	}

	keyValDiffs, err := m.GetStateKeyValsDiff(s.PostState.KeyVals, keyVals)
	if err != nil {
		return fmt.Errorf("get state keyValsDiff failed: %w", err)
	}

	err = m.DebugStateKeyValsDiff(keyValDiffs)
	if err != nil {
		return fmt.Errorf("debug state keyValsDiffs failed: %w", err)
	}

	return nil
}
