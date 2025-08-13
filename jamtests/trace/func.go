package jamtests

import (
	"fmt"

	"github.com/New-JAMneration/JAM-Protocol/internal/store"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/merklization"
)

func (s *TraceTestCase) Dump() error {
	// Add block, state
	st := store.GetInstance()
	st.AddBlock(s.Block)

	// Update timeslot
	st.GetPosteriorStates().SetTau(s.Block.Header.Slot)

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
	stateRoot := merklization.MerklizationState(store.GetInstance().GetPosteriorStates().GetState())

	if stateRoot != s.PostState.StateRoot {
		return fmt.Errorf("state root different")
	}

	return nil
}

func (s *TraceTestCase) CmpKeyVal() ([]types.StateKeyValDiff, error) {
	keyVals, err := merklization.StateEncoder(store.GetInstance().GetPosteriorStates().GetState())
	if err != nil {
		return nil, fmt.Errorf("state encode keyVals failed")
	}

	keyValDiffs, err := merklization.GetStateKeyValsDiff(keyVals, s.PostState.KeyVals)
	if err != nil {
		return nil, fmt.Errorf("get state keyValsDiff failed")
	}

	return keyValDiffs, nil
}
