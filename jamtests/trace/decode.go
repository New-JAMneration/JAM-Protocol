package jamtests

import "github.com/New-JAMneration/JAM-Protocol/internal/types"

// Decode StateKeyVals
func (s *StateKeyVals) Decode(d *types.Decoder) error {
	var err error

	// Decode the length of the array
	length, err := d.DecodeLength()
	if err != nil {
		return err
	}

	if length == 0 {
		return nil
	}

	// Allocate space for the array
	*s = make(StateKeyVals, length)

	// Decode each element in the array
	for i := uint64(0); i < length; i++ {
		if err = (*s)[i].Key.Decode(d); err != nil {
			return err
		}

		if err = (*s)[i].Value.Decode(d); err != nil {
			return err
		}
	}

	return nil
}

// Decode TraceState
func (s *TraceState) Decode(d *types.Decoder) error {
	var err error

	if err = s.StateRoot.Decode(d); err != nil {
		return err
	}

	if err = s.KeyVals.Decode(d); err != nil {
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
