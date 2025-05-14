package jamtests

import "github.com/New-JAMneration/JAM-Protocol/internal/types"

// Decode TraceState
func (s *TraceState) Decode(d *types.Decoder) error {
	var err error

	if err = s.StateRoot.Decode(d); err != nil {
		return err
	}

	return nil
}

// Decode TraceTestCase
func (t *TraceTestCase) Decode(d *types.Decoder) error {
	var err error

	if err = t.PreState.Decode(d); err != nil {
		return err
	}

	if err = t.Block.Decode(d); err != nil {
		return err
	}

	if err = t.PostState.Decode(d); err != nil {
		return err
	}

	return nil
}
