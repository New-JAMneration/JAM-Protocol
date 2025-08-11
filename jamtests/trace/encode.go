package jamtests

import "github.com/New-JAMneration/JAM-Protocol/internal/types"

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

// Encode Genesis
func (g *Genesis) Encode(e *types.Encoder) error {
	var err error

	if err = g.Header.Encode(e); err != nil {
		return err
	}

	if err = g.State.Encode(e); err != nil {
		return err
	}

	return nil
}
