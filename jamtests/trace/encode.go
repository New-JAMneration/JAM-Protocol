package jamtests

import "github.com/New-JAMneration/JAM-Protocol/internal/types"

// Encode StateKeyVals
func (s *StateKeyVals) Encode(e *types.Encoder) error {
	var err error

	// Encode the length of the array
	if err = e.EncodeLength(uint64(len(*s))); err != nil {
		return err
	}

	// Encode each element in the array
	for i := range *s {
		if err = (*s)[i].Key.Encode(e); err != nil {
			return err
		}

		if err = (*s)[i].Value.Encode(e); err != nil {
			return err
		}
	}

	return nil
}

// Encode TraceState
func (s *TraceState) Encode(e *types.Encoder) error {
	var err error

	if err = s.StateRoot.Encode(e); err != nil {
		return err
	}

	if err = s.KeyVals.Encode(e); err != nil {
		return err
	}

	return nil
}

// Encode TraceTestCase
func (t *TraceTestCase) Encode(e *types.Encoder) error {
	var err error

	if err = t.PreState.Encode(e); err != nil {
		return err
	}

	if err = t.Block.Encode(e); err != nil {
		return err
	}

	if err = t.PostState.Encode(e); err != nil {
		return err
	}

	return nil
}
